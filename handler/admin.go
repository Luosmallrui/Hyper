package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/models"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Admin struct {
	Config       *config.Config
	AdminService service.IAdminService
}

func (a *Admin) RegisterRouter(r gin.IRouter) {
	adminAuth := middleware.AdminAuth([]byte(a.Config.Jwt.Secret))
	// Login pages may read branding before an administrator has a token.
	r.GET("/v1/system-config", context.Wrap(a.GetPublicSystemConfig))

	g := r.Group("/v1/admin")
	g.POST("/login", context.Wrap(a.Login))

	authorized := g.Group("")
	authorized.Use(adminAuth)
	authorized.Use(a.permissionMiddleware())
	authorized.Use(a.operationLogMiddleware())
	{
		// 入驻申请管理
		authorized.GET("/profile", context.Wrap(a.GetProfile))
		authorized.PUT("/profile", context.Wrap(a.UpdateProfile))
		authorized.PUT("/profile/password", context.Wrap(a.UpdatePassword))
		authorized.GET("/admins", context.Wrap(a.ListAdmins))
		authorized.POST("/admins", context.Wrap(a.CreateAdmin))
		authorized.PUT("/admins/:id", context.Wrap(a.UpdateAdmin))
		authorized.DELETE("/admins/:id", context.Wrap(a.DeleteAdmin))
		authorized.GET("/roles", context.Wrap(a.ListRoles))
		authorized.POST("/roles", context.Wrap(a.CreateRole))
		authorized.PUT("/roles/:id", context.Wrap(a.UpdateRole))
		authorized.DELETE("/roles/:id", context.Wrap(a.DeleteRole))
		authorized.GET("/logs", context.Wrap(a.ListOperationLogs))
		authorized.GET("/categories", context.Wrap(a.ListCategories))
		authorized.POST("/categories", context.Wrap(a.CreateCategory))
		authorized.PUT("/categories/:id", context.Wrap(a.UpdateCategory))
		authorized.DELETE("/categories/:id", context.Wrap(a.DeleteCategory))

		authorized.GET("/organizers", context.Wrap(a.GetOrganizerList))
		authorized.GET("/organizers/:id", context.Wrap(a.GetOrganizerDetail))
		authorized.PUT("/organizers/:id/audit", context.Wrap(a.AuditOrganizer))
		authorized.PATCH("/organizers/:id/status", context.Wrap(a.UpdateOrganizerEnabled))
		authorized.PUT("/organizers/:id/tags", context.Wrap(a.UpdateOrganizerTags))
		authorized.PUT("/venues/:id/tags", context.Wrap(a.UpdateOrganizerTags))
		authorized.DELETE("/organizers/:id", context.Wrap(a.DeleteOrganizer))
		authorized.POST("/wechat-subscribe", context.Wrap(a.BindWechatSubscribe))

		// 派对/场地管理
		authorized.GET("/parties", context.Wrap(a.GetPartyList))
		authorized.GET("/parties/:id", context.Wrap(a.GetPartyDetail))
		authorized.PUT("/parties/:id/status", context.Wrap(a.UpdatePartyStatus))
		authorized.DELETE("/parties/:id", context.Wrap(a.HideParty))
		authorized.PUT("/parties/:id/tags", context.Wrap(a.UpdatePartyTags))

		// 活动审核管理
		authorized.GET("/activities", context.Wrap(a.GetActivityList))
		authorized.GET("/activities/:id", context.Wrap(a.GetActivityDetail))
		authorized.PUT("/activities/:id/audit", context.Wrap(a.AuditActivity))
		authorized.DELETE("/activities/:id", context.Wrap(a.HideActivity))
		authorized.PATCH("/activities/:id/visibility", context.Wrap(a.UpdateActivityVisibility))
		authorized.PUT("/activities/:id/tags", context.Wrap(a.UpdateActivityTags))
		authorized.GET("/activity-collections", context.Wrap(a.ListActivityCollections))
		authorized.POST("/activity-collections", context.Wrap(a.CreateActivityCollection))
		authorized.PUT("/activity-collections/:id", context.Wrap(a.UpdateActivityCollection))
		authorized.DELETE("/activity-collections/:id", context.Wrap(a.DeleteActivityCollection))

		// 票券管理
		authorized.GET("/tickets", context.Wrap(a.GetAllTickets))
		authorized.GET("/events/:id/tickets", context.Wrap(a.GetEventTickets))

		// 订单管理
		authorized.GET("/orders", context.Wrap(a.GetOrderList))
		authorized.GET("/orders/:order_no", context.Wrap(a.GetOrderDetail))
		authorized.POST("/orders/:order_no/refund/approve", context.Wrap(a.ApproveOrderRefund))
		authorized.POST("/orders/:order_no/refund/reject", context.Wrap(a.RejectOrderRefund))
		authorized.GET("/refunds/:refund_no", context.Wrap(a.GetRefundDetail))

		// 财务结算
		authorized.GET("/finance/summary", context.Wrap(a.GetFinanceSummary))
		authorized.GET("/finance/platform-flows", context.Wrap(a.ListPlatformFinanceFlows))
		authorized.GET("/finance/settlements", context.Wrap(a.GetFinanceSettlements))
		authorized.GET("/finance/settlements/:organizer_id", context.Wrap(a.GetFinanceSettlementDetail))
		authorized.GET("/finance/settlements/:organizer_id/export", context.Wrap(a.ExportFinanceSettlement))

		// 用户管理
		authorized.GET("/users", context.Wrap(a.GetUserList))
		authorized.PUT("/users/:id/status", context.Wrap(a.UpdateUserStatus))
		authorized.GET("/users/:id/records/:type", context.Wrap(a.ListUserRecords))
		authorized.GET("/viewers", context.Wrap(a.ListViewers))
		authorized.GET("/verifiers", context.Wrap(a.ListVerifiers))
		authorized.POST("/verifiers", context.Wrap(a.CreateVerifier))
		authorized.PUT("/verifiers/:id", context.Wrap(a.UpdateVerifier))
		authorized.PATCH("/verifiers/:id/status", context.Wrap(a.UpdateVerifierStatus))
		authorized.DELETE("/verifiers/:id", context.Wrap(a.DeleteVerifier))
		authorized.GET("/verification-records", context.Wrap(a.ListVerificationRecords))
		authorized.GET("/notes", context.Wrap(a.ListNotes))
		authorized.PUT("/notes/:id/status", context.Wrap(a.UpdateNoteStatus))
		authorized.GET("/notes/:id/records/:type", context.Wrap(a.ListNoteRecords))
		authorized.PATCH("/notes/:id/comments/:comment_id/status", context.Wrap(a.UpdateCommentStatus))
		authorized.GET("/note-interactions", context.Wrap(a.ListNoteInteractions))
		authorized.GET("/note-comments", context.Wrap(a.ListNoteComments))
		authorized.PATCH("/note-comments/:comment_id/status", context.Wrap(a.UpdateGlobalCommentStatus))
		authorized.GET("/points/logs", context.Wrap(a.ListPointLogs))
		authorized.POST("/points/adjust", context.Wrap(a.AdjustPoints))
		authorized.GET("/points/rules", context.Wrap(a.GetPointsRule))
		authorized.PUT("/points/rules", context.Wrap(a.UpdatePointsRule))
		authorized.GET("/withdraws", context.Wrap(a.ListWithdraws))
		authorized.PUT("/withdraws/:id/audit", context.Wrap(a.AuditWithdraw))
		authorized.GET("/bank-account-audits", context.Wrap(a.ListBankAccountAudits))
		authorized.PUT("/bank-account-audits/:id/audit", context.Wrap(a.AuditBankAccount))
		authorized.GET("/organizer-level-rules", context.Wrap(a.ListOrganizerLevelRules))
		authorized.POST("/organizer-level-rules", context.Wrap(a.CreateOrganizerLevelRule))
		authorized.PUT("/organizer-level-rules/:id", context.Wrap(a.UpdateOrganizerLevelRule))
		authorized.DELETE("/organizer-level-rules/:id", context.Wrap(a.DeleteOrganizerLevelRule))
		authorized.GET("/messages", context.Wrap(a.ListMessages))
		authorized.POST("/messages", context.Wrap(a.CreateMessage))
		authorized.GET("/messages/:id/records", context.Wrap(a.ListMessageDeliveries))

		// 内容管理
		authorized.GET("/banners", context.Wrap(a.ListBanners))
		authorized.POST("/banners", context.Wrap(a.CreateBanner))
		authorized.PUT("/banners/sort", context.Wrap(a.SortBanners))
		authorized.PUT("/banners/:id", context.Wrap(a.UpdateBanner))
		authorized.DELETE("/banners/:id", context.Wrap(a.DeleteBanner))

		// 平台设置
		authorized.GET("/settings", context.Wrap(a.GetSettings))
		authorized.PUT("/settings", context.Wrap(a.UpdateSettings))
		authorized.GET("/system-config", context.Wrap(a.GetSystemConfig))
		authorized.PUT("/system-config", context.Wrap(a.UpdateSystemConfig))

		// 数据概览
		authorized.GET("/dashboard", context.Wrap(a.GetDashboard))
	}
}

