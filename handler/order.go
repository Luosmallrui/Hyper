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
	order.POST("/create-viewer", authorize, context.Wrap(o.CreateViewer))
	order.POST("/delete-viewer", authorize, context.Wrap(o.DeleteViewer))
	order.GET("/list-viewer", authorize, context.Wrap(o.GetViewerList))

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

func (o *Order) CreateViewer(c *gin.Context) error {
	var req types.CreateViewerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数错误")
	}
	var userId int
	authHeader := c.GetHeader("Authorization")
	if authHeader == "Bearer debug-mode" {
		userId = 6 // Debug 模式下使用固定用户ID
	} else {
		userId = c.GetInt("user_id")
	}
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
	var userId int
	authHeader := c.GetHeader("Authorization")
	if authHeader == "Bearer debug-mode" {
		userId = 6 // Debug 模式下使用固定用户ID
	} else {
		userId = c.GetInt("user_id")
	}
	err := o.OrderService.DeleteViewer(c.Request.Context(), userId, req)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, "删除观影人成功")
	return nil
}

func (o *Order) GetViewerList(c *gin.Context) error {
	var req types.ListViewerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数错误")
	}
	var userId int
	authHeader := c.GetHeader("Authorization")
	if authHeader == "Bearer debug-mode" {
		userId = 6 // Debug 模式下使用固定用户ID
	} else {
		userId = c.GetInt("user_id")
	}
	viewers_list, err := o.OrderService.GetViewerList(c.Request.Context(), userId)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}

	response.Success(c, viewers_list)
	return nil

}
