package service

import (
	"Hyper/models"
	"Hyper/pkg/encrypt"
	"Hyper/types"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *AdminService) GetAdminProfile(ctx context.Context, adminID int64) (*types.AdminProfileResponse, error) {
	var admin models.Admin
	if err := s.DB.WithContext(ctx).First(&admin, adminID).Error; err != nil {
		return nil, err
	}
	return &types.AdminProfileResponse{ID: admin.Id, Username: admin.Username, Avatar: admin.Avatar, Mobile: admin.Mobile, Email: admin.Email, Motto: admin.Motto, Status: admin.Status, CreatedAt: admin.CreatedAt, UpdatedAt: admin.UpdatedAt}, nil
}

func (s *AdminService) UpdateAdminProfile(ctx context.Context, adminID int64, req types.AdminProfileRequest) error {
	return s.DB.WithContext(ctx).Model(&models.Admin{}).Where("id = ?", adminID).Updates(map[string]any{
		"avatar": req.Avatar,
		"mobile": req.Mobile,
		"email":  req.Email,
		"motto":  req.Motto,
	}).Error
}

func (s *AdminService) UpdateAdminPassword(ctx context.Context, adminID int64, req types.AdminPasswordRequest) error {
	var admin models.Admin
	if err := s.DB.WithContext(ctx).First(&admin, adminID).Error; err != nil {
		return err
	}
	if !encrypt.VerifyPassword(admin.Password, req.OldPassword) {
		return errors.New("原密码错误")
	}
	return s.DB.WithContext(ctx).Model(&admin).Update("password", encrypt.HashPassword(req.NewPassword)).Error
}

