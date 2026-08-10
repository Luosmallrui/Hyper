package service

import (
	"Hyper/config"
	"Hyper/dao"
	"Hyper/models"
	"Hyper/pkg/log"
	"Hyper/pkg/snowflake"
	"Hyper/types"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	qrcode "github.com/skip2/go-qrcode"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ITicketingService interface {
	GetOrganizerInfo(ctx context.Context, userID int64) (*types.OrganizerInfoResponse, error)
	GetPointsRule(ctx context.Context) (*types.PointsRule, error)
	ApplyOrganizer(ctx context.Context, userID int64, req types.OrganizerApplyRequest) (*types.OrganizerApplyResponse, error)
	GetOrganizerAuditStatus(ctx context.Context, userID int64) (*types.OrganizerAuditStatusResponse, error)
	UpdateOrganizerBasic(ctx context.Context, userID int64, req types.OrganizerBasicRequest) error
	GetWithdrawInfo(ctx context.Context, userID int64) (*types.OrganizerWithdrawInfoResponse, error)
	UpdateWithdrawInfo(ctx context.Context, userID int64, req types.OrganizerWithdrawRequest) error
	ListOrganizerWithdraws(ctx context.Context, userID int64, status *int8, page, size int) (*types.PageResponse[models.OrganizerWithdraw], error)
	CreateOrganizerWithdraw(ctx context.Context, userID int64, req types.CreateOrganizerWithdrawRequest) (int64, error)
	ListOrganizerCollections(ctx context.Context, userID int64, page, size int, keyword string) (*types.PageResponse[types.ActivityCollectionItem], error)
	GetOrganizerCollection(ctx context.Context, userID, collectionID int64) (*types.OrganizerCollectionDetail, error)
	SaveOrganizerCollection(ctx context.Context, userID, collectionID int64, req types.OrganizerCollectionRequest) (int64, error)
	DeleteOrganizerCollection(ctx context.Context, userID, collectionID int64) error
	ListOrganizerMessages(ctx context.Context, userID int64, page, size int, unreadOnly bool) (*types.PageResponse[types.OrganizerMessageItem], error)
	GetOrganizerMessageDetail(ctx context.Context, userID, messageID int64) (*types.OrganizerMessageDetail, error)
	MarkOrganizerMessageRead(ctx context.Context, userID, messageID int64) error
	MarkAllOrganizerMessagesRead(ctx context.Context, userID int64) (*types.OrganizerReadAllResponse, error)
	GetOrganizerSubscriptionSummary(ctx context.Context, userID int64) (*types.OrganizerSubscriptionSummary, error)
	ListVenues(ctx context.Context, userID int64, keyword string, tagIDs []int64, page, size int) (*types.PageResponse[types.VenueListItem], error)
	GetVenueDetail(ctx context.Context, userID, venueID int64) (*types.VenueDetailResponse, error)
	ListVenueNotes(ctx context.Context, userID, venueID int64, cursor int64, pageSize int) (*types.VenueNotesResponse, error)
	FollowVenue(ctx context.Context, userID, venueID int64) error
	UnfollowVenue(ctx context.Context, userID, venueID int64) error
	SubscribeVenue(ctx context.Context, userID, venueID int64) error
	UnsubscribeVenue(ctx context.Context, userID, venueID int64) error
	ListSubscriptions(ctx context.Context, userID int64, subType string, page, size int) (*types.PageResponse[types.SubscriptionListItem], error)
	GetOrganizerProfile(ctx context.Context, userID int64) (*types.OrganizerProfileResponse, error)
	UpdateOrganizerProfile(ctx context.Context, userID int64, req types.OrganizerProfileRequest) error
	LookupOrganizerUser(ctx context.Context, userID int64, phone string) (*types.OrganizerUserLookupResponse, error)
	ListOrganizerPosts(ctx context.Context, userID int64, page, size int, keyword string, status *int, activityID int, storeID int64) (*types.PageResponse[types.OrganizerPostItem], error)
	SaveOrganizerPost(ctx context.Context, userID int64, postID uint64, req types.OrganizerPostRequest) (uint64, error)
	UpdateOrganizerPostVisibility(ctx context.Context, userID int64, postID uint64, req types.OrganizerPostVisibilityRequest) error
	DeleteOrganizerPost(ctx context.Context, userID int64, postID uint64) error
	GetOrganizerFinanceSummary(ctx context.Context, userID int64) (*types.OrganizerFinanceSummary, error)
	ListOrganizerFinanceFlows(ctx context.Context, userID int64, page, size int, flowType string) (*types.PageResponse[types.OrganizerFinanceFlowItem], error)
	ListOrganizerLevelRules(ctx context.Context, userID int64) ([]models.OrganizerLevelRule, error)
	SaveOrganizerLevelRule(ctx context.Context, userID, ruleID int64, req types.OrganizerLevelRuleRequest) (int64, error)
	DeleteOrganizerLevelRule(ctx context.Context, userID, ruleID int64) error
	ListOrganizerRoles(ctx context.Context, userID int64, page, size int) (*types.PageResponse[models.OrganizerRole], error)
	SaveOrganizerRole(ctx context.Context, userID, roleID int64, req types.OrganizerRoleRequest) (int64, error)
	DeleteOrganizerRole(ctx context.Context, userID, roleID int64) error
	ListOrganizerStaff(ctx context.Context, userID int64, page, size int) (*types.PageResponse[models.OrganizerStaff], error)
	SaveOrganizerStaff(ctx context.Context, userID, staffID int64, req types.OrganizerStaffRequest) (int64, error)
	DeleteOrganizerStaff(ctx context.Context, userID, staffID int64) error
	ListOrganizerOperationLogs(ctx context.Context, userID int64, page, size int, keyword string) (*types.PageResponse[models.OrganizerOperationLog], error)
	GetActivity(ctx context.Context, userID, activityID int64) (*types.ActivityDetailResponse, error)
	SubscribeActivity(ctx context.Context, userID, activityID int64) error
	UnsubscribeActivity(ctx context.Context, userID, activityID int64) error
	ListSubscribedActivities(ctx context.Context, userID int64, page, size int) (*types.PageResponse[types.ActivityListItem], error)
	SaveActivityStep(ctx context.Context, userID int64, req types.ActivityCreateRequest) (int64, error)
	GetMyActivities(ctx context.Context, userID int64, page, size int, filter types.ActivityListFilter) (*types.PageResponse[types.ActivityListItem], error)
	SearchActivities(ctx context.Context, userID int64, keyword string) ([]types.ActivityListItem, error)
	DeleteActivity(ctx context.Context, userID, activityID int64) error
	SubmitActivityAudit(ctx context.Context, userID, activityID int64) error
	GetActivityStatistics(ctx context.Context, userID, activityID int64) (*types.ActivityStatisticsResponse, error)
	GetActivityDailyStatistics(ctx context.Context, userID, activityID int64, days int) (*types.PageResponse[types.ActivityDailyStatisticsItem], error)
	GetTicketSpecs(ctx context.Context, activityID int64) ([]models.TicketSpec, error)
	SaveTicketSpecs(ctx context.Context, userID, activityID int64, specs []types.TicketSpecSaveItem) error
	DeleteTicketSpec(ctx context.Context, userID, specID int64) error
	CreateTicketOrder(ctx context.Context, userID int64, req types.CreateTicketOrderRequest) (*types.CreateTicketOrderResponse, error)
	ListTicketOrders(ctx context.Context, userID int64, status *int8, refundStatus string, page, size int) (*types.PageResponse[types.TicketOrderListItem], error)
	GetTicketOrderDetail(ctx context.Context, userID int64, orderNo string) (*types.TicketOrderDetailResponse, error)
	ListCancelReasons(ctx context.Context) ([]models.CancelReason, error)
	CancelTicketOrder(ctx context.Context, userID int64, orderNo string, reasonID int64) error
	CancelOrganizerTicketOrder(ctx context.Context, userID int64, orderNo string, reasonID int64) (*types.OrganizerCancelOrderResponse, error)
	DeleteTicketOrder(ctx context.Context, userID int64, orderNo string) error
	ListRefundReasons(ctx context.Context) ([]models.RefundReason, error)
	ApplyRefund(ctx context.Context, userID int64, req types.ApplyRefundRequest) (string, error)
	ListUserRefunds(ctx context.Context, userID int64, status *int8, page, size int) (*types.PageResponse[types.UserRefundListItem], error)
	ListOrganizerOrders(ctx context.Context, userID int64, activityID int64, status *int8, keyword, startDate, endDate string, page, size int) (*types.PageResponse[types.OrganizerOrderListItem], error)
	GetOrganizerOrderSummary(ctx context.Context, userID int64, startDate, endDate string) (*types.OrganizerOrderSummary, error)
	GetOrganizerOrderDetail(ctx context.Context, userID int64, orderNo string) (*types.OrganizerOrderDetailResponse, error)
	ListOrganizerRefunds(ctx context.Context, userID int64, status *int8, page, size int) (*types.PageResponse[types.OrganizerRefundListItem], error)
	GetOrganizerRefundDetail(ctx context.Context, userID int64, refundNo string) (*types.OrganizerRefundDetailResponse, error)
	GetRefundDetail(ctx context.Context, userID int64, refundNo string) (*types.RefundDetailResponse, error)
	RejectRefund(ctx context.Context, userID int64, refundNo string, req types.RejectRefundRequest) error
	CancelRefund(ctx context.Context, userID int64, refundNo string) error
	ListVerifiers(ctx context.Context, userID int64, page, size int) (*types.PageResponse[models.Verifier], error)
	AddVerifier(ctx context.Context, userID int64, req types.VerifierRequest) error
	UpdateVerifierStatus(ctx context.Context, userID, verifierID int64, status int8) error
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

var ErrOrganizerOrderCancelNotAllowed = errors.New("已支付订单请通过退款售后流程处理")

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
	resp := buildOrganizerInfo(org)
	if err := s.fillOrganizerLevelInfo(ctx, org.ID, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *TicketingService) ApplyOrganizer(ctx context.Context, userID int64, req types.OrganizerApplyRequest) (*types.OrganizerApplyResponse, error) {
	var org models.Organizer
	err := s.DB.WithContext(ctx).Where("user_id = ?", userID).First(&org).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	data := models.Organizer{
		UserID:   userID,
		Type:     models.OrganizerTypeMerchant,
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
			return nil, err
		}
		s.notifyOrganizerApply(ctx, data, userID)
		return &types.OrganizerApplyResponse{ApplicationID: data.ID, Status: data.Status, SubmittedAt: data.CreatedAt}, nil
	}
	if org.Status == models.OrganizerStatusAuditing {
		return nil, errors.New("入驻申请正在审核中，请勿重复提交")
	}
	if org.Status == models.OrganizerStatusApproved {
		return nil, errors.New("入驻申请已通过，无需重复提交")
	}
	if err := s.DB.WithContext(ctx).Model(&org).Updates(map[string]any{
		"name":          req.Name,
		"logo":          req.Logo,
		"status":        models.OrganizerStatusAuditing,
		"reject_reason": "",
		"level":         "LV1",
		"province":      req.Province,
		"city":          req.City,
		"district":      req.District,
	}).Error; err != nil {
		return nil, err
	}
	data.ID = org.ID
	data.CreatedAt = time.Now()
	s.notifyOrganizerApply(ctx, data, userID)
	return &types.OrganizerApplyResponse{ApplicationID: org.ID, Status: models.OrganizerStatusAuditing, SubmittedAt: time.Now()}, nil
}

func (s *TicketingService) GetOrganizerAuditStatus(ctx context.Context, userID int64) (*types.OrganizerAuditStatusResponse, error) {
	var org models.Organizer
	err := s.DB.WithContext(ctx).Where("user_id = ?", userID).First(&org).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &types.OrganizerAuditStatusResponse{Status: models.OrganizerStatusPending, RejectReason: ""}, nil
	}
	if err != nil {
		return nil, err
	}
	resp := &types.OrganizerAuditStatusResponse{
		OrganizerID:  org.ID,
		Type:         org.Type,
		Status:       org.Status,
		Enabled:      org.Enabled,
		RejectReason: org.RejectReason,
		SubmittedAt:  &org.CreatedAt,
	}
	if org.Status == models.OrganizerStatusApproved || org.Status == models.OrganizerStatusRejected {
		resp.ReviewedAt = &org.UpdatedAt
	}
	return resp, nil
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

func (s *TicketingService) GetWithdrawInfo(ctx context.Context, userID int64) (*types.OrganizerWithdrawInfoResponse, error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	funds, err := s.getOrganizerWithdrawFunds(ctx, s.DB.WithContext(ctx), org.ID)
	if err != nil {
		return nil, err
	}
	resp := &types.OrganizerWithdrawInfoResponse{
		BankAccountName:       org.BankAccountName,
		BankAccountNo:         org.BankAccountNo,
		BankName:              org.BankName,
		ContactName:           org.BankContactName,
		ContactPhone:          org.BankContactPhone,
		CanWithdraw:           org.BankAccountName != "" && org.BankAccountNo != "" && org.BankName != "" && funds.AvailableAmount > 0,
		GrossAmount:           funds.GrossAmount,
		RefundAmount:          funds.RefundAmount,
		WithdrawAmount:        funds.WithdrawAmount,
		PendingWithdrawAmount: funds.PendingWithdrawAmount,
		AvailableAmount:       funds.AvailableAmount,
		ArrivalCycle:          s.getPlatformSetting(ctx, "withdraw_arrival_cycle", "T+1 到 T+3 个工作日"),
	}

	var latest models.OrganizerBankAccountAudit
	err = s.DB.WithContext(ctx).
		Where("organizer_id = ?", org.ID).
		Order("id DESC").
		First(&latest).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, nil
		}
		return nil, err
	}
	info := buildOrganizerBankAuditInfo(latest)
	resp.LatestAudit = &info
	if latest.Status == models.OrganizerBankAuditStatusPending {
		resp.PendingAudit = &info
	}
	return resp, nil
}

