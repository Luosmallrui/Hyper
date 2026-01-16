package types

// SearchTopicsRequest - 搜索话题请求
type SearchTopicsRequest struct {
	Query string `json:"query" binding:"max=64"` // 搜索关键词，为空则返回热门话题
}

// CreateOrGetTopicResponse 创建或获取话题响应 - 返回话题ID、名字和相关帖子数
type CreateOrGetTopicResponse struct {
	ID        uint64 `json:"id"`         // 话题 ID
	Name      string `json:"name"`       // 话题名字
	ViewCount uint32 `json:"view_count"` // 相关帖子的关联数
}

// TopicSearchResult - 单个话题搜索结果
type TopicSearchResult struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	PostCount   uint32 `json:"post_count"`
	IsHot       bool   `json:"is_hot"`
	SortWeight  int32  `json:"sort_weight"`
	Description string `json:"description"`
}

// SearchTopicsResponse - 搜索话题响应
type SearchTopicsResponse struct {
	Topics []TopicSearchResult `json:"topics"` // 匹配的话题列表
}

// ============ 话题笔记相关 ============

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

// TopicNotesResponse 话题笔记列表响应
type TopicNotesResponse struct {
	Topic      *TopicInfo       `json:"topic"`
	Notes      []*NoteWithStats `json:"notes"`
	HasMore    bool             `json:"has_more"`
	NextCursor int64            `json:"next_cursor"`
}