func (a *Admin) permissionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := a.AdminService.CheckPermission(c.Request.Context(), int64(c.GetInt("admin_id")), c.Request.Method, c.FullPath()); err != nil {
			required := service.RequiredAdminPermission(c.Request.Method, c.FullPath())
			_ = a.AdminService.RecordOperationLog(c.Request.Context(), models.AdminOperationLog{
				AdminID: int64(c.GetInt("admin_id")), Action: "admin.permission.denied", Resource: "permission", ResourceType: "permission", ResourceID: required, ResourceName: required,
				Method: c.Request.Method, Path: c.Request.URL.Path, IP: c.ClientIP(), Result: "denied", ErrorCode: "ADMIN_PERMISSION_DENIED", ErrorMessage: err.Error(), Remark: "所需权限：" + required,
			})
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code": 403, "msg": "无权执行该操作", "message": "无权执行该操作",
				"error_code": "ADMIN_PERMISSION_DENIED", "data": gin.H{"required_permission": required},
			})
			return
		}
		c.Next()
	}
}

func (a *Admin) operationLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Request.Method == http.MethodGet {
			return
		}
		adminID := int64(c.GetInt("admin_id"))
		if adminID == 0 {
			return
		}
		meta := adminAuditMetaForRequest(c)
		result, errorCode, errorMessage := "success", "", ""
		if message, ok := c.Get("request_handler_error"); ok {
			result = "failed"
			errorCode = "ADMIN_OPERATION_FAILED"
			errorMessage, _ = message.(string)
			if code, ok := c.Get("request_handler_error_code"); ok {
				errorCode, _ = code.(string)
			}
		} else if c.Writer.Status() >= http.StatusBadRequest {
			result = "failed"
			errorCode = "HTTP_" + strconv.Itoa(c.Writer.Status())
			errorMessage = "管理接口处理失败"
		}
		_ = a.AdminService.RecordOperationLog(c.Request.Context(), models.AdminOperationLog{
			AdminID: adminID, Action: meta.Action, Resource: meta.ResourceType, ResourceType: meta.ResourceType, ResourceID: meta.ResourceID, ResourceName: meta.ResourceName,
			Method: c.Request.Method, Path: c.Request.URL.Path, IP: c.ClientIP(), Remark: meta.Remark, Result: result, ErrorCode: errorCode, ErrorMessage: errorMessage,
		})
	}
}

type adminAuditMeta struct {
	Action       string
	ResourceType string
	ResourceID   string
	ResourceName string
	Remark       string
}

const adminAuditMetaKey = "admin_audit_meta"

func setAdminAuditMeta(c *gin.Context, meta adminAuditMeta) { c.Set(adminAuditMetaKey, meta) }

