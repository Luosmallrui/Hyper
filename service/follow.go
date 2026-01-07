package service

import (
	"Hyper/dao"
	"context"
	"errors"
)

var _ IFollowService = (*FollowService)(nil)

type IFollowService interface {
	Follow(ctx context.Context, followerID, followeeID uint64) error
	Unfollow(ctx context.Context, followerID, followeeID uint64) error
	IsFollowing(ctx context.Context, followerID, followeeID uint64) (bool, error)
	GetFollowerCount(ctx context.Context, userID uint64) (int64, error)
	GetFollowingCount(ctx context.Context, userID uint64) (int64, error)
	GetFollowingList(ctx context.Context, userID uint64, limit, offset int) ([]map[string]interface{}, int64, error)
}

type FollowService struct {
	FollowDAO *dao.UserFollowDAO
	StatsDAO  *dao.UserStatsDAO
	UserDAO   *dao.Users
}

func (s *FollowService) Follow(ctx context.Context, followerID, followeeID uint64) error {
	// 不能关注自己
	if followerID == followeeID {
		return errors.New("不能关注自己")
	}

	// 校验被关注用户是否存在
	exist, err := s.UserDAO.IsExist(ctx, "id = ?", followeeID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New("用户不存在")
	}

	// 检查是否已经关注
	isFollowing, err := s.FollowDAO.IsFollowing(ctx, followerID, followeeID)
	if err != nil {
		return err
	}
	if isFollowing {
		// 已经关注过，直接返回成功
		return nil
	}

	// 设置关注状态
	if err := s.FollowDAO.SetStatus(ctx, followerID, followeeID, 1); err != nil {
		return err
	}

	// 更新统计：被关注人的粉丝数+1，关注人的关注数+1
	if err := s.StatsDAO.IncrFollowerCount(ctx, followeeID, 1); err != nil {
		return err
	}
	if err := s.StatsDAO.IncrFollowingCount(ctx, followerID, 1); err != nil {
		return err
	}

	return nil
}

func (s *FollowService) Unfollow(ctx context.Context, followerID, followeeID uint64) error {
	// 不能取消关注自己
	if followerID == followeeID {
		return errors.New("不能取消关注自己")
	}

	// 校验被关注用户是否存在
	exist, err := s.UserDAO.IsExist(ctx, "id = ?", followeeID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New("用户不存在")
	}

	// 检查是否已经关注
	isFollowing, err := s.FollowDAO.IsFollowing(ctx, followerID, followeeID)
	if err != nil {
		return err
	}
	if !isFollowing {
		// 没有关注过，直接返回成功
		return nil
	}

	// 设置取消关注状态
	if err := s.FollowDAO.SetStatus(ctx, followerID, followeeID, 0); err != nil {
		return err
	}

	// 更新统计：被关注人的粉丝数-1，关注人的关注数-1
	if err := s.StatsDAO.IncrFollowerCount(ctx, followeeID, -1); err != nil {
		return err
	}
	if err := s.StatsDAO.IncrFollowingCount(ctx, followerID, -1); err != nil {
		return err
	}

	return nil
}

func (s *FollowService) IsFollowing(ctx context.Context, followerID, followeeID uint64) (bool, error) {
	return s.FollowDAO.IsFollowing(ctx, followerID, followeeID)
}

func (s *FollowService) GetFollowerCount(ctx context.Context, userID uint64) (int64, error) {
	stats, err := s.StatsDAO.GetByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}
	if stats == nil {
		return 0, nil
	}
	return int64(stats.FollowerCount), nil
}

func (s *FollowService) GetFollowingCount(ctx context.Context, userID uint64) (int64, error) {
	stats, err := s.StatsDAO.GetByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}
	if stats == nil {
		return 0, nil
	}
	return int64(stats.FollowingCount), nil
}

func (s *FollowService) GetFollowingList(ctx context.Context, userID uint64, limit, offset int) ([]map[string]interface{}, int64, error) {
	return s.FollowDAO.GetFollowingList(ctx, userID, limit, offset)
}
