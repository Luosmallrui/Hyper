package service

import (
	"Hyper/dao"
	"Hyper/dao/cache"
	"Hyper/models"
	"Hyper/types"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var _ ISessionService = (*SessionService)(nil)

type SessionService struct {
	DB             *gorm.DB
	MessageStorage *cache.MessageStorage
	UnreadStorage  *cache.UnreadStorage
	UserService    IUserService
	SessionDAO     *dao.SessionDAO
	GroupDAO       *dao.Group
}

func SessionMapKey(sessionType int, peerId uint64) string {
	return fmt.Sprintf("%d:%d", sessionType, peerId)
}

type ISessionService interface {
	UpdateSingleSession(ctx context.Context, msg *types.Message) error
	ListUserSessions(ctx context.Context, userId uint64, limit int) ([]*types.SessionDTO, error)
	UpsertGroupSessions(ctx context.Context, msg *types.Message, memberIDs []int) error
	UpdateSessionSettings(ctx context.Context, userID uint64, req *types.SessionSettingRequest) error
	ClearUnread(ctx context.Context, userId uint64, sessionType int, peerId uint64, readTime int64) error
	GetUnreadNum(ctx context.Context, userId int) (int64, error)
	//CreateSession(ctx context.Context, tx *gorm.DB, userId int, groupId uint64) (uint64, error)
}

func (s *SessionService) GetUnreadNum(ctx context.Context, userId int) (int64, error) {
	return s.SessionDAO.GetUnreadNum(ctx, userId)
}

func (s *SessionService) UpdateSingleSession(
	ctx context.Context,
	msg *types.Message,
) error {

	// 自己发给自己，通常不计未读
	if msg.SenderID == msg.TargetID {
		return nil
	}

	summary := truncateContent(msg.Content, 50)

	// 发送方会话（unread = 0）
	if err := s.upsertConversation(
		ctx,
		uint64(msg.SenderID),
		types.SessionTypeSingle,
		uint64(msg.TargetID),
		msg,
		summary,
		0,
	); err != nil {
		return err
	}

	// 接收方会话（unread +1）
	if err := s.upsertConversation(
		ctx,
		uint64(msg.TargetID),
		types.SessionTypeSingle,
		uint64(msg.SenderID),
		msg,
		summary,
		1,
	); err != nil {
		return err
	}

	return nil
}
func (s *SessionService) UpsertGroupSessions(ctx context.Context, msg *types.Message, memberIDs []int) error {
	// msg.TargetID = group_id（数字）
	groupID := uint64(msg.TargetID)
	senderID := uint64(msg.SenderID)

	lastTimeMs := msg.Timestamp
	if lastTimeMs <= 0 {
		lastTimeMs = time.Now().UnixMilli()
	}

	// last_msg_content 最大 255
	lastContent := msg.Content
	if len([]rune(lastContent)) > 200 { // 留点余量，避免 emoji 等导致超长
		r := []rune(lastContent)
		lastContent = string(r[:200]) + "..."
	}

	now := time.Now()

	rows := make([]models.Session, 0, len(memberIDs))
	for _, mid := range memberIDs {
		uid := uint64(mid)
		if uid == 0 {
			continue
		}

		// 未读最小版：发送者不加，其他成员 +1
		var unread uint32 = 0
		if uid != senderID {
			unread = 1
		}

		rows = append(rows, models.Session{
			UserId:      uid,
			SessionType: types.GroupChatSessionTypeGroup,
			PeerId:      groupID,

			LastMsgId:      uint64(msg.Id),
			LastMsgType:    msg.MsgType,
			LastMsgContent: lastContent,
			LastMsgTime:    lastTimeMs,

			UnreadCount: unread,
			UpdatedAt:   now,
			CreatedAt:   now,
		})
	}

	return s.SessionDAO.BatchUpsert(ctx, rows)
}

func truncateContent(content string, maxLen int) string {
	runes := []rune(content)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return content
}

func (s *SessionService) upsertConversation(
	ctx context.Context,
	userId uint64,
	sessionType int,
	peerId uint64,
	msg *types.Message,
	summary string,
	unreadDelta int,
) error {

	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var conv models.Session

		err := tx.Where(
			"user_id = ? AND session_type = ? AND peer_id = ?",
			userId, sessionType, peerId,
		).Take(&conv).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 新会话
				conv = models.Session{
					UserId:         userId,
					SessionType:    sessionType,
					PeerId:         peerId,
					LastMsgId:      uint64(msg.Id),
					LastMsgType:    msg.MsgType,
					LastMsgContent: summary,
					LastMsgTime:    msg.Timestamp,
					UnreadCount:    uint32(max(unreadDelta, 0)),
				}
				return tx.Create(&conv).Error
			}
			return err
		}

		// 已存在会话，更新
		update := map[string]interface{}{
			"last_msg_id":      msg.Id,
			"last_msg_type":    msg.MsgType,
			"last_msg_content": summary,
			"last_msg_time":    msg.Timestamp,
			"updated_at":       time.Now(),
		}

		if unreadDelta > 0 {
			update["unread_count"] = gorm.Expr(
				"unread_count + ?", unreadDelta,
			)
		}

		return tx.Model(&models.Session{}).
			Where("id = ?", conv.Id).
			Updates(update).Error
	})
}

