package dao

import (
	"Hyper/models"
	"context"

	"gorm.io/gorm"
)

type Image struct {
	Repo[models.Image]
}

func NewImage(db *gorm.DB) *Image {
	return &Image{
		Repo: NewRepo[models.Image](db),
	}
}

func (u *Image) CreateImage(ctx context.Context, image *models.Image) error {
	return u.Repo.Db.WithContext(ctx).Create(image).Error
}
