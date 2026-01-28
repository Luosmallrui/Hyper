package handler

import (
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/types"
	"time"

	"github.com/gin-gonic/gin"
)

type Point struct {
}

func (p *Point) RegisterRouter(r gin.IRouter) {
	pointGroup := r.Group("/v1/points")
	pointGroup.GET("/balance", context.Wrap(p.Balance))
	pointGroup.GET("/records", context.Wrap(p.GetRecords))

}

func (p *Point) Balance(c *gin.Context) error {

	resp := types.PointsAccount{
		Balance:       3000,
		TotalEarned:   5000,
		TotalUsed:     2000,
		PendingCount:  3,
		PendingAmount: 20,
	}
	response.Success(c, resp)
	return nil
}

func (p *Point) GetRecords(c *gin.Context) error {
	l := make([]types.PointRecord, 0)
	l = append(l, types.PointRecord{
		ID:          1,
		Amount:      80,
		Description: "消费（门票）",
		OrderType:   "ticket",
		CreatedAt:   time.Now(),
		Status:      1,
	})
	l = append(l, types.PointRecord{
		ID:          2,
		Amount:      -190,
		Description: "积分使用（汽车改装）",
		OrderType:   "car_service",
		CreatedAt:   time.Now(),
		Status:      1,
	})
	resp := types.ListPointsRecord{
		Records:    l,
		NextCursor: 0,
		HasMore:    false,
	}
	response.Success(c, resp)
	return nil
}
