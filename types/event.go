package types

import "Hyper/models"

// CreateEventRequest 创建活动请求
type CreateEventRequest struct {
	Title           string            `json:"title" binding:"required,max=80"`
	ShareTitle      string            `json:"shareTitle" binding:"required,max=20"`
	StartAt         string            `json:"startAt" binding:"required"`
	EndAt           string            `json:"endAt" binding:"required"`
	Summary         string            `json:"summary"`
	SummaryHTML     string            `json:"summaryHtml"`
	StrongRealName  bool              `json:"strongRealName"`
	MinorProtection bool              `json:"minorProtection"`
	Address         EventAddress      `json:"address" binding:"required"`
	Posters         EventPosterReq    `json:"posters" binding:"required"`
	Tickets         []EventTicketReq  `json:"tickets" binding:"required,min=1"`
	Qualification   QualificationReq  `json:"qualification" binding:"required"`
}

type EventAddress struct {
	Mode     string `json:"mode"`
	Province string `json:"province"`
	City     string `json:"city"`
	District string `json:"district"`
	Location string `json:"location"`
}

type EventPosterReq struct {
	DetailPoster     string `json:"detailPoster"`
	DetailLongPoster string `json:"detailLongPoster"`
	SharePoster      string `json:"sharePoster" binding:"required"`
	GroupPoster      string `json:"groupPoster"`
}

type EventTicketReq struct {
	Name          string `json:"name" binding:"required"`
	Price         int64  `json:"price" binding:"min=0"`
	Stock         int    `json:"stock" binding:"min=1"`
	PurchaseLimit int    `json:"purchaseLimit" binding:"min=1"`
	StartAt       string `json:"startAt" binding:"required"`
	EndAt         string `json:"endAt" binding:"required"`
	Description   string `json:"description"`
}

type QualificationReq struct {
	PromoterConfirmed bool   `json:"promoterConfirmed"`
	SafetyAgreement   bool   `json:"safetyAgreement"`
	TicketStatement   bool   `json:"ticketStatement"`
	Notes             string `json:"notes"`
}

// EventDetail 活动详情响应
type EventDetail struct {
	models.Event
	Tickets []models.EventTicket `json:"tickets"`
}

// EventListResponse 活动列表响应
type EventListResponse struct {
	List     []models.EventListItem `json:"list"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"pageSize"`
}
