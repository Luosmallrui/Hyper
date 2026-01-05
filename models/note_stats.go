package models

import "time"

// NoteStats 笔记统计
// 对应表 note_stats
type NoteStats struct {
	NoteID       uint64    `gorm:"column:note_id;primary_key" json:"note_id"`
	LikeCount    uint64    `gorm:"column:like_count;not null;default:0" json:"like_count"`
	CollCount    uint64    `gorm:"column:coll_count;not null;default:0" json:"coll_count"`
	ShareCount   uint64    `gorm:"column:share_count;not null;default:0" json:"share_count"`
	CommentCount uint64    `gorm:"column:comment_count;not null;default:0" json:"comment_count"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (n NoteStats) TableName() string { return "note_stats" }
