package service

import (
	"Hyper/models"
	"Hyper/pkg/encrypt"
	"Hyper/types"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
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
	resp := &types.AdminProfileResponse{ID: admin.Id, Username: admin.Username, Avatar: admin.Avatar, Mobile: admin.Mobile, Email: admin.Email, Motto: admin.Motto, RoleID: admin.RoleID, Status: admin.Status, Permissions: []string{}, CreatedAt: admin.CreatedAt, UpdatedAt: admin.UpdatedAt}
	if admin.RoleID == 0 {
		return resp, nil
	}
	var role models.AdminRole
	if err := s.DB.WithContext(ctx).Where("id = ?", admin.RoleID).First(&role).Error; err != nil {
		return nil, err
	}
	permissions, err := normalizeAdminPermissions(role.Permissions)
	if err != nil {
		return nil, err
	}
	resp.RoleName = role.Name
	resp.Permissions = permissions
	return resp, nil
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

func (s *AdminService) ListAdmins(ctx context.Context, page, pageSize int, keyword string) (*types.AdminPageResponse[types.AdminAccountItem], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	query := s.DB.WithContext(ctx).Table("admin a").Joins("LEFT JOIN admin_roles r ON r.id = a.role_id")
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("a.username LIKE ? OR a.mobile LIKE ? OR a.email LIKE ?", like, like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []struct {
		ID                              int
		Username, Avatar, Mobile, Email string
		RoleID                          int64
		Status                          int8
		CreatedAt, UpdatedAt            time.Time
		RoleName                        string
		Permissions                     string
	}
	if err := query.Select("a.id, a.username, a.avatar, a.mobile, a.email, a.role_id, a.status, a.created_at, a.updated_at, r.name AS role_name, r.permissions").Order("a.id desc").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]types.AdminAccountItem, 0, len(rows))
	for _, row := range rows {
		permissions, _ := normalizeAdminPermissions(row.Permissions)
		list = append(list, types.AdminAccountItem{ID: row.ID, Username: row.Username, Avatar: row.Avatar, Mobile: row.Mobile, Email: row.Email, RoleID: row.RoleID, Role: types.AdminRoleSummary{ID: row.RoleID, Name: row.RoleName, Permissions: permissions}, Status: row.Status, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt})
	}
	return &types.AdminPageResponse[types.AdminAccountItem]{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *AdminService) CreateAdmin(ctx context.Context, req types.AdminAccountRequest) (int64, error) {
	if req.Password == "" {
		return 0, errors.New("密码不能为空")
	}
	if req.Status == 0 {
		req.Status = models.AdminStatusNormal
	}
	if err := s.validateAdminRole(ctx, req.RoleID); err != nil {
		return 0, err
	}
	admin := models.Admin{Username: req.Username, Password: encrypt.HashPassword(req.Password), Avatar: req.Avatar, Mobile: req.Mobile, Email: req.Email, RoleID: req.RoleID, Status: req.Status}
	if err := s.DB.WithContext(ctx).Create(&admin).Error; err != nil {
		return 0, err
	}
	return int64(admin.Id), nil
}

func (s *AdminService) UpdateAdmin(ctx context.Context, actorID, id int64, req types.AdminAccountRequest) error {
	if err := s.validateAdminRole(ctx, req.RoleID); err != nil {
		return err
	}
	if actorID == id && req.Status != 0 && req.Status != models.AdminStatusNormal {
		return errors.New("不能停用自己的管理员账号")
	}
	if err := s.ensureAdminChangeKeepsSuperAdmin(ctx, id, req.RoleID, req.Status); err != nil {
		return err
	}
	updates := map[string]any{"username": req.Username, "avatar": req.Avatar, "mobile": req.Mobile, "email": req.Email, "role_id": req.RoleID}
	if req.Status != 0 {
		updates["status"] = req.Status
	}
	if req.Password != "" {
		updates["password"] = encrypt.HashPassword(req.Password)
	}
	return s.DB.WithContext(ctx).Model(&models.Admin{}).Where("id = ?", id).Updates(updates).Error
}

func (s *AdminService) validateAdminRole(ctx context.Context, roleID int64) error {
	if roleID == 0 {
		return errors.New("管理员必须分配启用角色")
	}
	var count int64
	if err := s.DB.WithContext(ctx).Model(&models.AdminRole{}).Where("id = ? AND status = 1", roleID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("管理员角色不存在或已停用")
	}
	return nil
}

func (s *AdminService) DeleteAdmin(ctx context.Context, actorID, id int64) error {
	if actorID == id {
		return errors.New("不能删除自己的管理员账号")
	}
	if err := s.ensureAdminChangeKeepsSuperAdmin(ctx, id, -1, models.AdminStatusDeactivate); err != nil {
		return err
	}
	return s.DB.WithContext(ctx).Delete(&models.Admin{}, id).Error
}

func (s *AdminService) ListRoles(ctx context.Context, page, pageSize int, keyword string) (*types.AdminPageResponse[types.AdminRoleItem], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	query := s.DB.WithContext(ctx).Table("admin_roles r").Joins("LEFT JOIN admin a ON a.role_id = r.id")
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		query = query.Where("r.name LIKE ?", "%"+keyword+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []struct {
		ID                             int64
		Name, Description, Permissions string
		Status                         int8
		MemberCount                    int64
		CreatedAt, UpdatedAt           time.Time
	}
	if err := query.Select("r.id, r.name, r.description, r.permissions, r.status, r.created_at, r.updated_at, COUNT(a.id) AS member_count").Group("r.id, r.name, r.description, r.permissions, r.status, r.created_at, r.updated_at").Order("r.id desc").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]types.AdminRoleItem, 0, len(rows))
	for _, row := range rows {
		permissions, _ := normalizeAdminPermissions(row.Permissions)
		list = append(list, types.AdminRoleItem{ID: row.ID, Name: row.Name, Description: row.Description, Permissions: permissions, Status: row.Status, MemberCount: row.MemberCount, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt})
	}
	return &types.AdminPageResponse[types.AdminRoleItem]{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *AdminService) SaveRole(ctx context.Context, id int64, req types.AdminRoleRequest) (int64, error) {
	permissions, err := normalizeAdminPermissions(req.Permissions)
	if err != nil {
		return 0, err
	}
	if id == 0 && req.Status == 0 {
		req.Status = 1
	}
	if req.Status != 0 && req.Status != 1 {
		return 0, errors.New("角色状态仅支持 0禁用 或 1启用")
	}
	if id > 0 {
		if err := s.ensureRoleChangeKeepsSuperAdmin(ctx, id, permissions, req.Status); err != nil {
			return 0, err
		}
	}
	encoded, _ := json.Marshal(permissions)
	role := models.AdminRole{Name: req.Name, Description: req.Description, Permissions: string(encoded), Status: req.Status}
	if id > 0 {
		return id, s.DB.WithContext(ctx).Model(&models.AdminRole{}).Where("id = ?", id).Updates(role).Error
	}
	if err := s.DB.WithContext(ctx).Create(&role).Error; err != nil {
		return 0, err
	}
	return role.ID, nil
}

func (s *AdminService) DeleteRole(ctx context.Context, id int64) error {
	var members int64
	if err := s.DB.WithContext(ctx).Model(&models.Admin{}).Where("role_id = ?", id).Count(&members).Error; err != nil {
		return err
	}
	if members > 0 {
		return errors.New("该角色仍有关联管理员，请先迁移成员")
	}
	if err := s.ensureRoleChangeKeepsSuperAdmin(ctx, id, nil, models.AdminStatusDeactivate); err != nil {
		return err
	}
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

func (s *AdminService) RecordOperationLog(ctx context.Context, item models.AdminOperationLog) error {
	return s.DB.WithContext(ctx).Create(&item).Error
}

func (s *AdminService) CheckPermission(ctx context.Context, adminID int64, method, path string) error {
	var admin models.Admin
	if err := s.DB.WithContext(ctx).First(&admin, adminID).Error; err != nil {
		return err
	}
	if admin.Status != models.AdminStatusNormal {
		return errors.New("管理员账号已停用")
	}
	if isAdminSelfServiceRoute(method, path) {
		return nil
	}
	requiredPermission := RequiredAdminPermission(method, path)
	if requiredPermission == "" {
		return errors.New("管理接口未配置权限")
	}
	if admin.RoleID == 0 {
		return fmt.Errorf("无权执行该操作：需要 %s", requiredPermission)
	}
	var role models.AdminRole
	if err := s.DB.WithContext(ctx).Where("id = ? AND status = 1", admin.RoleID).First(&role).Error; err != nil {
		return errors.New("管理员角色不存在或已停用")
	}
	permissions, err := normalizeAdminPermissions(role.Permissions)
	if err != nil {
		return errors.New("管理员角色权限配置无效")
	}
	for _, permission := range permissions {
		if permission == "*" || permission == requiredPermission {
			return nil
		}
	}
	return fmt.Errorf("无权执行该操作：需要 %s", requiredPermission)
}

var adminPermissionWhitelist = map[string]struct{}{
	"admin.dashboard": {}, "admin.system": {}, "admin.users": {}, "admin.merchants": {},
	"admin.organizers": {}, "admin.activities": {}, "admin.tickets": {}, "admin.orders": {},
	"admin.verifications": {}, "admin.content": {}, "admin.finance": {},
}

func normalizeAdminPermissions(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{}, nil
	}
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		values = strings.Split(raw, ",")
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if value != "*" {
			if _, ok := adminPermissionWhitelist[value]; !ok {
				return nil, fmt.Errorf("无效的权限值：%s", value)
			}
		}
		if _, ok := seen[value]; !ok {
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}
	_, hasStar := seen["*"]
	if len(result) > 1 && hasStar {
		return nil, errors.New("超级管理员权限只能单独使用 *")
	}
	return result, nil
}

func isAdminSelfServiceRoute(method, path string) bool {
	return ((method == "GET" || method == "PUT") && path == "/v1/admin/profile") || (method == "PUT" && path == "/v1/admin/profile/password")
}

// RequiredAdminPermission is the single route-to-module mapping used by the RBAC middleware.
func RequiredAdminPermission(method, path string) string {
	if (method == "POST" || method == "PUT" || method == "DELETE") && (strings.HasPrefix(path, "/v1/admin/admins") || strings.HasPrefix(path, "/v1/admin/roles")) {
		return "*"
	}
	switch {
	case path == "/v1/admin/dashboard":
		return "admin.dashboard"
	case strings.HasPrefix(path, "/v1/admin/settings"), strings.HasPrefix(path, "/v1/admin/categories"), strings.HasPrefix(path, "/v1/admin/logs"), strings.HasPrefix(path, "/v1/admin/admins"), strings.HasPrefix(path, "/v1/admin/roles"), strings.HasPrefix(path, "/v1/admin/wechat-subscribe"):
		return "admin.system"
	case strings.HasPrefix(path, "/v1/admin/users"), strings.HasPrefix(path, "/v1/admin/viewers"):
		return "admin.users"
	case strings.HasPrefix(path, "/v1/admin/organizers"):
		return "admin.organizers"
	case strings.HasPrefix(path, "/v1/admin/activities"), strings.HasPrefix(path, "/v1/admin/parties"), strings.HasPrefix(path, "/v1/admin/activity-collections"):
		return "admin.activities"
	case strings.HasPrefix(path, "/v1/admin/tickets") || strings.HasPrefix(path, "/v1/admin/events/"):
		return "admin.tickets"
	case strings.HasPrefix(path, "/v1/admin/orders"):
		return "admin.orders"
	case strings.HasPrefix(path, "/v1/admin/verifiers"), strings.HasPrefix(path, "/v1/admin/verification-records"):
		return "admin.verifications"
	case strings.HasPrefix(path, "/v1/admin/notes"), strings.HasPrefix(path, "/v1/admin/messages"), strings.HasPrefix(path, "/v1/admin/banners"):
		return "admin.content"
	case strings.HasPrefix(path, "/v1/admin/finance"), strings.HasPrefix(path, "/v1/admin/withdraws"), strings.HasPrefix(path, "/v1/admin/bank-account-audits"), strings.HasPrefix(path, "/v1/admin/points"):
		return "admin.finance"
	case strings.HasPrefix(path, "/v1/admin/organizer-level-rules"):
		return "admin.merchants"
	default:
		return "*"
	}
}

func (s *AdminService) ensureAdminChangeKeepsSuperAdmin(ctx context.Context, targetID, newRoleID int64, newStatus int8) error {
	var target models.Admin
	if err := s.DB.WithContext(ctx).First(&target, targetID).Error; err != nil {
		return err
	}
	if !s.isAdminSuper(ctx, target) {
		return nil
	}
	if newStatus == 0 {
		newStatus = target.Status
	}
	if newStatus == models.AdminStatusNormal && s.isRoleSuper(ctx, newRoleID) {
		return nil
	}
	var count int64
	if err := s.DB.WithContext(ctx).Table("admin a").Joins("JOIN admin_roles r ON r.id = a.role_id").Where("a.id <> ? AND a.status = ? AND r.status = ? AND r.permissions = ?", targetID, models.AdminStatusNormal, 1, `["*"]`).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("不能删除、停用或降权最后一个启用的超级管理员")
	}
	return nil
}

