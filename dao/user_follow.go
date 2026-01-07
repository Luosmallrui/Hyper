package dao

import (
	"Hyper/models"
	"context"
	"time"

	"gorm.io/gorm"
)

type UserFollowDAO struct {
	Repo[models.UserFollow]
}

func NewUserFollowDAO(db *gorm.DB) *UserFollowDAO {
	return &UserFollowDAO{
		Repo: NewRepo[models.UserFollow](db),
	}
}

// IsFollowing 检查是否已关注
func (d *UserFollowDAO) IsFollowing(ctx context.Context, followerID, followeeID uint64) (bool, error) {
	var follow models.UserFollow
	err := d.Db.WithContext(ctx).
		Where("follower_id = ? AND followee_id = ? AND status = 1", followerID, followeeID).
		First(&follow).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// SetStatus 设置关注状态（如不存在则创建）
func (d *UserFollowDAO) SetStatus(ctx context.Context, followerID, followeeID uint64, status int) error {
	now := time.Now()

	// 优先更新已有记录，避免 OnConflict 未命中导致不更新的情况
	res := d.Db.WithContext(ctx).
		Model(&models.UserFollow{}).
		Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		Updates(map[string]any{
			"status":     status,
			"updated_at": now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		return nil
	}

	// 不存在则插入
	follow := models.UserFollow{
		FollowerID: followerID,
		FolloweeID: followeeID,
		Status:     status,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	return d.Db.WithContext(ctx).Create(&follow).Error
}

// GetFollowerCount 获取粉丝数
func (d *UserFollowDAO) GetFollowerCount(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	err := d.Db.WithContext(ctx).
		Model(&models.UserFollow{}).
		Where("followee_id = ? AND status = 1", userID).
		Count(&count).Error
	return count, err
}

// GetFollowingCount 获取关注数
func (d *UserFollowDAO) GetFollowingCount(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	err := d.Db.WithContext(ctx).
		Model(&models.UserFollow{}).
		Where("follower_id = ? AND status = 1", userID).
		Count(&count).Error
	return count, err
}

// GetFollowingList 获取用户关注的用户列表（按关注时间倒序）
func (d *UserFollowDAO) GetFollowingList(ctx context.Context, userID uint64, limit, offset int) ([]map[string]interface{}, int64, error) {
	var follows []map[string]interface{}
	var total int64

	// 查询关注关系总数
	err := d.Db.WithContext(ctx).
		Model(&models.UserFollow{}).
		Where("follower_id = ? AND status = 1", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 联接用户表获取用户信息，按创建时间倒序
	err = d.Db.WithContext(ctx).
		Table("user_follow uf").
		Select("u.id as user_id, u.nickname").
		Joins("LEFT JOIN users u ON uf.followee_id = u.id").
		Where("uf.follower_id = ? AND uf.status = 1", userID).
		Order("uf.created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&follows).Error

	return follows, total, err
}