func (s *SessionService) ListUserSessions(ctx context.Context, userId uint64, limit int) ([]*types.SessionDTO, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// 1. 批量查询会话表
	var convs []models.Session
	err := s.DB.WithContext(ctx).
		Where("user_id = ?", userId).
		Order("is_top DESC, last_msg_time DESC").
		Limit(limit).
		Find(&convs).Error
	if err != nil {
		return nil, err
	}

	if len(convs) == 0 {
		return []*types.SessionDTO{}, nil
	}

	// 2. 收集id
	peerIds := make([]uint64, 0, len(convs))
	groupIds := make([]uint64, 0, len(convs))

	for _, c := range convs {
		if c.SessionType == types.SessionTypeSingle {
			peerIds = append(peerIds, c.PeerId)
		} else if c.SessionType == types.GroupChatSessionTypeGroup {
			groupIds = append(groupIds, c.PeerId) // peer_id 就是 group_id
		}
	}

	// 3. 批量获取用户信息 (从 Redis + DB 回源)
	// 建议封装成我们之前讨论的 BatchGetUserInfo
	userInfoMap := s.UserService.BatchGetUserInfo(ctx, peerIds)
	groupInfoMap := map[uint64]*models.Group{}
	if s.GroupDAO != nil {
		m, err := s.GroupDAO.BatchGetByIDs(ctx, groupIds)
		if err == nil {
			groupInfoMap = m
		}
	}
	// 4. 批量获取 Redis 中的最后一条消息
	// 注意：这里推荐在 MessageService 内部实现一个 Pipeline 批量获取方法
	//DB作为权威未读
	//unreadMap := s.UnreadStorage.BatchGet(ctx, userId, convs)
	lastMsgMap := s.MessageStorage.BatchGet(ctx, userId, convs)

	// 5. 在内存中组装结果
	result := make([]*types.SessionDTO, 0, len(convs))
	for _, c := range convs {
		dto := &types.SessionDTO{
			SessionType: c.SessionType,
			PeerId:      c.PeerId,
			IsTop:       c.IsTop,
			IsMute:      c.IsMute,
			Unread:      c.UnreadCount, //DB作为权威未读
		}
		// A) 私聊：peer_id 是对方 uid，才去 userInfoMap 拿昵称头像
		if c.SessionType == types.SessionTypeSingle {
			if info, ok := userInfoMap[c.PeerId]; ok {
				dto.PeerName = info.Nickname
				dto.PeerAvatar = info.Avatar
			}
		} else {
			if g, ok := groupInfoMap[c.PeerId]; ok {
				dto.PeerName = g.Name
				dto.PeerAvatar = g.Avatar
			} else {
				dto.PeerName = fmt.Sprintf("群聊(%d)", c.PeerId)
				dto.PeerAvatar = ""
			}
		}

		// 优先使用 Redis 中的实时数据，但必须保证不被旧值覆盖 DB
		k := SessionMapKey(c.SessionType, c.PeerId)

		if last, ok := lastMsgMap[k]; ok {
			// 只有当 Redis 的时间 >= DB 才采用 Redis，避免 Redis 陈旧覆盖 DB
			if last.Timestamp >= c.LastMsgTime {
				dto.LastMsg = last.Content
				dto.LastMsgTime = last.Timestamp
			} else {
				dto.LastMsg = c.LastMsgContent
				dto.LastMsgTime = c.LastMsgTime
			}
		} else {
			dto.LastMsg = c.LastMsgContent
			dto.LastMsgTime = c.LastMsgTime
		}

		result = append(result, dto)
	}
	return result, nil
}
func (s *SessionService) UpdateSessionSettings(ctx context.Context, userID uint64, req *types.SessionSettingRequest) error {
	// service 再做一次防御性校验
	if req.SessionType != 1 && req.SessionType != 2 {
		return errors.New("session_type 必须是 1 或 2")
	}

	if req.PeerID == 0 {
		return errors.New("peer_id 不能为空")
	}

	// 指针校验：既要“必须传”，又要“允许 0”
	if req.IsTop == nil {
		return errors.New("is_top 不能为空")
	}
	if *req.IsTop != 0 && *req.IsTop != 1 {
		return errors.New("is_top 只能是 0 或 1")
	}

	if req.IsMute == nil {
		return errors.New("is_mute 不能为空")
	}
	if *req.IsMute != 0 && *req.IsMute != 1 {
		return errors.New("is_mute 只能是 0 或 1")
	}

	isTop := *req.IsTop
	isMute := *req.IsMute

	return s.SessionDAO.UpsertSessionSettings(ctx, userID, req.SessionType, req.PeerID, isTop, isMute)
}