func (s *AdminService) isAdminSuper(ctx context.Context, admin models.Admin) bool {
	return s.isRoleSuper(ctx, admin.RoleID)
}

func (s *AdminService) isRoleSuper(ctx context.Context, roleID int64) bool {
	if roleID == 0 {
		return false
	}
	var role models.AdminRole
	if s.DB.WithContext(ctx).First(&role, roleID).Error != nil {
		return false
	}
	permissions, err := normalizeAdminPermissions(role.Permissions)
	return err == nil && len(permissions) == 1 && permissions[0] == "*"
}

func (s *AdminService) ensureRoleChangeKeepsSuperAdmin(ctx context.Context, roleID int64, permissions []string, status int8) error {
	var role models.AdminRole
	if err := s.DB.WithContext(ctx).First(&role, roleID).Error; err != nil {
		return err
	}
	oldPermissions, err := normalizeAdminPermissions(role.Permissions)
	if err != nil || len(oldPermissions) != 1 || oldPermissions[0] != "*" {
		return nil
	}
	if status == 1 && len(permissions) == 1 && permissions[0] == "*" {
		return nil
	}
	var count int64
	if err := s.DB.WithContext(ctx).Table("admin a").Joins("JOIN admin_roles r ON r.id = a.role_id").Where("a.role_id <> ? AND a.status = ? AND r.status = ? AND r.permissions = ?", roleID, models.AdminStatusNormal, 1, `["*"]`).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("不能删除、停用或降权最后一个启用的超级管理员角色")
	}
	return nil
}

