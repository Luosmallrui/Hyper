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
func (d *NoteCollectionDAO) SetStatus(ctx context.Context, noteID uint64, userID uint64, status uint8) error {
	return d.Db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var item models.NoteCollection
		err := tx.Where("note_id = ? AND user_id = ?", noteID, int(userID)).Limit(1).Find(&item).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
		}
		if err != nil {
			return err
		}
		if item.ID == 0 {
			item = models.NoteCollection{NoteID: noteID, UserID: int(userID), Status: status}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
			return nil
		}
		return tx.Model(&models.NoteCollection{}).Where("id = ?", item.ID).Update("status", status).Error
	})
}

// IsCollected 是否已收藏（status=1）
func (d *NoteCollectionDAO) IsCollected(ctx context.Context, noteID uint64, userID uint64) (bool, error) {
	exist, err := d.IsExist(ctx, "note_id = ? AND user_id = ? AND status = 1", noteID, int(userID))
	if err != nil {
		return false, err
	}
	return exist, nil
}
