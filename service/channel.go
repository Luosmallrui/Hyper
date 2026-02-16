package service

import (
	"Hyper/models"
	"Hyper/types"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var _ IChannelService = (*ChannelService)(nil)

type IChannelService interface {
	CreateChannel(ctx context.Context, req *types.CreateChannelReq) (*types.CreateChannelResp, error)
	ListChannels(ctx context.Context, req *types.ListChannelsReq) (*types.ListChannelsResp, error)
	GetGlobalChannels(ctx context.Context) ([]models.Channel, error)
	UnsubscribeChannel(ctx context.Context, userID int, channelID int) error
	SubscribeChannel(ctx context.Context, userID int, channelID int) error
	GetUserChannelView(ctx context.Context, userID int, AllGlobalChannels []models.Channel) (*types.ChannelViewResponse, error)
}

func (s *ChannelService) ListChannels(ctx context.Context, req *types.ListChannelsReq) (*types.ListChannelsResp, error) {
	var channels []models.Channel

	// 构建查询：只查可见的
	db := s.Db.WithContext(ctx).Where("is_visible = ?", true)

	// 按父类目过滤（可选）
	if req.ParentId > 0 {
		db = db.Where("parent_id = ?", req.ParentId)
	}

	// 排序逻辑：权重越大越靠前 (sort_weight DESC)
	err := db.Order("sort_weight DESC, created_at ASC").Find(&channels).Error
	if err != nil {
		return nil, err
	}

	// 将数据库模型转换为返回的 types 类型
	data := make([]*types.ChannelInfo, 0, len(channels))
	for _, v := range channels {
		data = append(data, &types.ChannelInfo{
			Id:         v.ID,
			Name:       v.Name,
			EnName:     v.EnName,
			IconUrl:    v.IconURL,
			SortWeight: v.SortWeight,
			ParentId:   v.ParentID,
		})
	}

	return &types.ListChannelsResp{
		Channels: data,
	}, nil
}

type ChannelService struct {
	Db    *gorm.DB
	Redis *redis.Client
}

// CreateChannel 实现 IDL 定义的接口
func (s *ChannelService) CreateChannel(ctx context.Context, req *types.CreateChannelReq) (*types.CreateChannelResp, error) {

	// 1. 构造模型
	newChannel := &models.Channel{
		Name:        req.Name,
		EnName:      req.EnName,
		IconURL:     req.IconUrl,
		Description: req.Description,
		SortWeight:  req.SortWeight,
		ParentID:    req.ParentId,
		IsVisible:   true,
	}

	// 2. 写入数据库 (GORM)
	if err := s.Db.WithContext(ctx).Create(newChannel).Error; err != nil {
		// 处理唯一索引冲突 (名称重复)
		return nil, kerrors.NewBizStatusError(50001, "频道名称已存在")
	}
	resp := &types.CreateChannelResp{
		ChannelId: newChannel.ID,
	}
	return resp, nil
}

func (s *ChannelService) GenerateKey(useID int) string {
	return fmt.Sprintf("user:channel:%d", useID)
}

func (s *ChannelService) GetUserChannelView(ctx context.Context, userID int, AllGlobalChannels []models.Channel) (*types.ChannelViewResponse, error) {
	key := s.GenerateKey(userID)

	count, err := s.Redis.ZCard(ctx, key).Result()
	if err != nil {
		return nil, kerrors.NewBizStatusError(50001, "获取用户频道失败")
	}
	if count == 0 && len(AllGlobalChannels) == 0 {
		limit := 8
		if len(AllGlobalChannels) < 8 {
			limit = len(AllGlobalChannels)
		}
		pipe := s.Redis.Pipeline()
		for i := 0; i < limit; i++ {
			pipe.ZAdd(ctx, key, redis.Z{
				Score:  float64(i),
				Member: AllGlobalChannels[i].ID,
			})
		}
	}

	subscribedIDsStrs, err := s.Redis.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, kerrors.NewBizStatusError(50002, "获取用户频道失败")
	}

	myChannels := make([]*models.Channel, 0)
	otherChannels := make([]*models.Channel, 0)
	// 构建全局频道的 ID 到 Channel 的映射
	globalChannels := make(map[int]models.Channel, 0)
	for _, ch := range AllGlobalChannels {
		globalChannels[ch.ID] = ch
	}
	subscribedSet := make(map[int]bool)

	for _, idStr := range subscribedIDsStrs {
		id, _ := strconv.Atoi(idStr) //redis 返回的是字符串，需要转换为整数
		if ch, exists := globalChannels[id]; exists {
			tempCh := ch
			myChannels = append(myChannels, &tempCh)
			subscribedSet[id] = true
		}
	}

	for _, ch := range AllGlobalChannels {
		if !subscribedSet[ch.ID] {
			tempCh := ch
			otherChannels = append(otherChannels, &tempCh)
		}
	}
	return &types.ChannelViewResponse{
		Mychannels:    myChannels,
		Otherchannels: otherChannels,
	}, nil
}

func (s *ChannelService) SubscribeChannel(ctx context.Context, userID int, channelID int) error {
	key := s.GenerateKey(userID)

	socre := float64(time.Now().UnixNano())

	err := s.Redis.ZAdd(ctx, key, redis.Z{Score: socre, Member: channelID}).Err()
	if err != nil {
		return kerrors.NewBizStatusError(50003, "订阅频道失败")
	}
	return nil
}

func (s *ChannelService) UnsubscribeChannel(ctx context.Context, userID int, channelID int) error {
	key := s.GenerateKey(userID)

	err := s.Redis.ZRem(ctx, key, channelID).Err()
	if err != nil {
		return kerrors.NewBizStatusError(50004, "取消订阅频道失败")
	}
	return nil
}

func (s *ChannelService) GetGlobalChannels(ctx context.Context) ([]models.Channel, error) {
	var channels []models.Channel
	val, err := s.Redis.Get(ctx, "app:global:channels").Result()

	if err == nil {
		if jsonErr := json.Unmarshal([]byte(val), &channels); jsonErr == nil {
			// 从缓存获取成功
			return channels, nil
		}
	} else if err != redis.Nil {
		// 其他 Redis 错误
	}
	err = s.Db.WithContext(ctx).
		Where("is_visible = ?", true).
		Order("sort_weight DESC, id ASC").
		Find(&channels).Error
	if err != nil {
		return nil, kerrors.NewBizStatusError(50000, "获取频道列表失败")
	}
	if len(channels) > 0 {
		data, _ := json.Marshal(channels)
		s.Redis.Set(ctx, "app:global:channels", string(data), 24*time.Hour)
	}
	return channels, nil
}
