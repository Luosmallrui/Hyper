package user

import (
	"context"
	"gorm.io/gorm"
)

// Repository 接口定义
type Repository interface {
	FindByID(ctx context.Context, id uint) (*UserModel, error)
}

// repository 具体实现
type repository struct {
	db *gorm.DB
}

// NewRepository 构造函数
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindByID(ctx context.Context, id uint) (*UserModel, error) {
	var u UserModel
	err := r.db.WithContext(ctx).First(&u, id).Error
	return &u, err
}
