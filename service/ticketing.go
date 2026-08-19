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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"math"
	"net/url"
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
	ListContentTags(ctx context.Context) ([]types.ContentTagItem, error)
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
	ListOrganizerFollowers(ctx context.Context, userID int64, page, size int, keyword string) (*types.PageResponse[types.OrganizerFollowerItem], error)
	ListVenues(ctx context.Context, userID int64, keyword string, tagIDs []int64, page, size int) (*types.PageResponse[types.VenueListItem], error)
	GetVenueDetail(ctx context.Context, userID, venueID int64) (*types.VenueDetailResponse, error)
	ListVenueNotes(ctx context.Context, userID, venueID int64, cursor int64, pageSize int) (*types.VenueNotesResponse, error)
	FollowVenue(ctx context.Context, userID, venueID int64) error
	UnfollowVenue(ctx context.Context, userID, venueID int64) error
	SubscribeVenue(ctx context.Context, userID, venueID int64) error
	UnsubscribeVenue(ctx context.Context, userID, venueID int64) error
	ListSubscriptions(ctx context.Context, userID int64, subType string, page, size int) (*types.PageResponse[types.SubscriptionListItem], error)
	GetPublicOrganizerHome(ctx context.Context, userID, organizerID int64, activityPage, activitySize, venuePage, venueSize int) (*types.PublicOrganizerHomeResponse, error)
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
	RecordActivityView(ctx context.Context, userID, activityID int64, visitorID string) error
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
	ListOrganizerOrders(ctx context.Context, userID int64, activityID int64, status *int8, keyword, salesChannel, withdrawStatus, startDate, endDate string, page, size int) (*types.PageResponse[types.OrganizerOrderListItem], error)
	GetOrganizerOrderSummary(ctx context.Context, userID int64, startDate, endDate string) (*types.OrganizerOrderSummary, error)
	GetOrganizerOrderDetail(ctx context.Context, userID int64, orderNo string) (*types.OrganizerOrderDetailResponse, error)
	ListOrganizerRefunds(ctx context.Context, userID int64, status *int8, page, size int) (*types.PageResponse[types.OrganizerRefundListItem], error)
	GetOrganizerRefundDetail(ctx context.Context, userID int64, refundNo string) (*types.OrganizerRefundDetailResponse, error)
	GetRefundDetail(ctx context.Context, userID int64, refundNo string) (*types.RefundDetailResponse, error)
	RejectRefund(ctx context.Context, userID int64, refundNo string, req types.RejectRefundRequest) error
	CancelRefund(ctx context.Context, userID int64, refundNo string) error
	ListVerifiers(ctx context.Context, userID int64, page, size int) (*types.PageResponse[types.OrganizerVerifierItem], error)
	GetVerifierOrderDetail(ctx context.Context, userID int64, orderNo string) (*types.OrganizerOrderDetailResponse, error)
	AddVerifier(ctx context.Context, userID int64, req types.VerifierRequest) error
	UpdateVerifierStatus(ctx context.Context, userID, verifierID int64, status int8) error
	DeleteVerifier(ctx context.Context, userID, verifierID int64) error
	GetVerifierActivationQR(ctx context.Context, userID, verifierID int64) (map[string]string, error)
	GetVerifierActivationInfo(ctx context.Context, verifierID int64) (*types.VerifierActivationInfoResponse, error)
	ActivateVerifier(ctx context.Context, userID int64, req types.ActivateVerifierRequest) (*types.ActivateVerifierResponse, error)
	ScanOrder(ctx context.Context, req types.ScanOrderRequest) (*types.ScanOrderResponse, error)
	ConfirmVerify(ctx context.Context, verifierID int64, req types.ConfirmVerifyRequest) error
	ListVerified(ctx context.Context, verifierID int64, page, size int) (*types.PageResponse[types.VerifiedListItem], error)
	ListVerifiedByUser(ctx context.Context, userID int64, page, size int) (*types.PageResponse[types.VerifiedListItem], error)
	ListOrganizerVerificationRecords(ctx context.Context, userID int64, filter types.VerificationRecordFilter) (*types.PageResponse[types.VerifiedListItem], error)
	ResolveBoundVerifierID(ctx context.Context, userID int64) (int64, error)
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

// ListContentTags returns the currently enabled marketing tags for the
// activity/venue publishing flow. Tag configuration remains admin-owned.
func (s *TicketingService) ListContentTags(ctx context.Context) ([]types.ContentTagItem, error) {
	tags, err := dao.ListActiveCouponTags(ctx, s.DB)
	if err != nil {
		return nil, err
	}
	return types.BuildContentTagItems(tags), nil
}

func (s *TicketingService) ApplyOrganizer(ctx context.Context, userID int64, req types.OrganizerApplyRequest) (*types.OrganizerApplyResponse, error) {
	organizerType, err := normalizeOrganizerType(req.Type)
	if err != nil {
		return nil, err
	}
	markerIcon, err := normalizeMarkerIcon(req.MarkerIcon)
	if err != nil {
		return nil, err
	}
	var venueProfile *types.OrganizerVenueProfileRevision
	if organizerType == models.OrganizerTypeVenue {
		if req.VenueProfile == nil {
			return nil, errors.New("场地入驻必须填写固定地址、经纬度和营业时间")
		}
		if err := validateVenueProfileInput(*req.VenueProfile); err != nil {
			return nil, err
		}
		venueProfile = &types.OrganizerVenueProfileRevision{
			Name: req.Name, Logo: req.Logo, MarkerIcon: markerIcon,
			Province: req.Province, City: req.City, District: req.District,
			OrganizerVenueProfileInput: *req.VenueProfile,
		}
	}
	var org models.Organizer
	err = s.DB.WithContext(ctx).Where("user_id = ?", userID).First(&org).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	data := models.Organizer{
		UserID:     userID,
		Type:       organizerType,
		Name:       req.Name,
		Logo:       req.Logo,
		MarkerIcon: markerIcon,
		Status:     models.OrganizerStatusAuditing,
		Level:      "LV1",
		Province:   req.Province,
		City:       req.City,
		District:   req.District,
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&data).Error; err != nil {
				return err
			}
			return upsertOrganizerVenueProfile(tx, data.ID, venueProfile)
		}); err != nil {
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
	updates := map[string]any{
		"name":          req.Name,
		"logo":          req.Logo,
		"type":          organizerType,
		"status":        models.OrganizerStatusAuditing,
		"reject_reason": "",
		"level":         "LV1",
		"province":      req.Province,
		"city":          req.City,
		"district":      req.District,
	}
	if markerIcon != "" {
		updates["marker_icon"] = markerIcon
	}
	if err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&org).Updates(updates).Error; err != nil {
			return err
		}
		return upsertOrganizerVenueProfile(tx, org.ID, venueProfile)
	}); err != nil {
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
		OrganizerID:               org.ID,
		Type:                      org.Type,
		MarkerIcon:                org.MarkerIcon,
		Status:                    org.Status,
		Enabled:                   org.Enabled,
		RejectReason:              org.RejectReason,
		HasPendingProfileRevision: org.PendingProfileStatus == models.OrganizerStatusAuditing || org.PendingProfileStatus == models.OrganizerStatusRejected,
		PendingProfileReason:      org.PendingProfileReason,
		SubmittedAt:               &org.CreatedAt,
	}
	if (org.Status == models.OrganizerStatusApproved || org.Status == models.OrganizerStatusRejected) && org.PendingProfileStatus != models.OrganizerStatusAuditing {
		resp.ReviewedAt = &org.UpdatedAt
	}
	return resp, nil
}

