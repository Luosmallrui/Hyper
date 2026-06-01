package models

import "time"

type AdminWechatSubscriber struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	AdminID   int64     `gorm:"column:admin_id;not null;uniqueIndex:uk_admin_openid" json:"admin_id"`
	OpenID    string    `gorm:"column:open_id;size:64;not null;uniqueIndex:uk_admin_openid" json:"open_id"`
	Enabled   int8      `gorm:"column:enabled;not null;default:1" json:"enabled"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (AdminWechatSubscriber) TableName() string {
	return "admin_wechat_subscribers"
}
