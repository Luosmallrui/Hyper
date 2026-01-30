package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/service"
	"Hyper/types"

	"github.com/gin-gonic/gin"
)

type Channel struct {
	Config     *config.Config
	ChannelSrv service.IChannelService
}

func (ch *Channel) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(ch.Config.Jwt.Secret))
	channel := r.Group("/v1/channel")
	channel.GET("/list", authorize, context.Wrap(ch.GetChannelsList)) //创建
	channel.POST("/create", authorize, context.Wrap(ch.CreateChannel))

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

	c.JSON(200, res)
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

	c.JSON(200, res)
	return nil
}
