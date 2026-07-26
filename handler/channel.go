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

type Channel struct {
	Config     *config.Config
	ChannelSrv service.IChannelService
	OssService service.IOssService
}

func (ch *Channel) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(ch.Config.Jwt.Secret))
	optionalAuth := middleware.OptionalAuth([]byte(ch.Config.Jwt.Secret))
	channel := r.Group("/v1/channel")
	channel.GET("/list", optionalAuth, context.Wrap(ch.GetChannelsList))
	channel.POST("/create", authorize, context.Wrap(ch.CreateChannel))
	channel.POST("/upload", authorize, context.Wrap(ch.UploadIcon))
	channel.GET("/", optionalAuth, context.Wrap(ch.GetUserChannelView))
	channel.GET("", optionalAuth, context.Wrap(ch.GetUserChannelView))
	channel.POST("/subscribe", authorize, context.Wrap(ch.SubscribeChannel))
	channel.POST("/unsubscribe", authorize, context.Wrap(ch.UnsubscribeChannel))
}

func (ch *Channel) UploadIcon(c *gin.Context) error {
	header, err := c.FormFile("image")
	if err != nil {
		return response.NewError(400, "missing image")
	}
	img, err := ch.OssService.UploadIcon(c.Request.Context(), header)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, img)
	return nil
}

func (ch *Channel) GetChannelsList(c *gin.Context) error {
	var req types.ListChannelsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return err
	}
	res, err := ch.ChannelSrv.ListChannels(c.Request.Context(), &req)
	if err != nil {
		return err
	}
	response.Success(c, res)
	return nil
}

func (ch *Channel) CreateChannel(c *gin.Context) error {
	var req types.CreateChannelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	res, err := ch.ChannelSrv.CreateChannel(c.Request.Context(), &req)
	if err != nil {
		return err
	}

	response.Success(c, res)
	return nil
}
func (ch *Channel) GetUserChannelView(c *gin.Context) error {
	// 获取用户ID
	var userId int
	authHeader := c.GetHeader("Authorization")
	if authHeader == "Bearer debug-mode" {
		userId = 6 // Debug 模式下使用固定用户ID
	} else {
		userId = c.GetInt("user_id")
	}
	globalchannels, err := ch.ChannelSrv.GetGlobalChannels(c.Request.Context())
	if err != nil {
		return err
	}
	res, err := ch.ChannelSrv.GetUserChannelView(c.Request.Context(), userId, globalchannels)
	if err != nil {
		return err
	}
	response.Success(c, res)
	return nil
}

func (ch *Channel) SubscribeChannel(c *gin.Context) error {
	var req types.SubscribeChannelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数格式错误")
	}
	var userId int
	authHeader := c.GetHeader("Authorization")
	if authHeader == "Bearer debug-mode" {
		userId = 6 // Debug 模式下使用固定用户ID
	} else {
		userId = c.GetInt("user_id")
	}

	err := ch.ChannelSrv.SubscribeChannel(c.Request.Context(), userId, req.ChannelId)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, "订阅成功")
	return nil
}

func (ch *Channel) UnsubscribeChannel(c *gin.Context) error {
	var req types.UnSubscribeChannelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数格式错误")
	}
	var userId int
	authHeader := c.GetHeader("Authorization")
	if authHeader == "Bearer debug-mode" {
		userId = 6 // Debug 模式下使用固定用户ID
	} else {
		userId = c.GetInt("user_id")
	}
	err := ch.ChannelSrv.UnsubscribeChannel(c.Request.Context(), userId, req.ChannelId)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, "取消订阅成功")
	return nil
}
