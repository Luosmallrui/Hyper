package service

import (
	"Hyper/models"
	"Hyper/types"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ITicketingService interface {
	GetOrganizerInfo(ctx context.Context, userID int64) (*types.OrganizerInfoResponse, error)
	ApplyOrganizer(ctx context.Context, userID int64, req types.OrganizerApplyRequest) error
	UpdateOrganizerBasic(ctx context.Context, userID int64, req types.OrganizerBasicRequest) error
	UpdateWithdrawInfo(ctx context.Context, userID int64, req types.OrganizerWithdrawRequest) error
	GetActivity(ctx context.Context, activityID int64) (*types.ActivityDetailResponse, error)
	SaveActivityStep(ctx context.Context, userID int64, req types.ActivityCreateRequest) (int64, error)
	GetMyActivities(ctx context.Context, userID int64, page, size int) (*types.PageResponse[types.ActivityListItem], error)
	SearchActivities(ctx context.Context, keyword string) ([]types.ActivityListItem, error)
	DeleteActivity(ctx context.Context, userID, activityID int64) error
	SubmitActivityAudit(ctx context.Context, userID, activityID int64) error
	GetTicketSpecs(ctx context.Context, activityID int64) ([]models.TicketSpec, error)
	SaveTicketSpecs(ctx context.Context, userID, activityID int64, specs []types.TicketSpecSaveItem) error
	DeleteTicketSpec(ctx context.Context, userID, specID int64) error
	CreateTicketOrder(ctx context.Context, userID int64, req types.CreateTicketOrderRequest) (string, error)
	GetTicketOrderDetail(ctx context.Context, userID int64, orderNo string) (*types.TicketOrderDetailResponse, error)
	ListCancelReasons(ctx context.Context) ([]models.CancelReason, error)
	CancelTicketOrder(ctx context.Context, userID int64, orderNo string, reasonID int64) error
	ListRefundReasons(ctx context.Context) ([]models.RefundReason, error)
	ApplyRefund(ctx context.Context, userID int64, req types.ApplyRefundRequest) (string, error)
	ListOrganizerRefunds(ctx context.Context, userID int64, status *int8, page, size int) (*types.PageResponse[types.OrganizerRefundListItem], error)
	GetRefundDetail(ctx context.Context, userID int64, refundNo string) (*types.RefundDetailResponse, error)
	RejectRefund(ctx context.Context, userID int64, refundNo string, req types.RejectRefundRequest) error
	CancelRefund(ctx context.Context, userID int64, refundNo string) error
	ListVerifiers(ctx context.Context, userID int64, page, size int) (*types.PageResponse[models.Verifier], error)
	AddVerifier(ctx context.Context, userID int64, req types.VerifierRequest) error
	DeleteVerifier(ctx context.Context, userID, verifierID int64) error
	GetVerifierActivationQR(ctx context.Context, userID, verifierID int64) (map[string]string, error)
	ActivateVerifier(ctx context.Context, req types.ActivateVerifierRequest) error
	ScanOrder(ctx context.Context, req types.ScanOrderRequest) (*types.ScanOrderResponse, error)
	ConfirmVerify(ctx context.Context, verifierID int64, req types.ConfirmVerifyRequest) error
	ListVerified(ctx context.Context, verifierID int64, page, size int) (*types.PageResponse[types.VerifiedListItem], error)
}

type TicketingService struct {
	DB *gorm.DB
}

var _ ITicketingService = (*TicketingService)(nil)

func (s *TicketingService) GetOrganizerInfo(ctx context.Context, userID int64) (*types.OrganizerInfoResponse, error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return buildOrganizerInfo(org), nil
}

func (s *TicketingService) ApplyOrganizer(ctx context.Context, userID int64, req types.OrganizerApplyRequest) error {
	var org models.Organizer
	err := s.DB.WithContext(ctx).Where("user_id = ?", userID).First(&org).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	data := models.Organizer{
		UserID:   userID,
		Name:     req.Name,
		Logo:     req.Logo,
		Status:   models.OrganizerStatusAuditing,
		Level:    "LV1",
		Province: req.Province,
		City:     req.City,
		District: req.District,
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.DB.WithContext(ctx).Create(&data).Error
	}
	return s.DB.WithContext(ctx).Model(&org).Updates(data).Error
}

func (s *TicketingService) UpdateOrganizerBasic(ctx context.Context, userID int64, req types.OrganizerBasicRequest) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	updates := map[string]any{}
	putString(updates, "name", req.Name)
	putString(updates, "logo", req.Logo)
	putString(updates, "province", req.Province)
	putString(updates, "city", req.City)
	putString(updates, "district", req.District)
	if len(updates) == 0 {
		return nil
	}
	return s.DB.WithContext(ctx).Model(org).Updates(updates).Error
}

