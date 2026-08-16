package service

import (
	"Hyper/dao"
	"Hyper/models"
	"Hyper/pkg/encrypt"
	"Hyper/pkg/jwt"
	"Hyper/types"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IAdminService interface {
	Login(ctx context.Context, username, password string) (string, string, error)
	GetOrganizerList(ctx context.Context, page, pageSize int, status *int8, organizerType string) (*types.AdminOrganizerListResponse, error)
	GetOrganizerDetail(ctx context.Context, organizerID int64) (*types.AdminOrganizerDetail, error)
	AuditOrganizer(ctx context.Context, organizerID int64, req types.AdminAuditOrganizerRequest) error
	DeleteOrganizer(ctx context.Context, organizerID int64) error
	BindAdminWechatSubscriber(ctx context.Context, adminID int64, code string) error
	GetPartyList(ctx context.Context, page, pageSize int, keyword, partyType string) (*types.AdminPartyListResponse, error)
	GetPartyDetail(ctx context.Context, partyID int64) (*types.AdminPartyDetail, error)
	UpdatePartyStatus(ctx context.Context, partyID int64, status string) error
	GetActivityList(ctx context.Context, page, pageSize int, filter types.AdminActivityFilter) (*types.AdminActivityListResponse, error)
	UpdateOrganizerEnabled(ctx context.Context, organizerID int64, enabled int8) error
	GetActivityDetail(ctx context.Context, activityID int64) (*types.AdminActivityDetail, error)
	AuditActivity(ctx context.Context, activityID int64, req types.AdminAuditActivityRequest) error
	SetActivityVisibility(ctx context.Context, activityID int64, visible bool, reason string) (*models.Activity, error)
	GetEventTicketList(ctx context.Context, eventID int64, page, pageSize int) (*types.AdminTicketListResponse, error)
	GetAllTickets(ctx context.Context, page, pageSize int, keyword string) (*types.AdminTicketListResponse, error)
	GetOrderList(ctx context.Context, page, pageSize int, eventID int64) (*types.AdminOrderListResponse, error)
	GetTicketOrderList(ctx context.Context, page, pageSize int, activityID int64, status, refundStatus *int8, keyword, salesChannel string) (*types.AdminTicketOrderListResponse, error)
	GetTicketOrderDetail(ctx context.Context, orderNo string) (*types.AdminTicketOrderDetail, error)
	GetRefundDetail(ctx context.Context, refundNo string) (*types.AdminRefundDetail, error)
	ApproveOrderRefund(ctx context.Context, orderNo string) error
	RejectOrderRefund(ctx context.Context, orderNo string, reason string) error
	GetFinanceSummary(ctx context.Context) (*types.AdminFinanceSummary, error)
	GetFinanceSettlements(ctx context.Context, page, pageSize int, organizerID int64) (*types.AdminSettlementListResponse, error)
	ListPlatformFinanceFlows(ctx context.Context, page, pageSize int, filter types.AdminPlatformFlowFilter) (*types.AdminPlatformFlowListResponse, error)
	GetUserList(ctx context.Context, page, pageSize int, keyword string) (*types.AdminUserListResponse, error)
	UpdateUserStatus(ctx context.Context, userID int, status int8) error
	ListBanners(ctx context.Context) ([]models.PlatformBanner, error)
	CreateBanner(ctx context.Context, req types.AdminBannerRequest) (int64, error)
	UpdateBanner(ctx context.Context, id int64, req types.AdminBannerRequest) error
	DeleteBanner(ctx context.Context, id int64) error
	SortBanners(ctx context.Context, req types.AdminBannerSortRequest) error
	GetSettings(ctx context.Context) ([]types.AdminSettingItem, error)
	UpdateSettings(ctx context.Context, settings []types.AdminSettingItem) error
	GetSystemConfig(ctx context.Context) (*types.AdminSystemConfig, error)
	UpdateSystemConfig(ctx context.Context, config types.AdminSystemConfig) error
	GetDashboardStats(ctx context.Context) (*types.AdminDashboardStats, error)
	GetAdminProfile(ctx context.Context, adminID int64) (*types.AdminProfileResponse, error)
	UpdateAdminProfile(ctx context.Context, adminID int64, req types.AdminProfileRequest) error
	UpdateAdminPassword(ctx context.Context, adminID int64, req types.AdminPasswordRequest) error
	ListAdmins(ctx context.Context, page, pageSize int, keyword string) (*types.AdminPageResponse[types.AdminAccountItem], error)
	CreateAdmin(ctx context.Context, req types.AdminAccountRequest) (int64, error)
	UpdateAdmin(ctx context.Context, actorID, id int64, req types.AdminAccountRequest) error
	DeleteAdmin(ctx context.Context, actorID, id int64) error
	ListRoles(ctx context.Context, page, pageSize int, keyword string) (*types.AdminPageResponse[types.AdminRoleItem], error)
	SaveRole(ctx context.Context, id int64, req types.AdminRoleRequest) (int64, error)
	DeleteRole(ctx context.Context, id int64) error
	ListOperationLogs(ctx context.Context, page, pageSize int, filter types.AdminOperationLogFilter) (*types.AdminPageResponse[types.AdminOperationLogItem], error)
	RecordOperationLog(ctx context.Context, log models.AdminOperationLog) error
	CheckPermission(ctx context.Context, adminID int64, method, path string) error
	ListCategories(ctx context.Context, page, pageSize int, categoryType string) (*types.AdminPageResponse[models.AdminCategory], error)
	SaveCategory(ctx context.Context, id int64, req types.AdminCategoryRequest) (int64, error)
	DeleteCategory(ctx context.Context, id int64) error
	SaveContentTags(ctx context.Context, targetType string, targetID int64, tagIDs []int64) error
	ListUserRecords(ctx context.Context, userID int64, recordType string, page, pageSize int) (*types.AdminPageResponse[map[string]any], error)
	ListViewers(ctx context.Context, page, pageSize int, keyword string) (*types.AdminPageResponse[map[string]any], error)
	ListActivityCollections(ctx context.Context, page, pageSize int, keyword string, organizerID int64) (*types.AdminPageResponse[types.ActivityCollectionItem], error)
	SaveActivityCollection(ctx context.Context, id int64, req types.ActivityCollectionRequest) (int64, error)
	DeleteActivityCollection(ctx context.Context, id int64) error
	ListVerifiers(ctx context.Context, page, pageSize int, keyword string, organizerID int64) (*types.AdminPageResponse[map[string]any], error)
	SaveVerifier(ctx context.Context, id int64, req types.AdminVerifierRequest) (int64, error)
	UpdateVerifierStatus(ctx context.Context, id int64, status int8) error
	DeleteVerifier(ctx context.Context, id int64) error
	ListVerificationRecords(ctx context.Context, page, pageSize int, keyword string, organizerID int64) (*types.AdminPageResponse[map[string]any], error)
	ListNotes(ctx context.Context, page, pageSize int, status *int, keyword string) (*types.AdminPageResponse[map[string]any], error)
	UpdateNoteStatus(ctx context.Context, noteID int64, status int) error
	ListNoteRecords(ctx context.Context, recordType string, page, pageSize int, noteID int64) (*types.AdminPageResponse[map[string]any], error)
	ListNoteInteractions(ctx context.Context, page, pageSize int, filter types.AdminNoteInteractionFilter) (*types.AdminPageResponse[map[string]any], error)
	ListNoteComments(ctx context.Context, page, pageSize int, filter types.AdminNoteCommentFilter) (*types.AdminPageResponse[map[string]any], error)
	UpdateCommentStatus(ctx context.Context, noteID, commentID, adminID int64, status int8, reason string) error
	ListPointLogs(ctx context.Context, page, pageSize int, userID int64) (*types.AdminPageResponse[map[string]any], error)
	AdjustPoints(ctx context.Context, adminID int64, req types.AdminPointsAdjustRequest) (*types.AdminPointsAdjustResponse, error)
	ListWithdraws(ctx context.Context, page, pageSize int, status *int8, organizerID int64) (*types.AdminPageResponse[map[string]any], error)
	AuditWithdraw(ctx context.Context, id int64, req types.WithdrawAuditRequest) error
	ListBankAccountAudits(ctx context.Context, page, pageSize int, status *int8, organizerID int64) (*types.AdminPageResponse[map[string]any], error)
	AuditBankAccount(ctx context.Context, id int64, req types.BankAccountAuditRequest) error
	ListOrganizerLevelRules(ctx context.Context) ([]models.OrganizerLevelRule, error)
	SaveOrganizerLevelRule(ctx context.Context, id int64, req types.OrganizerLevelRuleRequest) (int64, error)
	DeleteOrganizerLevelRule(ctx context.Context, id int64) error
	ListMessages(ctx context.Context, page, pageSize int, target, messageType string) (*types.AdminPageResponse[map[string]any], error)
	ListMessageDeliveries(ctx context.Context, messageID int64, page, pageSize int, filter types.AdminMessageDeliveryFilter) (*types.AdminPageResponse[map[string]any], error)
	CreateMessage(ctx context.Context, adminID int64, req types.PlatformMessageRequest) (int64, error)
	ListCustomerServiceSessions(ctx context.Context, page, pageSize int, keyword string) (*types.AdminCustomerServiceSessionListResponse, error)
	ListCustomerServiceMessages(ctx context.Context, customerID uint64, cursor, since int64, limit int) (*types.AdminCustomerServiceMessageListResponse, error)
	SendCustomerServiceMessage(ctx context.Context, customerID uint64, req types.AdminCustomerServiceSendMessageRequest) (*types.Message, error)
	MarkCustomerServiceSessionRead(ctx context.Context, customerID uint64, readTime int64) error
	GetPointsRule(ctx context.Context) (*types.PointsRule, error)
	UpdatePointsRule(ctx context.Context, req types.UpdatePointsRuleRequest) error
}

type AdminService struct {
	AdminDAO       *dao.Admin
	DB             *gorm.DB
	Secret         []byte
	WeChatService  IWeChatService
	MqProducer     rmq_client.Producer
	MessageService IMessageService
}

var _ IAdminService = (*AdminService)(nil)

func (s *AdminService) Login(ctx context.Context, username, password string) (string, string, error) {
	admin, err := s.AdminDAO.FindByWhere(ctx, "username = ?", username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", errors.New("管理员账号不存在")
		}
		return "", "", err
	}

	if admin.Status == models.AdminStatusDeactivate {
		return "", "", errors.New("该管理员账号已停用")
	}

	if !encrypt.VerifyPassword(admin.Password, password) {
		return "", "", errors.New("密码错误")
	}

	accessToken, err := jwt.GenerateToken(s.Secret, uint(admin.Id), "admin", "access", 2*time.Hour)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := jwt.GenerateToken(s.Secret, uint(admin.Id), "admin", "refresh", 7*24*time.Hour)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AdminService) GetOrganizerList(ctx context.Context, page, pageSize int, status *int8, organizerType string) (*types.AdminOrganizerListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	query := s.DB.WithContext(ctx).Model(&models.Organizer{})
	if status != nil {
		if *status == models.OrganizerStatusAuditing {
			// An approved venue can also have a profile revision waiting for
			// review. Surface it in the same audit queue without taking the
			// live venue offline.
			query = query.Where("status = ? OR pending_profile_status = ?", *status, models.OrganizerStatusAuditing)
		} else {
			query = query.Where("status = ?", *status)
		}
	}
	if organizerType != "" {
		if organizerType != models.OrganizerTypeVenue && organizerType != models.OrganizerTypeMerchant {
			return nil, errors.New("入驻类型无效，仅支持 venue 或 merchant")
		}
		query = query.Where("type = ?", organizerType)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var organizers []models.Organizer
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&organizers).Error; err != nil {
		return nil, err
	}
	userIDs := make([]int, 0, len(organizers))
	for _, org := range organizers {
		userIDs = append(userIDs, int(org.UserID))
	}
	userMap := make(map[int]*models.Users)
	if len(userIDs) > 0 {
		var users []models.Users
		s.DB.WithContext(ctx).Where("id IN ?", userIDs).Find(&users)
		for i := range users {
			userMap[users[i].Id] = &users[i]
		}
	}
	list := make([]types.AdminOrganizerItem, 0, len(organizers))
	for _, org := range organizers {
		item := types.AdminOrganizerItem{
			ID:                        org.ID,
			UserID:                    org.UserID,
			Type:                      org.Type,
			Name:                      org.Name,
			Logo:                      org.Logo,
			Status:                    org.Status,
			Enabled:                   org.Enabled,
			RejectReason:              org.RejectReason,
			AuditKind:                 "initial",
			HasPendingProfileRevision: org.PendingProfileStatus == models.OrganizerStatusAuditing || org.PendingProfileStatus == models.OrganizerStatusRejected,
			PendingProfileReason:      org.PendingProfileReason,
			Level:                     org.Level,
			ServiceFeeRate:            org.ServiceFeeRate,
			Province:                  org.Province,
			City:                      org.City,
			District:                  org.District,
			CreatedAt:                 org.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:                 org.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		if org.Status == models.OrganizerStatusApproved && org.PendingProfileStatus == models.OrganizerStatusAuditing {
			item.AuditKind = "profile_revision"
		}
		if u, ok := userMap[int(org.UserID)]; ok {
			item.UserName = u.Nickname
			item.UserAvatar = u.Avatar
			item.UserMobile = u.Mobile
		}
		list = append(list, item)
	}
	return &types.AdminOrganizerListResponse{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *AdminService) UpdateOrganizerEnabled(ctx context.Context, organizerID int64, enabled int8) error {
	if enabled != 0 && enabled != 1 {
		return errors.New("商家启停状态仅支持 0 或 1")
	}
	var org models.Organizer
	if err := s.DB.WithContext(ctx).Where("id = ?", organizerID).First(&org).Error; err != nil {
		return err
	}
	if enabled == 1 && org.Status != models.OrganizerStatusApproved {
		return errors.New("只有审核通过的商家可以启用")
	}
	result := s.DB.WithContext(ctx).Model(&org).Updates(map[string]any{
		"enabled":    enabled,
		"updated_at": time.Now(),
	})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *AdminService) GetOrganizerDetail(ctx context.Context, organizerID int64) (*types.AdminOrganizerDetail, error) {
	var org models.Organizer
	if err := s.DB.WithContext(ctx).Where("id = ?", organizerID).First(&org).Error; err != nil {
		return nil, err
	}
	detail := &types.AdminOrganizerDetail{Organizer: org}
	if org.PendingProfileStatus == models.OrganizerStatusAuditing || org.PendingProfileStatus == models.OrganizerStatusRejected {
		revision, err := decodeOrganizerVenueProfileRevision(org)
		if err != nil {
			return nil, err
		}
		detail.PendingProfileRevision = revision
	}
	var user models.Users
	if err := s.DB.WithContext(ctx).Where("id = ?", org.UserID).First(&user).Error; err == nil {
		detail.UserName = user.Nickname
		detail.UserAvatar = user.Avatar
		detail.UserMobile = user.Mobile
	}
	tagMap, err := dao.LoadContentTags(ctx, s.DB, models.ContentTagTargetVenue, []int64{organizerID}, true)
	if err != nil {
		return nil, err
	}
	detail.TagIDs = types.ContentTagIDs(tagMap[organizerID])
	detail.Tags = types.BuildContentTagItems(tagMap[organizerID])
	return detail, nil
}

func (s *AdminService) AuditOrganizer(ctx context.Context, organizerID int64, req types.AdminAuditOrganizerRequest) error {
	if req.Status != models.OrganizerStatusApproved && req.Status != models.OrganizerStatusRejected {
		return errors.New("审核状态无效，仅支持 2通过 或 3拒绝")
	}
	if req.Status == models.OrganizerStatusRejected && req.RejectReason == "" {
		return errors.New("拒绝时必须填写 reject_reason")
	}
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var org models.Organizer
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", organizerID).First(&org).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("入驻申请不存在")
			}
			return err
		}

		if org.Status == models.OrganizerStatusApproved && org.PendingProfileStatus == models.OrganizerStatusAuditing {
			return s.auditVenueProfileRevision(tx, org, req)
		}

		updates := map[string]any{"status": req.Status, "reject_reason": "", "updated_at": time.Now()}
		if req.Status == models.OrganizerStatusApproved {
			if err := ensureDefaultOrganizerLevelRules(tx); err != nil {
				return err
			}
			level, feeRate, _, err := organizerLevelByCompletedCount(tx, 0)
			if err != nil {
				return err
			}
			updates["enabled"] = 1
			updates["level"] = fmt.Sprintf("LV%d", level)
			updates["service_fee_rate"] = feeRate
		}
		if req.Status == models.OrganizerStatusRejected {
			updates["reject_reason"] = req.RejectReason
		}
		return tx.Model(&models.Organizer{}).Where("id = ?", organizerID).Updates(updates).Error
	})
}

