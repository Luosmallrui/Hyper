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
	err := g.Db.WithContext(ctx).
		Where("id = ? AND is_dismiss = 0", gid).
		First(&group).Error
	if err != nil {
		return nil, errors.New("群不存在或已解散")
	}
	return &group, nil
}

func (g *Group) SetMuteAll(ctx context.Context, gid int, mute int) error {
	res := g.Db.WithContext(ctx).
		Model(&models.Group{}).
		Where("id = ? AND is_dismiss = 0", gid).
		Update("is_mute_all", mute)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("群不存在或已解散")
	}
	return nil
}
func (g *Group) GetGroup(ctx context.Context, groupId int) (*models.Group, error) {
	var group models.Group
	if err := g.Db.WithContext(ctx).
		Where("id = ? AND is_dismiss = 0", groupId).
		First(&group).Error; err != nil {
		return nil, errors.New("群不存在或已解散")
	}
	return &group, nil
}

func (g *Group) UpdateOwnerId(ctx context.Context, groupId int, newOwnerId int) error {
	res := g.Db.WithContext(ctx).
		Model(&models.Group{}).
		Where("id = ? AND is_dismiss = 0", groupId).
		Update("owner_id", newOwnerId)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("群不存在或已解散")
	}
	return nil
}
func (g *Group) BatchGetByIDs(ctx context.Context, gids []uint64) (map[uint64]*models.Group, error) {
	res := make(map[uint64]*models.Group)
	if len(gids) == 0 {
		return res, nil
	}

	var rows []models.Group
	if err := g.Db.WithContext(ctx).Where("id IN ? AND is_dismiss = 0", gids).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	for i := range rows {
		gg := rows[i] // 注意取地址的写法，避免指针复用坑
		res[uint64(gg.Id)] = &gg
	}

	return res, nil
}