func (s *TicketingService) UpdateWithdrawInfo(ctx context.Context, userID int64, req types.OrganizerWithdrawRequest) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	return s.DB.WithContext(ctx).Model(org).Updates(map[string]any{
		"bank_account_name": req.BankAccountName,
		"bank_account_no":   req.BankAccountNo,
		"bank_name":         req.BankName,
	}).Error
}

func (s *TicketingService) SaveActivityStep(ctx context.Context, userID int64, req types.ActivityCreateRequest) (int64, error) {
	org, err := s.ensureOrganizer(ctx, userID)
	if err != nil {
		return 0, err
	}
	var act models.Activity
	if req.ActivityID > 0 {
		if err := s.DB.WithContext(ctx).Where("id = ? AND organizer_id = ?", req.ActivityID, org.ID).First(&act).Error; err != nil {
			return 0, err
		}
	} else {
		act = models.Activity{
			OrganizerID: org.ID,
			Name:        "未命名活动",
			StartTime:   time.Now(),
			EndTime:     time.Now(),
			Status:      models.ActivityStatusDraft,
		}
		if err := s.DB.WithContext(ctx).Create(&act).Error; err != nil {
			return 0, err
		}
	}
	if req.Step == 4 {
		if err := s.SaveTicketSpecs(ctx, userID, act.ID, req.TicketSpecs); err != nil {
			return 0, err
		}
	}
	updates, err := activityUpdates(req)
	if err != nil {
		return 0, err
	}
	if len(updates) > 0 {
		if err := s.DB.WithContext(ctx).Model(&act).Updates(updates).Error; err != nil {
			return 0, err
		}
	}
	return act.ID, nil
}

func (s *TicketingService) GetActivity(ctx context.Context, activityID int64) (*types.ActivityDetailResponse, error) {
	var act models.Activity
	if err := s.DB.WithContext(ctx).First(&act, activityID).Error; err != nil {
		return nil, err
	}
	var specs []models.TicketSpec
	if err := s.DB.WithContext(ctx).Where("activity_id = ?", activityID).Order("id asc").Find(&specs).Error; err != nil {
		return nil, err
	}
	var org models.Organizer
	_ = s.DB.WithContext(ctx).First(&org, act.OrganizerID).Error
	return &types.ActivityDetailResponse{Activity: act, TicketSpecs: specs, Organizer: &org}, nil
}

func (s *TicketingService) GetMyActivities(ctx context.Context, userID int64, page, size int) (*types.PageResponse[types.ActivityListItem], error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &types.PageResponse[types.ActivityListItem]{List: []types.ActivityListItem{}}, nil
		}
		return nil, err
	}
	query := s.DB.WithContext(ctx).Model(&models.Activity{}).Where("organizer_id = ?", org.ID)
	return s.listActivities(query, page, size)
}

func (s *TicketingService) SearchActivities(ctx context.Context, keyword string) ([]types.ActivityListItem, error) {
	query := s.DB.WithContext(ctx).Model(&models.Activity{}).Where("status = ?", models.ActivityStatusOnline)
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}
	resp, err := s.listActivities(query, 1, 50)
	if err != nil {
		return nil, err
	}
	return resp.List, nil
}

func (s *TicketingService) DeleteActivity(ctx context.Context, userID, activityID int64) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	return s.DB.WithContext(ctx).Where("id = ? AND organizer_id = ?", activityID, org.ID).Delete(&models.Activity{}).Error
}

func (s *TicketingService) SubmitActivityAudit(ctx context.Context, userID, activityID int64) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	return s.DB.WithContext(ctx).Model(&models.Activity{}).
		Where("id = ? AND organizer_id = ? AND status IN ?", activityID, org.ID, []int8{models.ActivityStatusDraft, models.ActivityStatusRejected}).
		Update("status", models.ActivityStatusPending).Error
}

func (s *TicketingService) GetTicketSpecs(ctx context.Context, activityID int64) ([]models.TicketSpec, error) {
	var specs []models.TicketSpec
	err := s.DB.WithContext(ctx).Where("activity_id = ?", activityID).Order("id asc").Find(&specs).Error
	return specs, err
}

