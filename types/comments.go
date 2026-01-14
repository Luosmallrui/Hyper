// types/comment.go
package types

import "time"

// 创建评论请求
type CreateCommentRequest struct {
	NoteID        uint64 `json:"note_id,string" binding:"required"`
	Content       string `json:"content" binding:"required,min=1,max=1000"`
	RootID        uint64 `json:"root_id,string"`   // 根评论ID(回复评论时需要)
	ParentID      uint64 `json:"parent_id,string"` // 父评论ID(回复评论时需要)
	ReplyToUserID int    `json:"reply_to_user_id"` // 回复的目标用户ID
}

// 删除评论请求
type DeleteCommentRequest struct {
	CommentID uint64 `json:"comment_id,string" binding:"required"`
}

// 点赞评论请求
type LikeCommentRequest struct {
	CommentID uint64 `json:"comment_id,string" binding:"required"`
}

// 取消点赞评论请求
type UnlikeCommentRequest struct {
	CommentID uint64 `json:"comment_id,string" binding:"required"`
}

// 评论响应(一级评论)
type CommentResponse struct {
	ID         uint64    `json:"id"`
	NoteID     uint64    `json:"note_id"`
	UserID     uint64    `json:"user_id"`
	Content    string    `json:"content"`
	LikeCount  int       `json:"like_count"`
	ReplyCount int       `json:"reply_count"` // 回复数
	IPLocation string    `json:"ip_location"`
	IsLiked    bool      `json:"is_liked"` // 当前用户是否点赞
	CreatedAt  time.Time `json:"created_at"`

	// 用户信息
	User UserProfile `json:"user"`

	// 最新3条回复
	LatestReplies []*ReplyResponse `json:"latest_replies,omitempty"`
}

type ReplyResponse struct {
	ID         uint64    `json:"id"`
	RootID     uint64    `json:"root_id"`
	ParentID   uint64    `json:"parent_id"`
	Content    string    `json:"content"`
	LikeCount  int       `json:"like_count"`
	IsLiked    bool      `json:"is_liked"`
	IPLocation string    `json:"ip_location"`
	CreatedAt  time.Time `json:"created_at"`

	// 评论者信息
	User UserProfile `json:"user"`

	// 回复目标用户信息(如果是回复别人的回复)
	ReplyToUser UserProfile `json:"reply_to_user,omitempty"`
}

type UserInfo struct {
	UserID   uint64 `json:"user_id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

type GetCommentsRequest struct {
	NoteID   uint64 `form:"note_id" binding:"required"`
	Cursor   int64  `form:"cursor"`    // 游标(时间戳纳秒)
	PageSize int    `form:"page_size"` // 每页数量
}

type GetRepliesRequest struct {
	RootID   uint64 `form:"root_id" binding:"required"`
	Cursor   int64  `form:"cursor"`
	PageSize int    `form:"page_size"`
}

type CommentsListResponse struct {
	Comments   []*CommentResponse `json:"comments"`
	NextCursor int64              `json:"next_cursor"` // 下一页游标
	HasMore    bool               `json:"has_more"`    // 是否还有更多
}

type RepliesListResponse struct {
	Replies    []*ReplyResponse `json:"replies"`
	NextCursor int64            `json:"next_cursor"`
	HasMore    bool             `json:"has_more"`
}
