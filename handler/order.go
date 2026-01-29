package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Order struct {
	Config       *config.Config
	OrderService service.IOrderService
}

func (o *Order) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(o.Config.Jwt.Secret))
	order := r.Group("/v1/order")
	order.Use(authorize)
	order.GET("/list", context.Wrap(o.GetOrder))

}

func (o *Order) GetOrder(c *gin.Context) error {
	var req types.FeedRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数错误")
	}
	userId := c.GetInt("user_id")
	orders, nextCursor, hasMore, err := o.OrderService.GetOrderList(c, userId, req.Cursor, req.PageSize)
	if err != nil {
		return response.NewError(500, err.Error())
	}

	resp := types.ListOrderReq{
		Orders:     orders,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}
	response.Success(c, resp)
	return nil
}