func (s *TicketingService) UpdateOrganizerBasic(ctx context.Context, userID int64, req types.OrganizerBasicRequest) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	if org.Type == models.OrganizerTypeVenue && org.Status == models.OrganizerStatusApproved {
		return errors.New("已通过场地请通过主办方资料接口提交完整资料，修改后需重新审核")
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
		if err := s.allocateOrganizerWithdrawOrders(ctx, tx, org.ID, withdraw); err != nil {
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

// allocateOrganizerWithdrawOrders reserves order revenue in FIFO payment order
// for a newly created withdrawal. Existing withdrawals from before this table
// was introduced are treated as an opaque FIFO reserve, so they cannot be
// accidentally allocated again.
func (s *TicketingService) allocateOrganizerWithdrawOrders(ctx context.Context, tx *gorm.DB, organizerID int64, withdraw models.OrganizerWithdraw) error {
	type orderRow struct {
		ID          int64
		OrderNo     string
		ActivityID  int64
		ActualPrice int64
	}
	var orders []orderRow
	if err := tx.WithContext(ctx).Table("ticket_orders o").
		Select("o.id, o.order_no, o.activity_id, o.actual_price").
		Joins("JOIN activities a ON a.id = o.activity_id").
		Where("a.organizer_id = ? AND o.status IN ? AND o.actual_price > 0", organizerID, []int8{
			models.TicketOrderStatusUsable,
			models.TicketOrderStatusUsed,
			models.TicketOrderStatusRefundReject,
		}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Order("o.pay_time ASC, o.id ASC").Find(&orders).Error; err != nil {
		return err
	}

	var allocationRows []struct {
		OrderID int64
		Amount  int64
	}
	if err := tx.WithContext(ctx).Model(&models.OrganizerWithdrawAllocation{}).
		Select("order_id, COALESCE(SUM(amount), 0) AS amount").
		Where("organizer_id = ? AND status IN ?", organizerID, []int8{
			models.OrganizerWithdrawAllocationStatusPending,
			models.OrganizerWithdrawAllocationStatusSettled,
		}).
		Group("order_id").Scan(&allocationRows).Error; err != nil {
		return err
	}
	allocated := make(map[int64]int64, len(allocationRows))
	for _, row := range allocationRows {
		allocated[row.OrderID] = row.Amount
	}

	var legacyReserved int64
	if err := tx.WithContext(ctx).Table("organizer_withdraws w").
		Select("COALESCE(SUM(w.amount), 0)").
		Where("w.organizer_id = ? AND w.status IN ? AND w.id <> ?", organizerID, []int8{0, 1}, withdraw.ID).
		Where("NOT EXISTS (SELECT 1 FROM organizer_withdraw_allocations owa WHERE owa.withdraw_id = w.id)").
		Scan(&legacyReserved).Error; err != nil {
		return err
	}

	remaining := withdraw.Amount
	allocations := make([]models.OrganizerWithdrawAllocation, 0)
	for _, order := range orders {
		available := order.ActualPrice - allocated[order.ID]
		if available <= 0 {
			continue
		}
		if legacyReserved > 0 {
			consumed := minInt64(available, legacyReserved)
			available -= consumed
			legacyReserved -= consumed
		}
		if available <= 0 {
			continue
		}
		amount := minInt64(available, remaining)
		allocations = append(allocations, models.OrganizerWithdrawAllocation{
			WithdrawID:  withdraw.ID,
			OrganizerID: organizerID,
			OrderID:     order.ID,
			OrderNo:     order.OrderNo,
			ActivityID:  order.ActivityID,
			Amount:      amount,
			Status:      models.OrganizerWithdrawAllocationStatusPending,
		})
		remaining -= amount
		if remaining == 0 {
			break
		}
	}
	if remaining > 0 {
		return errors.New("可分配订单资金不足，请刷新可提现金额后重试")
	}
	return tx.WithContext(ctx).Create(&allocations).Error
}

func minInt64(left, right int64) int64 {
	if left < right {
		return left
	}
	return right
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
				AuditType:  act.AuditType,
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

// ListOrganizerFollowers returns followers of the organizer's public entity.
// A fixed venue is followed as target_type=venue; a regular organizer is
// followed as target_type=organizer. Personal user_follow relations are not
// part of the merchant audience.
func (s *TicketingService) ListOrganizerFollowers(ctx context.Context, userID int64, page, size int, keyword string) (*types.PageResponse[types.OrganizerFollowerItem], error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	page, size = normalizePage(page, size)
	targetType, err := resolveOrganizerFollowTarget(ctx, s.DB, *org)
	if err != nil {
		return nil, err
	}
	sourceSQL := `
		SELECT cf.user_id, cf.target_type, cf.created_at AS followed_at
		FROM content_follows cf
		WHERE cf.target_id = ? AND cf.target_type = ?`
	sourceArgs := []any{org.ID, targetType}

	base := s.DB.WithContext(ctx).Table("("+sourceSQL+") AS f", sourceArgs...).
		Joins("JOIN users u ON u.id = f.user_id").
		Where("f.user_id <> ?", org.UserID)
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		base = base.Where("u.nickname LIKE ? OR u.mobile LIKE ?", like, like)
	}

	var total int64
	countQuery := base.Session(&gorm.Session{}).Select("f.user_id").Group("f.user_id")
	if err := s.DB.WithContext(ctx).Table("(?) AS organizer_followers", countQuery).Count(&total).Error; err != nil {
		return nil, err
	}

	type followerRow struct {
		UserID      int64
		Nickname    string
		Avatar      string
		Signature   string
		Mobile      string
		UserStatus  int8
		TargetTypes string
		FollowedAt  time.Time
	}
	var rows []followerRow
	if err := base.Select(`f.user_id, COALESCE(u.nickname, '') AS nickname, COALESCE(u.avatar, '') AS avatar,
		COALESCE(u.motto, '') AS signature, COALESCE(u.mobile, '') AS mobile, u.status AS user_status,
		GROUP_CONCAT(DISTINCT f.target_type ORDER BY f.target_type SEPARATOR ',') AS target_types,
		MAX(f.followed_at) AS followed_at`).
		Group("f.user_id, u.nickname, u.avatar, u.motto, u.mobile, u.status").
		Order("followed_at DESC").Offset((page - 1) * size).Limit(size).Scan(&rows).Error; err != nil {
		return nil, err
	}

	list := make([]types.OrganizerFollowerItem, 0, len(rows))
	for _, row := range rows {
		item := types.OrganizerFollowerItem{
			UserID: row.UserID, Nickname: row.Nickname, Avatar: row.Avatar, Signature: row.Signature,
			Mobile: maskOrganizerFollowerMobile(row.Mobile), UserStatus: row.UserStatus, FollowedAt: row.FollowedAt,
		}
		if row.TargetTypes != "" {
			item.TargetTypes = strings.Split(row.TargetTypes, ",")
		} else {
			item.TargetTypes = []string{}
		}
		list = append(list, item)
	}
	return &types.PageResponse[types.OrganizerFollowerItem]{List: list, Total: total}, nil
}

func maskOrganizerFollowerMobile(mobile string) string {
	mobile = strings.TrimSpace(mobile)
	if len(mobile) != 11 {
		return ""
	}
	return mobile[:3] + "****" + mobile[7:]
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

// GetPublicOrganizerHome returns the public storefront of an approved organizer.
// Activities and venues keep separate pagination because the client displays
// them as independent sections or tabs.
func (s *TicketingService) GetPublicOrganizerHome(ctx context.Context, userID, organizerID int64, activityPage, activitySize, venuePage, venueSize int) (*types.PublicOrganizerHomeResponse, error) {
	activityPage, activitySize = normalizePage(activityPage, activitySize)
	venuePage, venueSize = normalizePage(venuePage, venueSize)

	var row struct {
		ID            int64
		UserID        int64
		Type          string
		Name          string
		Logo          string
		OwnerNickname string
		OwnerAvatar   string
		CoverImage    string
		Gallery       string
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
	}
	if err := s.DB.WithContext(ctx).Table("organizers o").
		Select(`o.id, o.user_id, o.type, o.name, o.logo, o.province, o.city, o.district,
			COALESCE(u.nickname, '') AS owner_nickname,
			COALESCE(u.avatar, '') AS owner_avatar,
			COALESCE(p.cover_image, '') AS cover_image,
			COALESCE(p.gallery, '') AS gallery,
			COALESCE(p.description, '') AS description,
			COALESCE(p.business_hours, '') AS business_hours,
			COALESCE(p.service_phone, '') AS service_phone,
			COALESCE(p.address, '') AS address,
			COALESCE(p.latitude, 0) AS latitude,
			COALESCE(p.longitude, 0) AS longitude,
			COALESCE(p.average_spend, 0) AS average_spend`).
		Joins("LEFT JOIN organizer_profiles p ON p.organizer_id = o.id").
		Joins("LEFT JOIN users u ON u.id = o.user_id").
		Where("o.id = ? AND o.status = ? AND o.enabled = 1", organizerID, models.OrganizerStatusApproved).
		First(&row).Error; err != nil {
		return nil, err
	}

	followTargetType, err := resolveOrganizerFollowTarget(ctx, s.DB, models.Organizer{ID: row.ID, Type: row.Type})
	if err != nil {
		return nil, err
	}
	followCounts, followed, err := dao.LoadContentFollowStatsExcludingOwners(
		ctx, s.DB, followTargetType, []int64{row.ID}, userID, map[int64]int64{row.ID: row.UserID},
	)
	if err != nil {
		return nil, err
	}
	resp := &types.PublicOrganizerHomeResponse{
		ID:               row.ID,
		UserID:           row.UserID,
		Type:             row.Type,
		Name:             row.Name,
		Logo:             row.Logo,
		OwnerNickname:    row.OwnerNickname,
		OwnerAvatar:      row.OwnerAvatar,
		CoverImage:       row.CoverImage,
		Gallery:          []string{},
		Description:      row.Description,
		BusinessHours:    row.BusinessHours,
		ServicePhone:     row.ServicePhone,
		Province:         row.Province,
		City:             row.City,
		District:         row.District,
		Address:          row.Address,
		Latitude:         row.Latitude,
		Longitude:        row.Longitude,
		AverageSpend:     row.AverageSpend,
		FollowCount:      followCounts[row.ID],
		IsFollow:         followed[row.ID],
		FollowTargetType: followTargetType,
		FollowTargetID:   row.ID,
		Activities:       types.PageResponse[types.ActivityListItem]{List: []types.ActivityListItem{}},
		Venues:           types.PageResponse[types.VenueListItem]{List: []types.VenueListItem{}},
	}
	if row.Gallery != "" {
		_ = json.Unmarshal([]byte(row.Gallery), &resp.Gallery)
	}

	activities, err := s.listActivities(
		s.DB.WithContext(ctx).Model(&models.Activity{}).
			Where("organizer_id = ? AND status = ? AND is_hidden = 0 AND type <> ?", row.ID, models.ActivityStatusOnline, models.ActivityTypeVenue),
		activityPage, activitySize, userID,
	)
	if err != nil {
		return nil, err
	}
	if err := s.fillActivitySubscriptionStatus(ctx, userID, activities.List); err != nil {
		return nil, err
	}
	resp.Activities = *activities
	resp.ActivityCount = activities.Total

	venues := []types.VenueListItem{}
	var venueTotal int64
	if row.Type == models.OrganizerTypeVenue {
		venueTotal = 1
		if venuePage == 1 {
			venues = append(venues, types.VenueListItem{ID: row.ID, UserID: row.UserID, Name: row.Name, Logo: row.Logo,
				CoverImage: row.CoverImage, Description: row.Description, BusinessHours: row.BusinessHours,
				ServicePhone: row.ServicePhone, Province: row.Province, City: row.City, District: row.District,
				Address: row.Address, Latitude: row.Latitude, Longitude: row.Longitude, AverageSpend: row.AverageSpend})
		}
	} else {
		// Compatibility fallback for historical venue activity rows. New venue
		// organizers are represented by their organizer profile above.
		venueQuery := s.DB.WithContext(ctx).Table("activities a").
			Select(`o.id, o.user_id, o.name, o.logo, a.id AS activity_id, a.name AS activity_name,
				COALESCE(NULLIF(a.poster_list, ''), p.cover_image, '') AS cover_image,
				COALESCE(NULLIF(a.description, ''), p.description, '') AS description,
				COALESCE(p.business_hours, '') AS business_hours, COALESCE(p.service_phone, '') AS service_phone,
				COALESCE(NULLIF(a.province, ''), o.province, '') AS province, COALESCE(NULLIF(a.city, ''), o.city, '') AS city,
				COALESCE(NULLIF(a.district, ''), o.district, '') AS district, COALESCE(NULLIF(a.address, ''), p.address, '') AS address,
				COALESCE(NULLIF(a.latitude, 0), p.latitude, 0) AS latitude, COALESCE(NULLIF(a.longitude, 0), p.longitude, 0) AS longitude,
				COALESCE(p.average_spend, 0) AS average_spend, a.created_at`).
			Joins("JOIN organizers o ON o.id = a.organizer_id").
			Joins("LEFT JOIN organizer_profiles p ON p.organizer_id = o.id").
			Where("a.organizer_id = ? AND a.type = ? AND a.status = ? AND a.is_hidden = 0", row.ID, models.ActivityTypeVenue, models.ActivityStatusOnline).
			Where("o.status = ? AND o.enabled = 1", models.OrganizerStatusApproved)
		if err := venueQuery.Count(&venueTotal).Error; err != nil {
			return nil, err
		}
		if err := venueQuery.Order("a.created_at DESC, a.id DESC").Offset((venuePage - 1) * venueSize).Limit(venueSize).Scan(&venues).Error; err != nil {
			return nil, err
		}
	}
	if err := s.fillVenueStats(ctx, userID, venues); err != nil {
		return nil, err
	}
	resp.Venues = types.PageResponse[types.VenueListItem]{List: venues, Total: venueTotal}
	resp.VenueCount = venueTotal
	return resp, nil
}

func (s *TicketingService) fillActivitySubscriptionStatus(ctx context.Context, userID int64, activities []types.ActivityListItem) error {
	if userID <= 0 || len(activities) == 0 {
		return nil
	}
	activityIDs := make([]int64, 0, len(activities))
	for _, activity := range activities {
		activityIDs = append(activityIDs, activity.ID)
	}
	var subscribedIDs []int64
	if err := s.DB.WithContext(ctx).Model(&models.ActivitySubscription{}).
		Where("user_id = ? AND activity_id IN ?", userID, activityIDs).
		Pluck("activity_id", &subscribedIDs).Error; err != nil {
		return err
	}
	subscribed := make(map[int64]bool, len(subscribedIDs))
	for _, activityID := range subscribedIDs {
		subscribed[activityID] = true
	}
	for i := range activities {
		activities[i].IsSubscribe = subscribed[activities[i].ID]
	}
	return nil
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
			Where("sub.user_id = ? AND a.status = ? AND a.is_hidden = 0 AND a.type <> ?", userID, models.ActivityStatusOnline, models.ActivityTypeVenue).
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
	return `o.status = ? AND o.enabled = 1 AND (
		o.type = 'venue' OR EXISTS (
			SELECT 1 FROM activities va
			WHERE va.organizer_id = o.id AND va.type = ? AND va.status = ? AND va.is_hidden = 0
		)
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
		Type:          org.Type,
		Name:          org.Name,
		Logo:          org.Logo,
		MarkerIcon:    org.MarkerIcon,
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
	if org.Type == models.OrganizerTypeVenue {
		resp.VenueProfile = &types.OrganizerVenueProfileInput{
			CoverImage: profile.CoverImage, Gallery: resp.Gallery, Description: profile.Description,
			BusinessHours: profile.BusinessHours, ContactName: profile.ContactName, ServicePhone: profile.ServicePhone,
			Address: profile.Address, Latitude: profile.Latitude, Longitude: profile.Longitude, AverageSpend: profile.AverageSpend,
		}
	}
	if org.PendingProfileStatus == models.OrganizerStatusAuditing || org.PendingProfileStatus == models.OrganizerStatusRejected {
		revision, err := decodeOrganizerVenueProfileRevision(*org)
		if err != nil {
			return nil, err
		}
		resp.HasPendingProfileRevision = revision != nil
		resp.PendingProfileReason = org.PendingProfileReason
		resp.PendingProfileRevision = revision
	}
	return resp, nil
}

func (s *TicketingService) UpdateOrganizerProfile(ctx context.Context, userID int64, req types.OrganizerProfileRequest) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	markerIcon, err := normalizeMarkerIcon(req.MarkerIcon)
	if err != nil {
		return err
	}
	profileInput := types.OrganizerVenueProfileInput{
		CoverImage: req.CoverImage, Gallery: req.Gallery, Description: req.Description,
		BusinessHours: req.BusinessHours, ContactName: req.ContactName, ServicePhone: req.ServicePhone,
		Address: req.Address, Latitude: req.Latitude, Longitude: req.Longitude, AverageSpend: req.AverageSpend,
	}
	if req.VenueProfile != nil {
		profileInput = *req.VenueProfile
	}
	if org.Type == models.OrganizerTypeVenue {
		if err := validateVenueProfileInput(profileInput); err != nil {
			return err
		}
	} else if req.Latitude != 0 || req.Longitude != 0 {
		if err := validateChinaCoordinate(req.Latitude, req.Longitude); err != nil {
			return err
		}
	}
	revision := &types.OrganizerVenueProfileRevision{
		Name: req.Name, Logo: req.Logo, MarkerIcon: markerIcon, Province: req.Province, City: req.City, District: req.District,
		OrganizerVenueProfileInput: profileInput,
	}
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked models.Organizer
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", org.ID).First(&locked).Error; err != nil {
			return err
		}
		if locked.Type == models.OrganizerTypeVenue && locked.Status == models.OrganizerStatusApproved {
			if locked.PendingProfileStatus == models.OrganizerStatusAuditing {
				return errors.New("场地资料正在审核中，请勿重复提交")
			}
			payload, err := json.Marshal(revision)
			if err != nil {
				return err
			}
			if err := tx.Model(&models.Organizer{}).Where("id = ?", locked.ID).Updates(map[string]any{
				"pending_profile_revision": string(payload),
				"pending_profile_status":   models.OrganizerStatusAuditing,
				"pending_profile_reason":   "",
				"updated_at":               time.Now(),
			}).Error; err != nil {
				return err
			}
			return s.createOrganizerLog(tx, locked.ID, userID, "submit_profile_revision", "organizer_profile", "", "", "")
		}

		gallery, _ := json.Marshal(profileInput.Gallery)
		orgUpdates := map[string]any{}
		if strings.TrimSpace(req.Name) != "" {
			orgUpdates["name"] = req.Name
		}
		if req.Logo != "" {
			orgUpdates["logo"] = req.Logo
		}
		if markerIcon != "" {
			orgUpdates["marker_icon"] = markerIcon
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
		profile := models.OrganizerProfile{OrganizerID: org.ID, CoverImage: profileInput.CoverImage, Gallery: string(gallery), Description: profileInput.Description, BusinessHours: profileInput.BusinessHours, ContactName: profileInput.ContactName, ServicePhone: profileInput.ServicePhone, Address: profileInput.Address, Latitude: profileInput.Latitude, Longitude: profileInput.Longitude, AverageSpend: profileInput.AverageSpend}
		updates := map[string]any{
			"cover_image": profileInput.CoverImage, "gallery": string(gallery), "description": profileInput.Description,
			"business_hours": profileInput.BusinessHours, "contact_name": profileInput.ContactName, "service_phone": profileInput.ServicePhone,
			"address": profileInput.Address, "latitude": profileInput.Latitude, "longitude": profileInput.Longitude, "average_spend": profileInput.AverageSpend,
			"updated_at": time.Now(),
		}
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "organizer_id"}}, DoUpdates: clause.Assignments(updates)}).Create(&profile).Error; err != nil {
			return err
		}
		return s.createOrganizerLog(tx, org.ID, userID, "update_profile", "organizer_profile", "", "", "")
	})
}

func validateVenueProfileInput(input types.OrganizerVenueProfileInput) error {
	if strings.TrimSpace(input.Address) == "" {
		return errors.New("场地地址不能为空")
	}
	if strings.TrimSpace(input.BusinessHours) == "" {
		return errors.New("场地营业时间不能为空")
	}
	return validateChinaCoordinate(input.Latitude, input.Longitude)
}

func upsertOrganizerVenueProfile(tx *gorm.DB, organizerID int64, revision *types.OrganizerVenueProfileRevision) error {
	if revision == nil {
		return nil
	}
	gallery, err := json.Marshal(revision.Gallery)
	if err != nil {
		return err
	}
	profile := models.OrganizerProfile{OrganizerID: organizerID, CoverImage: revision.CoverImage, Gallery: string(gallery), Description: revision.Description, BusinessHours: revision.BusinessHours, ContactName: revision.ContactName, ServicePhone: revision.ServicePhone, Address: revision.Address, Latitude: revision.Latitude, Longitude: revision.Longitude, AverageSpend: revision.AverageSpend}
	return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "organizer_id"}}, DoUpdates: clause.Assignments(map[string]any{
		"cover_image": profile.CoverImage, "gallery": profile.Gallery, "description": profile.Description,
		"business_hours": profile.BusinessHours, "contact_name": profile.ContactName, "service_phone": profile.ServicePhone,
		"address": profile.Address, "latitude": profile.Latitude, "longitude": profile.Longitude, "average_spend": profile.AverageSpend, "updated_at": time.Now(),
	})}).Create(&profile).Error
}

func decodeOrganizerVenueProfileRevision(org models.Organizer) (*types.OrganizerVenueProfileRevision, error) {
	if strings.TrimSpace(org.PendingProfileRevision) == "" {
		return nil, nil
	}
	var revision types.OrganizerVenueProfileRevision
	if err := json.Unmarshal([]byte(org.PendingProfileRevision), &revision); err != nil {
		return nil, fmt.Errorf("场地待审核修改数据损坏: %w", err)
	}
	return &revision, nil
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
	activityType := models.ActivityTypeParty
	if req.Type != nil {
		activityType, err = normalizeActivityType(*req.Type)
		if err != nil {
			return 0, err
		}
	}
	if req.ActivityID == 0 && activityType == models.ActivityTypeVenue {
		return 0, errors.New("新场地请在入驻申请中选择 venue 并填写固定资料，活动发布仅支持 party")
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
			AuditType:   models.ActivityAuditTypeInitial,
		}
		if err := s.DB.WithContext(ctx).Create(&act).Error; err != nil {
			return 0, err
		}
	}
	if req.Type == nil {
		activityType = defaultActivityType(act.Type)
	}
	wasVenue := defaultActivityType(act.Type) == models.ActivityTypeVenue
	if req.BusinessHours != nil && activityType != models.ActivityTypeVenue {
		return 0, errors.New("仅场地可填写每日经营时间")
	}
	if activityType == models.ActivityTypeVenue && (req.Step == 4 || req.TicketSpecs != nil) {
		return 0, errors.New("场地不支持票券配置")
	}
	if activityType == models.ActivityTypeVenue && !wasVenue {
		var ticketSpecCount int64
		if err := s.DB.WithContext(ctx).Model(&models.TicketSpec{}).Where("activity_id = ?", act.ID).Count(&ticketSpecCount).Error; err != nil {
			return 0, err
		}
		if ticketSpecCount > 0 {
			return 0, errors.New("已有票券的派对不能转换为场地")
		}
	}
	updates, err := activityUpdates(req, activityType)
	if err != nil {
		return 0, err
	}
	// Venue organizers inherit their reviewed venue address unless the activity
	// creator explicitly chose another event location in step 2. The organizer
	// profile itself remains fixed and is never changed by this request.
	if org.Type == models.OrganizerTypeVenue && activityType == models.ActivityTypeParty {
		if hasExplicitActivityLocation(req) {
			if err := validateExplicitActivityLocation(req); err != nil {
				return 0, err
			}
		} else {
			var profile models.OrganizerProfile
			if err := s.DB.WithContext(ctx).Where("organizer_id = ?", org.ID).First(&profile).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return 0, errors.New("场地固定资料未完善，暂不能发布活动")
				}
				return 0, err
			}
			updates["province"] = org.Province
			updates["city"] = org.City
			updates["district"] = org.District
			updates["address"] = profile.Address
			updates["latitude"] = profile.Latitude
			updates["longitude"] = profile.Longitude
		}
	}
	if activityType == models.ActivityTypeVenue && !wasVenue {
		startTime, endTime := venueValidityWindow(time.Now())
		updates["start_time"] = startTime
		updates["end_time"] = endTime
	}
	if req.ActivityID > 0 && act.Status == models.ActivityStatusOnline && activityEditNeedsReaudit(req, updates) {
		if err := s.saveActivityRevision(ctx, org.ID, &act, req, updates); err != nil {
			return 0, err
		}
		return act.ID, nil
	}
	if req.Step == 4 {
		if err := s.SaveTicketSpecs(ctx, userID, act.ID, req.TicketSpecs); err != nil {
			return 0, err
		}
	}
	if len(updates) > 0 {
		if err := s.DB.WithContext(ctx).Model(&act).Updates(updates).Error; err != nil {
			return 0, err
		}
	}
	if req.BusinessHours != nil {
		if err := s.DB.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "organizer_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"business_hours": *req.BusinessHours,
				"updated_at":     time.Now(),
			}),
		}).Create(&models.OrganizerProfile{OrganizerID: org.ID, BusinessHours: *req.BusinessHours}).Error; err != nil {
			return 0, err
		}
	}
	if req.TagIDs != nil {
		targetType, targetID := models.ContentTagTargetActivity, act.ID
		if activityType == models.ActivityTypeVenue {
			targetType, targetID = models.ContentTagTargetVenue, org.ID
		}
		if err := dao.ReplaceContentTags(ctx, s.DB, targetType, targetID, req.TagIDs); err != nil {
			return 0, err
		}
	}
	return act.ID, nil
}

func (s *TicketingService) GetActivity(ctx context.Context, userID, activityID int64) (*types.ActivityDetailResponse, error) {
	var act models.Activity
	if err := s.DB.WithContext(ctx).First(&act, activityID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			canView, historyErr := s.hasHistoricalActivityOrder(ctx, userID, activityID)
			if historyErr != nil {
				return nil, historyErr
			}
			if canView {
				return unavailableActivityDetail(activityID), nil
			}
		}
		return nil, err
	}
	var org models.Organizer
	if err := s.DB.WithContext(ctx).First(&org, act.OrganizerID).Error; err != nil {
		return nil, err
	}
	if userID != org.UserID && !isActivityPublic(act) {
		// A platform-hidden activity remains available to ticket holders for
		// historical order lookups. A merchant's pending revision, however, is
		// not public until the administrator approves it again.
		if act.IsHidden == 1 {
			canView, err := s.hasHistoricalActivityOrder(ctx, userID, act.ID)
			if err != nil {
				return nil, err
			}
			if !canView {
				return nil, gorm.ErrRecordNotFound
			}
		} else {
			return nil, gorm.ErrRecordNotFound
		}
	}
	revision, err := decodeActivityRevision(act)
	if err != nil {
		return nil, err
	}
	pendingRevisionStatus := act.PendingRevisionStatus
	pendingRevisionReason := act.PendingRevisionReason
	showPendingRevision := userID == org.UserID && act.IsHidden == 0 && revision != nil && pendingRevisionStatus == models.ActivityStatusPending
	if showPendingRevision {
		act = revision.Activity
		act.Status = models.ActivityStatusPending
		act.AuditType = models.ActivityAuditTypeReaudit
	}

	specs := make([]models.TicketSpec, 0)
	if showPendingRevision {
		specs = revision.TicketSpecs
	} else if defaultActivityType(act.Type) != models.ActivityTypeVenue {
		if err := s.DB.WithContext(ctx).Where("activity_id = ?", activityID).Order("id asc").Find(&specs).Error; err != nil {
			return nil, err
		}
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
	hasPendingRevision := userID == org.UserID && revision != nil && (pendingRevisionStatus == models.ActivityStatusPending || pendingRevisionStatus == models.ActivityStatusRejected)
	if userID != org.UserID {
		pendingRevisionReason = ""
	}
	resp := &types.ActivityDetailResponse{
		Activity:              act,
		UserID:                org.UserID,
		TagIDs:                types.ContentTagIDs(tags),
		Tags:                  types.BuildContentTagItems(tags),
		TicketSpecs:           specs,
		Organizer:             &org,
		IsFollow:              followed[contentFollowIDForActivity(act)],
		FollowCount:           followCounts[contentFollowIDForActivity(act)],
		FollowTargetType:      contentFollowTargetForActivity(act),
		FollowTargetID:        contentFollowIDForActivity(act),
		HasPendingRevision:    hasPendingRevision,
		PendingRevisionReason: pendingRevisionReason,
	}
	if showPendingRevision {
		resp.TagIDs = revision.TagIDs
	}
	if defaultActivityType(act.Type) == models.ActivityTypeVenue {
		var profile models.OrganizerProfile
		if err := s.DB.WithContext(ctx).Where("organizer_id = ?", act.OrganizerID).First(&profile).Error; err == nil {
			resp.BusinessHours = profile.BusinessHours
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if showPendingRevision {
			resp.BusinessHours = revision.BusinessHours
		}
	}
	if userID > 0 {
		var count int64
		if defaultActivityType(act.Type) == models.ActivityTypeVenue {
			_ = s.DB.WithContext(ctx).Model(&models.VenueSubscription{}).
				Where("organizer_id = ? AND user_id = ?", act.OrganizerID, userID).
				Count(&count).Error
		} else {
			_ = s.DB.WithContext(ctx).Model(&models.ActivitySubscription{}).
				Where("activity_id = ? AND user_id = ?", activityID, userID).
				Count(&count).Error
		}
		resp.IsSubscribe = count > 0
	}
	return resp, nil
}

// RecordActivityView records a successful public activity-detail load. Logged
// in users are naturally deduplicated; guests are deduplicated by a hashed
// client-generated visitor ID and never persist the raw identifier.
func (s *TicketingService) RecordActivityView(ctx context.Context, userID, activityID int64, visitorID string) error {
	if activityID <= 0 {
		return nil
	}
	now := time.Now()
	statDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	visitorKey := activityVisitorKey(userID, visitorID)

	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		stat := models.ActivityDailyStat{ActivityID: activityID, StatDate: statDate, ViewCount: 1}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "activity_id"}, {Name: "stat_date"}},
			DoUpdates: clause.Assignments(map[string]any{
				"view_count": gorm.Expr("view_count + ?", 1),
				"updated_at": now,
			}),
		}).Create(&stat).Error; err != nil {
			return err
		}
		if visitorKey == "" {
			return nil
		}

		visitor := models.ActivityDailyVisitor{
			ActivityID: activityID,
			StatDate:   statDate,
			VisitorKey: visitorKey,
			UserID:     userID,
		}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&visitor)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		return tx.Model(&models.ActivityDailyStat{}).
			Where("activity_id = ? AND stat_date = ?", activityID, statDate).
			Updates(map[string]any{
				"visitor_count": gorm.Expr("visitor_count + ?", 1),
				"updated_at":    now,
			}).Error
	})
}

func activityVisitorKey(userID int64, visitorID string) string {
	if userID > 0 {
		return fmt.Sprintf("u:%d", userID)
	}
	visitorID = strings.TrimSpace(visitorID)
	if visitorID == "" || len(visitorID) > 256 {
		return ""
	}
	hash := sha256.Sum256([]byte(visitorID))
	return "g:" + hex.EncodeToString(hash[:])
}

func (s *TicketingService) SubscribeActivity(ctx context.Context, userID, activityID int64) error {
	var activity models.Activity
	if err := s.DB.WithContext(ctx).
		Where("id = ? AND status = ? AND is_hidden = 0", activityID, models.ActivityStatusOnline).
		First(&activity).Error; err != nil {
		return err
	}
	if defaultActivityType(activity.Type) == models.ActivityTypeVenue {
		return s.SubscribeVenue(ctx, userID, activity.OrganizerID)
	}
	sub := models.ActivitySubscription{ActivityID: activityID, UserID: userID}
	return s.DB.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&sub).Error
}

func (s *TicketingService) UnsubscribeActivity(ctx context.Context, userID, activityID int64) error {
	var activity models.Activity
	err := s.DB.WithContext(ctx).Select("id", "type", "organizer_id").Where("id = ?", activityID).First(&activity).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err == nil && defaultActivityType(activity.Type) == models.ActivityTypeVenue {
		return s.UnsubscribeVenue(ctx, userID, activity.OrganizerID)
	}
	return s.DB.WithContext(ctx).
		Where("activity_id = ? AND user_id = ?", activityID, userID).
		Delete(&models.ActivitySubscription{}).Error
}

func (s *TicketingService) ListSubscribedActivities(ctx context.Context, userID int64, page, size int) (*types.PageResponse[types.ActivityListItem], error) {
	page, size = normalizePage(page, size)
	type subscribedActivityRow struct {
		ID           int64
		OrganizerID  int64
		Type         string
		Name         string
		PosterList   string
		StartTime    time.Time
		EndTime      time.Time
		Status       int8
		AuditType    string
		SubscribedAt time.Time
	}

	// Venue subscriptions are organizer-scoped for compatibility, while party
	// subscriptions are activity-scoped. Merge both sources here so the legacy
	// "subscribed activities" page does not silently lose subscribed venues.
	rowsByActivityID := make(map[int64]subscribedActivityRow)
	mergeRows := func(rows []subscribedActivityRow) {
		for _, row := range rows {
			current, exists := rowsByActivityID[row.ID]
			if !exists || row.SubscribedAt.After(current.SubscribedAt) {
				rowsByActivityID[row.ID] = row
			}
		}
	}
	loadRows := func(query *gorm.DB, subscribedAtColumn string) error {
		var rows []subscribedActivityRow
		if err := query.Select("a.id, a.organizer_id, a.type, a.name, a.poster_list, a.start_time, a.end_time, a.status, a.audit_type, " + subscribedAtColumn + " AS subscribed_at").Scan(&rows).Error; err != nil {
			return err
		}
		mergeRows(rows)
		return nil
	}

	if err := loadRows(
		s.DB.WithContext(ctx).Table("activity_subscriptions AS sub").
			Joins("JOIN activities a ON a.id = sub.activity_id").
			Where("sub.user_id = ? AND a.type <> ? AND a.status = ? AND a.is_hidden = 0", userID, models.ActivityTypeVenue, models.ActivityStatusOnline),
		"sub.created_at",
	); err != nil {
		return nil, err
	}
	// Keep venue rows written by an early activity-subscription implementation
	// visible during the storage migration.
	if err := loadRows(
		s.DB.WithContext(ctx).Table("activity_subscriptions AS sub").
			Joins("JOIN activities a ON a.id = sub.activity_id").
			Where("sub.user_id = ? AND a.type = ? AND a.status = ? AND a.is_hidden = 0", userID, models.ActivityTypeVenue, models.ActivityStatusOnline),
		"sub.created_at",
	); err != nil {
		return nil, err
	}
	if err := loadRows(
		s.DB.WithContext(ctx).Table("venue_subscriptions AS sub").
			Joins("JOIN activities a ON a.organizer_id = sub.organizer_id").
			Joins("JOIN organizers o ON o.id = a.organizer_id").
			Where("sub.user_id = ? AND a.type = ? AND a.status = ? AND a.is_hidden = 0 AND o.status = ? AND o.enabled = 1", userID, models.ActivityTypeVenue, models.ActivityStatusOnline, models.OrganizerStatusApproved),
		"sub.created_at",
	); err != nil {
		return nil, err
	}

	rows := make([]subscribedActivityRow, 0, len(rowsByActivityID))
	for _, row := range rowsByActivityID {
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].SubscribedAt.After(rows[j].SubscribedAt) })
	total := int64(len(rows))
	start := (page - 1) * size
	if start >= len(rows) {
		return &types.PageResponse[types.ActivityListItem]{List: []types.ActivityListItem{}, Total: total}, nil
	}
	end := start + size
	if end > len(rows) {
		end = len(rows)
	}
	rows = rows[start:end]

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
			AuditType:        row.AuditType,
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
	org, err := s.findApprovedOrganizerForActivityList(ctx, userID)
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
	resp, err := s.listActivities(query, page, size, userID)
	if err != nil {
		return nil, err
	}
	if len(resp.List) == 0 {
		return resp, nil
	}
	activityIDs := make([]int64, 0, len(resp.List))
	for _, item := range resp.List {
		activityIDs = append(activityIDs, item.ID)
	}
	var revisions []models.Activity
	if err := s.DB.WithContext(ctx).Select("id", "pending_revision_status", "pending_revision_reason").
		Where("organizer_id = ? AND id IN ? AND pending_revision_status IN ?", org.ID, activityIDs, []int8{models.ActivityStatusPending, models.ActivityStatusRejected}).Find(&revisions).Error; err != nil {
		return nil, err
	}
	byID := make(map[int64]models.Activity, len(revisions))
	for _, activity := range revisions {
		byID[activity.ID] = activity
	}
	for i := range resp.List {
		if activity, ok := byID[resp.List[i].ID]; ok {
			resp.List[i].HasPendingRevision = true
			resp.List[i].PendingRevisionReason = activity.PendingRevisionReason
		}
	}
	return resp, nil
}

func (s *TicketingService) findApprovedOrganizerForActivityList(ctx context.Context, userID int64) (*models.Organizer, error) {
	var org models.Organizer
	if err := s.DB.WithContext(ctx).
		Where("user_id = ? AND status = ? AND enabled = ?", userID, models.OrganizerStatusApproved, 1).
		Order("id DESC").
		First(&org).Error; err == nil {
		return &org, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return s.findOrganizerByUser(ctx, userID)
}

func (s *TicketingService) SearchActivities(ctx context.Context, userID int64, keyword string) ([]types.ActivityListItem, error) {
	query := s.DB.WithContext(ctx).Table("activities AS a").
		Select("a.*").
		Joins("JOIN organizers o ON o.id = a.organizer_id").
		Where("a.type <> ? AND a.status = ? AND a.is_hidden = 0 AND a.end_time >= ? AND o.status = ? AND o.enabled = 1", models.ActivityTypeVenue, models.ActivityStatusOnline, time.Now(), models.OrganizerStatusApproved)
	if keyword != "" {
		query = query.Where("a.name LIKE ? OR o.name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
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
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var activity models.Activity
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND organizer_id = ?", activityID, org.ID).
			First(&activity).Error; err != nil {
			return err
		}

		var orderCount int64
		if err := tx.Model(&models.TicketOrder{}).Where("activity_id = ?", activityID).Count(&orderCount).Error; err != nil {
			return err
		}
		if orderCount > 0 {
			now := time.Now()
			return tx.Model(&activity).Updates(map[string]any{
				"is_hidden":     1,
				"hidden_at":     now,
				"hidden_reason": "主办方已下架活动",
			}).Error
		}
		return tx.Delete(&activity).Error
	})
}

const unavailableActivityName = "活动已下架"

func unavailableActivityDetail(activityID int64) *types.ActivityDetailResponse {
	return &types.ActivityDetailResponse{
		Activity: models.Activity{
			ID:           activityID,
			Name:         unavailableActivityName,
			Status:       models.ActivityStatusOnline,
			IsHidden:     1,
			HiddenReason: unavailableActivityName,
		},
		Tags:        []types.ContentTagItem{},
		TicketSpecs: []models.TicketSpec{},
	}
}

func applyOrderActivityListItem(item *types.TicketOrderListItem, activityID, activityRecordID int64, name string, startTime, endTime time.Time, posterList string, hidden int8, hiddenReason string) {
	item.Activity.ID = activityID
	if activityRecordID == 0 {
		item.Activity.Name = unavailableActivityName
		item.Activity.IsHidden = true
		item.Activity.HiddenReason = unavailableActivityName
		return
	}
	item.Activity.Name = name
	item.Activity.StartTime = startTime
	item.Activity.EndTime = endTime
	item.Activity.PosterList = posterList
	item.Activity.IsHidden = hidden == 1
	item.Activity.HiddenReason = hiddenReason
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
	switch activity.Status {
	case models.ActivityStatusPending:
		return nil
	case models.ActivityStatusDraft, models.ActivityStatusRejected:
		return s.DB.WithContext(ctx).Model(&activity).Updates(map[string]any{
			"status":        models.ActivityStatusPending,
			"audit_type":    normalizeActivityAuditType(activity.AuditType),
			"reject_reason": "",
			"updated_at":    time.Now(),
		}).Error
	case models.ActivityStatusOnline:
		return errors.New("已上架内容修改后会自动提交二次审核")
	default:
		return errors.New("活动正在审核中，请勿重复提交")
	}
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
		Row().Scan(&resp.TicketCount, &resp.GrossAmount, &resp.BuyerCount, &resp.PaidOrderCount); err != nil {
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
	traffic, err := s.loadActivityTrafficStats(ctx, []int64{activityID}, time.Time{}, time.Time{})
	if err != nil {
		return nil, err
	}
	resp.ViewCount = traffic[activityID].ViewCount
	resp.VisitorCount = traffic[activityID].VisitorCount
	if resp.VisitorCount > 0 {
		resp.ConversionRate = float64(resp.PaidOrderCount) / float64(resp.VisitorCount)
	}
	withdraws, err := s.loadActivityWithdrawAmounts(ctx, []int64{activityID})
	if err != nil {
		return nil, err
	}
	resp.PendingWithdrawAmount = withdraws[activityID].PendingAmount
	resp.WithdrawnAmount = withdraws[activityID].SettledAmount
	resp.AvailableWithdrawAmount = maxInt64(resp.NetAmount-resp.PendingWithdrawAmount-resp.WithdrawnAmount, 0)
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

	byDate := make(map[string]*types.ActivityDailyStatisticsItem, len(rows))
	for i := range rows {
		byDate[rows[i].Date] = &rows[i]
	}
	var trafficRows []struct {
		StatDate     time.Time
		ViewCount    int64
		VisitorCount int64
	}
	if err := s.DB.WithContext(ctx).Model(&models.ActivityDailyStat{}).
		Select("stat_date, view_count, visitor_count").
		Where("activity_id = ? AND stat_date >= ?", activityID, start).
		Order("stat_date ASC").Scan(&trafficRows).Error; err != nil {
		return nil, err
	}
	for _, traffic := range trafficRows {
		date := traffic.StatDate.Format("2006-01-02")
		item, ok := byDate[date]
		if !ok {
			item = &types.ActivityDailyStatisticsItem{Date: date}
			byDate[date] = item
		}
		item.ViewCount = traffic.ViewCount
		item.VisitorCount = traffic.VisitorCount
	}
	merged := make([]types.ActivityDailyStatisticsItem, 0, len(byDate))
	for _, item := range byDate {
		merged = append(merged, *item)
	}
	sort.Slice(merged, func(i, j int) bool { return merged[i].Date < merged[j].Date })
	return &types.PageResponse[types.ActivityDailyStatisticsItem]{List: merged, Total: int64(len(merged))}, nil
}

type activityTrafficAmounts struct {
	ViewCount    int64
	VisitorCount int64
}

func (s *TicketingService) loadActivityTrafficStats(ctx context.Context, activityIDs []int64, start, end time.Time) (map[int64]activityTrafficAmounts, error) {
	result := make(map[int64]activityTrafficAmounts, len(activityIDs))
	if len(activityIDs) == 0 {
		return result, nil
	}
	statQuery := s.DB.WithContext(ctx).Model(&models.ActivityDailyStat{}).
		Select("activity_id, COALESCE(SUM(view_count), 0) AS view_count").
		Where("activity_id IN ?", activityIDs)
	visitorQuery := s.DB.WithContext(ctx).Model(&models.ActivityDailyVisitor{}).
		Select("activity_id, COUNT(DISTINCT visitor_key) AS visitor_count").
		Where("activity_id IN ?", activityIDs)
	if !start.IsZero() {
		statQuery = statQuery.Where("stat_date >= ?", start)
		visitorQuery = visitorQuery.Where("stat_date >= ?", start)
	}
	if !end.IsZero() {
		statQuery = statQuery.Where("stat_date < ?", end)
		visitorQuery = visitorQuery.Where("stat_date < ?", end)
	}
	var views []struct {
		ActivityID int64
		ViewCount  int64
	}
	if err := statQuery.Group("activity_id").Scan(&views).Error; err != nil {
		return nil, err
	}
	for _, row := range views {
		amounts := result[row.ActivityID]
		amounts.ViewCount = row.ViewCount
		result[row.ActivityID] = amounts
	}
	var visitors []struct {
		ActivityID   int64
		VisitorCount int64
	}
	if err := visitorQuery.Group("activity_id").Scan(&visitors).Error; err != nil {
		return nil, err
	}
	for _, row := range visitors {
		amounts := result[row.ActivityID]
		amounts.VisitorCount = row.VisitorCount
		result[row.ActivityID] = amounts
	}
	return result, nil
}

func (s *TicketingService) loadOrganizerTrafficTotals(ctx context.Context, organizerID int64, start, end time.Time) (activityTrafficAmounts, error) {
	result := activityTrafficAmounts{}
	statQuery := s.DB.WithContext(ctx).Table("activity_daily_stats ads").
		Joins("JOIN activities a ON a.id = ads.activity_id").
		Where("a.organizer_id = ?", organizerID)
	visitorQuery := s.DB.WithContext(ctx).Table("activity_daily_visitors adv").
		Joins("JOIN activities a ON a.id = adv.activity_id").
		Where("a.organizer_id = ?", organizerID)
	if !start.IsZero() {
		statQuery = statQuery.Where("ads.stat_date >= ?", start)
		visitorQuery = visitorQuery.Where("adv.stat_date >= ?", start)
	}
	if !end.IsZero() {
		statQuery = statQuery.Where("ads.stat_date < ?", end)
		visitorQuery = visitorQuery.Where("adv.stat_date < ?", end)
	}
	if err := statQuery.Select("COALESCE(SUM(ads.view_count), 0)").Scan(&result.ViewCount).Error; err != nil {
		return result, err
	}
	if err := visitorQuery.Select("COUNT(DISTINCT adv.visitor_key)").Scan(&result.VisitorCount).Error; err != nil {
		return result, err
	}
	return result, nil
}

type activityWithdrawAmounts struct {
	PendingAmount int64
	SettledAmount int64
}

func (s *TicketingService) loadActivityWithdrawAmounts(ctx context.Context, activityIDs []int64) (map[int64]activityWithdrawAmounts, error) {
	result := make(map[int64]activityWithdrawAmounts, len(activityIDs))
	if len(activityIDs) == 0 {
		return result, nil
	}
	var rows []struct {
		ActivityID    int64
		PendingAmount int64
		SettledAmount int64
	}
	if err := s.DB.WithContext(ctx).Model(&models.OrganizerWithdrawAllocation{}).
		Select(`activity_id,
			COALESCE(SUM(CASE WHEN status = 0 THEN amount ELSE 0 END), 0) AS pending_amount,
			COALESCE(SUM(CASE WHEN status = 1 THEN amount ELSE 0 END), 0) AS settled_amount`).
		Where("activity_id IN ? AND status IN ?", activityIDs, []int8{
			models.OrganizerWithdrawAllocationStatusPending,
			models.OrganizerWithdrawAllocationStatusSettled,
		}).
		Group("activity_id").Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.ActivityID] = activityWithdrawAmounts{PendingAmount: row.PendingAmount, SettledAmount: row.SettledAmount}
	}
	return result, nil
}

func maxInt64(value, min int64) int64 {
	if value < min {
		return min
	}
	return value
}

func (s *TicketingService) GetTicketSpecs(ctx context.Context, activityID int64) ([]models.TicketSpec, error) {
	var activity models.Activity
	if err := s.DB.WithContext(ctx).First(&activity, activityID).Error; err != nil {
		return nil, err
	}
	if defaultActivityType(activity.Type) == models.ActivityTypeVenue {
		return []models.TicketSpec{}, nil
	}
	var specs []models.TicketSpec
	err := s.DB.WithContext(ctx).Where("activity_id = ?", activityID).Order("id asc").Find(&specs).Error
	return specs, err
}

func (s *TicketingService) SaveTicketSpecs(ctx context.Context, userID, activityID int64, specs []types.TicketSpecSaveItem) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	var activity models.Activity
	if err := s.DB.WithContext(ctx).Where("id = ? AND organizer_id = ?", activityID, org.ID).First(&activity).Error; err != nil {
		return err
	}
	if defaultActivityType(activity.Type) == models.ActivityTypeVenue {
		return errors.New("场地不支持票券配置")
	}
	if activity.Status == models.ActivityStatusOnline {
		return s.saveActivityRevision(ctx, org.ID, &activity, types.ActivityCreateRequest{
			ActivityID:  activityID,
			Step:        4,
			TicketSpecs: specs,
		}, map[string]any{})
	}
	if err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range specs {
			spec, err := buildTicketSpec(activityID, item)
			if err != nil {
				return err
			}
			if item.ID > 0 {
				// Map updates deliberately retain zero values: a merchant must be
				// able to disable a ticket or set its stock/price to zero.
				if err := tx.Model(&models.TicketSpec{}).Where("id = ? AND activity_id = ?", item.ID, activityID).Updates(ticketSpecUpdates(spec)).Error; err != nil {
					return err
				}
				continue
			}
			if err := tx.Create(spec).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *TicketingService) DeleteTicketSpec(ctx context.Context, userID, specID int64) error {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return err
	}
	var spec models.TicketSpec
	if err := s.DB.WithContext(ctx).
		Where("id = ? AND activity_id IN (?)", specID, s.DB.Model(&models.Activity{}).Select("id").Where("organizer_id = ?", org.ID)).
		First(&spec).Error; err != nil {
		return err
	}
	var activity models.Activity
	if err := s.DB.WithContext(ctx).Where("id = ?", spec.ActivityID).First(&activity).Error; err != nil {
		return err
	}
	if activity.Status == models.ActivityStatusOnline {
		return s.deleteRevisionTicketSpec(ctx, org.ID, &activity, specID)
	}
	if err := s.DB.WithContext(ctx).Delete(&spec).Error; err != nil {
		return err
	}
	return nil
}

func (s *TicketingService) CreateTicketOrder(ctx context.Context, userID int64, req types.CreateTicketOrderRequest) (*types.CreateTicketOrderResponse, error) {
	salesChannel, err := NormalizeSalesChannel(req.SalesChannel, true)
	if err != nil {
		return nil, err
	}
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
		if !isActivityPublic(act) {
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
			SalesChannel:   salesChannel,
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
		result.SalesChannel = salesChannel
		result.Status = orderStatus
		result.QRCode = qrContent
		result.QRCodeURL = qrURL
		result.Viewers = orderViewerItems(orderViewers, false)
		return nil
	})
	return result, err
}

func isActivityPublic(activity models.Activity) bool {
	return activity.Status == models.ActivityStatusOnline && activity.IsHidden == 0
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
	query := s.DB.WithContext(ctx).Model(&models.TicketOrder{}).
		Where("ticket_orders.user_id = ? AND ticket_orders.user_deleted_at IS NULL", userID)
	if status != nil {
		switch *status {
		case models.TicketOrderStatusPending:
			query = query.Where("ticket_orders.status = ? AND ticket_orders.actual_price > 0", *status)
		case models.TicketOrderStatusUsable:
			query = query.Where("(ticket_orders.status = ? OR (ticket_orders.status = ? AND ticket_orders.actual_price = 0))", *status, models.TicketOrderStatusPending)
		default:
			query = query.Where("ticket_orders.status = ?", *status)
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
		ID               int64
		OrderNo          string
		Status           int8
		TotalPrice       int64
		ActualPrice      int64
		PointsAmount     int64
		Quantity         int
		ActivityID       int64
		ActivityRecordID int64
		ActivityName     string
		StartTime        time.Time
		EndTime          time.Time
		PosterList       string
		ActivityHidden   int8
		HiddenReason     string
		TicketSpecID     int64
		TicketSpecName   string
		BuyerName        string
		BuyerIDCard      string
		CreatedAt        time.Time
		ExpireTime       time.Time
		PayTime          *time.Time
		RefundNo         string
		RefundStatus     *int8
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
			COALESCE(activities.id, 0) AS activity_record_id,
			activities.name AS activity_name,
			activities.start_time,
			activities.end_time,
			activities.poster_list,
			COALESCE(activities.is_hidden, 0) AS activity_hidden,
			COALESCE(activities.hidden_reason, '') AS hidden_reason,
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
		applyOrderActivityListItem(&item, r.ActivityID, r.ActivityRecordID, r.ActivityName, r.StartTime, r.EndTime, r.PosterList, r.ActivityHidden, r.HiddenReason)
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

func (s *TicketingService) ListOrganizerOrders(ctx context.Context, userID int64, activityID int64, status *int8, keyword, salesChannel, withdrawStatus, startDate, endDate string, page, size int) (*types.PageResponse[types.OrganizerOrderListItem], error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	page, size = normalizePage(page, size)
	query := s.DB.WithContext(ctx).Table("ticket_orders o").
		Joins("JOIN activities a ON a.id = o.activity_id").
		Joins("JOIN ticket_specs ts ON ts.id = o.ticket_spec_id").
		Joins("LEFT JOIN users u ON u.id = o.user_id").
		Joins("LEFT JOIN (SELECT order_id, MAX(verified_at) AS verified_at FROM verification_records GROUP BY order_id) vr ON vr.order_id = o.id").
		Where("a.organizer_id = ?", org.ID)
	if activityID > 0 {
		query = query.Where("o.activity_id = ?", activityID)
	}
	if status != nil {
		query = query.Where("o.status = ?", *status)
	}
	salesChannel, err = NormalizeSalesChannel(salesChannel, false)
	if err != nil {
		return nil, err
	}
	if salesChannel != "" {
		query = query.Where("o.sales_channel = ?", salesChannel)
	}
	query, err = applyOrganizerOrderWithdrawStatusFilter(query, strings.TrimSpace(withdrawStatus))
	if err != nil {
		return nil, err
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
		ID             int64
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
		PosterList     string
		TicketSpecID   int64
		TicketSpecName string
		PayMethod      string
		SalesChannel   string
		PayTime        *time.Time
		VerifiedAt     *time.Time
		CreatedAt      time.Time
		ExpireTime     time.Time
	}
	if err := query.Select(`
			o.id,
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
			COALESCE(NULLIF(a.poster_list, ''), NULLIF(a.poster_detail, ''), NULLIF(a.poster_wechat, ''), NULLIF(a.poster_long, '')) AS poster_list,
			o.ticket_spec_id,
			ts.name AS ticket_spec_name,
			o.pay_method,
			o.sales_channel,
			o.pay_time,
			vr.verified_at,
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
	orderIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		orderNos = append(orderNos, row.OrderNo)
		orderIDs = append(orderIDs, row.ID)
	}
	viewersByOrderNo, err := s.orderViewersByOrderNo(ctx, orderNos)
	if err != nil {
		return nil, err
	}
	withdrawAmounts, err := s.loadOrderWithdrawAmounts(ctx, orderIDs)
	if err != nil {
		return nil, err
	}
	list := make([]types.OrganizerOrderListItem, 0, len(rows))
	for _, row := range rows {
		withdraw := withdrawAmounts[row.ID]
		list = append(list, types.OrganizerOrderListItem{
			OrderNo:          row.OrderNo,
			Status:           row.Status,
			TotalPrice:       row.TotalPrice,
			ActualPrice:      row.ActualPrice,
			PointsAmount:     row.PointsAmount,
			PointsDiscount:   row.PointsDiscount,
			Quantity:         row.Quantity,
			UserID:           row.UserID,
			UserName:         row.UserName,
			UserMobile:       maskPhone(row.UserMobile),
			BuyerPhoneMasked: maskPhone(row.UserMobile),
			UserAvatar:       row.UserAvatar,
			BuyerName:        row.BuyerName,
			BuyerIDCard:      maskIDCard(row.BuyerIDCard),
			Viewers:          orderViewerItems(viewersByOrderNo[row.OrderNo], false),
			ActivityID:       row.ActivityID,
			ActivityName:     row.ActivityName,
			PosterList:       row.PosterList,
			TicketSpecID:     row.TicketSpecID,
			TicketSpecName:   row.TicketSpecName,
			PayMethod:        row.PayMethod,
			SalesChannel:     row.SalesChannel,
			PayTime:          row.PayTime,
			VerifiedAt:       row.VerifiedAt,
			CreatedAt:        row.CreatedAt,
			ExpireTime:       row.ExpireTime,
			WithdrawStatus:   withdrawStatusForOrder(row.Status, withdraw),
			WithdrawAmount:   withdraw.PendingAmount + withdraw.SettledAmount,
		})
	}
	return &types.PageResponse[types.OrganizerOrderListItem]{List: list, Total: total}, nil
}

type orderWithdrawAmounts struct {
	PendingAmount int64
	SettledAmount int64
}

func (s *TicketingService) loadOrderWithdrawAmounts(ctx context.Context, orderIDs []int64) (map[int64]orderWithdrawAmounts, error) {
	result := make(map[int64]orderWithdrawAmounts, len(orderIDs))
	if len(orderIDs) == 0 {
		return result, nil
	}
	var rows []struct {
		OrderID       int64
		PendingAmount int64
		SettledAmount int64
	}
	if err := s.DB.WithContext(ctx).Model(&models.OrganizerWithdrawAllocation{}).
		Select(`order_id,
			COALESCE(SUM(CASE WHEN status = 0 THEN amount ELSE 0 END), 0) AS pending_amount,
			COALESCE(SUM(CASE WHEN status = 1 THEN amount ELSE 0 END), 0) AS settled_amount`).
		Where("order_id IN ? AND status IN ?", orderIDs, []int8{
			models.OrganizerWithdrawAllocationStatusPending,
			models.OrganizerWithdrawAllocationStatusSettled,
		}).
		Group("order_id").Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.OrderID] = orderWithdrawAmounts{PendingAmount: row.PendingAmount, SettledAmount: row.SettledAmount}
	}
	return result, nil
}

func withdrawStatusForOrder(orderStatus int8, amounts orderWithdrawAmounts) string {
	if amounts.PendingAmount > 0 {
		return "pending_withdraw"
	}
	if amounts.SettledAmount > 0 {
		return "withdrawn"
	}
	switch orderStatus {
	case models.TicketOrderStatusUsable, models.TicketOrderStatusUsed, models.TicketOrderStatusRefundReject:
		return "available"
	default:
		return "unavailable"
	}
}

func applyOrganizerOrderWithdrawStatusFilter(query *gorm.DB, status string) (*gorm.DB, error) {
	switch status {
	case "":
		return query, nil
	case "pending_withdraw":
		return query.Where(`EXISTS (
			SELECT 1 FROM organizer_withdraw_allocations owa
			WHERE owa.order_id = o.id AND owa.status = ?
		)`, models.OrganizerWithdrawAllocationStatusPending), nil
	case "withdrawn":
		return query.Where(`EXISTS (
			SELECT 1 FROM organizer_withdraw_allocations owa
			WHERE owa.order_id = o.id AND owa.status = ?
		)`, models.OrganizerWithdrawAllocationStatusSettled), nil
	case "available":
		return query.Where("o.status IN ?", []int8{
			models.TicketOrderStatusUsable,
			models.TicketOrderStatusUsed,
			models.TicketOrderStatusRefundReject,
		}).Where(`NOT EXISTS (
			SELECT 1 FROM organizer_withdraw_allocations owa
			WHERE owa.order_id = o.id AND owa.status IN (?, ?)
		)`, models.OrganizerWithdrawAllocationStatusPending, models.OrganizerWithdrawAllocationStatusSettled), nil
	default:
		return nil, errors.New("withdraw_status 仅支持 available、pending_withdraw、withdrawn")
	}
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
	var totals struct {
		TotalAmount int64
		OrderCount  int64
		TicketCount int64
	}
	if err := base().
		Select("COALESCE(SUM(o.actual_price), 0) AS total_amount, COUNT(o.id) AS order_count, COALESCE(SUM(o.quantity), 0) AS ticket_count").
		Scan(&totals).Error; err != nil {
		return nil, err
	}
	resp.TotalAmount = totals.TotalAmount
	resp.OrderCount = totals.OrderCount
	resp.TicketCount = totals.TicketCount
	if resp.OrderCount > 0 {
		resp.AverageOrderAmount = resp.TotalAmount / resp.OrderCount
	}
	traffic, err := s.loadOrganizerTrafficTotals(ctx, org.ID, start, end)
	if err != nil {
		return nil, err
	}
	resp.ViewCount = traffic.ViewCount
	resp.VisitorCount = traffic.VisitorCount
	paidOrderCounts, err := s.loadActivityPaidOrderCounts(ctx, org.ID, nil, start, end)
	if err != nil {
		return nil, err
	}
	for _, count := range paidOrderCounts {
		resp.PaidOrderCount += count
	}
	if resp.VisitorCount > 0 {
		resp.ConversionRate = float64(resp.PaidOrderCount) / float64(resp.VisitorCount)
	}

	if err := base().
		Select("o.activity_id, a.name AS activity_name, COUNT(o.id) AS order_count, COALESCE(SUM(o.quantity), 0) AS ticket_count, COALESCE(SUM(o.actual_price), 0) AS total_amount").
		Group("o.activity_id, a.name").
		Order("total_amount DESC, order_count DESC, o.activity_id DESC").
		Scan(&resp.ActivityRanks).Error; err != nil {
		return nil, err
	}
	rankActivityIDs := make([]int64, 0, len(resp.ActivityRanks))
	for i := range resp.ActivityRanks {
		rankActivityIDs = append(rankActivityIDs, resp.ActivityRanks[i].ActivityID)
	}
	activityTraffic, err := s.loadActivityTrafficStats(ctx, rankActivityIDs, start, end)
	if err != nil {
		return nil, err
	}
	activityWithdraws, err := s.loadActivityWithdrawAmounts(ctx, rankActivityIDs)
	if err != nil {
		return nil, err
	}
	for i := range resp.ActivityRanks {
		rank := &resp.ActivityRanks[i]
		traffic := activityTraffic[rank.ActivityID]
		withdraw := activityWithdraws[rank.ActivityID]
		rank.ViewCount = traffic.ViewCount
		rank.VisitorCount = traffic.VisitorCount
		rank.PaidOrderCount = paidOrderCounts[rank.ActivityID]
		if rank.VisitorCount > 0 {
			rank.ConversionRate = float64(rank.PaidOrderCount) / float64(rank.VisitorCount)
		}
		rank.PendingWithdrawAmount = withdraw.PendingAmount
		rank.WithdrawnAmount = withdraw.SettledAmount
		rank.AvailableWithdrawAmount = maxInt64(rank.TotalAmount-rank.PendingWithdrawAmount-rank.WithdrawnAmount, 0)
	}
	return resp, nil
}

func (s *TicketingService) loadActivityPaidOrderCounts(ctx context.Context, organizerID int64, activityIDs []int64, start, end time.Time) (map[int64]int64, error) {
	result := make(map[int64]int64, len(activityIDs))
	query := s.DB.WithContext(ctx).Table("ticket_orders o").
		Select("o.activity_id, COUNT(o.id) AS order_count").
		Joins("JOIN activities a ON a.id = o.activity_id").
		Where("a.organizer_id = ? AND o.status IN ?", organizerID, []int8{
			models.TicketOrderStatusUsable,
			models.TicketOrderStatusUsed,
			models.TicketOrderStatusRefunding,
			models.TicketOrderStatusRefundSuccess,
			models.TicketOrderStatusRefundReject,
		})
	if len(activityIDs) > 0 {
		query = query.Where("o.activity_id IN ?", activityIDs)
	}
	if !start.IsZero() {
		query = query.Where("o.pay_time >= ?", start)
	}
	if !end.IsZero() {
		query = query.Where("o.pay_time < ?", end)
	}
	var rows []struct {
		ActivityID int64
		OrderCount int64
	}
	if err := query.Group("o.activity_id").Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.ActivityID] = row.OrderCount
	}
	return result, nil
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
	withdrawAmounts, err := s.loadOrderWithdrawAmounts(ctx, []int64{order.ID})
	if err != nil {
		return nil, err
	}
	withdraw := withdrawAmounts[order.ID]
	resp.WithdrawStatus = withdrawStatusForOrder(order.Status, withdraw)
	resp.WithdrawAmount = withdraw.PendingAmount + withdraw.SettledAmount
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

