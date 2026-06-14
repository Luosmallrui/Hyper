package service

import (
	"Hyper/config"
	"Hyper/models"
	"Hyper/pkg/log"
	"Hyper/types"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"

	qrcode "github.com/skip2/go-qrcode"
	"go.uber.org/zap"
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
	GetActivityStatistics(ctx context.Context, userID, activityID int64) (*types.ActivityStatisticsResponse, error)
	GetActivityDailyStatistics(ctx context.Context, userID, activityID int64, days int) (*types.PageResponse[types.ActivityDailyStatisticsItem], error)
	GetTicketSpecs(ctx context.Context, activityID int64) ([]models.TicketSpec, error)
	SaveTicketSpecs(ctx context.Context, userID, activityID int64, specs []types.TicketSpecSaveItem) error
	DeleteTicketSpec(ctx context.Context, userID, specID int64) error
	CreateTicketOrder(ctx context.Context, userID int64, req types.CreateTicketOrderRequest) (*types.CreateTicketOrderResponse, error)
	ListTicketOrders(ctx context.Context, userID int64, status *int8, page, size int) (*types.PageResponse[types.TicketOrderListItem], error)
	GetTicketOrderDetail(ctx context.Context, userID int64, orderNo string) (*types.TicketOrderDetailResponse, error)
	ListCancelReasons(ctx context.Context) ([]models.CancelReason, error)
	CancelTicketOrder(ctx context.Context, userID int64, orderNo string, reasonID int64) error
	ListRefundReasons(ctx context.Context) ([]models.RefundReason, error)
	ApplyRefund(ctx context.Context, userID int64, req types.ApplyRefundRequest) (string, error)
	ListUserRefunds(ctx context.Context, userID int64, status *int8, page, size int) (*types.PageResponse[types.UserRefundListItem], error)
	ListOrganizerOrders(ctx context.Context, userID int64, activityID int64, status *int8, keyword, startDate, endDate string, page, size int) (*types.PageResponse[types.OrganizerOrderListItem], error)
	ListOrganizerRefunds(ctx context.Context, userID int64, status *int8, page, size int) (*types.PageResponse[types.OrganizerRefundListItem], error)
	GetRefundDetail(ctx context.Context, userID int64, refundNo string) (*types.RefundDetailResponse, error)
	RejectRefund(ctx context.Context, userID int64, refundNo string, req types.RejectRefundRequest) error
	CancelRefund(ctx context.Context, userID int64, refundNo string) error
	ListVerifiers(ctx context.Context, userID int64, page, size int) (*types.PageResponse[models.Verifier], error)
	AddVerifier(ctx context.Context, userID int64, req types.VerifierRequest) error
	DeleteVerifier(ctx context.Context, userID, verifierID int64) error
	GetVerifierActivationQR(ctx context.Context, userID, verifierID int64) (map[string]string, error)
	GetVerifierActivationInfo(ctx context.Context, verifierID int64) (*types.VerifierActivationInfoResponse, error)
	ActivateVerifier(ctx context.Context, userID int64, req types.ActivateVerifierRequest) (*types.ActivateVerifierResponse, error)
	ScanOrder(ctx context.Context, req types.ScanOrderRequest) (*types.ScanOrderResponse, error)
	ConfirmVerify(ctx context.Context, verifierID int64, req types.ConfirmVerifyRequest) error
	ListVerified(ctx context.Context, verifierID int64, page, size int) (*types.PageResponse[types.VerifiedListItem], error)
	ListViewers(ctx context.Context, userID int64) (*types.PageResponse[types.ViewerItem], error)
	CreateViewer(ctx context.Context, userID int64, req types.CreateViewerReq) (int64, error)
	UpdateViewer(ctx context.Context, userID, viewerID int64, req types.UpdateViewerReq) error
	DeleteViewer(ctx context.Context, userID, viewerID int64) error
	ListStores(ctx context.Context, userID int64, keyword string, page, size int) (*types.PageResponse[models.OrganizerStore], error)
	CreateStore(ctx context.Context, userID int64, req types.StoreRequest) (int64, error)
	UpdateStore(ctx context.Context, userID, storeID int64, req types.StoreRequest) error
	DeleteStore(ctx context.Context, userID, storeID int64) error
}

type TicketingService struct {
	DB            *gorm.DB
	Config        *config.Config
	WeChatService IWeChatService
	OssService    IOssService
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
	if req.Type != models.OrganizerTypeVenue && req.Type != models.OrganizerTypeMerchant {
		return errors.New("入驻类型无效，仅支持 venue 或 merchant")
	}
	var org models.Organizer
	err := s.DB.WithContext(ctx).Where("user_id = ?", userID).First(&org).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	data := models.Organizer{
		UserID:   userID,
		Type:     req.Type,
		Name:     req.Name,
		Logo:     req.Logo,
		Status:   models.OrganizerStatusAuditing,
		Level:    "LV1",
		Province: req.Province,
		City:     req.City,
		District: req.District,
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := s.DB.WithContext(ctx).Create(&data).Error; err != nil {
			return err
		}
		s.notifyOrganizerApply(ctx, data, userID)
		return nil
	}
	if org.Status == models.OrganizerStatusApproved {
		return errors.New("入驻申请已通过，无需重复提交")
	}
	if err := s.DB.WithContext(ctx).Model(&org).Updates(map[string]any{
		"type":          req.Type,
		"name":          req.Name,
		"logo":          req.Logo,
		"status":        models.OrganizerStatusAuditing,
		"reject_reason": "",
		"level":         "LV1",
		"province":      req.Province,
		"city":          req.City,
		"district":      req.District,
	}).Error; err != nil {
		return err
	}
	data.ID = org.ID
	s.notifyOrganizerApply(ctx, data, userID)
	return nil
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
	var activity models.Activity
	if err := s.DB.WithContext(ctx).
		Where("id = ? AND organizer_id = ? AND status IN ?", activityID, org.ID, []int8{models.ActivityStatusDraft, models.ActivityStatusRejected}).
		First(&activity).Error; err != nil {
		return err
	}
	if err := validateChinaCoordinate(activity.Latitude, activity.Longitude); err != nil {
		return err
	}
	return s.DB.WithContext(ctx).Model(&activity).Update("status", models.ActivityStatusPending).Error
}

