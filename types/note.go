package types

import (
	"time"
)

// VisibleConf 笔记可见性常量
const (
	VisibleConfPublic        int = 1 // 公开
	VisibleConfFollowersOnly int = 2 // 粉丝可见
	VisibleConfPrivate       int = 3 // 自己可见
)

// Pagination 分页常量
const (
	DefaultPage     int = 1  // 默认页码
	DefaultPageSize int = 20 // 默认每页数量
)

// NoteStatus 笔记状态常量
const (
	NoteStatusDefaultQuery int = 1 // 查询笔记列表时的默认状态（公开）
)

// Note 笔记主表：存储核心文字和状态
type Note struct {
	ID       int64    `gorm:"primaryKey" json:"id"`           // 雪花算法ID
	UserID   int64    `gorm:"index" json:"user_id"`           // 作者ID
	Title    string   `gorm:"type:varchar(100)" json:"title"` // 标题
	Content  string   `gorm:"type:text" json:"content"`       // 正文内容
	TopicIDs []int64  `gorm:"type:json" json:"topic_ids"`     // 话题列表
	Location Location `gorm:"type:json" json:"location"`      // 地理位置{lat, lng, name}

	MediaData []NoteMedia `gorm:"type:json" json:"media_data"`

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
	Url    string `json:"url"`    // OSS 路径
	Width  int    `json:"width"`  // 原始宽度
	Height int    `json:"height"` // 原始高度
}

// CreateNoteRequest 创建笔记请求
type CreateNoteRequest struct {
	Title       string      `json:"title" binding:"required,max=100"`   // 标题
	Content     string      `json:"content"`                            // 正文内容
	TopicIDs    []int64     `json:"topic_ids"`                          // 话题列表
	Location    *Location   `json:"location"`                           // 地理位置
	MediaData   []NoteMedia `json:"media_data"`                         // 媒体资源列表
	Type        int         `json:"type" binding:"required,oneof=1 2"`  // 1-图文, 2-视频
	VisibleConf int         `json:"visible_conf" binding:"oneof=1 2 3"` // 1-公开, 2-粉丝可见, 3-自己可见
}

// Location 地理位置
type Location struct {
	Lat  float64 `json:"lat"`  // 纬度
	Lng  float64 `json:"lng"`  // 经度
	Name string  `json:"name"` // 地点名称
}

// CreateNoteResponse 创建笔记响应
type CreateNoteResponse struct {
	NoteID uint64 `json:"note_id"` // 笔记ID
}

// GetMyNotesRequest 查询自己笔记的请求
type GetMyNotesRequest struct {
	Status   int `form:"status" binding:"omitempty,oneof=0 1 2 3"`   // 笔记状态筛选（可选）
	Page     int `form:"page" binding:"omitempty,min=1"`             // 页码（从1开始）
	PageSize int `form:"pagesize" binding:"omitempty,min=1,max=100"` // 每页数量
}

// GetMyNotesResponse 笔记列表响应
type GetMyNotesResponse struct {
	Notes []*Note `json:"notes"` // 笔记列表
	Total int     `json:"total"` // 总数
}

