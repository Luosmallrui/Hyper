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

// GetOfflineMessages 查询需要补发的消息（发送成功但未读）
func (d *MessageDAO) GetOfflineMessages(userID string, limit int) ([]models.Message, error) {
	var msgs []models.Message
	err := d.db.
		Where("target_id = ? AND status = ?", userID, 1).
		Order("timestamp asc").
		Limit(limit).
		Find(&msgs).Error
	return msgs, err
}

// ProviderSet 单独命名，避免与其他 ProviderSet 冲突
var MessageProviderSet = wire.NewSet(
	NewMessageDAO,
	NewGroupDAO,
)

// ack
func (d *MessageDAO) MarkMessagesRead(msgIDs []string) error {
	return d.db.
		Model(&models.Message{}).
		Where("msg_id IN ?", msgIDs).
		Update("status", 2).Error
}

// 已读回执
func (d *MessageDAO) GetByID(msgID string) (*models.Message, error) {
	var msg models.Message
	err := d.db.Where("msg_id = ?", msgID).First(&msg).Error
	return &msg, err
}
