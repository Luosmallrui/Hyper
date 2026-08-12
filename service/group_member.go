package service

import (
	"Hyper/dao"
	"Hyper/dao/cache"
	"Hyper/models"
	"Hyper/pkg/response"
	"Hyper/types"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var _ IGroupMemberService = (*GroupMemberService)(nil)

type IGroupMemberService interface {
	InviteMembers(ctx context.Context, groupId int, InvitedUsersIds []int, userId int) (*types.InviteMemberResponse, error)
	FindInviteCandidate(ctx context.Context, groupID, operatorID int, mobile string) (*types.GroupInviteCandidateResponse, error)
	KickMember(ctx context.Context, GroupId int, KickedUserIds int, userId int) error
	ListMembers(ctx context.Context, groupId int, userId int) (*types.GroupMemberListResponse, error)
	QuitGroup(ctx context.Context, groupId int, userId int) (*types.QuitGroupResponse, error)
	MuteMember(ctx context.Context, groupId int, operatorId int, targetUserId int, mute bool) error
	SetMuteAll(ctx context.Context, groupId int, operatorId int, mute bool) (*types.MuteAllResponse, error)
	SetAdmin(ctx context.Context, groupId int, operatorId int, targetUserId int, admin bool) error
	TransferOwner(ctx context.Context, groupId int, operatorId int, newOwnerId int) (*types.TransferOwnerResponse, error)
}

type GroupMemberService struct {
	Redis          *redis.Client
	GroupRepo      *dao.Group
	Relation       *cache.Relation
	DB             *gorm.DB
	GroupMemberDAO *dao.GroupMember
	SessionDAO     *dao.SessionDAO
	UnreadStorage  *cache.UnreadStorage
}

func (s *GroupMemberService) ensureGroupActive(ctx context.Context, gid int) error {
	if s.GroupRepo == nil {
		return errors.New("GroupRepo 未初始化")
	}
	_, err := s.GroupRepo.GetGroup(ctx, gid)
	return err
}
func (s *GroupMemberService) InviteMembers(ctx context.Context, groupId int, InvitedUsersIds []int, userId int) (*types.InviteMemberResponse, error) {
	if s.SessionDAO == nil {
		return nil, errors.New("SessionDAO 未初始化")
	}
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
	// The public invite API accepts selected user IDs. Ensure every requested
	// account is a normal registered user, even when a caller bypasses the
	// mobile lookup UI and calls this endpoint directly.
	validUserIDs := make(map[int]struct{}, len(InvitedUsersIds))
	var users []models.Users
	if err := s.DB.WithContext(ctx).
		Where("id IN ? AND status = ?", InvitedUsersIds, 1).
		Find(&users).Error; err != nil {
		return nil, errors.New("查询邀请用户失败: " + err.Error())
	}
	for _, user := range users {
		validUserIDs[user.Id] = struct{}{}
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
	processedUserIDs := make(map[int]struct{}, len(InvitedUsersIds))

	// 3. 开启事务
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, invUid := range InvitedUsersIds {
			if _, seen := processedUserIDs[invUid]; seen {
				resp.FailedCount++
				resp.FailedUserIds = append(resp.FailedUserIds, invUid)
				continue
			}
			processedUserIDs[invUid] = struct{}{}
			if invUid <= 0 || invUid == userId {
				resp.FailedCount++
				resp.FailedUserIds = append(resp.FailedUserIds, invUid)
				continue
			}
			if _, ok := validUserIDs[invUid]; !ok {
				resp.FailedCount++
				resp.FailedUserIds = append(resp.FailedUserIds, invUid)
				continue
			}
			// 是否成功加入（新加/恢复）用一个标记
			joined := false
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
				joined = true
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
				joined = true
			}
			if joined {
				//为新加入的人创建群会话（幂等）
				if _, err := s.SessionDAO.EnsureSession(ctx, tx, uint64(invUid), 2, uint64(groupId)); err != nil {
					return err
				}

				resp.SuccessCount++
				actualSuccessIds = append(actualSuccessIds, invUid)
			}
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

	// 事务成功提交后再写缓存：缓存失败不影响邀请成功
	if s.Relation != nil && len(actualSuccessIds) > 0 {
		for _, uid := range actualSuccessIds {
			s.Relation.SetGroupRelation(ctx, uid, groupId)
		}
	}

	return resp, nil
}

// FindInviteCandidate resolves an exact mobile number for a group owner or
// admin. It returns only a masked number and requires group management rights
// before performing the lookup, so it cannot be used as a public phone search.
func (s *GroupMemberService) FindInviteCandidate(ctx context.Context, groupID, operatorID int, mobile string) (*types.GroupInviteCandidateResponse, error) {
	group, err := s.GroupRepo.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	var operator models.GroupMember
	if err := s.DB.WithContext(ctx).
		Where("group_id = ? AND user_id = ? AND is_quit = 0 AND role IN ?", groupID, operatorID, []int{
			models.GroupMemberLeaderOwner,
			models.GroupMemberLeaderAdmin,
		}).
		First(&operator).Error; err != nil {
		return nil, errors.New("操作者不在群内或无权限邀请")
	}

	var user models.Users
	if err := s.DB.WithContext(ctx).
		Where("mobile = ? AND status = ?", mobile, 1).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &types.GroupInviteCandidateResponse{Found: false}, nil
		}
		return nil, errors.New("查询用户失败: " + err.Error())
	}

	candidate := &types.GroupInviteCandidate{
		UserID:       user.Id,
		MobileMasked: maskGroupMemberMobile(user.Mobile),
		Nickname:     user.Nickname,
		Avatar:       user.Avatar,
		Motto:        user.Motto,
		CanInvite:    true,
	}

	var membership models.GroupMember
	err = s.DB.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, user.Id).
		First(&membership).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		candidate.MembershipStatus = "not_member"
	case err != nil:
		return nil, errors.New("查询群成员状态失败: " + err.Error())
	case membership.IsQuit == 1:
		candidate.MembershipStatus = "left"
	default:
		candidate.MembershipStatus = "active"
		candidate.CanInvite = false
		candidate.InviteDisabledReason = "用户已在群内"
	}

	if candidate.CanInvite && s.GroupMemberDAO.CountMemberTotal(ctx, groupID) >= int64(group.MaxMembers) {
		candidate.CanInvite = false
		candidate.InviteDisabledReason = "群人数已达上限"
	}

	return &types.GroupInviteCandidateResponse{
		Found:     true,
		Candidate: candidate,
	}, nil
}

