package types

import "Hyper/models"

type GetDistrictsListRequest struct {
	CityID int `json:"city_id" binding:"required"`
}

type GetDistrictDetailRequest struct {
	DistrictID int `json:"district_id" form:"district_id" binding:"required"`
	Cursor     int `json:"cursor" form:"cursor"`          // 传上一次结果中最后一条记录的 ID
	Limit      int `json:"limit" form:"limit,default=10"` // 每次获取多少个点
}

type GetDistrictDetailResponse struct {
	List       []models.Areas `json:"list"`
	NextCursor int            `json:"next_cursor"` // 下一次请求带上的游标
	HasMore    bool           `json:"has_more"`    // 是否还有更多数据
}

type MapMarker struct {
	ID              string  `json:"id"`
	Source          string  `json:"source"`    // party | activity
	SourceID        int64   `json:"source_id"` // 原表 ID
	UserID          int64   `json:"user_id"`
	User            string  `json:"user"`
	UserName        string  `json:"username"`
	UserAvatar      string  `json:"user_avatar"`
	UserAvatarCamel string  `json:"userAvatar"`
	Title           string  `json:"title"`
	Type            string  `json:"type"` // 派对 | 场地 | activity
	Location        string  `json:"location"`
	Address         string  `json:"address"`
	Lat             float64 `json:"lat"`
	Lng             float64 `json:"lng"`
	CoverImage      string  `json:"cover_image"`
	CreatedAt       string  `json:"created_at"`
	AvgPrice        int64   `json:"avg_price"`
	CurrentCount    int64   `json:"current_count"`
	PostCount       int64   `json:"post_count"`
	Icon            string  `json:"icon"`
	IsSubscriber    bool    `json:"is_subscribe"`
	IsFollow        bool    `json:"is_follow"`
	StartTime       string  `json:"start_time,omitempty"`
	EndTime         string  `json:"end_time,omitempty"`
	Status          any     `json:"status"`
}

type MapMarkerResponse struct {
	List  []MapMarker `json:"list"`
	Total int         `json:"total"`
}