func (s *TicketingService) GetActivityStatistics(ctx context.Context, userID, activityID int64) (*types.ActivityStatisticsResponse, error) {
	if err := s.ensureActivityOwner(ctx, userID, activityID); err != nil {
		return nil, err
	}
	paidStatuses := []int8{models.TicketOrderStatusUsable, models.TicketOrderStatusUsed, models.TicketOrderStatusRefunding, models.TicketOrderStatusRefundSuccess, models.TicketOrderStatusRefundReject}
	resp := &types.ActivityStatisticsResponse{}
	if err := s.DB.WithContext(ctx).Model(&models.TicketOrder{}).
		Where("activity_id = ? AND status IN ?", activityID, paidStatuses).
		Select("COALESCE(SUM(quantity),0), COALESCE(SUM(actual_price),0), COUNT(DISTINCT user_id), COUNT(*)").
		Row().Scan(&resp.TicketCount, &resp.GrossAmount, &resp.BuyerCount, new(int64)); err != nil {
		return nil, err
	}
	if err := s.DB.WithContext(ctx).Model(&models.TicketOrder{}).
		Where("activity_id = ? AND status = ?", activityID, models.TicketOrderStatusUsed).
		Select("COALESCE(SUM(quantity),0)").Scan(&resp.VerifiedCount).Error; err != nil {
		return nil, err
	}
	if err := s.DB.WithContext(ctx).Table("refunds r").
		Joins("JOIN ticket_orders o ON o.id = r.order_id").
		Where("o.activity_id = ? AND r.status = ?", activityID, models.RefundStatusSuccess).
		Select("COALESCE(SUM(r.refund_amount),0)").Scan(&resp.RefundAmount).Error; err != nil {
		return nil, err
	}
	resp.NetAmount = resp.GrossAmount - resp.RefundAmount
	if resp.TicketCount > 0 {
		resp.AverageTicketPrice = resp.GrossAmount / resp.TicketCount
		resp.VerifyRate = float64(resp.VerifiedCount) / float64(resp.TicketCount)
	}
	return resp, nil
}

func (s *TicketingService) GetActivityDailyStatistics(ctx context.Context, userID, activityID int64, days int) (*types.PageResponse[types.ActivityDailyStatisticsItem], error) {
	if err := s.ensureActivityOwner(ctx, userID, activityID); err != nil {
		return nil, err
	}
	if days <= 0 {
		days = 7
	}
	if days > 90 {
		days = 90
	}
	start := time.Now().AddDate(0, 0, -days+1).Format("2006-01-02")
	var rows []types.ActivityDailyStatisticsItem
	if err := s.DB.WithContext(ctx).Model(&models.TicketOrder{}).
		Select("DATE(created_at) AS date, COALESCE(SUM(actual_price),0) AS amount, COALESCE(SUM(quantity),0) AS ticket_count, COUNT(*) AS order_count").
		Where("activity_id = ? AND status IN ? AND created_at >= ?", activityID, []int8{models.TicketOrderStatusUsable, models.TicketOrderStatusUsed, models.TicketOrderStatusRefunding, models.TicketOrderStatusRefundSuccess, models.TicketOrderStatusRefundReject}, start).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return &types.PageResponse[types.ActivityDailyStatisticsItem]{List: rows, Total: int64(len(rows))}, nil
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

func (s *TicketingService) CreateTicketOrder(ctx context.Context, userID int64, req types.CreateTicketOrderRequest) (*types.CreateTicketOrderResponse, error) {
	orderNo := "T" + time.Now().Format("20060102150405") + randomHex(4)
	expireTime := time.Now().Add(15 * time.Minute)
	result := &types.CreateTicketOrderResponse{OrderNo: orderNo}
	qrContent := "TICKET:" + orderNo + ":" + randomHex(8)
	qrURL, err := s.generateTicketQRCodeURL(ctx, orderNo, qrContent)
	if err != nil {
		return nil, err
	}
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
		orderViewers, err := s.resolveOrderViewers(tx, userID, req, req.Quantity, act.RealNameMode == 1, act.MinorCheck == 1)
		if err != nil {
			return err
		}
		buyerName, buyerIDCard := req.BuyerName, req.BuyerIDCard
		if len(orderViewers) > 0 {
			buyerName = orderViewers[0].RealName
			buyerIDCard = orderViewers[0].IDCard
		}
		if err := tx.Model(&spec).UpdateColumn("sold_count", gorm.Expr("sold_count + ?", req.Quantity)).Error; err != nil {
			return err
		}
		totalPrice := spec.Price * int64(req.Quantity)
		pointsAmount, pointsDiscount := int64(0), int64(0)
		if req.UsePoints || req.PointsAmount > 0 {
			if req.PointsAmount <= 0 {
				return errors.New("积分抵扣数量必须大于0")
			}
			pointsAmount = req.PointsAmount
			pointsDiscount = pointsAmount * 10
			if pointsDiscount > totalPrice {
				return errors.New("积分抵扣金额超过订单金额")
			}
			var account models.UserPoint
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("user_id = ?", uint64(userID)).
				First(&account).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("积分余额不足")
				}
				return err
			}
			if account.Balance < pointsAmount {
				return errors.New("积分余额不足")
			}
			newBalance := account.Balance - pointsAmount
			if err := tx.Model(&models.UserPoint{}).
				Where("user_id = ?", uint64(userID)).
				Updates(map[string]any{
					"balance":    newBalance,
					"total_used": gorm.Expr("total_used + ?", pointsAmount),
				}).Error; err != nil {
				return err
			}
			if err := tx.Create(&models.PointsLog{
				UserID:     uint64(userID),
				Amount:     -pointsAmount,
				Balance:    newBalance,
				ChangeType: models.TypeOrderDeduct,
				SourceID:   orderNo,
				Remark:     "票务订单积分抵扣",
				Status:     1,
			}).Error; err != nil {
				return err
			}
		}
		actualPrice := totalPrice - pointsDiscount
		order := models.TicketOrder{
			OrderNo:        orderNo,
			UserID:         userID,
			ActivityID:     req.ActivityID,
			TicketSpecID:   req.TicketSpecID,
			Quantity:       req.Quantity,
			TotalPrice:     totalPrice,
			ActualPrice:    actualPrice,
			PointsAmount:   pointsAmount,
			PointsDiscount: pointsDiscount,
			BuyerName:      buyerName,
			BuyerIDCard:    buyerIDCard,
			Status:         models.TicketOrderStatusPending,
			ExpireTime:     expireTime,
			QRCode:         qrContent,
			QRCodeURL:      qrURL,
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		for i := range orderViewers {
			orderViewers[i].OrderID = order.ID
			orderViewers[i].OrderNo = order.OrderNo
		}
		if len(orderViewers) > 0 {
			if err := tx.Create(&orderViewers).Error; err != nil {
				return err
			}
		}
		result.TotalPrice = totalPrice
		result.PointsAmount = pointsAmount
		result.PointsDiscount = pointsDiscount
		result.ActualPrice = actualPrice
		result.QRCode = qrContent
		result.QRCodeURL = qrURL
		result.Viewers = orderViewerItems(orderViewers, false)
		return nil
	})
	return result, err
}

