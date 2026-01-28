package types

import "time"

type Order struct {
	Id         int       `json:"id"`
	UserId     int       `json:"user_id"`
	Type       string    `json:"type"`
	ImageUrl   string    `json:"image_url"`
	Name       string    `json:"name"`
	Price      int       `json:"price"`
	Created    time.Time `json:"created_at"`
	PaidAt     time.Time `json:"paid_at"`
	Status     int8      `json:"status"`      //支付状态
	Quantity   int       `json:"quantity"`    // 购买数量
	SellerName string    `json:"seller_name"` //商家名字
}

type Seller struct {
	Id   int    `gorm:"column:id" json:"id"`
	Name string `gorm:"column:title" json:"name"`
}
