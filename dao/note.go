package dao

import (
	"Hyper/models"
	"context"

	"gorm.io/gorm"
)

type NoteDAO struct {
	Repo[models.Note]
}

func NewNoteDAO(db *gorm.DB) *NoteDAO {
	return &NoteDAO{Repo: NewRepo[models.Note](db)}
}

// Create 创建笔记
func (d *NoteDAO) Create(ctx context.Context, note *models.Note) error {
	return d.Db.WithContext(ctx).Create(note).Error
}

// FindByUserID 根据用户ID查询笔记列表
func (d *NoteDAO) FindByUserID(ctx context.Context, userID uint64, status int, limit, offset int) ([]*models.Note, error) {
	var notes []*models.Note
	err := d.Db.WithContext(ctx).
		// Debug().
		Where("user_id = ? AND status = ?", userID, status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notes).Error
	return notes, err
}

// UpdateStatus 更新笔记状态
func (d *NoteDAO) UpdateStatus(ctx context.Context, noteID uint64, status int) error {
	return d.Db.WithContext(ctx).
		Model(&models.Note{}).
		Where("id = ?", noteID).
		Update("status", status).Error
}

// FindByIDs 根据 ID 列表查询笔记
func (d *NoteDAO) FindByIDs(ctx context.Context, ids []uint64) ([]*models.Note, error) {
	if len(ids) == 0 {
		return []*models.Note{}, nil
	}
	var notes []*models.Note
	err := d.Db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&notes).Error
	return notes, err
}
