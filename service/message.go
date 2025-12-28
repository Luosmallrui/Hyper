package service

import (
	"Hyper/dao"
	"Hyper/models"
	"time"

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

// 查询某个用户/群的最近消息
func (s *MessageService) GetRecentMessages(targetID string, limit int) ([]models.Message, error) {
	return s.dao.GetMessagesByTarget(targetID, limit)
}

var MessageProviderSet = wire.NewSet(NewMessageService)
