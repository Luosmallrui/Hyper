package dao

import (
	"Hyper/models"
	"context"
	"errors"

	"gorm.io/gorm"
)

type GroupDAO struct {
	db *gorm.DB
}
type Group struct {
	Repo[models.Group]
}

func NewGroupDAO(db *gorm.DB) *GroupDAO { return &GroupDAO{db: db} }

func (d *GroupDAO) FindByID(ctx context.Context, gid int) (*models.Group, error) {
	var g models.Group
	err := d.db.WithContext(ctx).Where("id = ?", gid).First(&g).Error
	return &g, err
}
func NewGroup(db *gorm.DB) *Group {
	return &Group{Repo: NewRepo[models.Group](db)}
}

func (d *GroupDAO) SetMuteAll(ctx context.Context, gid int, mute int) error {
	return d.db.WithContext(ctx).Model(&models.Group{}).
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

func (d *GroupDAO) WithDB(db *gorm.DB) *GroupDAO {
	nd := *d
	nd.db = db
	return &nd
}

func (d *GroupDAO) UpdateOwnerId(ctx context.Context, groupId int, newOwnerId int) error {
	return d.db.WithContext(ctx).
		Model(&models.Group{}).
		Where("id = ?", groupId).
		Update("owner_id", newOwnerId).Error
}