func adminAuditMetaForRequest(c *gin.Context) adminAuditMeta {
	if value, ok := c.Get(adminAuditMetaKey); ok {
		if meta, ok := value.(adminAuditMeta); ok {
			return meta
		}
	}
	path := c.FullPath()
	meta := adminAuditMeta{Action: "admin.unknown." + strings.ToLower(c.Request.Method), ResourceType: "admin"}
	switch {
	case path == "/v1/admin/profile":
		meta = adminAuditMeta{Action: "admin.profile.update", ResourceType: "admin"}
	case path == "/v1/admin/settings":
		meta = adminAuditMeta{Action: "admin.settings.update", ResourceType: "settings"}
	case strings.HasPrefix(path, "/v1/admin/categories"):
		meta = adminAuditMeta{Action: "admin.category." + requestAction(c.Request.Method), ResourceType: "category"}
	case strings.HasPrefix(path, "/v1/admin/admins"):
		meta = adminAuditMeta{Action: "admin.account." + requestAction(c.Request.Method), ResourceType: "admin"}
	case strings.HasPrefix(path, "/v1/admin/roles"):
		meta = adminAuditMeta{Action: "admin.role." + requestAction(c.Request.Method), ResourceType: "role"}
	case strings.HasPrefix(path, "/v1/admin/organizers"):
		meta = adminAuditMeta{Action: "admin.organizer." + requestAction(c.Request.Method), ResourceType: "organizer"}
	case strings.HasPrefix(path, "/v1/admin/venues"):
		meta = adminAuditMeta{Action: "admin.venue." + requestAction(c.Request.Method), ResourceType: "venue"}
	case strings.HasPrefix(path, "/v1/admin/activities"):
		meta = adminAuditMeta{Action: "admin.activity." + requestAction(c.Request.Method), ResourceType: "activity"}
	case strings.HasPrefix(path, "/v1/admin/activity-collections"):
		meta = adminAuditMeta{Action: "admin.activity_collection." + requestAction(c.Request.Method), ResourceType: "activity_collection"}
	case strings.HasPrefix(path, "/v1/admin/parties"):
		meta = adminAuditMeta{Action: "admin.party." + requestAction(c.Request.Method), ResourceType: "party"}
	case strings.HasPrefix(path, "/v1/admin/orders"), strings.HasPrefix(path, "/v1/admin/refunds"):
		meta = adminAuditMeta{Action: "admin.refund." + requestAction(c.Request.Method), ResourceType: "refund"}
	case strings.HasPrefix(path, "/v1/admin/withdraws"):
		meta = adminAuditMeta{Action: "admin.withdraw." + requestAction(c.Request.Method), ResourceType: "withdraw"}
	case strings.HasPrefix(path, "/v1/admin/users"):
		meta = adminAuditMeta{Action: "admin.user." + requestAction(c.Request.Method), ResourceType: "user"}
	case strings.HasPrefix(path, "/v1/admin/verifiers"):
		meta = adminAuditMeta{Action: "admin.verifier." + requestAction(c.Request.Method), ResourceType: "verifier"}
	case strings.HasPrefix(path, "/v1/admin/points/adjust"):
		meta = adminAuditMeta{Action: "admin.points.adjust", ResourceType: "points"}
	case strings.HasPrefix(path, "/v1/admin/points"):
		meta = adminAuditMeta{Action: "admin.points.update", ResourceType: "points"}
	case strings.HasPrefix(path, "/v1/admin/bank-account-audits"):
		meta = adminAuditMeta{Action: "admin.bank_account.audit", ResourceType: "bank_account"}
	case strings.HasPrefix(path, "/v1/admin/organizer-level-rules"):
		meta = adminAuditMeta{Action: "admin.organizer_level_rule." + requestAction(c.Request.Method), ResourceType: "organizer_level_rule"}
	case strings.HasPrefix(path, "/v1/admin/messages"):
		meta = adminAuditMeta{Action: "admin.message." + requestAction(c.Request.Method), ResourceType: "message"}
	case strings.HasPrefix(path, "/v1/admin/banners"):
		meta = adminAuditMeta{Action: "admin.banner." + requestAction(c.Request.Method), ResourceType: "banner"}
	case strings.HasPrefix(path, "/v1/admin/notes"), strings.HasPrefix(path, "/v1/admin/note-comments"):
		meta = adminAuditMeta{Action: "admin.note." + requestAction(c.Request.Method), ResourceType: "note"}
	}
	for _, key := range []string{"id", "order_no", "organizer_id", "comment_id"} {
		if id := c.Param(key); id != "" {
			meta.ResourceID = id
			break
		}
	}
	return meta
}

func requestAction(method string) string {
	switch method {
	case http.MethodPost:
		return "create"
	case http.MethodPut, http.MethodPatch:
		return "update"
	case http.MethodDelete:
		return "delete"
	default:
		return strings.ToLower(method)
	}
}

