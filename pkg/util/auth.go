package util

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	AppID     = "你的AppID"
	AppSecret = "你的AppSecret"
)

func GetAccessToken() (string, error) {
	// 伪代码：先查 Redis
	// if token, ok := redis.Get("wx_access_token"); ok { return token, nil }

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", AppID, AppSecret)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	// 伪代码：存入 Redis
	// redis.Set("wx_access_token", result.AccessToken, result.ExpiresIn - 200)

	return result.AccessToken, nil
}