func maskGroupMemberMobile(mobile string) string {
	if len(mobile) != 11 {
		return ""
	}
	return mobile[:3] + "****" + mobile[7:]
}

// 踢出成员
func (s *GroupMemberService) KickMember(ctx context.Context, GroupId int, KickedUserId int, userId int) error {
	if err := s.ensureGroupActive(ctx, GroupId); err != nil {
		return err
	}
	if s.SessionDAO == nil {
		return errors.New("SessionDAO 未初始化")
	}
	var members []models.GroupMember
	if err := s.DB.WithContext(ctx).
		Where("group_id = ? AND user_id IN ? AND is_quit = 0", GroupId, []int{KickedUserId, userId}).
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
	kicked := false
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.GroupMember{}).
			Where("group_id = ? AND user_id = ? AND is_quit = 0", GroupId, KickedUserId).
			Updates(map[string]any{
				"is_quit":    1,
				"updated_at": time.Now(),
			})
		if result.Error != nil {
			return errors.New("踢出成员失败: " + result.Error.Error())
		}
		if result.RowsAffected == 0 {
			return errors.New("状态已经变更请勿重复操作")
		}
		kicked = true // 事务内标记成功（事务成功提交后再删缓存）
		if err := tx.Model(&models.Group{}).
			Where("id = ? AND is_dismiss = 0 AND member_count > 0", GroupId).
			Update("member_count", gorm.Expr("member_count - 1")).Error; err != nil {
			return errors.New("更新群成员数失败: " + err.Error())
		}
		if err := s.SessionDAO.WithDB(tx).
			DeleteSession(ctx, uint64(KickedUserId), 2, uint64(GroupId)); err != nil {
			return errors.New("删除被踢用户会话失败: " + err.Error())
		}
		return nil
	})
	if err != nil {
		return err
	}
	// 事务成功后再做缓存清理
	if kicked {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		s.GroupMemberDAO.ClearGroupRelation(cleanupCtx, KickedUserId, GroupId)
		// 清理被踢用户的群未读（Redis）
		if s.UnreadStorage != nil {
			s.UnreadStorage.Reset(cleanupCtx, KickedUserId, 2, GroupId)
		}
	}

	return nil
}

