package service

import (
	"Hyper/models"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type IGroupService interface {
	CreateGroup(ctx context.Context, req *models.CreateGroupRequest, userId int) (*models.Group, error)
}

type GroupService struct {
	DB *gorm.DB
}

func NewGroupService(db *gorm.DB) *GroupService {
	return &GroupService{DB: db}
}

// 创建群
func (s *GroupService) CreateGroup(ctx context.Context, req *models.CreateGroupRequest, userId int) (*models.Group, error) {
	group := &models.Group{
		Name:        req.Name,
		Avatar:      req.Avatar,
		Description: req.Description,
		OwnerId:     userId, //设置群主为创建者
		MemberCount: 1,      //初始成员数量为1
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
		Leader:    1,
		IsQuit:    1,
		IsMute:    1,
		JoinTime:  time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.DB.WithContext(ctx).Create(groupMember).Error; err != nil {
		return nil, errors.New("添加群成员失败: " + err.Error())
	}
	return group, nil
}

// 邀请成员
func (s *GroupService) InviteMembers(ctx context.Context, req *models.InviteMemberRequest) error {
	return nil
}

// 踢出成员
func (s *GroupService) KickMember(ctx context.Context, req *models.KickMemberRequest) error {
	return nil
}

// 转移群主
func (s *GroupService) TransferOwner(ctx context.Context, req *models.TransferOwnerRequest) error {
	return nil
}

// 更新群信息
func (s *GroupService) UpdateGroup(ctx context.Context, req *models.UpdateGroupRequest) error {
	return nil
}
