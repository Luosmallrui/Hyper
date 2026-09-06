package service

import (
	"Hyper/config"
	"Hyper/types"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

var _ IWeChatService = (*WeChatService)(nil)

var (
	// ErrContentUnsafe is deliberately generic so callers never expose moderation details.
	ErrContentUnsafe = errors.New("内容含违规信息")
	// ErrContentSafetyUnavailable prevents unverified user-generated content from being published.
	ErrContentSafetyUnavailable = errors.New("内容安全验证暂不可用")
)

type IWeChatService interface {
	Code2Session(ctx context.Context, code string) (*types.WxLoginResponse, error)
	GetAccessToken() (string, error)
	GetUserPhoneNumber(code string) (string, error)
	CheckTextSecurity(ctx context.Context, content, openID, nickname, title string, scene int) error
	GenerateUnlimitedQRCode(ctx context.Context, scene, page string) ([]byte, error)
	SendSubscribeMessage(ctx context.Context, req types.WeChatSubscribeMessageRequest) error
}

// CheckTextSecurity invokes WeChat msgSecCheck V2 before user-generated text is persisted.
// scene: 1 profile, 2 comment, 3 forum, 4 social feed.
func (w *WeChatService) CheckTextSecurity(ctx context.Context, content, openID, nickname, title string, scene int) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}
	if len([]rune(content)) > 2500 {
		return fmt.Errorf("%w: 文本长度超过内容安全接口限制", ErrContentSafetyUnavailable)
	}
	if strings.TrimSpace(openID) == "" {
		return fmt.Errorf("%w: 当前用户缺少微信 openid", ErrContentSafetyUnavailable)
	}
	if scene < 1 || scene > 4 {
		return fmt.Errorf("%w: 内容场景无效", ErrContentSafetyUnavailable)
	}

	accessToken, err := w.GetAccessToken()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrContentSafetyUnavailable, err)
	}
	body, err := json.Marshal(map[string]any{
		"content":  content,
		"version":  2,
		"openid":   openID,
		"scene":    scene,
		"nickname": strings.TrimSpace(nickname),
		"title":    strings.TrimSpace(title),
	})
	if err != nil {
		return fmt.Errorf("%w: 请求编码失败", ErrContentSafetyUnavailable)
	}
	url := fmt.Sprintf("https://api.weixin.qq.com/wxa/msg_sec_check?access_token=%s", accessToken)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("%w: 请求创建失败", ErrContentSafetyUnavailable)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrContentSafetyUnavailable, err)
	}
	defer resp.Body.Close()

	var wxResp struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		Result  struct {
			Suggest string `json:"suggest"`
		} `json:"result"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&wxResp); err != nil {
		return fmt.Errorf("%w: 响应解析失败", ErrContentSafetyUnavailable)
	}
	if wxResp.ErrCode == 87014 {
		return ErrContentUnsafe
	}
	if wxResp.ErrCode != 0 {
		return fmt.Errorf("%w: 微信返回 %d", ErrContentSafetyUnavailable, wxResp.ErrCode)
	}
	if wxResp.Result.Suggest != "" && wxResp.Result.Suggest != "pass" {
		return ErrContentUnsafe
	}
	return nil
}

type WeChatService struct {
	Config *config.Config
	Redis  *redis.Client
}

// wechatHTTPClient 调用微信开放接口统一使用带超时的 http.Client，避免请求无限挂起
var wechatHTTPClient = &http.Client{Timeout: 10 * time.Second}

// accessTokenGroup 单flight：并发刷新 access_token 时只放行一次真实请求，避免互踢已缓存 token
var accessTokenGroup singleflight.Group

func (w *WeChatService) Code2Session(ctx context.Context, code string) (*types.WxLoginResponse, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		w.Config.App.AppID,
		w.Config.App.AppSecret,
		code,
	)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := wechatHTTPClient.Do(httpReq)
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cacheKey := "wechat:access_token:" + w.Config.App.AppID
	if w.Redis != nil {
		if token, err := w.Redis.Get(ctx, cacheKey).Result(); err == nil && token != "" {
			return token, nil
		}
	}
	// 单flight：并发刷新只放行一次真实请求，避免频繁调用 token 接口互踢
	v, err, _ := accessTokenGroup.Do(cacheKey, func() (any, error) {
		// 拿到刷新权后再查一次缓存，可能已被其他协程/节点刷新
		if w.Redis != nil {
			if token, err := w.Redis.Get(ctx, cacheKey).Result(); err == nil && token != "" {
				return token, nil
			}
		}
		token, expiresIn, err := w.fetchAccessToken(ctx)
		if err != nil {
			return "", err
		}
		if w.Redis != nil {
			// TTL 比微信有效期短 5 分钟，防止用到临界过期 token
			ttl := time.Duration(expiresIn)*time.Second - 5*time.Minute
			if ttl < time.Minute {
				ttl = time.Minute
			}
			if err := w.Redis.Set(ctx, cacheKey, token, ttl).Err(); err != nil {
				return "", fmt.Errorf("缓存 access_token 失败: %w", err)
			}
		}
		return token, nil
	})
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

// fetchAccessToken 向微信请求新的 access_token，返回 token 与有效期（秒）
func (w *WeChatService) fetchAccessToken(ctx context.Context) (string, int, error) {
	appID := w.Config.App.AppID
	appSecret := w.Config.App.AppSecret
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appID, appSecret)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", 0, err
	}
	resp, err := wechatHTTPClient.Do(httpReq)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", 0, err
	}
	if tokenResp.ErrCode != 0 {
		return "", 0, fmt.Errorf("微信 access_token 获取失败: %d %s", tokenResp.ErrCode, tokenResp.ErrMsg)
	}
	if tokenResp.AccessToken == "" {
		return "", 0, errors.New("微信 access_token 为空")
	}

	return tokenResp.AccessToken, tokenResp.ExpiresIn, nil
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

	// 签名无 ctx（兼容现有调用方），通过 http.Client 超时兜底
	httpReq, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := wechatHTTPClient.Do(httpReq)
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

func (w *WeChatService) GenerateUnlimitedQRCode(ctx context.Context, scene, page string) ([]byte, error) {
	accessToken, err := w.GetAccessToken()
	if err != nil {
		return nil, err
	}
	// env_version 可通过 app.qr_code_env_version 配置（develop/trial/release）；
	// 缺省 "trial" 以保持线上现状，小程序正式发布后应配置为 release
	envVersion := "trial"
	if w.Config != nil && w.Config.App != nil && w.Config.App.QRCodeEnvVersion != "" {
		envVersion = w.Config.App.QRCodeEnvVersion
	}
	body, err := json.Marshal(map[string]any{
		"scene":       scene,
		"page":        page,
		"check_path":  false,
		"env_version": envVersion,
		"width":       430,
		"auto_color":  true,
		"is_hyaline":  false,
	})
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=%s", accessToken)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	if bytes.HasPrefix(bytes.TrimSpace(data), []byte("{")) {
		var wxResp struct {
			ErrCode int    `json:"errcode"`
			ErrMsg  string `json:"errmsg"`
		}
		if err := json.Unmarshal(data, &wxResp); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("微信小程序码生成失败: %d %s", wxResp.ErrCode, wxResp.ErrMsg)
	}
	return data, nil
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