// GetVerifierOrderDetail only grants an activated verifier access to orders
// under its own organizer. User-facing /order/:order_no remains purchaser-only.
func (s *TicketingService) GetVerifierOrderDetail(ctx context.Context, userID int64, orderNo string) (*types.OrganizerOrderDetailResponse, error) {
	var verifier models.Verifier
	if err := s.DB.WithContext(ctx).Where("user_id = ? AND status = ?", userID, models.VerifierStatusActive).Order("id DESC").First(&verifier).Error; err != nil {
		return nil, err
	}
	var order models.TicketOrder
	if err := s.DB.WithContext(ctx).Table("ticket_orders o").
		Select("o.*").
		Joins("JOIN activities a ON a.id = o.activity_id").
		Where("o.order_no = ? AND a.organizer_id = ?", orderNo, verifier.OrganizerID).
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
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
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

func (s *TicketingService) ListVerifiers(ctx context.Context, userID int64, page, size int) (*types.PageResponse[types.OrganizerVerifierItem], error) {
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
	if err := query.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, err
	}
	verifierIDs := make([]int64, 0, len(list))
	for _, verifier := range list {
		verifierIDs = append(verifierIDs, verifier.ID)
	}
	verifiedCounts := make(map[int64]int64, len(verifierIDs))
	if len(verifierIDs) > 0 {
		var rows []struct {
			VerifierID    int64
			VerifiedCount int64
		}
		if err := s.DB.WithContext(ctx).Model(&models.VerificationRecord{}).
			Select("verifier_id, COUNT(*) AS verified_count").Where("verifier_id IN ?", verifierIDs).
			Group("verifier_id").Scan(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			verifiedCounts[row.VerifierID] = row.VerifiedCount
		}
	}
	items := make([]types.OrganizerVerifierItem, 0, len(list))
	for _, verifier := range list {
		items = append(items, types.OrganizerVerifierItem{Verifier: verifier, VerifiedCount: verifiedCounts[verifier.ID]})
	}
	return &types.PageResponse[types.OrganizerVerifierItem]{List: items, Total: total}, nil
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
	var buyer models.Users
	_ = s.DB.WithContext(ctx).Select("mobile").First(&buyer, order.UserID).Error
	item := struct {
		OrderNo           string                  `json:"order_no"`
		ActivityID        int64                   `json:"activity_id"`
		ActivityName      string                  `json:"activity_name"`
		PosterList        string                  `json:"poster_list"`
		TicketSpecName    string                  `json:"ticket_spec_name"`
		Quantity          int                     `json:"quantity"`
		BuyerNameMasked   string                  `json:"buyer_name_masked"`
		BuyerIDCardMasked string                  `json:"buyer_id_card_masked"`
		BuyerPhoneMasked  string                  `json:"buyer_phone_masked"`
		Viewers           []types.OrderViewerItem `json:"viewers,omitempty"`
	}{
		OrderNo:           order.OrderNo,
		ActivityID:        order.ActivityID,
		ActivityName:      activity.Name,
		PosterList:        firstNonEmpty(activity.PosterList, activity.PosterDetail, activity.PosterWechat, activity.PosterLong),
		TicketSpecName:    spec.Name,
		Quantity:          order.Quantity,
		BuyerNameMasked:   maskName(order.BuyerName),
		BuyerIDCardMasked: maskIDCard(order.BuyerIDCard),
		BuyerPhoneMasked:  maskPhone(buyer.Mobile),
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
	if verifierID <= 0 {
		return nil, errors.New("核销员身份未识别")
	}
	return s.listVerified(ctx, "vr.verifier_id = ?", verifierID, types.VerificationRecordFilter{Page: page, Size: size})
}

// ListVerifiedByUser returns records from every verifier identity bound to the
// same logged-in user. A verifier may be disabled and bound again later, which
// creates another verifier row but must not make earlier verification history
// disappear from that user's history.
func (s *TicketingService) ListVerifiedByUser(ctx context.Context, userID int64, page, size int) (*types.PageResponse[types.VerifiedListItem], error) {
	if _, err := s.ResolveBoundVerifierID(ctx, userID); err != nil {
		return nil, err
	}
	return s.listVerified(ctx, "v.user_id = ?", userID, types.VerificationRecordFilter{Page: page, Size: size})
}

// ListOrganizerVerificationRecords returns every verification performed for
// the current organizer, regardless of which of its verifiers performed it.
func (s *TicketingService) ListOrganizerVerificationRecords(ctx context.Context, userID int64, filter types.VerificationRecordFilter) (*types.PageResponse[types.VerifiedListItem], error) {
	org, err := s.findOrganizerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.listVerified(ctx, "v.organizer_id = ?", org.ID, filter)
}

func (s *TicketingService) listVerified(ctx context.Context, condition string, conditionValue any, filter types.VerificationRecordFilter) (*types.PageResponse[types.VerifiedListItem], error) {
	page, size := normalizePage(filter.Page, filter.Size)
	var total int64
	base := s.DB.WithContext(ctx).Table("verification_records vr").
		Joins("LEFT JOIN verifiers v ON v.id = vr.verifier_id").
		Joins("LEFT JOIN organizers og ON og.id = v.organizer_id").
		Joins("LEFT JOIN ticket_orders o ON o.id = vr.order_id").
		Joins("LEFT JOIN activities a ON a.id = vr.activity_id").
		Joins("LEFT JOIN ticket_specs ts ON ts.id = o.ticket_spec_id").
		Joins("LEFT JOIN users u ON u.id = o.user_id").
		Where(condition, conditionValue)
	if filter.VerifierID > 0 {
		base = base.Where("vr.verifier_id = ?", filter.VerifierID)
	}
	if filter.ActivityID > 0 {
		base = base.Where("vr.activity_id = ?", filter.ActivityID)
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		base = base.Where("o.order_no LIKE ? OR a.name LIKE ? OR v.name LIKE ? OR v.phone LIKE ?", like, like, like, like)
	}
	if startDate := strings.TrimSpace(filter.StartDate); startDate != "" {
		start, err := parseDateStart(startDate)
		if err != nil {
			return nil, errors.New("start_date 格式错误")
		}
		base = base.Where("vr.verified_at >= ?", start)
	}
	if endDate := strings.TrimSpace(filter.EndDate); endDate != "" {
		end, err := parseDateEnd(endDate)
		if err != nil {
			return nil, errors.New("end_date 格式错误")
		}
		base = base.Where("vr.verified_at < ?", end)
	}
	if err := base.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []struct {
		ID             int64
		OrderNo        string
		ActivityID     int64
		ActivityName   string
		PosterList     string
		TicketSpecName string
		Quantity       int
		VerifierID     int64
		VerifierName   string
		VerifierPhone  string
		OrganizerID    int64
		OrganizerName  string
		BuyerName      string
		BuyerIDCard    string
		BuyerPhone     string
		VerifiedAt     time.Time
	}
	if err := base.Select(`vr.id, COALESCE(o.order_no, '') AS order_no, vr.activity_id,
		COALESCE(NULLIF(a.name, ''), '活动已删除') AS activity_name,
		COALESCE(NULLIF(a.poster_list, ''), NULLIF(a.poster_detail, ''), NULLIF(a.poster_wechat, ''), NULLIF(a.poster_long, '')) AS poster_list,
		COALESCE(NULLIF(ts.name, ''), '票种已删除') AS ticket_spec_name,
		COALESCE(o.quantity, 0) AS quantity, COALESCE(o.buyer_name, '') AS buyer_name,
		COALESCE(o.buyer_id_card, '') AS buyer_id_card, COALESCE(u.mobile, '') AS buyer_phone,
		vr.verifier_id, COALESCE(v.name, '') AS verifier_name, COALESCE(v.phone, '') AS verifier_phone,
		COALESCE(v.organizer_id, 0) AS organizer_id, COALESCE(og.name, '') AS organizer_name, vr.verified_at`).
		Order("vr.id desc").Offset((page - 1) * size).Limit(size).Scan(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]types.VerifiedListItem, 0, len(rows))
	for _, r := range rows {
		list = append(list, types.VerifiedListItem{
			ID:                r.ID,
			OrderNo:           r.OrderNo,
			ActivityID:        r.ActivityID,
			ActivityName:      r.ActivityName,
			PosterList:        r.PosterList,
			TicketSpecName:    r.TicketSpecName,
			Quantity:          r.Quantity,
			VerifierID:        r.VerifierID,
			VerifierName:      r.VerifierName,
			VerifierPhone:     maskPhone(r.VerifierPhone),
			OrganizerID:       r.OrganizerID,
			OrganizerName:     r.OrganizerName,
			BuyerNameMasked:   maskName(r.BuyerName),
			BuyerIDCardMasked: maskIDCard(r.BuyerIDCard),
			BuyerPhoneMasked:  maskPhone(r.BuyerPhone),
			VerifiedAt:        r.VerifiedAt,
		})
	}
	return &types.PageResponse[types.VerifiedListItem]{List: list, Total: total}, nil
}