func (s *TicketingService) SaveTicketSpecs(ctx context.Context, userID, activityID int64, specs []types.TicketSpecSaveItem) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	if err := s.DB.WithContext(ctx).Where("id = ? AND organizer_id = ?", activityID, org.ID).First(&models.Activity{}).Error; err != nil {
		return err
	}
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range specs {
			spec, err := buildTicketSpec(activityID, item)
			if err != nil {
				return err
			}
			if item.ID > 0 {
				if err := tx.Model(&models.TicketSpec{}).Where("id = ? AND activity_id = ?", item.ID, activityID).Updates(spec).Error; err != nil {
					return err
				}
				continue
			}
			if err := tx.Create(spec).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *TicketingService) DeleteTicketSpec(ctx context.Context, userID, specID int64) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	return s.DB.WithContext(ctx).
		Where("id = ? AND activity_id IN (?)", specID, s.DB.Model(&models.Activity{}).Select("id").Where("organizer_id = ?", org.ID)).
		Delete(&models.TicketSpec{}).Error
}

func (s *TicketingService) CreateTicketOrder(ctx context.Context, userID int64, req types.CreateTicketOrderRequest) (string, error) {
	orderNo := "T" + time.Now().Format("20060102150405") + randomHex(4)
	expireTime := time.Now().Add(15 * time.Minute)
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var act models.Activity
		if err := tx.First(&act, req.ActivityID).Error; err != nil {
			return err
		}
		if act.Status != models.ActivityStatusOnline {
			return errors.New("活动未上架")
		}
		var spec models.TicketSpec
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND activity_id = ? AND is_enabled = 1", req.TicketSpecID, req.ActivityID).
			First(&spec).Error; err != nil {
			return err
		}
		now := time.Now()
		if !spec.SaleStart.IsZero() && now.Before(spec.SaleStart) {
			return errors.New("票券未开售")
		}
		if !spec.SaleEnd.IsZero() && now.After(spec.SaleEnd) {
			return errors.New("票券已停售")
		}
		if req.Quantity > spec.PurchaseLimit && spec.PurchaseLimit > 0 {
			return errors.New("超过限购数量")
		}
		if spec.Stock-spec.SoldCount < req.Quantity {
			return errors.New("库存不足")
		}
		if act.RealNameMode == 1 && (req.BuyerName == "" || req.BuyerIDCard == "") {
			return errors.New("实名模式下需要填写购票人姓名和身份证")
		}
		if act.MinorCheck == 1 && isMinor(req.BuyerIDCard) {
			return errors.New("未成年人不可购票")
		}
		if err := tx.Model(&spec).UpdateColumn("sold_count", gorm.Expr("sold_count + ?", req.Quantity)).Error; err != nil {
			return err
		}
		order := models.TicketOrder{
			OrderNo:      orderNo,
			UserID:       userID,
			ActivityID:   req.ActivityID,
			TicketSpecID: req.TicketSpecID,
			Quantity:     req.Quantity,
			TotalPrice:   spec.Price * int64(req.Quantity),
			ActualPrice:  spec.Price * int64(req.Quantity),
			BuyerName:    req.BuyerName,
			BuyerIDCard:  req.BuyerIDCard,
			Status:       models.TicketOrderStatusPending,
			ExpireTime:   expireTime,
			QRCode:       "TICKET:" + orderNo + ":" + randomHex(8),
		}
		return tx.Create(&order).Error
	})
	return orderNo, err
}

func (s *TicketingService) GetTicketOrderDetail(ctx context.Context, userID int64, orderNo string) (*types.TicketOrderDetailResponse, error) {
	var order models.TicketOrder
	if err := s.DB.WithContext(ctx).Where("order_no = ? AND user_id = ?", orderNo, userID).First(&order).Error; err != nil {
		return nil, err
	}
	return s.buildOrderDetail(ctx, order)
}

func (s *TicketingService) ListCancelReasons(ctx context.Context) ([]models.CancelReason, error) {
	var reasons []models.CancelReason
	err := s.DB.WithContext(ctx).Order("sort asc,id asc").Find(&reasons).Error
	return reasons, err
}

func (s *TicketingService) CancelTicketOrder(ctx context.Context, userID int64, orderNo string, reasonID int64) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order models.TicketOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_no = ? AND user_id = ?", orderNo, userID).First(&order).Error; err != nil {
			return err
		}
		if order.Status != models.TicketOrderStatusPending {
			return errors.New("当前订单不可取消")
		}
		var reason models.CancelReason
		_ = tx.First(&reason, reasonID).Error
		if err := tx.Model(&models.TicketSpec{}).Where("id = ?", order.TicketSpecID).
			UpdateColumn("sold_count", gorm.Expr("GREATEST(sold_count - ?, 0)", order.Quantity)).Error; err != nil {
			return err
		}
		return tx.Model(&order).Updates(map[string]any{"status": models.TicketOrderStatusCancelled, "cancel_reason": reason.Reason}).Error
	})
}

