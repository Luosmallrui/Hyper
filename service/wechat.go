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
	SendSubscribeMessage(ctx context.Context, req types.WeChatSubscribeMessageRequest) error
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
		return nil, errors.New(wxResp.ErrMsg)
	}

	return &wxResp, nil
}

func (w *WeChatService) GetAccessToken() (string, error) {
	appID := w.Config.App.AppID
	appSecret := w.Config.App.AppSecret
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appID, appSecret)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	json.NewDecoder(resp.Body).Decode(&tokenResp)

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

func (w *WeChatService) SendSubscribeMessage(ctx context.Context, req types.WeChatSubscribeMessageRequest) error {
	accessToken, err := w.GetAccessToken()
	if err != nil {
		return err
	}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=%s", accessToken)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var wxResp struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wxResp); err != nil {
		return err
	}
	if wxResp.ErrCode != 0 {
		return fmt.Errorf("微信订阅消息发送失败: %d %s", wxResp.ErrCode, wxResp.ErrMsg)
	}
	return nil
}
