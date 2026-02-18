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