func (s *SessionService) ClearUnread(ctx context.Context, userId uint64, sessionType int, peerId uint64, readTime int64) error {
	q := s.DB.WithContext(ctx).
		Model(&models.Session{}).
		Where("user_id = ? AND session_type = ? AND peer_id = ?", userId, sessionType, peerId)

	// 关键：只在 last_msg_time <= readTime 时清零，避免清掉之后到来的新消息
	if readTime > 0 {
		q = q.Where("last_msg_time <= ?", readTime)
	}

	if err := q.Update("unread_count", 0).Error; err != nil {
		return err
	}
	// DB 权威未读：Redis unread 仅用于清理历史残留 key（兼容旧逻辑/防鬼未读）
	if s.UnreadStorage != nil {
		s.UnreadStorage.Reset(ctx, int(userId), sessionType, int(peerId))
	}
	return nil
}

//func (s *SessionService) CreateSession(ctx context.Context, tx *gorm.DB, userId int, groupId uint64) (uint64, error) {
//	// 允许外部不传事务：不传就用默认 DB
//	if tx == nil {
//		tx = s.DB
//	}
//	now := time.Now()
//	session := &models.Session{
//		UserId:         uint64(userId),
//		SessionType:    2,       // 2 表示群聊（1 表示单聊）
//		PeerId:         groupId, // 群ID
//		LastMsgId:      0,
//		LastMsgType:    1,
//		LastMsgContent: "创建了群聊",
//		LastMsgTime:    now.UnixMilli(),
//		UnreadCount:    0,
//		IsTop:          0,
//		IsMute:         0,
//		CreatedAt:      now,
//		UpdatedAt:      now,
//	}
//	err := tx.WithContext(ctx).Create(session).Error
//	if err == nil {
//		return session.Id, nil
//	}
//
//	//幂等：如果是唯一键冲突，说明会话已经存在 -> 查出来返回 id
//	if isMySQLDuplicateKey(err) {
//		var exist models.Session
//		qErr := tx.WithContext(ctx).
//			Select("id").
//			Where("user_id = ? AND session_type = ? AND peer_id = ?", uint64(userId), 2, groupId).
//			First(&exist).Error
//		if qErr != nil {
//			return 0, errors.New("创建会话失败: 查询已存在会话失败")
//		}
//		return exist.Id, nil
//	}
//
//	return 0, errors.New("创建会话失败")
//}

func isMySQLDuplicateKey(err error) bool {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		// 1062 = Duplicate entry（唯一键冲突）
		return me.Number == 1062
	}
	return false
}
