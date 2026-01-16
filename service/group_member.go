package service

import (
	"Hyper/models"
	"Hyper/types"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type IGroupMemberService interface {
	InviteMembers(ctx context.Context, groupId int, InvitedUsersIds []int, userId int) (*types.InviteMemberResponse, error)
	KickMember(ctx context.Context, GroupId int, KickedUserIds int, userId int) error
	GetGroupId(ctx context.Context, GroupID int) (*models.Group, error)
}

type GroupMemberService struct {
	DB *gorm.DB
	//Redis           *redis.Client
	//GroupMemberRepo *dao.GroupMember
}

func NewGroupMemberService(db *gorm.DB) *GroupMemberService {
	return &GroupMemberService{
		DB: db,
		//Redis:           redisClient,
		//GroupMemberRepo: dao.NewGroupMemberDao(db),
	}
}

func (s *GroupMemberService) GetGroupId(ctx context.Context, GroupID int) (*models.Group, error) {
	var group models.Group
	if err := s.DB.WithContext(ctx).Where("id = ?", GroupID).First(&group).Error; err != nil {
		return nil, errors.New("获取群消息失败: " + err.Error())
	}
	return &group, nil
}

// 邀请成员,ctx,群ID,被邀请用户ID列表,邀请者用户ID（默认邀请就进来，不需要审批）（需要改逻辑）
func (s *GroupMemberService) InviteMembers(ctx context.Context, groupId int, InvitedUsersIds []int, userId int) (*types.InviteMemberResponse, error) {

	//1、群存在且邀请者在群内且未退出（群成员都可以邀请人）
	var member models.GroupMember
	if err := s.DB.WithContext(ctx).
		Where("group_id = ? AND user_id = ? AND is_quit = 0", groupId, userId).
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
	if group.MemberCount+len(InvitedUsersIds) > group.MaxMembers {
		return nil, errors.New("群成员已满，无法邀请更多成员")
	}
	//4、邀请成员入群
	resp := &types.InviteMemberResponse{
		FailedUserIds: []int{},
	}
	//去重
	uniqueUserIds := make(map[int]bool)
	for _, uid := range InvitedUsersIds {
		uniqueUserIds[uid] = true
	}
	//邀请成员入群
	for newUserId := range uniqueUserIds {
		//不能邀请自己
		if newUserId == userId {
			resp.FailedCount++
			resp.FailedUserIds = append(resp.FailedUserIds, newUserId)
			continue
		}

		var gm models.GroupMember
		err := s.DB.WithContext(ctx).
			Where("group_id = ? AND user_id = ?", groupId, newUserId).
			First(&gm).Error

		if err == nil {
			// 已存在记录：判断是否退群
			if gm.IsQuit == 0 {
				// 未退群：已经是成员
				resp.FailedCount++
				resp.FailedUserIds = append(resp.FailedUserIds, newUserId)
				continue
			}

			// 已退群：恢复入群
			if err := s.DB.WithContext(ctx).
				Model(&models.GroupMember{}).
				Where("group_id = ? AND user_id = ?", groupId, newUserId).
				Updates(map[string]interface{}{
					"is_quit":   0,
					"is_mute":   0,
					"join_time": time.Now(),
				}).Error; err != nil {
				resp.FailedCount++
				resp.FailedUserIds = append(resp.FailedUserIds, newUserId)
				continue
			}

			resp.SuccessCount++
			continue
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("查询群成员失败: " + err.Error())
		}

		// 不存在记录：创建新成员
		groupMember := &models.GroupMember{
			GroupId:   groupId,
			UserId:    newUserId,
			Role:      3,
			IsQuit:    0,
			IsMute:    0,
			JoinTime:  time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := s.DB.WithContext(ctx).Create(groupMember).Error; err != nil {
			resp.FailedCount++
			resp.FailedUserIds = append(resp.FailedUserIds, newUserId)
			continue
		}

		resp.SuccessCount++
	}
	if resp.SuccessCount > 0 {
		if err := s.DB.WithContext(ctx).
			Model(&models.Group{}).
			Where("id = ?", groupId).
			Update("member_count", gorm.Expr("member_count + ?", resp.SuccessCount)).Error; err != nil {
			return nil, errors.New("更新群成员数量失败: " + err.Error())
		}
	}
	return resp, nil
}

// 踢出成员
func (s *GroupMemberService) KickMember(ctx context.Context, GroupId int, KickedUserIds int, userId int) error {
	//1、群存在且操作者在群内且未退出且是管理员或群主
	var member models.GroupMember
	if err := s.DB.WithContext(ctx).
		Where("group_id = ? AND user_id = ? AND is_quit = 0 AND role IN (1,2)", GroupId, userId).
		First(&member).
		Error; err != nil {
		return errors.New("操作者不在群内或无权限")
	}

	//2、被踢成员在群内且未退出
	var kickedMember models.GroupMember
	if err := s.DB.WithContext(ctx).
		Where("group_id = ? AND user_id = ? AND is_quit = 0", GroupId, KickedUserIds).
		First(&kickedMember).Error; err != nil {
		return errors.New("被踢成员不在群内或已退出")
	}

	//3、执行踢出操作
	if err := s.DB.WithContext(ctx).
		Model(&models.GroupMember{}).
		Where("group_id = ? AND user_id = ?", GroupId, KickedUserIds).
		Update("is_quit", 1).Error; err != nil {
		return errors.New("踢出成员失败: " + err.Error())
	}

	//4、更新群成员数量
	if err := s.DB.WithContext(ctx).
		Model(&models.Group{}).
		Where("id = ?", GroupId).
		Update("member_count", gorm.Expr("member_count - ?", 1)).Error; err != nil {
		return errors.New("更新群成员数量失败: " + err.Error())
	}
	return nil
}