// Login 管理员登录
// POST /api/v1/admin/login
func (a *Admin) Login(c *gin.Context) error {
	var req types.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, "参数格式错误")
	}

	accessToken, refreshToken, err := a.AdminService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		return response.NewError(400, err.Error())
	}

	response.Success(c, types.UserToken{
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		AccessExpire:  time.Now().Add(2 * time.Hour).Unix(),
		RefreshExpire: time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	return nil
}

// BindWechatSubscribe 绑定管理员微信订阅通知
// POST /api/v1/admin/wechat-subscribe
func (a *Admin) BindWechatSubscribe(c *gin.Context) error {
	var req types.AdminWechatSubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, "参数格式错误")
	}
	adminID := int64(c.GetInt("admin_id"))
	if adminID == 0 {
		return response.NewError(401, "管理员身份无效")
	}
	if err := a.AdminService.BindAdminWechatSubscriber(c.Request.Context(), adminID, req.Code); err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) GetProfile(c *gin.Context) error {
	resp, err := a.AdminService.GetAdminProfile(c.Request.Context(), int64(c.GetInt("admin_id")))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) UpdateProfile(c *gin.Context) error {
	var req types.AdminProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	if err := a.AdminService.UpdateAdminProfile(c.Request.Context(), int64(c.GetInt("admin_id")), req); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) UpdatePassword(c *gin.Context) error {
	var req types.AdminPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	if err := a.AdminService.UpdateAdminPassword(c.Request.Context(), int64(c.GetInt("admin_id")), req); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) ListAdmins(c *gin.Context) error {
	resp, err := a.AdminService.ListAdmins(c.Request.Context(), adminPage(c), adminPageSize(c), c.Query("keyword"))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) CreateAdmin(c *gin.Context) error {
	var req types.AdminAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	id, err := a.AdminService.CreateAdmin(c.Request.Context(), req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": id})
	return nil
}

func (a *Admin) UpdateAdmin(c *gin.Context) error {
	var req types.AdminAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	if req.Status == models.AdminStatusDeactivate {
		setAdminAuditMeta(c, adminAuditMeta{Action: "admin.account.disable", ResourceType: "admin", ResourceID: c.Param("id"), Remark: "停用管理员账号"})
	}
	if err := a.AdminService.UpdateAdmin(c.Request.Context(), int64(c.GetInt("admin_id")), adminParamID(c), req); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) DeleteAdmin(c *gin.Context) error {
	if err := a.AdminService.DeleteAdmin(c.Request.Context(), int64(c.GetInt("admin_id")), adminParamID(c)); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) ListRoles(c *gin.Context) error {
	resp, err := a.AdminService.ListRoles(c.Request.Context(), adminPage(c), adminPageSize(c), c.Query("keyword"))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) CreateRole(c *gin.Context) error {
	var req types.AdminRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	id, err := a.AdminService.SaveRole(c.Request.Context(), 0, req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": id})
	return nil
}

func (a *Admin) UpdateRole(c *gin.Context) error {
	var req types.AdminRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	id, err := a.AdminService.SaveRole(c.Request.Context(), adminParamID(c), req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": id})
	return nil
}

func (a *Admin) DeleteRole(c *gin.Context) error {
	if err := a.AdminService.DeleteRole(c.Request.Context(), adminParamID(c)); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) ListOperationLogs(c *gin.Context) error {
	filter := types.AdminOperationLogFilter{
		AdminID: adminQueryInt64(c, "admin_id"), Action: c.Query("action"), ResourceType: c.Query("resource_type"),
		Result: c.Query("result"), Keyword: c.Query("keyword"), StartDate: adminQueryTime(c, "start_date"), EndDate: adminQueryTime(c, "end_date"),
	}
	resp, err := a.AdminService.ListOperationLogs(c.Request.Context(), adminPage(c), adminPageSize(c), filter)
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) ListCategories(c *gin.Context) error {
	resp, err := a.AdminService.ListCategories(c.Request.Context(), adminPage(c), adminPageSize(c), c.Query("type"))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) CreateCategory(c *gin.Context) error {
	var req types.AdminCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	id, err := a.AdminService.SaveCategory(c.Request.Context(), 0, req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": id})
	return nil
}

func (a *Admin) UpdateCategory(c *gin.Context) error {
	var req types.AdminCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	id, err := a.AdminService.SaveCategory(c.Request.Context(), adminParamID(c), req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": id})
	return nil
}

func (a *Admin) DeleteCategory(c *gin.Context) error {
	if err := a.AdminService.DeleteCategory(c.Request.Context(), adminParamID(c)); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) UpdateActivityTags(c *gin.Context) error {
	return a.updateContentTags(c, models.ContentTagTargetActivity, "activity")
}

func (a *Admin) UpdateOrganizerTags(c *gin.Context) error {
	return a.updateContentTags(c, models.ContentTagTargetVenue, "venue")
}

func (a *Admin) UpdatePartyTags(c *gin.Context) error {
	return a.updateContentTags(c, models.ContentTagTargetParty, "party")
}

func (a *Admin) updateContentTags(c *gin.Context, targetType, resourceType string) error {
	id := adminParamID(c)
	if id <= 0 {
		return response.NewError(http.StatusBadRequest, "无效的ID")
	}
	var req types.ContentTagBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数格式错误")
	}
	if err := a.AdminService.SaveContentTags(c.Request.Context(), targetType, id, req.TagIDs); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewError(http.StatusNotFound, "目标不存在")
		}
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	setAdminAuditMeta(c, adminAuditMeta{Action: "admin." + resourceType + ".tags.update", ResourceType: resourceType, ResourceID: strconv.FormatInt(id, 10), Remark: "更新优惠标签"})
	response.Success(c, gin.H{"success": true, "tag_ids": req.TagIDs})
	return nil
}

func (a *Admin) ListUserRecords(c *gin.Context) error {
	resp, err := a.AdminService.ListUserRecords(c.Request.Context(), adminParamID(c), c.Param("type"), adminPage(c), adminPageSize(c))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) ListViewers(c *gin.Context) error {
	resp, err := a.AdminService.ListViewers(c.Request.Context(), adminPage(c), adminPageSize(c), c.Query("keyword"))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) ListActivityCollections(c *gin.Context) error {
	resp, err := a.AdminService.ListActivityCollections(c.Request.Context(), adminPage(c), adminPageSize(c), c.Query("keyword"), adminQueryInt64(c, "organizer_id"))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) CreateActivityCollection(c *gin.Context) error {
	var req types.ActivityCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	id, err := a.AdminService.SaveActivityCollection(c.Request.Context(), 0, req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": id})
	return nil
}

func (a *Admin) UpdateActivityCollection(c *gin.Context) error {
	var req types.ActivityCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	id, err := a.AdminService.SaveActivityCollection(c.Request.Context(), adminParamID(c), req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": id})
	return nil
}

func (a *Admin) DeleteActivityCollection(c *gin.Context) error {
	if err := a.AdminService.DeleteActivityCollection(c.Request.Context(), adminParamID(c)); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) ListVerifiers(c *gin.Context) error {
	resp, err := a.AdminService.ListVerifiers(c.Request.Context(), adminPage(c), adminPageSize(c), c.Query("keyword"), adminQueryInt64(c, "organizer_id"))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) CreateVerifier(c *gin.Context) error {
	var req types.AdminVerifierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	id, err := a.AdminService.SaveVerifier(c.Request.Context(), 0, req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": id})
	return nil
}

func (a *Admin) UpdateVerifier(c *gin.Context) error {
	var req types.AdminVerifierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	id, err := a.AdminService.SaveVerifier(c.Request.Context(), adminParamID(c), req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": id})
	return nil
}

func (a *Admin) UpdateVerifierStatus(c *gin.Context) error {
	var req struct {
		Status int8 `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	if err := a.AdminService.UpdateVerifierStatus(c.Request.Context(), adminParamID(c), req.Status); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) DeleteVerifier(c *gin.Context) error {
	if err := a.AdminService.DeleteVerifier(c.Request.Context(), adminParamID(c)); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) ListVerificationRecords(c *gin.Context) error {
	resp, err := a.AdminService.ListVerificationRecords(c.Request.Context(), adminPage(c), adminPageSize(c), c.Query("keyword"), adminQueryInt64(c, "organizer_id"))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) ListNotes(c *gin.Context) error {
	var status *int
	if raw := c.Query("status"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil {
			return err
		}
		status = &value
	}
	resp, err := a.AdminService.ListNotes(c.Request.Context(), adminPage(c), adminPageSize(c), status, c.Query("keyword"))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) UpdateNoteStatus(c *gin.Context) error {
	var req types.AdminNoteStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "动态状态仅支持 -1删除、0隐藏、1公开")
	}
	if err := a.AdminService.UpdateNoteStatus(c.Request.Context(), adminParamID(c), int(req.Status)); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) ListNoteRecords(c *gin.Context) error {
	resp, err := a.AdminService.ListNoteRecords(c.Request.Context(), c.Param("type"), adminPage(c), adminPageSize(c), adminParamID(c))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) UpdateCommentStatus(c *gin.Context) error {
	var req types.AdminCommentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, "参数格式错误")
	}
	commentID, err := strconv.ParseInt(c.Param("comment_id"), 10, 64)
	if err != nil {
		return response.NewError(400, "评论ID无效")
	}
	setAdminAuditMeta(c, adminAuditMeta{Action: "admin.comment.moderate", ResourceType: "comment", ResourceID: strconv.FormatInt(commentID, 10), Remark: req.Reason})
	if err := a.AdminService.UpdateCommentStatus(c.Request.Context(), adminParamID(c), commentID, int64(c.GetInt("admin_id")), req.Status, req.Reason); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewError(404, "评论不存在")
		}
		return response.NewError(400, err.Error())
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) ListNoteInteractions(c *gin.Context) error {
	filter := types.AdminNoteInteractionFilter{Type: c.Query("type"), NoteID: adminQueryInt64(c, "note_id"), UserID: adminQueryInt64(c, "user_id"), Keyword: c.Query("keyword"), Channel: c.Query("channel"), StartDate: adminQueryTime(c, "start_date"), EndDate: adminQueryTime(c, "end_date")}
	resp, err := a.AdminService.ListNoteInteractions(c.Request.Context(), adminPage(c), adminPageSize(c), filter)
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) ListNoteComments(c *gin.Context) error {
	var status *int8
	if raw := c.Query("status"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 8)
		if err != nil {
			return response.NewError(400, "评论状态无效")
		}
		parsed := int8(value)
		status = &parsed
	}
	filter := types.AdminNoteCommentFilter{Status: status, NoteID: adminQueryInt64(c, "note_id"), UserID: adminQueryInt64(c, "user_id"), Keyword: c.Query("keyword"), StartDate: adminQueryTime(c, "start_date"), EndDate: adminQueryTime(c, "end_date")}
	resp, err := a.AdminService.ListNoteComments(c.Request.Context(), adminPage(c), adminPageSize(c), filter)
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) UpdateGlobalCommentStatus(c *gin.Context) error {
	commentID, err := strconv.ParseInt(c.Param("comment_id"), 10, 64)
	if err != nil {
		return response.NewError(400, "评论ID无效")
	}
	var req types.AdminCommentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, "参数格式错误")
	}
	setAdminAuditMeta(c, adminAuditMeta{Action: "admin.comment.moderate", ResourceType: "comment", ResourceID: strconv.FormatInt(commentID, 10), Remark: req.Reason})
	if err := a.AdminService.UpdateCommentStatus(c.Request.Context(), 0, commentID, int64(c.GetInt("admin_id")), req.Status, req.Reason); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) ListPointLogs(c *gin.Context) error {
	resp, err := a.AdminService.ListPointLogs(c.Request.Context(), adminPage(c), adminPageSize(c), adminQueryInt64(c, "user_id"))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) AdjustPoints(c *gin.Context) error {
	var req types.AdminPointsAdjustRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, "参数格式错误")
	}
	setAdminAuditMeta(c, adminAuditMeta{Action: "admin.points.adjust", ResourceType: "points", ResourceID: strconv.FormatInt(req.UserID, 10), Remark: req.Reason})
	resp, err := a.AdminService.AdjustPoints(c.Request.Context(), int64(c.GetInt("admin_id")), req)
	if err != nil {
		return response.NewError(400, err.Error())
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) GetPointsRule(c *gin.Context) error {
	resp, err := a.AdminService.GetPointsRule(c.Request.Context())
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) UpdatePointsRule(c *gin.Context) error {
	var req types.UpdatePointsRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, "参数格式错误")
	}
	if err := a.AdminService.UpdatePointsRule(c.Request.Context(), req); err != nil {
		return response.NewError(400, err.Error())
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) ListWithdraws(c *gin.Context) error {
	var status *int8
	if raw := c.Query("status"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 8)
		if err != nil {
			return err
		}
		v := int8(value)
		status = &v
	}
	resp, err := a.AdminService.ListWithdraws(c.Request.Context(), adminPage(c), adminPageSize(c), status, adminQueryInt64(c, "organizer_id"))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) AuditWithdraw(c *gin.Context) error {
	var req types.WithdrawAuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	action := "admin.withdraw.reject"
	if req.Status == 1 {
		action = "admin.withdraw.approve"
	}
	setAdminAuditMeta(c, adminAuditMeta{Action: action, ResourceType: "withdraw", ResourceID: c.Param("id"), Remark: req.Remark})
	if err := a.AdminService.AuditWithdraw(c.Request.Context(), adminParamID(c), req); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) ListBankAccountAudits(c *gin.Context) error {
	var status *int8
	if raw := c.Query("status"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 8)
		if err != nil {
			return err
		}
		v := int8(value)
		status = &v
	}
	resp, err := a.AdminService.ListBankAccountAudits(c.Request.Context(), adminPage(c), adminPageSize(c), status, adminQueryInt64(c, "organizer_id"))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) AuditBankAccount(c *gin.Context) error {
	var req types.BankAccountAuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	action := "admin.bank_account.reject"
	if req.Status == models.OrganizerBankAuditStatusApproved {
		action = "admin.bank_account.approve"
	}
	setAdminAuditMeta(c, adminAuditMeta{Action: action, ResourceType: "bank_account", ResourceID: c.Param("id"), Remark: req.RejectReason})
	if err := a.AdminService.AuditBankAccount(c.Request.Context(), adminParamID(c), req); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) ListOrganizerLevelRules(c *gin.Context) error {
	resp, err := a.AdminService.ListOrganizerLevelRules(c.Request.Context())
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"list": resp})
	return nil
}

func (a *Admin) CreateOrganizerLevelRule(c *gin.Context) error {
	var req types.OrganizerLevelRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	id, err := a.AdminService.SaveOrganizerLevelRule(c.Request.Context(), 0, req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": id})
	return nil
}

func (a *Admin) UpdateOrganizerLevelRule(c *gin.Context) error {
	var req types.OrganizerLevelRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	id, err := a.AdminService.SaveOrganizerLevelRule(c.Request.Context(), adminParamID(c), req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": id})
	return nil
}

func (a *Admin) DeleteOrganizerLevelRule(c *gin.Context) error {
	if err := a.AdminService.DeleteOrganizerLevelRule(c.Request.Context(), adminParamID(c)); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) ListMessages(c *gin.Context) error {
	resp, err := a.AdminService.ListMessages(c.Request.Context(), adminPage(c), adminPageSize(c), c.Query("target"), c.Query("type"))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) CreateMessage(c *gin.Context) error {
	var req types.PlatformMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	id, err := a.AdminService.CreateMessage(c.Request.Context(), int64(c.GetInt("admin_id")), req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": id})
	return nil
}

func (a *Admin) ListMessageDeliveries(c *gin.Context) error {
	filter := types.AdminMessageDeliveryFilter{DeliveryStatus: c.Query("delivery_status"), ReadStatus: c.Query("read_status")}
	resp, err := a.AdminService.ListMessageDeliveries(c.Request.Context(), adminParamID(c), adminPage(c), adminPageSize(c), filter)
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

// GetOrganizerList 获取入驻申请列表
// GET /api/v1/admin/organizers?page=1&pageSize=20&status=1&type=venue
func (a *Admin) GetOrganizerList(c *gin.Context) error {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	organizerType := c.Query("type")
	var status *int8
	if raw := c.Query("status"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 8)
		if err != nil {
			return response.NewError(400, "无效的状态")
		}
		s := int8(v)
		status = &s
	}
	result, err := a.AdminService.GetOrganizerList(c.Request.Context(), page, pageSize, status, organizerType)
	if err != nil {
		return response.NewError(500, "查询失败: "+err.Error())
	}
	response.Success(c, result)
	return nil
}

// GetOrganizerDetail 获取入驻申请详情
// GET /api/v1/admin/organizers/:id
func (a *Admin) GetOrganizerDetail(c *gin.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.NewError(400, "无效的ID")
	}
	detail, err := a.AdminService.GetOrganizerDetail(c.Request.Context(), id)
	if err != nil {
		return response.NewError(404, "入驻申请不存在")
	}
	response.Success(c, detail)
	return nil
}

// AuditOrganizer 审核入驻申请
// PUT /api/v1/admin/organizers/:id/audit
func (a *Admin) AuditOrganizer(c *gin.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.NewError(400, "无效的ID")
	}
	var req types.AdminAuditOrganizerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, "参数格式错误")
	}
	action := "admin.organizer.reject"
	if req.Status == models.OrganizerStatusApproved {
		action = "admin.organizer.approve"
	}
	setAdminAuditMeta(c, adminAuditMeta{Action: action, ResourceType: "organizer", ResourceID: strconv.FormatInt(id, 10), Remark: req.RejectReason})
	if err := a.AdminService.AuditOrganizer(c.Request.Context(), id, req); err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, "审核完成")
	return nil
}

func (a *Admin) UpdateOrganizerEnabled(c *gin.Context) error {
	var req struct {
		Enabled int8 `json:"enabled" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, "参数格式错误")
	}
	if err := a.AdminService.UpdateOrganizerEnabled(c.Request.Context(), adminParamID(c), req.Enabled); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewError(404, "商家不存在")
		}
		return response.NewError(400, err.Error())
	}
	response.Success(c, gin.H{"success": true, "enabled": req.Enabled})
	return nil
}

// DeleteOrganizer 删除入驻商家
// DELETE /api/v1/admin/organizers/:id
func (a *Admin) DeleteOrganizer(c *gin.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.NewError(400, "无效的ID")
	}
	if err := a.AdminService.DeleteOrganizer(c.Request.Context(), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewError(404, "商家不存在")
		}
		return response.NewError(500, err.Error())
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

// GetPartyList 获取所有派对/场地列表
// GET /api/v1/admin/parties?page=1&pageSize=20&keyword=xxx&type=派对
func (a *Admin) GetPartyList(c *gin.Context) error {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	partyType := c.Query("type") // "派对" 或 "场地"，空=全部

	result, err := a.AdminService.GetPartyList(c.Request.Context(), page, pageSize, keyword, partyType)
	if err != nil {
		return response.NewError(500, "查询失败: "+err.Error())
	}

	response.Success(c, result)
	return nil
}

// GetPartyDetail 获取派对/场地详情
// GET /api/v1/admin/parties/:id
func (a *Admin) GetPartyDetail(c *gin.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.NewError(400, "无效的ID")
	}

	detail, err := a.AdminService.GetPartyDetail(c.Request.Context(), id)
	if err != nil {
		return response.NewError(404, "记录不存在")
	}

	response.Success(c, detail)
	return nil
}

// UpdatePartyStatus 更新派对状态
// PUT /api/v1/admin/parties/:id/status
func (a *Admin) UpdatePartyStatus(c *gin.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.NewError(400, "无效的ID")
	}

	var req types.AdminUpdatePartyStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, "参数格式错误")
	}

	if err := a.AdminService.UpdatePartyStatus(c.Request.Context(), id, req.Status); err != nil {
		return response.NewError(500, err.Error())
	}

	response.Success(c, "状态已更新")
	return nil
}