func (s *TicketingService) UpdateWithdrawInfo(ctx context.Context, userID int64, req types.OrganizerWithdrawRequest) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	if org.Status != models.OrganizerStatusApproved {
		return errors.New("商家入驻通过后才能提交收款账户审核")
	}
	var pending int64
	if err := s.DB.WithContext(ctx).Model(&models.OrganizerBankAccountAudit{}).
		Where("organizer_id = ? AND status = ?", org.ID, models.OrganizerBankAuditStatusPending).
		Count(&pending).Error; err != nil {
		return err
	}
	if pending > 0 {
		return errors.New("已有待审核的收款账户申请，请勿重复提交")
	}
	audit := models.OrganizerBankAccountAudit{
		OrganizerID:      org.ID,
		UserID:           org.UserID,
		BankAccountName:  strings.TrimSpace(req.BankAccountName),
		BankAccountNo:    strings.TrimSpace(req.BankAccountNo),
		BankName:         strings.TrimSpace(req.BankName),
		BankContactName:  strings.TrimSpace(req.ContactName),
		BankContactPhone: strings.TrimSpace(req.ContactPhone),
		Status:           models.OrganizerBankAuditStatusPending,
	}
	if audit.BankAccountName == "" || audit.BankAccountNo == "" || audit.BankName == "" {
		return errors.New("收款人、收款账户、银行信息不能为空")
	}
	return s.DB.WithContext(ctx).Create(&audit).Error
}

func (s *TicketingService) ListOrganizerWithdraws(ctx context.Context, userID int64, status *int8, page, size int) (*types.PageResponse[models.OrganizerWithdraw], error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	page, size = normalizePage(page, size)
	query := s.DB.WithContext(ctx).Model(&models.OrganizerWithdraw{}).Where("organizer_id = ?", org.ID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []models.OrganizerWithdraw
	if err := query.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, err
	}
	return &types.PageResponse[models.OrganizerWithdraw]{List: list, Total: total}, nil
}

func (s *TicketingService) CreateOrganizerWithdraw(ctx context.Context, userID int64, req types.CreateOrganizerWithdrawRequest) (int64, error) {
	var withdrawID int64
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var org models.Organizer
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND status = ? AND enabled = 1", userID, models.OrganizerStatusApproved).
			First(&org).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("商家不存在、未通过审核或已被停用")
			}
			return err
		}
		if org.BankAccountName == "" || org.BankAccountNo == "" || org.BankName == "" {
			return errors.New("收款账户审核通过后才能提现")
		}
		var pending int64
		if err := tx.Model(&models.OrganizerWithdraw{}).
			Where("organizer_id = ? AND status = 0", org.ID).
			Count(&pending).Error; err != nil {
			return err
		}
		if pending > 0 {
			return errors.New("已有待审核提现申请，请勿重复提交")
		}
		funds, err := s.getOrganizerWithdrawFunds(ctx, tx, org.ID)
		if err != nil {
			return err
		}
		if funds.AvailableAmount <= 0 {
			return errors.New("暂无可提现金额")
		}
		if req.Amount > funds.AvailableAmount {
			return fmt.Errorf("提现金额不能超过可提现金额，可提现金额为%.2f元", float64(funds.AvailableAmount)/100)
		}
		withdraw := models.OrganizerWithdraw{
			OrganizerID:     org.ID,
			Amount:          req.Amount,
			BankAccountName: org.BankAccountName,
			BankAccountNo:   org.BankAccountNo,
			BankName:        org.BankName,
			Status:          0,
			Remark:          req.Remark,
		}
		if err := tx.Create(&withdraw).Error; err != nil {
			return err
		}
		withdrawID = withdraw.ID
		return nil
	})
	if err != nil {
		return 0, err
	}
	return withdrawID, nil
}

func (s *TicketingService) ListOrganizerCollections(ctx context.Context, userID int64, page, size int, keyword string) (*types.PageResponse[types.ActivityCollectionItem], error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	page, size = normalizePage(page, size)
	query := s.DB.WithContext(ctx).Model(&models.ActivityCollection{}).Where("organizer_id = ?", org.ID)
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		query = query.Where("title LIKE ? OR share_title LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []types.ActivityCollectionItem
	if err := query.Select("id, organizer_id, title, share_title, description, share_image, status, created_at, updated_at").
		Order("created_at DESC").Offset((page - 1) * size).Limit(size).Scan(&rows).Error; err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	if len(ids) > 0 {
		var counts []struct {
			CollectionID int64
			Count        int
		}
		if err := s.DB.WithContext(ctx).Model(&models.ActivityCollectionItem{}).
			Select("collection_id, COUNT(*) AS count").Where("collection_id IN ?", ids).Group("collection_id").Scan(&counts).Error; err != nil {
			return nil, err
		}
		countMap := map[int64]int{}
		for _, c := range counts {
			countMap[c.CollectionID] = c.Count
		}
		for i := range rows {
			rows[i].OrganizerName = org.Name
			rows[i].ActivityCount = countMap[rows[i].ID]
		}
	}
	return &types.PageResponse[types.ActivityCollectionItem]{List: rows, Total: total}, nil
}

func (s *TicketingService) GetOrganizerCollection(ctx context.Context, userID, collectionID int64) (*types.OrganizerCollectionDetail, error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	var item types.ActivityCollectionItem
	if err := s.DB.WithContext(ctx).Model(&models.ActivityCollection{}).
		Select("id, organizer_id, title, share_title, description, share_image, status, created_at, updated_at").
		Where("id = ? AND organizer_id = ?", collectionID, org.ID).Scan(&item).Error; err != nil {
		return nil, err
	}
	if item.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	item.OrganizerName = org.Name

	var links []models.ActivityCollectionItem
	if err := s.DB.WithContext(ctx).Where("collection_id = ?", collectionID).Order("sort ASC, id ASC").Find(&links).Error; err != nil {
		return nil, err
	}
	activityIDs := make([]int64, 0, len(links))
	for _, link := range links {
		activityIDs = append(activityIDs, link.ActivityID)
	}
	item.ActivityCount = len(activityIDs)

	detail := &types.OrganizerCollectionDetail{ActivityCollectionItem: item, ActivityIDs: activityIDs, Activities: []types.ActivityListItem{}}
	if len(activityIDs) == 0 {
		return detail, nil
	}
	var acts []models.Activity
	if err := s.DB.WithContext(ctx).Where("id IN ?", activityIDs).Find(&acts).Error; err != nil {
		return nil, err
	}
	actMap := make(map[int64]models.Activity, len(acts))
	for _, act := range acts {
		actMap[act.ID] = act
	}
	for _, id := range activityIDs {
		if act, ok := actMap[id]; ok {
			detail.Activities = append(detail.Activities, types.ActivityListItem{
				ID:         act.ID,
				Type:       defaultActivityType(act.Type),
				Name:       act.Name,
				PosterList: act.PosterList,
				StartTime:  act.StartTime,
				EndTime:    act.EndTime,
				Status:     act.Status,
			})
		}
	}
	return detail, nil
}

func (s *TicketingService) SaveOrganizerCollection(ctx context.Context, userID, collectionID int64, req types.OrganizerCollectionRequest) (int64, error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(req.Title) == "" {
		return 0, errors.New("合集标题不能为空")
	}
	status := req.Status
	if status == 0 {
		status = 1
	}
	if err := s.ensureActivitiesBelongToOrganizer(ctx, org.ID, req.ActivityIDs); err != nil {
		return 0, err
	}
	var savedID int64
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		collection := models.ActivityCollection{
			OrganizerID: org.ID,
			Title:       req.Title,
			ShareTitle:  req.ShareTitle,
			Description: req.Description,
			ShareImage:  req.ShareImage,
			Status:      status,
		}
		if collectionID > 0 {
			result := tx.Model(&models.ActivityCollection{}).Where("id = ? AND organizer_id = ?", collectionID, org.ID).Updates(map[string]any{
				"title": req.Title, "share_title": req.ShareTitle, "description": req.Description, "share_image": req.ShareImage, "status": status,
			})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
			savedID = collectionID
			if err := tx.Where("collection_id = ?", collectionID).Delete(&models.ActivityCollectionItem{}).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Create(&collection).Error; err != nil {
				return err
			}
			savedID = collection.ID
		}
		for i, activityID := range req.ActivityIDs {
			if err := tx.Create(&models.ActivityCollectionItem{CollectionID: savedID, ActivityID: activityID, Sort: i}).Error; err != nil {
				return err
			}
		}
		return s.createOrganizerLog(tx, org.ID, userID, "save_collection", "activity_collection", "", "", fmt.Sprintf("collection_id=%d", savedID))
	})
	return savedID, err
}

func (s *TicketingService) DeleteOrganizerCollection(ctx context.Context, userID, collectionID int64) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var collection models.ActivityCollection
		if err := tx.Where("id = ? AND organizer_id = ?", collectionID, org.ID).First(&collection).Error; err != nil {
			return err
		}
		if err := tx.Where("collection_id = ?", collectionID).Delete(&models.ActivityCollectionItem{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&collection).Error; err != nil {
			return err
		}
		return s.createOrganizerLog(tx, org.ID, userID, "delete_collection", "activity_collection", "", "", fmt.Sprintf("collection_id=%d", collectionID))
	})
}

func (s *TicketingService) ListOrganizerMessages(ctx context.Context, userID int64, page, size int, unreadOnly bool) (*types.PageResponse[types.OrganizerMessageItem], error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	page, size = normalizePage(page, size)
	query := s.DB.WithContext(ctx).Table("platform_messages pm").
		Joins("LEFT JOIN organizer_message_reads omr ON omr.message_id = pm.id AND omr.organizer_id = ?", org.ID).
		Joins("LEFT JOIN platform_message_deliveries pmd ON pmd.message_id = pm.id AND pmd.user_id = ?", userID).
		Where("pm.status = 1 AND (pm.target IN ? OR pm.target = ?) AND (pmd.id IS NOT NULL OR NOT EXISTS (SELECT 1 FROM platform_message_deliveries any_delivery WHERE any_delivery.message_id = pm.id))", []string{"merchant", "organizer", "business", "all"}, "")
	if unreadOnly {
		query = query.Where("COALESCE(omr.is_read,0) = 0")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []struct {
		ID          int64
		Title       string
		Content     string
		ContentType string
		CoverImage  string
		Type        string
		Target      string
		IsRead      int8
		ReadAt      *time.Time
		CreatedAt   time.Time
	}
	if err := query.Select("pm.id, pm.title, pm.content, pm.content_type, pm.cover_image, pm.type, pm.target, COALESCE(omr.is_read,0) AS is_read, omr.read_at, pm.created_at").
		Order("pm.created_at DESC").Offset((page - 1) * size).Limit(size).Scan(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]types.OrganizerMessageItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, types.OrganizerMessageItem{ID: row.ID, Title: row.Title, Content: row.Content, ContentType: row.ContentType, CoverImage: row.CoverImage, Type: row.Type, Target: row.Target, IsRead: row.IsRead == 1, ReadAt: row.ReadAt, CreatedAt: row.CreatedAt})
	}
	return &types.PageResponse[types.OrganizerMessageItem]{List: list, Total: total}, nil
}

func (s *TicketingService) GetOrganizerMessageDetail(ctx context.Context, userID, messageID int64) (*types.OrganizerMessageDetail, error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	var row struct {
		ID          int64
		Title       string
		Content     string
		ContentType string
		CoverImage  string
		MediaData   string
		Type        string
		Target      string
		IsRead      int8
		ReadAt      *time.Time
		CreatedAt   time.Time
	}
	query := s.DB.WithContext(ctx).Table("platform_messages pm").
		Joins("LEFT JOIN organizer_message_reads omr ON omr.message_id = pm.id AND omr.organizer_id = ?", org.ID).
		Joins("LEFT JOIN platform_message_deliveries pmd ON pmd.message_id = pm.id AND pmd.user_id = ?", userID).
		Where("pm.id = ? AND pm.status = 1 AND (pm.target IN ? OR pm.target = ?) AND (pmd.id IS NOT NULL OR NOT EXISTS (SELECT 1 FROM platform_message_deliveries any_delivery WHERE any_delivery.message_id = pm.id))", messageID, []string{"merchant", "organizer", "business", "all"}, "")
	if err := query.Select("pm.id, pm.title, pm.content, pm.content_type, pm.cover_image, pm.media_data, pm.type, pm.target, COALESCE(omr.is_read,0) AS is_read, omr.read_at, pm.created_at").Limit(1).Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	if row.IsRead == 0 {
		if err := s.MarkOrganizerMessageRead(ctx, userID, messageID); err != nil {
			return nil, err
		}
		now := time.Now()
		row.IsRead = 1
		row.ReadAt = &now
	}
	resp := &types.OrganizerMessageDetail{OrganizerMessageItem: types.OrganizerMessageItem{
		ID: row.ID, Title: row.Title, Content: row.Content, ContentType: row.ContentType, CoverImage: row.CoverImage,
		Type: row.Type, Target: row.Target, IsRead: row.IsRead == 1, ReadAt: row.ReadAt, CreatedAt: row.CreatedAt,
	}, MediaData: []string{}}
	if row.MediaData != "" {
		_ = json.Unmarshal([]byte(row.MediaData), &resp.MediaData)
	}
	return resp, nil
}

func (s *TicketingService) MarkOrganizerMessageRead(ctx context.Context, userID, messageID int64) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	now := time.Now()
	row := models.OrganizerMessageRead{OrganizerID: org.ID, MessageID: messageID, IsRead: 1, ReadAt: &now}
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "organizer_id"}, {Name: "message_id"}}, DoUpdates: clause.Assignments(map[string]any{"is_read": 1, "read_at": now, "updated_at": now})}).Create(&row).Error; err != nil {
			return err
		}
		return tx.Model(&models.PlatformMessageDelivery{}).
			Where("message_id = ? AND user_id = ?", messageID, userID).
			Updates(map[string]any{"status": 3, "read_at": now}).Error
	})
}

