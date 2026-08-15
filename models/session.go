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

	UnreadCount   uint32
	IsTop         int
	IsMute        int
	ClearedAt     int64 // 当前用户清空该会话历史的时间戳（毫秒）
	LastMsgHidden int   // 当前用户删除最新消息或清空会话时忽略共享消息摘要缓存

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Session) TableName() string {
	return "im_session"
}
