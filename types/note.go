package types

import "time"

// Note 笔记主表：存储核心文字和状态
type Note struct {
	ID       int64   `gorm:"primaryKey" json:"id"`           // 雪花算法ID
	UserID   int64   `gorm:"index" json:"user_id"`           // 作者ID
	Title    string  `gorm:"type:varchar(100)" json:"title"` // 标题
	Content  string  `gorm:"type:text" json:"content"`       // 正文内容
	TopicIDs []int64 `gorm:"type:json" json:"topic_ids"`     // 话题列表
	Location string  `gorm:"type:json" json:"location"`      // 地理位置{lat, lng, name}

	MediaData string `gorm:"type:json" json:"media_data"`

	Type        int `json:"type"`         // 1-图文, 2-视频
	Status      int `json:"status"`       // 0-审核中, 1-公开, 2-私密, 3-违规
	VisibleConf int `json:"visible_conf"` // 1-公开, 2-粉丝可见, 3-自己可见

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NoteMedia 媒体资源明细
type NoteMedia struct {
	URL          string `json:"url"`           // 主图/视频地址
	ThumbnailURL string `json:"thumbnail_url"` // 缩略图
	Width        int    `json:"width"`         // 宽高比，前端排版布局用
	Height       int    `json:"height"`
	Duration     int    `json:"duration"` // 视频时长(秒)
}

// NoteStat
type NoteStat struct {
	NoteID     int64 `gorm:"primaryKey"`
	LikeCount  int64 `json:"like_count"`    // 点赞数
	CollCount  int64 `json:"coll_count"`    // 收藏数
	ShareCount int64 `json:"share_count"`   // 分享数
	CommentCnt int64 `json:"comment_count"` // 评论总数
}

type UploadResponse struct {
	Key    string `json:"key"`    // OSS 路径
	Width  int    `json:"width"`  // 原始宽度
	Height int    `json:"height"` // 原始高度
}