// HideParty keeps legacy party/venue records and removes them from public
// discovery by setting the existing status field to offline.
// DELETE /api/v1/admin/parties/:id
func (a *Admin) HideParty(c *gin.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.NewError(400, "无效的ID")
	}
	setAdminAuditMeta(c, adminAuditMeta{Action: "admin.party.hide", ResourceType: "party", ResourceID: strconv.FormatInt(id, 10)})
	if err := a.AdminService.UpdatePartyStatus(c.Request.Context(), id, "offline"); err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, gin.H{"id": id, "status": "offline"})
	return nil
}

// GetActivityList 获取活动审核列表
// GET /api/v1/admin/activities?page=1&pageSize=20&status=1&keyword=xxx&organizer_id=1
func (a *Admin) GetActivityList(c *gin.Context) error {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")

	var status *int8
	if raw := c.Query("status"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 8)
		if err != nil {
			return response.NewError(400, "无效的状态")
		}
		s := int8(v)
		status = &s
	}

	var organizerID int64
	if raw := c.Query("organizer_id"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return response.NewError(400, "无效的主办方ID")
		}
		organizerID = v
	}

	var isHidden *int8
	if raw := c.Query("is_hidden"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 8)
		if err != nil || (v != 0 && v != 1) {
			return response.NewError(400, "is_hidden 仅支持 0 或 1")
		}
		value := int8(v)
		isHidden = &value
	}

	filter := types.AdminActivityFilter{
		Status:        status,
		IsHidden:      isHidden,
		Keyword:       keyword,
		OrganizerID:   organizerID,
		PublishedFrom: adminQueryTime(c, "published_from"),
		PublishedTo:   adminQueryTime(c, "published_to"),
		ActivityFrom:  adminQueryTime(c, "activity_from"),
		ActivityTo:    adminQueryTime(c, "activity_to"),
	}
	result, err := a.AdminService.GetActivityList(c.Request.Context(), page, pageSize, filter)
	if err != nil {
		return response.NewError(500, "查询失败: "+err.Error())
	}
	response.Success(c, result)
	return nil
}

