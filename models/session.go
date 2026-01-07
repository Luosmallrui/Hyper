package models

import "time"

type Session struct {
	Id          uint64 `gorm:"primaryKey"`
	UserId      uint64
	SessionType int
	PeerId      uint64

	LastMsgId      uint64
	LastMsgType    int
	LastMsgContent string
	LastMsgTime    int64

	UnreadCount uint32
	IsTop       int
	IsMute      int

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Session) TableName() string {
	return "im_session"
}
