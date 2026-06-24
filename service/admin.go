package service

import (
	"Hyper/dao"
	"Hyper/models"
	"Hyper/pkg/encrypt"
	"Hyper/pkg/jwt"
	"Hyper/types"
	"context"
	"errors"
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
	GetEventTicketList(ctx context.Context, eventID int64, page, pageSize int) (*types.AdminTicketListResponse, error)
	GetAllTickets(ctx context.Context, page, pageSize int, keyword string) (*types.AdminTicketListResponse, error)
	GetOrderList(ctx context.Context, page, pageSize int, eventID int64) (*types.AdminOrderListResponse, error)
	GetTicketOrderList(ctx context.Context, page, pageSize int, activityID int64, status *int8, keyword string) (*types.AdminTicketOrderListResponse, error)
	GetTicketOrderDetail(ctx context.Context, orderNo string) (*types.AdminTicketOrderDetail, error)
	ApproveOrderRefund(ctx context.Context, orderNo string) error
	RejectOrderRefund(ctx context.Context, orderNo string, reason string) error
	GetFinanceSummary(ctx context.Context) (*types.AdminFinanceSummary, error)
	GetFinanceSettlements(ctx context.Context, page, pageSize int, organizerID int64) (*types.AdminSettlementListResponse, error)
	GetUserList(ctx context.Context, page, pageSize int, keyword string) (*types.AdminUserListResponse, error)
	UpdateUserStatus(ctx context.Context, userID int, status int8) error
	ListBanners(ctx context.Context) ([]models.PlatformBanner, error)
	CreateBanner(ctx context.Context, req types.AdminBannerRequest) (int64, error)
	UpdateBanner(ctx context.Context, id int64, req types.AdminBannerRequest) error
	DeleteBanner(ctx context.Context, id int64) error
	SortBanners(ctx context.Context, req types.AdminBannerSortRequest) error
	GetSettings(ctx context.Context) ([]types.AdminSettingItem, error)
	UpdateSettings(ctx context.Context, settings []types.AdminSettingItem) error
	GetDashboardStats(ctx context.Context) (*types.AdminDashboardStats, error)
	GetAdminProfile(ctx context.Context, adminID int64) (*types.AdminProfileResponse, error)
	UpdateAdminProfile(ctx context.Context, adminID int64, req types.AdminProfileRequest) error
	UpdateAdminPassword(ctx context.Context, adminID int64, req types.AdminPasswordRequest) error
	ListAdmins(ctx context.Context, page, pageSize int, keyword string) (*types.AdminPageResponse[models.Admin], error)
	CreateAdmin(ctx context.Context, req types.AdminAccountRequest) (int64, error)
	UpdateAdmin(ctx context.Context, id int64, req types.AdminAccountRequest) error
	DeleteAdmin(ctx context.Context, id int64) error
	ListRoles(ctx context.Context, page, pageSize int, keyword string) (*types.AdminPageResponse[models.AdminRole], error)
	SaveRole(ctx context.Context, id int64, req types.AdminRoleRequest) (int64, error)
	DeleteRole(ctx context.Context, id int64) error
	ListOperationLogs(ctx context.Context, page, pageSize int, adminID int64, keyword string) (*types.AdminPageResponse[models.AdminOperationLog], error)
	RecordOperationLog(ctx context.Context, log models.AdminOperationLog) error
	CheckPermission(ctx context.Context, adminID int64, method, path string) error
	ListCategories(ctx context.Context, page, pageSize int, categoryType string) (*types.AdminPageResponse[models.AdminCategory], error)
	SaveCategory(ctx context.Context, id int64, req types.AdminCategoryRequest) (int64, error)
	DeleteCategory(ctx context.Context, id int64) error
	ListUserRecords(ctx context.Context, userID int64, recordType string, page, pageSize int) (*types.AdminPageResponse[map[string]any], error)
	ListViewers(ctx context.Context, page, pageSize int, keyword string) (*types.AdminPageResponse[map[string]any], error)
	ListActivityCollections(ctx context.Context, page, pageSize int, keyword string, organizerID int64) (*types.AdminPageResponse[types.ActivityCollectionItem], error)
	SaveActivityCollection(ctx context.Context, id int64, req types.ActivityCollectionRequest) (int64, error)
	DeleteActivityCollection(ctx context.Context, id int64) error
	ListVerifiers(ctx context.Context, page, pageSize int, keyword string, organizerID int64) (*types.AdminPageResponse[map[string]any], error)
	UpdateVerifierStatus(ctx context.Context, id int64, status int8) error
	ListVerificationRecords(ctx context.Context, page, pageSize int, keyword string, organizerID int64) (*types.AdminPageResponse[map[string]any], error)
	ListNotes(ctx context.Context, page, pageSize int, status *int, keyword string) (*types.AdminPageResponse[map[string]any], error)
	UpdateNoteStatus(ctx context.Context, noteID int64, status int) error
	ListNoteRecords(ctx context.Context, recordType string, page, pageSize int, noteID int64) (*types.AdminPageResponse[map[string]any], error)
	UpdateCommentStatus(ctx context.Context, noteID, commentID int64, status int8) error
	ListPointLogs(ctx context.Context, page, pageSize int, userID int64) (*types.AdminPageResponse[map[string]any], error)
	ListWithdraws(ctx context.Context, page, pageSize int, status *int8, organizerID int64) (*types.AdminPageResponse[map[string]any], error)
	AuditWithdraw(ctx context.Context, id int64, req types.WithdrawAuditRequest) error
	ListBankAccountAudits(ctx context.Context, page, pageSize int, status *int8, organizerID int64) (*types.AdminPageResponse[map[string]any], error)
	AuditBankAccount(ctx context.Context, id int64, req types.BankAccountAuditRequest) error
	ListOrganizerLevelRules(ctx context.Context) ([]models.OrganizerLevelRule, error)
	SaveOrganizerLevelRule(ctx context.Context, id int64, req types.OrganizerLevelRuleRequest) (int64, error)
	DeleteOrganizerLevelRule(ctx context.Context, id int64) error
	ListMessages(ctx context.Context, page, pageSize int) (*types.AdminPageResponse[models.PlatformMessage], error)
	ListMessageDeliveries(ctx context.Context, messageID int64, page, pageSize int, status *int8) (*types.AdminPageResponse[map[string]any], error)
	CreateMessage(ctx context.Context, req types.PlatformMessageRequest) (int64, error)
	GetPointsRule(ctx context.Context) (*types.PointsRule, error)
	UpdatePointsRule(ctx context.Context, req types.UpdatePointsRuleRequest) error
}

