package service

import (
	"Hyper/models"
	"Hyper/pkg/snowflake"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	platformFlowOrderIncome = "order_income"
	platformFlowRefund      = "refund"
	platformFlowServiceFee  = "service_fee"
	platformFlowSettlement  = "settlement"
	platformFlowWithdraw    = "withdraw"
)

func appendPlatformFlow(tx *gorm.DB, flow models.PlatformFinanceFlow) error {
	if flow.BusinessKey == "" {
		return fmt.Errorf("平台流水业务键不能为空")
	}
	if flow.FlowNo == "" {
		flow.FlowNo = fmt.Sprintf("PF%d", snowflake.GenID())
	}
	if flow.OccurredAt.IsZero() {
		flow.OccurredAt = time.Now()
	}
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "business_key"}},
		DoNothing: true,
	}).Create(&flow).Error
}

func platformFlowOrganizer(tx *gorm.DB, activityID int64) (models.Activity, models.Organizer, error) {
	var activity models.Activity
	if err := tx.First(&activity, activityID).Error; err != nil {
		return activity, models.Organizer{}, err
	}
	var organizer models.Organizer
	if err := tx.First(&organizer, activity.OrganizerID).Error; err != nil {
		return activity, organizer, err
	}
	return activity, organizer, nil
}

func platformServiceFeeRate(tx *gorm.DB, organizerID int64, occurredAt time.Time) (float64, error) {
	var completed int64
	if err := tx.Model(&models.Activity{}).
		Where("organizer_id = ? AND end_time < ?", organizerID, occurredAt).
		Count(&completed).Error; err != nil {
		return 0, err
	}
	_, feeRate, _ := organizerLevelByCompletedCount(completed)
	return feeRate, nil
}

// recordPlatformTicketPayment snapshots the accounting facts at payment time.
func recordPlatformTicketPayment(tx *gorm.DB, order models.TicketOrder, payMethod string) error {
	if order.ActualPrice <= 0 {
		return nil
	}
	activity, organizer, err := platformFlowOrganizer(tx, order.ActivityID)
	if err != nil {
		return err
	}
	occurredAt := time.Now()
	if order.PayTime != nil {
		occurredAt = *order.PayTime
	}
	feeRate, err := platformServiceFeeRate(tx, organizer.ID, occurredAt)
	if err != nil {
		return err
	}
	serviceFee := int64(math.Round(float64(order.ActualPrice) * feeRate))
	settlementAmount := order.ActualPrice - serviceFee
	base := models.PlatformFinanceFlow{
		OrderNo: order.OrderNo, OrganizerID: organizer.ID, OrganizerName: organizer.Name,
		PayMethod: payMethod, OccurredAt: occurredAt,
	}
	if err := appendPlatformFlow(tx, withPlatformFlow(base, "order_income:"+order.OrderNo, platformFlowOrderIncome, "income", order.ActualPrice, "订单票款收入")); err != nil {
		return err
	}
	if err := appendPlatformFlow(tx, withPlatformFlow(base, "service_fee:"+order.OrderNo, platformFlowServiceFee, "income", serviceFee, fmt.Sprintf("活动%d服务费，费率%.2f%%", activity.ID, feeRate*100))); err != nil {
		return err
	}
	return appendPlatformFlow(tx, withPlatformFlow(base, "settlement:"+order.OrderNo, platformFlowSettlement, "expense", settlementAmount, "商家可结算金额"))
}

func recordPlatformTicketRefund(tx *gorm.DB, refund models.Refund, order models.TicketOrder) error {
	if refund.RefundAmount <= 0 {
		return nil
	}
	_, organizer, err := platformFlowOrganizer(tx, order.ActivityID)
	if err != nil {
		return err
	}
	return appendPlatformFlow(tx, models.PlatformFinanceFlow{
		BusinessKey:   "refund:" + refund.RefundNo,
		Type:          platformFlowRefund,
		Direction:     "expense",
		Amount:        refund.RefundAmount,
		OrderNo:       order.OrderNo,
		RefundNo:      refund.RefundNo,
		OrganizerID:   organizer.ID,
		OrganizerName: organizer.Name,
		PayMethod:     order.PayMethod,
		Remark:        "订单退款",
		OccurredAt:    time.Now(),
	})
}

func recordPlatformWithdraw(tx *gorm.DB, withdraw models.OrganizerWithdraw) error {
	var organizer models.Organizer
	if err := tx.First(&organizer, withdraw.OrganizerID).Error; err != nil {
		return err
	}
	return appendPlatformFlow(tx, models.PlatformFinanceFlow{
		BusinessKey:   fmt.Sprintf("withdraw:%d", withdraw.ID),
		Type:          platformFlowWithdraw,
		Direction:     "expense",
		Amount:        withdraw.Amount,
		WithdrawID:    withdraw.ID,
		OrganizerID:   organizer.ID,
		OrganizerName: organizer.Name,
		Remark:        "商家提现审核通过",
		OccurredAt:    time.Now(),
	})
}

func withPlatformFlow(base models.PlatformFinanceFlow, businessKey, flowType, direction string, amount int64, remark string) models.PlatformFinanceFlow {
	base.BusinessKey = businessKey
	base.Type = flowType
	base.Direction = direction
	base.Amount = amount
	base.Remark = remark
	return base
}