// ResolveBoundVerifierID derives the verifier identity from the authenticated
// Mini Program user. This avoids silently querying verifier_id=0 when an old
// client loses its X-Verifier-Id header after a restart.
func (s *TicketingService) ResolveBoundVerifierID(ctx context.Context, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, errors.New("请先登录核销员账号")
	}
	var verifier models.Verifier
	if err := s.DB.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, models.VerifierStatusActive).
		Order("id DESC").First(&verifier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("当前账号不是已激活核销员")
		}
		return 0, err
	}
	return verifier.ID, nil
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
		return &org, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return &org, err
	}
	// A staff account acts on behalf of the same organizer and must therefore
	// resolve to the organizer identity, not its personal profile.
	err = s.DB.WithContext(ctx).
		Table("organizer_staff AS os").
		Select("o.*").
		Joins("JOIN organizers o ON o.id = os.organizer_id").
		Where("os.user_id = ? AND os.status = ? AND o.status = ? AND o.enabled = ?", userID, 1, models.OrganizerStatusApproved, 1).
		Order("os.id DESC").
		Limit(1).
		Scan(&org).Error
	if err != nil {
		return &org, err
	}
	if org.ID == 0 {
		return &org, gorm.ErrRecordNotFound
	}
	return &org, nil
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
		list = append(list, types.ActivityListItem{ID: a.ID, Type: activityType, Name: a.Name, PosterList: a.PosterList, StartTime: a.StartTime, EndTime: a.EndTime, Status: a.Status, AuditType: a.AuditType, TagIDs: types.ContentTagIDs(tags), Tags: types.BuildContentTagItems(tags), IsFollow: isFollow, FollowCount: followCount, FollowTargetType: contentFollowTargetForActivity(a), FollowTargetID: contentFollowIDForActivity(a)})
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
	activityMissing := false
	if err := s.DB.WithContext(ctx).First(&act, order.ActivityID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			activityMissing = true
		} else {
			return nil, err
		}
	}
	var spec models.TicketSpec
	if err := s.DB.WithContext(ctx).First(&spec, order.TicketSpecID).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
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
		SalesChannel:   order.SalesChannel,
		PayTime:        order.PayTime,
		CreatedAt:      order.CreatedAt,
		QRCode:         order.QRCode,
		QRCodeURL:      order.QRCodeURL,
		ExpireTime:     order.ExpireTime,
	}
	if activityMissing {
		resp.Activity.ID = order.ActivityID
		resp.Activity.Name = unavailableActivityName
		resp.Activity.IsHidden = true
		resp.Activity.HiddenReason = unavailableActivityName
	} else {
		activityAddress := act.Address
		if activityAddress == "" {
			var profile models.OrganizerProfile
			if err := s.DB.WithContext(ctx).Where("organizer_id = ?", act.OrganizerID).First(&profile).Error; err == nil {
				activityAddress = profile.Address
			}
		}
		resp.Activity.ID = act.ID
		resp.Activity.Name = act.Name
		resp.Activity.Address = activityAddress
		resp.Activity.StartTime = act.StartTime
		resp.Activity.EndTime = act.EndTime
		resp.Activity.PosterList = firstNonEmpty(act.PosterList, act.PosterDetail, act.PosterWechat, act.PosterLong)
		resp.Activity.IsHidden = act.IsHidden == 1
		resp.Activity.HiddenReason = act.HiddenReason
	}
	if spec.ID == 0 {
		resp.TicketSpec.Name = "票券信息不可用"
	} else {
		resp.TicketSpec.Name = spec.Name
	}
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