func (s *AdminService) auditVenueProfileRevision(tx *gorm.DB, org models.Organizer, req types.AdminAuditOrganizerRequest) error {
	if req.Status == models.OrganizerStatusRejected {
		return tx.Model(&models.Organizer{}).Where("id = ?", org.ID).Updates(map[string]any{
			"pending_profile_status": models.OrganizerStatusRejected,
			"pending_profile_reason": req.RejectReason,
			"updated_at":             time.Now(),
		}).Error
	}

	revision, err := decodeOrganizerVenueProfileRevision(org)
	if err != nil {
		return err
	}
	if revision == nil {
		return errors.New("场地资料待审核数据不存在")
	}
	if err := upsertOrganizerVenueProfile(tx, org.ID, revision); err != nil {
		return err
	}
	return tx.Model(&models.Organizer{}).Where("id = ?", org.ID).Updates(map[string]any{
		"name": revision.Name, "logo": revision.Logo, "marker_icon": revision.MarkerIcon,
		"province": revision.Province, "city": revision.City, "district": revision.District,
		"pending_profile_revision": "", "pending_profile_status": 0, "pending_profile_reason": "", "updated_at": time.Now(),
	}).Error
}

func (s *AdminService) DeleteOrganizer(ctx context.Context, organizerID int64) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var org models.Organizer
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", organizerID).First(&org).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Activity{}).
			Where("organizer_id = ? AND status IN ?", organizerID, []int8{
				models.ActivityStatusDraft,
				models.ActivityStatusPending,
				models.ActivityStatusAuditing,
				models.ActivityStatusOnline,
			}).
			Updates(map[string]any{
				"status":        models.ActivityStatusRejected,
				"reject_reason": "主办方已被管理员删除",
				"updated_at":    time.Now(),
			}).Error; err != nil {
			return err
		}
		if err := tx.Where("organizer_id = ?", organizerID).Delete(&models.Verifier{}).Error; err != nil {
			return err
		}
		if err := tx.Where("organizer_id = ?", organizerID).Delete(&models.OrganizerStore{}).Error; err != nil {
			return err
		}
		var collections []models.ActivityCollection
		if err := tx.Where("organizer_id = ?", organizerID).Find(&collections).Error; err != nil {
			return err
		}
		collectionIDs := make([]int64, 0, len(collections))
		for _, collection := range collections {
			collectionIDs = append(collectionIDs, collection.ID)
		}
		if len(collectionIDs) > 0 {
			if err := tx.Where("collection_id IN ?", collectionIDs).Delete(&models.ActivityCollectionItem{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("organizer_id = ?", organizerID).Delete(&models.ActivityCollection{}).Error; err != nil {
			return err
		}
		result := tx.Delete(&org)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (s *AdminService) BindAdminWechatSubscriber(ctx context.Context, adminID int64, code string) error {
	if s.WeChatService == nil {
		return errors.New("微信服务未初始化")
	}
	wxResp, err := s.WeChatService.Code2Session(ctx, code)
	if err != nil {
		return err
	}
	if wxResp.OpenID == "" {
		return errors.New("微信 openid 为空")
	}
	record := models.AdminWechatSubscriber{
		AdminID: adminID,
		OpenID:  wxResp.OpenID,
		Enabled: 1,
	}
	return s.DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "admin_id"}, {Name: "open_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"enabled":    1,
			"updated_at": time.Now(),
		}),
	}).Create(&record).Error
}

