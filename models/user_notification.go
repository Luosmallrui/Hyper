package models

import "time"

// UserNotification 用户通知收件箱：系统消息/互动通知/支付消息统一落库，
// WebSocket 只负责实时提醒，离线消息从这里补拉。
type UserNotification struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"column:user_id;not null;index:idx_user_id,priority:1;index:idx_user_read,priority:1" json:"user_id"`
	Type      string    `gorm:"column:type;size:20;not null" json:"type"` // system / interaction / payment
	Title     string    `gorm:"column:title;size:100;not null" json:"title"`
	Content   string    `gorm:"column:content;size:500;not null;default:''" json:"content"`
	Payload   string    `gorm:"column:payload;size:500;not null;default:''" json:"payload"` // JSON 字符串，跳转参数
	IsRead    int8      `gorm:"column:is_read;not null;default:0;index:idx_user_read,priority:2" json:"is_read"` // 0 未读 1 已读
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (UserNotification) TableName() string { return "user_notifications" }
