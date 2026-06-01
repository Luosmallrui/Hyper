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

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IAdminService interface {
	Login(ctx context.Context, username, password string) (string, string, error)
	GetOrganizerList(ctx context.Context, page, pageSize int, status *int8, organizerType string) (*types.AdminOrganizerListResponse, error)
	GetOrganizerDetail(ctx context.Context, organizerID int64) (*types.AdminOrganizerDetail, error)
	AuditOrganizer(ctx context.Context, organizerID int64, req types.AdminAuditOrganizerRequest) error
	BindAdminWechatSubscriber(ctx context.Context, adminID int64, code string) error
	GetPartyList(ctx context.Context, page, pageSize int, keyword, partyType string) (*types.AdminPartyListResponse, error)
	GetPartyDetail(ctx context.Context, partyID int64) (*types.AdminPartyDetail, error)
	UpdatePartyStatus(ctx context.Context, partyID int64, status string) error
	GetEventTicketList(ctx context.Context, eventID int64, page, pageSize int) (*types.AdminTicketListResponse, error)
	GetAllTickets(ctx context.Context, page, pageSize int, keyword string) (*types.AdminTicketListResponse, error)
	GetOrderList(ctx context.Context, page, pageSize int, eventID int64) (*types.AdminOrderListResponse, error)
	GetDashboardStats(ctx context.Context) (*types.AdminDashboardStats, error)
}

type AdminService struct {
	AdminDAO      *dao.Admin
	DB            *gorm.DB
	Secret        []byte
	WeChatService IWeChatService
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
