package service

import (
	"Hyper/dao"
	"Hyper/models"
	"Hyper/pkg/snowflake"
	"Hyper/types"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type MessageService struct {
	MessageDao     *dao.MessageDAO
	UserService    IUserService
	GroupMemberDAO *dao.GroupMember
	GroupDAO       *dao.Group
	MqProducer     rmq_client.Producer
	Redis          *redis.Client
	DB             *gorm.DB
	NoteDAO        *dao.NoteDAO

	// 私信平台开关的短 TTL 进程内缓存，避免每条消息都查 platform_setting 表
	dmCacheMu      sync.Mutex
	dmCacheExpires time.Time
	dmCacheEnabled bool
	dmCacheSvcUID  int64
}

var (
	ErrGroupNotFound         = errors.New("群不存在或已解散")
	ErrNotGroupMember        = errors.New("你不在群内或已退群")
	ErrDirectMessageDisabled = errors.New("平台已关闭私信功能")
)

const (
	// sendMessageTimeout 约束 SendMessage 全链路（DB 校验/卡片补全/MQ 发送）
	sendMessageTimeout = 10 * time.Second
	// dmSettingsCacheTTL 私信开关进程内缓存的有效期
	dmSettingsCacheTTL = 30 * time.Second
)

var _ IMessageService = (*MessageService)(nil)

type IMessageService interface {
	SaveMessage(msg *models.ImSingleMessage) error
	SaveSingleMessage(msg *models.ImSingleMessage) error
	SaveGroupMessage(msg *models.ImGroupMessage) error
	SendMessage(msg *types.Message) error
	DeleteMessageForUser(ctx context.Context, userID uint64, sessionType int, peerID uint64, messageID int64) error
	ListMessages(ctx context.Context, userId, peerId uint64, sessionType int, cursor int64, since int64, limit int) ([]types.ListMessageReq, error)
}

func (s *MessageService) SaveMessage(msg *models.ImSingleMessage) error {
	// 执行插入
	return s.MessageDao.Save(msg)
}

func (s *MessageService) ListMessages(ctx context.Context, userId, peerId uint64, sessionType int, cursor int64, since int64, limit int) ([]types.ListMessageReq, error) {
	// 1) 参数兜底：接口必须有硬限制，避免被人传超大 limit 拖垮 DB
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	// 2) 分流：私聊查 im_single_messages；群聊查 im_group_messages
	switch sessionType {

	case types.SessionTypeSingle:
		// 私聊：用双方 uid 算 session_hash，保证 A->B 与 B->A 一致
		sessionHash := GetSessionHash(int64(userId), int64(peerId))

		q := s.DB.WithContext(ctx).
			Model(&models.ImSingleMessage{}).
			Where("session_hash = ?", sessionHash)
		q, visibilityErr := s.applyMessageVisibilityFilters(ctx, q, "im_single_messages", userId, sessionType, peerId)
		if visibilityErr != nil {
			return nil, visibilityErr
		}

		// 分页逻辑：since 模式向前拉新；cursor 模式向上翻旧
		if since > 0 {
			q = q.Where("created_at > ?", since).Order("created_at ASC")
		} else {
			if cursor > 0 {
				q = q.Where("created_at < ?", cursor)
			}
			q = q.Order("created_at DESC")
		}

		var msgs []models.ImSingleMessage
		if err := q.Limit(limit).Find(&msgs).Error; err != nil {
			return nil, err
		}
		// cursor 模式查的是 DESC，需要翻转为时间正序返回给前端
		if since <= 0 {
			for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
				msgs[i], msgs[j] = msgs[j], msgs[i]
			}
		}
		// 统一映射为 types.ListMessageReq
		result := make([]types.ListMessageReq, 0, len(msgs))
		for _, m := range msgs {
			result = append(result, types.ListMessageReq{
				Id:       strconv.FormatInt(m.Id, 10),
				SenderId: uint64(m.SenderId),
				Content:  m.Content,
				MsgType:  m.MsgType,
				Ext:      decodeMessageExt(m.Ext),
				Time:     m.CreatedAt,
				IsSelf:   m.SenderId == int64(userId),
			})
		}
		return result, nil
	case types.GroupChatSessionTypeGroup:
		// 群聊：peerId 在这里就是 groupId
		// 群消息不是公开资源。读取历史与发消息一样，必须校验群仍有效且
		// 当前用户是未退群成员，不能只凭猜到的 group_id 查询消息。
		if s.GroupDAO == nil || s.GroupMemberDAO == nil {
			return nil, errors.New("群聊服务未初始化")
		}
		if _, err := s.GroupDAO.GetGroup(ctx, int(peerId)); err != nil {
			return nil, ErrGroupNotFound
		}
		if !s.GroupMemberDAO.IsMember(ctx, int(peerId), int(userId), false) {
			return nil, ErrNotGroupMember
		}

		// conn-server 写库时就是按 GetGroupSessionHash(groupId) 填的 SessionHash
		sessionHash := GetGroupSessionHash(int64(peerId))

		q := s.DB.WithContext(ctx).
			Model(&models.ImGroupMessage{}).
			Where("session_hash = ?", sessionHash)
		q, visibilityErr := s.applyMessageVisibilityFilters(ctx, q, "im_group_messages", userId, sessionType, peerId)
		if visibilityErr != nil {
			return nil, visibilityErr
		}

		if since > 0 {
			q = q.Where("created_at > ?", since).Order("created_at ASC")
		} else {
			if cursor > 0 {
				q = q.Where("created_at < ?", cursor)
			}
			q = q.Order("created_at DESC")
		}

		var msgs []models.ImGroupMessage
		if err := q.Limit(limit).Find(&msgs).Error; err != nil {
			return nil, err
		}

		if since <= 0 {
			for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
				msgs[i], msgs[j] = msgs[j], msgs[i]
			}
		}

		result := make([]types.ListMessageReq, 0, len(msgs))

		userIds := make([]uint64, 0, len(msgs))
		userMapT := make(map[int64]struct{})

		for i := 0; i < len(msgs); i++ {
			uID := msgs[i].SenderId
			if _, exists := userMapT[uID]; !exists {
				userMapT[uID] = struct{}{}
				userIds = append(userIds, uint64(uID))
			}
		}
		userInfo := s.UserService.BatchGetUserInfo(ctx, userIds)
		for _, m := range msgs {
			result = append(result, types.ListMessageReq{
				Id:       strconv.FormatInt(m.Id, 10),
				Nickname: userInfo[uint64(m.SenderId)].Nickname,
				Avatar:   userInfo[uint64(m.SenderId)].Avatar,
				SenderId: uint64(m.SenderId),
				Content:  m.Content,
				MsgType:  m.MsgType,
				Ext:      decodeMessageExt(m.Ext),
				Time:     m.CreatedAt,
				IsSelf:   m.SenderId == int64(userId),
			})
		}
		return result, nil

	default:
		return nil, fmt.Errorf("invalid session_type=%d (only 1 or 2)", sessionType)
	}
}