func (s *TicketingService) resolveOrderViewers(tx *gorm.DB, userID int64, req types.CreateTicketOrderRequest, quantity int, realNameMode, minorCheck bool) ([]models.TicketOrderViewer, error) {
	viewers := make([]models.TicketOrderViewer, 0, len(req.ViewerIDs)+len(req.Viewers))
	seenIDs := make(map[int64]struct{})
	if len(req.ViewerIDs) > 0 {
		var saved []models.Viewer
		if err := tx.Where("user_id = ? AND id IN ?", int(userID), req.ViewerIDs).Find(&saved).Error; err != nil {
			return nil, err
		}
		savedByID := make(map[int64]models.Viewer, len(saved))
		for _, viewer := range saved {
			savedByID[viewer.ID] = viewer
		}
		for _, id := range req.ViewerIDs {
			if id <= 0 {
				return nil, errors.New("观演人ID无效")
			}
			if _, ok := seenIDs[id]; ok {
				return nil, errors.New("观演人不能重复选择")
			}
			viewer, ok := savedByID[id]
			if !ok {
				return nil, errors.New("观演人不存在")
			}
			seenIDs[id] = struct{}{}
			viewers = append(viewers, models.TicketOrderViewer{
				ViewerID: viewer.ID,
				RealName: strings.TrimSpace(viewer.RealName),
				IDCard:   strings.TrimSpace(viewer.IDCard),
				Phone:    strings.TrimSpace(viewer.Phone),
				Type:     viewer.Type,
			})
		}
	}
	for _, input := range req.Viewers {
		viewerID := input.ViewerID
		if viewerID == 0 {
			viewerID = input.ID
		}
		if viewerID > 0 {
			if _, ok := seenIDs[viewerID]; ok {
				continue
			}
			var saved models.Viewer
			if err := tx.Where("user_id = ? AND id = ?", int(userID), viewerID).First(&saved).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, errors.New("观演人不存在")
				}
				return nil, err
			}
			seenIDs[viewerID] = struct{}{}
			viewers = append(viewers, models.TicketOrderViewer{
				ViewerID: saved.ID,
				RealName: strings.TrimSpace(saved.RealName),
				IDCard:   strings.TrimSpace(saved.IDCard),
				Phone:    strings.TrimSpace(saved.Phone),
				Type:     saved.Type,
			})
			continue
		}
		viewers = append(viewers, models.TicketOrderViewer{
			RealName: strings.TrimSpace(input.RealName),
			IDCard:   strings.TrimSpace(input.IDCard),
			Phone:    strings.TrimSpace(input.Phone),
			Type:     input.Type,
		})
	}
	if len(viewers) == 0 && (strings.TrimSpace(req.BuyerName) != "" || strings.TrimSpace(req.BuyerIDCard) != "") {
		viewers = append(viewers, models.TicketOrderViewer{
			RealName: strings.TrimSpace(req.BuyerName),
			IDCard:   strings.TrimSpace(req.BuyerIDCard),
			Type:     1,
		})
	}
	if realNameMode {
		if len(viewers) != quantity {
			return nil, errors.New("实名模式下观演人数必须等于购票数量")
		}
	}
	seenIDCards := make(map[string]struct{})
	for i := range viewers {
		if viewers[i].Type == 0 {
			viewers[i].Type = 1
		}
		if viewers[i].RealName == "" || viewers[i].IDCard == "" {
			return nil, errors.New("实名模式下需要填写每位观演人的姓名和身份证")
		}
		if _, ok := seenIDCards[viewers[i].IDCard]; ok {
			return nil, errors.New("观演人不能重复选择")
		}
		seenIDCards[viewers[i].IDCard] = struct{}{}
		if minorCheck && isMinor(viewers[i].IDCard) {
			return nil, errors.New("未成年人不可购票")
		}
	}
	return viewers, nil
}

func (s *TicketingService) GetTicketOrderDetail(ctx context.Context, userID int64, orderNo string) (*types.TicketOrderDetailResponse, error) {
	var order models.TicketOrder
	if err := s.DB.WithContext(ctx).Where("order_no = ? AND user_id = ?", orderNo, userID).First(&order).Error; err != nil {
		return nil, err
	}
	return s.buildOrderDetail(ctx, order)
}

