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

// GenerateOrderSn 生成订单号：时间(14位) + 用户ID后4位 + 6位随机数
// 结果示例: 202601241530058899123456（纯数字字符串，长度可变，向后兼容）
func GenerateOrderSn(userId int) string {
	// 1. 获取当前时间 (14位: YYYYMMDDHHMMSS)
	now := time.Now().Format("20060102150405")

	// 2. 取用户ID后4位 (不足4位补0)
	// 这样可以确保同一个用户的订单在数据库物理分布上更趋近，利于分库分表
	userSuffix := fmt.Sprintf("%04d", userId%10000)

	// 3. 生成6位随机数，避免同秒同用户并发下单撞号
	randomNum := rand.IntN(1000000)

	// 4. 拼接并返回
	return fmt.Sprintf("%s%s%06d", now, userSuffix, randomNum)
}