// GetActivityDetail 获取活动审核详情
// GET /api/v1/admin/activities/:id
func (a *Admin) GetActivityDetail(c *gin.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.NewError(400, "无效的活动ID")
	}
	detail, err := a.AdminService.GetActivityDetail(c.Request.Context(), id)
	if err != nil {
		return response.NewError(404, "活动不存在")
	}
	response.Success(c, detail)
	return nil
}

// AuditActivity 审核活动
// PUT /api/v1/admin/activities/:id/audit
func (a *Admin) AuditActivity(c *gin.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.NewError(400, "无效的活动ID")
	}
	var req types.AdminAuditActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, "参数格式错误")
	}
	action := "admin.activity.reject"
	if req.Status == models.ActivityStatusOnline {
		action = "admin.activity.approve"
	}
	setAdminAuditMeta(c, adminAuditMeta{Action: action, ResourceType: "activity", ResourceID: strconv.FormatInt(id, 10), Remark: req.RejectReason})
	if err := a.AdminService.AuditActivity(c.Request.Context(), id, req); err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, "审核完成")
	return nil
}

// HideActivity uses soft hiding as the default management-side delete action.
// DELETE /api/v1/admin/activities/:id?reason=违规或下架原因
func (a *Admin) HideActivity(c *gin.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.NewError(400, "无效的活动ID")
	}
	setAdminAuditMeta(c, adminAuditMeta{
		Action: "admin.activity.hide", ResourceType: "activity", ResourceID: strconv.FormatInt(id, 10), Remark: strings.TrimSpace(c.Query("reason")),
	})
	activity, err := a.AdminService.SetActivityVisibility(c.Request.Context(), id, false, c.Query("reason"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewError(404, "活动不存在")
		}
		return response.NewError(500, err.Error())
	}
	setAdminAuditMeta(c, adminAuditMeta{
		Action: "admin.activity.hide", ResourceType: "activity", ResourceID: strconv.FormatInt(id, 10), ResourceName: activity.Name,
		Remark: activity.HiddenReason,
	})
	response.Success(c, gin.H{"id": id, "is_hidden": true, "hidden_at": activity.HiddenAt, "hidden_reason": activity.HiddenReason})
	return nil
}

