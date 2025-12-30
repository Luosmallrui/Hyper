package service

import "Hyper/dao"

type MessageReadService struct {
	dao *dao.MessageReadDAO
}

func NewMessageReadService(d *dao.MessageReadDAO) *MessageReadService {
	return &MessageReadService{dao: d}
}

func (s *MessageReadService) MarkGroupRead(msgID, userID string) error {
	return s.dao.MarkRead(msgID, userID)
}