func (s *TicketingService) MarkAllOrganizerMessagesRead(ctx context.Context, userID int64) (*types.OrganizerReadAllResponse, error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	var messageIDs []int64
	if err := s.DB.WithContext(ctx).Model(&models.PlatformMessage{}).
		Where("status = 1 AND (target IN ? OR target = ?) AND (EXISTS (SELECT 1 FROM platform_message_deliveries d WHERE d.message_id = platform_messages.id AND d.user_id = ?) OR NOT EXISTS (SELECT 1 FROM platform_message_deliveries any_delivery WHERE any_delivery.message_id = platform_messages.id))", []string{"merchant", "organizer", "business", "all"}, "", userID).
		Pluck("id", &messageIDs).Error; err != nil {
		return nil, err
	}
	if len(messageIDs) == 0 {
		return &types.OrganizerReadAllResponse{UpdatedCount: 0}, nil
	}
	now := time.Now()
	rows := make([]models.OrganizerMessageRead, 0, len(messageIDs))
	for _, messageID := range messageIDs {
		rows = append(rows, models.OrganizerMessageRead{OrganizerID: org.ID, MessageID: messageID, IsRead: 1, ReadAt: &now})
	}
	if err := s.DB.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "organizer_id"}, {Name: "message_id"}}, DoUpdates: clause.Assignments(map[string]any{"is_read": 1, "read_at": now, "updated_at": now})}).Create(&rows).Error; err != nil {
		return nil, err
	}
	if err := s.DB.WithContext(ctx).Model(&models.PlatformMessageDelivery{}).
		Where("message_id IN ? AND user_id = ?", messageIDs, userID).
		Updates(map[string]any{"status": 3, "read_at": now}).Error; err != nil {
		return nil, err
	}
	return &types.OrganizerReadAllResponse{UpdatedCount: int64(len(messageIDs))}, nil
}

func (s *TicketingService) GetOrganizerSubscriptionSummary(ctx context.Context, userID int64) (*types.OrganizerSubscriptionSummary, error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp := &types.OrganizerSubscriptionSummary{}
	base := func() *gorm.DB {
		return s.DB.WithContext(ctx).Table("activity_subscriptions sub").
			Joins("JOIN activities a ON a.id = sub.activity_id").
			Where("a.organizer_id = ?", org.ID)
	}
	if err := base().Count(&resp.TotalSubscriptions).Error; err != nil {
		return nil, err
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if err := base().Where("sub.created_at >= ?", today).Count(&resp.TodaySubscriptions).Error; err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *TicketingService) ListVenues(ctx context.Context, userID int64, keyword string, tagIDs []int64, page, size int) (*types.PageResponse[types.VenueListItem], error) {
	page, size = normalizePage(page, size)
	query := s.DB.WithContext(ctx).Table("organizers o").
		Select(`o.id, o.user_id, o.name, o.logo, o.province, o.city, o.district, o.created_at,
			COALESCE(p.cover_image, '') AS cover_image,
			COALESCE(p.description, '') AS description,
			COALESCE(p.business_hours, '') AS business_hours,
			COALESCE(p.service_phone, '') AS service_phone,
			COALESCE(p.address, '') AS address,
			COALESCE(p.latitude, 0) AS latitude,
			COALESCE(p.longitude, 0) AS longitude,
			COALESCE(p.average_spend, 0) AS average_spend`).
		Joins("LEFT JOIN organizer_profiles p ON p.organizer_id = o.id").
		Where(visibleVenueOrganizerSQL(), models.OrganizerStatusApproved, models.ActivityTypeVenue, models.ActivityStatusOnline)
	query = dao.ApplyContentTagFilter(query, models.ContentTagTargetVenue, "o.id", tagIDs)
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("o.name LIKE ? OR p.address LIKE ? OR p.description LIKE ? OR o.district LIKE ?", like, like, like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []types.VenueListItem
	if err := query.Order("o.created_at DESC").Offset((page - 1) * size).Limit(size).Scan(&list).Error; err != nil {
		return nil, err
	}
	if err := s.fillVenueStats(ctx, userID, list); err != nil {
		return nil, err
	}
	return &types.PageResponse[types.VenueListItem]{List: list, Total: total}, nil
}

func (s *TicketingService) GetVenueDetail(ctx context.Context, userID, venueID int64) (*types.VenueDetailResponse, error) {
	var item types.VenueListItem
	if err := s.DB.WithContext(ctx).Table("organizers o").
		Select(`o.id, o.user_id, o.name, o.logo, o.province, o.city, o.district, o.created_at,
			COALESCE(p.cover_image, '') AS cover_image,
			COALESCE(p.description, '') AS description,
			COALESCE(p.business_hours, '') AS business_hours,
			COALESCE(p.service_phone, '') AS service_phone,
			COALESCE(p.address, '') AS address,
			COALESCE(p.latitude, 0) AS latitude,
			COALESCE(p.longitude, 0) AS longitude,
			COALESCE(p.average_spend, 0) AS average_spend`).
		Joins("LEFT JOIN organizer_profiles p ON p.organizer_id = o.id").
		Where("o.id = ?", venueID).
		Where(visibleVenueOrganizerSQL(), models.OrganizerStatusApproved, models.ActivityTypeVenue, models.ActivityStatusOnline).
		First(&item).Error; err != nil {
		return nil, err
	}
	list := []types.VenueListItem{item}
	if err := s.fillVenueStats(ctx, userID, list); err != nil {
		return nil, err
	}
	resp := &types.VenueDetailResponse{VenueListItem: list[0], Gallery: []string{}, Stores: []models.OrganizerStore{}}
	var profile models.OrganizerProfile
	if err := s.DB.WithContext(ctx).Where("organizer_id = ?", venueID).First(&profile).Error; err == nil && profile.Gallery != "" {
		_ = json.Unmarshal([]byte(profile.Gallery), &resp.Gallery)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err := s.DB.WithContext(ctx).Where("organizer_id = ?", venueID).Order("created_at DESC").Find(&resp.Stores).Error; err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *TicketingService) ListVenueNotes(ctx context.Context, userID, venueID int64, cursor int64, pageSize int) (*types.VenueNotesResponse, error) {
	if err := s.ensureVenueVisible(ctx, venueID); err != nil {
		return nil, err
	}
	if pageSize <= 0 {
		pageSize = types.DefaultPageSize
	}
	if pageSize > 100 {
		pageSize = 100
	}
	var storeIDs []int64
	if err := s.DB.WithContext(ctx).Model(&models.OrganizerStore{}).Where("organizer_id = ?", venueID).Pluck("id", &storeIDs).Error; err != nil {
		return nil, err
	}
	if len(storeIDs) == 0 {
		return &types.VenueNotesResponse{Notes: []types.VenueNoteItem{}}, nil
	}
	query := s.DB.WithContext(ctx).Table("notes n").
		Select(`n.id, n.user_id, n.title, n.content, n.type, n.media_data, n.activity_id, n.store_id,
			n.created_at, n.updated_at, COALESCE(u.avatar, '') AS avatar, COALESCE(u.nickname, '') AS nickname`).
		Joins("LEFT JOIN users u ON u.id = n.user_id").
		Where("n.store_id IN ? AND n.status <> ? AND n.visible_conf = ?", storeIDs, -1, types.VisibleConfPublic)
	if cursor > 0 {
		query = query.Where("n.created_at < ?", time.Unix(0, cursor))
	}
	var rows []struct {
		ID         int64
		UserID     int64
		Title      string
		Content    string
		Type       int
		MediaData  string
		ActivityID int
		StoreID    int64
		Avatar     string
		Nickname   string
		CreatedAt  time.Time
		UpdatedAt  time.Time
	}
	if err := query.Order("n.created_at DESC").Limit(pageSize + 1).Scan(&rows).Error; err != nil {
		return nil, err
	}
	resp := &types.VenueNotesResponse{Notes: []types.VenueNoteItem{}}
	displayCount := len(rows)
	if displayCount > pageSize {
		displayCount = pageSize
		resp.HasMore = true
	}
	for i := 0; i < displayCount; i++ {
		row := rows[i]
		media := []types.NoteMedia{}
		if row.MediaData != "" {
			_ = json.Unmarshal([]byte(row.MediaData), &media)
		}
		resp.Notes = append(resp.Notes, types.VenueNoteItem{
			ID: row.ID, UserID: row.UserID, Title: row.Title, Content: row.Content, Type: row.Type,
			MediaData: media, ActivityID: row.ActivityID, StoreID: row.StoreID,
			Avatar: row.Avatar, Nickname: row.Nickname, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
			TimeStamp: row.CreatedAt.UnixNano(),
		})
	}
	if displayCount > 0 {
		resp.NextCursor = rows[displayCount-1].CreatedAt.UnixNano()
	}
	return resp, nil
}

func (s *TicketingService) FollowVenue(ctx context.Context, userID, venueID int64) error {
	return dao.FollowContent(ctx, s.DB, userID, models.ContentFollowTargetVenue, venueID)
}

func (s *TicketingService) UnfollowVenue(ctx context.Context, userID, venueID int64) error {
	return dao.UnfollowContent(ctx, s.DB, userID, models.ContentFollowTargetVenue, venueID)
}

func (s *TicketingService) SubscribeVenue(ctx context.Context, userID, venueID int64) error {
	if err := s.ensureVenueVisible(ctx, venueID); err != nil {
		return err
	}
	sub := models.VenueSubscription{OrganizerID: venueID, UserID: userID}
	return s.DB.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&sub).Error
}

func (s *TicketingService) UnsubscribeVenue(ctx context.Context, userID, venueID int64) error {
	return s.DB.WithContext(ctx).Where("organizer_id = ? AND user_id = ?", venueID, userID).Delete(&models.VenueSubscription{}).Error
}

func (s *TicketingService) ListSubscriptions(ctx context.Context, userID int64, subType string, page, size int) (*types.PageResponse[types.SubscriptionListItem], error) {
	page, size = normalizePage(page, size)
	subType = strings.TrimSpace(subType)
	if subType == "" {
		subType = "all"
	}
	if subType != "all" && subType != "activity" && subType != "venue" {
		return nil, errors.New("type 仅支持 all、activity、venue")
	}

	items := make([]types.SubscriptionListItem, 0)
	if subType == "all" || subType == "activity" {
		var rows []struct {
			ID           int64
			Name         string
			PosterList   string
			Description  string
			StartTime    time.Time
			EndTime      time.Time
			Status       int8
			Address      string
			Latitude     float64
			Longitude    float64
			SubscribedAt time.Time
		}
		if err := s.DB.WithContext(ctx).Table("activity_subscriptions sub").
			Select(`a.id, a.name, a.poster_list, a.description, a.start_time, a.end_time, a.status,
				a.address, a.latitude, a.longitude, sub.created_at AS subscribed_at`).
			Joins("JOIN activities a ON a.id = sub.activity_id").
			Where("sub.user_id = ?", userID).
			Scan(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			startTime := row.StartTime
			endTime := row.EndTime
			items = append(items, types.SubscriptionListItem{
				ID: fmt.Sprintf("activity-%d", row.ID), Source: "activity", SourceID: row.ID, Title: row.Name,
				CoverImage: row.PosterList, Description: row.Description, StartTime: &startTime, EndTime: &endTime,
				Status: row.Status, Address: row.Address, Latitude: row.Latitude, Longitude: row.Longitude,
				SubscribedAt: row.SubscribedAt,
			})
		}
	}
	if subType == "all" || subType == "venue" {
		var rows []struct {
			ID            int64
			UserID        int64
			Name          string
			Logo          string
			CoverImage    string
			Description   string
			BusinessHours string
			ServicePhone  string
			Province      string
			City          string
			District      string
			Address       string
			Latitude      float64
			Longitude     float64
			AverageSpend  int64
			SubscribedAt  time.Time
		}
		if err := s.DB.WithContext(ctx).Table("venue_subscriptions sub").
			Select(`o.id, o.user_id, o.name, o.logo,
				COALESCE(p.cover_image, '') AS cover_image,
				COALESCE(p.description, '') AS description,
				COALESCE(p.business_hours, '') AS business_hours,
				COALESCE(p.service_phone, '') AS service_phone,
				o.province, o.city, o.district,
				COALESCE(p.address, '') AS address,
				COALESCE(p.latitude, 0) AS latitude,
				COALESCE(p.longitude, 0) AS longitude,
				COALESCE(p.average_spend, 0) AS average_spend,
				sub.created_at AS subscribed_at`).
			Joins("JOIN organizers o ON o.id = sub.organizer_id").
			Joins("LEFT JOIN organizer_profiles p ON p.organizer_id = o.id").
			Where("sub.user_id = ?", userID).
			Where(visibleVenueOrganizerSQL(), models.OrganizerStatusApproved, models.ActivityTypeVenue, models.ActivityStatusOnline).
			Scan(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			items = append(items, types.SubscriptionListItem{
				ID: fmt.Sprintf("venue-%d", row.ID), Source: "venue", SourceID: row.ID, Title: row.Name,
				CoverImage: firstNonEmpty(row.CoverImage, row.Logo), Description: row.Description,
				Address: row.Address, Latitude: row.Latitude, Longitude: row.Longitude,
				Extra: map[string]any{
					"user_id":        row.UserID,
					"logo":           row.Logo,
					"business_hours": row.BusinessHours,
					"service_phone":  row.ServicePhone,
					"province":       row.Province,
					"city":           row.City,
					"district":       row.District,
					"average_spend":  row.AverageSpend,
				},
				SubscribedAt: row.SubscribedAt,
			})
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].SubscribedAt.After(items[j].SubscribedAt)
	})
	total := int64(len(items))
	start := (page - 1) * size
	if start >= len(items) {
		return &types.PageResponse[types.SubscriptionListItem]{List: []types.SubscriptionListItem{}, Total: total}, nil
	}
	end := start + size
	if end > len(items) {
		end = len(items)
	}
	return &types.PageResponse[types.SubscriptionListItem]{List: items[start:end], Total: total}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func visibleVenueOrganizerSQL() string {
	return `o.status = ? AND o.enabled = 1 AND EXISTS (
		SELECT 1 FROM activities va
		WHERE va.organizer_id = o.id AND va.type = ? AND va.status = ?
	)`
}

func (s *TicketingService) fillVenueStats(ctx context.Context, userID int64, venues []types.VenueListItem) error {
	if len(venues) == 0 {
		return nil
	}
	venueIDs := make([]int64, 0, len(venues))
	for _, venue := range venues {
		venueIDs = append(venueIDs, venue.ID)
	}

	followCounts, followed, err := dao.LoadContentFollowStats(ctx, s.DB, models.ContentFollowTargetVenue, venueIDs, userID)
	if err != nil {
		return err
	}

	subCounts := map[int64]int64{}
	var subRows []struct {
		OrganizerID int64
		Count       int64
	}
	if err := s.DB.WithContext(ctx).Model(&models.VenueSubscription{}).
		Select("organizer_id, COUNT(*) AS count").
		Where("organizer_id IN ?", venueIDs).
		Group("organizer_id").
		Scan(&subRows).Error; err != nil {
		return err
	}
	for _, row := range subRows {
		subCounts[row.OrganizerID] = row.Count
	}

	postCounts := map[int64]int64{}
	var postRows []struct {
		OrganizerID int64
		Count       int64
	}
	if err := s.DB.WithContext(ctx).Table("notes n").
		Select("os.organizer_id, COUNT(n.id) AS count").
		Joins("JOIN organizer_stores os ON os.id = n.store_id").
		Where("os.organizer_id IN ? AND n.status <> ? AND n.visible_conf = ?", venueIDs, -1, types.VisibleConfPublic).
		Group("os.organizer_id").
		Scan(&postRows).Error; err != nil {
		return err
	}
	for _, row := range postRows {
		postCounts[row.OrganizerID] = row.Count
	}

	subscribed := map[int64]bool{}
	if userID > 0 {
		var subscribedIDs []int64
		if err := s.DB.WithContext(ctx).Model(&models.VenueSubscription{}).
			Where("user_id = ? AND organizer_id IN ?", userID, venueIDs).
			Pluck("organizer_id", &subscribedIDs).Error; err != nil {
			return err
		}
		for _, id := range subscribedIDs {
			subscribed[id] = true
		}
	}

	for i := range venues {
		venues[i].FollowCount = followCounts[venues[i].ID]
		venues[i].SubscribeCount = subCounts[venues[i].ID]
		venues[i].PostCount = postCounts[venues[i].ID]
		venues[i].IsFollow = followed[venues[i].ID]
		venues[i].FollowTargetType = models.ContentFollowTargetVenue
		venues[i].FollowTargetID = venues[i].ID
		venues[i].IsSubscribe = subscribed[venues[i].ID]
	}
	tagMap, err := dao.LoadContentTags(ctx, s.DB, models.ContentTagTargetVenue, venueIDs, false)
	if err != nil {
		return err
	}
	for i := range venues {
		tags := tagMap[venues[i].ID]
		venues[i].TagIDs = types.ContentTagIDs(tags)
		venues[i].Tags = types.BuildContentTagItems(tags)
	}
	return nil
}

func (s *TicketingService) ensureVenueVisible(ctx context.Context, venueID int64) error {
	_, err := s.getVisibleVenue(ctx, venueID)
	return err
}

func (s *TicketingService) getVisibleVenue(ctx context.Context, venueID int64) (*models.Organizer, error) {
	var venue models.Organizer
	if err := s.DB.WithContext(ctx).Table("organizers o").
		Where("o.id = ?", venueID).
		Where(visibleVenueOrganizerSQL(), models.OrganizerStatusApproved, models.ActivityTypeVenue, models.ActivityStatusOnline).
		First(&venue).Error; err != nil {
		return nil, err
	}
	return &venue, nil
}

func (s *TicketingService) GetOrganizerProfile(ctx context.Context, userID int64) (*types.OrganizerProfileResponse, error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	var profile models.OrganizerProfile
	err = s.DB.WithContext(ctx).Where("organizer_id = ?", org.ID).First(&profile).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	resp := &types.OrganizerProfileResponse{
		ID:            org.ID,
		Name:          org.Name,
		Logo:          org.Logo,
		Province:      org.Province,
		City:          org.City,
		District:      org.District,
		CoverImage:    profile.CoverImage,
		Description:   profile.Description,
		BusinessHours: profile.BusinessHours,
		ContactName:   profile.ContactName,
		ServicePhone:  profile.ServicePhone,
		Address:       profile.Address,
		Latitude:      profile.Latitude,
		Longitude:     profile.Longitude,
		AverageSpend:  profile.AverageSpend,
		Gallery:       []string{},
	}
	if profile.Gallery != "" {
		_ = json.Unmarshal([]byte(profile.Gallery), &resp.Gallery)
	}
	return resp, nil
}

func (s *TicketingService) UpdateOrganizerProfile(ctx context.Context, userID int64, req types.OrganizerProfileRequest) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	if req.Latitude != 0 || req.Longitude != 0 {
		if req.Latitude < 18 || req.Latitude > 54 || req.Longitude < 73 || req.Longitude > 135 {
			return errors.New("经纬度不在中国大陆常用范围内")
		}
	}
	gallery, _ := json.Marshal(req.Gallery)
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		orgUpdates := map[string]any{}
		if strings.TrimSpace(req.Name) != "" {
			orgUpdates["name"] = req.Name
		}
		if req.Logo != "" {
			orgUpdates["logo"] = req.Logo
		}
		if req.Province != "" {
			orgUpdates["province"] = req.Province
		}
		if req.City != "" {
			orgUpdates["city"] = req.City
		}
		if req.District != "" {
			orgUpdates["district"] = req.District
		}
		if len(orgUpdates) > 0 {
			if err := tx.Model(&models.Organizer{}).Where("id = ?", org.ID).Updates(orgUpdates).Error; err != nil {
				return err
			}
		}
		profile := models.OrganizerProfile{
			OrganizerID:   org.ID,
			CoverImage:    req.CoverImage,
			Gallery:       string(gallery),
			Description:   req.Description,
			BusinessHours: req.BusinessHours,
			ContactName:   req.ContactName,
			ServicePhone:  req.ServicePhone,
			Address:       req.Address,
			Latitude:      req.Latitude,
			Longitude:     req.Longitude,
			AverageSpend:  req.AverageSpend,
		}
		updates := map[string]any{
			"cover_image": req.CoverImage, "gallery": string(gallery), "description": req.Description,
			"business_hours": req.BusinessHours, "contact_name": req.ContactName, "service_phone": req.ServicePhone,
			"address": req.Address, "latitude": req.Latitude, "longitude": req.Longitude, "average_spend": req.AverageSpend,
			"updated_at": time.Now(),
		}
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "organizer_id"}}, DoUpdates: clause.Assignments(updates)}).Create(&profile).Error; err != nil {
			return err
		}
		return s.createOrganizerLog(tx, org.ID, userID, "update_profile", "organizer_profile", "", "", "")
	})
}