func (s *TicketingService) generateTicketQRCodeURL(ctx context.Context, orderNo, content string) (string, error) {
	if s.OssService == nil {
		return "", errors.New("OSS 服务未初始化")
	}
	qr, err := qrcode.New(content, qrcode.High)
	if err != nil {
		return "", err
	}
	qr.ForegroundColor = color.RGBA{R: 32, G: 18, B: 54, A: 255}
	qr.BackgroundColor = color.RGBA{R: 248, G: 252, B: 255, A: 255}
	png, err := qr.PNG(640)
	if err != nil {
		return "", err
	}
	objectKey := fmt.Sprintf("ticket/qrcode/%s/%s.png", time.Now().Format("2006/01/02"), orderNo)
	if err := s.OssService.UploadRaw(ctx, bytes.NewReader(png), objectKey); err != nil {
		return "", err
	}
	return "https://cdn.hypercn.cn/" + objectKey, nil
}

func (s *TicketingService) ListTicketOrders(ctx context.Context, userID int64, status *int8, page, size int) (*types.PageResponse[types.TicketOrderListItem], error) {
	page, size = normalizePage(page, size)
	query := s.DB.WithContext(ctx).Model(&models.TicketOrder{}).Where("user_id = ?", userID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var rows []struct {
		OrderNo        string
		Status         int8
		TotalPrice     int64
		ActualPrice    int64
		Quantity       int
		ActivityID     int64
		ActivityName   string
		StartTime      time.Time
		EndTime        time.Time
		PosterList     string
		TicketSpecID   int64
		TicketSpecName string
		BuyerName      string
		BuyerIDCard    string
		CreatedAt      time.Time
		ExpireTime     time.Time
		PayTime        *time.Time
	}
	err := query.
		Select(`ticket_orders.order_no,
			ticket_orders.status,
			ticket_orders.total_price,
			ticket_orders.actual_price,
			ticket_orders.quantity,
			ticket_orders.activity_id,
			activities.name AS activity_name,
			activities.start_time,
			activities.end_time,
			activities.poster_list,
			ticket_orders.ticket_spec_id,
			ticket_specs.name AS ticket_spec_name,
			ticket_orders.buyer_name,
			ticket_orders.buyer_id_card,
			ticket_orders.created_at,
			ticket_orders.expire_time,
			ticket_orders.pay_time`).
		Joins("LEFT JOIN activities ON activities.id = ticket_orders.activity_id").
		Joins("LEFT JOIN ticket_specs ON ticket_specs.id = ticket_orders.ticket_spec_id").
		Order("ticket_orders.created_at desc").
		Offset((page - 1) * size).
		Limit(size).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	list := make([]types.TicketOrderListItem, 0, len(rows))
	orderNos := make([]string, 0, len(rows))
	for _, r := range rows {
		orderNos = append(orderNos, r.OrderNo)
	}
	viewersByOrderNo, err := s.orderViewersByOrderNo(ctx, orderNos)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		item := types.TicketOrderListItem{
			OrderNo:     r.OrderNo,
			Status:      r.Status,
			TotalPrice:  r.TotalPrice,
			ActualPrice: r.ActualPrice,
			Quantity:    r.Quantity,
			BuyerName:   r.BuyerName,
			BuyerIDCard: maskIDCard(r.BuyerIDCard),
			CreatedAt:   r.CreatedAt,
			ExpireTime:  r.ExpireTime,
			PayTime:     r.PayTime,
		}
		item.Viewers = orderViewerItems(viewersByOrderNo[r.OrderNo], false)
		item.Activity.ID = r.ActivityID
		item.Activity.Name = r.ActivityName
		item.Activity.StartTime = r.StartTime
		item.Activity.EndTime = r.EndTime
		item.Activity.PosterList = r.PosterList
		item.TicketSpec.ID = r.TicketSpecID
		item.TicketSpec.Name = r.TicketSpecName
		list = append(list, item)
	}
	return &types.PageResponse[types.TicketOrderListItem]{List: list, Total: total}, nil
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

func (s *TicketingService) ListUserRefunds(ctx context.Context, userID int64, status *int8, page, size int) (*types.PageResponse[types.UserRefundListItem], error) {
	page, size = normalizePage(page, size)
	query := s.DB.WithContext(ctx).Table("refunds r").
		Joins("JOIN ticket_orders o ON o.id = r.order_id").
		Joins("JOIN activities a ON a.id = o.activity_id").
		Where("o.user_id = ?", userID)
	if status != nil {
		query = query.Where("r.status = ?", *status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []struct {
		RefundNo     string
		OrderNo      string
		Status       int8
		RefundAmount int64
		Reason       string
		CreatedAt    time.Time
		UpdatedAt    time.Time
		ActivityID   int64
		ActivityName string
		PosterList   string
	}
	if err := query.Select(`
			r.refund_no,
			o.order_no,
			r.status,
			r.refund_amount,
			r.reason,
			r.created_at,
			r.updated_at,
			a.id AS activity_id,
			a.name AS activity_name,
			a.poster_list
		`).
		Order("r.id desc").
		Offset((page - 1) * size).
		Limit(size).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]types.UserRefundListItem, 0, len(rows))
	for _, row := range rows {
		item := types.UserRefundListItem{
			RefundNo:     row.RefundNo,
			OrderNo:      row.OrderNo,
			Status:       row.Status,
			StatusText:   refundStatusText(row.Status),
			RefundAmount: row.RefundAmount,
			Reason:       row.Reason,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
		}
		item.Activity.ID = row.ActivityID
		item.Activity.Name = row.ActivityName
		item.Activity.PosterList = row.PosterList
		list = append(list, item)
	}
	return &types.PageResponse[types.UserRefundListItem]{List: list, Total: total}, nil
}

func (s *TicketingService) ListOrganizerOrders(ctx context.Context, userID int64, activityID int64, status *int8, keyword, startDate, endDate string, page, size int) (*types.PageResponse[types.OrganizerOrderListItem], error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	page, size = normalizePage(page, size)
	query := s.DB.WithContext(ctx).Table("ticket_orders o").
		Joins("JOIN activities a ON a.id = o.activity_id").
		Joins("JOIN ticket_specs ts ON ts.id = o.ticket_spec_id").
		Joins("LEFT JOIN users u ON u.id = o.user_id").
		Where("a.organizer_id = ?", org.ID)
	if activityID > 0 {
		query = query.Where("o.activity_id = ?", activityID)
	}
	if status != nil {
		query = query.Where("o.status = ?", *status)
	}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where(`
			o.order_no LIKE ?
			OR u.nickname LIKE ?
			OR u.mobile LIKE ?
			OR o.buyer_name LIKE ?
			OR o.buyer_id_card LIKE ?
			OR EXISTS (
				SELECT 1 FROM ticket_order_viewers tov
				WHERE tov.order_id = o.id
				AND (tov.real_name LIKE ? OR tov.id_card LIKE ? OR tov.phone LIKE ?)
			)
		`, like, like, like, like, like, like, like, like)
	}
	if startDate = strings.TrimSpace(startDate); startDate != "" {
		start, err := parseDateStart(startDate)
		if err != nil {
			return nil, fmt.Errorf("开始日期格式错误")
		}
		query = query.Where("o.created_at >= ?", start)
	}
	if endDate = strings.TrimSpace(endDate); endDate != "" {
		end, err := parseDateEnd(endDate)
		if err != nil {
			return nil, fmt.Errorf("结束日期格式错误")
		}
		query = query.Where("o.created_at < ?", end)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []struct {
		OrderNo        string
		Status         int8
		TotalPrice     int64
		ActualPrice    int64
		PointsAmount   int64
		PointsDiscount int64
		Quantity       int
		UserID         int64
		UserName       string
		UserMobile     string
		UserAvatar     string
		BuyerName      string
		BuyerIDCard    string
		ActivityID     int64
		ActivityName   string
		TicketSpecID   int64
		TicketSpecName string
		PayMethod      string
		PayTime        *time.Time
		CreatedAt      time.Time
		ExpireTime     time.Time
	}
	if err := query.Select(`
			o.order_no,
			o.status,
			o.total_price,
			o.actual_price,
			o.points_amount,
			o.points_discount,
			o.quantity,
			o.user_id,
			u.nickname AS user_name,
			u.mobile AS user_mobile,
			u.avatar AS user_avatar,
			o.buyer_name,
			o.buyer_id_card,
			o.activity_id,
			a.name AS activity_name,
			o.ticket_spec_id,
			ts.name AS ticket_spec_name,
			o.pay_method,
			o.pay_time,
			o.created_at,
			o.expire_time
		`).
		Order("o.created_at desc").
		Offset((page - 1) * size).
		Limit(size).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	orderNos := make([]string, 0, len(rows))
	for _, row := range rows {
		orderNos = append(orderNos, row.OrderNo)
	}
	viewersByOrderNo, err := s.orderViewersByOrderNo(ctx, orderNos)
	if err != nil {
		return nil, err
	}
	list := make([]types.OrganizerOrderListItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, types.OrganizerOrderListItem{
			OrderNo:        row.OrderNo,
			Status:         row.Status,
			TotalPrice:     row.TotalPrice,
			ActualPrice:    row.ActualPrice,
			PointsAmount:   row.PointsAmount,
			PointsDiscount: row.PointsDiscount,
			Quantity:       row.Quantity,
			UserID:         row.UserID,
			UserName:       row.UserName,
			UserMobile:     maskPhone(row.UserMobile),
			UserAvatar:     row.UserAvatar,
			BuyerName:      row.BuyerName,
			BuyerIDCard:    maskIDCard(row.BuyerIDCard),
			Viewers:        orderViewerItems(viewersByOrderNo[row.OrderNo], false),
			ActivityID:     row.ActivityID,
			ActivityName:   row.ActivityName,
			TicketSpecID:   row.TicketSpecID,
			TicketSpecName: row.TicketSpecName,
			PayMethod:      row.PayMethod,
			PayTime:        row.PayTime,
			CreatedAt:      row.CreatedAt,
			ExpireTime:     row.ExpireTime,
		})
	}
	return &types.PageResponse[types.OrganizerOrderListItem]{List: list, Total: total}, nil
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
		if err := tx.Model(&refund).Update("status", models.RefundStatusCancelled).Error; err != nil {
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
	resp := map[string]string{
		"wechat_qr_page": "pages/user-sub/verifier-bind/index",
		"wechat_scene":   fmt.Sprintf("v=%d", verifierID),
		"douyin_qr":      fmt.Sprintf("hyper://verifier/activate?verifier_id=%d&channel=douyin", verifierID),
	}
	if s.WeChatService == nil {
		return nil, errors.New("微信服务未初始化")
	}
	if s.OssService == nil {
		return nil, errors.New("OSS 服务未初始化")
	}
	if len(resp["wechat_scene"]) > 32 {
		return nil, errors.New("微信小程序码 scene 超过 32 个字符")
	}
	qrBytes, err := s.WeChatService.GenerateUnlimitedQRCode(ctx, resp["wechat_scene"], resp["wechat_qr_page"])
	if err != nil {
		return nil, err
	}
	objectKey := fmt.Sprintf("verifier/qrcode/%s/%d.jpg", time.Now().Format("2006/01/02"), verifierID)
	if err := s.OssService.UploadRaw(ctx, bytes.NewReader(qrBytes), objectKey); err != nil {
		return nil, err
	}
	qrURL := "https://cdn.hypercn.cn/" + objectKey
	resp["wechat_qr"] = qrURL
	resp["wechat_qr_url"] = qrURL
	resp["wechat_mini_program_code_url"] = qrURL
	return resp, nil
}

func (s *TicketingService) GetVerifierActivationInfo(ctx context.Context, verifierID int64) (*types.VerifierActivationInfoResponse, error) {
	var verifier models.Verifier
	if err := s.DB.WithContext(ctx).First(&verifier, verifierID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("核销员邀请不存在")
		}
		return nil, err
	}
	resp := &types.VerifierActivationInfoResponse{
		VerifierID:  verifier.ID,
		Name:        verifier.Name,
		Phone:       verifier.Phone,
		Status:      verifier.Status,
		IsBound:     verifier.UserID != 0,
		OrganizerID: verifier.OrganizerID,
	}
	var org models.Organizer
	if err := s.DB.WithContext(ctx).First(&org, verifier.OrganizerID).Error; err == nil {
		resp.OrganizerName = org.Name
	}
	return resp, nil
}

