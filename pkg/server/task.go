package server

import (
	"Hyper/models"
	"Hyper/pkg/log"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// StartOrderCancelTask 启动定时取消超时未支付订单的后台任务
// 每分钟检查一次，取消超过 expireMinutes 分钟未支付的订单并回滚库存
func StartOrderCancelTask(db *gorm.DB, expireMinutes int) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			cancelExpiredOrders(db, expireMinutes)
			cancelExpiredTicketOrders(db)
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

func cancelExpiredTicketOrders(db *gorm.DB) {
	var orders []models.TicketOrder
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("status = ? AND actual_price > 0 AND expire_time <= ?", models.TicketOrderStatusPending, time.Now()).
		Find(&orders).Error; err != nil {
		log.L.Error("查询过期票务订单失败", zap.Error(err))
		return
	}
	for i := range orders {
		order := orders[i]
		err := db.Transaction(func(tx *gorm.DB) error {
			var locked models.TicketOrder
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", order.ID).First(&locked).Error; err != nil {
				return err
			}
			if locked.Status != models.TicketOrderStatusPending {
				return nil
			}
			if err := tx.Model(&models.TicketSpec{}).Where("id = ?", locked.TicketSpecID).
				UpdateColumn("sold_count", gorm.Expr("GREATEST(sold_count - ?, 0)", locked.Quantity)).Error; err != nil {
				return err
			}
			if err := returnTicketOrderPoints(tx, locked); err != nil {
				return err
			}
			if err := tx.Model(&locked).Updates(map[string]any{
				"status":        models.TicketOrderStatusCancelled,
				"cancel_reason": "超时未支付自动取消",
			}).Error; err != nil {
				return err
			}
			tx.Model(&models.PayRecord{}).Where("order_sn = ? AND pay_status = 0", locked.OrderNo).Update("pay_status", 4)
			return nil
		})
		if err != nil {
			log.L.Error("取消过期票务订单失败", zap.String("order_no", order.OrderNo), zap.Error(err))
		}
	}
}

func returnTicketOrderPoints(tx *gorm.DB, order models.TicketOrder) error {
	if order.PointsAmount <= 0 {
		return nil
	}
	var exists int64
	if err := tx.Model(&models.PointsLog{}).
		Where("user_id = ? AND source_id = ? AND change_type = ?", uint64(order.UserID), order.OrderNo, models.TypeOrderRefund).
		Count(&exists).Error; err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}
	var account models.UserPoint
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", uint64(order.UserID)).First(&account).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		account = models.UserPoint{UserID: uint64(order.UserID), Balance: 0}
		if err := tx.Create(&account).Error; err != nil {
			return err
		}
	}
	newBalance := account.Balance + order.PointsAmount
	if err := tx.Model(&models.UserPoint{}).Where("user_id = ?", uint64(order.UserID)).Updates(map[string]any{
		"balance":    newBalance,
		"total_used": gorm.Expr("GREATEST(total_used - ?, 0)", order.PointsAmount),
	}).Error; err != nil {
		return err
	}
	return tx.Create(&models.PointsLog{
		UserID:     uint64(order.UserID),
		Amount:     order.PointsAmount,
		Balance:    newBalance,
		ChangeType: models.TypeOrderRefund,
		SourceID:   order.OrderNo,
		Remark:     "待支付订单超时取消返还积分抵扣",
		Status:     1,
	}).Error
}
