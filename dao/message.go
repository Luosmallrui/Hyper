package dao

import (
	"Hyper/models"

	"github.com/google/wire"
	"gorm.io/gorm"
)

type MessageDAO struct {
	db *gorm.DB
}

func NewMessageDAO(db *gorm.DB) *MessageDAO {
	return &MessageDAO{db: db}
}

// 保存消息
func (d *MessageDAO) Save(msg *models.Message) error {
	return d.db.Create(msg).Error
}

// 查询某个用户/群的最近消息
func (d *MessageDAO) GetMessagesByTarget(targetID string, limit int) ([]models.Message, error) {
	var msgs []models.Message
	err := d.db.Where("target_id = ?", targetID).Order("timestamp desc").Limit(limit).Find(&msgs).Error
	return msgs, err
}

// ProviderSet 单独命名，避免与其他 ProviderSet 冲突
var MessageProviderSet = wire.NewSet(NewMessageDAO)
