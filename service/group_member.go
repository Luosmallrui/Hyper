package service

import (
	"Hyper/dao"
	"Hyper/models"
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var _ IGroupMemberService = (*GroupMemberService)(nil)

type IGroupMemberService interface {
	Handover(ctx context.Context, groupId int, userId int, memberId int) error
	SetLeaderStatus(ctx context.Context, groupId int, userId int, leader int) error
	SetMuteStatus(ctx context.Context, groupId int, userId int, status int) error
}

type GroupMemberService struct {
	Db              *gorm.DB
	Redis           *redis.Client
	GroupMemberRepo *dao.GroupMember
}

func (g *GroupMemberService) Handover(ctx context.Context, groupId int, userId int, memberId int) error {
	return g.Db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		err := tx.Model(&models.GroupMember{}).Where("group_id = ? and user_id = ? and leader = ?", groupId, userId, models.GroupMemberLeaderOwner).Update("leader", models.GroupMemberLeaderOrdinary).Error
		if err != nil {
			return err
		}

		err = tx.Model(&models.GroupMember{}).Where("group_id = ? and user_id = ?", groupId, memberId).Update("leader", models.GroupMemberLeaderOwner).Error
		if err != nil {
			return err
		}

		return nil
	})
}

func (g *GroupMemberService) SetLeaderStatus(ctx context.Context, groupId int, userId int, leader int) error {
	return g.GroupMemberRepo.Model(ctx).Where("group_id = ? and user_id = ?", groupId, userId).UpdateColumn("leader", leader).Error
}

func (g *GroupMemberService) SetMuteStatus(ctx context.Context, groupId int, userId int, status int) error {
	return g.GroupMemberRepo.Model(ctx).Where("group_id = ? and user_id = ?", groupId, userId).UpdateColumn("is_mute", status).Error
}
