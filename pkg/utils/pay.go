package utils

import (
	"fmt"
	"time"
)

func GenerateOutTradeNo(prefix string, orderID int64) string {
	// 时间精确到毫秒
	now := time.Now().Format("20060102150405")
	return fmt.Sprintf("%s%s%d", prefix, now, orderID)
}