func (s *TicketingService) ActivateVerifier(ctx context.Context, userID int64, req types.ActivateVerifierRequest) (*types.ActivateVerifierResponse, error) {
	var user models.Users
	if err := s.DB.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, err
	}
	req.Phone = strings.TrimSpace(req.Phone)
	req.Channel = strings.TrimSpace(req.Channel)
	if req.Channel == "" {
		req.Channel = "wechat"
	}
	if strings.TrimSpace(user.Mobile) == "" {
		return nil, errors.New("请先绑定手机号")
	}
	if strings.TrimSpace(user.Mobile) != req.Phone {
		return nil, errors.New("登录手机号与核销员手机号不一致")
	}

	var verifier models.Verifier
	query := s.DB.WithContext(ctx).Where("phone = ?", req.Phone)
	if req.VerifierID > 0 {
		query = query.Where("id = ?", req.VerifierID)
	} else {
		query = query.Where("(user_id = 0 OR user_id = ?)", userID)
	}
	if err := query.Order("id desc").First(&verifier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("核销员邀请不存在")
		}
		return nil, err
	}
	if verifier.UserID != 0 && verifier.UserID != userID {
		return nil, errors.New("该核销员已绑定其他账号")
	}

	now := time.Now()
	if err := s.DB.WithContext(ctx).Model(&models.Verifier{}).
		Where("id = ?", verifier.ID).
		Updates(map[string]any{
			"user_id":  userID,
			"status":   models.VerifierStatusActive,
			"channel":  req.Channel,
			"bound_at": now,
		}).Error; err != nil {
		return nil, err
	}
	return &types.ActivateVerifierResponse{
		Success:    true,
		VerifierID: verifier.ID,
		UserID:     userID,
		Status:     models.VerifierStatusActive,
	}, nil
}

