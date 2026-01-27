package types

type CreateProductRequest struct {
	ProductName string `json:"product_name" binding:"required"`      // ProductName: 商品名称 (对应数据库 product_name)，必填且唯一
	Price       uint32 `json:"price" binding:"required,gt=0"`        // Price: 商品价格 (对应数据库 price)，单位：分，必须大于 0
	Stock       uint32 `json:"stock" binding:"omitempty,gte=0"`      // Stock: 库存数量 (对应数据库 stock)，不传默认为 0，不能为负数
	Description string `json:"description"`                          // Description: 商品详细描述 (对应数据库 description)，可选
	CoverImage  string `json:"cover_image"`                          // CoverImage: 商品封面图 URL (对应数据库 cover_image)，可选
	Status      int8   `json:"status" binding:"omitempty,oneof=0 1"` // Status: 状态 (对应数据库 status)，0-下架, 1-上架，默认为 1
}