func (s *AdminService) GetPointsRule(ctx context.Context) (*types.PointsRule, error) {
	return loadPointsRule(ctx, s.DB)
}

func (s *AdminService) UpdatePointsRule(ctx context.Context, req types.UpdatePointsRuleRequest) error {
	if req.DiscountCentsPerPoint <= 0 || req.RewardCentsPerPoint <= 0 {
		return errors.New("积分规则必须大于 0")
	}
	settings := []models.PlatformSetting{
		{Key: pointsDiscountSettingKey, Value: fmt.Sprintf("%d", req.DiscountCentsPerPoint), Remark: "每积分可抵扣金额，单位分"},
		{Key: pointsRewardSettingKey, Value: fmt.Sprintf("%d", req.RewardCentsPerPoint), Remark: "消费多少分奖励1积分"},
	}
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, setting := range settings {
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "setting_key"}},
				DoUpdates: clause.Assignments(map[string]any{
					"setting_value": setting.Value,
					"remark":        setting.Remark,
					"updated_at":    time.Now(),
				}),
			}).Create(&setting).Error; err != nil {
				return err
			}
		}
		return nil
	})
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
	if req.Type == "distance" {
		value := strings.TrimSpace(req.Value)
		if value == "" {
			return 0, errors.New("距离筛选项必须填写公里值")
		}
		distance, err := strconv.ParseFloat(value, 64)
		if err != nil || distance <= 0 {
			return 0, errors.New("距离筛选项只能填写大于0的数字公里值")
		}
		req.Value = value
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
		return adminMapPage(s.DB.WithContext(ctx).Table("ticket_orders o").
			Select(`o.id, o.order_no, o.status, o.quantity, o.actual_price, o.created_at,
				a.id AS activity_id, a.name AS activity_name, a.poster_list, ts.name AS ticket_spec_name`).
			Joins("LEFT JOIN activities a ON a.id = o.activity_id").
			Joins("LEFT JOIN ticket_specs ts ON ts.id = o.ticket_spec_id").
			Where("o.user_id = ? AND o.status IN ?", userID, []int8{models.TicketOrderStatusUsable, models.TicketOrderStatusUsed, models.TicketOrderStatusRefunding, models.TicketOrderStatusRefundSuccess, models.TicketOrderStatusRefundReject}).
			Order("o.id desc"), page, pageSize)
	case "subscribes":
		return adminMapPage(s.DB.WithContext(ctx).Table("activity_subscriptions sub").
			Select(`sub.id, sub.activity_id, sub.created_at, a.name AS activity_name, a.poster_list, a.status AS activity_status`).
			Joins("LEFT JOIN activities a ON a.id = sub.activity_id").
			Where("sub.user_id = ?", userID).
			Order("sub.id desc"), page, pageSize)
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

func (s *AdminService) UpdateVerifierStatus(ctx context.Context, id int64, status int8) error {
	if status != models.VerifierStatusInactive && status != models.VerifierStatusActive {
		return errors.New("核销员状态无效")
	}
	result := s.DB.WithContext(ctx).Model(&models.Verifier{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
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
	query := s.DB.WithContext(ctx).Table("notes n").Select("n.*, u.nickname, u.avatar, ns.like_count, ns.coll_count, ns.coll_count AS collect_count, ns.comment_count, ns.share_count").Joins("LEFT JOIN users u ON u.id = n.user_id").Joins("LEFT JOIN note_stats ns ON ns.note_id = n.id")
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
		return adminMapPage(s.DB.WithContext(ctx).Table("note_shares ns").
			Select("ns.id, ns.note_id, ns.user_id, u.nickname AS user_name, CASE WHEN u.mobile = '' THEN '' ELSE CONCAT(LEFT(u.mobile, 3), '****', RIGHT(u.mobile, 4)) END AS user_mobile, ns.channel, ns.created_at").
			Joins("LEFT JOIN users u ON u.id = ns.user_id").
			Where("ns.note_id = ?", noteID).
			Order("ns.id desc"), page, pageSize)
	default:
		return nil, errors.New("记录类型无效")
	}
}

func (s *AdminService) UpdateCommentStatus(ctx context.Context, noteID, commentID int64, status int8) error {
	if status != -1 && status != 0 && status != 1 {
		return errors.New("评论状态仅支持 -1删除、0隐藏、1公开")
	}
	result := s.DB.WithContext(ctx).Model(&models.Comment{}).
		Where("id = ? AND note_id = ?", commentID, noteID).
		Updates(map[string]any{"status": status, "updated_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *AdminService) ListPointLogs(ctx context.Context, page, pageSize int, userID int64) (*types.AdminPageResponse[map[string]any], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	query := s.DB.WithContext(ctx).Table("point_logs pl").Select("pl.*, u.nickname, u.mobile").Joins("LEFT JOIN users u ON u.id = pl.user_id")
	if userID > 0 {
		query = query.Where("pl.user_id = ?", userID)
	}
	return adminMapPage(query.Order("pl.id desc"), page, pageSize)
}

func (s *AdminService) ListPlatformFinanceFlows(ctx context.Context, page, pageSize int, filter types.AdminPlatformFlowFilter) (*types.AdminPlatformFlowListResponse, error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	query := s.DB.WithContext(ctx).Model(&models.PlatformFinanceFlow{})
	if flowType := strings.TrimSpace(filter.Type); flowType != "" {
		query = query.Where("type = ?", flowType)
	}
	if filter.OrganizerID > 0 {
		query = query.Where("organizer_id = ?", filter.OrganizerID)
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("flow_no LIKE ? OR order_no LIKE ? OR refund_no LIKE ? OR organizer_name LIKE ?", like, like, like, like)
	}
	if filter.StartDate != nil {
		query = query.Where("occurred_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("occurred_at < ?", filter.EndDate.AddDate(0, 0, 1))
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []types.AdminPlatformFlowItem
	if err := query.Select("id, flow_no, type, direction, amount, order_no, refund_no, withdraw_id, organizer_id, organizer_name, pay_method, remark, occurred_at").
		Order("occurred_at desc, id desc").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&list).Error; err != nil {
		return nil, err
	}
	return &types.AdminPlatformFlowListResponse{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *AdminService) AdjustPoints(ctx context.Context, adminID int64, req types.AdminPointsAdjustRequest) (*types.AdminPointsAdjustResponse, error) {
	req.Reason = strings.TrimSpace(req.Reason)
	req.RequestNo = strings.TrimSpace(req.RequestNo)
	if req.UserID <= 0 || req.Points == 0 || req.Reason == "" || req.RequestNo == "" {
		return nil, errors.New("用户、积分变动、调整原因和请求编号不能为空")
	}
	resp := &types.AdminPointsAdjustResponse{UserID: req.UserID, ChangePoints: req.Points, RequestNo: req.RequestNo}
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user models.Users
		if err := tx.First(&user, req.UserID).Error; err != nil {
			return errors.New("用户不存在")
		}
		var existing models.PointsLog
		if err := tx.Where("user_id = ? AND source_id = ? AND change_type = ?", req.UserID, req.RequestNo, models.TypeSystemCompensate).First(&existing).Error; err == nil {
			resp.ChangePoints = existing.Amount
			resp.Balance = existing.Balance
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		var account models.UserPoint
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", req.UserID).First(&account).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			account = models.UserPoint{UserID: uint64(req.UserID)}
			if err := tx.Create(&account).Error; err != nil {
				return err
			}
		}
		// A concurrent request may have passed the first idempotency check before
		// this account lock was acquired, so check again while holding the lock.
		if err := tx.Where("user_id = ? AND source_id = ? AND change_type = ?", req.UserID, req.RequestNo, models.TypeSystemCompensate).First(&existing).Error; err == nil {
			resp.ChangePoints = existing.Amount
			resp.Balance = existing.Balance
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		newBalance := account.Balance + req.Points
		if newBalance < 0 {
			return errors.New("扣减后积分余额不能小于0")
		}
		updates := map[string]any{"balance": newBalance}
		if req.Points > 0 {
			updates["total_earned"] = gorm.Expr("total_earned + ?", req.Points)
		} else {
			updates["total_used"] = gorm.Expr("total_used + ?", -req.Points)
		}
		if err := tx.Model(&models.UserPoint{}).Where("user_id = ?", req.UserID).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.PointsLog{UserID: uint64(req.UserID), Amount: req.Points, Balance: newBalance, ChangeType: models.TypeSystemCompensate, SourceID: req.RequestNo, Remark: req.Reason, Status: 1}).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.AdminOperationLog{AdminID: adminID, Action: "points_adjust", Resource: "points", Method: "POST", Path: "/api/v1/admin/points/adjust", Remark: fmt.Sprintf("user_id=%d,points=%d,request_no=%s,reason=%s", req.UserID, req.Points, req.RequestNo, req.Reason)}).Error; err != nil {
			return err
		}
		resp.Balance = newBalance
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
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
	if req.Status != 1 && req.Status != 2 {
		return errors.New("提现审核状态仅支持 1通过 或 2拒绝")
	}
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var withdraw models.OrganizerWithdraw
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&withdraw).Error; err != nil {
			return err
		}
		if withdraw.Status != 0 {
			return errors.New("该提现申请已审核")
		}
		if err := tx.Model(&withdraw).Updates(map[string]any{"status": req.Status, "remark": req.Remark, "updated_at": time.Now()}).Error; err != nil {
			return err
		}
		if req.Status == 1 {
			return recordPlatformWithdraw(tx, withdraw)
		}
		return nil
	})
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

func (s *AdminService) ListOrganizerLevelRules(ctx context.Context) ([]models.OrganizerLevelRule, error) {
	if err := s.ensureDefaultOrganizerLevelRules(ctx); err != nil {
		return nil, err
	}
	var rules []models.OrganizerLevelRule
	err := s.DB.WithContext(ctx).Order("level ASC").Find(&rules).Error
	return rules, err
}

func (s *AdminService) SaveOrganizerLevelRule(ctx context.Context, id int64, req types.OrganizerLevelRuleRequest) (int64, error) {
	if req.Level <= 0 {
		return 0, errors.New("等级必须大于0")
	}
	if req.Name == "" {
		req.Name = fmt.Sprintf("LV%d", req.Level)
	}
	rule := models.OrganizerLevelRule{
		Level:                 req.Level,
		Name:                  req.Name,
		FeeRate:               req.FeeRate,
		RequiredActivityCount: req.RequiredActivityCount,
		Description:           req.Description,
		Benefits:              req.Benefits,
		Status:                req.Status,
	}
	if id > 0 {
		result := s.DB.WithContext(ctx).Model(&models.OrganizerLevelRule{}).Where("id = ?", id).Updates(map[string]any{
			"level": req.Level, "name": req.Name, "fee_rate": req.FeeRate,
			"required_activity_count": req.RequiredActivityCount, "description": req.Description,
			"benefits": req.Benefits, "status": req.Status,
		})
		if result.Error != nil {
			return 0, result.Error
		}
		if result.RowsAffected == 0 {
			return 0, gorm.ErrRecordNotFound
		}
		return id, nil
	}
	if err := s.DB.WithContext(ctx).Create(&rule).Error; err != nil {
		return 0, err
	}
	return rule.ID, nil
}

func (s *AdminService) DeleteOrganizerLevelRule(ctx context.Context, id int64) error {
	result := s.DB.WithContext(ctx).Delete(&models.OrganizerLevelRule{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *AdminService) ensureDefaultOrganizerLevelRules(ctx context.Context) error {
	for _, rule := range defaultOrganizerLevelRules() {
		if err := s.DB.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "level"}}, DoNothing: true}).Create(&rule).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *AdminService) ListMessages(ctx context.Context, page, pageSize int, target, messageType string) (*types.AdminPageResponse[models.PlatformMessage], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	var total int64
	query := s.DB.WithContext(ctx).Model(&models.PlatformMessage{})
	if target = strings.TrimSpace(target); target != "" {
		query = query.Where("target = ?", target)
	}
	if messageType = strings.TrimSpace(messageType); messageType != "" {
		query = query.Where("type = ?", messageType)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []models.PlatformMessage
	err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return &types.AdminPageResponse[models.PlatformMessage]{List: list, Total: total, Page: page, PageSize: pageSize}, err
}

func (s *AdminService) ListMessageDeliveries(ctx context.Context, messageID int64, page, pageSize int, status *int8) (*types.AdminPageResponse[map[string]any], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	query := s.DB.WithContext(ctx).Table("platform_message_deliveries d").
		Select("d.*, u.nickname, u.avatar, u.mobile").
		Joins("LEFT JOIN users u ON u.id = d.user_id").
		Where("d.message_id = ?", messageID)
	if status != nil {
		query = query.Where("d.status = ?", *status)
	}
	return adminMapPage(query.Order("d.id desc"), page, pageSize)
}

func (s *AdminService) CreateMessage(ctx context.Context, req types.PlatformMessageRequest) (int64, error) {
	req.Title = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)
	req.Type = strings.TrimSpace(req.Type)
	req.Target = strings.TrimSpace(req.Target)
	if req.Title == "" {
		return 0, errors.New("消息标题不能为空")
	}
	if req.Content == "" {
		return 0, errors.New("消息内容不能为空")
	}
	if req.Type == "" {
		req.Type = "system"
	}
	if req.Target == "" {
		req.Target = "all"
	}
	msg := models.PlatformMessage{Title: req.Title, Content: req.Content, Type: req.Type, Target: req.Target, Status: req.Status}
	if err := s.DB.WithContext(ctx).Create(&msg).Error; err != nil {
		return 0, err
	}
	if req.Status == 1 {
		targetIDs, err := s.resolvePlatformMessageTargets(ctx, req)
		if err != nil {
			return msg.ID, err
		}
		deliveries := make([]models.PlatformMessageDelivery, 0, len(targetIDs))
		for _, targetID := range targetIDs {
			deliveries = append(deliveries, models.PlatformMessageDelivery{MessageID: msg.ID, UserID: targetID})
		}
		if err := s.DB.WithContext(ctx).Create(&deliveries).Error; err != nil {
			return msg.ID, err
		}
		if err := s.publishPlatformMessage(ctx, msg, targetIDs); err != nil {
			return msg.ID, err
		}
	}
	return msg.ID, nil
}

func (s *AdminService) publishPlatformMessage(ctx context.Context, msg models.PlatformMessage, targetIDs []int64) error {
	if s.MqProducer == nil {
		return errors.New("消息队列未初始化")
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
			_ = s.DB.WithContext(ctx).Model(&models.PlatformMessageDelivery{}).
				Where("message_id = ? AND user_id = ?", msg.ID, targetID).
				Updates(map[string]any{"status": 2, "error": err.Error()}).Error
			return err
		}
		now := time.Now()
		_ = s.DB.WithContext(ctx).Model(&models.PlatformMessageDelivery{}).
			Where("message_id = ? AND user_id = ?", msg.ID, targetID).
			Updates(map[string]any{"status": 1, "sent_at": now, "error": ""}).Error
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
