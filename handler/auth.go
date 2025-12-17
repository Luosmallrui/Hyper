package handler

import (
	"Hyper/config"
	"Hyper/service"
	"Hyper/types"
	"fmt"
	"github.com/gin-gonic/gin"
)

type Auth struct {
	Config        *config.Config
	UserService   service.IUserService
	WeChatService service.WeChatService
}

func (u *Auth) RegisterRouter(r gin.IRouter) {
	auth := r.Group("/")
	auth.POST("/api/auth/wx-login", u.Login) // 登录
	auth.POST("/api/auth/bind-phone", u.BindPhone)
}

func (u *Auth) Login(c *gin.Context) {
	var req types.WxLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.LoginCode == "" {
		c.JSON(400, gin.H{"error": "login_code 不能为空"})
		return
	}

	// 1. code 换 openid
	wxResp, err := u.WeChatService.Code2Session(c.Request.Context(), req.LoginCode)
	if err != nil {
		c.JSON(500, gin.H{"error": "微信登录失败"})
		return
	}

	// 2. TODO: 根据 openid 查用户 / 创建用户
	userID := int64(12345) // 示例

	// 3. TODO: 生成你自己的 token
	token := "123"

	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "登录成功",
		"data": gin.H{
			"openid":      wxResp.OpenID,
			"is_new_user": true,
			"user_id":     userID,
			"token":       token,
			"data":        wxResp,
		},
	})
}

func (u *Auth) BindPhone(c *gin.Context) {
	var req types.BindPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PhoneCode == "" {
		c.JSON(400, gin.H{"error": "phone_code 不能为空"})
		return
	}
	userPhoneNumber, err := u.WeChatService.GetUserPhoneNumber(req.PhoneCode)
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	// 3. TODO: 绑定手机号到当前登录用户（user_id 从 token 中取）
	phone := userPhoneNumber
	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "绑定手机号成功",
		"data": gin.H{
			"phone": phone,
		},
	})
}