func (s *TicketingService) ListRefundReasons(ctx context.Context) ([]models.RefundReason, error) {
	var reasons []models.RefundReason
	err := s.DB.WithContext(ctx).Order("sort asc,id asc").Find(&reasons).Error
	return reasons, err
}

func (s *TicketingService) ApplyRefund(ctx context.Context, userID int64, req types.ApplyRefundRequest) (string, error) {
	refundNo := "R" + time.Now().Format("20060102150405") + randomHex(4)
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order models.TicketOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_no = ? AND user_id = ?", req.OrderNo, userID).First(&order).Error; err != nil {
			return err
		}
		if order.Status != models.TicketOrderStatusUsable {
			return errors.New("当前订单不可退款")
		}
		var count int64
		if err := tx.Model(&models.Refund{}).Where("order_id = ? AND status IN ?", order.ID, []int8{0, 1}).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("该订单已有退款处理中")
		}
		var reason models.RefundReason
		_ = tx.First(&reason, req.ReasonID).Error
		refund := models.Refund{
			OrderID:          order.ID,
			RefundNo:         refundNo,
			RefundAmount:     order.ActualPrice,
			Reason:           reason.Reason,
			Status:           models.RefundStatusAuditing,
			ExpectArriveDate: time.Now().AddDate(0, 0, 3).Format("2006-01-02"),
		}
		if err := tx.Create(&refund).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.RefundLog{RefundID: refund.ID, Status: "审核中", Description: "退款申请已提交"}).Error; err != nil {
			return err
		}
		return tx.Model(&order).Update("status", models.TicketOrderStatusRefunding).Error
	})
	return refundNo, err
}

func (s *TicketingService) ListOrganizerRefunds(ctx context.Context, userID int64, status *int8, page, size int) (*types.PageResponse[types.OrganizerRefundListItem], error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	page, size = normalizePage(page, size)
	query := s.DB.WithContext(ctx).Table("refunds r").
		Joins("JOIN ticket_orders o ON o.id = r.order_id").
		Joins("JOIN activities a ON a.id = o.activity_id").
		Joins("JOIN ticket_specs ts ON ts.id = o.ticket_spec_id").
		Where("a.organizer_id = ?", org.ID)
	if status != nil {
		query = query.Where("r.status = ?", *status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []struct {
		RefundNo         string
		Status           int8
		RefundAmount     int64
		DeductAmount     int64
		Reason           string
		RejectReason     string
		ExpectArriveDate string
		WechatRefundID   string
		WechatStatus     string
		OrderNo          string
		UserID           int64
		BuyerName        string
		BuyerIDCard      string
		ActivityID       int64
		ActivityName     string
		TicketSpecID     int64
		TicketSpecName   string
		Quantity         int
		CreatedAt        time.Time
	}
	if err := query.Select(`
			r.refund_no,
			r.status,
			r.refund_amount,
			r.deduct_amount,
			r.reason,
			r.reject_reason,
			r.expect_arrive_date,
			r.wechat_refund_id,
			r.wechat_status,
			o.order_no,
			o.user_id,
			o.buyer_name,
			o.buyer_id_card,
			o.activity_id,
			a.name AS activity_name,
			o.ticket_spec_id,
			ts.name AS ticket_spec_name,
			o.quantity,
			r.created_at
		`).
		Order("r.id desc").
		Offset((page - 1) * size).
		Limit(size).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]types.OrganizerRefundListItem, 0, len(rows))
	for _, r := range rows {
		list = append(list, types.OrganizerRefundListItem{
			RefundNo:         r.RefundNo,
			Status:           r.Status,
			RefundAmount:     r.RefundAmount,
			DeductAmount:     r.DeductAmount,
			Reason:           r.Reason,
			RejectReason:     r.RejectReason,
			ExpectArriveDate: r.ExpectArriveDate,
			WechatRefundID:   r.WechatRefundID,
			WechatStatus:     r.WechatStatus,
			OrderNo:          r.OrderNo,
			UserID:           r.UserID,
			BuyerName:        r.BuyerName,
			BuyerIDCard:      r.BuyerIDCard,
			ActivityID:       r.ActivityID,
			ActivityName:     r.ActivityName,
			TicketSpecID:     r.TicketSpecID,
			TicketSpecName:   r.TicketSpecName,
			Quantity:         r.Quantity,
			CreatedAt:        r.CreatedAt,
		})
	}
	return &types.PageResponse[types.OrganizerRefundListItem]{List: list, Total: total}, nil
}