func (s *TicketingService) LookupOrganizerUser(ctx context.Context, userID int64, phone string) (*types.OrganizerUserLookupResponse, error) {
	if _, err := s.findOrganizerByUser(ctx, userID); err != nil {
		return nil, err
	}
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil, errors.New("手机号不能为空")
	}
	var user models.Users
	if err := s.DB.WithContext(ctx).Where("mobile = ?", phone).First(&user).Error; err != nil {
		return nil, err
	}
	return &types.OrganizerUserLookupResponse{
		UserID:   int64(user.Id),
		Phone:    user.Mobile,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Status:   user.Status,
	}, nil
}

func (s *TicketingService) ListOrganizerPosts(ctx context.Context, userID int64, page, size int, keyword string, status *int, activityID int, storeID int64) (*types.PageResponse[types.OrganizerPostItem], error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	page, size = normalizePage(page, size)
	query := s.DB.WithContext(ctx).Table("notes n").
		Select("n.*, ns.like_count, ns.coll_count, ns.share_count, ns.comment_count, a.name AS activity_name, os.name AS store_name").
		Joins("LEFT JOIN note_stats ns ON ns.note_id = n.id").
		Joins("LEFT JOIN activities a ON a.id = n.activity_id").
		Joins("LEFT JOIN organizer_stores os ON os.id = n.store_id").
		Where("n.user_id = ? AND n.status <> ?", org.UserID, -1)
	if status != nil {
		query = query.Where("n.status = ?", *status)
	}
	if activityID > 0 {
		query = query.Where("n.activity_id = ?", activityID)
	}
	if storeID > 0 {
		query = query.Where("n.store_id = ?", storeID)
	}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("n.title LIKE ? OR n.content LIKE ?", like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []struct {
		models.Note
		LikeCount    int64
		CollCount    int64
		ShareCount   int64
		CommentCount int64
		ActivityName string
		StoreName    string
	}
	if err := query.Order("n.created_at DESC").Offset((page - 1) * size).Limit(size).Scan(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]types.OrganizerPostItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, buildOrganizerPostItem(row.Note, row.LikeCount, row.CollCount, row.ShareCount, row.CommentCount, row.ActivityName, row.StoreName))
	}
	return &types.PageResponse[types.OrganizerPostItem]{List: list, Total: total}, nil
}

func (s *TicketingService) SaveOrganizerPost(ctx context.Context, userID int64, postID uint64, req types.OrganizerPostRequest) (uint64, error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(req.Title) == "" {
		return 0, errors.New("动态标题不能为空")
	}
	var publisher models.Users
	if err := s.DB.WithContext(ctx).First(&publisher, org.UserID).Error; err != nil {
		return 0, err
	}
	if s.WeChatService == nil {
		return 0, ErrContentSafetyUnavailable
	}
	contentForCheck := strings.TrimSpace(req.Title + "\n" + req.Content)
	if err := s.WeChatService.CheckTextSecurity(ctx, contentForCheck, publisher.OpenID, publisher.Nickname, req.Title, 4); err != nil {
		return 0, err
	}
	if err := s.ensureOrganizerPostRelations(ctx, org.ID, req.ActivityID, req.StoreID); err != nil {
		return 0, err
	}
	if req.Type == 0 {
		req.Type = 1
	}
	if req.Type != 1 && req.Type != 2 {
		return 0, errors.New("动态类型无效")
	}
	if req.VisibleConf == 0 {
		req.VisibleConf = types.VisibleConfPublic
	}
	if req.Status == 0 {
		req.Status = 1
	}
	if req.MediaData == nil && len(req.Images) > 0 {
		req.MediaData = make([]types.NoteMedia, 0, len(req.Images))
		for _, image := range req.Images {
			req.MediaData = append(req.MediaData, types.NoteMedia{URL: image})
		}
	}
	mediaJSON, _ := json.Marshal(req.MediaData)
	locationJSON := "{}"
	if req.Location != nil {
		if b, err := json.Marshal(req.Location); err == nil {
			locationJSON = string(b)
		}
	}
	now := time.Now()
	if postID > 0 {
		result := s.DB.WithContext(ctx).Model(&models.Note{}).
			Where("id = ? AND user_id = ? AND status <> ?", postID, org.UserID, -1).
			Updates(map[string]any{
				"title": req.Title, "content": req.Content, "media_data": string(mediaJSON), "location": locationJSON,
				"type": req.Type, "status": req.Status, "visible_conf": req.VisibleConf,
				"activity_id": req.ActivityID, "store_id": req.StoreID, "updated_at": now,
			})
		if result.Error != nil {
			return 0, result.Error
		}
		if result.RowsAffected == 0 {
			return 0, gorm.ErrRecordNotFound
		}
		_ = s.createOrganizerLog(s.DB.WithContext(ctx), org.ID, userID, "save_post", "post", "", "", fmt.Sprintf("post_id=%d", postID))
		return postID, nil
	}
	note := models.Note{
		ID:          uint64(snowflake.GenUserID()),
		UserID:      uint64(org.UserID),
		Title:       req.Title,
		Content:     req.Content,
		TopicIDs:    "[]",
		Location:    locationJSON,
		MediaData:   string(mediaJSON),
		Type:        req.Type,
		Status:      req.Status,
		VisibleConf: req.VisibleConf,
		ActivityID:  req.ActivityID,
		StoreID:     req.StoreID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&note).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.NoteStats{NoteID: note.ID, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
			return err
		}
		return s.createOrganizerLog(tx, org.ID, userID, "save_post", "post", "", "", fmt.Sprintf("post_id=%d", note.ID))
	})
	return note.ID, err
}

