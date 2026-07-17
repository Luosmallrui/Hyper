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
	"regexp"
	"strconv"
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
	SmsService     service.ISMSService
}

func (u *Auth) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(u.Config.Jwt.Secret))
	auth := r.Group("/v1/auth")
	auth.POST("/wx-login", context.Wrap(u.Login)) // 微信登录
	auth.POST("/login", context.Wrap(u.SMSLogin))
	auth.POST("/login-password", context.Wrap(u.PasswordLogin))
	auth.POST("/set-password", authorize, context.Wrap(u.SetPassword))
	auth.POST("/reset-password", context.Wrap(u.ResetPassword))
	auth.POST("/send-code", context.Wrap(u.SendCode))
	auth.POST("/bind-phone", authorize, context.Wrap(u.BindPhone)) //微信获取手机号
	auth.POST("/refresh", context.Wrap(u.Refresh))
	auth.GET("/profile", authorize, context.Wrap(u.GetProfile))
	auth.GET("/token", context.Wrap(u.GetToken))
	auth.POST("/update-profile", authorize, context.Wrap(u.UpdateUserProfile)) //更新用户信息
	auth.POST("/send-sms", authorize, context.Wrap(u.SendSms))                 //发送验证码
	auth.POST("/update-phone", authorize, context.Wrap(u.UpdatePhone))         //更新手机号
	auth.GET("/test1", authorize, context.Wrap(u.test))
}

// GetProfile supplies the minimal authenticated profile needed by PC account flows.
func (u *Auth) GetProfile(c *gin.Context) error {
	userID, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "获取用户身份失败")
	}
	user, err := u.UserService.GetUserInfo(c.Request.Context(), int(userID))
	if err != nil {
		return err
	}
	response.Success(c, gin.H{
		"user_id":  user.Id,
		"phone":    user.Mobile,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
	})
	return nil
}

func (u *Auth) test(c *gin.Context) error {
	userId, err := context.GetUserID(c)
	fmt.Println(userId)
	fmt.Println(err)
	return nil
}

func (u *Auth) SMSLogin(c *gin.Context) error {
	var req types.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, err.Error())
	}
	// 验证手机号格式
	if !isValidPhone(req.Phone) {
		return response.NewError(400, "手机号格式不正确")
	}
	// 验证验证码
	valid, err := u.SmsService.VerifyCode(c, req.Phone, req.Code)
	if err != nil || !valid {
		return response.NewError(400, "验证码错误或已过期")
	}

	user, err := u.UserService.RegisterOrLogin(c, req.Phone)
	if err != nil {
		return response.NewError(400, err.Error())
	}

	accessToken, err := jwt.GenerateToken([]byte(u.Config.Jwt.Secret), uint(user.Id), user.OpenID, "access", time.Duration(u.Config.Jwt.ExpiresTime)*time.Second)
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
		AccessToken:     accessToken,
		RefreshToken:    refreshToken,
		AccessExpire:    time.Now().Add(time.Duration(u.Config.Jwt.ExpiresTime) * time.Second).Unix(),
		RefreshExpire:   time.Now().Add(7 * 24 * time.Hour).Unix(),
		NeedSetPassword: user.Password == "",
	}
	response.Success(c, rep)
	return nil
}

func isValidPhone(phone string) bool {
	pattern := `^1[3-9]\d{9}$`
	match, _ := regexp.MatchString(pattern, phone)
	return match
}

func (u *Auth) SendCode(c *gin.Context) error {
	var req types.SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, err.Error())
	}
	if !isValidPhone(req.Phone) {
		return response.NewError(400, "手机号格式不正确")
	}

	err := u.SmsService.SendVerificationCode(c, req.Phone)
	if err != nil {
		return response.NewError(400, err.Error())
	}
	return nil
}

// 本地测试
func (u *Auth) GetToken(c *gin.Context) error {
	secret := []byte(u.Config.Jwt.Secret)

	// 本地测试：固定返回 8/9/10 三个用户的 token
	uids := []uint{8, 9, 10}

	result := make(map[string]gin.H, len(uids))
	for _, uid := range uids {
		access, err := jwt.GenerateToken(secret, uid, "XX", "access", 2*time.Hour)
		if err != nil {
			return response.NewError(http.StatusInternalServerError, err.Error())
		}
		refresh, err := jwt.GenerateToken(secret, uid, "XX", "refresh", 2*time.Hour)
		if err != nil {
			return response.NewError(http.StatusInternalServerError, err.Error())
		}

		// key 用字符串，方便前端/工具查看
		result[strconv.FormatUint(uint64(uid), 10)] = gin.H{
			"token":   access,
			"refresh": refresh,
		}
	}

	response.Success(c, gin.H{"tokens": result})
	return nil
}

