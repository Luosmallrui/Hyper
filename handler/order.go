package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Order struct {
	Config           *config.Config
	OrderService     service.IOrderService
	TicketingService service.ITicketingService
}

func (o *Order) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(o.Config.Jwt.Secret))
	order := r.Group("/v1/order")
	order.Use(authorize)
	order.GET("/list", context.Wrap(o.GetOrder))
	order.POST("/create-viewer", context.Wrap(o.CreateViewer))
	order.POST("/delete-viewer", context.Wrap(o.DeleteViewer))
	order.GET("/list-viewer", context.Wrap(o.GetViewerList))
}

func (o *Order) GetOrder(c *gin.Context) error {
	if c.Query("legacy") != "1" && o.TicketingService != nil {
		var status *int8
		if rawStatus := c.Query("status"); rawStatus != "" {
			v, err := strconv.ParseInt(rawStatus, 10, 8)
			if err != nil {
				return response.NewError(http.StatusBadRequest, "订单状态参数错误")
			}
			parsed := int8(v)
			status = &parsed
		}
		userID := int64(c.GetInt("user_id"))
		orders, err := o.TicketingService.ListTicketOrders(c.Request.Context(), userID, status, orderPage(c), orderSize(c))
		if err != nil {
			return response.NewError(500, err.Error())
		}
		response.Success(c, orders)
		return nil
	}

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

func (o *Order) CreateViewer(c *gin.Context) error {
	var req types.CreateViewerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数错误")
	}
	userId := c.GetInt("user_id")
	err := o.OrderService.AddViewers(c.Request.Context(), userId, req)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, "添加观影人成功")
	return nil
}

func (o *Order) DeleteViewer(c *gin.Context) error {
	var req types.DeleteViewerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数错误")
	}
	userId := c.GetInt("user_id")
	err := o.OrderService.DeleteViewer(c.Request.Context(), userId, req)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, "删除观影人成功")
	return nil
}

func (o *Order) GetViewerList(c *gin.Context) error {
	userId := c.GetInt("user_id")
	viewers_list, err := o.OrderService.GetViewerList(c.Request.Context(), userId)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}

	response.Success(c, viewers_list)
	return nil
}

func orderPage(c *gin.Context) int {
	v, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	return v
}

func orderSize(c *gin.Context) int {
	v, _ := strconv.Atoi(c.DefaultQuery("size", c.DefaultQuery("pageSize", "10")))
	return v
}