type AdminService struct {
	AdminDAO      *dao.Admin
	DB            *gorm.DB
	Secret        []byte
	WeChatService IWeChatService
	MqProducer    rmq_client.Producer
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
		query = query.Where("status = ?", *status)
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
			ID:             org.ID,
			UserID:         org.UserID,
			Type:           org.Type,
			Name:           org.Name,
			Logo:           org.Logo,
			Status:         org.Status,
			Enabled:        org.Enabled,
			RejectReason:   org.RejectReason,
			Level:          org.Level,
			ServiceFeeRate: org.ServiceFeeRate,
			Province:       org.Province,
			City:           org.City,
			District:       org.District,
			CreatedAt:      org.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:      org.UpdatedAt.Format("2006-01-02 15:04:05"),
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
	var user models.Users
	if err := s.DB.WithContext(ctx).Where("id = ?", org.UserID).First(&user).Error; err == nil {
		detail.UserName = user.Nickname
		detail.UserAvatar = user.Avatar
		detail.UserMobile = user.Mobile
	}
	return detail, nil
}

func (s *AdminService) AuditOrganizer(ctx context.Context, organizerID int64, req types.AdminAuditOrganizerRequest) error {
	if req.Status != models.OrganizerStatusApproved && req.Status != models.OrganizerStatusRejected {
		return errors.New("审核状态无效，仅支持 2通过 或 3拒绝")
	}
	if req.Status == models.OrganizerStatusRejected && req.RejectReason == "" {
		return errors.New("拒绝时必须填写 reject_reason")
	}
	updates := map[string]any{
		"status":        req.Status,
		"reject_reason": "",
		"updated_at":    time.Now(),
	}
	if req.Status == models.OrganizerStatusApproved {
		updates["enabled"] = 1
	}
	if req.Status == models.OrganizerStatusRejected {
		updates["reject_reason"] = req.RejectReason
	}
	result := s.DB.WithContext(ctx).Model(&models.Organizer{}).Where("id = ?", organizerID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("入驻申请不存在")
	}
	return nil
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
		query = query.Where("status = ?", *filter.Status)
	} else {
		query = query.Where("status <> ?", models.ActivityStatusDraft)
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
			RejectReason:     activity.RejectReason,
			TicketSpecCount:  ticketSpecCountMap[activity.ID],
			CreatedAt:        activity.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:        activity.UpdatedAt.Format("2006-01-02 15:04:05"),
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

	detail := &types.AdminActivityDetail{Activity: activity}
	var organizer models.Organizer
	if err := s.DB.WithContext(ctx).Where("id = ?", activity.OrganizerID).First(&organizer).Error; err == nil {
		detail.Organizer = &organizer
	}
	if err := s.DB.WithContext(ctx).Where("activity_id = ?", activity.ID).Order("id ASC").Find(&detail.TicketSpecs).Error; err != nil {
		return nil, err
	}
	return detail, nil
}

