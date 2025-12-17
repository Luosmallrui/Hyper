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

	wxResp, err := u.WeChatService.Code2Session(c.Request.Context(), req.LoginCode)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"data": wxResp,
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
	phone := userPhoneNumber
	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "绑定手机号成功",
		"data": gin.H{
			"phone": phone,
		},
	})
}