// hasHistoricalActivityOrder grants a purchaser read-only access to a hidden
// activity's historical detail. It never re-enables subscribing or buying.
func (s *TicketingService) hasHistoricalActivityOrder(ctx context.Context, userID, activityID int64) (bool, error) {
	if userID <= 0 {
		return false, nil
	}
	statuses := []int8{
		models.TicketOrderStatusUsable,
		models.TicketOrderStatusUsed,
		models.TicketOrderStatusRefunding,
		models.TicketOrderStatusRefundSuccess,
		models.TicketOrderStatusRefundReject,
	}
	var count int64
	err := s.DB.WithContext(ctx).Model(&models.TicketOrder{}).
		Where("user_id = ? AND activity_id = ? AND user_deleted_at IS NULL", userID, activityID).
		Where("(status IN ? OR (status = ? AND actual_price = 0))", statuses, models.TicketOrderStatusPending).
		Count(&count).Error
	return count > 0, err
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
		MarkerIcon:     org.MarkerIcon,
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

func activityUpdates(req types.ActivityCreateRequest, activityType string) (map[string]any, error) {
	updates := map[string]any{}
	if req.Type != nil {
		activityType, err := normalizeActivityType(*req.Type)
		if err != nil {
			return nil, err
		}
		updates["type"] = activityType
	}
	putString(updates, "name", req.Name)
	if req.MarkerIcon != nil {
		markerIcon, err := normalizeMarkerIcon(*req.MarkerIcon)
		if err != nil {
			return nil, err
		}
		updates["marker_icon"] = markerIcon
	}
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
	if activityType == models.ActivityTypeParty {
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
	}
	return updates, nil
}

func hasExplicitActivityLocation(req types.ActivityCreateRequest) bool {
	return req.Address != nil || req.Latitude != nil || req.Longitude != nil
}

func validateExplicitActivityLocation(req types.ActivityCreateRequest) error {
	if req.Address == nil || strings.TrimSpace(*req.Address) == "" {
		return errors.New("自定义活动地址不能为空")
	}
	if req.Latitude == nil || req.Longitude == nil {
		return errors.New("自定义活动地址必须同时提供经纬度")
	}
	return validateChinaCoordinate(*req.Latitude, *req.Longitude)
}

// activityEditNeedsReaudit deliberately treats every writable content field
// (and a tag or business-hours change) as publishable. This keeps the rule
// simple for merchants and ensures review does not miss a material change.
func activityEditNeedsReaudit(req types.ActivityCreateRequest, updates map[string]any) bool {
	return len(updates) > 0 || req.TagIDs != nil || req.BusinessHours != nil || req.TicketSpecs != nil
}

func decodeActivityRevision(activity models.Activity) (*types.ActivityRevisionPayload, error) {
	if strings.TrimSpace(activity.PendingRevision) == "" {
		return nil, nil
	}
	var payload types.ActivityRevisionPayload
	if err := json.Unmarshal([]byte(activity.PendingRevision), &payload); err != nil {
		return nil, fmt.Errorf("待审核修改数据损坏: %w", err)
	}
	payload.Activity.ID = activity.ID
	payload.Activity.OrganizerID = activity.OrganizerID
	return &payload, nil
}

func (s *TicketingService) activityRevisionSnapshot(ctx context.Context, tx *gorm.DB, activity models.Activity) (*types.ActivityRevisionPayload, error) {
	if payload, err := decodeActivityRevision(activity); err != nil || payload == nil {
		if err != nil {
			return nil, err
		}
	} else {
		return payload, nil
	}

	payload := &types.ActivityRevisionPayload{Activity: activity, TicketSpecs: []models.TicketSpec{}, TagIDs: []int64{}}
	payload.Activity.PendingRevision = ""
	payload.Activity.PendingRevisionStatus = 0
	payload.Activity.PendingRevisionReason = ""
	if defaultActivityType(activity.Type) != models.ActivityTypeVenue {
		if err := tx.WithContext(ctx).Where("activity_id = ?", activity.ID).Order("id ASC").Find(&payload.TicketSpecs).Error; err != nil {
			return nil, err
		}
	}
	targetType, targetID := models.ContentTagTargetActivity, activity.ID
	if defaultActivityType(activity.Type) == models.ActivityTypeVenue {
		targetType, targetID = models.ContentTagTargetVenue, activity.OrganizerID
	}
	tags, err := dao.LoadContentTags(ctx, tx, targetType, []int64{targetID}, false)
	if err != nil {
		return nil, err
	}
	payload.TagIDs = types.ContentTagIDs(tags[targetID])
	if defaultActivityType(activity.Type) == models.ActivityTypeVenue {
		var profile models.OrganizerProfile
		if err := tx.WithContext(ctx).Where("organizer_id = ?", activity.OrganizerID).First(&profile).Error; err == nil {
			payload.BusinessHours = profile.BusinessHours
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	return payload, nil
}

func applyActivityUpdateSnapshot(activity *models.Activity, updates map[string]any) error {
	encoded, err := json.Marshal(updates)
	if err != nil {
		return err
	}
	return json.Unmarshal(encoded, activity)
}

func applyRevisionTicketSpecs(activityID int64, current []models.TicketSpec, items []types.TicketSpecSaveItem) ([]models.TicketSpec, error) {
	byID := make(map[int64]int, len(current))
	for i := range current {
		byID[current[i].ID] = i
	}
	for _, item := range items {
		spec, err := buildTicketSpec(activityID, item)
		if err != nil {
			return nil, err
		}
		if item.ID > 0 {
			index, ok := byID[item.ID]
			if !ok {
				return nil, errors.New("票券不存在")
			}
			spec.ID = item.ID
			spec.SoldCount = current[index].SoldCount
			current[index] = *spec
			continue
		}
		current = append(current, *spec)
	}
	return current, nil
}

// saveActivityRevision stages changes for an already-online activity. The
// public row is intentionally untouched until the administrator approves it.
func (s *TicketingService) saveActivityRevision(ctx context.Context, organizerID int64, activity *models.Activity, req types.ActivityCreateRequest, updates map[string]any) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked models.Activity
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND organizer_id = ? AND status = ?", activity.ID, organizerID, models.ActivityStatusOnline).First(&locked).Error; err != nil {
			return err
		}
		payload, err := s.activityRevisionSnapshot(ctx, tx, locked)
		if err != nil {
			return err
		}
		if err := applyActivityUpdateSnapshot(&payload.Activity, updates); err != nil {
			return err
		}
		payload.Activity.ID = locked.ID
		payload.Activity.OrganizerID = locked.OrganizerID
		payload.Activity.Status = models.ActivityStatusPending
		payload.Activity.AuditType = models.ActivityAuditTypeReaudit
		payload.Activity.RejectReason = ""
		if req.TicketSpecs != nil {
			payload.TicketSpecs, err = applyRevisionTicketSpecs(locked.ID, payload.TicketSpecs, req.TicketSpecs)
			if err != nil {
				return err
			}
		}
		if req.TagIDs != nil {
			payload.TagIDs = append([]int64(nil), req.TagIDs...)
		}
		if req.BusinessHours != nil {
			payload.BusinessHours = *req.BusinessHours
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		return tx.Model(&locked).Updates(map[string]any{
			"pending_revision":        string(encoded),
			"pending_revision_status": models.ActivityStatusPending,
			"pending_revision_reason": "",
			"updated_at":              time.Now(),
		}).Error
	})
}