// GetMyCollectionsRequest 查询自己收藏的笔记请求
type GetMyCollectionsRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"pagesize" binding:"omitempty,min=1,max=100"`
}

type GetMyCollectionsResponse struct {
	Notes []*Note `json:"notes"`
	Total int     `json:"total"`
}

type Leaf struct {
	Page     int `form:"page"`      // 页码（从1开始）
	PageSize int `form:"pageSize" ` // 每页数量
}

type Notes struct {
	ID        int64     `json:"id"`        // 雪花算法ID
	UserID    int64     `json:"user_id"`   // 作者ID
	Title     string    `json:"title"`     // 标题
	Content   string    ` json:"content"`  // 正文内容
	TopicIDs  []int64   `json:"topic_ids"` // 话题列表
	Location  Location  `json:"location"`  // 地理位置{lat, lng, name}
	MediaData NoteMedia `json:"media_data"`

	Type        int       `json:"type"`         // 1-图文, 2-视频
	Status      int       `json:"status"`       // 0-审核中, 1-公开, 2-私密, 3-违规
	VisibleConf int       `json:"visible_conf"` // 1-公开, 2-粉丝可见, 3-自己可见
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	TimeStamp   int64     `json:"time_stamp"`
	Avatar      string    `json:"avatar"`
	Nickname    string    `json:"nickname"`

	// 统计数据
	LikeCount    int64 `json:"like_count"`
	CollCount    int64 `json:"coll_count"`
	ShareCount   int64 `json:"share_count"`
	CommentCount int64 `json:"comment_count"`
	ViewCount    int64 `json:"view_count,omitempty"`

	// 用户相关状态
	IsLiked     bool `json:"is_liked"`     // 当前用户是否点赞
	IsCollected bool `json:"is_collected"` // 当前用户是否收藏
	IsFollowed  bool `json:"is_followed"`  // 当前用户是否关注作者
}

type NoteBrief struct {
	ID        int64     `json:"id,string"` // 雪花算法ID
	UserID    int64     `json:"user_id"`   // 作者ID
	Title     string    `json:"title"`     // 标题
	Type      int       `json:"type"`      // 1-图文, 2-视频
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	TimeStamp int64     `json:"time_stamp"`
	MediaData NoteMedia `json:"media_data"`
	// 统计数据
	LikeCount    int64 `json:"like_count"`
	CollCount    int64 `json:"coll_count"`
	ShareCount   int64 `json:"share_count"`
	CommentCount int64 `json:"comment_count"`
	ViewCount    int64 `json:"view_count,omitempty"`

	// 用户相关状态
	IsLiked     bool `json:"is_liked"`     // 当前用户是否点赞
	IsCollected bool `json:"is_collected"` // 当前用户是否收藏
	IsFollowed  bool `json:"is_followed"`  // 当前用户是否关注作者
}

type ListNotesReq struct {
	Cursor     int64  `form:"cursor"` // 传入上次最后一条记录的时间戳（纳秒或秒）
	PageSize   int    `form:"pageSize"`
	SearchType string `form:"search_type"`
	ChannelID  int64  `form:"channel_id"`
}
type ListNotesRep struct {
	Notes      []*Notes `json:"notes"`
	NextCursor int64    `json:"next_cursor"` // 返回给前端，下次请求带上
	HasMore    bool     `json:"has_more"`    // 告诉前端是否还有更多
}

type NoteStats struct {
	NoteID       uint64 `json:"note_id"`
	LikeCount    int64  `json:"like_count"`
	CollCount    int64  `json:"coll_count"`
	ShareCount   int64  `json:"share_count"`
	CommentCount int64  `json:"comment_count"`
	ViewCount    int64  `json:"view_count,omitempty"` // 如果后续加浏览数
}

type NoteDetail struct {
	ID          int64       `json:"id"`
	UserID      int64       `json:"user_id"`
	Title       string      `json:"title"`
	Content     string      `json:"content"`
	TopicIDs    []int64     `json:"topic_ids"`
	Location    Location    `json:"location"`
	MediaData   []NoteMedia `json:"media_data"`
	Type        int         `json:"type"`
	Status      int         `json:"status"`
	VisibleConf int         `json:"visible_conf"`

	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`

	// 统计数据
	LikeCount    int64 `json:"like_count"`
	CollCount    int64 `json:"coll_count"`
	ShareCount   int64 `json:"share_count"`
	CommentCount int64 `json:"comment_count"`
	ViewCount    int64 `json:"view_count"`

	// 用户交互状态
	IsLiked        bool            `json:"is_liked"`
	IsCollected    bool            `json:"is_collected"`
	IsFollowed     bool            `json:"is_followed"` // 是否关注了作者
	CommentPreview *CommentPreview `json:"comment_preview"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type CommentPreview struct {
	TotalCount int64              `json:"total_count"` // 评论总数
	Comments   []*CommentResponse `json:"comments"`    // 前3条评论
}

type ListNotesBriefRep struct {
	Notes      []*NoteBrief `json:"notes"`
	Nickname   string       `json:"nickname"`
	Avatar     string       `json:"avatar"`
	NextCursor int64        `json:"next_cursor"` // 返回给前端，下次请求带上
	HasMore    bool         `json:"has_more"`    // 告诉前端是否还有更多
}
