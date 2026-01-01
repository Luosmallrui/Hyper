package service

import (
	"Hyper/dao"
	"Hyper/models"
	"time"

	"github.com/google/uuid"
)

type MessageService struct {
	MessageDao *dao.MessageDAO
}

var _ IMessageService = (*MessageService)(nil)

type IMessageService interface {
	SendMessage(msg *models.Message) error
	SendGroupMessage(msg *models.Message) error
	GetRecentMessages(targetID string, limit int) ([]models.Message, error)
	PullOfflineMessages(userID string) ([]models.Message, error)
	SendSystemMessage(targetID string, content string) error
	AckMessages(msgIDs []string) error
	GetMessageByID(msgID string) (*models.Message, error)
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
	return s.MessageDao.Save(msg)
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