// DeleteMessageForUser hides one message only for the current user. It does
// not revoke the message for the peer or remove it from operational storage.
func (s *MessageService) DeleteMessageForUser(ctx context.Context, userID uint64, sessionType int, peerID uint64, messageID int64) error {
	if userID == 0 || peerID == 0 || messageID <= 0 {
		return errors.New("消息或会话参数无效")
	}
	if err := s.ensureMessageSessionAccess(ctx, userID, sessionType, peerID); err != nil {
		return err
	}

	sessionHash := GetSessionHash(int64(userID), int64(peerID))
	if sessionType == types.GroupChatSessionTypeGroup {
		sessionHash = GetGroupSessionHash(int64(peerID))
	}

	var exists bool
	switch sessionType {
	case types.SessionTypeSingle:
		var message models.ImSingleMessage
		err := s.DB.WithContext(ctx).Where("id = ? AND session_hash = ?", messageID, sessionHash).First(&message).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("消息不存在或不属于当前会话")
		}
		if err != nil {
			return err
		}
		exists = true
	case types.GroupChatSessionTypeGroup:
		var message models.ImGroupMessage
		err := s.DB.WithContext(ctx).Where("id = ? AND session_hash = ?", messageID, sessionHash).First(&message).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("消息不存在或不属于当前会话")
		}
		if err != nil {
			return err
		}
		exists = true
	default:
		return fmt.Errorf("invalid session_type=%d (only 1 or 2)", sessionType)
	}
	if !exists {
		return nil
	}

	deletion := models.ImMessageUserDeletion{MessageID: messageID, SessionType: sessionType, UserID: userID}
	if err := s.DB.WithContext(ctx).Where("message_id = ? AND session_type = ? AND user_id = ?", messageID, sessionType, userID).FirstOrCreate(&deletion).Error; err != nil {
		return err
	}
	return s.refreshSessionPreviewAfterMessageDelete(ctx, userID, sessionType, peerID, messageID, sessionHash)
}