func (s *AdminService) GetPartyList(ctx context.Context, page, pageSize int, keyword, partyType string) (*types.AdminPartyListResponse, error) {
	query := s.DB.WithContext(ctx).Model(&models.Merchant{})

	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR location_name LIKE ?", like, like)
	}
	if partyType != "" {
		query = query.Where("type = ?", partyType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var parties []models.Merchant
	if err := query.
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&parties).Error; err != nil {
		return nil, err
	}

	// 批量获取用户信息
	userIDs := make([]int, 0, len(parties))
	for _, p := range parties {
		userIDs = append(userIDs, p.UserID)
	}
	userMap := make(map[int]*models.Users)
	if len(userIDs) > 0 {
		var users []models.Users
		s.DB.WithContext(ctx).Where("id IN ?", userIDs).Find(&users)
		for i := range users {
			userMap[users[i].Id] = &users[i]
		}
	}

	list := make([]types.AdminPartyItem, 0, len(parties))
	for _, p := range parties {
		item := types.AdminPartyItem{
			ID:           p.ID,
			UserID:       p.UserID,
			Title:        p.Title,
			Type:         p.Type,
			Status:       p.Status,
			LocationName: p.LocationName,
			Address:      p.Address,
			CoverImage:   p.CoverImage,
			Category:     p.Category,
			CreatedAt:    p.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    p.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		if u, ok := userMap[p.UserID]; ok {
			item.UserName = u.Nickname
			item.UserAvatar = u.Avatar
			item.UserMobile = u.Mobile
		}
		list = append(list, item)
	}

	return &types.AdminPartyListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *AdminService) GetPartyDetail(ctx context.Context, partyID int64) (*types.AdminPartyDetail, error) {
	var party models.Merchant
	if err := s.DB.WithContext(ctx).Where("id = ?", partyID).First(&party).Error; err != nil {
		return nil, err
	}

	detail := &types.AdminPartyDetail{Merchant: party}
	var user models.Users
	if err := s.DB.WithContext(ctx).Where("id = ?", party.UserID).First(&user).Error; err == nil {
		detail.UserName = user.Nickname
		detail.UserAvatar = user.Avatar
		detail.UserMobile = user.Mobile
	}
	tagMap, err := dao.LoadContentTags(ctx, s.DB, models.ContentTagTargetParty, []int64{partyID}, true)
	if err != nil {
		return nil, err
	}
	detail.TagIDs = types.ContentTagIDs(tagMap[partyID])
	detail.Tags = types.BuildContentTagItems(tagMap[partyID])

	return detail, nil
}

func (s *AdminService) UpdatePartyStatus(ctx context.Context, partyID int64, status string) error {
	if status != "active" && status != "offline" {
		return errors.New("状态值无效，仅支持 active 或 offline")
	}
	result := s.DB.WithContext(ctx).
		Model(&models.Merchant{}).
		Where("id = ?", partyID).
		Updates(map[string]any{
			"status":     status,
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("派对不存在")
	}
	return nil
}

func (s *AdminService) GetActivityList(ctx context.Context, page, pageSize int, filter types.AdminActivityFilter) (*types.AdminActivityListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := s.DB.WithContext(ctx).Model(&models.Activity{})
	if filter.Status != nil {
		if *filter.Status == models.ActivityStatusPending {
			query = query.Where("status = ? OR (status = ? AND pending_revision_status = ?)", models.ActivityStatusPending, models.ActivityStatusOnline, models.ActivityStatusPending)
		} else if *filter.Status == models.ActivityStatusOnline {
			query = query.Where("status = ? AND pending_revision_status <> ?", models.ActivityStatusOnline, models.ActivityStatusPending)
		} else {
			query = query.Where("status = ?", *filter.Status)
		}
	} else {
		query = query.Where("status <> ?", models.ActivityStatusDraft)
	}
	if filter.AuditType != "" {
		if filter.AuditType == models.ActivityAuditTypeReaudit {
			query = query.Where("audit_type = ? OR pending_revision_status = ?", filter.AuditType, models.ActivityStatusPending)
		} else {
			query = query.Where("audit_type = ? AND pending_revision_status <> ?", filter.AuditType, models.ActivityStatusPending)
		}
	}
	if filter.IsHidden != nil {
		query = query.Where("is_hidden = ?", *filter.IsHidden)
	}
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		query = query.Where("name LIKE ? OR address LIKE ?", like, like)
	}
	if filter.OrganizerID > 0 {
		query = query.Where("organizer_id = ?", filter.OrganizerID)
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

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var activities []models.Activity
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&activities).Error; err != nil {
		return nil, err
	}
	revisionTicketSpecCounts := make(map[int64]int)
	for i := range activities {
		if activities[i].PendingRevisionStatus != models.ActivityStatusPending {
			continue
		}
		payload, err := decodeActivityRevision(activities[i])
		if err != nil {
			return nil, err
		}
		if payload == nil {
			continue
		}
		candidate := payload.Activity
		candidate.ID = activities[i].ID
		candidate.OrganizerID = activities[i].OrganizerID
		candidate.Status = models.ActivityStatusPending
		candidate.AuditType = models.ActivityAuditTypeReaudit
		candidate.PendingRevisionStatus = models.ActivityStatusPending
		activities[i] = candidate
		revisionTicketSpecCounts[candidate.ID] = len(payload.TicketSpecs)
	}

	organizerIDs := make([]int64, 0, len(activities))
	activityIDs := make([]int64, 0, len(activities))
	for _, activity := range activities {
		organizerIDs = append(organizerIDs, activity.OrganizerID)
		activityIDs = append(activityIDs, activity.ID)
	}

	organizerMap := make(map[int64]models.Organizer)
	if len(organizerIDs) > 0 {
		var organizers []models.Organizer
		if err := s.DB.WithContext(ctx).Where("id IN ?", organizerIDs).Find(&organizers).Error; err != nil {
			return nil, err
		}
		for _, organizer := range organizers {
			organizerMap[organizer.ID] = organizer
		}
	}

	ticketSpecCountMap := make(map[int64]int)
	if len(activityIDs) > 0 {
		var rows []struct {
			ActivityID int64
			Count      int
		}
		if err := s.DB.WithContext(ctx).Model(&models.TicketSpec{}).
			Select("activity_id, COUNT(*) AS count").
			Where("activity_id IN ?", activityIDs).
			Group("activity_id").
			Scan(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			ticketSpecCountMap[row.ActivityID] = row.Count
		}
	}

	list := make([]types.AdminActivityItem, 0, len(activities))
	for _, activity := range activities {
		item := types.AdminActivityItem{
			ID:               activity.ID,
			OrganizerID:      activity.OrganizerID,
			Type:             activity.Type,
			Name:             activity.Name,
			ShareTitle:       activity.ShareTitle,
			StartTime:        formatAdminTime(activity.StartTime),
			EndTime:          formatAdminTime(activity.EndTime),
			RealNameMode:     activity.RealNameMode,
			MinorCheck:       activity.MinorCheck,
			Description:      activity.Description,
			Province:         activity.Province,
			City:             activity.City,
			District:         activity.District,
			Address:          activity.Address,
			Latitude:         activity.Latitude,
			Longitude:        activity.Longitude,
			PosterDetail:     activity.PosterDetail,
			PosterLong:       activity.PosterLong,
			PosterList:       activity.PosterList,
			PosterWechat:     activity.PosterWechat,
			QualificationDoc: activity.QualificationDoc,
			Status:           activity.Status,
			AuditType:        activity.AuditType,
			RejectReason:     activity.RejectReason,
			IsHidden:         activity.IsHidden,
			HiddenAt:         formatAdminPtrTime(activity.HiddenAt),
			HiddenReason:     activity.HiddenReason,
			TicketSpecCount:  ticketSpecCountMap[activity.ID],
			CreatedAt:        activity.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:        activity.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		if count, ok := revisionTicketSpecCounts[activity.ID]; ok {
			item.TicketSpecCount = count
		}
		if organizer, ok := organizerMap[activity.OrganizerID]; ok {
			item.OrganizerName = organizer.Name
			item.OrganizerType = organizer.Type
		}
		list = append(list, item)
	}

	return &types.AdminActivityListResponse{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *AdminService) GetActivityDetail(ctx context.Context, activityID int64) (*types.AdminActivityDetail, error) {
	var activity models.Activity
	if err := s.DB.WithContext(ctx).Where("id = ?", activityID).First(&activity).Error; err != nil {
		return nil, err
	}

	payload, err := decodeActivityRevision(activity)
	if err != nil {
		return nil, err
	}
	showPendingRevision := payload != nil && activity.PendingRevisionStatus == models.ActivityStatusPending
	if showPendingRevision {
		candidate := payload.Activity
		candidate.ID = activity.ID
		candidate.OrganizerID = activity.OrganizerID
		candidate.Status = models.ActivityStatusPending
		candidate.AuditType = models.ActivityAuditTypeReaudit
		candidate.PendingRevisionStatus = models.ActivityStatusPending
		activity = candidate
	}
	detail := &types.AdminActivityDetail{Activity: activity}
	var organizer models.Organizer
	if err := s.DB.WithContext(ctx).Where("id = ?", activity.OrganizerID).First(&organizer).Error; err == nil {
		detail.Organizer = &organizer
	}
	if showPendingRevision {
		detail.TicketSpecs = payload.TicketSpecs
	} else if err := s.DB.WithContext(ctx).Where("activity_id = ?", activity.ID).Order("id ASC").Find(&detail.TicketSpecs).Error; err != nil {
		return nil, err
	}
	targetType, targetID := models.ContentTagTargetActivity, activityID
	if activity.Type == models.ActivityTypeVenue {
		targetType, targetID = models.ContentTagTargetVenue, activity.OrganizerID
	}
	tagMap, err := dao.LoadContentTags(ctx, s.DB, targetType, []int64{targetID}, true)
	if err != nil {
		return nil, err
	}
	detail.TagIDs = types.ContentTagIDs(tagMap[targetID])
	detail.Tags = types.BuildContentTagItems(tagMap[targetID])
	if showPendingRevision {
		detail.TagIDs = payload.TagIDs
	}
	return detail, nil
}

func (s *AdminService) SaveContentTags(ctx context.Context, targetType string, targetID int64, tagIDs []int64) error {
	if targetID <= 0 {
		return errors.New("标签绑定目标无效")
	}
	var count int64
	switch targetType {
	case models.ContentTagTargetActivity:
		var activity models.Activity
		if err := s.DB.WithContext(ctx).Where("id = ?", targetID).First(&activity).Error; err != nil {
			return err
		}
		if activity.Type == models.ActivityTypeVenue {
			targetType = models.ContentTagTargetVenue
			targetID = activity.OrganizerID
			count = 1
		} else {
			count = 1
		}
	case models.ContentTagTargetVenue:
		if err := s.DB.WithContext(ctx).Table("organizers o").
			Where("o.id = ?", targetID).
			Where("EXISTS (SELECT 1 FROM activities a WHERE a.organizer_id = o.id AND a.type = ?)", models.ActivityTypeVenue).
			Count(&count).Error; err != nil {
			return err
		}
	case models.ContentTagTargetParty:
		if err := s.DB.WithContext(ctx).Model(&models.Merchant{}).Where("id = ?", targetID).Count(&count).Error; err != nil {
			return err
		}
	default:
		return errors.New("标签绑定目标类型无效")
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}
	return dao.ReplaceContentTags(ctx, s.DB, targetType, targetID, tagIDs)
}

func (s *AdminService) AuditActivity(ctx context.Context, activityID int64, req types.AdminAuditActivityRequest) error {
	if req.Status != models.ActivityStatusOnline && req.Status != models.ActivityStatusRejected {
		return errors.New("审核状态无效，仅支持 3通过上架 或 4拒绝")
	}
	if req.Status == models.ActivityStatusRejected && req.RejectReason == "" {
		return errors.New("拒绝时必须填写 reject_reason")
	}
	var activity models.Activity
	if err := s.DB.WithContext(ctx).Where("id = ?", activityID).First(&activity).Error; err != nil {
		return err
	}
	if activity.Status == models.ActivityStatusOnline && activity.PendingRevisionStatus == models.ActivityStatusPending {
		if req.Status == models.ActivityStatusRejected {
			return s.DB.WithContext(ctx).Model(&activity).Updates(map[string]any{
				"pending_revision_status": models.ActivityStatusRejected,
				"pending_revision_reason": req.RejectReason,
				"updated_at":              time.Now(),
			}).Error
		}
		payload, err := decodeActivityRevision(activity)
		if err != nil {
			return err
		}
		if payload == nil {
			return errors.New("待审核修改不存在")
		}
		if err := validateChinaCoordinate(payload.Activity.Latitude, payload.Activity.Longitude); err != nil {
			return err
		}
		return s.publishActivityRevision(ctx, activity, payload)
	}
	if req.Status == models.ActivityStatusOnline {
		if err := validateChinaCoordinate(activity.Latitude, activity.Longitude); err != nil {
			return err
		}
	}

	updates := map[string]any{
		"status":        req.Status,
		"reject_reason": "",
		"updated_at":    time.Now(),
	}
	if req.Status == models.ActivityStatusRejected {
		updates["reject_reason"] = req.RejectReason
	}
	result := s.DB.WithContext(ctx).Model(&models.Activity{}).Where("id = ?", activityID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("活动不存在")
	}
	return nil
}

func (s *AdminService) publishActivityRevision(ctx context.Context, activity models.Activity, payload *types.ActivityRevisionPayload) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked models.Activity
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND status = ? AND pending_revision_status = ?", activity.ID, models.ActivityStatusOnline, models.ActivityStatusPending).First(&locked).Error; err != nil {
			return err
		}
		candidate := payload.Activity
		updates := map[string]any{
			"type": candidate.Type, "name": candidate.Name, "share_title": candidate.ShareTitle,
			"start_time": candidate.StartTime, "end_time": candidate.EndTime,
			"real_name_mode": candidate.RealNameMode, "minor_check": candidate.MinorCheck,
			"description": candidate.Description, "province": candidate.Province, "city": candidate.City,
			"district": candidate.District, "address": candidate.Address, "latitude": candidate.Latitude,
			"longitude": candidate.Longitude, "poster_detail": candidate.PosterDetail,
			"poster_long": candidate.PosterLong, "poster_list": candidate.PosterList,
			"poster_wechat": candidate.PosterWechat, "qualification_doc": candidate.QualificationDoc,
			"audit_type": models.ActivityAuditTypeReaudit, "reject_reason": "",
			"pending_revision": "", "pending_revision_status": 0, "pending_revision_reason": "",
			"updated_at": time.Now(),
		}
		if err := tx.Model(&locked).Updates(updates).Error; err != nil {
			return err
		}
		if defaultActivityType(candidate.Type) != models.ActivityTypeVenue {
			if err := replacePublishedTicketSpecs(ctx, tx, locked.ID, payload.TicketSpecs); err != nil {
				return err
			}
		}
		targetType, targetID := models.ContentTagTargetActivity, locked.ID
		if defaultActivityType(candidate.Type) == models.ActivityTypeVenue {
			targetType, targetID = models.ContentTagTargetVenue, locked.OrganizerID
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "organizer_id"}}, DoUpdates: clause.Assignments(map[string]any{"business_hours": payload.BusinessHours, "updated_at": time.Now()})}).Create(&models.OrganizerProfile{OrganizerID: locked.OrganizerID, BusinessHours: payload.BusinessHours}).Error; err != nil {
				return err
			}
		}
		return dao.ReplaceContentTags(ctx, tx, targetType, targetID, payload.TagIDs)
	})
}

