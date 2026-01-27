package models

import (
	"time"

	"gorm.io/gorm"
)

// Product 对应数据库中的 products 表
type Product struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement;column:id" json:"id"`                                  // ID: 自增主键
	ProductName string         `gorm:"uniqueIndex:idx_product_name;not null;column:product_name" json:"product_name"` // ProductName: 商品名称
	Price       uint32         `gorm:"not null;column:price" json:"price"`                                            // Price: 价格（单位：分）
	Stock       uint32         `gorm:"default:0;not null;column:stock" json:"stock"`                                  // Stock: 库存数量
	Description string         `gorm:"type:text;column:description" json:"description"`                               // Description: 商品详细描述
	CoverImage  string         `gorm:"size:512;default:'';column:cover_image" json:"cover_image"`                     // CoverImage: 封面图 URL
	Status      int8           `gorm:"default:1;not null;index:idx_status;column:status" json:"status"`               // Status: 状态 (0-下架, 1-上架)
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`                            // CreatedAt: 创建时间
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`                            // UpdatedAt: 更新时间
	DeletedAt   gorm.DeletedAt `gorm:"index:idx_products_deleted_at;column:deleted_at" json:"-"`                      // DeletedAt: 软删除标记
}

func (Product) TableName() string {
	return "products"
}

type OrderItem struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`                      // ID: 自增主键，明细唯一标识
	OrderID        uint64    `gorm:"not null;index:idx_order_id;column:order_id" json:"order_id"`       // OrderID: 所属订单ID，关联主订单
	ProductID      uint64    `gorm:"not null;index:idx_product_id;column:product_id" json:"product_id"` // ProductID: 商品ID，关联原始商品
	ProductName    string    `gorm:"size:255;not null;column:product_name" json:"product_name"`         // ProductName: 冗余商品名称，防止原商品删除/更名
	ProductPrice   uint32    `gorm:"not null;column:product_price" json:"product_price"`                // ProductPrice: 冗余下单单价（分），锁定成交价
	Quantity       uint32    `gorm:"default:1;not null;column:quantity" json:"quantity"`                // Quantity: 购买数量
	SubtotalAmount uint32    `gorm:"not null;column:subtotal_amount" json:"subtotal_amount"`            // SubtotalAmount: 小计金额（分），单价 * 数量
	CoverImage     string    `gorm:"size:512;default:'';column:cover_image" json:"cover_image"`         // CoverImage: 冗余商品封面图，防止原图失效
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`                // CreatedAt: 明细创建时间
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`                // UpdatedAt: 最后更新时间
}

func (OrderItem) TableName() string {
	return "order_items"
}
