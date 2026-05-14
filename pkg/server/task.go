package server

import (
	"Hyper/pkg/log"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// StartOrderCancelTask 启动定时取消超时未支付订单的后台任务
// 每分钟检查一次，取消超过 expireMinutes 分钟未支付的订单并回滚库存
func StartOrderCancelTask(db *gorm.DB, expireMinutes int) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			cancelExpiredOrders(db, expireMinutes)
		}
	}()
}

func cancelExpiredOrders(db *gorm.DB, expireMinutes int) {
	deadline := time.Now().Add(-time.Duration(expireMinutes) * time.Minute)

	var orderSns []string
	if err := db.Raw("SELECT order_sn FROM orders WHERE status = 10 AND created_at < ?", deadline).
		Scan(&orderSns).Error; err != nil {
		log.L.Error("查询过期订单失败", zap.Error(err))
		return
	}

	for _, orderSn := range orderSns {
		err := db.Transaction(func(tx *gorm.DB) error {
			// 更新订单状态为已取消(30)
			result := tx.Exec("UPDATE orders SET status = 30 WHERE order_sn = ? AND status = 10", orderSn)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return nil // 已被其他goroutine处理
			}

			// 回滚库存
			if err := tx.Exec(
				"UPDATE products p INNER JOIN order_items oi ON p.id = oi.product_id SET p.stock = p.stock + oi.quantity WHERE oi.order_sn = ?",
				orderSn,
			).Error; err != nil {
				return err
			}

			// 更新支付流水状态为已关闭(4)
			tx.Exec("UPDATE pay_records SET pay_status = 4 WHERE order_sn = ? AND pay_status = 0", orderSn)

			log.L.Info("已取消超时订单", zap.String("order_sn", orderSn))
			return nil
		})
		if err != nil {
			log.L.Error("取消过期订单失败", zap.String("order_sn", orderSn), zap.Error(err))
		}
	}
}
