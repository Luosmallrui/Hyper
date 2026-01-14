package dao

import (
	"Hyper/models"
	"golang.org/x/net/context"
	"gorm.io/gorm"
)

type CommentLike struct {
	Repo[models.CommentLike]
}

func NewCommentLike(db *gorm.DB) *CommentLike {
	return &CommentLike{
		Repo: NewRepo[models.CommentLike](db),
	}
}

// Create 创建点赞记录
func (d *CommentLike) Create(ctx context.Context, like *models.CommentLike) error {
	return d.Db.WithContext(ctx).Create(like).Error
}

// Delete 删除点赞记录
func (d *CommentLike) Delete(ctx context.Context, commentID, userID uint64) error {
	return d.Db.WithContext(ctx).
		Where("comment_id = ? AND user_id = ?", commentID, userID).
		Delete(&models.CommentLike{}).Error
}

// CheckExists 检查是否点赞
func (d *CommentLike) CheckExists(ctx context.Context, commentID, userID uint64) (bool, error) {
	var count int64
	err := d.Db.WithContext(ctx).
		Model(&models.CommentLike{}).
		Where("comment_id = ? AND user_id = ?", commentID, userID).
		Count(&count).Error
	return count > 0, err
}

// BatchCheckExists 批量检查点赞状态
func (d *CommentLike) BatchCheckExists(ctx context.Context, commentIDs []uint64, userID uint64) (map[uint64]bool, error) {
	result := make(map[uint64]bool)
	if len(commentIDs) == 0 {
		return result, nil
	}

	var likes []*models.CommentLike
	err := d.Db.WithContext(ctx).
		Where("comment_id IN ? AND user_id = ?", commentIDs, userID).
		Find(&likes).Error

	if err != nil {
		return nil, err
	}

	for _, like := range likes {
		result[like.CommentID] = true
	}

	return result, nil
}
