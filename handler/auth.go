package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/jwt"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Auth struct {
	Config        *config.Config
	UserService   service.IUserService
	WeChatService service.IWeChatService
	OssService    service.IOssService
}

func (u *Auth) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(u.Config.Jwt.Secret))
	auth := r.Group("/")
	auth.POST("/api/auth/wx-login", context.Wrap(u.Login))                     // 微信登录
	auth.POST("/api/auth/bind-phone", authorize, context.Wrap(u.BindPhone))    //微信获取手机号
	auth.POST("api/auth/send-sms", authorize, context.Wrap(u.SendSms))         //发送验证码
	auth.POST("api/auth/update-phone", authorize, context.Wrap(u.UpdatePhone)) //更新手机号
	//auth.POST("/api/auth/update-phone", authorize, context.Wrap(u.UpdatePhone)) //更新手机号
	auth.GET("/token", context.Wrap(u.GetToken))
}

func (u *Auth) GetToken(c *gin.Context) error {
	token, err := jwt.GenerateToken([]byte(u.Config.Jwt.Secret), 1, "XXX")
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, token)
	return nil
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

	rep := types.UserProfileResp{
		User: types.UserBasicInfo{
			UserID:      231255123123,
			Nickname:    "邪修的马路路",
			PhoneNumber: user.Mobile,
		},
		Stats: types.UserStats{
			Following: 25,
			Follower:  115,
			Likes:     25,
		},
		Token: token,
	}
	response.Success(c, rep)
	return nil
}

func (u *Auth) BindPhone(c *gin.Context) error {
	userId, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	var req types.BindPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PhoneCode == "" {
		return response.NewError(http.StatusInternalServerError, "phone_code 不能为空")
	}
	userPhoneNumber, err := u.WeChatService.GetUserPhoneNumber(req.PhoneCode)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	phone := userPhoneNumber
	err = u.UserService.UpdateMobile(c.Request.Context(), int(userId), phone)
	if err != nil {
		return response.NewError(500, err.Error())
	}

	response.Success(c, types.BindPhoneRep{PhoneNumber: phone})
	return nil
}

// 发送验证码
func (u *Auth) SendSms(c *gin.Context) error {
	var req types.SendSmsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	err := u.UserService.SendVerifyCode(c.Request.Context(), req.Mobile)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, "验证码发送成功")
	return nil
}

// 更新手机号
func (u *Auth) UpdatePhone(c *gin.Context) error {
	userId, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	var req types.UpdatePhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	err = u.UserService.UpdateMobileWithSms(c.Request.Context(), req.Mobile, int(userId), req.Code)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, "手机号更新成功")
	return nil
}
