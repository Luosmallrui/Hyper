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
	QuitGroup(ctx context.Context, groupId int, userId int) (*types.QuitGroupResponse, error)
	MuteMember(ctx context.Context, groupId int, operatorId int, targetUserId int, mute bool) error
	SetMuteAll(ctx context.Context, groupId int, operatorId int, mute bool) (*types.MuteAllResponse, error)
}

type GroupMemberService struct {
	GroupMemberDAO  *dao.GroupMember
	DB              *gorm.DB
	Redis           *redis.Client
	GroupMemberRepo *dao.GroupMember
	GroupRepo       *dao.Group
	Relation        *cache.Relation
	DB             *gorm.DB
	GroupMemberDAO *dao.GroupMember
	SessionDAO     *dao.SessionDAO
	UnreadStorage  *cache.UnreadStorage
	GroupDAO       *dao.GroupDAO
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
func (s *GroupMemberService) QuitGroup(ctx context.Context, groupId int, userId int) (*types.QuitGroupResponse, error) {
	// 1) 先查成员记录，判断是否在群里
	member, err := s.GroupMemberDAO.FindByUserId(ctx, groupId, userId)
	if err != nil {
		return nil, errors.New("你不在该群，无法退群")
	}
	if member.IsQuit == 1 {
		return nil, errors.New("你已经退过该群了")
	}

	// 2) 开事务处理：退群/解散 + 清会话 + 清未读
	resp := &types.QuitGroupResponse{Disbanded: false}

	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// 群主退群 => 解散群
		if member.Role == models.GroupMemberLeaderOwner { // 1=群主 :contentReference[oaicite:10]{index=10}
			resp.Disbanded = true

			// 2.1 先拿到所有未退群成员（用于清未读/删会话）
			memberIDs, err := s.GroupMemberDAO.GetMemberIds(ctx, groupId)
			if err != nil {
				return err
			}

			// 2.2 全员标记退群
			if err := tx.Table("group_member").
				Where("group_id = ? AND is_quit = 0", groupId).
				Updates(map[string]any{
					"is_quit":    1,
					"updated_at": time.Now(),
				}).Error; err != nil {
				return err
			}

			// 2.3 删除群（表结构没软删字段，这里用硬删）
			if err := tx.Table("groups").Where("id = ?", groupId).Delete(nil).Error; err != nil {
				return err
			}

			// 2.4 删除所有人的群会话
			if err := s.SessionDAO.DeleteSessionsByPeer(ctx, 2, uint64(groupId)); err != nil {
				return err
			}

			// 2.5 清所有人该群的未读（Redis）
			for _, uid := range memberIDs {
				s.UnreadStorage.Reset(ctx, uid, 2, groupId)
				// 同时建议删关系缓存（否则 IsMember cache 命中会误判）
				s.GroupMemberDAO.ClearGroupRelation(ctx, uid, groupId)
			}

			return nil
		}

		// 普通成员退群
		if err := tx.Table("group_member").
			Where("group_id = ? AND user_id = ? AND is_quit = 0", groupId, userId).
			Updates(map[string]any{
				"is_quit":    1,
				"updated_at": time.Now(),
			}).Error; err != nil {
			return err
		}

		// member_count - 1
		_ = tx.Table("groups").
			Where("id = ? AND member_count > 0", groupId).
			UpdateColumn("member_count", gorm.Expr("member_count - 1")).Error

		// 删除该用户的群会话
		if err := s.SessionDAO.DeleteSession(ctx, uint64(userId), 2, uint64(groupId)); err != nil {
			return err
		}

		// 清该用户该群未读
		s.UnreadStorage.Reset(ctx, userId, 2, groupId)

		// 删关系缓存，避免缓存命中仍被当成员
		s.GroupMemberDAO.ClearGroupRelation(ctx, userId, groupId)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// 个人禁言
func (s *GroupMemberService) MuteMember(ctx context.Context, gid int, operatorId int, targetId int, mute bool) error {
	// 1) 操作者必须在群里且未退群
	op, err := s.GroupMemberDAO.FindByUserId(ctx, gid, operatorId)
	if err != nil || op.IsQuit == 1 {
		return errors.New("你不在群内或已退群")
	}

	// 2) 目标必须在群里且未退群
	target, err := s.GroupMemberDAO.FindByUserId(ctx, gid, targetId)
	if err != nil || target.IsQuit == 1 {
		return errors.New("对方不在群内或已退群")
	}

	// 3) 权限判断（只允许群主/管理员操作）
	if op.Role != 1 && op.Role != 2 {
		return errors.New("无权限操作")
	}

	// 4) 不能禁言群主
	if target.Role == 1 {
		return errors.New("不能禁言群主")
	}

	// 5) 管理员不能禁言管理员
	if op.Role == 2 && target.Role == 2 {
		return errors.New("管理员不能禁言其他管理员")
	}

	val := 0
	if mute {
		val = 1
	}

	return s.GroupMemberDAO.SetMemberMute(ctx, gid, targetId, val)
}

// 群禁言开关
func (s *GroupMemberService) SetMuteAll(ctx context.Context, gid int, operatorId int, mute bool) (*types.MuteAllResponse, error) {
	op, err := s.GroupMemberDAO.FindByUserId(ctx, gid, operatorId)
	if err != nil || op.IsQuit == 1 {
		return nil, errors.New("你不在群内或已退群")
	}
	if op.Role != 1 && op.Role != 2 {
		return nil, errors.New("无权限操作")
	}

	val := 0
	if mute {
		val = 1
	}
	if err := s.GroupDAO.SetMuteAll(ctx, gid, val); err != nil {
		return nil, err
	}
	return &types.MuteAllResponse{IsMuteAll: mute}, nil
}