func replacePublishedTicketSpecs(ctx context.Context, tx *gorm.DB, activityID int64, desired []models.TicketSpec) error {
	var current []models.TicketSpec
	if err := tx.WithContext(ctx).Where("activity_id = ?", activityID).Find(&current).Error; err != nil {
		return err
	}
	desiredIDs := make(map[int64]bool, len(desired))
	for _, spec := range desired {
		if spec.ID == 0 {
			spec.ActivityID = activityID
			if err := tx.Create(&spec).Error; err != nil {
				return err
			}
			continue
		}
		desiredIDs[spec.ID] = true
		if err := tx.Model(&models.TicketSpec{}).Where("id = ? AND activity_id = ?", spec.ID, activityID).Updates(ticketSpecUpdates(&spec)).Error; err != nil {
			return err
		}
	}
	for _, spec := range current {
		if desiredIDs[spec.ID] {
			continue
		}
		if spec.SoldCount > 0 {
			if err := tx.Model(&spec).Updates(map[string]any{"is_enabled": 0, "updated_at": time.Now()}).Error; err != nil {
				return err
			}
			continue
		}
		if err := tx.Delete(&spec).Error; err != nil {
			return err
		}
	}
	return nil
}

// SetActivityVisibility implements the management-side delete behavior. It
// deliberately keeps the activity row and all related business records so
// orders, refunds, settlements and verification history remain traceable.
func (s *AdminService) SetActivityVisibility(ctx context.Context, activityID int64, visible bool, reason string) (*models.Activity, error) {
	var activity models.Activity
	if err := s.DB.WithContext(ctx).Where("id = ?", activityID).First(&activity).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	updates := map[string]any{
		"is_hidden":     0,
		"hidden_at":     nil,
		"hidden_reason": "",
		"updated_at":    now,
	}
	if !visible {
		updates["is_hidden"] = 1
		updates["hidden_at"] = now
		updates["hidden_reason"] = strings.TrimSpace(reason)
	}
	if err := s.DB.WithContext(ctx).Model(&activity).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := s.DB.WithContext(ctx).Where("id = ?", activityID).First(&activity).Error; err != nil {
		return nil, err
	}
	return &activity, nil
}