func (s *TicketingService) GetRefundDetail(ctx context.Context, userID int64, refundNo string) (*types.RefundDetailResponse, error) {
	var refund models.Refund
	if err := s.DB.WithContext(ctx).Where("refund_no = ?", refundNo).First(&refund).Error; err != nil {
		return nil, err
	}
	var order models.TicketOrder
	if err := s.DB.WithContext(ctx).Where("id = ? AND user_id = ?", refund.OrderID, userID).First(&order).Error; err != nil {
		return nil, err
	}
	var activity models.Activity
	_ = s.DB.WithContext(ctx).First(&activity, order.ActivityID).Error
	var logs []models.RefundLog
	_ = s.DB.WithContext(ctx).Where("refund_id = ?", refund.ID).Order("id asc").Find(&logs).Error
	resp := &types.RefundDetailResponse{
		RefundNo:         refund.RefundNo,
		Status:           refund.Status,
		RefundAmount:     refund.RefundAmount,
		DeductAmount:     refund.DeductAmount,
		ExpectArriveDate: refund.ExpectArriveDate,
		Progress:         logs,
	}
	resp.Order.OrderNo = order.OrderNo
	resp.Order.TotalPrice = order.TotalPrice
	resp.Order.ActualPrice = order.ActualPrice
	resp.Order.Quantity = order.Quantity
	resp.Activity.Name = activity.Name
	resp.Activity.StartTime = activity.StartTime
	resp.Activity.EndTime = activity.EndTime
	return resp, nil
}

func (s *TicketingService) RejectRefund(ctx context.Context, userID int64, refundNo string, req types.RejectRefundRequest) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var refund models.Refund
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("refund_no = ?", refundNo).First(&refund).Error; err != nil {
			return err
		}
		if refund.Status != models.RefundStatusAuditing {
			return errors.New("当前退款状态不可拒绝")
		}
		var order models.TicketOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", refund.OrderID).First(&order).Error; err != nil {
			return err
		}
		var activity models.Activity
		if err := tx.Where("id = ? AND organizer_id = ?", order.ActivityID, org.ID).First(&activity).Error; err != nil {
			return errors.New("无权审核该退款")
		}
		if err := tx.Model(&refund).Updates(map[string]any{
			"status":        models.RefundStatusRejected,
			"reject_reason": req.RejectReason,
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&order).Update("status", models.TicketOrderStatusRefundReject).Error; err != nil {
			return err
		}
		return tx.Create(&models.RefundLog{
			RefundID:    refund.ID,
			Status:      "退款拒绝",
			Description: req.RejectReason,
		}).Error
	})
}

func (s *TicketingService) CancelRefund(ctx context.Context, userID int64, refundNo string) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var refund models.Refund
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("refund_no = ?", refundNo).First(&refund).Error; err != nil {
			return err
		}
		var order models.TicketOrder
		if err := tx.Where("id = ? AND user_id = ?", refund.OrderID, userID).First(&order).Error; err != nil {
			return err
		}
		if refund.Status != models.RefundStatusAuditing {
			return errors.New("当前退款不可取消")
		}
		if err := tx.Model(&refund).Update("status", models.RefundStatusRejected).Error; err != nil {
			return err
		}
		return tx.Model(&order).Update("status", models.TicketOrderStatusUsable).Error
	})
}

func (s *TicketingService) ListVerifiers(ctx context.Context, userID int64, page, size int) (*types.PageResponse[models.Verifier], error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	query := s.DB.WithContext(ctx).Model(&models.Verifier{}).Where("organizer_id = ?", org.ID)
	page, size = normalizePage(page, size)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []models.Verifier
	err = query.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return &types.PageResponse[models.Verifier]{List: list, Total: total}, err
}

func (s *TicketingService) AddVerifier(ctx context.Context, userID int64, req types.VerifierRequest) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	return s.DB.WithContext(ctx).Create(&models.Verifier{
		OrganizerID: org.ID,
		Name:        req.Name,
		Phone:       req.Phone,
		Status:      models.VerifierStatusInactive,
	}).Error
}

func (s *TicketingService) DeleteVerifier(ctx context.Context, userID, verifierID int64) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	return s.DB.WithContext(ctx).Where("id = ? AND organizer_id = ?", verifierID, org.ID).Delete(&models.Verifier{}).Error
}