func (s *TicketingService) deleteRevisionTicketSpec(ctx context.Context, organizerID int64, activity *models.Activity, specID int64) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked models.Activity
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND organizer_id = ? AND status = ?", activity.ID, organizerID, models.ActivityStatusOnline).First(&locked).Error; err != nil {
			return err
		}
		payload, err := s.activityRevisionSnapshot(ctx, tx, locked)
		if err != nil {
			return err
		}
		index := -1
		for i := range payload.TicketSpecs {
			if payload.TicketSpecs[i].ID == specID {
				index = i
				break
			}
		}
		if index < 0 {
			return gorm.ErrRecordNotFound
		}
		payload.TicketSpecs = append(payload.TicketSpecs[:index], payload.TicketSpecs[index+1:]...)
		payload.Activity.Status = models.ActivityStatusPending
		payload.Activity.AuditType = models.ActivityAuditTypeReaudit
		encoded, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		return tx.Model(&locked).Updates(map[string]any{
			"pending_revision":        string(encoded),
			"pending_revision_status": models.ActivityStatusPending,
			"pending_revision_reason": "",
			"updated_at":              time.Now(),
		}).Error
	})
}

// markActivityPendingReview is idempotent: it only transitions an online
// activity. Drafts, rejected records and entries already awaiting review keep
// their existing workflow state.
func (s *TicketingService) markActivityPendingReview(ctx context.Context, activityID int64, auditType string) (bool, error) {
	auditType = normalizeActivityAuditType(auditType)
	result := s.DB.WithContext(ctx).Model(&models.Activity{}).
		Where("id = ? AND status = ?", activityID, models.ActivityStatusOnline).
		Updates(map[string]any{
			"status":        models.ActivityStatusPending,
			"audit_type":    auditType,
			"reject_reason": "",
			"updated_at":    time.Now(),
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func normalizeActivityAuditType(auditType string) string {
	if auditType == models.ActivityAuditTypeReaudit {
		return models.ActivityAuditTypeReaudit
	}
	return models.ActivityAuditTypeInitial
}

// venueValidityWindow only preserves compatibility with the activities table.
// Venue availability is governed by organizer business_hours, not this range.
func venueValidityWindow(now time.Time) (time.Time, time.Time) {
	return now, now.AddDate(20, 0, 0)
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

// normalizeOrganizerType keeps historical merchant records compatible while
// allowing the application UI to use the clearer party/venue wording.
func normalizeOrganizerType(raw string) (string, error) {
	switch strings.TrimSpace(raw) {
	case "", models.OrganizerTypeMerchant, models.ActivityTypeParty:
		return models.OrganizerTypeMerchant, nil
	case models.OrganizerTypeVenue:
		return models.OrganizerTypeVenue, nil
	default:
		return "", errors.New("入驻类型无效，仅支持 party 或 venue")
	}
}

// normalizeMarkerIcon keeps the database as a passive URL store while
// preventing arbitrary external hosts from being rendered on the map.
func normalizeMarkerIcon(raw string) (string, error) {
	markerIcon := strings.TrimSpace(raw)
	if markerIcon == "" {
		return "", nil
	}
	if len(markerIcon) > 255 {
		return "", errors.New("地图图标地址过长")
	}
	parsed, err := url.Parse(markerIcon)
	if err != nil || parsed.Scheme != "https" || parsed.Host != "cdn.hypercn.cn" || parsed.Path == "" {
		return "", errors.New("地图图标必须是 cdn.hypercn.cn 的 HTTPS 地址")
	}
	return markerIcon, nil
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

func ticketSpecUpdates(spec *models.TicketSpec) map[string]any {
	return map[string]any{
		"name":           spec.Name,
		"description":    spec.Description,
		"is_enabled":     spec.IsEnabled,
		"sale_start":     spec.SaleStart,
		"sale_end":       spec.SaleEnd,
		"price":          spec.Price,
		"stock":          spec.Stock,
		"purchase_limit": spec.PurchaseLimit,
		"max_attendees":  spec.MaxAttendees,
		"updated_at":     time.Now(),
	}
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
