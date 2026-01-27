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

func (g *Group) GetGroup(ctx context.Context, groupId int) (*models.Group, error) {
	var group models.Group
	if err := g.Db.WithContext(ctx).Where("id = ?", groupId).First(&group).Error; err != nil {
		return nil, errors.New("获取群消息失败: " + err.Error())
	}
	return &group, nil
}
