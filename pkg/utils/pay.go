package utils

import (
	"fmt"
	"math/rand/v2"
	"time"
)

func GenerateOutTradeNo(prefix string, orderID int64) string {
	// 时间精确到毫秒
	now := time.Now().Format("20060102150405")
	return fmt.Sprintf("%s%s%d", prefix, now, orderID)
}

// GenerateOrderSn 生成你喜欢的格式：ORD + 20260124153005 + 8899 + 123
func GenerateOrderSn(userId int) string {
	// 1. 获取当前时间 (14位: YYYYMMDDHHMMSS)
	// 如果觉得长，可以用 "060102150405" 缩减到 12 位
	now := time.Now().Format("20060102150405")

	// 2. 取用户ID后4位 (不足4位补0)
	// 这样可以确保同一个用户的订单在数据库物理分布上更趋近，利于分库分表
	userSuffix := fmt.Sprintf("%04d", userId%10000)

	// 3. 生成3位随机数
	// rand.IntN(900) + 100 产生 100-999
	randomNum := rand.IntN(900) + 100

	// 4. 拼接并返回
	// 结果示例: ORD202601241530058899123
	return fmt.Sprintf("%s%s%d", now, userSuffix, randomNum)
}
