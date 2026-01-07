package dao

import (
	"Hyper/models"
	"context"
	"fmt"

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

func (u *Users) GetOrCreateByOpenID(ctx context.Context, openid string) (*models.Users, error) {
	user := &models.Users{OpenID: openid}
	err := u.Repo.Db.WithContext(ctx).
		Where("open_id = ?", openid).
		FirstOrCreate(user).Error
	return user, err
}

func (u *Users) UpdateById(
	ctx context.Context,
	id int64,
	data map[string]any,
) error {

	if id <= 0 {
		return gorm.ErrRecordNotFound
	}
	return u.Db.WithContext(ctx).
		Model(&models.Users{}).
		Where("id = ?", id).
		Updates(data).Error

}

func (u *Users) Update(ctx context.Context, userID uint, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	err := u.Db.WithContext(ctx).
		Model(&models.Users{}).
		Where("id = ?", userID).
		Updates(updates).Error

	if err != nil {
		return fmt.Errorf("dao.User.Update error: %w", err)
	}

	return nil
}
