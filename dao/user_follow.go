package dao

import (
	"Hyper/models"
	"context"
	"errors"
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
	if errors.Is(err, gorm.ErrRecordNotFound) {
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

func (d *UserFollowDAO) GetFollowingFeed(ctx context.Context, userID uint64, cursor int64, limit int) ([]*models.FollowingQueryResult, error) {
	var follows []*models.FollowingQueryResult

	db := d.Db.WithContext(ctx).
		Table("user_follow uf").
		Select("u.id as user_id, u.nickname, u.avatar, uf.created_at as follow_time").
		Joins("LEFT JOIN users u ON uf.followee_id = u.id").
		Where("uf.follower_id = ? AND uf.status = 1", userID)

	if cursor > 0 {
		// 纳秒转 time.Time
		cursorTime := time.Unix(0, cursor)
		db = db.Where("uf.created_at < ?", cursorTime)
	}

	// 执行查询，GORM 会自动根据 tag 映射到结构体
	err := db.Order("uf.created_at DESC").
		Limit(limit).
		Scan(&follows).Error // 注意：Scan 或 Find 都可以，Find 更语义化

	return follows, err
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
		Select("u.id as user_id, u.nickname, u.avatar as avatar, uf.updated_at as updated_at").
		Joins("LEFT JOIN users u ON uf.followee_id = u.id").
		Where("uf.follower_id = ? AND uf.status = 1", userID).
		Order("uf.created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&follows).Error

	return follows, total, err
}

func (d *UserFollowDAO) CheckExists(ctx context.Context, followerID, followeeID uint64) (bool, error) {
	var uf models.UserFollow

	err := d.Db.WithContext(ctx).
		Where("follower_id = ? AND followee_id = ? AND status = 1", followerID, followeeID).
		Limit(1).
		Find(&uf).Error

	if err != nil {
		return false, err
	}
	return uf.ID != 0, nil
}

func (d *UserFollowDAO) GetFollowerFeed(ctx context.Context, userID uint64, cursor int64, limit int) ([]*models.FollowingQueryResult, error) {
	var results []*models.FollowingQueryResult

	query := d.Db.WithContext(ctx).
		Table("user_follow AS uf").
		Select(`
            uf.follower_id AS user_id,
            u.nickname,
            u.avatar AS avatar,
            u.signature,
            uf.updated_at
        `).
		Joins("LEFT JOIN users AS u ON uf.follower_id = u.id").
		Where("uf.followee_id = ? AND uf.status = 1", userID).
		Order("uf.updated_at DESC")

	// 如果有 cursor，则从该时间点开始查询
	if cursor > 0 {
		cursorTime := time.Unix(cursor, 0)
		query = query.Where("uf.updated_at < ?", cursorTime)
	}

	err := query.Limit(limit).Find(&results).Error
	return results, err
}

func (d *UserFollowDAO) GetFollowingIDs(ctx context.Context, userID int) ([]int, error) {
	followingIds := make([]int, 0)
	err := d.Db.WithContext(ctx).Model(&models.UserFollow{}).
		Where("follower_id = ?", userID).
		Pluck("followee_id", &followingIds).Error
	if err != nil {
		return followingIds, err
	}
	return followingIds, nil
}
