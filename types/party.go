package types

import "Hyper/models"

type CreatePartyRequest struct {
	Title        string   `json:"title" binding:"required"`
	Type         string   `json:"type" binding:"required"`
	Description  string   `json:"description"`
	LocationName string   `json:"locationName" binding:"required"`
	Address      string   `json:"address"`
	Lat          float64  `json:"lat" binding:"required"`
	Lng          float64  `json:"lng" binding:"required"`
	Price        float64  `json:"price"`                        // 接收元，存入分
	MaxAttendees int      `json:"maxAttendees"`                 // 人数限制
	StartTime    string   `json:"startTime" binding:"required"` // 2026-02-01 20:00:00
	EndTime      string   `json:"endTime"`
	CoverImage   string   `json:"coverImage"`
	Images       []string `json:"images"` // 图片数组
	Tags         []string `json:"tags"`   // 标签数组
}

type PartyList struct {
	List     []models.MerchantListItem `json:"list"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"pageSize"`
}

type MerchantDetail struct {
	Id           int64            `json:"id"`
	Name         string           `json:"name"`
	AvgPrice     int64            `json:"avg_price"` //人均价格
	LocationName string           `json:"location_name"`
	Images       []string         `json:"images"`
	Goods        []models.Product `json:"goods"`
	ListNotesBriefRep
	UserName      string `json:"user_name,omitempty"`
	UserAvatar    string `json:"user_avatar,omitempty"`
	IsFollow      bool   `json:"is_follow"`
	BusinessHours string `json:"business_hours"`
}