func (s *TicketingService) GetVerifierActivationQR(ctx context.Context, userID, verifierID int64) (map[string]string, error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.DB.WithContext(ctx).Where("id = ? AND organizer_id = ?", verifierID, org.ID).First(&models.Verifier{}).Error; err != nil {
		return nil, err
	}
	return map[string]string{
		"wechat_qr": fmt.Sprintf("hyper://verifier/activate?id=%d&channel=wechat", verifierID),
		"douyin_qr": fmt.Sprintf("hyper://verifier/activate?id=%d&channel=douyin", verifierID),
	}, nil
}

func (s *TicketingService) ActivateVerifier(ctx context.Context, req types.ActivateVerifierRequest) error {
	return s.DB.WithContext(ctx).Model(&models.Verifier{}).
		Where("phone = ?", req.Phone).
		Updates(map[string]any{"status": models.VerifierStatusActive, "channel": req.Channel}).Error
}

func (s *TicketingService) ScanOrder(ctx context.Context, req types.ScanOrderRequest) (*types.ScanOrderResponse, error) {
	var order models.TicketOrder
	if err := s.DB.WithContext(ctx).Where("qr_code = ?", req.QRCode).First(&order).Error; err != nil {
		return &types.ScanOrderResponse{Success: false, ErrorCode: "ORDER_NOT_FOUND"}, nil
	}
	if order.ActivityID != req.ActivityID {
		return &types.ScanOrderResponse{Success: false, ErrorCode: "WRONG_ACTIVITY"}, nil
	}
	if order.Status == models.TicketOrderStatusUsed {
		return &types.ScanOrderResponse{Success: false, ErrorCode: "ALREADY_VERIFIED"}, nil
	}
	if order.Status == models.TicketOrderStatusCancelled || order.Status == models.TicketOrderStatusRefundSuccess {
		return &types.ScanOrderResponse{Success: false, ErrorCode: "ORDER_CANCELLED"}, nil
	}
	if order.Status != models.TicketOrderStatusUsable {
		return &types.ScanOrderResponse{Success: false, ErrorCode: "INVALID_QR"}, nil
	}
	var activity models.Activity
	_ = s.DB.WithContext(ctx).First(&activity, order.ActivityID).Error
	if !activity.StartTime.IsZero() && time.Now().Before(activity.StartTime.Add(-24*time.Hour)) {
		return &types.ScanOrderResponse{Success: false, ErrorCode: "NOT_VERIFIABLE_TIME"}, nil
	}
	var spec models.TicketSpec
	_ = s.DB.WithContext(ctx).First(&spec, order.TicketSpecID).Error
	item := struct {
		ActivityName      string `json:"activity_name"`
		TicketSpecName    string `json:"ticket_spec_name"`
		Quantity          int    `json:"quantity"`
		BuyerNameMasked   string `json:"buyer_name_masked"`
		BuyerIDCardMasked string `json:"buyer_id_card_masked"`
	}{
		ActivityName:      activity.Name,
		TicketSpecName:    spec.Name,
		Quantity:          order.Quantity,
		BuyerNameMasked:   maskName(order.BuyerName),
		BuyerIDCardMasked: maskIDCard(order.BuyerIDCard),
	}
	return &types.ScanOrderResponse{Success: true, Order: &item}, nil
}

func (s *TicketingService) ConfirmVerify(ctx context.Context, verifierID int64, req types.ConfirmVerifyRequest) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var verifier models.Verifier
		if err := tx.First(&verifier, verifierID).Error; err != nil {
			return err
		}
		if verifier.Status != models.VerifierStatusActive {
			return errors.New("核销员未激活")
		}
		var order models.TicketOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_no = ?", req.OrderNo).First(&order).Error; err != nil {
			return err
		}
		if order.Status != models.TicketOrderStatusUsable {
			return errors.New("订单不可核销")
		}
		if err := tx.Model(&order).Update("status", models.TicketOrderStatusUsed).Error; err != nil {
			return err
		}
		return tx.Create(&models.VerificationRecord{
			OrderID:    order.ID,
			VerifierID: verifier.ID,
			ActivityID: order.ActivityID,
			VerifiedAt: time.Now(),
		}).Error
	})
}