func (s *AdminService) AuditActivity(ctx context.Context, activityID int64, req types.AdminAuditActivityRequest) error {
	if req.Status != models.ActivityStatusOnline && req.Status != models.ActivityStatusRejected {
		return errors.New("审核状态无效，仅支持 3通过上架 或 4拒绝")
	}
	if req.Status == models.ActivityStatusRejected && req.RejectReason == "" {
		return errors.New("拒绝时必须填写 reject_reason")
	}
	if req.Status == models.ActivityStatusOnline {
		var activity models.Activity
		if err := s.DB.WithContext(ctx).Where("id = ?", activityID).First(&activity).Error; err != nil {
			return err
		}
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

func (s *AdminService) GetTicketOrderList(ctx context.Context, page, pageSize int, activityID int64, status *int8, keyword string) (*types.AdminTicketOrderListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	query := s.DB.WithContext(ctx).Model(&models.TicketOrder{})
	if activityID > 0 {
		query = query.Where("activity_id = ?", activityID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("order_no LIKE ? OR buyer_name LIKE ? OR buyer_id_card LIKE ?", like, like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var orders []models.TicketOrder
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&orders).Error; err != nil {
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
	return &types.AdminTicketOrderDetail{Order: items[0], Refunds: refunds, RefundLogs: logs}, nil
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
		Joins("LEFT JOIN refunds r ON r.order_id = o.id AND r.status = ?", models.RefundStatusSuccess)
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
	if err := base.Select("org.id AS organizer_id, org.name AS organizer_name, COALESCE(SUM(o.actual_price),0) AS gross_amount, COALESCE(SUM(r.refund_amount),0) AS refund_amount, COALESCE(SUM(o.actual_price),0) - COALESCE(SUM(r.refund_amount),0) AS settle_amount, COUNT(o.id) AS order_count").
		Group("org.id, org.name").
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
	query := s.DB.WithContext(ctx).Model(&models.Users{})
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("nickname LIKE ? OR mobile LIKE ?", like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var users []models.Users
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
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
	list := make([]types.AdminUserItem, 0, len(users))
	for _, user := range users {
		item := types.AdminUserItem{ID: user.Id, Nickname: user.Nickname, Avatar: user.Avatar, Mobile: user.Mobile, Status: user.Status, CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05")}
		if stat, ok := stats[user.Id]; ok {
			item.TotalAmount = stat.Amount
			item.OrderCount = stat.Count
		}
		list = append(list, item)
	}
	return &types.AdminUserListResponse{List: list, Total: total, Page: page, PageSize: pageSize}, nil
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
		item := types.AdminTicketOrderItem{OrderNo: order.OrderNo, Status: order.Status, TotalPrice: order.TotalPrice, ActualPrice: order.ActualPrice, Quantity: order.Quantity, UserID: order.UserID, BuyerName: order.BuyerName, BuyerIDCard: order.BuyerIDCard, ActivityID: order.ActivityID, TicketSpecID: order.TicketSpecID, PayMethod: order.PayMethod, PayTime: formatAdminPtrTime(order.PayTime), CreatedAt: order.CreatedAt.Format("2006-01-02 15:04:05")}
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