func (s *AdminService) GetEventTicketList(ctx context.Context, eventID int64, page, pageSize int) (*types.AdminTicketListResponse, error) {
	query := s.DB.WithContext(ctx).Model(&models.EventTicket{}).Where("event_id = ?", eventID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var tickets []models.EventTicket
	if err := query.
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&tickets).Error; err != nil {
		return nil, err
	}

	// 获取活动标题
	var event models.Event
	s.DB.WithContext(ctx).Where("id = ?", eventID).First(&event)

	list := make([]types.AdminTicketItem, 0, len(tickets))
	for _, t := range tickets {
		list = append(list, types.AdminTicketItem{
			EventTicket: t,
			EventTitle:  event.Title,
		})
	}

	return &types.AdminTicketListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *AdminService) GetAllTickets(ctx context.Context, page, pageSize int, keyword string) (*types.AdminTicketListResponse, error) {
	query := s.DB.WithContext(ctx).Model(&models.EventTicket{})

	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Joins("LEFT JOIN events ON events.id = event_tickets.event_id").
			Where("event_tickets.name LIKE ? OR events.title LIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var tickets []models.EventTicket
	if err := query.
		Order("event_tickets.created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&tickets).Error; err != nil {
		return nil, err
	}

	// 批量获取活动标题
	eventIDs := make([]int64, 0, len(tickets))
	for _, t := range tickets {
		eventIDs = append(eventIDs, t.EventID)
	}
	eventTitleMap := make(map[int64]string)
	if len(eventIDs) > 0 {
		var events []models.Event
		s.DB.WithContext(ctx).Where("id IN ?", eventIDs).Find(&events)
		for _, e := range events {
			eventTitleMap[e.ID] = e.Title
		}
	}

	list := make([]types.AdminTicketItem, 0, len(tickets))
	for _, t := range tickets {
		list = append(list, types.AdminTicketItem{
			EventTicket: t,
			EventTitle:  eventTitleMap[t.EventID],
		})
	}

	return &types.AdminTicketListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *AdminService) GetOrderList(ctx context.Context, page, pageSize int, eventID int64) (*types.AdminOrderListResponse, error) {
	query := s.DB.WithContext(ctx).Model(&models.Order{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var orders []models.Order
	if err := query.
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&orders).Error; err != nil {
		return nil, err
	}

	// 批量获取用户信息
	userIDs := make([]int, 0, len(orders))
	for _, o := range orders {
		userIDs = append(userIDs, o.UserID)
	}
	userMap := make(map[int]*models.Users)
	if len(userIDs) > 0 {
		var users []models.Users
		s.DB.WithContext(ctx).Where("id IN ?", userIDs).Find(&users)
		for i := range users {
			userMap[users[i].Id] = &users[i]
		}
	}

	list := make([]types.AdminOrderItem, 0, len(orders))
	for _, o := range orders {
		item := types.AdminOrderItem{Order: o}
		if u, ok := userMap[o.UserID]; ok {
			item.UserName = u.Nickname
			item.UserMobile = u.Mobile
		}
		list = append(list, item)
	}

	return &types.AdminOrderListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *AdminService) GetTicketOrderList(ctx context.Context, page, pageSize int, activityID int64, status, refundStatus *int8, keyword, salesChannel string) (*types.AdminTicketOrderListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	query := s.DB.WithContext(ctx).Model(&models.TicketOrder{}).
		Joins("LEFT JOIN users u ON u.id = ticket_orders.user_id").
		Joins("LEFT JOIN activities a ON a.id = ticket_orders.activity_id").
		Joins("LEFT JOIN ticket_specs ts ON ts.id = ticket_orders.ticket_spec_id")
	if activityID > 0 {
		query = query.Where("ticket_orders.activity_id = ?", activityID)
	}
	if status != nil {
		query = query.Where("ticket_orders.status = ?", *status)
	}
	if refundStatus != nil {
		query = query.Where(`EXISTS (
			SELECT 1 FROM refunds rf
			WHERE rf.order_id = ticket_orders.id
			AND rf.status = ?
			AND rf.id = (
				SELECT MAX(rf2.id) FROM refunds rf2 WHERE rf2.order_id = ticket_orders.id
			)
		)`, *refundStatus)
	}
	normalizedSalesChannel, err := NormalizeSalesChannel(salesChannel, false)
	if err != nil {
		return nil, err
	}
	if normalizedSalesChannel != "" {
		query = query.Where("ticket_orders.sales_channel = ?", normalizedSalesChannel)
	}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where(`ticket_orders.order_no LIKE ?
			OR ticket_orders.buyer_name LIKE ?
			OR ticket_orders.buyer_id_card LIKE ?
			OR u.nickname LIKE ?
			OR u.mobile LIKE ?
			OR a.name LIKE ?
			OR ts.name LIKE ?`, like, like, like, like, like, like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var orders []models.TicketOrder
	if err := query.Select("ticket_orders.*").Order("ticket_orders.created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&orders).Error; err != nil {
		return nil, err
	}
	list, err := s.buildAdminTicketOrderItems(ctx, orders)
	if err != nil {
		return nil, err
	}
	return &types.AdminTicketOrderListResponse{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *AdminService) GetTicketOrderDetail(ctx context.Context, orderNo string) (*types.AdminTicketOrderDetail, error) {
	var order models.TicketOrder
	if err := s.DB.WithContext(ctx).Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return nil, err
	}
	items, err := s.buildAdminTicketOrderItems(ctx, []models.TicketOrder{order})
	if err != nil {
		return nil, err
	}
	var refunds []models.Refund
	if err := s.DB.WithContext(ctx).Where("order_id = ?", order.ID).Order("id DESC").Find(&refunds).Error; err != nil {
		return nil, err
	}
	refundIDs := make([]int64, 0, len(refunds))
	for _, refund := range refunds {
		refundIDs = append(refundIDs, refund.ID)
	}
	var logs []models.RefundLog
	if len(refundIDs) > 0 {
		if err := s.DB.WithContext(ctx).Where("refund_id IN ?", refundIDs).Order("id ASC").Find(&logs).Error; err != nil {
			return nil, err
		}
	}
	var viewers []models.TicketOrderViewer
	if err := s.DB.WithContext(ctx).Where("order_id = ?", order.ID).Order("id ASC").Find(&viewers).Error; err != nil {
		return nil, err
	}
	var verificationRecords []map[string]any
	if err := s.DB.WithContext(ctx).Table("verification_records vr").
		Select("vr.*, v.name AS verifier_name, v.phone AS verifier_phone, a.name AS activity_name").
		Joins("LEFT JOIN verifiers v ON v.id = vr.verifier_id").
		Joins("LEFT JOIN activities a ON a.id = vr.activity_id").
		Where("vr.order_id = ?", order.ID).
		Order("vr.id DESC").
		Find(&verificationRecords).Error; err != nil {
		return nil, err
	}
	var payRecords []models.PayRecord
	if err := s.DB.WithContext(ctx).Where("order_sn = ?", order.OrderNo).Order("id DESC").Find(&payRecords).Error; err != nil {
		return nil, err
	}
	return &types.AdminTicketOrderDetail{
		Order:               items[0],
		Viewers:             orderViewerItems(viewers, true),
		Refunds:             refunds,
		RefundLogs:          logs,
		VerificationRecords: verificationRecords,
		PayRecords:          payRecords,
	}, nil
}

// GetRefundDetail returns one refund and the order context required by the admin refund detail page.
func (s *AdminService) GetRefundDetail(ctx context.Context, refundNo string) (*types.AdminRefundDetail, error) {
	var refund models.Refund
	if err := s.DB.WithContext(ctx).Where("refund_no = ?", refundNo).First(&refund).Error; err != nil {
		return nil, err
	}
	var order models.TicketOrder
	if err := s.DB.WithContext(ctx).Where("id = ?", refund.OrderID).First(&order).Error; err != nil {
		return nil, err
	}
	orderDetail, err := s.GetTicketOrderDetail(ctx, order.OrderNo)
	if err != nil {
		return nil, err
	}
	logs := make([]models.RefundLog, 0)
	for _, log := range orderDetail.RefundLogs {
		if log.RefundID == refund.ID {
			logs = append(logs, log)
		}
	}
	return &types.AdminRefundDetail{
		Refund:              refund,
		Order:               orderDetail.Order,
		Viewers:             orderDetail.Viewers,
		RefundLogs:          logs,
		VerificationRecords: orderDetail.VerificationRecords,
		PayRecords:          orderDetail.PayRecords,
	}, nil
}

func (s *AdminService) ApproveOrderRefund(ctx context.Context, orderNo string) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order models.TicketOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_no = ?", orderNo).First(&order).Error; err != nil {
			return err
		}
		var refund models.Refund
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_id = ?", order.ID).Order("id DESC").First(&refund).Error; err != nil {
			return err
		}
		if refund.Status != models.RefundStatusAuditing {
			return errors.New("退款单状态不可审核通过")
		}
		if err := tx.Model(&refund).Updates(map[string]any{"status": models.RefundStatusRunning, "wechat_status": "ADMIN_APPROVED"}).Error; err != nil {
			return err
		}
		if err := tx.Model(&order).Update("status", models.TicketOrderStatusRefunding).Error; err != nil {
			return err
		}
		return tx.Create(&models.RefundLog{RefundID: refund.ID, Status: "退款中", Description: "管理员审核通过，等待退款处理"}).Error
	})
}