func (s *MessageService) applyMessageVisibilityFilters(ctx context.Context, query *gorm.DB, table string, userID uint64, sessionType int, peerID uint64) (*gorm.DB, error) {
	var session models.Session
	err := s.DB.WithContext(ctx).
		Select("cleared_at").
		Where("user_id = ? AND session_type = ? AND peer_id = ?", userID, sessionType, peerID).
		First(&session).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if session.ClearedAt > 0 {
		query = query.Where("created_at > ?", session.ClearedAt)
	}
	query = query.Where(
		"NOT EXISTS (SELECT 1 FROM im_message_user_deletions d WHERE d.message_id = "+table+".id AND d.session_type = ? AND d.user_id = ?)",
		sessionType,
		userID,
	)
	return query, nil
}

func (s *MessageService) ensureMessageSessionAccess(ctx context.Context, userID uint64, sessionType int, peerID uint64) error {
	if sessionType == types.SessionTypeSingle {
		return nil
	}
	if sessionType != types.GroupChatSessionTypeGroup {
		return fmt.Errorf("invalid session_type=%d (only 1 or 2)", sessionType)
	}
	if s.GroupDAO == nil || s.GroupMemberDAO == nil {
		return errors.New("群聊服务未初始化")
	}
	if _, err := s.GroupDAO.GetGroup(ctx, int(peerID)); err != nil {
		return ErrGroupNotFound
	}
	if !s.GroupMemberDAO.IsMember(ctx, int(peerID), int(userID), false) {
		return ErrNotGroupMember
	}
	return nil
}

