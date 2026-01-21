package dao

import (
	"Hyper/models"

	"gorm.io/gorm"
)

type MessageDAO struct {
	db *gorm.DB
}

func NewMessageDAO(db *gorm.DB) *MessageDAO {
	return &MessageDAO{db: db}
}

func (d *MessageDAO) Save(msg *models.ImSingleMessage) error {
	//table := "im_single_messages"
	//if msg.SessionType == 2 {
	//	table = "im_group_messages"
	//}
	return d.db.Create(msg).Error
}

// 保存单聊消息（im_single_messages）
func (d *MessageDAO) SaveSingle(msg *models.ImSingleMessage) error {
	return d.db.Create(msg).Error
}

// 保存群聊消息（ im_group_messages）
func (d *MessageDAO) SaveGroup(msg *models.ImGroupMessage) error {
	return d.db.Create(msg).Error
}