func (s *TicketingService) ScanOrder(ctx context.Context, req types.ScanOrderRequest) (*types.ScanOrderResponse, error) {
	var order models.TicketOrder
	if err := s.DB.WithContext(ctx).Where("qr_code = ?", req.QRCode).First(&order).Error; err != nil {
		return &types.ScanOrderResponse{Success: false, ErrorCode: "ORDER_NOT_FOUND"}, nil
	}
	if req.ActivityID > 0 && order.ActivityID != req.ActivityID {
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
		ActivityName      string                  `json:"activity_name"`
		TicketSpecName    string                  `json:"ticket_spec_name"`
		Quantity          int                     `json:"quantity"`
		BuyerNameMasked   string                  `json:"buyer_name_masked"`
		BuyerIDCardMasked string                  `json:"buyer_id_card_masked"`
		Viewers           []types.OrderViewerItem `json:"viewers,omitempty"`
	}{
		ActivityName:      activity.Name,
		TicketSpecName:    spec.Name,
		Quantity:          order.Quantity,
		BuyerNameMasked:   maskName(order.BuyerName),
		BuyerIDCardMasked: maskIDCard(order.BuyerIDCard),
	}
	if viewers, err := s.orderViewers(ctx, order.ID); err == nil {
		item.Viewers = orderViewerItems(viewers, false)
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

func (s *TicketingService) ListViewers(ctx context.Context, userID int64) (*types.PageResponse[types.ViewerItem], error) {
	var viewers []models.Viewer
	if err := s.DB.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Find(&viewers).Error; err != nil {
		return nil, err
	}
	list := make([]types.ViewerItem, 0, len(viewers))
	for _, viewer := range viewers {
		list = append(list, types.ViewerItem{
			ID:        viewer.ID,
			RealName:  viewer.RealName,
			IDCard:    maskIDCard(viewer.IDCard),
			Phone:     maskPhone(viewer.Phone),
			Type:      viewer.Type,
			CreatedAt: viewer.CreatedAt,
			UpdatedAt: viewer.UpdatedAt,
		})
	}
	return &types.PageResponse[types.ViewerItem]{List: list, Total: int64(len(list))}, nil
}

func (s *TicketingService) CreateViewer(ctx context.Context, userID int64, req types.CreateViewerReq) (int64, error) {
	var viewerID int64
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		viewerType, err := GetAgeTypeByIDCard(req.IDCard)
		if err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&models.Viewer{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
			return err
		}
		if count >= 5 {
			return errors.New("常用观演人已达上限")
		}
		var exists models.Viewer
		if err := tx.Where("user_id = ? AND id_card = ?", userID, req.IDCard).First(&exists).Error; err == nil {
			return errors.New("该观演人已存在")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := tx.Where("user_id = ? AND phone = ?", userID, req.Phone).First(&exists).Error; err == nil {
			return errors.New("该手机号已被当前账号其他观演人绑定")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		viewer := models.Viewer{
			UserID:   int(userID),
			RealName: req.RealName,
			IDCard:   req.IDCard,
			Phone:    req.Phone,
			Type:     viewerType,
		}
		if err := tx.Create(&viewer).Error; err != nil {
			return err
		}
		viewerID = viewer.ID
		return nil
	})
	return viewerID, err
}

func (s *TicketingService) UpdateViewer(ctx context.Context, userID, viewerID int64, req types.UpdateViewerReq) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var viewer models.Viewer
		if err := tx.Where("id = ? AND user_id = ?", viewerID, userID).First(&viewer).Error; err != nil {
			return err
		}
		updates := make(map[string]any)
		if req.RealName != "" {
			updates["real_name"] = req.RealName
		}
		if req.Phone != "" {
			var exists models.Viewer
			if err := tx.Where("user_id = ? AND phone = ? AND id <> ?", userID, req.Phone, viewerID).First(&exists).Error; err == nil {
				return errors.New("该手机号已被其他观演人绑定")
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			updates["phone"] = req.Phone
		}
		if len(updates) == 0 {
			return nil
		}
		return tx.Model(&viewer).Updates(updates).Error
	})
}