func (s *MessageService) refreshSessionPreviewAfterMessageDelete(ctx context.Context, userID uint64, sessionType int, peerID uint64, messageID, sessionHash int64) error {
	var session models.Session
	if err := s.DB.WithContext(ctx).Where("user_id = ? AND session_type = ? AND peer_id = ?", userID, sessionType, peerID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if int64(session.LastMsgId) != messageID {
		return nil
	}

	updates := map[string]interface{}{"last_msg_hidden": 1}
	if sessionType == types.SessionTypeSingle {
		var previous models.ImSingleMessage
		query := s.DB.WithContext(ctx).Where("session_hash = ?", sessionHash).Where("created_at > ?", session.ClearedAt).
			Where("NOT EXISTS (SELECT 1 FROM im_message_user_deletions d WHERE d.message_id = im_single_messages.id AND d.session_type = ? AND d.user_id = ?)", sessionType, userID).
			Order("created_at DESC")
		err := query.First(&previous).Error
		if err == nil {
			updates["last_msg_id"] = previous.Id
			updates["last_msg_type"] = previous.MsgType
			updates["last_msg_content"] = previous.Content
			updates["last_msg_time"] = previous.CreatedAt
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			updates["last_msg_id"] = 0
			updates["last_msg_type"] = 0
			updates["last_msg_content"] = ""
			updates["last_msg_time"] = 0
		} else {
			return err
		}
	} else {
		var previous models.ImGroupMessage
		query := s.DB.WithContext(ctx).Where("session_hash = ?", sessionHash).Where("created_at > ?", session.ClearedAt).
			Where("NOT EXISTS (SELECT 1 FROM im_message_user_deletions d WHERE d.message_id = im_group_messages.id AND d.session_type = ? AND d.user_id = ?)", sessionType, userID).
			Order("created_at DESC")
		err := query.First(&previous).Error
		if err == nil {
			updates["last_msg_id"] = previous.Id
			updates["last_msg_type"] = previous.MsgType
			updates["last_msg_content"] = previous.Content
			updates["last_msg_time"] = previous.CreatedAt
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			updates["last_msg_id"] = 0
			updates["last_msg_type"] = 0
			updates["last_msg_content"] = ""
			updates["last_msg_time"] = 0
		} else {
			return err
		}
	}
	return s.DB.WithContext(ctx).Model(&models.Session{}).Where("id = ?", session.Id).Updates(updates).Error
}

func (s *MessageService) SaveSingleMessage(msg *models.ImSingleMessage) error {
	return s.MessageDao.SaveSingle(msg)
}
func (s *MessageService) SaveGroupMessage(msg *models.ImGroupMessage) error {
	return s.MessageDao.SaveGroup(msg)
}
func (s *MessageService) SendMessage(msg *types.Message) error {
	if msg == nil {
		return errors.New("消息不能为空")
	}
	if err := normalizeMessagePayload(msg); err != nil {
		return err
	}
	// SendMessage 没有请求级 ctx（由 ws/内部调用触发），但不能裸用
	// context.Background()：DB 校验/卡片补全/MQ Send 任一环节挂死都会永久阻塞。
	ctx, cancel := context.WithTimeout(context.Background(), sendMessageTimeout)
	defer cancel()
	if msg.SessionType == types.SessionTypeSingle {
		allowed, err := s.directMessageAllowed(ctx, msg.SenderID, msg.TargetID)
		if err != nil {
			return err
		}
		if !allowed {
			return ErrDirectMessageDisabled
		}
	}

	// 1) 服务端统一补字段：时间、状态、雪花ID、Ext兜底
	msg.Timestamp = time.Now().UnixMilli()
	msg.Status = types.MsgStatusSending
	msg.Id = snowflake.GenID()

	if msg.Ext == nil {
		msg.Ext = make(map[string]interface{})
	}

	// 2) 生成“会话标识”
	// SessionID：用于展示/调试，稳定可读
	// SessionHash：用于数据库索引/分表/查询（通常是整型）
	switch msg.SessionType {
	case types.SessionTypeSingle:
		// 单聊：用双方uid生成稳定会话（A->B 和 B->A 一样）
		msg.SessionID = s.generateSessionID(msg.SenderID, msg.TargetID)
		msg.SessionHash = GetSessionHash(msg.SenderID, msg.TargetID)

	case types.GroupChatSessionTypeGroup:
		// 群聊：TargetID 此时就是 groupId
		msg.SessionID = fmt.Sprintf("g_%d", msg.TargetID)
		msg.SessionHash = GetGroupSessionHash(msg.TargetID)

	default:
		return fmt.Errorf("unknown session_type=%d", msg.SessionType)
	}
	// 3) 群聊禁言校验：必须在发 MQ 之前做
	if msg.SessionType == types.GroupChatSessionTypeGroup {

		gid := int(msg.TargetID) // 群ID
		uid := int(msg.SenderID) // 发送者ID

		// 3.1) 查群成员记录：是否成员/是否退群/角色/个人禁言
		m, err := s.GroupMemberDAO.FindByUserId(ctx, gid, uid)
		if err != nil || m.IsQuit == 1 {
			return fmt.Errorf("你不在群内或已退群")
		}

		// 3.2) 个人禁言：优先级最高
		if m.IsMute == 1 {
			return fmt.Errorf("你已被禁言")
		}

		// 3.3) 群全员禁言：只禁普通成员(role=3)，群主/管理员仍可发言
		g, err := s.GroupDAO.FindByID(ctx, gid)
		if err != nil {
			return fmt.Errorf("群不存在")
		}
		if g.IsMuteAll == 1 && m.Role == 3 {
			return fmt.Errorf("群已开启全员禁言")
		}
	}

	// 4) 频道（给 ws / 路由用）
	msg.Channel = types.ChannelChat

	// 5) 卡片消息：转发帖子卡片（服务端补全卡片信息，防止前端伪造）
	if msg.MsgType == types.MsgTypeCard {
		if err := s.fillNoteForwardCard(ctx, msg); err != nil {
			return err
		}
	}
	if msg.MsgType == types.MsgTypeActivity {
		if err := s.fillActivityCard(ctx, msg); err != nil {
			return err
		}
	}

	// 6) 发 MQ
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	mqMsg := &rmq_client.Message{
		Topic: types.ImTopicChat,
		Body:  body,
	}

	_, err = s.MqProducer.Send(ctx, mqMsg)
	if err != nil {
		return err
	}

	return nil
}

// directMessageAllowed reads the platform switch with a short in-process TTL
// cache (30s) so per-message sends do not hit platform_setting every time;
// configuration changes take effect within the TTL. The configured
// customer-service account remains available when ordinary user DMs are closed.
func (s *MessageService) directMessageAllowed(ctx context.Context, senderID, targetID int64) (bool, error) {
	s.dmCacheMu.Lock()
	if time.Now().Before(s.dmCacheExpires) {
		enabled, serviceUserID := s.dmCacheEnabled, s.dmCacheSvcUID
		s.dmCacheMu.Unlock()
		if enabled {
			return true, nil
		}
		return serviceUserID > 0 && (senderID == serviceUserID || targetID == serviceUserID), nil
	}
	s.dmCacheMu.Unlock()

	var settings []models.PlatformSetting
	if err := s.DB.WithContext(ctx).Where("setting_key IN ?", []string{"direct_message_enabled", "customer_service_user_id"}).Find(&settings).Error; err != nil {
		return false, err
	}
	values := make(map[string]string, len(settings))
	for _, setting := range settings {
		values[setting.Key] = setting.Value
	}
	enabled := platformBoolEnabled(values["direct_message_enabled"], true)
	serviceUserID, _ := strconv.ParseInt(values["customer_service_user_id"], 10, 64)

	s.dmCacheMu.Lock()
	s.dmCacheEnabled = enabled
	s.dmCacheSvcUID = serviceUserID
	s.dmCacheExpires = time.Now().Add(dmSettingsCacheTTL)
	s.dmCacheMu.Unlock()

	if enabled {
		return true, nil
	}
	return serviceUserID > 0 && (senderID == serviceUserID || targetID == serviceUserID), nil
}

// normalizeMessagePayload keeps image messages compatible with the existing
// Content-based protocol while making image_url available to newer clients.
func normalizeMessagePayload(msg *types.Message) error {
	if msg.MsgType < types.MsgTypeText || msg.MsgType > types.MsgTypeActivity {
		return errors.New("不支持的消息类型")
	}
	if msg.Ext == nil {
		msg.Ext = make(map[string]interface{})
	}

	msg.Content = strings.TrimSpace(msg.Content)
	if msg.MsgType == types.MsgTypeImage {
		imageURL := msg.Content
		if rawURL, ok := msg.Ext[types.ExtKeyImageURL].(string); ok && strings.TrimSpace(rawURL) != "" {
			imageURL = strings.TrimSpace(rawURL)
		}
		if imageURL == "" {
			return errors.New("图片消息缺少 image_url")
		}
		parsedURL, err := url.ParseRequestURI(imageURL)
		if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" {
			return errors.New("图片消息 image_url 必须是有效的 http 或 https 地址")
		}
		if len([]rune(imageURL)) > 5000 {
			return errors.New("图片消息 image_url 不能超过 5000 个字符")
		}
		msg.Content = imageURL
		msg.Ext[types.ExtKeyImageURL] = imageURL
		return nil
	}

	if msg.Content == "" {
		return errors.New("消息内容不能为空")
	}
	if len([]rune(msg.Content)) > 5000 {
		return errors.New("消息内容不能超过 5000 个字符")
	}
	return nil
}

