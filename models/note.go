package models

import (
	"time"
)

type Note struct {
	ID          uint64    `gorm:"column:id;primary_key" json:"id"`
	UserID      uint64    `gorm:"column:user_id;not null;index:idx_userid_status" json:"user_id"`
	Title       string    `gorm:"column:title;type:varchar(100);not null;default:''" json:"title"`
	Content     string    `gorm:"column:content;type:text" json:"content"`
	TopicIDs    string    `gorm:"column:topic_ids;type:json" json:"topic_ids"`
	Location    string    `gorm:"column:location;type:json" json:"location"`
	MediaData   string    `gorm:"column:media_data;type:json" json:"media_data"`
	Type        int8      `gorm:"column:type;not null;default:1" json:"type"`
	Status      int8      `gorm:"column:status;not null;default:0;index:idx_userid_status" json:"status"`
	VisibleConf int8      `gorm:"column:visible_conf;not null;default:1" json:"visible_conf"`
	CreatedAt   time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (n Note) TableName() string {
	return "notes"
}