func (s *GroupMemberService) ListMembers(ctx context.Context, groupID int, userID int) (*types.GroupMemberListResponse, error) {
	group, err := s.GroupRepo.GetGroup(ctx, groupID)
	if err != nil {
		return nil, response.NewError(404, "群不存在或已解散")
	}
	if s.GroupMemberDAO == nil {
		return nil, errors.New("群成员服务未初始化")
	}
	// 1) 权限：必须是群成员
	// 权限判断不能依赖 Redis 缓存，避免缓存过期或残留导致误判。
	if !s.GroupMemberDAO.IsMember(ctx, groupID, userID, false) {
		return nil, response.NewError(403, "你不在群内或已退出")
	}

	// 2) DAO 拿到 models.MemberItem 列表
	items, err := s.GroupMemberDAO.GetMembers(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("查询群成员失败: %w", err)
	}

	// 3) 映射成员信息，同时按当前登录成员的角色预计算操作权限。
	dtos := make([]types.GroupMemberItemDTO, 0, len(items))
	avatarList := make([]string, 0, len(items))
	viewerRole := 0
	viewerUserCard := ""
	viewerDisplayName := ""
	viewerMuted := false
	for _, it := range items {
		if it == nil {
			continue
		}
		displayName := groupMemberDisplayName(it.UserCard, it.Nickname, it.UserId)
		isCurrentUser := it.UserId == userID
		if isCurrentUser {
			viewerRole = it.Role
			viewerUserCard = it.UserCard
			viewerDisplayName = displayName
			viewerMuted = it.IsMute == 1
		}
		if it.Avatar != "" && len(avatarList) < 9 {
			avatarList = append(avatarList, it.Avatar)
		}
		dtos = append(dtos, types.GroupMemberItemDTO{
			UserId:         it.UserId,
			Avatar:         it.Avatar,
			Nickname:       it.Nickname,
			DisplayName:    displayName,
			Gender:         it.Gender,
			Motto:          it.Motto,
			Role:           it.Role,
			RoleName:       groupMemberRoleName(it.Role),
			IsMute:         it.IsMute,
			IsMuted:        it.IsMute == 1,
			CanSendMessage: groupMemberCanSend(it.Role, it.IsMute == 1, group.IsMuteAll == 1),
			UserCard:       it.UserCard,
			JoinTime:       formatGroupMemberTime(it.JoinTime),
			IsCurrentUser:  isCurrentUser,
		})
	}
	if viewerRole == 0 {
		return nil, response.NewError(403, "你不在群内或已退出")
	}
	permissions := groupMemberPermissions(viewerRole)
	for i := range dtos {
		dtos[i].CanKick = canManageGroupMember(viewerRole, dtos[i].Role, dtos[i].UserId == userID)
		dtos[i].CanMute = canManageGroupMember(viewerRole, dtos[i].Role, dtos[i].UserId == userID)
		dtos[i].CanSetAdmin = viewerRole == models.GroupMemberLeaderOwner && dtos[i].UserId != userID && dtos[i].Role != models.GroupMemberLeaderOwner
		dtos[i].CanTransfer = viewerRole == models.GroupMemberLeaderOwner && dtos[i].UserId != userID
	}

	return &types.GroupMemberListResponse{
		Group: types.GroupMemberGroupInfo{
			ID:               group.Id,
			Name:             group.Name,
			Avatar:           group.Avatar,
			Description:      group.Description,
			OwnerID:          group.OwnerId,
			MemberCount:      len(dtos),
			MaxMembers:       group.MaxMembers,
			IsMuteAll:        group.IsMuteAll == 1,
			CreatedAt:        formatGroupMemberTime(group.CreatedAt),
			MemberAvatarList: avatarList,
		},
		CurrentUser: types.GroupCurrentMemberInfo{
			UserID:         userID,
			Role:           viewerRole,
			RoleName:       groupMemberRoleName(viewerRole),
			UserCard:       viewerUserCard,
			DisplayName:    viewerDisplayName,
			IsOwner:        viewerRole == models.GroupMemberLeaderOwner,
			IsAdmin:        viewerRole == models.GroupMemberLeaderAdmin,
			IsMuted:        viewerMuted,
			CanSendMessage: groupMemberCanSend(viewerRole, viewerMuted, group.IsMuteAll == 1),
			Permissions:    permissions,
		},
		Members: dtos,
	}, nil
}