func (s *TicketingService) UpdateOrganizerPostVisibility(ctx context.Context, userID int64, postID uint64, req types.OrganizerPostVisibilityRequest) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	status := 0
	if req.Status != nil {
		status = *req.Status
	} else if req.Visible != nil {
		if *req.Visible {
			status = 1
		} else {
			status = 2
		}
	} else {
		return errors.New("请提交 visible 或 status")
	}
	if status != 0 && status != 1 && status != 2 && status != 3 {
		return errors.New("动态状态无效")
	}
	result := s.DB.WithContext(ctx).Model(&models.Note{}).
		Where("id = ? AND user_id = ? AND status <> ?", postID, org.UserID, -1).
		Updates(map[string]any{"status": status, "updated_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return s.createOrganizerLog(s.DB.WithContext(ctx), org.ID, userID, "update_post_visibility", "post", "", "", fmt.Sprintf("post_id=%d,status=%d", postID, status))
}

func (s *TicketingService) DeleteOrganizerPost(ctx context.Context, userID int64, postID uint64) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	result := s.DB.WithContext(ctx).Model(&models.Note{}).
		Where("id = ? AND user_id = ? AND status <> ?", postID, org.UserID, -1).
		Updates(map[string]any{"status": -1, "updated_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return s.createOrganizerLog(s.DB.WithContext(ctx), org.ID, userID, "delete_post", "post", "", "", fmt.Sprintf("post_id=%d", postID))
}

func (s *TicketingService) GetOrganizerFinanceSummary(ctx context.Context, userID int64) (*types.OrganizerFinanceSummary, error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	funds, err := s.getOrganizerWithdrawFunds(ctx, s.DB.WithContext(ctx), org.ID)
	if err != nil {
		return nil, err
	}
	resp := &types.OrganizerFinanceSummary{
		GrossAmount:    funds.GrossAmount,
		RefundAmount:   funds.RefundAmount,
		WithdrawAmount: funds.WithdrawAmount,
		SettleAmount:   funds.GrossAmount - funds.RefundAmount - funds.WithdrawAmount,
	}
	paidStatuses := []int8{models.TicketOrderStatusUsable, models.TicketOrderStatusUsed, models.TicketOrderStatusRefunding, models.TicketOrderStatusRefundSuccess, models.TicketOrderStatusRefundReject}
	if err := s.DB.WithContext(ctx).Table("ticket_orders o").Joins("JOIN activities a ON a.id = o.activity_id").
		Where("a.organizer_id = ? AND o.status IN ?", org.ID, paidStatuses).
		Select("COUNT(o.id)").Scan(&resp.OrderCount).Error; err != nil {
		return nil, err
	}
	today := time.Now().Truncate(24 * time.Hour)
	if err := s.DB.WithContext(ctx).Table("ticket_orders o").Joins("JOIN activities a ON a.id = o.activity_id").
		Where("a.organizer_id = ? AND o.status IN ? AND o.pay_time >= ?", org.ID, paidStatuses, today).
		Select("COUNT(o.id) AS today_order_count, COALESCE(SUM(o.actual_price),0) AS today_order_amount, COALESCE(SUM(o.quantity),0) AS today_ticket_count").
		Scan(resp).Error; err != nil {
		return nil, err
	}
	return resp, nil
}

type organizerWithdrawFunds struct {
	GrossAmount           int64
	RefundAmount          int64
	WithdrawAmount        int64
	PendingWithdrawAmount int64
	AvailableAmount       int64
}

func (s *TicketingService) getOrganizerWithdrawFunds(ctx context.Context, db *gorm.DB, organizerID int64) (*organizerWithdrawFunds, error) {
	resp := &organizerWithdrawFunds{}
	paidStatuses := []int8{models.TicketOrderStatusUsable, models.TicketOrderStatusUsed, models.TicketOrderStatusRefunding, models.TicketOrderStatusRefundSuccess, models.TicketOrderStatusRefundReject}
	if err := db.WithContext(ctx).Table("ticket_orders o").Joins("JOIN activities a ON a.id = o.activity_id").
		Where("a.organizer_id = ? AND o.status IN ?", organizerID, paidStatuses).
		Select("COALESCE(SUM(o.actual_price),0)").Scan(&resp.GrossAmount).Error; err != nil {
		return nil, err
	}
	if err := db.WithContext(ctx).Table("refunds r").Joins("JOIN ticket_orders o ON o.id = r.order_id").Joins("JOIN activities a ON a.id = o.activity_id").
		Where("a.organizer_id = ? AND r.status = ?", organizerID, models.RefundStatusSuccess).
		Select("COALESCE(SUM(r.refund_amount),0)").Scan(&resp.RefundAmount).Error; err != nil {
		return nil, err
	}
	if err := db.WithContext(ctx).Model(&models.OrganizerWithdraw{}).Where("organizer_id = ? AND status = 1", organizerID).Select("COALESCE(SUM(amount),0)").Scan(&resp.WithdrawAmount).Error; err != nil {
		return nil, err
	}
	if err := db.WithContext(ctx).Model(&models.OrganizerWithdraw{}).Where("organizer_id = ? AND status = 0", organizerID).Select("COALESCE(SUM(amount),0)").Scan(&resp.PendingWithdrawAmount).Error; err != nil {
		return nil, err
	}
	resp.AvailableAmount = resp.GrossAmount - resp.RefundAmount - resp.WithdrawAmount - resp.PendingWithdrawAmount
	if resp.AvailableAmount < 0 {
		resp.AvailableAmount = 0
	}
	return resp, nil
}

func (s *TicketingService) ListOrganizerFinanceFlows(ctx context.Context, userID int64, page, size int, flowType string) (*types.PageResponse[types.OrganizerFinanceFlowItem], error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	page, size = normalizePage(page, size)
	flows := make([]types.OrganizerFinanceFlowItem, 0)
	if flowType == "" || flowType == "order" {
		var orders []models.TicketOrder
		_ = s.DB.WithContext(ctx).Joins("JOIN activities a ON a.id = ticket_orders.activity_id").
			Where("a.organizer_id = ? AND ticket_orders.status IN ?", org.ID, []int8{models.TicketOrderStatusUsable, models.TicketOrderStatusUsed, models.TicketOrderStatusRefunding, models.TicketOrderStatusRefundSuccess, models.TicketOrderStatusRefundReject}).
			Order("ticket_orders.created_at DESC").Find(&orders).Error
		for _, order := range orders {
			flows = append(flows, types.OrganizerFinanceFlowItem{ID: fmt.Sprintf("order-%s", order.OrderNo), Type: "order", Amount: order.ActualPrice, OrderNo: order.OrderNo, ActivityID: order.ActivityID, Description: "票务订单收入", CreatedAt: order.CreatedAt})
		}
	}
	if flowType == "" || flowType == "refund" {
		var rows []struct {
			RefundNo     string
			RefundAmount int64
			OrderNo      string
			ActivityID   int64
			CreatedAt    time.Time
		}
		_ = s.DB.WithContext(ctx).Table("refunds r").Select("r.refund_no, r.refund_amount, o.order_no, o.activity_id, r.created_at").
			Joins("JOIN ticket_orders o ON o.id = r.order_id").Joins("JOIN activities a ON a.id = o.activity_id").
			Where("a.organizer_id = ? AND r.status = ?", org.ID, models.RefundStatusSuccess).Scan(&rows).Error
		for _, row := range rows {
			flows = append(flows, types.OrganizerFinanceFlowItem{ID: fmt.Sprintf("refund-%s", row.RefundNo), Type: "refund", Amount: -row.RefundAmount, OrderNo: row.OrderNo, RefundNo: row.RefundNo, ActivityID: row.ActivityID, Description: "订单退款", CreatedAt: row.CreatedAt})
		}
	}
	if flowType == "" || flowType == "withdraw" {
		var withdraws []models.OrganizerWithdraw
		_ = s.DB.WithContext(ctx).Where("organizer_id = ?", org.ID).Find(&withdraws).Error
		for _, w := range withdraws {
			flows = append(flows, types.OrganizerFinanceFlowItem{ID: fmt.Sprintf("withdraw-%d", w.ID), Type: "withdraw", Amount: -w.Amount, Description: "商家提现", CreatedAt: w.CreatedAt})
		}
	}
	sortFinanceFlows(flows)
	total := int64(len(flows))
	start := (page - 1) * size
	if start >= len(flows) {
		return &types.PageResponse[types.OrganizerFinanceFlowItem]{List: []types.OrganizerFinanceFlowItem{}, Total: total}, nil
	}
	end := start + size
	if end > len(flows) {
		end = len(flows)
	}
	return &types.PageResponse[types.OrganizerFinanceFlowItem]{List: flows[start:end], Total: total}, nil
}

func (s *TicketingService) ListOrganizerLevelRules(ctx context.Context, userID int64) ([]models.OrganizerLevelRule, error) {
	if _, err := s.findOrganizerByUser(ctx, userID); err != nil {
		return nil, err
	}
	if err := s.ensureDefaultLevelRules(ctx); err != nil {
		return nil, err
	}
	var rules []models.OrganizerLevelRule
	err := s.DB.WithContext(ctx).Order("level ASC").Find(&rules).Error
	return rules, err
}

func (s *TicketingService) SaveOrganizerLevelRule(ctx context.Context, userID, ruleID int64, req types.OrganizerLevelRuleRequest) (int64, error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return 0, err
	}
	if req.Level <= 0 {
		return 0, errors.New("等级必须大于0")
	}
	if req.Name == "" {
		req.Name = fmt.Sprintf("LV%d", req.Level)
	}
	status := req.Status
	if status == 0 {
		status = 1
	}
	rule := models.OrganizerLevelRule{Level: req.Level, Name: req.Name, FeeRate: req.FeeRate, RequiredActivityCount: req.RequiredActivityCount, Description: req.Description, Benefits: req.Benefits, Status: status}
	if ruleID > 0 {
		result := s.DB.WithContext(ctx).Model(&models.OrganizerLevelRule{}).Where("id = ?", ruleID).Updates(rule)
		if result.Error != nil {
			return 0, result.Error
		}
		if result.RowsAffected == 0 {
			return 0, gorm.ErrRecordNotFound
		}
		_ = s.createOrganizerLog(s.DB.WithContext(ctx), org.ID, userID, "save_level_rule", "level_rule", "", "", fmt.Sprintf("rule_id=%d", ruleID))
		return ruleID, nil
	}
	if err := s.DB.WithContext(ctx).Create(&rule).Error; err != nil {
		return 0, err
	}
	_ = s.createOrganizerLog(s.DB.WithContext(ctx), org.ID, userID, "save_level_rule", "level_rule", "", "", fmt.Sprintf("rule_id=%d", rule.ID))
	return rule.ID, nil
}

func (s *TicketingService) DeleteOrganizerLevelRule(ctx context.Context, userID, ruleID int64) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	if err := s.DB.WithContext(ctx).Delete(&models.OrganizerLevelRule{}, ruleID).Error; err != nil {
		return err
	}
	return s.createOrganizerLog(s.DB.WithContext(ctx), org.ID, userID, "delete_level_rule", "level_rule", "", "", fmt.Sprintf("rule_id=%d", ruleID))
}

func (s *TicketingService) ListOrganizerRoles(ctx context.Context, userID int64, page, size int) (*types.PageResponse[models.OrganizerRole], error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	page, size = normalizePage(page, size)
	query := s.DB.WithContext(ctx).Model(&models.OrganizerRole{}).Where("organizer_id = ?", org.ID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []models.OrganizerRole
	err = query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return &types.PageResponse[models.OrganizerRole]{List: list, Total: total}, err
}

func (s *TicketingService) SaveOrganizerRole(ctx context.Context, userID, roleID int64, req types.OrganizerRoleRequest) (int64, error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return 0, err
	}
	permissions, _ := json.Marshal(req.Permissions)
	status := req.Status
	if status == 0 {
		status = 1
	}
	role := models.OrganizerRole{OrganizerID: org.ID, Name: req.Name, Description: req.Description, Permissions: string(permissions), Status: status}
	if roleID > 0 {
		result := s.DB.WithContext(ctx).Model(&models.OrganizerRole{}).Where("id = ? AND organizer_id = ?", roleID, org.ID).Updates(map[string]any{"name": req.Name, "description": req.Description, "permissions": string(permissions), "status": status})
		if result.Error != nil {
			return 0, result.Error
		}
		if result.RowsAffected == 0 {
			return 0, gorm.ErrRecordNotFound
		}
		_ = s.createOrganizerLog(s.DB.WithContext(ctx), org.ID, userID, "save_role", "role", "", "", fmt.Sprintf("role_id=%d", roleID))
		return roleID, nil
	}
	if err := s.DB.WithContext(ctx).Create(&role).Error; err != nil {
		return 0, err
	}
	_ = s.createOrganizerLog(s.DB.WithContext(ctx), org.ID, userID, "save_role", "role", "", "", fmt.Sprintf("role_id=%d", role.ID))
	return role.ID, nil
}

func (s *TicketingService) DeleteOrganizerRole(ctx context.Context, userID, roleID int64) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.OrganizerStaff{}).Where("organizer_id = ? AND role_id = ?", org.ID, roleID).Update("role_id", 0).Error; err != nil {
			return err
		}
		result := tx.Where("id = ? AND organizer_id = ?", roleID, org.ID).Delete(&models.OrganizerRole{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return s.createOrganizerLog(tx, org.ID, userID, "delete_role", "role", "", "", fmt.Sprintf("role_id=%d", roleID))
	})
}

func (s *TicketingService) ListOrganizerStaff(ctx context.Context, userID int64, page, size int) (*types.PageResponse[models.OrganizerStaff], error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	page, size = normalizePage(page, size)
	query := s.DB.WithContext(ctx).Model(&models.OrganizerStaff{}).Where("organizer_id = ?", org.ID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []models.OrganizerStaff
	err = query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return &types.PageResponse[models.OrganizerStaff]{List: list, Total: total}, err
}

func (s *TicketingService) SaveOrganizerStaff(ctx context.Context, userID, staffID int64, req types.OrganizerStaffRequest) (int64, error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return 0, err
	}
	if req.RoleID > 0 {
		var count int64
		if err := s.DB.WithContext(ctx).Model(&models.OrganizerRole{}).Where("id = ? AND organizer_id = ?", req.RoleID, org.ID).Count(&count).Error; err != nil {
			return 0, err
		}
		if count == 0 {
			return 0, errors.New("角色不存在")
		}
	}
	status := req.Status
	if status == 0 {
		status = 1
	}
	staff := models.OrganizerStaff{OrganizerID: org.ID, UserID: req.UserID, RoleID: req.RoleID, Name: req.Name, Phone: req.Phone, Status: status}
	if staffID > 0 {
		result := s.DB.WithContext(ctx).Model(&models.OrganizerStaff{}).Where("id = ? AND organizer_id = ?", staffID, org.ID).Updates(map[string]any{"user_id": req.UserID, "role_id": req.RoleID, "name": req.Name, "phone": req.Phone, "status": status})
		if result.Error != nil {
			return 0, result.Error
		}
		if result.RowsAffected == 0 {
			return 0, gorm.ErrRecordNotFound
		}
		_ = s.createOrganizerLog(s.DB.WithContext(ctx), org.ID, userID, "save_staff", "staff", "", "", fmt.Sprintf("staff_id=%d", staffID))
		return staffID, nil
	}
	if err := s.DB.WithContext(ctx).Create(&staff).Error; err != nil {
		return 0, err
	}
	_ = s.createOrganizerLog(s.DB.WithContext(ctx), org.ID, userID, "save_staff", "staff", "", "", fmt.Sprintf("staff_id=%d", staff.ID))
	return staff.ID, nil
}

