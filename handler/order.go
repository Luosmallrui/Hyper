package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"

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
	userId := c.GetInt("user_id")
	resp, err := o.OrderService.GetOrderList(c, userId)
	if err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, resp)
	return nil
}
