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
	Type        int       `gorm:"column:type;not null;default:1" json:"type"`
	Status      int       `gorm:"column:status;not null;default:0;index:idx_userid_status" json:"status"`
	VisibleConf int       `gorm:"column:visible_conf;not null;default:1" json:"visible_conf"`
	CreatedAt   time.Time `gorm:"column:created_at;index:idx_created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
	ActivityID  int       `gorm:"column:activity_id;not null;index:idx_activity_id" json:"activity_id"`
	StoreID     int64     `gorm:"column:store_id;not null;default:0;index:idx_store_id" json:"store_id"`
}

func (n Note) TableName() string {
	return "notes"
}

// NoteShare stores every share action. Unlike note_stats, it is an auditable event log.
type NoteShare struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	NoteID    uint64    `gorm:"column:note_id;not null;index:idx_note_share_note_created" json:"note_id"`
	UserID    uint64    `gorm:"column:user_id;not null;index:idx_note_share_user" json:"user_id"`
	Channel   string    `gorm:"column:channel;size:50;not null;default:''" json:"channel"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;index:idx_note_share_note_created" json:"created_at"`
}

func (NoteShare) TableName() string {
	return "note_shares"
}
