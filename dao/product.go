package dao

import (
	"Hyper/models"
	"context"
	"time"

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

func (d *Product) GetProductsByCursor(ctx context.Context, partyID uint64, limit int, cursor int) ([]*models.Product, error) {
	var products []*models.Product
	db := d.Db.WithContext(ctx).Where("party_id = ? AND status = 1", partyID)

	// 如果有游标，查询创建时间早于游标的数据 (向下滚动)
	if cursor > 0 {
		// 将纳秒时间戳转回 time.Time
		cursorTime := time.Unix(0, int64(cursor))
		db = db.Where("created_at < ?", cursorTime)
	}

	err := db.Order("sales_volume DESC").
		Order("created_at DESC").
		Limit(limit).
		Find(&products).Error
	if err != nil {
		return nil, err
	}
	return products, err
}

func (d *Product) FindProductById(ctx context.Context, productID uint64, partyId int64) (*models.Product, error) {
	var product models.Product
	err := d.Db.WithContext(ctx).Where("id = ? AND party_id = ?", productID, partyId).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}