// UpdateActivityVisibility restores or hides an activity without deleting its
// orders or other historical data.
// PATCH /api/v1/admin/activities/:id/visibility
func (a *Admin) UpdateActivityVisibility(c *gin.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.NewError(400, "无效的活动ID")
	}
	var req types.AdminActivityVisibilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, "参数格式错误")
	}
	action := "admin.activity.hide"
	if *req.Visible {
		action = "admin.activity.restore"
	}
	setAdminAuditMeta(c, adminAuditMeta{
		Action: action, ResourceType: "activity", ResourceID: strconv.FormatInt(id, 10), Remark: strings.TrimSpace(req.Reason),
	})
	activity, err := a.AdminService.SetActivityVisibility(c.Request.Context(), id, *req.Visible, req.Reason)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewError(404, "活动不存在")
		}
		return response.NewError(500, err.Error())
	}
	setAdminAuditMeta(c, adminAuditMeta{
		Action: action, ResourceType: "activity", ResourceID: strconv.FormatInt(id, 10), ResourceName: activity.Name,
		Remark: activity.HiddenReason,
	})
	response.Success(c, gin.H{"id": id, "is_hidden": activity.IsHidden == 1, "hidden_at": activity.HiddenAt, "hidden_reason": activity.HiddenReason})
	return nil
}

// GetAllTickets 获取所有票券列表
// GET /api/v1/admin/tickets?page=1&pageSize=20&keyword=xxx
func (a *Admin) GetAllTickets(c *gin.Context) error {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")

	result, err := a.AdminService.GetAllTickets(c.Request.Context(), page, pageSize, keyword)
	if err != nil {
		return response.NewError(500, "查询失败: "+err.Error())
	}

	response.Success(c, result)
	return nil
}

// GetEventTickets 获取某个活动的票券列表
// GET /api/v1/admin/events/:id/tickets?page=1&pageSize=20
func (a *Admin) GetEventTickets(c *gin.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.NewError(400, "无效的活动ID")
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	result, err := a.AdminService.GetEventTicketList(c.Request.Context(), id, page, pageSize)
	if err != nil {
		return response.NewError(500, "查询失败: "+err.Error())
	}

	response.Success(c, result)
	return nil
}

// GetOrderList 获取订单列表
// GET /api/v1/admin/orders?page=1&pageSize=20
func (a *Admin) GetOrderList(c *gin.Context) error {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	activityID, _ := strconv.ParseInt(c.Query("activity_id"), 10, 64)
	keyword := c.Query("keyword")
	var status *int8
	if raw := c.Query("status"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 8)
		if err != nil {
			return response.NewError(400, "无效的订单状态")
		}
		s := int8(v)
		status = &s
	}
	refundStatus, err := adminRefundStatusQuery(c.Query("refund_status"))
	if err != nil {
		return response.NewError(400, err.Error())
	}
	if _, err := service.NormalizeSalesChannel(c.Query("sales_channel"), false); err != nil {
		return response.NewError(400, err.Error())
	}

	result, err := a.AdminService.GetTicketOrderList(c.Request.Context(), page, pageSize, activityID, status, refundStatus, keyword, c.Query("sales_channel"))
	if err != nil {
		return response.NewError(500, "查询失败: "+err.Error())
	}

	response.Success(c, result)
	return nil
}