func (s *AdminService) RejectOrderRefund(ctx context.Context, orderNo string, reason string) error {
	if reason == "" {
		return errors.New("拒绝原因不能为空")
	}
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order models.TicketOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_no = ?", orderNo).First(&order).Error; err != nil {
			return err
		}
		var refund models.Refund
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_id = ?", order.ID).Order("id DESC").First(&refund).Error; err != nil {
			return err
		}
		if refund.Status != models.RefundStatusAuditing {
			return errors.New("退款单状态不可拒绝")
		}
		if err := tx.Model(&refund).Updates(map[string]any{"status": models.RefundStatusRejected, "reject_reason": reason}).Error; err != nil {
			return err
		}
		if err := tx.Model(&order).Update("status", models.TicketOrderStatusRefundReject).Error; err != nil {
			return err
		}
		return tx.Create(&models.RefundLog{RefundID: refund.ID, Status: "退款拒绝", Description: reason}).Error
	})
}

func (s *AdminService) GetFinanceSummary(ctx context.Context) (*types.AdminFinanceSummary, error) {
	resp := &types.AdminFinanceSummary{}
	paidStatuses := []int8{models.TicketOrderStatusUsable, models.TicketOrderStatusUsed, models.TicketOrderStatusRefunding, models.TicketOrderStatusRefundSuccess, models.TicketOrderStatusRefundReject}
	if err := s.DB.WithContext(ctx).Model(&models.TicketOrder{}).Where("status IN ?", paidStatuses).
		Select("COALESCE(SUM(actual_price),0), COUNT(*)").Row().Scan(&resp.TotalRevenue, &resp.OrderCount); err != nil {
		return nil, err
	}
	if err := s.DB.WithContext(ctx).Model(&models.Refund{}).Where("status = ?", models.RefundStatusSuccess).
		Select("COALESCE(SUM(refund_amount),0)").Scan(&resp.RefundAmount).Error; err != nil {
		return nil, err
	}
	resp.PendingSettle = resp.TotalRevenue - resp.RefundAmount
	resp.SettledAmount = 0
	return resp, nil
}

