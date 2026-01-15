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

// GetUserTotalLikes 统计用户所有笔记的总点赞数
func (d *NoteStatsDAO) GetUserTotalLikes(ctx context.Context, userID uint64) (int64, error) {
	var total int64
	err := d.Db.WithContext(ctx).
		Raw(`
			SELECT COALESCE(SUM(ns.like_count), 0) as total
			FROM note_stats ns
			INNER JOIN notes n ON ns.note_id = n.id
			WHERE n.user_id = ?
		`, userID).
		Scan(&total).Error
	return total, err
}

// GetUserTotalCollects 统计用户所有笔记的总收藏数
func (d *NoteStatsDAO) GetUserTotalCollects(ctx context.Context, userID uint64) (int64, error) {
	var total int64
	err := d.Db.WithContext(ctx).
		Raw(`
			SELECT COALESCE(SUM(ns.coll_count), 0) as total
			FROM note_stats ns
			INNER JOIN notes n ON ns.note_id = n.id
			WHERE n.user_id = ?
		`, userID).
		Scan(&total).Error
	return total, err
}

func (d *NoteStatsDAO) BatchGetByNoteIDs(ctx context.Context, noteIDs []uint64) ([]*models.NoteStats, error) {
	var stats []*models.NoteStats
	err := d.Db.WithContext(ctx).
		Where("note_id IN ?", noteIDs).
		Find(&stats).Error
	return stats, err
}

func (d *NoteStatsDAO) GetByNoteID(ctx context.Context, noteID uint64) (*models.NoteStats, error) {
	var stats models.NoteStats
	err := d.Db.WithContext(ctx).
		Where("note_id = ?", noteID).
		First(&stats).Error
	return &stats, err
}
