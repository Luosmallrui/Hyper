package types

import "Hyper/models"

type CreateProductRequest struct {
	PartyId       uint64 `json:"party_id" binding:"required"`     // 所属商家ID
	ParentId      uint64 `json:"parent_id"`                       // 父级ID (0:独立商品/套餐, >0:子单品)
	ProductName   string `json:"product_name" binding:"required"` // 商品名称
	Price         uint32 `json:"price" binding:"min=0"`           // 售卖价 (分)
	OriginalPrice uint32 `json:"original_price"`                  // 展示原价 (分)
	Stock         uint32 `json:"stock"`                           // 库存
	Description   string `json:"description"`                     // 描述
	CoverImage    string `json:"cover_image"`                     // 封面图URL
	Status        int8   `json:"status"`                          // 状态: 1-上架, 0-下架
}

// BatchGetProductsResponse 批量获取（滑动加载）响应体
type BatchGetProductsResponse struct {
	Products   []*models.Product `json:"products"`    // 商品列表
	HasMore    bool              `json:"has_more"`    // 是否还有更多数据
	NextCursor int64             `json:"next_cursor"` // 下一次请求带上的游标 (纳秒时间戳)
}

type GetDetailProductResponse struct {
	ID            uint64 `json:"id,string"`      // 商品ID，转字符串防止精度丢失
	SalesVolume   uint32 `json:"sales_volume"`   // 销量
	ProductName   string `json:"product_name"`   // 商品名称
	Price         uint32 `json:"price"`          // 售卖价 (分)
	OriginalPrice uint32 `json:"original_price"` // 原价 (分)
	CoverImage    string `json:"cover_image"`    // 封面图URL
	Description   string `json:"description"`    // 商品详情
	Stock         uint32 `json:"stock"`          // 当前库存
	Status        int8   `json:"status"`         // 状态
}
