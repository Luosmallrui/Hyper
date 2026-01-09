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
	GetGroupId(ctx context.Context, GroupID int) (*models.Group, error)
}

type GroupService struct {
	DB *gorm.DB
}

func NewGroupService(db *gorm.DB) *GroupService {
	return &GroupService{DB: db}
}

func (s *GroupService) GetGroupId(ctx context.Context, GroupID int) (*models.Group, error) {
	var group models.Group
	if err := s.DB.WithContext(ctx).Where("id = ?", GroupID).First(&group).Error; err != nil {
		return nil, errors.New("获取群消息失败: " + err.Error())
	}
	return &group, nil
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
func (s *GroupService) InviteMembers(ctx context.Context, groupId int, usersIds []int, userId int) (*models.InviteMemberResponse, error) {

	//1、群存在且邀请者在群内且未退出（群成员都可以邀请人）
	var member models.GroupMember
	if err := s.DB.WithContext(ctx).
		Where("group_id = ? AND user_id = ? AND is_quit = 1", groupId, userId).
		First(&member).
		Error; err != nil {
		return nil, errors.New("邀请者不在群内或已退出")
	}
	//2、群信息
	group, err := s.GetGroupId(ctx, groupId)
	if err != nil {
		return nil, err
	}
	//3、群是否满了
	if group.MemberCount+len(usersIds) > group.MaxMembers {
		return nil, errors.New("群成员已满，无法邀请更多成员")
	}
	//4、邀请成员入群
	resp := &models.InviteMemberResponse{
		FailedUserIds: []int{},
	}
	//去重
	uniqueUserIds := make(map[int]bool)
	for _, uid := range usersIds {
		uniqueUserIds[uid] = true
	}
	//邀请成员入群
	for newUserId := range uniqueUserIds {
		//不能邀请自己
		if newUserId != userId {
			resp.FailedCount++
			resp.FailedUserIds = append(resp.FailedUserIds, newUserId)
			continue
		}

		var existMember models.GroupMember
		if err := s.DB.WithContext(ctx).
			Where("group_id = ? AND user_id = ? AND is_quit = 1", groupId, newUserId).
			First(&existMember).Error; err == nil {
			//已经是群成员
			resp.FailedCount++
			resp.FailedUserIds = append(resp.FailedUserIds, newUserId)
			continue
		}

		groupMember := &models.GroupMember{
			GroupId:   groupId,
			UserId:    newUserId,
			Leader:    3, //普通成员
			IsQuit:    1,
			IsMute:    1,
			JoinTime:  time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		//添加成员
		if err := s.DB.WithContext(ctx).
			Assign(groupMember).
			FirstOrCreate(&groupMember).Error; err != nil {
			resp.FailedCount++
			resp.FailedUserIds = append(resp.FailedUserIds, newUserId)
			continue
		}
		resp.SuccessCount++
	}
	if resp.SuccessCount > 0 {
		if err := s.DB.WithContext(ctx).
			Where("id = ?", groupId).
			Update("member_count", gorm.Expr("member_count + ?", resp.SuccessCount)).Error; err != nil {
			return nil, errors.New("更新群成员数量失败: " + err.Error())
		}
	}
	return resp, nil
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
