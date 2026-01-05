package dao

import (
	"Hyper/models"
	"context"
	"errors"

	"gorm.io/gorm"
)

type NoteLikeDAO struct {
	Repo[models.NoteLike]
}

func NewNoteLikeDAO(db *gorm.DB) *NoteLikeDAO {
	return &NoteLikeDAO{Repo: NewRepo[models.NoteLike](db)}
}

// GetByNoteUser 查询指定用户对指定笔记的点赞记录
func (d *NoteLikeDAO) GetByNoteUser(ctx context.Context, noteID uint64, userID uint64) (*models.NoteLike, error) {
	var item models.NoteLike
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

// SetStatus 设置点赞状态，如果不存在则创建
func (d *NoteLikeDAO) SetStatus(ctx context.Context, noteID uint64, userID uint64, status uint8) error {
	return d.Db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var item models.NoteLike
		err := tx.Where("note_id = ? AND user_id = ?", noteID, int(userID)).Limit(1).Find(&item).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
		}
		if err != nil {
			return err
		}
		if item.ID == 0 { // create
			item = models.NoteLike{NoteID: noteID, UserID: int(userID), Status: status}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
			return nil
		}
		// update
		return tx.Model(&models.NoteLike{}).Where("id = ?", item.ID).Update("status", status).Error
	})
}

// IsLiked 是否点赞（status=1）
func (d *NoteLikeDAO) IsLiked(ctx context.Context, noteID uint64, userID uint64) (bool, error) {
	exist, err := d.IsExist(ctx, "note_id = ? AND user_id = ? AND status = 1", noteID, int(userID))
	if err != nil {
		return false, err
	}
	return exist, nil
}
