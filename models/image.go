package models

import "time"

type Image struct {
	ID        int64     `gorm:"column:id;primaryKey" json:"id"`
	UserID    int       `gorm:"column:user_id;not null;index:idx_user_status,priority:1" json:"user_id"`
	OssKey    string    `gorm:"column:oss_key;type:varchar(255);not null" json:"oss_key"`
	Width     int       `gorm:"column:width;not null" json:"width"`
	Height    int       `gorm:"column:height;not null" json:"height"`
	Status    int       `gorm:"column:status;not null;index:idx_user_status,priority:2" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at;not null;index:idx_created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
}

// TableName 显式指定表名（推荐）
func (Image) TableName() string {
	return "image"
}
