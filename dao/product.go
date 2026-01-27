package dao

import (
	"Hyper/models"
	"context"

	"gorm.io/gorm"
)

type Product struct {
	Repo[models.Product]
}

func NewProduct(db *gorm.DB) *Product {
	return &Product{
		Repo: NewRepo[models.Product](db),
	}
}

func (p *Product) CreateProduct(ctx context.Context, product *models.Product) error {
	return p.Db.WithContext(ctx).Create(product).Error
}

func (p *Product) DeleteProduct(ctx context.Context, productID uint64) error {
	return p.Db.WithContext(ctx).Where("id = ?", productID).Delete(&models.Product{}).Error
}