//func (u *Auth) GetToken(c *gin.Context) error {
//
//	token, err := jwt.GenerateToken([]byte(u.Config.Jwt.Secret), 1, "XX", "access", 2*time.Hour)
//	if err != nil {
//		return response.NewError(http.StatusInternalServerError, err.Error())
//	}
//	fmt.Println(u.Config.Jwt.Secret, 123)
//	f, err := jwt.GenerateToken([]byte(u.Config.Jwt.Secret), 1, "XX", "refresh", 2*time.Hour)
//	if err != nil {
//		return response.NewError(http.StatusInternalServerError, err.Error())
//	}
//	response.Success(c, gin.H{"token": token, "refresh": f})
//	return nil
//}

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
	expireDuration := time.Duration(u.Config.Jwt.ExpiresTime) * time.Second
	expireAt := time.Now().Add(expireDuration)

	newAccessToken, _ := jwt.GenerateToken([]byte(u.Config.Jwt.Secret), claims.UserID, claims.OpenID, "access", expireDuration)
	resp := gin.H{
		"access_token":   newAccessToken,
		"refresh_token":  "",
		"access_expire":  expireAt.Unix(),
		"refresh_expire": claims.ExpiresAt,
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
	accessToken, err := jwt.GenerateToken([]byte(u.Config.Jwt.Secret), uint(user.Id), user.OpenID, "access", time.Duration(u.Config.Jwt.ExpiresTime)*time.Second)
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
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		AccessExpire:  time.Now().Add(time.Duration(u.Config.Jwt.ExpiresTime) * time.Second).Unix(),
		RefreshExpire: time.Now().Add(7 * 24 * time.Hour).Unix(),
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

// SendSms 发送验证码
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

// UpdatePhone 更新手机号
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

// PasswordLogin 手机号+密码登录
// POST /api/v1/auth/login-password
func (u *Auth) PasswordLogin(c *gin.Context) error {
	var req types.PasswordLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, err.Error())
	}
	if !isValidPhone(req.Phone) {
		return response.NewError(400, "手机号格式不正确")
	}

	user, err := u.UserService.Login(req.Phone, req.Password)
	if err != nil {
		return response.NewError(400, err.Error())
	}

	accessToken, err := jwt.GenerateToken([]byte(u.Config.Jwt.Secret), uint(user.Id), user.OpenID, "access", time.Duration(u.Config.Jwt.ExpiresTime)*time.Second)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	refreshToken, err := jwt.GenerateToken([]byte(u.Config.Jwt.Secret), uint(user.Id), user.OpenID, "refresh", 7*24*time.Hour)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	rep := types.UserToken{
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		AccessExpire:  time.Now().Add(time.Duration(u.Config.Jwt.ExpiresTime) * time.Second).Unix(),
		RefreshExpire: time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	response.Success(c, rep)
	return nil
}

// SetPassword 设置密码（短信登录后设置，后续可用密码登录）
// POST /api/v1/auth/set-password
func (u *Auth) SetPassword(c *gin.Context) error {
	var req types.SetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, err.Error())
	}

	userId, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "请先登录")
	}

	err = u.UserService.SetPassword(c.Request.Context(), int(userId), req.Password)
	if err != nil {
		return response.NewError(500, err.Error())
	}

	response.Success(c, "密码设置成功")
	return nil
}

// ResetPassword 找回密码（手机号 + 验证码 + 新密码）
// POST /api/v1/auth/reset-password
func (u *Auth) ResetPassword(c *gin.Context) error {
	var req types.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, err.Error())
	}
	if !isValidPhone(req.Phone) {
		return response.NewError(400, "手机号格式不正确")
	}

	valid, err := u.SmsService.VerifyCode(c.Request.Context(), req.Phone, req.Code)
	if err != nil || !valid {
		return response.NewError(400, "验证码错误或已过期")
	}

	if err := u.UserService.ResetPassword(c.Request.Context(), req.Phone, req.Password); err != nil {
		return response.NewError(400, err.Error())
	}

	response.Success(c, "密码重置成功")
	return nil
}

func (u *Auth) UpdateUserProfile(c *gin.Context) error {
	var req types.UpdateUserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	userId, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	err = u.UserService.UpdateUserProfile(c.Request.Context(), int(userId), &req)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, req)
	return nil
}
