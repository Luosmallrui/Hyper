package dao

import (
	"Hyper/models"
	"context"

	"gorm.io/gorm"
)

type NoteStatsDAO struct {
	Repo[models.NoteStats]
}

func NewNoteStatsDAO(db *gorm.DB) *NoteStatsDAO {
	return &NoteStatsDAO{Repo: NewRepo[models.NoteStats](db)}
}

// IncrLikeCount 点赞计数增减，避免负数
func (d *NoteStatsDAO) IncrLikeCount(ctx context.Context, noteID uint64, delta int64) error {
	// 使用原生 SQL 做 UPSERT 并限制不为负
	return d.Db.WithContext(ctx).Exec(
		"INSERT INTO note_stats (note_id, like_count, updated_at) VALUES (?, GREATEST(?, 0), NOW()) "+
			"ON DUPLICATE KEY UPDATE like_count = GREATEST(like_count + ?, 0), updated_at = NOW()",
		noteID, delta, delta,
	).Error
}

// IncrCollCount 收藏计数增减，避免负数
func (d *NoteStatsDAO) IncrCollCount(ctx context.Context, noteID uint64, delta int64) error {
	return d.Db.WithContext(ctx).Exec(
		"INSERT INTO note_stats (note_id, coll_count, updated_at) VALUES (?, GREATEST(?, 0), NOW()) "+
			"ON DUPLICATE KEY UPDATE coll_count = GREATEST(coll_count + ?, 0), updated_at = NOW()",
		noteID, delta, delta,
	).Error
}

func (d *NoteStatsDAO) GetByNoteID(ctx context.Context, noteID uint64) (*models.NoteStats, error) {
	var item models.NoteStats
	err := d.Db.WithContext(ctx).Where("note_id = ?", noteID).Limit(1).Find(&item).Error
	if err != nil {
		return nil, err
	}
	if item.NoteID == 0 {
		return &models.NoteStats{NoteID: noteID, LikeCount: 0, CollCount: 0, ShareCount: 0, CommentCount: 0}, nil
	}
	return &item, nil
}
