package types

type CreateTopicRequest struct {
	// Name 是核心，不能为空，限制长度
	Name string `json:"name" binding:"required,min=1,max=64"`

	// Description 和 CoverURL 是可选的，给默认值或允许为空
	Description string `json:"description" binding:"max=255"`
	CoverURL    string `json:"cover_url" binding:"omitempty,url"`

	// CategoryID 用于将话题归类
	CategoryID uint32 `json:"category_id"`
}

type CreateTopicResponse struct {
	ID uint64 `json:"id"` // 返回新创建的话题 ID
}

// TopicNotesResponse 话题笔记列表响应
type TopicNotesResponse struct {
	Topic      *TopicInfo       `json:"topic"`
	Notes      []*NoteWithStats `json:"notes"`
	HasMore    bool             `json:"has_more"`
	NextCursor int64            `json:"next_cursor"`
}

// TopicInfo 话题信息
type TopicInfo struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	PostCount   uint32 `json:"post_count"`
	ViewCount   uint32 `json:"view_count"`
	FollowCount uint32 `json:"follow_count"`
	IsHot       bool   `json:"is_hot"`
}

// NoteWithStats 带统计的笔记
type NoteWithStats struct {
	ID           int64  `json:"id"`
	UserID       int64  `json:"user_id"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	Type         int    `json:"type"`
	Nickname     string `json:"nickname"`
	Avatar       string `json:"avatar"`
	LikeCount    int64  `json:"like_count"`
	CommentCount int64  `json:"comment_count"`
	ViewCount    int64  `json:"view_count"`
	CreatedAt    int64  `json:"created_at"`
}
