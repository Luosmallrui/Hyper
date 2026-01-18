package service

import (
	"Hyper/dao"
	"Hyper/dao/cache"
	"Hyper/models"
	"Hyper/pkg/response"
	"Hyper/types"
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var _ IGroupMemberService = (*GroupMemberService)(nil)

type IGroupMemberService interface {
	InviteMembers(ctx context.Context, groupId int, InvitedUsersIds []int, userId int) (*types.InviteMemberResponse, error)
	KickMember(ctx context.Context, GroupId int, KickedUserIds int, userId int) error
	ListMembers(ctx context.Context, groupId int, userId int) ([]types.GroupMemberItemDTO, error)
}

type GroupMemberService struct {
	GroupMemberDAO  *dao.GroupMember
	DB              *gorm.DB
	Redis           *redis.Client
	GroupMemberRepo *dao.GroupMember
	GroupRepo       *dao.Group
	Relation        *cache.Relation
}

func (s *GroupMemberService) InviteMembers(ctx context.Context, groupId int, InvitedUsersIds []int, userId int) (*types.InviteMemberResponse, error) {
	var operator models.GroupMember
	if err := s.DB.WithContext(ctx).
		Where("group_id = ? AND user_id = ? AND is_quit = 0 AND role IN (1,2)", groupId, userId).
		First(&operator).Error; err != nil {
		return nil, errors.New("操作者不在群内或无权限邀请")
	}

	_, err := s.GroupRepo.GetGroup(ctx, groupId)
	if err != nil {
		return nil, errors.New("群不存在")
	}

	var existingMembers []models.GroupMember
	s.DB.WithContext(ctx).
		Where("group_id = ? AND user_id IN ?", groupId, InvitedUsersIds).
		Find(&existingMembers)

	memberMap := make(map[int]models.GroupMember)
	for _, m := range existingMembers {
		memberMap[m.UserId] = m
	}

	resp := &types.InviteMemberResponse{
		FailedUserIds: []int{},
	}
	actualSuccessIds := make([]int, 0)

	// 3. 开启事务
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, invUid := range InvitedUsersIds {
			if invUid == userId {
				continue
			} // 不能邀请自己

			if gm, ok := memberMap[invUid]; ok {
				if gm.IsQuit == 0 {
					resp.FailedCount++
					resp.FailedUserIds = append(resp.FailedUserIds, invUid)
					continue
				}
				// 情况 A：已退群，执行恢复
				if err := tx.Model(&models.GroupMember{}).
					Where("id = ?", gm.Id).
					Updates(map[string]interface{}{
						"is_quit":   0,
						"is_mute":   0,
						"join_time": time.Now(),
					}).Error; err != nil {
					return err
				}
			} else {
				// 情况 B：全新成员，执行创建
				newmember := models.GroupMember{
					GroupId:   groupId,
					UserId:    invUid,
					Role:      3,
					IsQuit:    0,
					IsMute:    0,
					JoinTime:  time.Now(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				if err := tx.Create(&newmember).Error; err != nil {
					return err
				}
			}
			resp.SuccessCount++
			actualSuccessIds = append(actualSuccessIds, invUid)
		}
		if resp.SuccessCount > 0 {
			result := tx.Model(&models.Group{}).
				Where("id = ? AND member_count + ? <= max_members", groupId, resp.SuccessCount).
				Update("member_count", gorm.Expr("member_count + ?", resp.SuccessCount))

			if result.RowsAffected == 0 {
				return errors.New("群人数已达上限，邀请失败")
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// 踢出成员
func (s *GroupMemberService) KickMember(ctx context.Context, GroupId int, KickedUserId int, userId int) error {
	var members []models.GroupMember
	if err := s.DB.WithContext(ctx).
		Where("group_id = ? AND user_id IN ?", GroupId, []int{KickedUserId, userId}).
		Find(&members).Error; err != nil {
		return errors.New("查询成员失败: " + err.Error())
	}

	var operator, target *models.GroupMember
	for i := range members {
		if members[i].UserId == userId {
			operator = &members[i]
		}
		if members[i].UserId == KickedUserId {
			target = &members[i]
		}
	}

	if operator == nil || operator.IsQuit == 1 || operator.Role > 2 {
		return errors.New("操作者不在群内或无权限踢出成员")
	}
	if target == nil || target.IsQuit == 1 {
		return errors.New("被踢成员不在群内")
	}
	//以前逻辑错误
	if operator.Role >= target.Role {
		return errors.New("无权限踢出该成员")
	}
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.GroupMember{}).
			Where("group_id = ? AND user_id = ?", GroupId, KickedUserId).
			Update("is_quit", 1)
		if result.Error != nil {
			return errors.New("踢出成员失败: " + result.Error.Error())
		}
		if result.RowsAffected == 0 {
			return errors.New("状态已经变更请勿重复操作")
		}
		if err := tx.Model(&models.Group{}).
			Where("id = ? AND member_count - 1 >= 0", GroupId).
			Update("member_count", gorm.Expr("member_count - 1")).Error; err != nil {
			return errors.New("更新群成员数失败: " + err.Error())
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *GroupMemberService) ListMembers(ctx context.Context, groupId int, userId int) ([]types.GroupMemberItemDTO, error) {
	// 1) 权限：必须是群成员
	if !s.GroupMemberDAO.IsMember(ctx, groupId, userId, true) {
		return nil, response.NewError(403, "你不在群内或已退出")
	}

	// 2) DAO 拿到 models.MemberItem 列表
	items := s.GroupMemberDAO.GetMembers(ctx, groupId)

	// 3) 映射成 DTO
	dtos := make([]types.GroupMemberItemDTO, 0, len(items))
	for _, it := range items {
		if it == nil {
			continue
		}
		dtos = append(dtos, types.GroupMemberItemDTO{
			Role:     it.Role,
			UserCard: it.UserCard,
			UserId:   it.UserId,
			IsMute:   it.IsMute,
			Avatar:   it.Avatar,
			Nickname: it.Nickname,
			Gender:   it.Gender,
			Motto:    it.Motto,
		})
	}
	return dtos, nil
}
