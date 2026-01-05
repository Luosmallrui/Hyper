package models

import "time"

// NoteLike 点赞记录
// 对应表 note_likes
// 唯一键: note_id + user_id
// status: 1=已点赞, 0=已取消
type NoteLike struct {
	ID        uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"id"`
	NoteID    uint64    `gorm:"column:note_id;not null;index:uk_note_user,priority:1" json:"note_id"`
	UserID    int       `gorm:"column:user_id;not null;index:uk_note_user,priority:2" json:"user_id"`
	Status    uint8     `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (n NoteLike) TableName() string { return "note_likes" }
