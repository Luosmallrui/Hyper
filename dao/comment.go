package dao

import (
	"Hyper/models"
	"context"
	"gorm.io/gorm"
	"time"
)

type Comment struct {
	Repo[models.Comment]
}

func NewComment(db *gorm.DB) *Comment {
	return &Comment{
		Repo: NewRepo[models.Comment](db),
	}
}

// GetRootCommentsByCursor 使用游标获取一级评论
func (d *Comment) GetRootCommentsByCursor(ctx context.Context, noteID uint64, cursor int64, limit int) ([]*models.Comment, error) {
	var comments []*models.Comment
	query := d.Db.WithContext(ctx).
		Where("note_id = ? AND root_id = 0 AND status = 1", noteID)

	// 如果有游标,则查询游标之前的数据
	if cursor > 0 {
		cursorTime := time.Unix(0, cursor)
		query = query.Where("created_at < ?", cursorTime)
	}

	err := query.
		Order("created_at DESC").
		Limit(limit).
		Find(&comments).Error

	return comments, err
}

// GetRepliesByCursor 使用游标获取回复(按时间正序)
func (d *Comment) GetRepliesByCursor(ctx context.Context, rootID uint64, cursor int64, limit int) ([]*models.Comment, error) {
	var replies []*models.Comment
	query := d.Db.WithContext(ctx).
		Where("root_id = ? AND status = 1", rootID)

	// 如果有游标,则查询游标之后的数据(因为回复是正序)
	if cursor > 0 {
		cursorTime := time.Unix(0, cursor)
		query = query.Where("created_at > ?", cursorTime)
	}

	err := query.
		Order("created_at ASC"). // 回复按时间正序
		Limit(limit).
		Find(&replies).Error

	return replies, err
}

func (d *Comment) Create(ctx context.Context, comment *models.Comment) error {
	return d.Db.WithContext(ctx).Create(comment).Error
}

// GetByID 根据ID获取评论
func (d *Comment) GetByID(ctx context.Context, commentID uint64) (*models.Comment, error) {
	var comment models.Comment
	err := d.Db.WithContext(ctx).
		Where("id = ? AND status = 1", commentID).
		First(&comment).Error
	return &comment, err
}

// GetRootComments 获取一级评论列表(按时间倒序)
func (d *Comment) GetRootComments(ctx context.Context, noteID uint64, offset, limit int) ([]*models.Comment, error) {
	var comments []*models.Comment
	err := d.Db.WithContext(ctx).
		Where("note_id = ? AND root_id = 0 AND status = 1", noteID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&comments).Error
	return comments, err
}

// GetRootCommentCount 获取一级评论总数
func (d *Comment) GetRootCommentCount(ctx context.Context, noteID uint64) (int64, error) {
	var count int64
	err := d.Db.WithContext(ctx).
		Model(&models.Comment{}).
		Where("note_id = ? AND root_id = 0 AND status = 1", noteID).
		Count(&count).Error
	return count, err
}

// GetLatestReplies 获取指定评论的最新N条回复
func (d *Comment) GetLatestReplies(ctx context.Context, rootID uint64, limit int) ([]*models.Comment, error) {
	var replies []*models.Comment
	err := d.Db.WithContext(ctx).
		Where("root_id = ? AND status = 1", rootID).
		Order("created_at DESC").
		Limit(limit).
		Find(&replies).Error
	return replies, err
}

// BatchGetLatestReplies 批量获取多个评论的最新回复
func (d *Comment) BatchGetLatestReplies(ctx context.Context, rootIDs []uint64, limit int) (map[uint64][]*models.Comment, error) {
	var replies []*models.Comment
	err := d.Db.WithContext(ctx).
		Where("root_id IN ? AND status = 1", rootIDs).
		Order("root_id, created_at DESC").
		Find(&replies).Error

	if err != nil {
		return nil, err
	}

	// 分组
	result := make(map[uint64][]*models.Comment)
	for _, reply := range replies {
		if result[reply.RootID] == nil {
			result[reply.RootID] = make([]*models.Comment, 0)
		}
		if len(result[reply.RootID]) < limit {
			result[reply.RootID] = append(result[reply.RootID], reply)
		}
	}

	return result, nil
}

// GetReplies 获取某条评论的所有回复(分页)
func (d *Comment) GetReplies(ctx context.Context, rootID uint64, offset, limit int) ([]*models.Comment, error) {
	var replies []*models.Comment
	err := d.Db.WithContext(ctx).
		Where("root_id = ? AND status = 1", rootID).
		Order("created_at ASC"). // 回复按时间正序
		Offset(offset).
		Limit(limit).
		Find(&replies).Error
	return replies, err
}

// GetReplyCount 获取回复总数
func (d *Comment) GetReplyCount(ctx context.Context, rootID uint64) (int64, error) {
	var count int64
	err := d.Db.WithContext(ctx).
		Model(&models.Comment{}).
		Where("root_id = ? AND status = 1", rootID).
		Count(&count).Error
	return count, err
}

// Delete 删除评论(软删除)
func (d *Comment) Delete(ctx context.Context, commentID uint64) error {
	return d.Db.WithContext(ctx).
		Model(&models.Comment{}).
		Where("id = ?", commentID).
		Update("status", 0).Error
}

// IncrementReplyCount 增加回复数
func (d *Comment) IncrementReplyCount(ctx context.Context, commentID uint64, delta int) error {
	return d.Db.WithContext(ctx).
		Model(&models.Comment{}).
		Where("id = ?", commentID).
		UpdateColumn("reply_count", gorm.Expr("reply_count + ?", delta)).
		Error
}

// IncrementLikeCount 增加点赞数
func (d *Comment) IncrementLikeCount(ctx context.Context, commentID uint64, delta int) error {
	return d.Db.WithContext(ctx).
		Model(&models.Comment{}).
		Where("id = ?", commentID).
		UpdateColumn("like_count", gorm.Expr("like_count + ?", delta)).
		Error
}

// Transaction 事务
func (d *Comment) Transaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return d.Db.WithContext(ctx).Transaction(fn)
}