func (s *AdminService) GetFinanceSettlements(ctx context.Context, page, pageSize int, organizerID int64) (*types.AdminSettlementListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	base := s.DB.WithContext(ctx).Table("organizers org").
		Joins("LEFT JOIN activities a ON a.organizer_id = org.id").
		Joins("LEFT JOIN ticket_orders o ON o.activity_id = a.id AND o.status IN ?", []int8{models.TicketOrderStatusUsable, models.TicketOrderStatusUsed, models.TicketOrderStatusRefunding, models.TicketOrderStatusRefundSuccess, models.TicketOrderStatusRefundReject}).
		Joins("LEFT JOIN refunds r ON r.order_id = o.id AND r.status = ?", models.RefundStatusSuccess).
		Joins(`LEFT JOIN (
			SELECT organizer_id,
				COALESCE(SUM(CASE WHEN status = 1 THEN amount ELSE 0 END), 0) AS withdraw_amount,
				COALESCE(SUM(CASE WHEN status = 0 THEN amount ELSE 0 END), 0) AS pending_withdraw_amount,
				MAX(updated_at) AS withdraw_updated_at
			FROM organizer_withdraws
			GROUP BY organizer_id
		) w ON w.organizer_id = org.id`)
	if organizerID > 0 {
		base = base.Where("org.id = ?", organizerID)
	}
	var total int64
	countQuery := s.DB.WithContext(ctx).Model(&models.Organizer{})
	if organizerID > 0 {
		countQuery = countQuery.Where("id = ?", organizerID)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []types.AdminSettlementItem
	if err := base.Select(`org.id AS organizer_id,
			org.name AS organizer_name,
			COALESCE(SUM(o.actual_price),0) AS gross_amount,
			COALESCE(SUM(r.refund_amount),0) AS refund_amount,
			COALESCE(MAX(w.withdraw_amount),0) AS withdraw_amount,
			COALESCE(MAX(w.pending_withdraw_amount),0) AS pending_withdraw_amount,
			COALESCE(SUM(o.actual_price),0) - COALESCE(SUM(r.refund_amount),0) AS settle_amount,
			COALESCE(SUM(o.actual_price),0) - COALESCE(SUM(r.refund_amount),0) - COALESCE(MAX(w.withdraw_amount),0) - COALESCE(MAX(w.pending_withdraw_amount),0) AS pending_settle_amount,
			COUNT(o.id) AS order_count,
			DATE_FORMAT(GREATEST(org.updated_at, COALESCE(MAX(w.withdraw_updated_at), org.updated_at)), '%Y-%m-%d %H:%i:%s') AS updated_at`).
		Group("org.id, org.name, org.updated_at").
		Order("org.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return &types.AdminSettlementListResponse{List: rows, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *AdminService) GetUserList(ctx context.Context, page, pageSize int, keyword string) (*types.AdminUserListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	likesSubquery := "(SELECT COUNT(*) FROM note_likes nl JOIN notes n ON n.id = nl.note_id WHERE n.user_id = u.id AND n.status <> -1 AND nl.status = 1)"
	query := s.DB.WithContext(ctx).Table("users u")
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("u.nickname LIKE ? OR u.mobile LIKE ?", like, like)
	}
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, err
	}
	var users []models.Users
	if err := query.Select("u.*, " + likesSubquery + " AS likes_cnt").
		Order("likes_cnt DESC, u.created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Scan(&users).Error; err != nil {
		return nil, err
	}
	userIDs := make([]int, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.Id)
	}
	stats := map[int]struct {
		Amount int64
		Count  int64
	}{}
	if len(userIDs) > 0 {
		var rows []struct {
			UserID int
			Amount int64
			Count  int64
		}
		_ = s.DB.WithContext(ctx).Model(&models.TicketOrder{}).Select("user_id, COALESCE(SUM(actual_price),0) AS amount, COUNT(*) AS count").Where("user_id IN ?", userIDs).Group("user_id").Scan(&rows).Error
		for _, row := range rows {
			stats[row.UserID] = struct {
				Amount int64
				Count  int64
			}{Amount: row.Amount, Count: row.Count}
		}
	}
	counts := s.countUserBehavior(ctx, userIDs)
	list := make([]types.AdminUserItem, 0, len(users))
	for _, user := range users {
		item := types.AdminUserItem{ID: user.Id, Nickname: user.Nickname, Avatar: user.Avatar, Mobile: user.Mobile, Status: user.Status, CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05")}
		if stat, ok := stats[user.Id]; ok {
			item.TotalAmount = stat.Amount
			item.OrderCount = stat.Count
		}
		if c, ok := counts[user.Id]; ok {
			item.LikesCount = c.Likes
			item.CollectionsCount = c.Collections
			item.FollowingCount = c.Following
			item.FollowersCount = c.Followers
			item.AttendCount = c.Attends
			item.SubscribeCount = c.Subscribes
		}
		list = append(list, item)
	}
	return &types.AdminUserListResponse{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

type userBehaviorCounts struct {
	Likes       int64
	Collections int64
	Following   int64
	Followers   int64
	Attends     int64
	Subscribes  int64
}

// countUserBehavior aggregates likes/collections/follow/attend/subscribe counts
// for the given page of users, matching the semantics of ListUserRecords.
func (s *AdminService) countUserBehavior(ctx context.Context, userIDs []int) map[int]userBehaviorCounts {
	result := make(map[int]userBehaviorCounts, len(userIDs))
	if len(userIDs) == 0 {
		return result
	}

	type countRow struct {
		UserID int
		Cnt    int64
	}

	// likes received on this user's published notes
	var likeRows []countRow
	_ = s.DB.WithContext(ctx).Table("note_likes nl").
		Select("n.user_id AS user_id, COUNT(*) AS cnt").
		Joins("JOIN notes n ON n.id = nl.note_id").
		Where("n.user_id IN ? AND n.status <> ? AND nl.status = 1", userIDs, -1).
		Group("n.user_id").Scan(&likeRows).Error
	for _, row := range likeRows {
		c := result[row.UserID]
		c.Likes = row.Cnt
		result[row.UserID] = c
	}

	// collections by this user
	var collectionRows []countRow
	_ = s.DB.WithContext(ctx).Table("note_collections").
		Select("user_id, COUNT(*) AS cnt").
		Where("user_id IN ? AND status = 1", userIDs).
		Group("user_id").Scan(&collectionRows).Error
	for _, row := range collectionRows {
		c := result[row.UserID]
		c.Collections = row.Cnt
		result[row.UserID] = c
	}

	// following
	var followingRows []countRow
	_ = s.DB.WithContext(ctx).Table("user_follow").
		Select("follower_id AS user_id, COUNT(*) AS cnt").
		Where("follower_id IN ? AND status = 1", userIDs).
		Group("follower_id").Scan(&followingRows).Error
	for _, row := range followingRows {
		c := result[row.UserID]
		c.Following = row.Cnt
		result[row.UserID] = c
	}

	// followers
	var followerRows []countRow
	_ = s.DB.WithContext(ctx).Table("user_follow").
		Select("followee_id AS user_id, COUNT(*) AS cnt").
		Where("followee_id IN ? AND status = 1", userIDs).
		Group("followee_id").Scan(&followerRows).Error
	for _, row := range followerRows {
		c := result[row.UserID]
		c.Followers = row.Cnt
		result[row.UserID] = c
	}

	// attended orders
	var attendRows []countRow
	_ = s.DB.WithContext(ctx).Table("ticket_orders").
		Select("user_id, COUNT(*) AS cnt").
		Where("user_id IN ? AND status IN ?", userIDs, []int8{models.TicketOrderStatusUsable, models.TicketOrderStatusUsed, models.TicketOrderStatusRefunding, models.TicketOrderStatusRefundSuccess, models.TicketOrderStatusRefundReject}).
		Group("user_id").Scan(&attendRows).Error
	for _, row := range attendRows {
		c := result[row.UserID]
		c.Attends = row.Cnt
		result[row.UserID] = c
	}

	// activity subscriptions
	var subscribeRows []countRow
	_ = s.DB.WithContext(ctx).Table("activity_subscriptions").
		Select("user_id, COUNT(*) AS cnt").
		Where("user_id IN ?", userIDs).
		Group("user_id").Scan(&subscribeRows).Error
	for _, row := range subscribeRows {
		c := result[row.UserID]
		c.Subscribes = row.Cnt
		result[row.UserID] = c
	}

	return result
}

func (s *AdminService) UpdateUserStatus(ctx context.Context, userID int, status int8) error {
	result := s.DB.WithContext(ctx).Model(&models.Users{}).Where("id = ?", userID).Updates(map[string]any{"status": status, "updated_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("用户不存在")
	}
	return nil
}

func (s *AdminService) ListBanners(ctx context.Context) ([]models.PlatformBanner, error) {
	var banners []models.PlatformBanner
	err := s.DB.WithContext(ctx).Order("sort ASC,id DESC").Find(&banners).Error
	return banners, err
}

func (s *AdminService) CreateBanner(ctx context.Context, req types.AdminBannerRequest) (int64, error) {
	banner := models.PlatformBanner{Title: req.Title, Image: req.Image, Link: req.Link, Position: req.Position, Sort: req.Sort, Status: req.Status}
	if banner.Position == "" {
		banner.Position = "home"
	}
	if banner.Status == 0 {
		banner.Status = 1
	}
	if err := s.DB.WithContext(ctx).Create(&banner).Error; err != nil {
		return 0, err
	}
	return banner.ID, nil
}

func (s *AdminService) UpdateBanner(ctx context.Context, id int64, req types.AdminBannerRequest) error {
	updates := map[string]any{"title": req.Title, "image": req.Image, "link": req.Link, "position": req.Position, "sort": req.Sort, "status": req.Status, "updated_at": time.Now()}
	if updates["position"] == "" {
		updates["position"] = "home"
	}
	return s.DB.WithContext(ctx).Model(&models.PlatformBanner{}).Where("id = ?", id).Updates(updates).Error
}

func (s *AdminService) DeleteBanner(ctx context.Context, id int64) error {
	return s.DB.WithContext(ctx).Delete(&models.PlatformBanner{}, id).Error
}

func (s *AdminService) SortBanners(ctx context.Context, req types.AdminBannerSortRequest) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range req.List {
			if err := tx.Model(&models.PlatformBanner{}).Where("id = ?", item.ID).Update("sort", item.Sort).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *AdminService) GetSettings(ctx context.Context) ([]types.AdminSettingItem, error) {
	var rows []models.PlatformSetting
	if err := s.DB.WithContext(ctx).Order("setting_key ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	resp := make([]types.AdminSettingItem, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, types.AdminSettingItem{Key: row.Key, Value: row.Value, Remark: row.Remark})
	}
	return resp, nil
}

func (s *AdminService) UpdateSettings(ctx context.Context, settings []types.AdminSettingItem) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range settings {
			row := models.PlatformSetting{Key: item.Key, Value: item.Value, Remark: item.Remark}
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "setting_key"}}, DoUpdates: clause.Assignments(map[string]any{"setting_value": item.Value, "remark": item.Remark, "updated_at": time.Now()})}).Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *AdminService) GetSystemConfig(ctx context.Context) (*types.AdminSystemConfig, error) {
	keys := []string{"system_name", "icp_record_no", "customer_service_phone", "customer_service_wechat", "customer_service_email", "customer_service_hours", "customer_service_user_id", "withdraw_arrival_cycle", "direct_message_enabled"}
	var rows []models.PlatformSetting
	if err := s.DB.WithContext(ctx).Where("setting_key IN ?", keys).Find(&rows).Error; err != nil {
		return nil, err
	}
	values := make(map[string]string, len(rows))
	for _, row := range rows {
		values[row.Key] = row.Value
	}
	customerServiceUserID, _ := strconv.ParseInt(values["customer_service_user_id"], 10, 64)
	directMessageEnabled := platformBoolEnabled(values["direct_message_enabled"], true)
	return &types.AdminSystemConfig{
		SystemName: values["system_name"], ICPRecordNo: values["icp_record_no"], CustomerServicePhone: values["customer_service_phone"],
		CustomerServiceWechat: values["customer_service_wechat"], CustomerServiceEmail: values["customer_service_email"], CustomerServiceHours: values["customer_service_hours"],
		CustomerServiceUserID: customerServiceUserID, WithdrawArrivalCycle: values["withdraw_arrival_cycle"], DirectMessageEnabled: &directMessageEnabled,
	}, nil
}

