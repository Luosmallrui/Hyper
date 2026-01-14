package models

import "time"

type Comment struct {
	ID            uint64    `gorm:"column:id;primaryKey"`
	NoteID        uint64    `gorm:"column:note_id;not null"`
	UserID        uint64    `gorm:"column:user_id;not null"`
	RootID        uint64    `gorm:"column:root_id;default:0"`
	ParentID      uint64    `gorm:"column:parent_id;default:0"`
	ReplyToUserID uint64    `gorm:"column:reply_to_user_id;default:0"`
	Content       string    `gorm:"column:content;type:text;not null"`
	LikeCount     int       `gorm:"column:like_count;default:0"`
	ReplyCount    int       `gorm:"column:reply_count;default:0"`
	IPLocation    string    `gorm:"column:ip_location;size:50"`
	Status        int8      `gorm:"column:status;default:1"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (Comment) TableName() string {
	return "comments"
}

type CommentLike struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	CommentID uint64    `gorm:"column:comment_id;not null"`
	UserID    uint64    `gorm:"column:user_id;not null"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (CommentLike) TableName() string {
	return "comment_likes"
}
