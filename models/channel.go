package models

import (
	"time"
)

// Channel 频道表结构体
type Channel struct {
	ID          int       `gorm:"primaryKey;column:id" json:"id"`
	Name        string    `gorm:"column:name;type:varchar(50);not null;uniqueIndex" json:"name"`
	EnName      string    `gorm:"column:en_name;type:varchar(50)" json:"en_name"`
	IconURL     string    `gorm:"column:icon_url;type:varchar(255)" json:"icon_url"`
	Description string    `gorm:"column:description;type:varchar(255)" json:"description"`
	SortWeight  int32     `gorm:"column:sort_weight;type:int;default:0;index" json:"sort_weight"`
	IsVisible   bool      `gorm:"column:is_visible;type:tinyint(1);default:1" json:"is_visible"`
	ParentID    uint32    `gorm:"column:parent_id;type:int;default:0;index" json:"parent_id"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (Channel) TableName() string {
	return "channels"
}
