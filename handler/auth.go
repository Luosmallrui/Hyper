package handler

import (
	"Hyper/config"
	"Hyper/service"
	"Hyper/types"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

type Auth struct {
	Config *config.Config
	//AdminRepo       *dao.Admin
	//UserRepo        *dao.Users
	//JwtTokenStorage *dao.JwtTokenStorage
	//ICaptcha        *base64Captcha.Captcha
	UserService service.IUserService
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
	wxResp, err := Code2Session(req.LoginCode)
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
		},
	})
}

func (u *Auth) BindPhone(c *gin.Context) {
	var req types.BindPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PhoneCode == "" {
		fmt.Println(err)
		c.JSON(400, gin.H{"error": "phone_code 不能为空"})
		return
	}

	// 1. 获取 access_token
	accessToken, err := u.getAccessToken()
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{"error": "获取 access_token 失败"})
		return
	}

	// 2. 调用微信换手机号接口
	wxAPI := fmt.Sprintf(
		"https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s",
		accessToken,
	)

	body, _ := json.Marshal(map[string]string{
		"code": req.PhoneCode, // ⚠️ 只能是 getPhoneNumber 的 code
	})

	resp, err := http.Post(wxAPI, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{"error": "调用微信接口失败"})
		return
	}
	defer resp.Body.Close()

	var wxResp types.WxPhoneResponse
	if err := json.NewDecoder(resp.Body).Decode(&wxResp); err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{"error": "解析微信返回失败"})
		return
	}

	if wxResp.ErrCode != 0 {
		c.JSON(400, gin.H{"error": wxResp.ErrMsg})
		return
	}

	// 3. TODO: 绑定手机号到当前登录用户（user_id 从 token 中取）
	phone := wxResp.PhoneInfo.PhoneNumber
	fmt.Println(err)
	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "绑定手机号成功",
		"data": gin.H{
			"phone": phone,
			"data":  wxResp,
		},
	})
}
func (u *Auth) getAccessToken() (string, error) {
	// 实际上你应该先检查 Redis 里有没有缓存的 token，如果有直接返回

	appID := u.Config.App.AppID
	appSecret := u.Config.App.AppSecret
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appID, appSecret)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
	}
	json.NewDecoder(resp.Body).Decode(&tokenResp)

	if tokenResp.ErrCode != 0 {
		return "", fmt.Errorf("token error")
	}

	return tokenResp.AccessToken, nil
}

func Code2Session(code string) (*types.WxLoginResponse, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		"wx27c89241d93f10d3",
		"5f8833e38748aa345e3b8e919241d2ce",
		code,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var wxResp types.WxLoginResponse
	if err := json.Unmarshal(body, &wxResp); err != nil {
		return nil, err
	}

	if wxResp.ErrCode != 0 {
		return nil, fmt.Errorf("wx error: %s", wxResp.ErrMsg)
	}

	return &wxResp, nil
}
