package dao

import (
	"time"

	"github.com/google/wire"
	"gorm.io/gorm"
)

type MessageReadDAO struct {
	db *gorm.DB
}

func NewMessageReadDAO(db *gorm.DB) *MessageReadDAO {
	return &MessageReadDAO{db: db}
}

// 插入一条“已读记录”
func (d *MessageReadDAO) MarkRead(msgID, userID string) error {
	return d.db.Exec(
		`INSERT IGNORE INTO im_message_read (msg_id, user_id, read_at)
		 VALUES (?, ?, ?)`,
		msgID,
		userID,
		time.Now().UnixMilli(),
	).Error
}

// ProviderSet（注意名字）
var MessageReadProviderSet = wire.NewSet(NewMessageReadDAO)
