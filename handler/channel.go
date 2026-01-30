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
	channel := r.Group("/v1/channel")
	channel.GET("/list", authorize, context.Wrap(ch.GetChannelsList)) //创建
	channel.POST("/create", authorize, context.Wrap(ch.CreateChannel))
	channel.POST("/upload", authorize, context.Wrap(ch.UploadIcon))

}

func (ch *Channel) UploadIcon(c *gin.Context) error {
	userID, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	header, err := c.FormFile("image")
	if err != nil {
		return response.NewError(400, "missing image")
	}
	img, err := ch.OssService.UploadImage(c.Request.Context(), int(userID), header)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, img)
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
