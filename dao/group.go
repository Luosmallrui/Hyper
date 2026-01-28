package dao

import (
	"Hyper/models"
	"context"
	"errors"

	"gorm.io/gorm"
)

type Group struct {
	Repo[models.Group]
}

func NewGroup(db *gorm.DB) *Group {
	return &Group{Repo: NewRepo[models.Group](db)}
}

func (g *Group) FindByID(ctx context.Context, gid int) (*models.Group, error) {
	var group models.Group
	err := g.Db.WithContext(ctx).Where("id = ?", gid).First(&group).Error
	return &group, err
}

func (g *Group) SetMuteAll(ctx context.Context, gid int, mute int) error {
	return g.Db.WithContext(ctx).Model(&models.Group{}).
		Where("id = ?", gid).
		Update("is_mute_all", mute).Error
}

func (g *Group) GetGroup(ctx context.Context, groupId int) (*models.Group, error) {
	var group models.Group
	if err := g.Db.WithContext(ctx).Where("id = ?", groupId).First(&group).Error; err != nil {
		return nil, errors.New("获取群消息失败: " + err.Error())
	}
	return &group, nil
}

func (g *Group) UpdateOwnerId(ctx context.Context, groupId int, newOwnerId int) error {
	return g.Db.WithContext(ctx).
		Model(&models.Group{}).
		Where("id = ?", groupId).
		Update("owner_id", newOwnerId).Error
}
