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
	CreateGroup(ctx context.Context, req *types.CreateGroupRequest, userId int) (*types.Group, error)
	DismissGroup(ctx context.Context, groupId int, userId int) error
	//UnMuteAllMembers(ctx context.Context, groupId int, userId int) error
	//MuteAllMembers(ctx context.Context, userId int, groupId int) error
	UpdateGroupName(ctx context.Context, groupId int, userId int, req *types.UpdateGroupNameRequest) error
	UpdateGroupAvatar(ctx context.Context, groupId int, userId int, req *types.UpdateGroupAvatarRequest) error
	UpdateGroupDescription(ctx context.Context, groupId int, userId int, req *types.UpdateGroupDescriptionRequest) error
}

var _ IGroupService = (*GroupService)(nil)

type GroupService struct {
	DB             *gorm.DB
	GroupMemberDAO *dao.GroupMember
	GroupDAO       *dao.Group
	Relation       *cache.Relation
	SessionDAO     *dao.SessionDAO
	UnreadStorage  *cache.UnreadStorage
}

// 创建群
func (s *GroupService) CreateGroup(ctx context.Context, req *types.CreateGroupRequest, userId int) (*types.Group, error) {
	if s.SessionDAO == nil {
		return nil, errors.New("SessionDAO 未初始化")
	}
	var groupModel models.Group // 只用于落库
	var resp types.Group        // 返回给上层（包含 SessionId）
	var groupID int
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		groupModel = models.Group{
			Name:        req.Name,
			Avatar:      req.Avatar,
			Description: req.Description,
			OwnerId:     userId,
			MemberCount: 1,
			IsDismiss:   0,
			MaxMembers:  200,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// 1) 创建群组记录（只落 models.Group）
		if err := tx.Create(&groupModel).Error; err != nil {
			return err
		}
		groupID = groupModel.Id

		// 2) 插入群主成员
		groupMember := &models.GroupMember{
			GroupId:   groupModel.Id,
			UserId:    userId,
			Role:      1,
			IsQuit:    0,
			IsMute:    0,
			JoinTime:  time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := tx.Create(groupMember).Error; err != nil {
			return err
		}

		// 3) 确保群会话存在（幂等）
		sid, err := s.SessionDAO.EnsureSession(ctx, tx, uint64(userId), 2, uint64(groupModel.Id))
		if err != nil {
			return err
		}
		resp.SessionId = int(sid)

		// 4) 初始化会话展示文案（避免 last_msg_content 为空）
		if err := tx.Model(&models.Session{}).
			Where("id = ?", sid).
			Updates(map[string]any{
				"last_msg_type":    1,
				"last_msg_content": "创建了群聊",
				"last_msg_time":    time.Now().UnixMilli(),
				"updated_at":       time.Now(),
			}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, errors.New("创建群聊失败: " + err.Error())
	}
	// DB 已提交后再写缓存：缓存失败不影响建群成功
	if s.Relation != nil {
		s.Relation.SetGroupRelation(ctx, userId, groupID)
	}

	// 组装返回 DTO
	resp.Group = groupModel
	return &resp, nil
}

func (s *GroupService) DismissGroup(ctx context.Context, groupId int, userId int) error {
	group, err := s.GroupDAO.GetGroup(ctx, groupId)
	if err != nil || group.OwnerId != userId {
		return errors.New("只有群主才能解散群")
	}
	if group.IsDismiss == 1 {
		return errors.New("群已解散")
	}
	var memberIDs []int
	// 2. 开启事务
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 2.1 再次用 tx 确认群仍未解散（并发下避免重复解散）  [改动]
		var g models.Group
		if err := tx.
			Where("id = ? AND is_dismiss = 0", groupId).
			First(&g).Error; err != nil {
			return errors.New("群不存在或已解散")
		}

		// 2.2 先查出所有未退群成员（必须在更新 is_quit 之前）
		if err := tx.Model(&models.GroupMember{}).
			Where("group_id = ? AND is_quit = 0", groupId).
			Pluck("user_id", &memberIDs).Error; err != nil {
			return err
		}
		// 2.3 标记群解散（带条件，保证幂等/并发安全）
		res := tx.Model(&models.Group{}).
			Where("id = ? AND is_dismiss = 0", groupId).
			Update("is_dismiss", 1)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("群不存在或已解散")
		}

		// 2.4 全员标记退群
		if err := tx.Model(&models.GroupMember{}).
			Where("group_id = ? AND is_quit = 0", groupId).
			Updates(map[string]any{
				"is_quit":    1,
				"updated_at": time.Now(),
			}).Error; err != nil {
			return errors.New("移除群成员失败: " + err.Error())
		}
		if err := tx.Model(&models.Group{}).
			Where("id = ?", groupId).
			Update("member_count", 0).Error; err != nil {
			return err
		}

		// 2.5 删除所有人的群会话（DB）放进事务，保证列表立刻不再出现
		if s.SessionDAO != nil {
			if err := s.SessionDAO.WithDB(tx).DeleteSessionsByPeer(ctx, 2, uint64(groupId)); err != nil {
				return errors.New("删除群会话失败: " + err.Error())
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	// 3) 事务提交后：Redis/缓存 best-effort（可异步/可超时 ctx）
	go func(ids []int) {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// 3.1 关系缓存
		if s.Relation != nil && len(ids) > 0 {
			s.Relation.BatchDelGroupRelation(cleanupCtx, ids, groupId)
		}

		// 3.2 Redis unread：DB 权威未读，Redis 仅清历史残留 key
		if s.UnreadStorage != nil && len(ids) > 0 {
			for _, uid := range ids {
				s.UnreadStorage.Reset(cleanupCtx, uid, 2, groupId)
			}
		}
	}(append([]int(nil), memberIDs...))

	return nil
}

//func (s *GroupService) MuteAllMembers(ctx context.Context, userId int, groupId int) error {
//	group, err := s.GroupDAO.GetGroup(ctx, groupId)
//	if err != nil {
//		return errors.New("群组不存在")
//	}
//	if group.IsMuteAll == 1 {
//		return errors.New("群已处于全员禁言状态")
//	}
//	if group.OwnerId != userId {
//		var member models.GroupMember
//		err := s.DB.WithContext(ctx).
//			Where("group_id = ? AND user_id = ? AND is_quit = 0", groupId, userId).
//			First(&member).Error
//		if err != nil {
//			return errors.New("用户不是群成员")
//		}
//		if member.Role > 2 {
//			return errors.New("只有群主和管理员可以设置全员禁言")
//		}
//	}
//	err = s.DB.WithContext(ctx).
//		Model(&models.Group{}).
//		Where("id = ?", groupId).
//		Update("is_mute_all", 1).Error
//	if err != nil {
//		return errors.New("设置全员禁言失败: " + err.Error())
//	}
//	return nil
//}

//func (s *GroupService) UnMuteAllMembers(ctx context.Context, groupId int, userId int) error {
//	group, err := s.GroupDAO.GetGroup(ctx, groupId)
//	if err != nil {
//		return errors.New("群组不存在")
//	}
//	if group.IsMuteAll == 0 {
//		return errors.New("群未处于全员禁言状态")
//	}
//
//	if group.OwnerId != userId {
//		var member models.GroupMember
//		err := s.DB.WithContext(ctx).
//			Where("group_id = ? AND user_id = ? AND is_quit = 0", groupId, userId).
//			First(&member).Error
//		if err != nil {
//			return errors.New("用户不是群成员")
//		}
//		if member.Role > 2 {
//			return errors.New("只有群主和管理员可以取消全员禁言")
//		}
//	}
//	err = s.DB.WithContext(ctx).
//		Model(&models.Group{}).
//		Where("id = ?", groupId).
//		Update("is_mute_all", 0).Error
//	if err != nil {
//		return errors.New("取消全员禁言失败: " + err.Error())
//	}
//	return nil
//}

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
