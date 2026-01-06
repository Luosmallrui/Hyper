package models

import (
	"time"
)

type UserStats struct {
	ID             uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"id"`
	UserID         uint64    `gorm:"column:user_id;not null;uniqueIndex" json:"user_id"`
	FollowerCount  uint32    `gorm:"column:follower_count;not null;default:0" json:"follower_count"`
	FollowingCount uint32    `gorm:"column:following_count;not null;default:0" json:"following_count"`
	LikeCount      uint32    `gorm:"column:like_count;not null;default:0" json:"like_count"`
	CreatedAt      time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
}

func (UserStats) TableName() string {
	return "user_stats"
}
