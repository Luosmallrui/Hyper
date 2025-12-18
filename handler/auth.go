package handler

import (
	"Hyper/config"
	"Hyper/pkg/context"
	"Hyper/pkg/jwt"
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
	auth.POST("/api/auth/wx-login", context.Wrap(u.Login))       // 微信登录
	auth.POST("/api/auth/bind-phone", context.Wrap(u.BindPhone)) //微信获取手机号
}

func (u *Auth) Login(c *gin.Context) error {
	var req types.WxLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数格式错误")
	}

	if req.LoginCode == "" {
		return response.NewError(http.StatusInternalServerError, "login_code 不能为空")
	}

	wxResp, err := u.WeChatService.Code2Session(c.Request.Context(), req.LoginCode)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}

	user, err := u.UserService.GetOrCreateByOpenID(c.Request.Context(), wxResp.OpenID)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}

	token, err := jwt.GenerateToken([]byte(u.Config.Jwt.Secret), uint(user.Id), user.OpenID)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}

	rep := types.LoginRep{
		Token:       token,
		UserId:      user.Id,
		OpenId:      user.OpenID,
		PhoneNumber: user.Mobile,
	}
	response.Success(c, rep)
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
	response.Success(c, types.BindPhoneRep{PhoneNumber: phone})
	return nil
}
