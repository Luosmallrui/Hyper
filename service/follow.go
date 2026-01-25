package service

import (
	"Hyper/dao"
	"Hyper/models"
	"Hyper/types"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
)

var _ IFollowService = (*FollowService)(nil)

type IFollowService interface {
	Follow(ctx context.Context, followerID, followeeID uint64) error
	Unfollow(ctx context.Context, followerID, followeeID uint64) error
	IsFollowing(ctx context.Context, followerID, followeeID uint64) (bool, error)
	GetFollowerCount(ctx context.Context, userID uint64) (int64, error)
	GetFollowingCount(ctx context.Context, userID uint64) (int64, error)
	GetFollowingList(ctx context.Context, userID uint64, limit int64, offset int) ([]*models.FollowingQueryResult, error)
	CheckFollowStatus(ctx context.Context, followerID, followeeID uint64) (bool, error)
	GetMyFollowingListWithStatus(ctx context.Context, myID uint64, cursor int64, limit int) ([]*models.FollowingQueryResult, error)
}

type FollowService struct {
	FollowDAO *dao.UserFollowDAO
	StatsDAO  *dao.UserStatsDAO
	UserDAO   *dao.Users
	Producer  rmq_client.Producer
	Redis     *redis.Client
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

	// 发送 MQ 通知
	go func() {
		// 查询关注者信息
		follower, err := s.UserDAO.FindById(ctx, followerID)
		if err != nil {
			log.Printf("[MQ] 获取关注者信息失败: %v", err)
			return
		}

		payload := &types.FollowPayload{
			UserId:    int(followerID),
			TargetId:  int(followeeID),
			Avatar:    follower.Avatar,
			Nickname:  follower.Nickname,
			CreatedAt: time.Now().Format(time.RFC3339),
		}
		dataBytes, _ := json.Marshal(payload)

		msgMap := &types.SystemMessage{
			Type: "follow",
			Data: json.RawMessage(dataBytes),
		}
		body, _ := json.Marshal(msgMap)

		msg := &rmq_client.Message{
			Topic: "hyper_system_messages",
			Body:  body,
		}

		if _, err := s.Producer.Send(ctx, msg); err != nil {
			log.Printf("[MQ] 发送关注通知失败: %v", err)
		}
	}()

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

func (s *FollowService) GetFollowingList(ctx context.Context, userID uint64, limit int64, offset int) ([]*models.FollowingQueryResult, error) {
	return s.FollowDAO.GetFollowingFeed(ctx, userID, limit, offset)
}

func (s *FollowService) GetMyFollowingListWithStatus(ctx context.Context, myID uint64, cursor int64, limit int) ([]*models.FollowingQueryResult, error) {
	// 1. 获取我关注的人（DAO 层逻辑不变）
	list, err := s.FollowDAO.GetFollowingFeed(ctx, myID, cursor, limit)
	if err != nil || len(list) == 0 {
		return list, err
	}

	// 2. 提取列表中的用户 ID
	targetIDs := make([]uint64, 0, len(list))
	for _, item := range list {
		targetIDs = append(targetIDs, item.UserID)
	}

	// 3. 只需要查一步：这 20 个人里，谁关注了我？
	// SQL: SELECT follower_id FROM user_follow WHERE follower_id IN (...) AND followee_id = myID
	followMeMap := make(map[uint64]bool)
	var followMeIDs []uint64

	err = s.FollowDAO.Db.WithContext(ctx).
		Model(&models.UserFollow{}).
		Where("follower_id IN ? AND followee_id = ? AND status = 1", targetIDs, myID).
		Pluck("follower_id", &followMeIDs).Error

	if err == nil {
		for _, id := range followMeIDs {
			followMeMap[id] = true
		}
	}

	// 4. 填充状态
	for _, item := range list {
		// 因为是“我的关注”列表，我肯定关注了他们
		item.IsFollowing = true

		// 如果他们也关注了我，那就是互关
		if followMeMap[item.UserID] {
			item.IsMutual = true
		}
		item.Signature = "这个人很懒，什么都没有留下.."
	}

	return list, nil
}
func (s *FollowService) CheckFollowStatus(ctx context.Context, followerID, followeeID uint64) (bool, error) {
	if followerID == 0 || followerID == followeeID {
		return false, nil
	}

	// 类似的逻辑
	key := fmt.Sprintf("user:following:%d", followerID)
	exists, err := s.Redis.SIsMember(ctx, key, followeeID).Result()
	if err == nil {
		return exists, nil
	}

	return s.FollowDAO.CheckExists(ctx, followerID, followeeID)
}
