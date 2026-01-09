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

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type MessageService struct {
	MessageDao *dao.MessageDAO
	MqProducer rocketmq.Producer
	Redis      *redis.Client
	DB         *gorm.DB
}

var _ IMessageService = (*MessageService)(nil)

type IMessageService interface {
	SaveSingleMessage(msg *models.ImSingleMessage) error
	SaveGroupMessage(msg *models.ImGroupMessage) error
	SendMessage(msg *types.Message) error
	GetRecentMessages(targetID string, limit int) ([]models.Message, error)
	PullOfflineMessages(userID string) ([]models.Message, error)
	SendSystemMessage(targetID string, content string) error
	AckMessages(msgIDs []string) error
	GetMessageByID(msgID string) (*models.Message, error)
	ListMessages(ctx context.Context, userId, peerId uint64, cursor int64, limit int) ([]types.ListMessageReq, error)
}

func (s *MessageService) ListMessages(ctx context.Context, userId, peerId uint64, cursor int64, limit int) ([]types.ListMessageReq, error) {

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	sessionHash := GetSessionHash(int64(userId), int64(peerId))

	q := s.DB.WithContext(ctx).
		Model(&models.ImSingleMessage{}).
		Where("session_hash = ?", sessionHash)
	if cursor > 0 {
		q = q.Where("created_at < ?", cursor)
	}

	var msgs []models.ImSingleMessage
	if err := q.
		Order("created_at DESC").
		Limit(limit).
		Find(&msgs).Error; err != nil {
		return nil, err
	}

	result := make([]types.ListMessageReq, 0, len(msgs))

	for i := len(msgs) - 1; i >= 0; i-- {
		m := msgs[i]

		ext := make(map[string]interface{})
		if m.Ext != "" {
			if err := json.Unmarshal([]byte(m.Ext), &ext); err != nil {
				ext = make(map[string]interface{})
			}
		}
		if ext == nil {
			ext = make(map[string]interface{})
		}

		item := types.ListMessageReq{
			Id:       uint64(m.Id),
			SenderId: uint64(m.SenderId),
			Content:  m.Content,
			MsgType:  m.MsgType,
			Ext:      ext,
			Time:     m.CreatedAt,
			IsSelf:   false,
		}
		if m.SenderId == int64(userId) {
			item.IsSelf = true
		}

		result = append(result, item)
	}

	return result, nil
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

	// 3) 频道（给 ws / 路由用）
	msg.Channel = types.ChannelChat

	// 4) 发 MQ
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	mqMsg := &primitive.Message{
		Topic: types.ImTopicChat,
		Body:  body,
	}

	// 5) ShardingKey：保证“同一会话/同一群”的消息尽量落在同一分片上，顺序更稳定
	switch msg.SessionType {
	case types.SessionTypeSingle:
		// 单聊按目标用户分片
		mqMsg.WithShardingKey(fmt.Sprintf("%d", msg.TargetID))
	case types.GroupChatSessionTypeGroup:
		// 群聊按群分片（加 g_ 前缀避免和用户ID字符串混淆）
		mqMsg.WithShardingKey(fmt.Sprintf("g_%d", msg.TargetID))
	}

	_, err = s.MqProducer.SendSync(context.Background(), mqMsg)
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
	_, _ = h.Write([]byte(rawID))

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
