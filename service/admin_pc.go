package service

import (
	"Hyper/models"
	"Hyper/pkg/encrypt"
	"Hyper/types"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
	resp := &types.AdminProfileResponse{ID: admin.Id, Username: admin.Username, Nickname: admin.Nickname, Avatar: admin.Avatar, Mobile: admin.Mobile, Email: admin.Email, Motto: admin.Motto, RoleID: admin.RoleID, Status: admin.Status, Permissions: []string{}, CreatedAt: admin.CreatedAt, UpdatedAt: admin.UpdatedAt}
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
		"avatar":   req.Avatar,
		"nickname": req.Nickname,
		"mobile":   req.Mobile,
		"email":    req.Email,
		"motto":    req.Motto,
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
		ID                                        int
		Username, Nickname, Avatar, Mobile, Email string
		RoleID                                    int64
		Status                                    int8
		CreatedAt, UpdatedAt                      time.Time
		RoleName                                  string
		Permissions                               string
	}
	if err := query.Select("a.id, a.username, a.nickname, a.avatar, a.mobile, a.email, a.role_id, a.status, a.created_at, a.updated_at, r.name AS role_name, r.permissions").Order("a.id desc").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]types.AdminAccountItem, 0, len(rows))
	for _, row := range rows {
		permissions, _ := normalizeAdminPermissions(row.Permissions)
		list = append(list, types.AdminAccountItem{ID: row.ID, Username: row.Username, Nickname: row.Nickname, Avatar: row.Avatar, Mobile: row.Mobile, Email: row.Email, RoleID: row.RoleID, Role: types.AdminRoleSummary{ID: row.RoleID, Name: row.RoleName, Permissions: permissions}, Status: row.Status, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt})
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
	admin := models.Admin{Username: req.Username, Nickname: req.Nickname, Password: encrypt.HashPassword(req.Password), Avatar: req.Avatar, Mobile: req.Mobile, Email: req.Email, RoleID: req.RoleID, Status: req.Status}
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
	updates := map[string]any{"username": req.Username, "nickname": req.Nickname, "avatar": req.Avatar, "mobile": req.Mobile, "email": req.Email, "role_id": req.RoleID}
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

func (s *AdminService) ListOperationLogs(ctx context.Context, page, pageSize int, filter types.AdminOperationLogFilter) (*types.AdminPageResponse[types.AdminOperationLogItem], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	query := s.DB.WithContext(ctx).Table("admin_operation_logs l").Joins("LEFT JOIN admin a ON a.id = l.admin_id")
	if filter.AdminID > 0 {
		query = query.Where("l.admin_id = ?", filter.AdminID)
	}
	if filter.Action = strings.TrimSpace(filter.Action); filter.Action != "" {
		query = query.Where("l.action = ?", filter.Action)
	}
	if filter.ResourceType = strings.TrimSpace(filter.ResourceType); filter.ResourceType != "" {
		query = query.Where("COALESCE(NULLIF(l.resource_type, ''), l.resource) = ?", filter.ResourceType)
	}
	if filter.Result = strings.TrimSpace(filter.Result); filter.Result != "" {
		query = query.Where("l.result = ?", filter.Result)
	}
	if filter.StartDate != nil {
		query = query.Where("l.created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("l.created_at < ?", filter.EndDate.AddDate(0, 0, 1))
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("l.action LIKE ? OR l.resource_type LIKE ? OR l.resource_id LIKE ? OR l.resource_name LIKE ? OR l.remark LIKE ? OR l.error_message LIKE ? OR a.username LIKE ?", like, like, like, like, like, like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []types.AdminOperationLogItem
	if err := query.Select("l.id, l.admin_id, COALESCE(a.username, '') AS admin_name, COALESCE(a.username, '') AS admin_username, l.action, COALESCE(NULLIF(l.resource_type, ''), l.resource) AS resource_type, l.resource_id, l.resource_name, l.method, l.path, l.remark, COALESCE(NULLIF(l.result, ''), 'success') AS result, l.error_code, l.error_message, l.ip, l.created_at").Order("l.id desc").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&list).Error; err != nil {
		return nil, err
	}
	for i := range list {
		list[i].ActionName = adminAuditActionName(list[i].Action)
		if list[i].ResourceName == "" {
			list[i].ResourceName = adminAuditResourceName(list[i].ResourceType, list[i].ResourceID)
		}
	}
	return &types.AdminPageResponse[types.AdminOperationLogItem]{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *AdminService) RecordOperationLog(ctx context.Context, item models.AdminOperationLog) error {
	if item.Result == "" {
		item.Result = "success"
	}
	if item.ResourceType == "" {
		item.ResourceType = item.Resource
	}
	if item.Resource == "" {
		item.Resource = item.ResourceType
	}
	if item.ResourceName == "" {
		item.ResourceName = adminAuditResourceName(item.ResourceType, item.ResourceID)
	}
	return s.DB.WithContext(ctx).Create(&item).Error
}

func adminAuditActionName(action string) string {
	names := map[string]string{
		"admin.settings.update": "更新系统配置", "admin.category.create": "新增分类", "admin.category.update": "更新分类", "admin.category.delete": "删除分类",
		"admin.role.create": "新增角色", "admin.role.update": "更新角色", "admin.role.delete": "删除角色",
		"admin.account.create": "新增管理员", "admin.account.update": "更新管理员", "admin.account.delete": "删除管理员",
		"admin.account.disable": "停用管理员", "admin.withdraw.approve": "审核通过提现申请", "admin.withdraw.reject": "驳回提现申请",
		"admin.refund.approve": "审核通过退款申请", "admin.refund.reject": "驳回退款申请", "admin.organizer.approve": "审核通过入驻申请",
		"admin.organizer.reject": "驳回入驻申请", "admin.activity.approve": "审核通过活动", "admin.activity.reject": "驳回活动", "admin.activity.hide": "下架隐藏活动", "admin.activity.restore": "恢复活动展示",
		"admin.bank_account.approve": "审核通过收款账户", "admin.bank_account.reject": "驳回收款账户", "admin.points.adjust": "人工调整积分", "admin.permission.denied": "权限拒绝",
		"admin.comment.moderate": "审核动态评论",
		"admin.profile.update":   "更新个人资料", "admin.activity_collection.create": "新增活动合集", "admin.activity_collection.update": "更新活动合集", "admin.activity_collection.delete": "删除活动合集",
		"admin.party.update": "更新派对或场地状态", "admin.party.hide": "下架隐藏派对或场地", "admin.party.tags.update": "更新派对标签", "admin.venue.tags.update": "更新场地标签", "admin.activity.tags.update": "更新活动标签", "admin.user.update": "更新用户状态", "admin.verifier.update": "更新核销员状态",
		"admin.customer_service.reply": "回复客服会话", "admin.customer_service.read": "标记客服会话已读",
		"admin.message.create": "发布平台消息", "admin.banner.create": "新增轮播图", "admin.banner.update": "更新轮播图", "admin.banner.delete": "删除轮播图",
	}
	if name, ok := names[action]; ok {
		return name
	}
	return action
}

func adminAuditResourceName(resourceType, resourceID string) string {
	base := map[string]string{"settings": "系统配置", "category": "分类", "role": "角色", "admin": "管理员", "withdraw": "提现申请", "refund": "退款申请", "organizer": "入驻申请", "venue": "场地", "party": "派对", "activity": "活动", "points": "积分账户", "permission": "权限校验", "customer_service_session": "客服会话"}[resourceType]
	if base == "" {
		base = resourceType
	}
	if resourceID == "" {
		return base
	}
	return fmt.Sprintf("%s %s", base, resourceID)
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
	"admin.verifications": {}, "admin.content": {}, "admin.finance": {}, "admin.customer_service": {},
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
	case strings.HasPrefix(path, "/v1/admin/settings"), strings.HasPrefix(path, "/v1/admin/system-config"), strings.HasPrefix(path, "/v1/admin/categories"), strings.HasPrefix(path, "/v1/admin/logs"), strings.HasPrefix(path, "/v1/admin/admins"), strings.HasPrefix(path, "/v1/admin/roles"), strings.HasPrefix(path, "/v1/admin/wechat-subscribe"):
		return "admin.system"
	case strings.HasPrefix(path, "/v1/admin/users"), strings.HasPrefix(path, "/v1/admin/viewers"):
		return "admin.users"
	case strings.HasPrefix(path, "/v1/admin/organizers"), strings.HasPrefix(path, "/v1/admin/venues"):
		return "admin.organizers"
	case strings.HasPrefix(path, "/v1/admin/activities"), strings.HasPrefix(path, "/v1/admin/parties"), strings.HasPrefix(path, "/v1/admin/activity-collections"):
		return "admin.activities"
	case strings.HasPrefix(path, "/v1/admin/tickets") || strings.HasPrefix(path, "/v1/admin/events/"):
		return "admin.tickets"
	case strings.HasPrefix(path, "/v1/admin/orders"), strings.HasPrefix(path, "/v1/admin/refunds"):
		return "admin.orders"
	case strings.HasPrefix(path, "/v1/admin/verifiers"), strings.HasPrefix(path, "/v1/admin/verification-records"):
		return "admin.verifications"
	case strings.HasPrefix(path, "/v1/admin/notes"), strings.HasPrefix(path, "/v1/admin/note-interactions"), strings.HasPrefix(path, "/v1/admin/note-comments"), strings.HasPrefix(path, "/v1/admin/messages"), strings.HasPrefix(path, "/v1/admin/banners"):
		return "admin.content"
	case strings.HasPrefix(path, "/v1/admin/finance"), strings.HasPrefix(path, "/v1/admin/withdraws"), strings.HasPrefix(path, "/v1/admin/bank-account-audits"), strings.HasPrefix(path, "/v1/admin/points"):
		return "admin.finance"
	case strings.HasPrefix(path, "/v1/admin/organizer-level-rules"):
		return "admin.merchants"
	case strings.HasPrefix(path, "/v1/admin/customer-service"):
		return "admin.customer_service"
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
	status := int8(1)
	if req.Status != nil {
		status = *req.Status
	}
	if status != 0 && status != 1 {
		return 0, errors.New("分类状态仅支持 0停用 或 1启用")
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
	if req.Type == models.ContentTagTypeCoupon {
		var count int64
		query := s.DB.WithContext(ctx).Model(&models.AdminCategory{}).Where("type = ? AND name = ?", req.Type, strings.TrimSpace(req.Name))
		if id > 0 {
			query = query.Where("id <> ?", id)
		}
		if err := query.Count(&count).Error; err != nil {
			return 0, err
		}
		if count > 0 {
			return 0, errors.New("优惠标签名称已存在")
		}
	}
	row := models.AdminCategory{Type: req.Type, Name: req.Name, Image: req.Image, Value: req.Value, Sort: req.Sort, Status: status}
	if id > 0 {
		return id, s.DB.WithContext(ctx).Model(&models.AdminCategory{}).Where("id = ?", id).Updates(row).Error
	}
	if err := s.DB.WithContext(ctx).Create(&row).Error; err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (s *AdminService) DeleteCategory(ctx context.Context, id int64) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var category models.AdminCategory
		if err := tx.Where("id = ?", id).First(&category).Error; err != nil {
			return err
		}
		if category.Type == models.ContentTagTypeCoupon {
			if err := tx.Where("tag_id = ?", id).Delete(&models.ContentTagRelation{}).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&category).Error
	})
}

func (s *AdminService) ListUserRecords(ctx context.Context, userID int64, recordType string, page, pageSize int) (*types.AdminPageResponse[map[string]any], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	switch recordType {
	case "likes":
		// "获赞记录" is the likes received by this user's published notes, not the likes they gave.
		return adminMapPage(s.DB.WithContext(ctx).Table("note_likes nl").
			Select("nl.id, nl.note_id, nl.user_id AS liker_user_id, nl.created_at, nl.updated_at, n.title AS note_title, n.content AS note_content, u.nickname AS liker_name, u.avatar AS liker_avatar, u.mobile AS liker_mobile").
			Joins("JOIN notes n ON n.id = nl.note_id").
			Joins("LEFT JOIN users u ON u.id = nl.user_id").
			Where("n.user_id = ? AND n.status <> ? AND nl.status = 1", userID, -1).
			Order("nl.id desc"), page, pageSize)
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

func (s *AdminService) SaveVerifier(ctx context.Context, id int64, req types.AdminVerifierRequest) (int64, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Phone = strings.TrimSpace(req.Phone)
	req.PermissionScope = strings.TrimSpace(req.PermissionScope)
	req.Channel = strings.TrimSpace(req.Channel)
	if req.Name == "" || req.Phone == "" {
		return 0, errors.New("核销员姓名和手机号不能为空")
	}
	if req.Status != models.VerifierStatusInactive && req.Status != models.VerifierStatusActive {
		return 0, errors.New("核销员状态无效")
	}
	if req.PermissionScope == "" {
		req.PermissionScope = "活动"
	}
	var organizer models.Organizer
	if err := s.DB.WithContext(ctx).First(&organizer, req.OrganizerID).Error; err != nil {
		return 0, err
	}
	updates := map[string]any{
		"organizer_id":     req.OrganizerID,
		"user_id":          req.UserID,
		"name":             req.Name,
		"phone":            req.Phone,
		"status":           req.Status,
		"permission_scope": req.PermissionScope,
		"channel":          req.Channel,
	}
	if id == 0 {
		verifier := models.Verifier{OrganizerID: req.OrganizerID, UserID: req.UserID, Name: req.Name, Phone: req.Phone, Status: req.Status, PermissionScope: req.PermissionScope, Channel: req.Channel}
		if req.UserID > 0 && req.Status == models.VerifierStatusActive {
			now := time.Now()
			verifier.BoundAt = &now
		}
		if err := s.DB.WithContext(ctx).Create(&verifier).Error; err != nil {
			return 0, err
		}
		return verifier.ID, nil
	}
	result := s.DB.WithContext(ctx).Model(&models.Verifier{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, gorm.ErrRecordNotFound
	}
	return id, nil
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

func (s *AdminService) DeleteVerifier(ctx context.Context, id int64) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.VerificationRecord{}).Where("verifier_id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("该核销员已有核销记录，不能删除")
		}
		result := tx.Delete(&models.Verifier{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (s *AdminService) ListVerificationRecords(ctx context.Context, page, pageSize int, keyword string, organizerID int64) (*types.AdminPageResponse[map[string]any], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	query := s.DB.WithContext(ctx).Table("verification_records vr").
		Select(`vr.*, v.name AS verifier_name, v.phone AS verifier_phone, v.user_id AS verifier_user_id,
			o.name AS organizer_name, tor.order_no, tor.quantity, tor.buyer_name, tor.buyer_id_card,
			a.name AS activity_name, COALESCE(NULLIF(a.poster_list, ''), NULLIF(a.poster_detail, ''), NULLIF(a.poster_wechat, ''), NULLIF(a.poster_long, '')) AS poster_list,
			ts.name AS ticket_spec_name, u.mobile AS buyer_phone`).
		Joins("LEFT JOIN verifiers v ON v.id = vr.verifier_id").
		Joins("LEFT JOIN organizers o ON o.id = v.organizer_id").
		Joins("LEFT JOIN ticket_orders tor ON tor.id = vr.order_id").
		Joins("LEFT JOIN activities a ON a.id = vr.activity_id").
		Joins("LEFT JOIN ticket_specs ts ON ts.id = tor.ticket_spec_id").
		Joins("LEFT JOIN users u ON u.id = tor.user_id")
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
	if status != -1 && status != 0 && status != 1 {
		return errors.New("动态状态仅支持 -1删除、0隐藏、1公开")
	}
	updates := map[string]any{"status": status, "updated_at": time.Now()}
	// status=0 is the management-side hide command. Existing historical posts may
	// also have status=0, so visibility must be changed explicitly to hide them.
	if status == 0 {
		updates["visible_conf"] = types.VisibleConfPrivate
	} else if status == 1 {
		updates["visible_conf"] = types.VisibleConfPublic
	}
	result := s.DB.WithContext(ctx).Model(&models.Note{}).Where("id = ?", noteID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
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

func (s *AdminService) UpdateCommentStatus(ctx context.Context, noteID, commentID, adminID int64, status int8, reason string) error {
	if status != -1 && status != 0 && status != 1 {
		return errors.New("评论状态仅支持 -1删除、0隐藏、1公开")
	}
	query := s.DB.WithContext(ctx).Model(&models.Comment{}).Where("id = ?", commentID)
	if noteID > 0 {
		query = query.Where("note_id = ?", noteID)
	}
	now := time.Now()
	result := query.Updates(map[string]any{"status": status, "moderated_by": adminID, "moderated_at": now, "moderate_reason": strings.TrimSpace(reason), "updated_at": now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *AdminService) ListNoteInteractions(ctx context.Context, page, pageSize int, filter types.AdminNoteInteractionFilter) (*types.AdminPageResponse[map[string]any], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	parts := make([]string, 0, 3)
	args := make([]any, 0)
	appendPart := func(kind, table, channelExpr string) {
		where := " WHERE 1=1"
		localArgs := make([]any, 0)
		if filter.NoteID > 0 {
			where += " AND x.note_id = ?"
			localArgs = append(localArgs, filter.NoteID)
		}
		if filter.UserID > 0 {
			where += " AND x.user_id = ?"
			localArgs = append(localArgs, filter.UserID)
		}
		if filter.StartDate != nil {
			where += " AND x.created_at >= ?"
			localArgs = append(localArgs, *filter.StartDate)
		}
		if filter.EndDate != nil {
			where += " AND x.created_at < ?"
			localArgs = append(localArgs, filter.EndDate.AddDate(0, 0, 1))
		}
		if kind == "share" && strings.TrimSpace(filter.Channel) != "" {
			where += " AND x.channel = ?"
			localArgs = append(localArgs, strings.TrimSpace(filter.Channel))
		}
		if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
			like := "%" + keyword + "%"
			where += " AND (n.title LIKE ? OR n.content LIKE ? OR u.nickname LIKE ? OR u.mobile LIKE ?)"
			localArgs = append(localArgs, like, like, like, like)
		}
		parts = append(parts, "SELECT x.id, '"+kind+"' AS type, x.note_id, CONCAT(LEFT(COALESCE(n.content, ''), 120)) AS note_content, n.status AS note_status, x.user_id, u.nickname AS user_name, CASE WHEN u.mobile = '' THEN '' ELSE CONCAT(LEFT(u.mobile, 3), '****', RIGHT(u.mobile, 4)) END AS user_mobile, "+channelExpr+" AS channel, x.created_at FROM "+table+" x LEFT JOIN notes n ON n.id = x.note_id LEFT JOIN users u ON u.id = x.user_id"+where)
		args = append(args, localArgs...)
	}
	if filter.Type == "" || filter.Type == "like" {
		appendPart("like", "note_likes", "''")
	}
	if filter.Type == "" || filter.Type == "collection" {
		appendPart("collection", "note_collections", "''")
	}
	if filter.Type == "" || filter.Type == "share" {
		appendPart("share", "note_shares", "x.channel")
	}
	if len(parts) == 0 {
		return nil, errors.New("互动类型仅支持 like、collection、share")
	}
	// Likes and collections retain cancelled rows, so only active interactions are included.
	for i := range parts {
		if strings.Contains(parts[i], "FROM note_likes") || strings.Contains(parts[i], "FROM note_collections") {
			parts[i] += " AND x.status = 1"
		}
	}
	union := strings.Join(parts, " UNION ALL ")
	var total int64
	if err := s.DB.WithContext(ctx).Raw("SELECT COUNT(*) FROM ("+union+") interactions", args...).Scan(&total).Error; err != nil {
		return nil, err
	}
	queryArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)
	var list []map[string]any
	if err := s.DB.WithContext(ctx).Raw("SELECT * FROM ("+union+") interactions ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?", queryArgs...).Scan(&list).Error; err != nil {
		return nil, err
	}
	return &types.AdminPageResponse[map[string]any]{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *AdminService) ListNoteComments(ctx context.Context, page, pageSize int, filter types.AdminNoteCommentFilter) (*types.AdminPageResponse[map[string]any], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	query := s.DB.WithContext(ctx).Table("comments c").Select("c.id, c.note_id, LEFT(COALESCE(n.content, ''), 120) AS note_content, c.user_id, u.nickname AS user_name, CASE WHEN u.mobile = '' THEN '' ELSE CONCAT(LEFT(u.mobile, 3), '****', RIGHT(u.mobile, 4)) END AS user_mobile, c.content, c.status, c.created_at, c.updated_at, c.moderated_by, c.moderated_at, c.moderate_reason, ma.username AS moderated_by_name").Joins("LEFT JOIN notes n ON n.id = c.note_id").Joins("LEFT JOIN users u ON u.id = c.user_id").Joins("LEFT JOIN admin ma ON ma.id = c.moderated_by")
	if filter.Status != nil {
		query = query.Where("c.status = ?", *filter.Status)
	}
	if filter.NoteID > 0 {
		query = query.Where("c.note_id = ?", filter.NoteID)
	}
	if filter.UserID > 0 {
		query = query.Where("c.user_id = ?", filter.UserID)
	}
	if filter.StartDate != nil {
		query = query.Where("c.created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("c.created_at < ?", filter.EndDate.AddDate(0, 0, 1))
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("c.content LIKE ? OR n.title LIKE ? OR n.content LIKE ? OR u.nickname LIKE ? OR u.mobile LIKE ?", like, like, like, like, like)
	}
	return adminMapPage(query.Order("c.id desc"), page, pageSize)
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
		allocationStatus := models.OrganizerWithdrawAllocationStatusReleased
		if req.Status == 1 {
			allocationStatus = models.OrganizerWithdrawAllocationStatusSettled
		}
		if err := tx.Model(&models.OrganizerWithdrawAllocation{}).
			Where("withdraw_id = ? AND status = ?", withdraw.ID, models.OrganizerWithdrawAllocationStatusPending).
			Updates(map[string]any{"status": allocationStatus, "updated_at": time.Now()}).Error; err != nil {
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
				"bank_account_name":  audit.BankAccountName,
				"bank_account_no":    audit.BankAccountNo,
				"bank_name":          audit.BankName,
				"bank_contact_name":  audit.BankContactName,
				"bank_contact_phone": audit.BankContactPhone,
				"updated_at":         now,
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
	return ensureDefaultOrganizerLevelRules(s.DB.WithContext(ctx))
}

func (s *AdminService) ListMessages(ctx context.Context, page, pageSize int, target, messageType string) (*types.AdminPageResponse[map[string]any], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	var total int64
	query := s.DB.WithContext(ctx).Table("platform_messages pm").Joins("LEFT JOIN admin a ON a.id = pm.creator_id")
	if target = strings.TrimSpace(target); target != "" {
		query = query.Where("pm.target = ?", target)
	}
	if messageType = strings.TrimSpace(messageType); messageType != "" {
		query = query.Where("pm.type = ?", messageType)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []map[string]any
	err := query.Select(`pm.id, pm.title, pm.content, pm.target, pm.type, pm.status, pm.channel, pm.creator_id, COALESCE(a.username, '') AS creator_name, pm.created_at, pm.updated_at,
		(SELECT COUNT(*) FROM platform_message_deliveries d WHERE d.message_id = pm.id) AS target_count,
		(SELECT COUNT(*) FROM platform_message_deliveries d WHERE d.message_id = pm.id AND d.status IN (1,3)) AS sent_count,
		0 AS delivered_count,
		(SELECT COUNT(*) FROM platform_message_deliveries d WHERE d.message_id = pm.id AND d.status = 2) AS failed_count,
		(SELECT COUNT(*) FROM platform_message_deliveries d WHERE d.message_id = pm.id AND (d.status = 3 OR d.read_at IS NOT NULL)) AS read_count,
		(SELECT COUNT(*) FROM platform_message_deliveries d WHERE d.message_id = pm.id AND d.status IN (0,1)) AS unread_count`).Order("pm.id desc").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&list).Error
	return &types.AdminPageResponse[map[string]any]{List: list, Total: total, Page: page, PageSize: pageSize}, err
}

func (s *AdminService) ListMessageDeliveries(ctx context.Context, messageID int64, page, pageSize int, filter types.AdminMessageDeliveryFilter) (*types.AdminPageResponse[map[string]any], error) {
	page, pageSize = normalizeAdminPage(page, pageSize)
	query := s.DB.WithContext(ctx).Table("platform_message_deliveries d").
		Select("d.id, d.message_id, CASE WHEN o.id IS NULL THEN 'user' ELSE 'organizer' END AS target_type, d.user_id AS target_id, COALESCE(o.name, u.nickname, '') AS target_name, CASE WHEN u.mobile = '' THEN '' ELSE CONCAT(LEFT(u.mobile, 3), '****', RIGHT(u.mobile, 4)) END AS target_mobile, CASE d.status WHEN 0 THEN 'pending' WHEN 1 THEN 'sent' WHEN 2 THEN 'failed' WHEN 3 THEN 'read' ELSE 'pending' END AS delivery_status, d.sent_at AS delivered_at, d.error AS failed_reason, CASE WHEN d.status = 3 OR d.read_at IS NOT NULL THEN 'read' ELSE 'unread' END AS read_status, d.read_at, d.created_at").
		Joins("LEFT JOIN users u ON u.id = d.user_id").
		Joins("LEFT JOIN organizers o ON o.user_id = d.user_id AND o.status = 2").
		Where("d.message_id = ?", messageID)
	if status := strings.TrimSpace(filter.DeliveryStatus); status != "" {
		statuses := map[string]int8{"pending": 0, "sent": 1, "failed": 2, "read": 3}
		value, ok := statuses[status]
		if !ok {
			return nil, errors.New("投递状态仅支持 pending、sent、failed、read")
		}
		query = query.Where("d.status = ?", value)
	}
	if readStatus := strings.TrimSpace(filter.ReadStatus); readStatus != "" {
		if readStatus == "read" {
			query = query.Where("d.status = 3 OR d.read_at IS NOT NULL")
		} else if readStatus == "unread" {
			query = query.Where("d.status <> 3 AND d.read_at IS NULL")
		} else {
			return nil, errors.New("阅读状态仅支持 read、unread")
		}
	}
	return adminMapPage(query.Order("d.id desc"), page, pageSize)
}

func (s *AdminService) CreateMessage(ctx context.Context, adminID int64, req types.PlatformMessageRequest) (int64, error) {
	req.Title = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)
	req.ContentType = strings.TrimSpace(req.ContentType)
	req.CoverImage = strings.TrimSpace(req.CoverImage)
	req.Type = strings.TrimSpace(req.Type)
	req.Target = strings.TrimSpace(req.Target)
	req.Channel = strings.TrimSpace(req.Channel)
	if req.Title == "" {
		return 0, errors.New("消息标题不能为空")
	}
	if req.Content == "" {
		return 0, errors.New("消息内容不能为空")
	}
	if req.Type == "" {
		req.Type = "system"
	}
	if req.ContentType == "" {
		req.ContentType = "text"
	}
	if req.ContentType != "text" && req.ContentType != "rich_text" {
		return 0, errors.New("content_type 仅支持 text 或 rich_text")
	}
	if req.Target == "" {
		req.Target = "all"
	}
	if req.Channel == "" {
		req.Channel = "in_app"
	}
	if req.Channel != "in_app" {
		return 0, errors.New("当前仅支持 in_app 站内消息渠道")
	}
	mediaData, err := json.Marshal(req.MediaData)
	if err != nil {
		return 0, err
	}
	msg := models.PlatformMessage{Title: req.Title, Content: req.Content, ContentType: req.ContentType, CoverImage: req.CoverImage, MediaData: string(mediaData), Type: req.Type, Target: req.Target, Channel: req.Channel, CreatorID: adminID, Status: req.Status}
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
	// C 端收件人直接落用户通知收件箱（organizers 类型的消息走主办方自己的已读模型，不在此落库）。
	// 实时提醒仍由下面的 platform_message WS 推送覆盖，这里只写库、不再额外发 MQ。
	if msg.Target != "merchant" && msg.Target != "organizer" && msg.Target != "business" && len(targetIDs) > 0 {
		content := msg.Content
		if rs := []rune(content); len(rs) > 500 { // user_notifications.content 上限 500
			content = string(rs[:500])
		}
		payload := fmt.Sprintf(`{"message_id":%d}`, msg.ID)
		notifications := make([]models.UserNotification, 0, len(targetIDs))
		for _, targetID := range targetIDs {
			notifications = append(notifications, models.UserNotification{
				UserID:  targetID,
				Type:    types.NotifyTypeSystem,
				Title:   msg.Title,
				Content: content,
				Payload: payload,
			})
		}
		// 落库失败不阻塞发布流程，仅记日志
		if err := s.DB.WithContext(ctx).Create(&notifications).Error; err != nil {
			log.Printf("[platform_message] 写入用户通知收件箱失败: message_id=%d err=%v", msg.ID, err)
		}
	}
	mediaData := []string{}
	if msg.MediaData != "" {
		_ = json.Unmarshal([]byte(msg.MediaData), &mediaData)
	}
	for _, targetID := range targetIDs {
		payload := types.PlatformMessagePayload{
			MessageID:   msg.ID,
			TargetID:    targetID,
			Title:       msg.Title,
			Content:     msg.Content,
			ContentType: msg.ContentType,
			CoverImage:  msg.CoverImage,
			MediaData:   mediaData,
			Type:        msg.Type,
			Target:      msg.Target,
			CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
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