func (s *TicketingService) DeleteOrganizerStaff(ctx context.Context, userID, staffID int64) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	result := s.DB.WithContext(ctx).Where("id = ? AND organizer_id = ?", staffID, org.ID).Delete(&models.OrganizerStaff{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return s.createOrganizerLog(s.DB.WithContext(ctx), org.ID, userID, "delete_staff", "staff", "", "", fmt.Sprintf("staff_id=%d", staffID))
}

func (s *TicketingService) ListOrganizerOperationLogs(ctx context.Context, userID int64, page, size int, keyword string) (*types.PageResponse[models.OrganizerOperationLog], error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	page, size = normalizePage(page, size)
	query := s.DB.WithContext(ctx).Model(&models.OrganizerOperationLog{}).Where("organizer_id = ?", org.ID)
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		query = query.Where("action LIKE ? OR resource LIKE ? OR remark LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []models.OrganizerOperationLog
	err = query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return &types.PageResponse[models.OrganizerOperationLog]{List: list, Total: total}, err
}

func (s *TicketingService) SaveActivityStep(ctx context.Context, userID int64, req types.ActivityCreateRequest) (int64, error) {
	org, err := s.ensureOrganizer(ctx, userID)
	if err != nil {
		return 0, err
	}
	if req.Type != nil {
		if _, err := normalizeActivityType(*req.Type); err != nil {
			return 0, err
		}
	}
	var act models.Activity
	if req.ActivityID > 0 {
		if err := s.DB.WithContext(ctx).Where("id = ? AND organizer_id = ?", req.ActivityID, org.ID).First(&act).Error; err != nil {
			return 0, err
		}
	} else {
		act = models.Activity{
			OrganizerID: org.ID,
			Type:        models.ActivityTypeParty,
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

func (s *TicketingService) GetActivity(ctx context.Context, userID, activityID int64) (*types.ActivityDetailResponse, error) {
	var act models.Activity
	if err := s.DB.WithContext(ctx).First(&act, activityID).Error; err != nil {
		return nil, err
	}
	var specs []models.TicketSpec
	if err := s.DB.WithContext(ctx).Where("activity_id = ?", activityID).Order("id asc").Find(&specs).Error; err != nil {
		return nil, err
	}
	var org models.Organizer
	if err := s.DB.WithContext(ctx).First(&org, act.OrganizerID).Error; err != nil {
		return nil, err
	}
	targetType, targetID := models.ContentTagTargetActivity, act.ID
	if defaultActivityType(act.Type) == models.ActivityTypeVenue {
		targetType, targetID = models.ContentTagTargetVenue, act.OrganizerID
	}
	tagMap, err := dao.LoadContentTags(ctx, s.DB, targetType, []int64{targetID}, false)
	if err != nil {
		return nil, err
	}
	tags := tagMap[targetID]
	followCounts, followed, err := dao.LoadContentFollowStats(ctx, s.DB, contentFollowTargetForActivity(act), []int64{contentFollowIDForActivity(act)}, userID)
	if err != nil {
		return nil, err
	}
	resp := &types.ActivityDetailResponse{
		Activity:         act,
		UserID:           org.UserID,
		TagIDs:           types.ContentTagIDs(tags),
		Tags:             types.BuildContentTagItems(tags),
		TicketSpecs:      specs,
		Organizer:        &org,
		IsFollow:         followed[contentFollowIDForActivity(act)],
		FollowCount:      followCounts[contentFollowIDForActivity(act)],
		FollowTargetType: contentFollowTargetForActivity(act),
		FollowTargetID:   contentFollowIDForActivity(act),
	}
	if userID > 0 {
		var count int64
		_ = s.DB.WithContext(ctx).Model(&models.ActivitySubscription{}).
			Where("activity_id = ? AND user_id = ?", activityID, userID).
			Count(&count).Error
		resp.IsSubscribe = count > 0
	}
	return resp, nil
}

func (s *TicketingService) SubscribeActivity(ctx context.Context, userID, activityID int64) error {
	var count int64
	if err := s.DB.WithContext(ctx).Model(&models.Activity{}).Where("id = ?", activityID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}
	sub := models.ActivitySubscription{ActivityID: activityID, UserID: userID}
	return s.DB.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&sub).Error
}

func (s *TicketingService) UnsubscribeActivity(ctx context.Context, userID, activityID int64) error {
	return s.DB.WithContext(ctx).
		Where("activity_id = ? AND user_id = ?", activityID, userID).
		Delete(&models.ActivitySubscription{}).Error
}

func (s *TicketingService) ListSubscribedActivities(ctx context.Context, userID int64, page, size int) (*types.PageResponse[types.ActivityListItem], error) {
	page, size = normalizePage(page, size)
	query := s.DB.WithContext(ctx).Table("activity_subscriptions AS sub").
		Joins("JOIN activities a ON a.id = sub.activity_id").
		Where("sub.user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var rows []struct {
		ID          int64
		OrganizerID int64
		Type        string
		Name        string
		PosterList  string
		StartTime   time.Time
		EndTime     time.Time
		Status      int8
	}
	if err := query.
		Select("a.id, a.organizer_id, a.type, a.name, a.poster_list, a.start_time, a.end_time, a.status").
		Order("sub.created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	activities := make([]models.Activity, 0, len(rows))
	for _, row := range rows {
		activities = append(activities, models.Activity{ID: row.ID, OrganizerID: row.OrganizerID, Type: row.Type})
	}
	followCounts, followed, err := s.loadActivityFollowStats(ctx, userID, activities)
	if err != nil {
		return nil, err
	}

	list := make([]types.ActivityListItem, 0, len(rows))
	for _, row := range rows {
		activity := models.Activity{ID: row.ID, OrganizerID: row.OrganizerID, Type: row.Type}
		list = append(list, types.ActivityListItem{
			ID:               row.ID,
			Type:             defaultActivityType(row.Type),
			Name:             row.Name,
			PosterList:       row.PosterList,
			StartTime:        row.StartTime,
			EndTime:          row.EndTime,
			Status:           row.Status,
			IsSubscribe:      true,
			IsFollow:         followed[row.ID],
			FollowCount:      followCounts[row.ID],
			FollowTargetType: contentFollowTargetForActivity(activity),
			FollowTargetID:   contentFollowIDForActivity(activity),
		})
	}
	return &types.PageResponse[types.ActivityListItem]{List: list, Total: total}, nil
}

func (s *TicketingService) GetMyActivities(ctx context.Context, userID int64, page, size int, filter types.ActivityListFilter) (*types.PageResponse[types.ActivityListItem], error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &types.PageResponse[types.ActivityListItem]{List: []types.ActivityListItem{}}, nil
		}
		return nil, err
	}
	query := s.DB.WithContext(ctx).Model(&models.Activity{}).Where("organizer_id = ?", org.ID)
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR address LIKE ?", like, like)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.PublishedFrom != nil {
		query = query.Where("created_at >= ?", *filter.PublishedFrom)
	}
	if filter.PublishedTo != nil {
		query = query.Where("created_at <= ?", *filter.PublishedTo)
	}
	if filter.ActivityFrom != nil {
		query = query.Where("end_time >= ?", *filter.ActivityFrom)
	}
	if filter.ActivityTo != nil {
		query = query.Where("start_time <= ?", *filter.ActivityTo)
	}
	return s.listActivities(query, page, size, userID)
}

func (s *TicketingService) SearchActivities(ctx context.Context, userID int64, keyword string) ([]types.ActivityListItem, error) {
	query := s.DB.WithContext(ctx).Model(&models.Activity{}).Where("status = ?", models.ActivityStatusOnline)
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}
	resp, err := s.listActivities(query, 1, 50, userID)
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
			rule, err := loadPointsRule(ctx, tx)
			if err != nil {
				return err
			}
			pointsDiscount = pointsAmount * rule.DiscountCentsPerPoint
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
		orderStatus := models.TicketOrderStatusPending
		payMethod := ""
		var payTime *time.Time
		if actualPrice == 0 {
			orderStatus = models.TicketOrderStatusUsable
			payMethod = zeroPayMethod(pointsAmount)
			nowPaid := time.Now()
			payTime = &nowPaid
		}
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
			PayMethod:      payMethod,
			PayTime:        payTime,
			BuyerName:      buyerName,
			BuyerIDCard:    buyerIDCard,
			Status:         orderStatus,
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
		result.Status = orderStatus
		result.QRCode = qrContent
		result.QRCodeURL = qrURL
		result.Viewers = orderViewerItems(orderViewers, false)
		return nil
	})
	return result, err
}

const (
	pointsDiscountSettingKey = "points_discount_cents_per_point"
	pointsRewardSettingKey   = "points_reward_cents_per_point"
)

func (s *TicketingService) GetPointsRule(ctx context.Context) (*types.PointsRule, error) {
	return loadPointsRule(ctx, s.DB)
}

func loadPointsRule(ctx context.Context, db *gorm.DB) (*types.PointsRule, error) {
	rule := &types.PointsRule{
		DiscountCentsPerPoint: 10,
		RewardCentsPerPoint:   1000,
	}
	var rows []models.PlatformSetting
	if err := db.WithContext(ctx).
		Where("setting_key IN ?", []string{pointsDiscountSettingKey, pointsRewardSettingKey}).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		value, err := strconv.ParseInt(strings.TrimSpace(row.Value), 10, 64)
		if err != nil || value <= 0 {
			continue
		}
		switch row.Key {
		case pointsDiscountSettingKey:
			rule.DiscountCentsPerPoint = value
		case pointsRewardSettingKey:
			rule.RewardCentsPerPoint = value
		}
	}
	return rule, nil
}

