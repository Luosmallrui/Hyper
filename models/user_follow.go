package models

import (
	"time"
)

type UserFollow struct {
	ID         uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"id"`
	FollowerID uint64    `gorm:"column:follower_id;not null" json:"follower_id"` // 关注人
	FolloweeID uint64    `gorm:"column:followee_id;not null" json:"followee_id"` // 被关注人
	Status     int       `gorm:"column:status;not null;default:1" json:"status"` // 1:关注中 0:已取消
	CreatedAt  time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
}

func (UserFollow) TableName() string {
	return "user_follow"
}