func groupMemberRoleName(role int) string {
	switch role {
	case models.GroupMemberLeaderOwner:
		return "群主"
	case models.GroupMemberLeaderAdmin:
		return "管理员"
	default:
		return "群成员"
	}
}

func groupMemberPermissions(role int) types.GroupMemberPermissions {
	isManager := role == models.GroupMemberLeaderOwner || role == models.GroupMemberLeaderAdmin
	isOwner := role == models.GroupMemberLeaderOwner
	return types.GroupMemberPermissions{
		CanInvite:          isManager,
		CanManageMembers:   isManager,
		CanMuteMembers:     isManager,
		CanMuteAll:         isManager,
		CanSetAdmin:        isOwner,
		CanTransferOwner:   isOwner,
		CanUpdateGroupInfo: isOwner,
		CanDismissGroup:    isOwner,
		CanQuit:            true,
	}
}

func groupMemberDisplayName(userCard, nickname string, userID int) string {
	if userCard != "" {
		return userCard
	}
	if nickname != "" {
		return nickname
	}
	return fmt.Sprintf("用户%d", userID)
}

func canManageGroupMember(viewerRole, memberRole int, isCurrentUser bool) bool {
	return !isCurrentUser && viewerRole <= models.GroupMemberLeaderAdmin && viewerRole < memberRole
}

func groupMemberCanSend(role int, isMuted, isMuteAll bool) bool {
	if role == models.GroupMemberLeaderOwner || role == models.GroupMemberLeaderAdmin {
		return true
	}
	return !isMuted && !isMuteAll
}

func formatGroupMemberTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}
func (s *GroupMemberService) QuitGroup(ctx context.Context, groupId int, userId int) (*types.QuitGroupResponse, error) {
	if err := s.ensureGroupActive(ctx, groupId); err != nil {
		return nil, err
	}
	if s.SessionDAO == nil {
		return nil, errors.New("SessionDAO 未初始化")
	}
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
	//  这些变量用于：事务提交后做 Redis 清理（best-effort）
	var (
		memberIDsForCleanup []int // 解散时需要清的所有成员
		needClearUserID     int   // 普通退群需要清的用户
	)
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		//事务内重新查 member，避免并发导致外面查到的 member 过期
		member, err := s.GroupMemberDAO.WithDB(tx).FindByUserId(ctx, groupId, userId)
		if err != nil {
			return errors.New("你不在该群，无法退群")
		}
		if member.IsQuit == 1 {
			return errors.New("你已经退过该群了")
		}
		// 群主退群 => 解散群
		if member.Role == models.GroupMemberLeaderOwner {
			resp.Disbanded = true

			// 2.1 用 tx 读取未退群成员，保证事务内一致性
			ids, err := s.GroupMemberDAO.WithDB(tx).GetMemberIds(ctx, groupId)
			if err != nil {
				return err
			}
			memberIDsForCleanup = ids

			// 2.2 全员标记退群
			if err := tx.Table("group_member").
				Where("group_id = ? AND is_quit = 0", groupId).
				Updates(map[string]any{
					"is_quit":    1,
					"updated_at": time.Now(),
				}).Error; err != nil {
				return err
			}

			// 2.3 软解散群：统一语义 is_dismiss=1
			res := tx.Model(&models.Group{}).
				Where("id = ? AND is_dismiss = 0", groupId).
				Update("is_dismiss", 1)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return errors.New("群不存在或已解散")
			}

			// 2.4 删除所有人的群会话
			if err := s.SessionDAO.WithDB(tx).DeleteSessionsByPeer(ctx, 2, uint64(groupId)); err != nil {
				return err
			}

			return nil
		}

		// 普通成员退群
		res := tx.Table("group_member").
			Where("group_id = ? AND user_id = ? AND is_quit = 0", groupId, userId).
			Updates(map[string]any{
				"is_quit":    1,
				"updated_at": time.Now(),
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("状态已经变更请勿重复操作")
		}
		// member_count - 1
		res = tx.Table("groups").
			Where("id = ? AND is_dismiss = 0 AND member_count > 0", groupId).
			UpdateColumn("member_count", gorm.Expr("member_count - 1"))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("更新群成员数失败")
		}

		// 删除该用户的群会话
		if err := s.SessionDAO.WithDB(tx).DeleteSession(ctx, uint64(userId), 2, uint64(groupId)); err != nil {
			return err
		}
		needClearUserID = userId

		return nil
	})
	if err != nil {
		return nil, err
	}
	// 事务提交后再做 Redis 清理，并且不要用请求 ctx
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if resp.Disbanded {
		for _, uid := range memberIDsForCleanup {
			// Redis unread 仅用于清理历史残留 key（DB 才是权威未读）
			if s.UnreadStorage != nil {
				s.UnreadStorage.Reset(cleanupCtx, uid, 2, groupId)
			}
			s.GroupMemberDAO.ClearGroupRelation(cleanupCtx, uid, groupId)
		}
	} else {
		if needClearUserID != 0 {
			if s.UnreadStorage != nil {
				s.UnreadStorage.Reset(cleanupCtx, needClearUserID, 2, groupId)
			}
			s.GroupMemberDAO.ClearGroupRelation(cleanupCtx, needClearUserID, groupId)
		}
	}

	return resp, nil
}