func (s *TicketingService) DeleteViewer(ctx context.Context, userID, viewerID int64) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var viewer models.Viewer
		if err := tx.Where("id = ? AND user_id = ?", viewerID, userID).First(&viewer).Error; err != nil {
			return err
		}
		var activeOrders int64
		if err := tx.Model(&models.TicketOrder{}).
			Where("user_id = ? AND buyer_id_card = ? AND status IN ?", userID, viewer.IDCard, []int8{
				models.TicketOrderStatusPending,
				models.TicketOrderStatusUsable,
				models.TicketOrderStatusRefunding,
				models.TicketOrderStatusRefundReject,
			}).
			Count(&activeOrders).Error; err != nil {
			return err
		}
		if activeOrders > 0 {
			return errors.New("观演人已关联未完成订单，暂不可删除")
		}
		if err := tx.Table("ticket_order_viewers tov").
			Joins("JOIN ticket_orders o ON o.id = tov.order_id").
			Where("o.user_id = ? AND (tov.viewer_id = ? OR tov.id_card = ?) AND o.status IN ?", userID, viewer.ID, viewer.IDCard, []int8{
				models.TicketOrderStatusPending,
				models.TicketOrderStatusUsable,
				models.TicketOrderStatusRefunding,
				models.TicketOrderStatusRefundReject,
			}).
			Count(&activeOrders).Error; err != nil {
			return err
		}
		if activeOrders > 0 {
			return errors.New("观演人已关联未完成订单，暂不可删除")
		}
		return tx.Delete(&viewer).Error
	})
}

func (s *TicketingService) ListStores(ctx context.Context, userID int64, keyword string, page, size int) (*types.PageResponse[models.OrganizerStore], error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	page, size = normalizePage(page, size)
	query := s.DB.WithContext(ctx).Model(&models.OrganizerStore{}).Where("organizer_id = ?", org.ID)
	if keyword != "" {
		query = query.Where("name LIKE ? OR address LIKE ? OR phone LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var stores []models.OrganizerStore
	if err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&stores).Error; err != nil {
		return nil, err
	}
	return &types.PageResponse[models.OrganizerStore]{List: stores, Total: total}, nil
}

func (s *TicketingService) CreateStore(ctx context.Context, userID int64, req types.StoreRequest) (int64, error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return 0, err
	}
	store := models.OrganizerStore{
		OrganizerID: org.ID,
		Name:        req.Name,
		Logo:        req.Logo,
		Address:     req.Address,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Phone:       req.Phone,
	}
	if err := s.DB.WithContext(ctx).Create(&store).Error; err != nil {
		return 0, err
	}
	return store.ID, nil
}

func (s *TicketingService) UpdateStore(ctx context.Context, userID, storeID int64, req types.StoreRequest) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	updates := map[string]any{
		"name":       req.Name,
		"logo":       req.Logo,
		"address":    req.Address,
		"latitude":   req.Latitude,
		"longitude":  req.Longitude,
		"phone":      req.Phone,
		"updated_at": time.Now(),
	}
	result := s.DB.WithContext(ctx).Model(&models.OrganizerStore{}).Where("id = ? AND organizer_id = ?", storeID, org.ID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("门店不存在")
	}
	return nil
}

func (s *TicketingService) DeleteStore(ctx context.Context, userID, storeID int64) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	result := s.DB.WithContext(ctx).Where("id = ? AND organizer_id = ?", storeID, org.ID).Delete(&models.OrganizerStore{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("门店不存在")
	}
	return nil
}

func (s *TicketingService) findOrganizerByUser(ctx context.Context, userID int64) (*models.Organizer, error) {
	var org models.Organizer
	err := s.DB.WithContext(ctx).Where("user_id = ?", userID).First(&org).Error
	return &org, err
}

func (s *TicketingService) ensureActivityOwner(ctx context.Context, userID, activityID int64) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	var count int64
	if err := s.DB.WithContext(ctx).Model(&models.Activity{}).Where("id = ? AND organizer_id = ?", activityID, org.ID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("活动不存在或无权限")
	}
	return nil
}

func (s *TicketingService) ensureOrganizer(ctx context.Context, userID int64) (*models.Organizer, error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err == nil {
		return org, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	org = &models.Organizer{UserID: userID, Type: models.OrganizerTypeVenue, Name: "未命名主办方", Status: models.OrganizerStatusPending, Level: "LV1"}
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
		OrderNo:        order.OrderNo,
		Status:         order.Status,
		TotalPrice:     order.TotalPrice,
		ActualPrice:    order.ActualPrice,
		PointsAmount:   order.PointsAmount,
		PointsDiscount: order.PointsDiscount,
		Quantity:       order.Quantity,
		BuyerName:      order.BuyerName,
		BuyerIDCard:    order.BuyerIDCard,
		PayMethod:      order.PayMethod,
		PayTime:        order.PayTime,
		CreatedAt:      order.CreatedAt,
		QRCode:         order.QRCode,
		QRCodeURL:      order.QRCodeURL,
		ExpireTime:     order.ExpireTime,
	}
	resp.Activity.ID = act.ID
	resp.Activity.Name = act.Name
	resp.Activity.StartTime = act.StartTime
	resp.Activity.EndTime = act.EndTime
	resp.Activity.PosterList = act.PosterList
	resp.TicketSpec.Name = spec.Name
	if viewers, err := s.orderViewers(ctx, order.ID); err == nil {
		resp.Viewers = orderViewerItems(viewers, true)
	}
	var refund models.Refund
	if err := s.DB.WithContext(ctx).Where("order_id = ?", order.ID).Order("id desc").First(&refund).Error; err == nil {
		statusText := refundStatusText(refund.Status)
		resp.RefundNo = refund.RefundNo
		resp.RefundInfo = &struct {
			RefundNo         string `json:"refund_no"`
			RefundAmount     int64  `json:"refund_amount"`
			Status           int8   `json:"status"`
			StatusText       string `json:"status_text"`
			ExpectArriveDate string `json:"expect_arrive_date"`
		}{RefundNo: refund.RefundNo, RefundAmount: refund.RefundAmount, Status: refund.Status, StatusText: statusText, ExpectArriveDate: refund.ExpectArriveDate}
		resp.Refund = &struct {
			RefundNo   string `json:"refund_no"`
			Status     int8   `json:"status"`
			StatusText string `json:"status_text"`
		}{RefundNo: refund.RefundNo, Status: refund.Status, StatusText: statusText}
	}
	return resp, nil
}

