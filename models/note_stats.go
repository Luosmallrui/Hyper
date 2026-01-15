package models

import "time"

// NoteStats 笔记统计
// 对应表 note_stats
type NoteStats struct {
	NoteID       uint64    `gorm:"column:note_id;primaryKey" json:"note_id"`
	LikeCount    int64     `gorm:"column:like_count;default:0" json:"like_count"`
	CollCount    int64     `gorm:"column:coll_count;default:0" json:"coll_count"`
	ShareCount   int64     `gorm:"column:share_count;default:0" json:"share_count"`
	CommentCount int64     `gorm:"column:comment_count;default:0" json:"comment_count"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (NoteStats) TableName() string {
	return "note_stats"
}
