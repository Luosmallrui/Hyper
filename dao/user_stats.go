package dao

import (
	"Hyper/models"
	"context"
	"time"

	"gorm.io/gorm"
)

type UserStatsDAO struct {
	Repo[models.UserStats]
}

func NewUserStatsDAO(db *gorm.DB) *UserStatsDAO {
	return &UserStatsDAO{
		Repo: NewRepo[models.UserStats](db),
	}
}

// GetOrCreate 获取或创建用户统计
func (d *UserStatsDAO) GetOrCreate(ctx context.Context, userID uint64) (*models.UserStats, error) {
	stats := &models.UserStats{UserID: userID}
	err := d.Db.WithContext(ctx).
		Where("user_id = ?", userID).
		FirstOrCreate(stats).Error
	return stats, err
}

// IncrFollowerCount 增加粉丝数
func (d *UserStatsDAO) IncrFollowerCount(ctx context.Context, userID uint64, delta int) error {
	now := time.Now()
	return d.Db.WithContext(ctx).Exec(`
		INSERT INTO user_stats (user_id, follower_count, created_at, updated_at) 
		VALUES (?, GREATEST(?, 0), ?, ?)
		ON DUPLICATE KEY UPDATE 
			follower_count = GREATEST(follower_count + ?, 0),
			updated_at = VALUES(updated_at)
	`, userID, delta, now, now, delta).Error
}

// IncrFollowingCount 增加关注数
func (d *UserStatsDAO) IncrFollowingCount(ctx context.Context, userID uint64, delta int) error {
	now := time.Now()
	return d.Db.WithContext(ctx).Exec(`
		INSERT INTO user_stats (user_id, following_count, created_at, updated_at) 
		VALUES (?, GREATEST(?, 0), ?, ?)
		ON DUPLICATE KEY UPDATE 
			following_count = GREATEST(following_count + ?, 0),
			updated_at = VALUES(updated_at)
	`, userID, delta, now, now, delta).Error
}

// GetByUserID 根据用户ID获取统计
func (d *UserStatsDAO) GetByUserID(ctx context.Context, userID uint64) (*models.UserStats, error) {
	var stats models.UserStats
	err := d.Db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&stats).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &stats, err
}
