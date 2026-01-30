package service

import (
	"Hyper/models"
	"Hyper/types"
	"context"

	"github.com/cloudwego/kitex/pkg/kerrors"
	"gorm.io/gorm"
)

var _ IChannelService = (*ChannelService)(nil)

type IChannelService interface {
	CreateChannel(ctx context.Context, req *types.CreateChannelReq) (*types.CreateChannelResp, error)
	ListChannels(ctx context.Context, req *types.ListChannelsReq) (*types.ListChannelsResp, error)
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
	Db *gorm.DB
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
