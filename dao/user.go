package dao

import (
	"Hyper/models"
	"context"

	"gorm.io/gorm"
)

type Users struct {
	Repo[models.Users]
}

func NewUsers(db *gorm.DB) *Users {
	return &Users{
		Repo: NewRepo[models.Users](db),
	}
}

// FindByMobile 手机号查询
func (u *Users) FindByMobile(ctx context.Context, mobile string) (*models.Users, error) {
	return u.Repo.FindByWhere(ctx, "mobile = ?", mobile)
}

// IsMobileExist 判断手机号是否存在
func (u *Users) IsMobileExist(ctx context.Context, mobile string) bool {
	exist, _ := u.Repo.IsExist(ctx, "mobile = ?", mobile)
	return exist
}
