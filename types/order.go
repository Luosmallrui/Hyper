package types

import (
	"Hyper/models"
	"time"
)

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

type ListOrderReq struct {
	Orders     []*Order `json:"list"`
	HasMore    bool     `json:"has_more"`
	NextCursor int64    `json:"next_cursor,string"`
}

type CreateViewerReq struct {
	RealName string `json:"real_name" binding:"required,min=2,max=20"` // 真实姓名
	IDCard   string `json:"id_card" binding:"required,len=18"`         // 身份证号（固定18位）
	Phone    string `json:"phone" binding:"required,len=11"`           // 手机号（固定11位）
}

type UpdateViewerReq struct {
	ID       int64  `json:"id"`
	RealName string `json:"real_name" binding:"omitempty,min=2,max=20"`
	Phone    string `json:"phone" binding:"omitempty,len=11"`
}

type DeleteViewerReq struct {
	ID int64 `json:"id" binding:"required"`
}

type ListViewerReq struct {
	userId int `json:"user_id" binding:"required"`
}

type GetViewerListResp struct {
	Total   int64            `json:"total"`
	Viewers []*models.Viewer `json:"viewers"`
}
