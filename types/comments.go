package types

// CreateCommentRequest 创建评论请求
type CreateCommentRequest struct {
	PostID        uint64  `json:"post_id" binding:"required"` // 帖子ID（必填）
	Content       string  `json:"content" binding:"required"` // 评论内容（必填）
	RootID        uint64  `json:"root_id"`                    // 顶级评论ID（如果是回复，填这个）
	ParentID      uint64  `json:"parent_id"`                  // 直接父评论ID（如果是回复，填这个）
	ReplyToUserID *uint64 `json:"reply_to_user_id,omitempty"` // 被回复人ID（如果是回复，填这个）
}

// CommentResponse 评论响应（用于返回给前端）
type CommentResponse struct {
	ID            uint64  `json:"id"`                         // 评论ID
	PostID        uint64  `json:"post_id"`                    // 帖子ID
	UserID        uint64  `json:"user_id"`                    // 评论者ID
	Content       string  `json:"content"`                    // 评论内容
	LikeCount     uint32  `json:"like_count"`                 // 点赞数
	RootID        uint64  `json:"root_id"`                    // 顶级评论ID
	ParentID      uint64  `json:"parent_id"`                  // 父评论ID
	ReplyToUserID *uint64 `json:"reply_to_user_id,omitempty"` // 被回复人ID
	CreatedAt     string  `json:"created_at"`                 // 创建时间
}
