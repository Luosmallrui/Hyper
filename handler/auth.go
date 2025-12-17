package handler

import (
	"Hyper/config"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Auth struct {
	Config        *config.Config
	UserService   service.IUserService
	WeChatService service.WeChatService
}

func (u *Auth) RegisterRouter(r gin.IRouter) {
	auth := r.Group("/")
	auth.POST("/api/auth/wx-login", context.Wrap(u.Login)) // 登录
	auth.POST("/api/auth/bind-phone", context.Wrap(u.BindPhone))
}

func (u *Auth) Login(c *gin.Context) error {
	var req types.WxLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}

	if req.LoginCode == "" {
		return response.NewError(http.StatusInternalServerError, "login_code 不能为空")
	}

	wxResp, err := u.WeChatService.Code2Session(c.Request.Context(), req.LoginCode)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, wxResp)
	return nil
}

func (u *Auth) BindPhone(c *gin.Context) error {
	var req types.BindPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PhoneCode == "" {
		return response.NewError(http.StatusInternalServerError, "phone_code 不能为空")
	}
	userPhoneNumber, err := u.WeChatService.GetUserPhoneNumber(req.PhoneCode)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	phone := userPhoneNumber
	response.Success(c, gin.H{
		"code": 0,
		"msg":  "绑定手机号成功",
		"data": gin.H{
			"phone": phone,
		},
	})
	return nil
}