func (a *Admin) GetOrderDetail(c *gin.Context) error {
	resp, err := a.AdminService.GetTicketOrderDetail(c.Request.Context(), c.Param("order_no"))
	if err != nil {
		return response.NewError(404, err.Error())
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) GetRefundDetail(c *gin.Context) error {
	setAdminAuditMeta(c, adminAuditMeta{Action: "admin.refund.view", ResourceType: "refund", ResourceID: c.Param("refund_no"), ResourceName: "退款单 " + c.Param("refund_no")})
	resp, err := a.AdminService.GetRefundDetail(c.Request.Context(), c.Param("refund_no"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewError(http.StatusNotFound, "退款单不存在")
		}
		return response.NewError(http.StatusInternalServerError, "查询退款详情失败")
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) ApproveOrderRefund(c *gin.Context) error {
	setAdminAuditMeta(c, adminAuditMeta{Action: "admin.refund.approve", ResourceType: "refund", ResourceID: c.Param("order_no"), ResourceName: "订单 " + c.Param("order_no")})
	if err := a.AdminService.ApproveOrderRefund(c.Request.Context(), c.Param("order_no")); err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) RejectOrderRefund(c *gin.Context) error {
	var req types.RejectRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, "参数格式错误")
	}
	setAdminAuditMeta(c, adminAuditMeta{Action: "admin.refund.reject", ResourceType: "refund", ResourceID: c.Param("order_no"), ResourceName: "订单 " + c.Param("order_no"), Remark: req.RejectReason})
	if err := a.AdminService.RejectOrderRefund(c.Request.Context(), c.Param("order_no"), req.RejectReason); err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) GetFinanceSummary(c *gin.Context) error {
	resp, err := a.AdminService.GetFinanceSummary(c.Request.Context())
	if err != nil {
		return response.NewError(500, "查询失败: "+err.Error())
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) ListPlatformFinanceFlows(c *gin.Context) error {
	filter := types.AdminPlatformFlowFilter{
		Type:        c.Query("type"),
		Keyword:     c.Query("keyword"),
		OrganizerID: adminQueryInt64(c, "organizer_id"),
		StartDate:   adminQueryTime(c, "start_date"),
		EndDate:     adminQueryTime(c, "end_date"),
	}
	if (c.Query("start_date") != "" && filter.StartDate == nil) || (c.Query("end_date") != "" && filter.EndDate == nil) {
		return response.NewError(400, "日期格式应为 YYYY-MM-DD")
	}
	resp, err := a.AdminService.ListPlatformFinanceFlows(c.Request.Context(), adminPage(c), adminPageSize(c), filter)
	if err != nil {
		return response.NewError(500, "查询失败: "+err.Error())
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) GetFinanceSettlements(c *gin.Context) error {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	resp, err := a.AdminService.GetFinanceSettlements(c.Request.Context(), page, pageSize, adminQueryInt64(c, "organizer_id"))
	if err != nil {
		return response.NewError(500, "查询失败: "+err.Error())
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) GetFinanceSettlementDetail(c *gin.Context) error {
	organizerID, err := strconv.ParseInt(c.Param("organizer_id"), 10, 64)
	if err != nil {
		return response.NewError(400, "无效的主办方ID")
	}
	resp, err := a.AdminService.GetFinanceSettlements(c.Request.Context(), 1, 1, organizerID)
	if err != nil {
		return response.NewError(500, "查询失败: "+err.Error())
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) ExportFinanceSettlement(c *gin.Context) error {
	organizerID, err := strconv.ParseInt(c.Param("organizer_id"), 10, 64)
	if err != nil {
		return response.NewError(400, "无效的主办方ID")
	}
	resp, err := a.AdminService.GetFinanceSettlements(c.Request.Context(), 1, 1, organizerID)
	if err != nil {
		return response.NewError(500, "导出失败: "+err.Error())
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) GetUserList(c *gin.Context) error {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	resp, err := a.AdminService.GetUserList(c.Request.Context(), page, pageSize, c.Query("keyword"))
	if err != nil {
		return response.NewError(500, "查询失败: "+err.Error())
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) UpdateUserStatus(c *gin.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.NewError(400, "无效的用户ID")
	}
	var req types.AdminUpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, "参数格式错误")
	}
	if err := a.AdminService.UpdateUserStatus(c.Request.Context(), int(id), req.Status); err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) ListBanners(c *gin.Context) error {
	resp, err := a.AdminService.ListBanners(c.Request.Context())
	if err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, gin.H{"list": resp})
	return nil
}

func (a *Admin) CreateBanner(c *gin.Context) error {
	var req types.AdminBannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, "参数格式错误")
	}
	id, err := a.AdminService.CreateBanner(c.Request.Context(), req)
	if err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, gin.H{"success": true, "id": id})
	return nil
}

func (a *Admin) UpdateBanner(c *gin.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.NewError(400, "无效的Banner ID")
	}
	var req types.AdminBannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, "参数格式错误")
	}
	if err := a.AdminService.UpdateBanner(c.Request.Context(), id, req); err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) DeleteBanner(c *gin.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.NewError(400, "无效的Banner ID")
	}
	if err := a.AdminService.DeleteBanner(c.Request.Context(), id); err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) SortBanners(c *gin.Context) error {
	var req types.AdminBannerSortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, "参数格式错误")
	}
	if err := a.AdminService.SortBanners(c.Request.Context(), req); err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) GetSettings(c *gin.Context) error {
	resp, err := a.AdminService.GetSettings(c.Request.Context())
	if err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, gin.H{"settings": resp})
	return nil
}

func (a *Admin) UpdateSettings(c *gin.Context) error {
	var req types.AdminSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, "参数格式错误")
	}
	setAdminAuditMeta(c, adminAuditMeta{Action: "admin.settings.update", ResourceType: "settings", ResourceName: "系统配置", Remark: "更新系统配置"})
	if err := a.AdminService.UpdateSettings(c.Request.Context(), req.Settings); err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) GetSystemConfig(c *gin.Context) error {
	resp, err := a.AdminService.GetSystemConfig(c.Request.Context())
	if err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, resp)
	return nil
}

func (a *Admin) GetPublicSystemConfig(c *gin.Context) error {
	config, err := a.AdminService.GetSystemConfig(c.Request.Context())
	if err != nil {
		return response.NewError(500, "读取公开系统配置失败")
	}
	// The cache is intentionally short so configuration changes become visible promptly.
	c.Header("Cache-Control", "public, max-age=300")
	response.Success(c, types.PublicSystemConfig{SystemName: config.SystemName, ICPRecordNo: config.ICPRecordNo})
	return nil
}

func (a *Admin) UpdateSystemConfig(c *gin.Context) error {
	var req types.AdminSystemConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(400, "参数格式错误")
	}
	setAdminAuditMeta(c, adminAuditMeta{Action: "admin.settings.update", ResourceType: "settings", ResourceName: "系统配置", Remark: "更新系统名称、备案及客服信息"})
	if err := a.AdminService.UpdateSystemConfig(c.Request.Context(), req); err != nil {
		return response.NewError(400, err.Error())
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

// GetDashboard 获取后台数据概览
// GET /api/v1/admin/dashboard
func (a *Admin) GetDashboard(c *gin.Context) error {
	stats, err := a.AdminService.GetDashboardStats(c.Request.Context())
	if err != nil {
		return response.NewError(500, "查询失败: "+err.Error())
	}

	response.Success(c, stats)
	return nil
}

func adminPage(c *gin.Context) int {
	v, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if v <= 0 {
		return 1
	}
	return v
}

func adminPageSize(c *gin.Context) int {
	v, _ := strconv.Atoi(c.DefaultQuery("pageSize", c.DefaultQuery("size", "20")))
	if v <= 0 {
		return 20
	}
	if v > 100 {
		return 100
	}
	return v
}

func adminParamID(c *gin.Context) int64 {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	return id
}

func adminQueryInt64(c *gin.Context, key string) int64 {
	id, _ := strconv.ParseInt(c.Query(key), 10, 64)
	return id
}

func adminQueryTime(c *gin.Context, key string) *time.Time {
	raw := c.Query(key)
	if raw == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
		if value, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return &value
		}
	}
	return nil
}

func adminRefundStatusQuery(raw string) (*int8, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if value, err := strconv.ParseInt(raw, 10, 8); err == nil {
		parsed := int8(value)
		return &parsed, nil
	}
	var value int8
	switch raw {
	case "pending_review":
		value = models.RefundStatusAuditing
	case "refunding":
		value = models.RefundStatusRunning
	case "refunded":
		value = models.RefundStatusSuccess
	case "rejected":
		value = models.RefundStatusRejected
	case "cancelled":
		value = models.RefundStatusCancelled
	default:
		return nil, errors.New("无效的售后状态")
	}
	return &value, nil
}
