package service

import (
	"Hyper/dao"
	"Hyper/models"
	"Hyper/pkg/snowflake"
	"Hyper/types"
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"time"

	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type MessageService struct {
	MessageDao     *dao.MessageDAO
	GroupMemberDAO *dao.GroupMember
	GroupDAO       *dao.GroupDAO
	MqProducer     rmq_client.Producer
	Redis          *redis.Client
	DB             *gorm.DB
}

var _ IMessageService = (*MessageService)(nil)

type IMessageService interface {
	SaveMessage(msg *models.ImSingleMessage) error
	SaveSingleMessage(msg *models.ImSingleMessage) error
	SaveGroupMessage(msg *models.ImGroupMessage) error
	SendMessage(msg *types.Message) error
	GetRecentMessages(targetID string, limit int) ([]models.Message, error)
	PullOfflineMessages(userID string) ([]models.Message, error)
	SendSystemMessage(targetID string, content string) error
	AckMessages(msgIDs []string) error
	GetMessageByID(msgID string) (*models.Message, error)
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
			ext := map[string]interface{}{}
			if m.Ext != "" {
				_ = json.Unmarshal([]byte(m.Ext), &ext)
			}
			result = append(result, types.ListMessageReq{
				Id:       uint64(m.Id),
				SenderId: uint64(m.SenderId),
				Content:  m.Content,
				MsgType:  m.MsgType,
				Ext:      ext,
				Time:     m.CreatedAt,
				IsSelf:   m.SenderId == int64(userId),
			})
		}
		return result, nil
	case types.GroupChatSessionTypeGroup:
		// 群聊：peerId 在这里就是 groupId
		// conn-server 写库时就是按 GetGroupSessionHash(groupId) 填的 SessionHash
		sessionHash := GetGroupSessionHash(int64(peerId))

		q := s.DB.WithContext(ctx).
			Model(&models.ImGroupMessage{}).
			Where("session_hash = ?", sessionHash)

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
		for _, m := range msgs {
			ext := map[string]interface{}{}
			if m.Ext != "" {
				_ = json.Unmarshal([]byte(m.Ext), &ext)
			}
			result = append(result, types.ListMessageReq{
				Id:       uint64(m.Id),
				SenderId: uint64(m.SenderId),
				Content:  m.Content,
				MsgType:  m.MsgType,
				Ext:      ext,
				Time:     m.CreatedAt,
				IsSelf:   m.SenderId == int64(userId),
			})
		}
		return result, nil

	default:
		return nil, fmt.Errorf("invalid session_type=%d (only 1 or 2)", sessionType)
	}
}

func (s *MessageService) SaveSingleMessage(msg *models.ImSingleMessage) error {
	return s.MessageDao.SaveSingle(msg)
}
func (s *MessageService) SaveGroupMessage(msg *models.ImGroupMessage) error {
	return s.MessageDao.SaveGroup(msg)
}
func (s *MessageService) SendMessage(msg *types.Message) error {
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
	// 3.5) 群聊禁言校验：必须在发 MQ 之前做
	if msg.SessionType == types.GroupChatSessionTypeGroup {

		gid := int(msg.TargetID) // 群ID
		uid := int(msg.SenderID) // 发送者ID

		// 1) 查群成员记录：是否成员/是否退群/角色/个人禁言
		m, err := s.GroupMemberDAO.FindByUserId(context.Background(), gid, uid)
		if err != nil || m.IsQuit == 1 {
			return fmt.Errorf("你不在群内或已退群")
		}

		// 2) 个人禁言：优先级最高
		if m.IsMute == 1 {
			return fmt.Errorf("你已被禁言")
		}

		// 3) 群全员禁言：只禁普通成员(role=3)，群主/管理员仍可发言
		g, err := s.GroupDAO.FindByID(context.Background(), gid)
		if err != nil {
			return fmt.Errorf("群不存在")
		}
		if g.IsMuteAll == 1 && m.Role == 3 {
			return fmt.Errorf("群已开启全员禁言")
		}
	}

	// 3) 频道（给 ws / 路由用）
	msg.Channel = types.ChannelChat

	// 4) 发 MQ
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	mqMsg := &rmq_client.Message{
		Topic: types.ImTopicChat,
		Body:  body,
	}
	//fmt.Println(s.MqProducer, 44)
	//s.MqProducer.SendAsync(context.Background(), mqMsg, func(ctx context.Context, resp []*rmq_client.SendReceipt, err error) {
	//	for i := 0; i < len(resp); i++ {
	//		fmt.Printf("%#v\n", resp[i])
	//	}
	//})
	//fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
	//MQ重发了，前者异步发送了后面又同步发送了
	_, err = s.MqProducer.Send(context.Background(), mqMsg)
	if err != nil {
		return err
	}

	return nil
}

//func (s *MessageService) SendGroupMessage(msg *models.Message) error {
//	// 群消息仍然只存一条
//	return nil
//}

// 查询某个用户/群的最近消息
func (s *MessageService) GetRecentMessages(targetID string, limit int) ([]models.Message, error) {
	return s.MessageDao.GetMessagesByTarget(targetID, limit)
}

// PullOfflineMessages 拉取需要补发的消息
func (s *MessageService) PullOfflineMessages(userID string) ([]models.Message, error) {
	msgs, err := s.MessageDao.GetOfflineMessages(userID, 100)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

// 发送系统消息
func (s *MessageService) SendSystemMessage(targetID string, content string) error {
	//msg := &models.Message{
	//	MsgID:       uuid.NewString(),
	//	SenderID:    "system",
	//	TargetID:    targetID,
	//	SessionType: 3,
	//	MsgType:     1,
	//	Content:     content,
	//	Timestamp:   time.Now().UnixMilli(),
	//	Status:      1,
	//	Ext:         "{}",
	//}
	return nil
}

// ack
func (s *MessageService) AckMessages(msgIDs []string) error {
	return s.MessageDao.MarkMessagesRead(msgIDs)
}

// 已读回执
func (s *MessageService) GetMessageByID(msgID string) (*models.Message, error) {
	return s.MessageDao.GetByID(msgID)
}

func (s *MessageService) generateSessionID(uid1, uid2 int64) string {
	if uid1 < uid2 {
		return fmt.Sprintf("%d_%d", uid1, uid2)
	}
	return fmt.Sprintf("%d_%d", uid2, uid1)
}

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
