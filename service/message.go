package service

import (
	"Hyper/dao"
	"Hyper/models"
	"Hyper/pkg/rocketmq"
	"Hyper/pkg/snowflake"
	"Hyper/types"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type MessageService struct {
	MessageDao *dao.MessageDAO
	RocketMQ   *rocketmq.Rocketmq
	Redis      *redis.Client
}

var _ IMessageService = (*MessageService)(nil)

type IMessageService interface {
	SendMessage(msg *types.Message) error
	SendGroupMessage(msg *models.Message) error
	GetRecentMessages(targetID string, limit int) ([]models.Message, error)
	PullOfflineMessages(userID string) ([]models.Message, error)
	SendSystemMessage(targetID string, content string) error
	AckMessages(msgIDs []string) error
	GetMessageByID(msgID string) (*models.Message, error)
}

func (s *MessageService) SendMessage(msg *types.Message) error {
	msg.Timestamp = time.Now().UnixMilli()
	msg.Status = 0 // 发送中

	cacheKey := fmt.Sprintf("idempotent:%d:%s", msg.SenderID, msg.ClientMsgID)
	isNew, err := s.Redis.SetNX(context.Background(), cacheKey, "1", 24*time.Hour).Result()
	if err != nil {
		return err
	}
	if !isNew {

		return nil
	}

	msg.Id = snowflake.GenID()

	if msg.SessionType == 1 { // 假设 1 是单聊
		msg.Ext = s.generateSessionID(msg.SenderID, msg.TargetID)

	}

	body, _ := json.Marshal(msg)
	mqMsg := &primitive.Message{
		Topic: "IM_CHAT_MSGS",
		Body:  body,
	}
	mqMsg.WithShardingKey(fmt.Sprintf("%d", msg.TargetID))
	_, err = s.RocketMQ.RocketmqProducer.SendSync(context.Background(), mqMsg)
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		s.Redis.Del(context.Background(), cacheKey)
		return err
	}

	return nil
}

func (s *MessageService) SendGroupMessage(msg *models.Message) error {
	// 群消息仍然只存一条
	return s.MessageDao.Save(msg)
}

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
	msg := &models.Message{
		MsgID:       uuid.NewString(),
		SenderID:    "system",
		TargetID:    targetID,
		SessionType: 3,
		MsgType:     1,
		Content:     content,
		Timestamp:   time.Now().UnixMilli(),
		Status:      1,
		Ext:         "{}",
	}
	return s.MessageDao.Save(msg)
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
