package service

import (
	"Hyper/dao"
	"Hyper/models"
	"time"

	"github.com/google/uuid"
	"github.com/google/wire"
)

type MessageService struct {
	dao *dao.MessageDAO
}

func NewMessageService(dao *dao.MessageDAO) *MessageService {
	return &MessageService{dao: dao}
}

// 发送消息
func (s *MessageService) SendMessage(msg *models.Message) error {
	// 设置时间戳
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().UnixMilli()
	}
	// 初始状态 0-发送中
	if msg.Status == 0 {
		msg.Status = 0
	}
	return s.dao.Save(msg)
}

func (s *MessageService) SendGroupMessage(msg *models.Message) error {
	// 群消息仍然只存一条
	return s.dao.Save(msg)
}

// 查询某个用户/群的最近消息
func (s *MessageService) GetRecentMessages(targetID string, limit int) ([]models.Message, error) {
	return s.dao.GetMessagesByTarget(targetID, limit)
}

// PullOfflineMessages 拉取需要补发的消息
func (s *MessageService) PullOfflineMessages(userID string) ([]models.Message, error) {
	msgs, err := s.dao.GetOfflineMessages(userID, 100)
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
	return s.dao.Save(msg)
}

var MessageProviderSet = wire.NewSet(
	NewMessageService,
)

// ack
func (s *MessageService) AckMessages(msgIDs []string) error {
	return s.dao.MarkMessagesRead(msgIDs)
}

// 已读回执
func (s *MessageService) GetMessageByID(msgID string) (*models.Message, error) {
	return s.dao.GetByID(msgID)
}
