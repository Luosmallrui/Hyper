package service

import (
	"fmt"
	"strings"
)

const (
	salesChannelWeChat = "wechat"
	salesChannelDouyin = "douyin"
	salesChannelWeb    = "web"
	salesChannelOther  = "other"
)

// normalizeSalesChannel keeps order attribution independent from the payment
// provider. Empty is only valid when creating a legacy-compatible WeChat order.
func NormalizeSalesChannel(value string, defaultWeChat bool) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "":
		if defaultWeChat {
			return salesChannelWeChat, nil
		}
		return "", nil
	case salesChannelWeChat, "wechat_mini_program":
		return salesChannelWeChat, nil
	case salesChannelDouyin, "douyin_mini_program":
		return salesChannelDouyin, nil
	case salesChannelWeb:
		return salesChannelWeb, nil
	case salesChannelOther:
		return salesChannelOther, nil
	default:
		return "", fmt.Errorf("销售渠道参数错误，仅支持 wechat、douyin、web、other")
	}
}