func (s *TicketingService) ListVerified(ctx context.Context, verifierID int64, page, size int) (*types.PageResponse[types.VerifiedListItem], error) {
	page, size = normalizePage(page, size)
	var total int64
	base := s.DB.WithContext(ctx).Table("verification_records vr").
		Joins("JOIN ticket_orders o ON o.id = vr.order_id").
		Joins("JOIN activities a ON a.id = vr.activity_id").
		Joins("JOIN ticket_specs ts ON ts.id = o.ticket_spec_id").
		Where("vr.verifier_id = ?", verifierID)
	if err := base.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []struct {
		ActivityName   string
		TicketSpecName string
		Quantity       int
		BuyerName      string
		BuyerIDCard    string
		VerifiedAt     time.Time
	}
	if err := base.Select("a.name AS activity_name, ts.name AS ticket_spec_name, o.quantity, o.buyer_name, o.buyer_id_card, vr.verified_at").
		Order("vr.id desc").Offset((page - 1) * size).Limit(size).Scan(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]types.VerifiedListItem, 0, len(rows))
	for _, r := range rows {
		list = append(list, types.VerifiedListItem{
			ActivityName:      r.ActivityName,
			TicketSpecName:    r.TicketSpecName,
			Quantity:          r.Quantity,
			BuyerNameMasked:   maskName(r.BuyerName),
			BuyerIDCardMasked: maskIDCard(r.BuyerIDCard),
			VerifiedAt:        r.VerifiedAt,
		})
	}
	return &types.PageResponse[types.VerifiedListItem]{List: list, Total: total}, nil
}

func (s *TicketingService) findOrganizerByUser(ctx context.Context, userID int64) (*models.Organizer, error) {
	var org models.Organizer
	err := s.DB.WithContext(ctx).Where("user_id = ?", userID).First(&org).Error
	return &org, err
}