func (s *TicketingService) getPlatformSetting(ctx context.Context, key, fallback string) string {
	var setting models.PlatformSetting
	err := s.DB.WithContext(ctx).Where("setting_key = ?", key).First(&setting).Error
	if err != nil {
		return fallback
	}
	value := strings.TrimSpace(setting.Value)
	if value == "" {
		return fallback
	}
	return value
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
	if err := s.expireUserPendingTicketOrders(ctx, userID); err != nil {
		return nil, err
	}
	var order models.TicketOrder
	if err := s.DB.WithContext(ctx).Where("order_no = ? AND user_id = ? AND user_deleted_at IS NULL", orderNo, userID).First(&order).Error; err != nil {
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

func (s *TicketingService) ListTicketOrders(ctx context.Context, userID int64, status *int8, refundStatus string, page, size int) (*types.PageResponse[types.TicketOrderListItem], error) {
	if err := s.expireUserPendingTicketOrders(ctx, userID); err != nil {
		return nil, err
	}
	page, size = normalizePage(page, size)
	query := s.DB.WithContext(ctx).Model(&models.TicketOrder{}).Where("user_id = ? AND user_deleted_at IS NULL", userID)
	if status != nil {
		switch *status {
		case models.TicketOrderStatusPending:
			query = query.Where("status = ? AND actual_price > 0", *status)
		case models.TicketOrderStatusUsable:
			query = query.Where("(status = ? OR (status = ? AND actual_price = 0))", *status, models.TicketOrderStatusPending)
		default:
			query = query.Where("status = ?", *status)
		}
	}
	if refundStatus = strings.TrimSpace(refundStatus); refundStatus != "" {
		refundStatusValue, ok := refundStatusValue(refundStatus)
		if !ok {
			return nil, fmt.Errorf("售后状态参数错误")
		}
		query = query.Where(`EXISTS (
			SELECT 1 FROM refunds r
			WHERE r.order_id = ticket_orders.id
			AND r.id = (SELECT MAX(r2.id) FROM refunds r2 WHERE r2.order_id = ticket_orders.id)
			AND r.status = ?
		)`, refundStatusValue)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var rows []struct {
		ID             int64
		OrderNo        string
		Status         int8
		TotalPrice     int64
		ActualPrice    int64
		PointsAmount   int64
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
		RefundNo       string
		RefundStatus   *int8
	}
	err := query.
		Select(`ticket_orders.id,
			ticket_orders.order_no,
			ticket_orders.status,
			ticket_orders.total_price,
			ticket_orders.actual_price,
			ticket_orders.points_amount,
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
			ticket_orders.pay_time,
			lr.refund_no,
			lr.status AS refund_status`).
		Joins("LEFT JOIN activities ON activities.id = ticket_orders.activity_id").
		Joins("LEFT JOIN ticket_specs ON ticket_specs.id = ticket_orders.ticket_spec_id").
		Joins("LEFT JOIN refunds lr ON lr.order_id = ticket_orders.id AND lr.id = (SELECT MAX(r2.id) FROM refunds r2 WHERE r2.order_id = ticket_orders.id)").
		Order("ticket_orders.created_at desc").
		Offset((page - 1) * size).
		Limit(size).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	list := make([]types.TicketOrderListItem, 0, len(rows))
	orderNos := make([]string, 0, len(rows))
	orderIDs := make([]int64, 0, len(rows))
	for _, r := range rows {
		orderNos = append(orderNos, r.OrderNo)
		orderIDs = append(orderIDs, r.ID)
	}
	viewersByOrderNo, err := s.orderViewersByOrderNo(ctx, orderNos)
	if err != nil {
		return nil, err
	}
	refundByOrderID, err := s.latestRefundByOrderID(ctx, orderIDs)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		status := r.Status
		latestRefund, hasRefund := refundByOrderID[r.ID]
		if hasRefund && latestRefund.Status == models.RefundStatusSuccess && status != models.TicketOrderStatusRefundSuccess {
			status = models.TicketOrderStatusRefundSuccess
			_ = s.DB.WithContext(ctx).Model(&models.TicketOrder{}).
				Where("id = ? AND status <> ?", r.ID, models.TicketOrderStatusRefundSuccess).
				Update("status", models.TicketOrderStatusRefundSuccess).Error
		}
		if status == models.TicketOrderStatusPending && r.ActualPrice == 0 {
			status = models.TicketOrderStatusUsable
			_ = s.markZeroPayOrderUsable(ctx, r.ID, r.PointsAmount)
		}
		item := types.TicketOrderListItem{
			OrderNo:     r.OrderNo,
			Status:      status,
			RefundNo:    r.RefundNo,
			TotalPrice:  r.TotalPrice,
			ActualPrice: r.ActualPrice,
			Quantity:    r.Quantity,
			BuyerName:   r.BuyerName,
			BuyerIDCard: maskIDCard(r.BuyerIDCard),
			CreatedAt:   r.CreatedAt,
			ExpireTime:  r.ExpireTime,
			PayTime:     r.PayTime,
		}
		if hasRefund {
			item.RefundNo = latestRefund.RefundNo
			item.RefundStatus = refundStatusCode(latestRefund.Status)
			item.RefundStatusText = refundStatusText(latestRefund.Status)
		} else if r.RefundStatus != nil {
			item.RefundStatus = refundStatusCode(*r.RefundStatus)
			item.RefundStatusText = refundStatusText(*r.RefundStatus)
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
		return cancelPendingTicketOrderTx(tx, &order, reason.Reason)
	})
}

func (s *TicketingService) CancelOrganizerTicketOrder(ctx context.Context, userID int64, orderNo string, reasonID int64) (*types.OrganizerCancelOrderResponse, error) {
	var resp *types.OrganizerCancelOrderResponse
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		org, err := s.findOrganizerByUser(ctx, userID)
		if err != nil {
			return err
		}

		var order models.TicketOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Table("ticket_orders o").
			Select("o.*").
			Joins("JOIN activities a ON a.id = o.activity_id").
			Where("o.order_no = ? AND a.organizer_id = ?", orderNo, org.ID).
			First(&order).Error; err != nil {
			return err
		}

		// An already-cancelled order is a no-op so retries cannot release inventory twice.
		if order.Status == models.TicketOrderStatusCancelled {
			resp = &types.OrganizerCancelOrderResponse{
				OrderNo:        order.OrderNo,
				Status:         order.Status,
				CancelReasonID: reasonID,
				CancelledAt:    order.UpdatedAt,
			}
			return nil
		}
		if order.Status != models.TicketOrderStatusPending {
			return ErrOrganizerOrderCancelNotAllowed
		}

		var reason models.CancelReason
		_ = tx.First(&reason, reasonID).Error
		if err := cancelPendingTicketOrderTx(tx, &order, reason.Reason); err != nil {
			return err
		}
		now := time.Now()
		if err := s.createOrganizerLog(tx, org.ID, userID, "cancel_pending_order", "ticket_order", "POST", "/api/v1/organizer/orders/:order_no/cancel", fmt.Sprintf("order_no=%s,reason_id=%d", order.OrderNo, reasonID)); err != nil {
			return err
		}
		resp = &types.OrganizerCancelOrderResponse{
			OrderNo:        order.OrderNo,
			Status:         models.TicketOrderStatusCancelled,
			CancelReasonID: reasonID,
			CancelledAt:    now,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *TicketingService) DeleteTicketOrder(ctx context.Context, userID int64, orderNo string) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order models.TicketOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("order_no = ? AND user_id = ? AND user_deleted_at IS NULL", orderNo, userID).
			First(&order).Error; err != nil {
			return err
		}
		//switch order.Status {
		//case models.TicketOrderStatusUsed, models.TicketOrderStatusCancelled, models.TicketOrderStatusRefundSuccess:
		//default:
		//	return errors.New("当前订单不可删除")
		//}
		return tx.Model(&order).Update("user_deleted_at", time.Now()).Error
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
		return nil
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

// GetOrganizerOrderSummary aggregates every successful order owned by the
// current organizer. It deliberately does not reuse the paginated order list,
// so dashboard totals stay correct when the organizer has more than one page.
func (s *TicketingService) GetOrganizerOrderSummary(ctx context.Context, userID int64, startDate, endDate string) (*types.OrganizerOrderSummary, error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	startDate = strings.TrimSpace(startDate)
	endDate = strings.TrimSpace(endDate)
	var start, end time.Time
	if startDate != "" {
		start, err = parseDateStart(startDate)
		if err != nil {
			return nil, fmt.Errorf("开始日期格式错误")
		}
	}
	if endDate != "" {
		end, err = parseDateEnd(endDate)
		if err != nil {
			return nil, fmt.Errorf("结束日期格式错误")
		}
	}

	// "成交" only includes orders that remain paid and usable/used. Refund
	// states are excluded, so this value describes current completed sales.
	base := func() *gorm.DB {
		query := s.DB.WithContext(ctx).Table("ticket_orders o").
			Joins("JOIN activities a ON a.id = o.activity_id").
			Where("a.organizer_id = ? AND o.status IN ?", org.ID, []int8{
				models.TicketOrderStatusUsable,
				models.TicketOrderStatusUsed,
			})
		if !start.IsZero() {
			query = query.Where("o.pay_time >= ?", start)
		}
		if !end.IsZero() {
			query = query.Where("o.pay_time < ?", end)
		}
		return query
	}

	resp := &types.OrganizerOrderSummary{ActivityRanks: make([]types.OrganizerOrderActivityRank, 0)}
	if err := base().
		Select("COALESCE(SUM(o.actual_price), 0) AS total_amount, COUNT(o.id) AS order_count, COALESCE(SUM(o.quantity), 0) AS ticket_count").
		Scan(resp).Error; err != nil {
		return nil, err
	}
	if resp.OrderCount > 0 {
		resp.AverageOrderAmount = resp.TotalAmount / resp.OrderCount
	}

	if err := base().
		Select("o.activity_id, a.name AS activity_name, COUNT(o.id) AS order_count, COALESCE(SUM(o.quantity), 0) AS ticket_count, COALESCE(SUM(o.actual_price), 0) AS total_amount").
		Group("o.activity_id, a.name").
		Order("total_amount DESC, order_count DESC, o.activity_id DESC").
		Scan(&resp.ActivityRanks).Error; err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *TicketingService) GetOrganizerOrderDetail(ctx context.Context, userID int64, orderNo string) (*types.OrganizerOrderDetailResponse, error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	var order models.TicketOrder
	if err := s.DB.WithContext(ctx).Table("ticket_orders o").
		Select("o.*").
		Joins("JOIN activities a ON a.id = o.activity_id").
		Where("o.order_no = ? AND a.organizer_id = ?", orderNo, org.ID).
		First(&order).Error; err != nil {
		return nil, err
	}
	detail, err := s.buildOrderDetail(ctx, order)
	if err != nil {
		return nil, err
	}
	resp := &types.OrganizerOrderDetailResponse{TicketOrderDetailResponse: *detail, UserID: order.UserID}
	var user models.Users
	if err := s.DB.WithContext(ctx).Where("id = ?", order.UserID).First(&user).Error; err == nil {
		resp.UserName = user.Nickname
		resp.UserMobile = maskPhone(user.Mobile)
		resp.UserAvatar = user.Avatar
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return resp, nil
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

func (s *TicketingService) GetOrganizerRefundDetail(ctx context.Context, userID int64, refundNo string) (*types.OrganizerRefundDetailResponse, error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var refund models.Refund
	if err := s.DB.WithContext(ctx).Table("refunds r").
		Select("r.*").
		Joins("JOIN ticket_orders o ON o.id = r.order_id").
		Joins("JOIN activities a ON a.id = o.activity_id").
		Where("r.refund_no = ? AND a.organizer_id = ?", refundNo, org.ID).
		First(&refund).Error; err != nil {
		return nil, err
	}

	var order models.TicketOrder
	if err := s.DB.WithContext(ctx).Where("id = ?", refund.OrderID).First(&order).Error; err != nil {
		return nil, err
	}
	var activity models.Activity
	if err := s.DB.WithContext(ctx).Where("id = ? AND organizer_id = ?", order.ActivityID, org.ID).First(&activity).Error; err != nil {
		return nil, err
	}
	var ticketSpec models.TicketSpec
	if err := s.DB.WithContext(ctx).Where("id = ?", order.TicketSpecID).First(&ticketSpec).Error; err != nil {
		return nil, err
	}

	resp := &types.OrganizerRefundDetailResponse{}
	resp.Refund.RefundNo = refund.RefundNo
	resp.Refund.RefundAmount = refund.RefundAmount
	resp.Refund.DeductAmount = refund.DeductAmount
	resp.Refund.Reason = refund.Reason
	resp.Refund.Status = refund.Status
	resp.Refund.WechatRefundID = refund.WechatRefundID
	resp.Refund.WechatStatus = refund.WechatStatus
	resp.Refund.RejectReason = refund.RejectReason
	resp.Refund.ExpectArriveDate = refund.ExpectArriveDate
	resp.Refund.CreatedAt = refund.CreatedAt
	resp.Refund.UpdatedAt = refund.UpdatedAt
	resp.Order.OrderNo = order.OrderNo
	resp.Order.Status = order.Status
	resp.Order.ActualPrice = order.ActualPrice
	resp.Order.Quantity = order.Quantity
	resp.Order.ActivityName = activity.Name
	resp.Order.TicketSpecName = ticketSpec.Name
	resp.Order.PayMethod = order.PayMethod
	resp.Order.PayTime = order.PayTime

	var user models.Users
	if err := s.DB.WithContext(ctx).Where("id = ?", order.UserID).First(&user).Error; err == nil {
		resp.Order.UserName = user.Nickname
		resp.Order.UserMobile = maskPhone(user.Mobile)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	viewers, err := s.orderViewers(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	resp.Viewers = orderViewerItems(viewers, false)
	if err := s.DB.WithContext(ctx).Where("refund_id = ?", refund.ID).Order("id ASC").Find(&resp.RefundLogs).Error; err != nil {
		return nil, err
	}

	var payRecords []models.PayRecord
	if err := s.DB.WithContext(ctx).Where("order_sn = ?", order.OrderNo).Order("id DESC").Find(&payRecords).Error; err != nil {
		return nil, err
	}
	resp.PayRecords = make([]types.OrganizerRefundPayRecord, 0, len(payRecords))
	for _, record := range payRecords {
		resp.PayRecords = append(resp.PayRecords, types.OrganizerRefundPayRecord{
			ID: record.ID, PayPlatform: record.PayPlatform, PayMethod: record.PayMethod,
			TransactionID: record.TransactionId, AmountTotal: record.AmountTotal, Currency: record.Currency,
			PayStatus: record.PayStatus, TradeState: record.RawTradeState, FinishedAt: record.FinishedAt, CreatedAt: record.CreatedAt,
		})
	}

	var records []struct {
		ID            int64
		VerifierID    int64
		VerifierName  string
		VerifierPhone string
		ActivityID    int64
		ActivityName  string
		VerifiedAt    time.Time
	}
	if err := s.DB.WithContext(ctx).Table("verification_records vr").
		Select("vr.id, vr.verifier_id, v.name AS verifier_name, v.phone AS verifier_phone, vr.activity_id, a.name AS activity_name, vr.verified_at").
		Joins("LEFT JOIN verifiers v ON v.id = vr.verifier_id").
		Joins("LEFT JOIN activities a ON a.id = vr.activity_id").
		Where("vr.order_id = ?", order.ID).Order("vr.id DESC").Find(&records).Error; err != nil {
		return nil, err
	}
	resp.VerificationRecords = make([]types.OrganizerRefundVerificationItem, 0, len(records))
	for _, record := range records {
		resp.VerificationRecords = append(resp.VerificationRecords, types.OrganizerRefundVerificationItem{
			ID: record.ID, VerifierID: record.VerifierID, VerifierName: record.VerifierName,
			VerifierPhone: maskPhone(record.VerifierPhone), ActivityID: record.ActivityID,
			ActivityName: record.ActivityName, VerifiedAt: record.VerifiedAt,
		})
	}
	if err := s.createOrganizerLog(s.DB.WithContext(ctx), org.ID, userID, "organizer.refund.view", "refund", "GET", "/api/v1/organizer/refunds/:refund_no", "refund_no="+refund.RefundNo); err != nil {
		return nil, err
	}
	return resp, nil
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

func (s *TicketingService) UpdateVerifierStatus(ctx context.Context, userID, verifierID int64, status int8) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	if status != models.VerifierStatusInactive && status != models.VerifierStatusActive {
		return errors.New("核销员状态无效")
	}
	result := s.DB.WithContext(ctx).Model(&models.Verifier{}).Where("id = ? AND organizer_id = ?", verifierID, org.ID).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return s.createOrganizerLog(s.DB.WithContext(ctx), org.ID, userID, "update_verifier_status", "verifier", "", "", fmt.Sprintf("verifier_id=%d,status=%d", verifierID, status))
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
	if hasActiveRefund, err := s.hasActiveRefund(ctx, order.ID); err != nil {
		return &types.ScanOrderResponse{Success: false, ErrorCode: "INVALID_QR"}, nil
	} else if hasActiveRefund {
		return &types.ScanOrderResponse{Success: false, ErrorCode: "REFUND_PENDING"}, nil
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
		var refundCount int64
		if err := tx.Model(&models.Refund{}).Where("order_id = ? AND status IN ?", order.ID, []int8{models.RefundStatusAuditing, models.RefundStatusRunning}).Count(&refundCount).Error; err != nil {
			return err
		}
		if refundCount > 0 {
			return errors.New("订单售后处理中，暂不可核销")
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
	if err == nil {
		if org.Status != models.OrganizerStatusApproved {
			return &org, errors.New("商家尚未审核通过")
		}
		if org.Enabled != 1 {
			return &org, errors.New("商家账号已停用")
		}
	}
	return &org, err
}

func (s *TicketingService) ensureActivitiesBelongToOrganizer(ctx context.Context, organizerID int64, activityIDs []int64) error {
	if len(activityIDs) == 0 {
		return nil
	}
	var count int64
	if err := s.DB.WithContext(ctx).Model(&models.Activity{}).Where("id IN ? AND organizer_id = ?", activityIDs, organizerID).Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(activityIDs)) {
		return errors.New("合集包含无权限活动")
	}
	return nil
}

func (s *TicketingService) ensureDefaultLevelRules(ctx context.Context) error {
	return ensureDefaultOrganizerLevelRules(s.DB.WithContext(ctx))
}

func ensureDefaultOrganizerLevelRules(db *gorm.DB) error {
	for _, rule := range defaultOrganizerLevelRules() {
		if err := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "level"}}, DoNothing: true}).Create(&rule).Error; err != nil {
			return err
		}
	}
	return nil
}

func defaultOrganizerLevelRules() []models.OrganizerLevelRule {
	return []models.OrganizerLevelRule{
		{Level: 1, Name: "LV1", FeeRate: 0.05, RequiredActivityCount: 0, Description: "默认等级", Benefits: "基础商家权益", Status: 1},
		{Level: 2, Name: "LV2", FeeRate: 0.03, RequiredActivityCount: 5, Description: "办理5场活动升级", Benefits: "手续费降至3%", Status: 1},
		{Level: 3, Name: "LV3", FeeRate: 0, RequiredActivityCount: 10, Description: "办理10场活动升级", Benefits: "手续费0%", Status: 1},
	}
}

func (s *TicketingService) ensureOrganizerPostRelations(ctx context.Context, organizerID int64, activityID int, storeID int64) error {
	if activityID > 0 {
		var count int64
		if err := s.DB.WithContext(ctx).Model(&models.Activity{}).Where("id = ? AND organizer_id = ?", activityID, organizerID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errors.New("关联活动不存在或无权限")
		}
	}
	if storeID > 0 {
		var count int64
		if err := s.DB.WithContext(ctx).Model(&models.OrganizerStore{}).Where("id = ? AND organizer_id = ?", storeID, organizerID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errors.New("关联门店不存在或无权限")
		}
	}
	return nil
}

func buildOrganizerPostItem(note models.Note, likeCount, collCount, shareCount, commentCount int64, activityName, storeName string) types.OrganizerPostItem {
	item := types.OrganizerPostItem{
		ID:           note.ID,
		UserID:       note.UserID,
		Title:        note.Title,
		Content:      note.Content,
		Type:         note.Type,
		Status:       note.Status,
		VisibleConf:  note.VisibleConf,
		ActivityID:   note.ActivityID,
		ActivityName: activityName,
		StoreID:      note.StoreID,
		StoreName:    storeName,
		LikeCount:    likeCount,
		CollCount:    collCount,
		ShareCount:   shareCount,
		CommentCount: commentCount,
		CreatedAt:    note.CreatedAt,
		UpdatedAt:    note.UpdatedAt,
		MediaData:    []types.NoteMedia{},
	}
	if note.MediaData != "" {
		_ = json.Unmarshal([]byte(note.MediaData), &item.MediaData)
	}
	if note.Location != "" {
		_ = json.Unmarshal([]byte(note.Location), &item.Location)
	}
	return item
}

func sortFinanceFlows(flows []types.OrganizerFinanceFlowItem) {
	sort.Slice(flows, func(i, j int) bool {
		return flows[i].CreatedAt.After(flows[j].CreatedAt)
	})
}

func (s *TicketingService) createOrganizerLog(tx *gorm.DB, organizerID, operatorID int64, action, resource, method, path, remark string) error {
	return tx.Create(&models.OrganizerOperationLog{
		OrganizerID: organizerID,
		OperatorID:  operatorID,
		Action:      action,
		Resource:    resource,
		Method:      method,
		Path:        path,
		Remark:      remark,
	}).Error
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
	org = &models.Organizer{UserID: userID, Type: models.OrganizerTypeMerchant, Name: "未命名主办方", Status: models.OrganizerStatusPending, Level: "LV1"}
	if err := s.DB.WithContext(ctx).Create(org).Error; err != nil {
		return nil, err
	}
	return org, nil
}

func (s *TicketingService) listActivities(query *gorm.DB, page, size int, userID int64) (*types.PageResponse[types.ActivityListItem], error) {
	page, size = normalizePage(page, size)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var acts []models.Activity
	if err := query.Order("created_at desc").Offset((page - 1) * size).Limit(size).Find(&acts).Error; err != nil {
		return nil, err
	}
	activityIDs := make([]int64, 0, len(acts))
	venueIDs := make([]int64, 0, len(acts))
	for _, activity := range acts {
		if defaultActivityType(activity.Type) == models.ActivityTypeVenue {
			venueIDs = append(venueIDs, activity.OrganizerID)
		} else {
			activityIDs = append(activityIDs, activity.ID)
		}
	}
	activityTags, err := dao.LoadContentTags(query.Statement.Context, s.DB, models.ContentTagTargetActivity, activityIDs, false)
	if err != nil {
		return nil, err
	}
	venueTags, err := dao.LoadContentTags(query.Statement.Context, s.DB, models.ContentTagTargetVenue, venueIDs, false)
	if err != nil {
		return nil, err
	}
	activityFollowCounts, activityFollowed, err := dao.LoadContentFollowStats(query.Statement.Context, s.DB, models.ContentFollowTargetActivity, activityIDs, userID)
	if err != nil {
		return nil, err
	}
	venueFollowCounts, venueFollowed, err := dao.LoadContentFollowStats(query.Statement.Context, s.DB, models.ContentFollowTargetVenue, venueIDs, userID)
	if err != nil {
		return nil, err
	}
	list := make([]types.ActivityListItem, 0, len(acts))
	for _, a := range acts {
		activityType := defaultActivityType(a.Type)
		tags := activityTags[a.ID]
		followCount := activityFollowCounts[a.ID]
		isFollow := activityFollowed[a.ID]
		if activityType == models.ActivityTypeVenue {
			tags = venueTags[a.OrganizerID]
			followCount = venueFollowCounts[a.OrganizerID]
			isFollow = venueFollowed[a.OrganizerID]
		}
		list = append(list, types.ActivityListItem{ID: a.ID, Type: activityType, Name: a.Name, PosterList: a.PosterList, StartTime: a.StartTime, EndTime: a.EndTime, Status: a.Status, TagIDs: types.ContentTagIDs(tags), Tags: types.BuildContentTagItems(tags), IsFollow: isFollow, FollowCount: followCount, FollowTargetType: contentFollowTargetForActivity(a), FollowTargetID: contentFollowIDForActivity(a)})
	}
	return &types.PageResponse[types.ActivityListItem]{List: list, Total: total}, nil
}

func (s *TicketingService) buildOrderDetail(ctx context.Context, order models.TicketOrder) (*types.TicketOrderDetailResponse, error) {
	if order.Status == models.TicketOrderStatusPending && order.ActualPrice == 0 {
		if err := s.markZeroPayOrderUsable(ctx, order.ID, order.PointsAmount); err != nil {
			return nil, err
		}
		order.Status = models.TicketOrderStatusUsable
		order.PayMethod = zeroPayMethod(order.PointsAmount)
		nowPaid := time.Now()
		order.PayTime = &nowPaid
	}
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
		if refund.Status == models.RefundStatusSuccess && order.Status != models.TicketOrderStatusRefundSuccess {
			order.Status = models.TicketOrderStatusRefundSuccess
			resp.Status = order.Status
			_ = s.DB.WithContext(ctx).Model(&models.TicketOrder{}).
				Where("id = ? AND status <> ?", order.ID, models.TicketOrderStatusRefundSuccess).
				Update("status", models.TicketOrderStatusRefundSuccess).Error
		}
		statusText := refundStatusText(refund.Status)
		resp.RefundNo = refund.RefundNo
		resp.RefundStatus = refundStatusCode(refund.Status)
		resp.RefundStatusText = statusText
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

func (s *TicketingService) markZeroPayOrderUsable(ctx context.Context, orderID int64, pointsAmount int64) error {
	payMethod := zeroPayMethod(pointsAmount)
	nowPaid := time.Now()
	return s.DB.WithContext(ctx).Model(&models.TicketOrder{}).
		Where("id = ? AND status = ? AND actual_price = 0", orderID, models.TicketOrderStatusPending).
		Updates(map[string]any{
			"status":     models.TicketOrderStatusUsable,
			"pay_method": payMethod,
			"pay_time":   nowPaid,
		}).Error
}

func (s *TicketingService) expireUserPendingTicketOrders(ctx context.Context, userID int64) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var orders []models.TicketOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND status = ? AND actual_price > 0 AND expire_time <= ?", userID, models.TicketOrderStatusPending, time.Now()).
			Find(&orders).Error; err != nil {
			return err
		}
		for i := range orders {
			if err := cancelPendingTicketOrderTx(tx, &orders[i], "超时未支付自动取消"); err != nil {
				return err
			}
		}
		return nil
	})
}

func cancelPendingTicketOrderTx(tx *gorm.DB, order *models.TicketOrder, reason string) error {
	if order.Status != models.TicketOrderStatusPending {
		return errors.New("当前订单不可取消")
	}
	if err := tx.Model(&models.TicketSpec{}).Where("id = ?", order.TicketSpecID).
		UpdateColumn("sold_count", gorm.Expr("GREATEST(sold_count - ?, 0)", order.Quantity)).Error; err != nil {
		return err
	}
	if err := returnOrderDeductedPointsTx(tx, *order); err != nil {
		return err
	}
	if err := tx.Model(&models.PayRecord{}).Where("order_sn = ? AND pay_status = 0", order.OrderNo).Update("pay_status", 4).Error; err != nil {
		return err
	}
	return tx.Model(order).Updates(map[string]any{
		"status":        models.TicketOrderStatusCancelled,
		"cancel_reason": reason,
	}).Error
}

func returnOrderDeductedPointsTx(tx *gorm.DB, order models.TicketOrder) error {
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
		Remark:     "待支付订单取消返还积分抵扣",
		Status:     1,
	}).Error
}

