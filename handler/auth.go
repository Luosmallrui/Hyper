package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/jwt"
	"Hyper/pkg/log"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Auth struct {
	Config         *config.Config
	UserService    service.IUserService
	WeChatService  service.IWeChatService
	OssService     service.IOssService
	FollowService  service.IFollowService
	LikeService    service.ILikeService
	CollectService service.ICollectService
}

func (u *Auth) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(u.Config.Jwt.Secret))
	auth := r.Group("/")
	auth.POST("/auth/wx-login", context.Wrap(u.Login))                  // 微信登录
	auth.POST("/auth/bind-phone", authorize, context.Wrap(u.BindPhone)) //微信获取手机号
	auth.POST("/auth/refresh", context.Wrap(u.Refresh))
	auth.GET("/token", context.Wrap(u.GetToken))
	auth.GET("/test1", authorize, context.Wrap(u.test))
}

func (u *Auth) test(c *gin.Context) error {
	userId, err := context.GetUserID(c)
	fmt.Println(userId)
	fmt.Println(err)
	return nil
}

func (u *Auth) GetToken(c *gin.Context) error {
	token, err := jwt.GenerateToken([]byte(u.Config.Jwt.Secret), 1, "XX", "access", 2*time.Hour)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	fmt.Println(u.Config.Jwt.Secret, 123)
	f, err := jwt.GenerateToken([]byte(u.Config.Jwt.Secret), 1, "XX", "refresh", 2*time.Hour)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, gin.H{"token": token, "refresh": f})
	return nil
}

func (u *Auth) Refresh(c *gin.Context) error {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return response.NewError(401, "Authorization not find")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return response.NewError(400, "Authorization 格式错误")
	}

	claims, err := jwt.ParseToken([]byte(u.Config.Jwt.Secret), "refresh", parts[1])
	if err != nil {
		return response.NewError(http.StatusUnauthorized, err.Error())
	}

	newAccessToken, _ := jwt.GenerateToken([]byte(u.Config.Jwt.Secret), claims.UserID, claims.OpenID, "access", 60*time.Second)
	resp := gin.H{
		"access_token":  newAccessToken,
		"refresh_token": "",
	}
	if jwt.ShouldRotateRefreshToken(claims, 24*time.Hour) {
		newRefreshToken, err := jwt.GenerateToken([]byte(u.Config.Jwt.Secret), claims.UserID, claims.OpenID, "refresh", 7*24*time.Hour)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "failed to rotate refresh token"})
			return response.NewError(500, err.Error())
		}
		resp["refresh_token"] = newRefreshToken
	}
	response.Success(c, resp)
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
	accessToken, err := jwt.GenerateToken([]byte(u.Config.Jwt.Secret), uint(user.Id), user.OpenID, "access", 60*time.Second)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	log.L.Info("generating access token", zap.String("token", accessToken))
	refreshToken, err := jwt.GenerateToken([]byte(u.Config.Jwt.Secret), uint(user.Id), user.OpenID, "refresh", 7*24*time.Hour)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	log.L.Info("generating refresh token", zap.String("token", refreshToken))

	rep := types.UserToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
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
