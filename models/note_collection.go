package models

import "time"

// NoteCollection 收藏记录，对应 note_collections
// status: 1=已收藏，0=已取消
// 唯一键: note_id + user_id
type NoteCollection struct {
	ID        uint64    `gorm:"column:id;primaryKey;AUTO_INCREMENT" json:"id"`
	NoteID    uint64    `gorm:"column:note_id;not null;index:uk_note_user,priority:1" json:"note_id"`
	UserID    int       `gorm:"column:user_id;not null;index:uk_note_user,priority:2" json:"user_id"`
	Status    uint8     `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (n NoteCollection) TableName() string { return "note_collections" }