func (s *TicketingService) hasActiveRefund(ctx context.Context, orderID int64) (bool, error) {
	var count int64
	if err := s.DB.WithContext(ctx).Model(&models.Refund{}).
		Where("order_id = ? AND status IN ?", orderID, []int8{models.RefundStatusAuditing, models.RefundStatusRunning}).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func zeroPayMethod(pointsAmount int64) string {
	if pointsAmount > 0 {
		return "points"
	}
	return "free"
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

func refundStatusCode(status int8) string {
	switch status {
	case models.RefundStatusAuditing:
		return "pending_review"
	case models.RefundStatusRunning:
		return "refunding"
	case models.RefundStatusSuccess:
		return "refunded"
	case models.RefundStatusRejected:
		return "rejected"
	case models.RefundStatusCancelled:
		return "cancelled"
	default:
		return ""
	}
}

func refundStatusValue(code string) (int8, bool) {
	switch code {
	case "pending_review":
		return models.RefundStatusAuditing, true
	case "refunding":
		return models.RefundStatusRunning, true
	case "refunded":
		return models.RefundStatusSuccess, true
	case "rejected":
		return models.RefundStatusRejected, true
	case "cancelled":
		return models.RefundStatusCancelled, true
	default:
		return 0, false
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

func (s *TicketingService) latestRefundByOrderID(ctx context.Context, orderIDs []int64) (map[int64]models.Refund, error) {
	result := make(map[int64]models.Refund)
	if len(orderIDs) == 0 {
		return result, nil
	}
	var refunds []models.Refund
	if err := s.DB.WithContext(ctx).
		Where("order_id IN ?", orderIDs).
		Order("id desc").
		Find(&refunds).Error; err != nil {
		return nil, err
	}
	for _, refund := range refunds {
		if _, ok := result[refund.OrderID]; !ok {
			result[refund.OrderID] = refund
		}
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

func (s *TicketingService) fillOrganizerLevelInfo(ctx context.Context, organizerID int64, resp *types.OrganizerInfoResponse) error {
	var completed int64
	if err := s.DB.WithContext(ctx).Model(&models.Activity{}).
		Where("organizer_id = ? AND status = ? AND end_time < ?", organizerID, models.ActivityStatusOnline, time.Now()).
		Count(&completed).Error; err != nil {
		return err
	}
	level, feeRate, next, err := organizerLevelByCompletedCount(s.DB.WithContext(ctx), completed)
	if err != nil {
		return err
	}
	resp.LevelValue = level
	resp.Level = fmt.Sprintf("LV%d", level)
	resp.FeeRate = feeRate
	resp.ServiceFeeRate = feeRate
	resp.CompletedActivityCount = completed
	resp.NextLevelRequiredCount = next
	return s.DB.WithContext(ctx).Model(&models.Organizer{}).Where("id = ?", organizerID).Updates(map[string]any{
		"level":            resp.Level,
		"service_fee_rate": feeRate,
		"updated_at":       time.Now(),
	}).Error
}

// organizerLevelByCompletedCount resolves the active, platform-configured level rule.
func organizerLevelByCompletedCount(db *gorm.DB, completed int64) (level int, feeRate float64, nextRequired int64, err error) {
	if err := ensureDefaultOrganizerLevelRules(db); err != nil {
		return 0, 0, 0, err
	}
	var current models.OrganizerLevelRule
	if err := db.Where("status = ? AND required_activity_count <= ?", 1, completed).
		Order("required_activity_count DESC, level DESC").First(&current).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, 0, err
		}
		// Keep a safe default even if every configurable rule has been disabled.
		return 1, 0.05, 5, nil
	}
	var next models.OrganizerLevelRule
	nextRequired = 0
	if err := db.Where("status = ? AND required_activity_count > ?", 1, completed).
		Order("required_activity_count ASC, level ASC").First(&next).Error; err == nil {
		nextRequired = next.RequiredActivityCount
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, 0, 0, err
	}
	return current.Level, current.FeeRate, nextRequired, nil
}

func buildOrganizerBankAuditInfo(audit models.OrganizerBankAccountAudit) types.OrganizerBankAuditInfo {
	return types.OrganizerBankAuditInfo{
		ID:              audit.ID,
		BankAccountName: audit.BankAccountName,
		BankAccountNo:   audit.BankAccountNo,
		BankName:        audit.BankName,
		ContactName:     audit.BankContactName,
		ContactPhone:    audit.BankContactPhone,
		Status:          audit.Status,
		RejectReason:    audit.RejectReason,
		ReviewedAt:      audit.ReviewedAt,
		CreatedAt:       audit.CreatedAt,
		UpdatedAt:       audit.UpdatedAt,
	}
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
	if req.Type != nil {
		activityType, err := normalizeActivityType(*req.Type)
		if err != nil {
			return nil, err
		}
		updates["type"] = activityType
	}
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

func normalizeActivityType(raw string) (string, error) {
	switch strings.TrimSpace(raw) {
	case "", models.ActivityTypeParty:
		return models.ActivityTypeParty, nil
	case models.ActivityTypeVenue:
		return models.ActivityTypeVenue, nil
	default:
		return "", errors.New("活动类型无效，仅支持 party 或 venue")
	}
}

func defaultActivityType(raw string) string {
	activityType, err := normalizeActivityType(raw)
	if err != nil {
		return models.ActivityTypeParty
	}
	return activityType
}

func contentFollowTargetForActivity(activity models.Activity) string {
	if defaultActivityType(activity.Type) == models.ActivityTypeVenue {
		return models.ContentFollowTargetVenue
	}
	return models.ContentFollowTargetActivity
}

func contentFollowIDForActivity(activity models.Activity) int64 {
	if defaultActivityType(activity.Type) == models.ActivityTypeVenue {
		return activity.OrganizerID
	}
	return activity.ID
}

func (s *TicketingService) loadActivityFollowStats(ctx context.Context, userID int64, activities []models.Activity) (map[int64]int64, map[int64]bool, error) {
	counts := make(map[int64]int64, len(activities))
	followed := make(map[int64]bool, len(activities))
	activityIDs := make([]int64, 0, len(activities))
	venueIDs := make([]int64, 0, len(activities))
	for _, activity := range activities {
		if contentFollowTargetForActivity(activity) == models.ContentFollowTargetVenue {
			venueIDs = append(venueIDs, activity.OrganizerID)
		} else {
			activityIDs = append(activityIDs, activity.ID)
		}
	}
	activityCounts, activityFollowed, err := dao.LoadContentFollowStats(ctx, s.DB, models.ContentFollowTargetActivity, activityIDs, userID)
	if err != nil {
		return nil, nil, err
	}
	venueCounts, venueFollowed, err := dao.LoadContentFollowStats(ctx, s.DB, models.ContentFollowTargetVenue, venueIDs, userID)
	if err != nil {
		return nil, nil, err
	}
	for _, activity := range activities {
		if contentFollowTargetForActivity(activity) == models.ContentFollowTargetVenue {
			counts[activity.ID] = venueCounts[activity.OrganizerID]
			followed[activity.ID] = venueFollowed[activity.OrganizerID]
			continue
		}
		counts[activity.ID] = activityCounts[activity.ID]
		followed[activity.ID] = activityFollowed[activity.ID]
	}
	return counts, followed, nil
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
		Description:   item.Description,
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
