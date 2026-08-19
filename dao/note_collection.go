package dao

import (
	"Hyper/models"
	"context"
	"errors"

	"gorm.io/gorm"
)

type NoteCollectionDAO struct {
	Repo[models.NoteCollection]
}

func NewNoteCollectionDAO(db *gorm.DB) *NoteCollectionDAO {
	return &NoteCollectionDAO{Repo: NewRepo[models.NoteCollection](db)}
}

func (d *NoteCollectionDAO) CheckExists(ctx context.Context, userID, noteID uint64) (bool, error) {
	var count int64
	err := d.Db.WithContext(ctx).
		Model(&models.NoteCollection{}).
		Where("user_id = ? AND note_id = ?", userID, noteID).
		Count(&count).Error
	return count > 0, err
}

// GetByNoteUser 查询指定用户对指定笔记的收藏记录
func (d *NoteCollectionDAO) GetByNoteUser(ctx context.Context, noteID uint64, userID uint64) (*models.NoteCollection, error) {
	var item models.NoteCollection
	err := d.Db.WithContext(ctx).Where("note_id = ? AND user_id = ?", noteID, int(userID)).Limit(1).Find(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if item.ID == 0 {
		return nil, nil
	}
	return &item, nil
}

// SetStatus 设置收藏状态，不存在则创建
// 使用 ON DUPLICATE KEY UPDATE 保证原子性，依赖 note_collections 的 uk(note_id, user_id) 唯一键
func (d *NoteCollectionDAO) SetStatus(ctx context.Context, noteID uint64, userID uint64, status uint8) error {
	return d.Db.WithContext(ctx).Exec(
		"INSERT INTO note_collections (note_id, user_id, status, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW()) "+
			"ON DUPLICATE KEY UPDATE status = VALUES(status), updated_at = NOW()",
		noteID, int(userID), status,
	).Error
}

// IsCollected 是否已收藏（status=1）
func (d *NoteCollectionDAO) IsCollected(ctx context.Context, noteID uint64, userID uint64) (bool, error) {
	exist, err := d.IsExist(ctx, "note_id = ? AND user_id = ? AND status = 1", noteID, int(userID))
	if err != nil {
		return false, err
	}
	return exist, nil
}

// ListNoteIDsByUser 查询用户收藏的笔记ID列表，按收藏时间倒序
func (d *NoteCollectionDAO) ListNoteIDsByUser(ctx context.Context, userID uint64, limit, offset int) ([]uint64, int64, error) {
	var total int64
	base := d.Db.WithContext(ctx).Model(&models.NoteCollection{}).Where("user_id = ? AND status = 1", int(userID))
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var ids []uint64
	err := base.Select("note_id").Order("created_at DESC").Limit(limit).Offset(offset).Scan(&ids).Error
	return ids, total, err
}