func (s *MessageService) generateSessionID(uid1, uid2 int64) string {
	if uid1 < uid2 {
		return fmt.Sprintf("%d_%d", uid1, uid2)
	}
	return fmt.Sprintf("%d_%d", uid2, uid1)
}

// GetSessionHash 用 FNV-1a 64bit 作为会话唯一键。
// 兼容性说明：该值已落库为历史消息的 session_hash，任何算法调整都会让
// 既有会话失效，因此【禁止修改】。需注意 FNV-1a 64bit 非加密哈希，理论
// 上存在碰撞导致串话的风险（64 位空间下概率极低）；如需根治，应通过新增
// 字段/迁移方案而不是改动此函数。
func GetSessionHash(uid1, uid2 int64) int64 {
	// 1. 保证 uid 顺序（从小到大），确保 A_B 和 B_A 生成的哈希一致
	var rawID string
	if uid1 < uid2 {
		rawID = fmt.Sprintf("%d_%d", uid1, uid2)
	} else {
		rawID = fmt.Sprintf("%d_%d", uid2, uid1)
	}

	// 2. 使用 FNV-1a 算法计算
	h := fnv.New64a()
	h.Write([]byte(rawID))

	// 3. 返回 int64 类型（强转 uint64 为 int64）
	return int64(h.Sum64())
}

func GetGroupSessionHash(groupID int64) int64 {
	// 给群加个固定前缀，避免和 "私聊" 的 hash 产生碰撞
	rawID := fmt.Sprintf("g_%d", groupID)
	h := fnv.New64a()
	_, _ = h.Write([]byte(rawID))
	return int64(h.Sum64())
}