func (s *AdminService) UpdateSystemConfig(ctx context.Context, config types.AdminSystemConfig) error {
	config.SystemName = strings.TrimSpace(config.SystemName)
	config.ICPRecordNo = strings.TrimSpace(config.ICPRecordNo)
	config.CustomerServicePhone = strings.TrimSpace(config.CustomerServicePhone)
	config.CustomerServiceWechat = strings.TrimSpace(config.CustomerServiceWechat)
	config.CustomerServiceEmail = strings.TrimSpace(config.CustomerServiceEmail)
	config.CustomerServiceHours = strings.TrimSpace(config.CustomerServiceHours)
	config.WithdrawArrivalCycle = strings.TrimSpace(config.WithdrawArrivalCycle)
	if config.SystemName == "" {
		return errors.New("系统名称不能为空")
	}
	if len(config.SystemName) > 100 || len(config.ICPRecordNo) > 100 || len(config.CustomerServicePhone) > 50 || len(config.CustomerServiceWechat) > 100 || len(config.CustomerServiceEmail) > 100 || len(config.CustomerServiceHours) > 100 || len(config.WithdrawArrivalCycle) > 100 {
		return errors.New("系统配置字段长度超限")
	}
	settings := []types.AdminSettingItem{
		{Key: "system_name", Value: config.SystemName, Remark: "平台系统名称"}, {Key: "icp_record_no", Value: config.ICPRecordNo, Remark: "ICP备案号"},
		{Key: "customer_service_phone", Value: config.CustomerServicePhone, Remark: "客服电话"}, {Key: "customer_service_wechat", Value: config.CustomerServiceWechat, Remark: "客服微信"},
		{Key: "customer_service_email", Value: config.CustomerServiceEmail, Remark: "客服邮箱"}, {Key: "customer_service_hours", Value: config.CustomerServiceHours, Remark: "客服服务时间"},
		{Key: "withdraw_arrival_cycle", Value: config.WithdrawArrivalCycle, Remark: "商家提现到账周期展示文案"},
	}
	// Keep an existing service account when older PC clients do not yet submit
	// this optional field. A positive value explicitly replaces the account.
	if config.CustomerServiceUserID > 0 {
		settings = append(settings, types.AdminSettingItem{Key: "customer_service_user_id", Value: strconv.FormatInt(config.CustomerServiceUserID, 10), Remark: "客服聊天用户 ID"})
	}
	if config.DirectMessageEnabled != nil {
		value := "0"
		if *config.DirectMessageEnabled {
			value = "1"
		}
		settings = append(settings, types.AdminSettingItem{Key: "direct_message_enabled", Value: value, Remark: "普通用户私信开关"})
	}
	return s.UpdateSettings(ctx, settings)
}

func platformBoolEnabled(value string, fallback bool) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "on"
}

func (s *AdminService) buildAdminTicketOrderItems(ctx context.Context, orders []models.TicketOrder) ([]types.AdminTicketOrderItem, error) {
	userIDs := make([]int, 0, len(orders))
	activityIDs := make([]int64, 0, len(orders))
	specIDs := make([]int64, 0, len(orders))
	for _, order := range orders {
		userIDs = append(userIDs, int(order.UserID))
		activityIDs = append(activityIDs, order.ActivityID)
		specIDs = append(specIDs, order.TicketSpecID)
	}
	userMap := map[int]models.Users{}
	if len(userIDs) > 0 {
		var users []models.Users
		if err := s.DB.WithContext(ctx).Where("id IN ?", userIDs).Find(&users).Error; err != nil {
			return nil, err
		}
		for _, user := range users {
			userMap[user.Id] = user
		}
	}
	activityMap := map[int64]models.Activity{}
	if len(activityIDs) > 0 {
		var activities []models.Activity
		if err := s.DB.WithContext(ctx).Where("id IN ?", activityIDs).Find(&activities).Error; err != nil {
			return nil, err
		}
		for _, activity := range activities {
			activityMap[activity.ID] = activity
		}
	}
	specMap := map[int64]models.TicketSpec{}
	if len(specIDs) > 0 {
		var specs []models.TicketSpec
		if err := s.DB.WithContext(ctx).Where("id IN ?", specIDs).Find(&specs).Error; err != nil {
			return nil, err
		}
		for _, spec := range specs {
			specMap[spec.ID] = spec
		}
	}
	list := make([]types.AdminTicketOrderItem, 0, len(orders))
	for _, order := range orders {
		item := types.AdminTicketOrderItem{
			OrderNo:        order.OrderNo,
			Status:         order.Status,
			TotalPrice:     order.TotalPrice,
			ActualPrice:    order.ActualPrice,
			PointsAmount:   order.PointsAmount,
			PointsDiscount: order.PointsDiscount,
			Quantity:       order.Quantity,
			UserID:         order.UserID,
			BuyerName:      order.BuyerName,
			BuyerIDCard:    order.BuyerIDCard,
			ActivityID:     order.ActivityID,
			TicketSpecID:   order.TicketSpecID,
			PayMethod:      order.PayMethod,
			SalesChannel:   order.SalesChannel,
			PayTime:        formatAdminPtrTime(order.PayTime),
			ExpireTime:     formatAdminTime(order.ExpireTime),
			CreatedAt:      order.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if user, ok := userMap[int(order.UserID)]; ok {
			item.UserName = user.Nickname
			item.UserMobile = user.Mobile
		}
		if activity, ok := activityMap[order.ActivityID]; ok {
			item.ActivityName = activity.Name
		}
		if spec, ok := specMap[order.TicketSpecID]; ok {
			item.TicketSpecName = spec.Name
		}
		list = append(list, item)
	}
	return list, nil
}

func (s *AdminService) GetDashboardStats(ctx context.Context) (*types.AdminDashboardStats, error) {
	stats := &types.AdminDashboardStats{}
	db := s.DB.WithContext(ctx)
	if err := db.Model(&models.Organizer{}).Where("status = ?", models.OrganizerStatusApproved).Count(&stats.TotalMerchants).Error; err != nil {
		return nil, err
	}

	if err := db.Model(&models.Merchant{}).Where("type = ?", "场地").Count(&stats.TotalParties).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&models.Merchant{}).Where("type = ?", "派对").Count(&stats.TotalEvents).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&models.EventTicket{}).Count(&stats.TotalTickets).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&models.Order{}).Count(&stats.TotalOrders).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&models.Users{}).Count(&stats.TotalUsers).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&models.Order{}).
		Where("status >= ?", 20).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&stats.TotalRevenue).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

func formatAdminTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func formatAdminPtrTime(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
