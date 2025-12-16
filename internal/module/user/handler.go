package user

import (
	"Hyper/pkg/util"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler 用户模块的 HTTP 处理器
type Handler struct {
	svc Service
}

// NewHandler 构造函数
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	userGroup := r.Group("/users")
	{
		userGroup.GET("/:id", h.GetUser)
		userGroup.POST("/login", h.Login)
	}
}
func (h *Handler) Login(c *gin.Context) {
	var req WxLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	// 1. 获取 AccessToken
	accessToken, err := util.GetAccessToken()
	if err != nil {
		c.JSON(500, gin.H{"error": "获取微信Token失败"})
		return
	}

	// 2. 解析手机号 (调用微信 getuserphonenumber 接口)
	// 官方文档: https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/user-info/phone-number/getPhoneNumber.html
	phoneUrl := fmt.Sprintf("https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s", accessToken)
	phoneBody := map[string]string{"code": req.PhoneCode}
	jsonData, _ := json.Marshal(phoneBody)

	phoneResp, err := http.Post(phoneUrl, "application/json", bytes.NewBuffer(jsonData))
	// ... 错误处理 ...

	var wxPhoneRes WxPhoneResponse
	json.NewDecoder(phoneResp.Body).Decode(&wxPhoneRes)

	if wxPhoneRes.ErrCode != 0 {
		c.JSON(400, gin.H{"error": "手机号解析失败: " + wxPhoneRes.ErrMsg})
		return
	}

	//phoneNumber := wxPhoneRes.PhoneInfo.PurePhoneNumber

	// 3. (可选但推荐) 获取 OpenID
	// 虽然手机号能唯一标识用户，但 OpenID 是微信生态的唯一ID，建议同时获取并关联
	// url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code", AppID, AppSecret, req.LoginCode)
	// ... 请求并解析出 openid ...

	// 4. 数据库逻辑 (Find or Create)
	// user := DB.FindUserByPhone(phoneNumber)
	// if user == nil {
	//     user = DB.CreateUser(phoneNumber, openid)
	// }

	// 5. 生成系统 Token (JWT)
	token := "eyJhbGciOiJIUzI1NiIsInR5c..." // 使用 JWT 库生成

	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "登录成功",
		"data": gin.H{
			"token":       token,
			"user_id":     12345, // user.ID
			"is_new_user": true,  // 告诉前端是否是新用户，决定是否跳转设置头像
		},
	})
}

// GetUser 具体处理函数
func (h *Handler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	user, err := h.svc.GetUser(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	user, err := h.svc.CreateUser(c.Request.Context(), 5)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, UserResponse{
		ID:    user.ID,
		Email: user.Email,
	})
}
