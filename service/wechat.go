package service

import (
	"Hyper/config"
	"Hyper/types"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var _ IWeChatService = (*WeChatService)(nil)

type IWeChatService interface {
	Code2Session(ctx context.Context, code string) (*types.WxLoginResponse, error)
	GetAccessToken() (string, error)
	GetUserPhoneNumber(code string) (string, error)
}

type WeChatService struct {
	Config *config.Config
}

func (w *WeChatService) Code2Session(ctx context.Context, code string) (*types.WxLoginResponse, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		w.Config.App.AppID,
		w.Config.App.AppSecret,
		code,
	)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var wxResp types.WxLoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&wxResp); err != nil {
		return nil, err
	}

	if wxResp.ErrCode != 0 {
		return nil, fmt.Errorf(wxResp.ErrMsg)
	}

	return &wxResp, nil
}

func (w *WeChatService) GetAccessToken() (string, error) {
	// 实际上你应该先检查 Redis 里有没有缓存的 token，如果有直接返回

	appID := w.Config.App.AppID
	appSecret := w.Config.App.AppSecret
	fmt.Println("appID:", appID)
	fmt.Println("appSecret:", appSecret)
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appID, appSecret)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
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
		fmt.Println(tokenResp.ErrCode)
		return "", fmt.Errorf("token error")
	}

	return tokenResp.AccessToken, nil
}

func (w *WeChatService) GetUserPhoneNumber(code string) (string, error) {
	accessToken, err := w.GetAccessToken()
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s",
		accessToken,
	)
	body, _ := json.Marshal(map[string]string{
		"code": code,
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	rep := &types.WxPhoneResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&rep); err != nil {
		fmt.Println(err)
		return "", err
	}

	if rep.ErrCode != 0 {
		fmt.Println(rep.ErrMsg)
		return "", errors.New("微信获取手机号失败")
	}
	return rep.PhoneInfo.PhoneNumber, nil
}
