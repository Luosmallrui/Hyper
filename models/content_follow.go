package models

import "time"

const (
	ContentFollowTargetActivity  = "activity"
	ContentFollowTargetVenue     = "venue"
	ContentFollowTargetParty     = "party"
	ContentFollowTargetOrganizer = "organizer"
)

// ContentFollow is an object-level follow relation. It is intentionally
// separate from user_follow, which represents following another user.
type ContentFollow struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int64     `gorm:"column:user_id;not null;uniqueIndex:uk_content_follow,priority:1;index" json:"user_id"`
	TargetType string    `gorm:"column:target_type;size:20;not null;uniqueIndex:uk_content_follow,priority:2;index" json:"target_type"`
	TargetID   int64     `gorm:"column:target_id;not null;uniqueIndex:uk_content_follow,priority:3;index" json:"target_id"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ContentFollow) TableName() string { return "content_follows" }