func (s *MessageService) fillNoteForwardCard(ctx context.Context, msg *types.Message) error {
	if msg.Ext == nil {
		return errors.New("ext 不能为空")
	}

	// 1) 判断是不是 note_forward
	ct, _ := msg.Ext[types.ExtKeyCardType].(string)
	if ct != types.CardTypeNoteForward {
		// 不是转发帖子卡片，就不处理（以后还能扩展别的卡片）
		return nil
	}

	// 2) 取 note_id（注意前端可能传 string / float64）
	raw := msg.Ext[types.ExtKeyNoteID]
	var noteID uint64
	switch v := raw.(type) {
	case float64:
		noteID = uint64(v)
	case string:
		// string -> uint64
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return errors.New("note_id 非法")
		}
		noteID = parsed
	default:
		return errors.New("note_id 不能为空")
	}
	if noteID == 0 {
		return errors.New("note_id 不能为空")
	}
	msg.Ext[types.ExtKeyNoteID] = strconv.FormatUint(noteID, 10)
	if s.NoteDAO == nil {
		return errors.New("NoteDAO 未初始化")
	}

	// 3) 查 note
	note, err := s.NoteDAO.GetByID(ctx, noteID)
	if err != nil {
		return errors.New("帖子不存在")
	}

	cover := ""
	if note.MediaData != "" {
		var media []map[string]interface{}
		if err := json.Unmarshal([]byte(note.MediaData), &media); err == nil && len(media) > 0 {
			if u, ok := media[0]["thumbnail_url"].(string); ok && u != "" {
				cover = u
			} else if u, ok := media[0]["url"].(string); ok {
				cover = u
			}
		}
	}

	// 6) 查作者信息
	userMap := s.UserService.BatchGetUserInfo(ctx, []uint64{note.UserID})
	author := userMap[note.UserID]

	// 7) 回填 ext.note（注意：只由服务端写入，前端传的会被覆盖）
	msg.Ext["note"] = map[string]interface{}{
		"id":              fmt.Sprintf("%d", note.ID),
		"title":           note.Title,
		"cover":           cover,
		"author_id":       note.UserID,
		"author_avatar":   author.Avatar,
		"author_nickname": author.Nickname,
	}

	return nil
}

func (s *MessageService) fillActivityCard(ctx context.Context, msg *types.Message) error {
	if msg.Ext == nil {
		return errors.New("ext 不能为空")
	}
	// 1) 判断是不是 note_forward
	ct, _ := msg.Ext[types.ExtKeyCardType].(string)
	if ct != types.CardTypeActivityForward {
		// 不是转发帖子卡片，就不处理（以后还能扩展别的卡片）
		return nil
	}
	raw := msg.Ext[types.ExtKeyActivity]
	var id uint64
	switch v := raw.(type) {
	case float64:
		id = uint64(v)
	case string:
		// string -> uint64
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return errors.New("note_id 非法")
		}
		id = parsed
	default:
		return errors.New("note_id 不能为空")
	}
	var merchant models.Merchant

	if err := s.DB.Model(&models.Merchant{}).Where("id = ?", id).First(&merchant).Error; err != nil {
		return nil
	}
	msg.Ext[types.ExtKeyActivity] = strconv.FormatUint(id, 10)
	msg.Ext["party"] = map[string]interface{}{
		"id":            strconv.FormatUint(uint64(merchant.ID), 10), // 商家/活动ID
		"title":         merchant.Title,                              // 标题
		"type":          merchant.Type,                               // 类型 (如：剧本杀、徒步等)
		"cover_image":   merchant.CoverImage,                         // 封面图
		"location_name": merchant.LocationName,                       // 场地名称
		"address":       merchant.Address,                            // 详细地址
		"lat":           merchant.Latitude,                           // 纬度
		"lng":           merchant.Longitude,                          // 经度
		"status":        "active",                                    // 额外状态标识
	}
	return nil
}

// decodeMessageExt preserves JSON number tokens until known snowflake IDs are
// converted to strings. json.Unmarshal into interface{} would turn them into
// float64 and permanently lose precision before the HTTP response is written.
func decodeMessageExt(raw string) map[string]interface{} {
	ext := make(map[string]interface{})
	if raw == "" {
		return ext
	}
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&ext); err != nil {
		return map[string]interface{}{}
	}
	normalizeMessageExtIDs(ext)
	return ext
}

func normalizeMessageExtIDs(ext map[string]interface{}) {
	for _, key := range []string{types.ExtKeyNoteID, types.ExtKeyActivity} {
		if value, ok := ext[key]; ok {
			ext[key] = messageIDString(value)
		}
	}
	for _, key := range []string{types.ExtKeyNote, "party", "activity"} {
		item, ok := ext[key].(map[string]interface{})
		if !ok {
			continue
		}
		if value, ok := item["id"]; ok {
			item["id"] = messageIDString(value)
		}
	}
}

func messageIDString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	case float64:
		return strconv.FormatUint(uint64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case int:
		return strconv.Itoa(v)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	default:
		return fmt.Sprint(value)
	}
}
