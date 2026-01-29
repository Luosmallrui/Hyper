package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID            uint64 `gorm:"primaryKey;autoIncrement;column:id" json:"id,string"`
	PartyId       uint64 `gorm:"column:party_id;uniqueIndex:uk_party_product;not null" json:"party_id"` // 所属商家
	ParentId      uint64 `gorm:"column:parent_id;default:0;index" json:"parent_id"`                     // 0:独立商品/主套餐, >0:子商品
	ProductName   string `gorm:"column:product_name;uniqueIndex:uk_party_product;size:255;not null" json:"product_name"`
	Price         uint32 `gorm:"column:price;not null" json:"price"`                    // 售卖价(分)
	OriginalPrice uint32 `gorm:"column:original_price;default:0" json:"original_price"` // 划线价/原价(分)
	Stock         uint32 `gorm:"column:stock;not null;default:0" json:"stock"`          // 库存
	Description   string `gorm:"column:description;type:text" json:"description"`
	CoverImage    string `gorm:"column:cover_image;size:512" json:"cover_image"`
	Status        int8   `gorm:"column:status;type:tinyint;not null;default:1;index:idx_status" json:"status"` // 0-下架, 1-上架
	SalesVolume   uint32 `gorm:"column:sales_volume;default:0;index:idx_sales" json:"sales_volume"`            // 销量：用于排序
	// 关联字段：用于一次性查出套餐下的所有单品
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
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
	ConsumeType    string    `gorm:"column:consume_type" json:"consume_type"`                           // ConsumeType: 消费类型 票or商品
	SellerID       int       `gorm:"column:seller_id" json:"seller_id"`                                 // SellerID 商家ID
}

func (OrderItem) TableName() string {
	return "order_items"
}