func refundStatusText(status int8) string {
	switch status {
	case models.RefundStatusAuditing:
		return "待审核"
	case models.RefundStatusRunning:
		return "退款中"
	case models.RefundStatusSuccess:
		return "已退款"
	case models.RefundStatusRejected:
		return "已驳回"
	case models.RefundStatusCancelled:
		return "已取消"
	default:
		return "未知"
	}
}

func (s *TicketingService) orderViewers(ctx context.Context, orderID int64) ([]models.TicketOrderViewer, error) {
	var viewers []models.TicketOrderViewer
	err := s.DB.WithContext(ctx).Where("order_id = ?", orderID).Order("id asc").Find(&viewers).Error
	return viewers, err
}

func (s *TicketingService) orderViewersByOrderNo(ctx context.Context, orderNos []string) (map[string][]models.TicketOrderViewer, error) {
	result := make(map[string][]models.TicketOrderViewer)
	if len(orderNos) == 0 {
		return result, nil
	}
	var viewers []models.TicketOrderViewer
	if err := s.DB.WithContext(ctx).Where("order_no IN ?", orderNos).Order("id asc").Find(&viewers).Error; err != nil {
		return nil, err
	}
	for _, viewer := range viewers {
		result[viewer.OrderNo] = append(result[viewer.OrderNo], viewer)
	}
	return result, nil
}

func orderViewerItems(viewers []models.TicketOrderViewer, includeSensitive bool) []types.OrderViewerItem {
	if len(viewers) == 0 {
		return nil
	}
	items := make([]types.OrderViewerItem, 0, len(viewers))
	for _, viewer := range viewers {
		item := types.OrderViewerItem{
			ViewerID:     viewer.ViewerID,
			RealName:     viewer.RealName,
			IDCardMasked: maskIDCard(viewer.IDCard),
			PhoneMasked:  maskPhone(viewer.Phone),
			Type:         viewer.Type,
		}
		if includeSensitive {
			item.IDCard = viewer.IDCard
			item.Phone = viewer.Phone
		}
		items = append(items, item)
	}
	return items
}

func buildOrganizerInfo(org *models.Organizer) *types.OrganizerInfoResponse {
	resp := &types.OrganizerInfoResponse{
		ID:             org.ID,
		Type:           org.Type,
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

func (s *TicketingService) notifyOrganizerApply(ctx context.Context, org models.Organizer, applicantUserID int64) {
	if s.Config == nil || s.Config.App == nil || s.Config.App.OrganizerApplyTemplateID == "" || s.WeChatService == nil {
		return
	}
	var subscribers []models.AdminWechatSubscriber
	if err := s.DB.WithContext(ctx).Where("enabled = ?", 1).Find(&subscribers).Error; err != nil {
		log.L.Warn("query admin wechat subscribers failed", zap.Error(err))
		return
	}
	if len(subscribers) == 0 {
		return
	}
	applicant := fmt.Sprintf("用户%d", applicantUserID)
	var user models.Users
	if err := s.DB.WithContext(ctx).Where("id = ?", applicantUserID).First(&user).Error; err == nil {
		if user.Nickname != "" {
			applicant = user.Nickname
		} else if user.Mobile != "" {
			applicant = user.Mobile
		}
	}
	applyType := "商家入驻"
	if org.Type == models.OrganizerTypeVenue {
		applyType = "场地入驻"
	}
	for _, sub := range subscribers {
		req := types.WeChatSubscribeMessageRequest{
			ToUser:     sub.OpenID,
			TemplateID: s.Config.App.OrganizerApplyTemplateID,
			Page:       fmt.Sprintf("pages/admin/organizer/detail?id=%d", org.ID),
			Data: map[string]types.WeChatMessageItem{
				"thing1":  {Value: limitWeChatField(org.Name, 20)},
				"thing2":  {Value: limitWeChatField(applicant, 20)},
				"time3":   {Value: time.Now().Format("2006-01-02 15:04:05")},
				"phrase4": {Value: "待审核"},
				"thing5":  {Value: limitWeChatField(applyType+"申请，请及时处理", 20)},
			},
			Lang: "zh_CN",
		}
		if err := s.WeChatService.SendSubscribeMessage(ctx, req); err != nil {
			log.L.Warn("send organizer apply subscribe message failed", zap.Int64("admin_id", sub.AdminID), zap.Error(err))
		}
	}
}

func limitWeChatField(value string, max int) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= max {
		return string(runes)
	}
	return string(runes[:max])
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
	if req.Latitude != nil || req.Longitude != nil {
		if req.Latitude == nil || req.Longitude == nil {
			return nil, fmt.Errorf("经纬度必须同时填写")
		}
		if err := validateChinaCoordinate(*req.Latitude, *req.Longitude); err != nil {
			return nil, err
		}
	}
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

func validateChinaCoordinate(latitude, longitude float64) error {
	if latitude == 0 || longitude == 0 {
		return errors.New("活动经纬度不能为空，请在地图中选择准确位置")
	}
	if latitude < 3 || latitude > 54 || longitude < 73 || longitude > 136 {
		return errors.New("活动经纬度不在中国范围内，请重新选择位置")
	}
	return nil
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

func parseDateStart(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	if t, err := time.ParseInLocation("2006-01-02", value, time.Local); err == nil {
		return t, nil
	}
	return parseDatetime(value)
}

func parseDateEnd(value string) (time.Time, error) {
	t, err := parseDateStart(value)
	if err != nil || t.IsZero() {
		return t, err
	}
	if len(strings.TrimSpace(value)) == len("2006-01-02") {
		return t.AddDate(0, 0, 1), nil
	}
	return t, nil
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

func maskPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}