func (s *TicketingService) ensureOrganizer(ctx context.Context, userID int64) (*models.Organizer, error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err == nil {
		return org, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	org = &models.Organizer{UserID: userID, Name: "未命名主办方", Status: models.OrganizerStatusPending, Level: "LV1"}
	if err := s.DB.WithContext(ctx).Create(org).Error; err != nil {
		return nil, err
	}
	return org, nil
}

func (s *TicketingService) listActivities(query *gorm.DB, page, size int) (*types.PageResponse[types.ActivityListItem], error) {
	page, size = normalizePage(page, size)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var acts []models.Activity
	if err := query.Order("created_at desc").Offset((page - 1) * size).Limit(size).Find(&acts).Error; err != nil {
		return nil, err
	}
	list := make([]types.ActivityListItem, 0, len(acts))
	for _, a := range acts {
		list = append(list, types.ActivityListItem{ID: a.ID, Name: a.Name, PosterList: a.PosterList, StartTime: a.StartTime, EndTime: a.EndTime, Status: a.Status})
	}
	return &types.PageResponse[types.ActivityListItem]{List: list, Total: total}, nil
}

func (s *TicketingService) buildOrderDetail(ctx context.Context, order models.TicketOrder) (*types.TicketOrderDetailResponse, error) {
	var act models.Activity
	if err := s.DB.WithContext(ctx).First(&act, order.ActivityID).Error; err != nil {
		return nil, err
	}
	var spec models.TicketSpec
	if err := s.DB.WithContext(ctx).First(&spec, order.TicketSpecID).Error; err != nil {
		return nil, err
	}
	resp := &types.TicketOrderDetailResponse{
		OrderNo:     order.OrderNo,
		Status:      order.Status,
		TotalPrice:  order.TotalPrice,
		ActualPrice: order.ActualPrice,
		Quantity:    order.Quantity,
		BuyerName:   order.BuyerName,
		BuyerIDCard: order.BuyerIDCard,
		PayMethod:   order.PayMethod,
		PayTime:     order.PayTime,
		CreatedAt:   order.CreatedAt,
		QRCode:      order.QRCode,
		ExpireTime:  order.ExpireTime,
	}
	resp.Activity.ID = act.ID
	resp.Activity.Name = act.Name
	resp.Activity.StartTime = act.StartTime
	resp.Activity.EndTime = act.EndTime
	resp.Activity.PosterList = act.PosterList
	resp.TicketSpec.Name = spec.Name
	var refund models.Refund
	if err := s.DB.WithContext(ctx).Where("order_id = ?", order.ID).Order("id desc").First(&refund).Error; err == nil {
		resp.RefundInfo = &struct {
			RefundAmount     int64  `json:"refund_amount"`
			Status           int8   `json:"status"`
			ExpectArriveDate string `json:"expect_arrive_date"`
		}{RefundAmount: refund.RefundAmount, Status: refund.Status, ExpectArriveDate: refund.ExpectArriveDate}
	}
	return resp, nil
}

func buildOrganizerInfo(org *models.Organizer) *types.OrganizerInfoResponse {
	resp := &types.OrganizerInfoResponse{
		ID:             org.ID,
		Name:           org.Name,
		Logo:           org.Logo,
		Status:         org.Status,
		RejectReason:   org.RejectReason,
		Level:          org.Level,
		ServiceFeeRate: org.ServiceFeeRate,
		JoinDays:       int(math.Max(0, time.Since(org.CreatedAt).Hours()/24)),
	}
	resp.BasicInfo.Province = org.Province
	resp.BasicInfo.City = org.City
	resp.BasicInfo.District = org.District
	resp.AccountInfo.BankAccountName = org.BankAccountName
	resp.AccountInfo.BankAccountNo = org.BankAccountNo
	resp.AccountInfo.BankName = org.BankName
	return resp
}

func activityUpdates(req types.ActivityCreateRequest) (map[string]any, error) {
	updates := map[string]any{}
	putString(updates, "name", req.Name)
	putString(updates, "share_title", req.ShareTitle)
	putInt8(updates, "real_name_mode", req.RealNameMode)
	putInt8(updates, "minor_check", req.MinorCheck)
	putString(updates, "description", req.Description)
	putString(updates, "province", req.Province)
	putString(updates, "city", req.City)
	putString(updates, "district", req.District)
	putString(updates, "address", req.Address)
	putFloat64(updates, "latitude", req.Latitude)
	putFloat64(updates, "longitude", req.Longitude)
	putString(updates, "poster_detail", req.PosterDetail)
	putString(updates, "poster_long", req.PosterLong)
	putString(updates, "poster_list", req.PosterList)
	putString(updates, "poster_wechat", req.PosterWechat)
	putString(updates, "qualification_doc", req.QualificationDoc)
	if req.StartTime != nil {
		t, err := parseDatetime(*req.StartTime)
		if err != nil {
			return nil, fmt.Errorf("开始时间格式错误")
		}
		updates["start_time"] = t
	}
	if req.EndTime != nil {
		t, err := parseDatetime(*req.EndTime)
		if err != nil {
			return nil, fmt.Errorf("结束时间格式错误")
		}
		updates["end_time"] = t
	}
	return updates, nil
}

func buildTicketSpec(activityID int64, item types.TicketSpecSaveItem) (*models.TicketSpec, error) {
	start, err := parseDatetime(item.SaleStart)
	if err != nil {
		return nil, fmt.Errorf("开售时间格式错误")
	}
	end, err := parseDatetime(item.SaleEnd)
	if err != nil {
		return nil, fmt.Errorf("停售时间格式错误")
	}
	if item.PurchaseLimit <= 0 {
		item.PurchaseLimit = 1
	}
	if item.MaxAttendees <= 0 {
		item.MaxAttendees = 1
	}
	return &models.TicketSpec{
		ActivityID:    activityID,
		Name:          item.Name,
		IsEnabled:     item.IsEnabled,
		SaleStart:     start,
		SaleEnd:       end,
		Price:         item.Price,
		Stock:         item.Stock,
		PurchaseLimit: item.PurchaseLimit,
		MaxAttendees:  item.MaxAttendees,
	}, nil
}

func parseDatetime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	layouts := []string{time.RFC3339, "2006-01-02T15:04", "2006-01-02 15:04:05", "2006-01-02 15:04"}
	var last error
	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			return t, nil
		}
		last = err
	}
	return time.Time{}, last
}

func normalizePage(page, size int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	return page, size
}

func putString(updates map[string]any, key string, value *string) {
	if value != nil {
		updates[key] = *value
	}
}

func putInt8(updates map[string]any, key string, value *int8) {
	if value != nil {
		updates[key] = *value
	}
}

func putFloat64(updates map[string]any, key string, value *float64) {
	if value != nil {
		updates[key] = *value
	}
}

func randomHex(size int) string {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}

func isMinor(idCard string) bool {
	if len(idCard) < 14 {
		return false
	}
	birth, err := time.Parse("20060102", idCard[6:14])
	if err != nil {
		return false
	}
	return time.Now().Before(birth.AddDate(18, 0, 0))
}

func maskName(name string) string {
	if name == "" {
		return ""
	}
	r := []rune(name)
	if len(r) <= 1 {
		return "*"
	}
	return string(r[0]) + strings.Repeat("*", len(r)-1)
}

func maskIDCard(card string) string {
	if len(card) <= 8 {
		return card
	}
	return card[:4] + strings.Repeat("*", len(card)-8) + card[len(card)-4:]
}
