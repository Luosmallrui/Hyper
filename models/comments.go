package models

import (
	"time"
)

// Comment 评论表结构
type Comments struct {
	ID            uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                                                 // 评论唯一ID
	PostID        uint64    `gorm:"column:post_id;not null;index:idx_post_id_root_id" json:"post_id"`                             // 所属帖子ID
	UserID        uint64    `gorm:"column:user_id;not null;index:idx_user_id" json:"user_id"`                                     // 发布评论的用户ID
	RootID        uint64    `gorm:"column:root_id;not null;default:0;index:idx_post_id_root_id;index:idx_root_id" json:"root_id"` // 顶级评论ID (0表示本身是顶级评论)
	ParentID      uint64    `gorm:"column:parent_id;not null;default:0" json:"parent_id"`                                         // 直接上级评论ID
	ReplyToUserID *uint64   `gorm:"column:reply_to_user_id" json:"reply_to_user_id,omitempty"`                                    // 被回复人的用户ID (使用指针处理 NULL)
	Content       string    `gorm:"column:content;type:text;not null" json:"content"`                                             // 评论正文
	LikeCount     uint32    `gorm:"column:like_count;default:0" json:"like_count"`                                                // 点赞数
	Status        int8      `gorm:"column:status;default:1" json:"status"`                                                        // 状态: 1-正常, 0-已删除, 2-审核中
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`                                           // 创建时间
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`                                           // 更新时间
}

// TableName 指定 GORM 使用的表名
func (Comments) TableName() string {
	return "comments"
}

// CommentLike 评论点赞表
type CommentLike struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	CommentID uint64    `gorm:"index:idx_comment_user,unique" json:"comment_id"`
	UserID    uint64    `gorm:"index:idx_comment_user,unique" json:"user_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (CommentLike) TableName() string {
	return "comment_likes"
}