// 个人禁言
func (s *GroupMemberService) MuteMember(ctx context.Context, gid int, operatorId int, targetId int, mute bool) error {
	if err := s.ensureGroupActive(ctx, gid); err != nil {
		return err
	}
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
	if err := s.ensureGroupActive(ctx, gid); err != nil {
		return nil, err
	}
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
	if err := s.GroupRepo.SetMuteAll(ctx, gid, val); err != nil {
		return nil, err
	}
	return &types.MuteAllResponse{IsMuteAll: mute}, nil
}
func (s *GroupMemberService) SetAdmin(ctx context.Context, groupId int, operatorId int, targetUserId int, admin bool) error {
	if err := s.ensureGroupActive(ctx, groupId); err != nil {
		return err
	}
	// 1) 只有群主可操作
	if !s.GroupMemberDAO.IsMaster(ctx, groupId, operatorId) {
		return errors.New("无权限操作")
	}

	// 2) 查目标成员是否在群内（未退群）
	target, err := s.GroupMemberDAO.FindByUserId(ctx, groupId, targetUserId)
	if err != nil || target.IsQuit == 1 {
		return errors.New("对方不在群内或已退群")
	}

	// 3) 不能操作群主
	if target.Role == models.GroupMemberLeaderOwner {
		return errors.New("不能操作群主")
	}

	// 4) 计算目标角色
	newRole := models.GroupMemberLeaderOrdinary // 3 普通成员
	if admin {
		newRole = models.GroupMemberLeaderAdmin // 2 管理员
	} //admin=true → role=2;admin=false → role=3

	// 5) 幂等：已经是该角色就直接成功
	if int(target.Role) == newRole {
		return nil
	}

	// 6) 更新角色
	return s.GroupMemberDAO.UpdateRole(ctx, groupId, targetUserId, newRole)
}

func (s *GroupMemberService) TransferOwner(
	ctx context.Context,
	groupId int,
	operatorId int,
	newOwnerId int,
) (*types.TransferOwnerResponse, error) {
	if err := s.ensureGroupActive(ctx, groupId); err != nil {
		return nil, err
	}
	if groupId <= 0 || operatorId <= 0 || newOwnerId <= 0 {
		return nil, errors.New("参数错误")
	}
	// 1) 只有群主能转让
	if !s.GroupMemberDAO.IsMaster(ctx, groupId, operatorId) {
		return nil, errors.New("无权限操作")
	}
	if operatorId == newOwnerId {
		return nil, errors.New("不能转让给自己")
	}
	// 2) 新群主必须在群且未退群
	target, err := s.GroupMemberDAO.FindByUserId(ctx, groupId, newOwnerId)
	if err != nil || target.IsQuit == 1 {
		return nil, errors.New("对方不在群内或已退群")
	}

	// 3) 事务：三步必须原子
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		groupDAO := s.GroupRepo
		gmDAO := s.GroupMemberDAO.WithDB(tx)

		// 3.1 更新群主字段
		if err := groupDAO.UpdateOwnerId(ctx, groupId, newOwnerId); err != nil {
			return err
		}
		// 3.2 旧群主降级为普通成员(3)
		if err := gmDAO.UpdateRole(ctx, groupId, operatorId, 3); err != nil {
			return err
		}
		// 3.3 新群主升为群主(1)
		if err := gmDAO.UpdateRole(ctx, groupId, newOwnerId, 1); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &types.TransferOwnerResponse{Success: true}, nil
}