func (s *AdminService) ListAdmins(ctx context.Context, page, pageSize int, keyword string) (*types.AdminPageResponse[models.Admin], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	query := s.DB.WithContext(ctx).Model(&models.Admin{})
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("username LIKE ? OR mobile LIKE ? OR email LIKE ?", like, like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []models.Admin
	if err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	for i := range list {
		list[i].Password = ""
	}
	return &types.AdminPageResponse[models.Admin]{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *AdminService) CreateAdmin(ctx context.Context, req types.AdminAccountRequest) (int64, error) {
	if req.Password == "" {
		return 0, errors.New("密码不能为空")
	}
	if req.Status == 0 {
		req.Status = models.AdminStatusNormal
	}
	admin := models.Admin{Username: req.Username, Password: encrypt.HashPassword(req.Password), Avatar: req.Avatar, Mobile: req.Mobile, Email: req.Email, Status: req.Status}
	if err := s.DB.WithContext(ctx).Create(&admin).Error; err != nil {
		return 0, err
	}
	return int64(admin.Id), nil
}

func (s *AdminService) UpdateAdmin(ctx context.Context, id int64, req types.AdminAccountRequest) error {
	updates := map[string]any{"username": req.Username, "avatar": req.Avatar, "mobile": req.Mobile, "email": req.Email}
	if req.Status != 0 {
		updates["status"] = req.Status
	}
	if req.Password != "" {
		updates["password"] = encrypt.HashPassword(req.Password)
	}
	return s.DB.WithContext(ctx).Model(&models.Admin{}).Where("id = ?", id).Updates(updates).Error
}

func (s *AdminService) DeleteAdmin(ctx context.Context, id int64) error {
	return s.DB.WithContext(ctx).Delete(&models.Admin{}, id).Error
}

func (s *AdminService) ListRoles(ctx context.Context, page, pageSize int, keyword string) (*types.AdminPageResponse[models.AdminRole], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	query := s.DB.WithContext(ctx).Model(&models.AdminRole{})
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []models.AdminRole
	if err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	return &types.AdminPageResponse[models.AdminRole]{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *AdminService) SaveRole(ctx context.Context, id int64, req types.AdminRoleRequest) (int64, error) {
	if req.Status == 0 {
		req.Status = 1
	}
	role := models.AdminRole{Name: req.Name, Description: req.Description, Permissions: req.Permissions, Status: req.Status}
	if id > 0 {
		return id, s.DB.WithContext(ctx).Model(&models.AdminRole{}).Where("id = ?", id).Updates(role).Error
	}
	if err := s.DB.WithContext(ctx).Create(&role).Error; err != nil {
		return 0, err
	}
	return role.ID, nil
}

func (s *AdminService) DeleteRole(ctx context.Context, id int64) error {
	return s.DB.WithContext(ctx).Delete(&models.AdminRole{}, id).Error
}

func (s *AdminService) ListOperationLogs(ctx context.Context, page, pageSize int, adminID int64, keyword string) (*types.AdminPageResponse[models.AdminOperationLog], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	query := s.DB.WithContext(ctx).Model(&models.AdminOperationLog{})
	if adminID > 0 {
		query = query.Where("admin_id = ?", adminID)
	}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("action LIKE ? OR resource LIKE ? OR path LIKE ?", like, like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []models.AdminOperationLog
	if err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	return &types.AdminPageResponse[models.AdminOperationLog]{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *AdminService) ListCategories(ctx context.Context, page, pageSize int, categoryType string) (*types.AdminPageResponse[models.AdminCategory], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	query := s.DB.WithContext(ctx).Model(&models.AdminCategory{})
	if categoryType != "" {
		query = query.Where("type = ?", categoryType)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []models.AdminCategory
	if err := query.Order("sort asc,id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	return &types.AdminPageResponse[models.AdminCategory]{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *AdminService) SaveCategory(ctx context.Context, id int64, req types.AdminCategoryRequest) (int64, error) {
	if req.Status == 0 {
		req.Status = 1
	}
	row := models.AdminCategory{Type: req.Type, Name: req.Name, Image: req.Image, Value: req.Value, Sort: req.Sort, Status: req.Status}
	if id > 0 {
		return id, s.DB.WithContext(ctx).Model(&models.AdminCategory{}).Where("id = ?", id).Updates(row).Error
	}
	if err := s.DB.WithContext(ctx).Create(&row).Error; err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (s *AdminService) DeleteCategory(ctx context.Context, id int64) error {
	return s.DB.WithContext(ctx).Delete(&models.AdminCategory{}, id).Error
}

func (s *AdminService) ListUserRecords(ctx context.Context, userID int64, recordType string, page, pageSize int) (*types.AdminPageResponse[map[string]any], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	switch recordType {
	case "likes":
		return adminMapPage(s.DB.WithContext(ctx).Table("note_likes nl").Select("nl.*, n.title AS note_title").Joins("LEFT JOIN notes n ON n.id = nl.note_id").Where("nl.user_id = ? AND nl.status = 1", userID).Order("nl.id desc"), page, pageSize)
	case "collections":
		return adminMapPage(s.DB.WithContext(ctx).Table("note_collections nc").Select("nc.*, n.title AS note_title").Joins("LEFT JOIN notes n ON n.id = nc.note_id").Where("nc.user_id = ? AND nc.status = 1", userID).Order("nc.id desc"), page, pageSize)
	case "following":
		return adminMapPage(s.DB.WithContext(ctx).Table("user_follow uf").Select("uf.*, u.nickname, u.avatar").Joins("LEFT JOIN users u ON u.id = uf.followee_id").Where("uf.follower_id = ? AND uf.status = 1", userID).Order("uf.id desc"), page, pageSize)
	case "followers":
		return adminMapPage(s.DB.WithContext(ctx).Table("user_follow uf").Select("uf.*, u.nickname, u.avatar").Joins("LEFT JOIN users u ON u.id = uf.follower_id").Where("uf.followee_id = ? AND uf.status = 1", userID).Order("uf.id desc"), page, pageSize)
	case "attends":
		return adminMapPage(s.DB.WithContext(ctx).Table("party_attends pa").Where("pa.user_id = ?", userID).Order("pa.id desc"), page, pageSize)
	case "subscribes":
		return adminMapPage(s.DB.WithContext(ctx).Table("subscribes s").Where("s.user_id = ?", userID).Order("s.id desc"), page, pageSize)
	default:
		return nil, errors.New("记录类型无效")
	}
}

func (s *AdminService) ListViewers(ctx context.Context, page, pageSize int, keyword string) (*types.AdminPageResponse[map[string]any], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	query := s.DB.WithContext(ctx).Table("viewers v").Select("v.*, u.nickname AS user_name, u.mobile AS user_mobile").Joins("LEFT JOIN users u ON u.id = v.user_id")
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("v.real_name LIKE ? OR v.id_card LIKE ? OR v.phone LIKE ? OR u.nickname LIKE ? OR u.mobile LIKE ?", like, like, like, like, like)
	}
	return adminMapPage(query.Order("v.id desc"), page, pageSize)
}

func (s *AdminService) ListActivityCollections(ctx context.Context, page, pageSize int, keyword string, organizerID int64) (*types.AdminPageResponse[types.ActivityCollectionItem], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	query := s.DB.WithContext(ctx).Table("activity_collections ac").Joins("LEFT JOIN organizers o ON o.id = ac.organizer_id")
	if organizerID > 0 {
		query = query.Where("ac.organizer_id = ?", organizerID)
	}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		query = query.Where("ac.title LIKE ? OR o.name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []types.ActivityCollectionItem
	err := query.Select("ac.*, o.name AS organizer_name, (SELECT COUNT(1) FROM activity_collection_items ai WHERE ai.collection_id = ac.id) AS activity_count").
		Order("ac.id desc").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&list).Error
	return &types.AdminPageResponse[types.ActivityCollectionItem]{List: list, Total: total, Page: page, PageSize: pageSize}, err
}

func (s *AdminService) SaveActivityCollection(ctx context.Context, id int64, req types.ActivityCollectionRequest) (int64, error) {
	if req.Status == 0 {
		req.Status = 1
	}
	var savedID int64
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row := models.ActivityCollection{OrganizerID: req.OrganizerID, Title: req.Title, ShareTitle: req.ShareTitle, Description: req.Description, ShareImage: req.ShareImage, Status: req.Status}
		if id > 0 {
			savedID = id
			if err := tx.Model(&models.ActivityCollection{}).Where("id = ?", id).Updates(row).Error; err != nil {
				return err
			}
			if err := tx.Where("collection_id = ?", id).Delete(&models.ActivityCollectionItem{}).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			savedID = row.ID
		}
		for i, activityID := range req.ActivityIDs {
			if err := tx.Create(&models.ActivityCollectionItem{CollectionID: savedID, ActivityID: activityID, Sort: i + 1}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return savedID, err
}

func (s *AdminService) DeleteActivityCollection(ctx context.Context, id int64) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("collection_id = ?", id).Delete(&models.ActivityCollectionItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.ActivityCollection{}, id).Error
	})
}

func (s *AdminService) ListVerifiers(ctx context.Context, page, pageSize int, keyword string, organizerID int64) (*types.AdminPageResponse[map[string]any], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	query := s.DB.WithContext(ctx).Table("verifiers v").Select("v.*, o.name AS organizer_name, u.nickname AS user_name, u.mobile AS user_mobile").Joins("LEFT JOIN organizers o ON o.id = v.organizer_id").Joins("LEFT JOIN users u ON u.id = v.user_id")
	if organizerID > 0 {
		query = query.Where("v.organizer_id = ?", organizerID)
	}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("v.name LIKE ? OR v.phone LIKE ? OR o.name LIKE ? OR u.nickname LIKE ?", like, like, like, like)
	}
	return adminMapPage(query.Order("v.id desc"), page, pageSize)
}

func (s *AdminService) ListVerificationRecords(ctx context.Context, page, pageSize int, keyword string, organizerID int64) (*types.AdminPageResponse[map[string]any], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	query := s.DB.WithContext(ctx).Table("verification_records vr").
		Select("vr.*, v.name AS verifier_name, v.phone AS verifier_phone, o.name AS organizer_name, tor.order_no, a.name AS activity_name").
		Joins("LEFT JOIN verifiers v ON v.id = vr.verifier_id").
		Joins("LEFT JOIN organizers o ON o.id = v.organizer_id").
		Joins("LEFT JOIN ticket_orders tor ON tor.id = vr.order_id").
		Joins("LEFT JOIN activities a ON a.id = vr.activity_id")
	if organizerID > 0 {
		query = query.Where("v.organizer_id = ?", organizerID)
	}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("v.name LIKE ? OR v.phone LIKE ? OR o.name LIKE ? OR tor.order_no LIKE ? OR a.name LIKE ?", like, like, like, like, like)
	}
	return adminMapPage(query.Order("vr.id desc"), page, pageSize)
}

func (s *AdminService) ListNotes(ctx context.Context, page, pageSize int, status *int, keyword string) (*types.AdminPageResponse[map[string]any], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	query := s.DB.WithContext(ctx).Table("notes n").Select("n.*, u.nickname, u.avatar, ns.like_count, ns.collect_count, ns.comment_count, ns.share_count").Joins("LEFT JOIN users u ON u.id = n.user_id").Joins("LEFT JOIN note_stats ns ON ns.note_id = n.id")
	if status != nil {
		query = query.Where("n.status = ?", *status)
	}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		query = query.Where("n.title LIKE ? OR n.content LIKE ? OR u.nickname LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	return adminMapPage(query.Order("n.id desc"), page, pageSize)
}

func (s *AdminService) UpdateNoteStatus(ctx context.Context, noteID int64, status int) error {
	return s.DB.WithContext(ctx).Model(&models.Note{}).Where("id = ?", noteID).Update("status", status).Error
}

func (s *AdminService) ListNoteRecords(ctx context.Context, recordType string, page, pageSize int, noteID int64) (*types.AdminPageResponse[map[string]any], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	switch recordType {
	case "likes":
		return adminMapPage(s.DB.WithContext(ctx).Table("note_likes nl").Select("nl.*, u.nickname, u.avatar").Joins("LEFT JOIN users u ON u.id = nl.user_id").Where("nl.note_id = ? AND nl.status = 1", noteID).Order("nl.id desc"), page, pageSize)
	case "collections":
		return adminMapPage(s.DB.WithContext(ctx).Table("note_collections nc").Select("nc.*, u.nickname, u.avatar").Joins("LEFT JOIN users u ON u.id = nc.user_id").Where("nc.note_id = ? AND nc.status = 1", noteID).Order("nc.id desc"), page, pageSize)
	case "comments":
		return adminMapPage(s.DB.WithContext(ctx).Table("comments c").Select("c.*, u.nickname, u.avatar").Joins("LEFT JOIN users u ON u.id = c.user_id").Where("c.note_id = ?", noteID).Order("c.id desc"), page, pageSize)
	case "shares":
		return adminMapPage(s.DB.WithContext(ctx).Table("note_stats ns").Where("ns.note_id = ?", noteID), page, pageSize)
	default:
		return nil, errors.New("记录类型无效")
	}
}

func (s *AdminService) ListPointLogs(ctx context.Context, page, pageSize int, userID int64) (*types.AdminPageResponse[map[string]any], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	query := s.DB.WithContext(ctx).Table("point_logs pl").Select("pl.*, u.nickname, u.mobile").Joins("LEFT JOIN users u ON u.id = pl.user_id")
	if userID > 0 {
		query = query.Where("pl.user_id = ?", userID)
	}
	return adminMapPage(query.Order("pl.id desc"), page, pageSize)
}

func (s *AdminService) ListWithdraws(ctx context.Context, page, pageSize int, status *int8, organizerID int64) (*types.AdminPageResponse[map[string]any], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	query := s.DB.WithContext(ctx).Table("organizer_withdraws w").Select("w.*, o.name AS organizer_name").Joins("LEFT JOIN organizers o ON o.id = w.organizer_id")
	if status != nil {
		query = query.Where("w.status = ?", *status)
	}
	if organizerID > 0 {
		query = query.Where("w.organizer_id = ?", organizerID)
	}
	return adminMapPage(query.Order("w.id desc"), page, pageSize)
}

func (s *AdminService) AuditWithdraw(ctx context.Context, id int64, req types.WithdrawAuditRequest) error {
	return s.DB.WithContext(ctx).Model(&models.OrganizerWithdraw{}).Where("id = ?", id).Updates(map[string]any{"status": req.Status, "remark": req.Remark}).Error
}

func (s *AdminService) ListBankAccountAudits(ctx context.Context, page, pageSize int, status *int8, organizerID int64) (*types.AdminPageResponse[map[string]any], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	query := s.DB.WithContext(ctx).Table("organizer_bank_account_audits a").
		Select("a.*, o.name AS organizer_name, o.type AS organizer_type, u.nickname AS user_name, u.mobile AS user_mobile").
		Joins("LEFT JOIN organizers o ON o.id = a.organizer_id").
		Joins("LEFT JOIN users u ON u.id = a.user_id")
	if status != nil {
		query = query.Where("a.status = ?", *status)
	}
	if organizerID > 0 {
		query = query.Where("a.organizer_id = ?", organizerID)
	}
	return adminMapPage(query.Order("a.id desc"), page, pageSize)
}

func (s *AdminService) AuditBankAccount(ctx context.Context, id int64, req types.BankAccountAuditRequest) error {
	if req.Status != models.OrganizerBankAuditStatusApproved && req.Status != models.OrganizerBankAuditStatusRejected {
		return errors.New("审核状态无效，仅支持 1通过 或 2拒绝")
	}
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var audit models.OrganizerBankAccountAudit
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&audit).Error; err != nil {
			return err
		}
		if audit.Status != models.OrganizerBankAuditStatusPending {
			return errors.New("该收款账户申请已审核")
		}
		now := time.Now()
		updates := map[string]any{
			"status":        req.Status,
			"reject_reason": req.RejectReason,
			"reviewed_at":   now,
			"updated_at":    now,
		}
		if err := tx.Model(&audit).Updates(updates).Error; err != nil {
			return err
		}
		if req.Status == models.OrganizerBankAuditStatusApproved {
			if err := tx.Model(&models.Organizer{}).Where("id = ?", audit.OrganizerID).Updates(map[string]any{
				"bank_account_name": audit.BankAccountName,
				"bank_account_no":   audit.BankAccountNo,
				"bank_name":         audit.BankName,
				"updated_at":        now,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *AdminService) ListMessages(ctx context.Context, page, pageSize int) (*types.AdminPageResponse[models.PlatformMessage], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	var total int64
	query := s.DB.WithContext(ctx).Model(&models.PlatformMessage{})
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []models.PlatformMessage
	err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return &types.AdminPageResponse[models.PlatformMessage]{List: list, Total: total, Page: page, PageSize: pageSize}, err
}

func (s *AdminService) CreateMessage(ctx context.Context, req types.PlatformMessageRequest) (int64, error) {
	msg := models.PlatformMessage{Title: req.Title, Content: req.Content, Type: req.Type, Target: req.Target, Status: req.Status}
	if err := s.DB.WithContext(ctx).Create(&msg).Error; err != nil {
		return 0, err
	}
	if req.Status == 1 {
		if err := s.publishPlatformMessage(ctx, msg, req); err != nil {
			return msg.ID, err
		}
	}
	return msg.ID, nil
}

func (s *AdminService) publishPlatformMessage(ctx context.Context, msg models.PlatformMessage, req types.PlatformMessageRequest) error {
	if s.MqProducer == nil {
		return errors.New("消息队列未初始化")
	}
	targetIDs, err := s.resolvePlatformMessageTargets(ctx, req)
	if err != nil {
		return err
	}
	for _, targetID := range targetIDs {
		payload := types.PlatformMessagePayload{
			MessageID: msg.ID,
			TargetID:  targetID,
			Title:     msg.Title,
			Content:   msg.Content,
			Type:      msg.Type,
			Target:    msg.Target,
			CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		}
		data, _ := json.Marshal(payload)
		event := types.SystemMessage{Type: "platform_message", Data: json.RawMessage(data)}
		body, _ := json.Marshal(event)
		if _, err := s.MqProducer.Send(ctx, &rmq_client.Message{Topic: types.SystemMessageTopic, Body: body}); err != nil {
			return err
		}
	}
	return nil
}

func (s *AdminService) resolvePlatformMessageTargets(ctx context.Context, req types.PlatformMessageRequest) ([]int64, error) {
	seen := make(map[int64]struct{})
	add := func(id int64) {
		if id > 0 {
			seen[id] = struct{}{}
		}
	}
	for _, id := range req.TargetUserIDs {
		add(id)
	}
	if len(req.OrganizerIDs) > 0 {
		var organizers []models.Organizer
		if err := s.DB.WithContext(ctx).Where("id IN ?", req.OrganizerIDs).Find(&organizers).Error; err != nil {
			return nil, err
		}
		for _, organizer := range organizers {
			add(organizer.UserID)
		}
	}
	switch strings.TrimSpace(req.Target) {
	case "all":
		var users []models.Users
		if err := s.DB.WithContext(ctx).Select("id").Where("status = ?", 1).Find(&users).Error; err != nil {
			return nil, err
		}
		for _, user := range users {
			add(int64(user.Id))
		}
	case "merchant", "organizer":
		if len(req.OrganizerIDs) == 0 {
			var organizers []models.Organizer
			if err := s.DB.WithContext(ctx).Select("user_id").Where("status = ?", models.OrganizerStatusApproved).Find(&organizers).Error; err != nil {
				return nil, err
			}
			for _, organizer := range organizers {
				add(organizer.UserID)
			}
		}
	case "user", "":
	default:
		return nil, errors.New("消息目标类型无效")
	}
	ids := make([]int64, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, errors.New("没有可推送的目标用户")
	}
	return ids, nil
}

func normalizeAdminPage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func adminMapPage(query *gorm.DB, page, pageSize int) (*types.AdminPageResponse[map[string]any], error) {
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []map[string]any
	if err := query.Session(&gorm.Session{}).Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, err
	}
	return &types.AdminPageResponse[map[string]any]{List: rows, Total: total, Page: page, PageSize: pageSize}, nil
}

func upsertByID[T any](tx *gorm.DB, id int64, row *T) (int64, error) {
	if id > 0 {
		return id, tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(row).Error
	}
	return 0, fmt.Errorf("unsupported")
}
