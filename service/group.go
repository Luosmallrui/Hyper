package service

import (
	"Hyper/dao"
	"Hyper/dao/cache"
	"Hyper/models"
	"Hyper/types"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type IGroupService interface {
	CreateGroup(ctx context.Context, req *types.CreateGroupRequest, userId int) (*models.Group, error)
	DismissGroup(ctx context.Context, groupId int, userId int) error
	UnMuteAllMembers(ctx context.Context, groupId int, userId int) error
	MuteAllMembers(ctx context.Context, userId int, groupId int) error
	UpdateGroupName(ctx context.Context, groupId int, userId int, req *types.UpdateGroupNameRequest) error
	UpdateGroupAvatar(ctx context.Context, groupId int, userId int, req *types.UpdateGroupAvatarRequest) error
	UpdateGroupDescription(ctx context.Context, groupId int, userId int, req *types.UpdateGroupDescriptionRequest) error
}

type GroupService struct {
	DB             *gorm.DB
	GroupMemberDAO *dao.GroupMember
	GroupDAO       *dao.Group
	Relation       *cache.Relation
}

// 创建群
func (s *GroupService) CreateGroup(ctx context.Context, req *types.CreateGroupRequest, userId int) (*models.Group, error) {
	group := &models.Group{
		Name:        req.Name,
		Avatar:      req.Avatar,
		Description: req.Description,
		OwnerId:     userId, //设置群主为创建者
		MemberCount: 1,      //初始成员数量为1
		IsDismiss:   0,
		MaxMembers:  200,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := s.DB.WithContext(ctx).Create(group).Error; err != nil {
		return nil, errors.New("创建群失败: " + err.Error())
	}

	groupMember := &models.GroupMember{
		GroupId:   group.Id,
		UserId:    userId,
		Role:      1,
		IsQuit:    0,
		IsMute:    0,
		JoinTime:  time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.DB.WithContext(ctx).Create(groupMember).Error; err != nil {
		return nil, errors.New("添加群成员失败: " + err.Error())
	}
	return group, nil
}

func (s *GroupService) DismissGroup(ctx context.Context, groupId int, userId int) error {
	group, err := s.GroupDAO.GetGroup(ctx, groupId)
	if err != nil || group.OwnerId != userId {
		return errors.New("只有群主才能解散群")
	}
	if group.IsDismiss == 1 {
		return errors.New("群已解散")
	}
	// 2. 开启事务
	err = s.DB.Transaction(func(tx *gorm.DB) error {
		// 更新群组状态
		if err := tx.Model(&models.Group{}).
			Where("id = ?", groupId).
			Update("is_dismiss", 1).
			Error; err != nil {
			return errors.New("解散群失败: " + err.Error())
		}
		// 批量移除成员（逻辑删除）
		if err := tx.Model(&models.GroupMember{}).
			Where("group_id = ?", groupId).
			Update("is_quit", 1).Error; err != nil {
			return errors.New("移除群成员失败: " + err.Error())
		}
		return nil
	})
	if err != nil {
		return err
	}
	// 异步删除群关系缓存
	go func() {
		bgCtx := context.Background()
		uids, _ := s.GroupMemberDAO.GetMemberIds(bgCtx, groupId)
		s.Relation.BatchDelGroupRelation(ctx, uids, groupId)
	}()
	return nil
}

func (s *GroupService) MuteAllMembers(ctx context.Context, userId int, groupId int) error {
	group, err := s.GroupDAO.GetGroup(ctx, groupId)
	if err != nil {
		return errors.New("群组不存在")
	}
	if group.IsAllMute == 1 {
		return errors.New("群已处于全员禁言状态")
	}
	if group.OwnerId != userId {
		var member models.GroupMember
		err := s.DB.WithContext(ctx).
			Where("group_id = ? AND user_id = ? AND is_quit = 0", groupId, userId).
			First(&member).Error
		if err != nil {
			return errors.New("用户不是群成员")
		}
		if member.Role > 2 {
			return errors.New("只有群主和管理员可以设置全员禁言")
		}
	}
	err = s.DB.WithContext(ctx).
		Model(&models.Group{}).
		Where("id = ?", groupId).
		Update("is_all_mute", 1).Error
	if err != nil {
		return errors.New("设置全员禁言失败: " + err.Error())
	}
	return nil
}

func (s *GroupService) UnMuteAllMembers(ctx context.Context, groupId int, userId int) error {
	group, err := s.GroupDAO.GetGroup(ctx, groupId)
	if err != nil {
		return errors.New("群组不存在")
	}
	if group.IsAllMute == 0 {
		return errors.New("群未处于全员禁言状态")
	}

	if group.OwnerId != userId {
		var member models.GroupMember
		err := s.DB.WithContext(ctx).
			Where("group_id = ? AND user_id = ? AND is_quit = 0", groupId, userId).
			First(&member).Error
		if err != nil {
			return errors.New("用户不是群成员")
		}
		if member.Role > 2 {
			return errors.New("只有群主和管理员可以取消全员禁言")
		}
	}
	err = s.DB.WithContext(ctx).
		Model(&models.Group{}).
		Where("id = ?", groupId).
		Update("is_all_mute", 0).Error
	if err != nil {
		return errors.New("取消全员禁言失败: " + err.Error())
	}
	return nil
}

func (s *GroupService) UpdateGroupAvatar(ctx context.Context, groupId int, userId int, req *types.UpdateGroupAvatarRequest) error {
	group, err := s.GroupDAO.GetGroup(ctx, groupId)
	if err != nil {
		return errors.New("群组不存在")
	}
	if group.OwnerId != userId {
		return errors.New("只有群主才能修改群头像")
	}
	err = s.DB.WithContext(ctx).
		Model(&models.Group{}).
		Where("id = ?", groupId).
		Update("avatar", req.Avatar).Error
	if err != nil {
		return errors.New("修改群头像失败: " + err.Error())
	}
	return nil
}
func (s *GroupService) UpdateGroupName(ctx context.Context, groupId int, userId int, req *types.UpdateGroupNameRequest) error {
	group, err := s.GroupDAO.GetGroup(ctx, groupId)
	if err != nil {
		return errors.New("群组不存在")
	}
	if group.OwnerId != userId {
		return errors.New("只有群主才能修改群名称")
	}
	err = s.DB.WithContext(ctx).
		Model(&models.Group{}).
		Where("id = ?", groupId).
		Update("name", req.Name).Error
	if err != nil {
		return errors.New("修改群名称失败: " + err.Error())
	}
	return nil
}
func (s *GroupService) UpdateGroupDescription(ctx context.Context, groupId int, userId int, req *types.UpdateGroupDescriptionRequest) error {
	group, err := s.GroupDAO.GetGroup(ctx, groupId)
	if err != nil {
		return errors.New("群组不存在")
	}
	if group.OwnerId != userId {
		return errors.New("只有群主才能修改群描述")
	}
	err = s.DB.WithContext(ctx).
		Model(&models.Group{}).
		Where("id = ?", groupId).
		Update("description", req.Description).Error
	if err != nil {
		return errors.New("修改群描述失败: " + err.Error())
	}
	return nil
}
