package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"errors"
	"strconv"
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

	g := r.Group("/v1/admin")
	g.POST("/login", context.Wrap(a.Login))

	authorized := g.Group("")
	authorized.Use(adminAuth)
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
		authorized.DELETE("/organizers/:id", context.Wrap(a.DeleteOrganizer))
		authorized.POST("/wechat-subscribe", context.Wrap(a.BindWechatSubscribe))

		// 派对/场地管理
		authorized.GET("/parties", context.Wrap(a.GetPartyList))
		authorized.GET("/parties/:id", context.Wrap(a.GetPartyDetail))
		authorized.PUT("/parties/:id/status", context.Wrap(a.UpdatePartyStatus))

		// 活动审核管理
		authorized.GET("/activities", context.Wrap(a.GetActivityList))
		authorized.GET("/activities/:id", context.Wrap(a.GetActivityDetail))
		authorized.PUT("/activities/:id/audit", context.Wrap(a.AuditActivity))
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

		// 财务结算
		authorized.GET("/finance/summary", context.Wrap(a.GetFinanceSummary))
		authorized.GET("/finance/settlements", context.Wrap(a.GetFinanceSettlements))
		authorized.GET("/finance/settlements/:organizer_id", context.Wrap(a.GetFinanceSettlementDetail))
		authorized.GET("/finance/settlements/:organizer_id/export", context.Wrap(a.ExportFinanceSettlement))

		// 用户管理
		authorized.GET("/users", context.Wrap(a.GetUserList))
		authorized.PUT("/users/:id/status", context.Wrap(a.UpdateUserStatus))
		authorized.GET("/users/:id/records/:type", context.Wrap(a.ListUserRecords))
		authorized.GET("/viewers", context.Wrap(a.ListViewers))
		authorized.GET("/verifiers", context.Wrap(a.ListVerifiers))
		authorized.GET("/verification-records", context.Wrap(a.ListVerificationRecords))
		authorized.GET("/notes", context.Wrap(a.ListNotes))
		authorized.PUT("/notes/:id/status", context.Wrap(a.UpdateNoteStatus))
		authorized.GET("/notes/:id/records/:type", context.Wrap(a.ListNoteRecords))
		authorized.GET("/points/logs", context.Wrap(a.ListPointLogs))
		authorized.GET("/withdraws", context.Wrap(a.ListWithdraws))
		authorized.PUT("/withdraws/:id/audit", context.Wrap(a.AuditWithdraw))
		authorized.GET("/messages", context.Wrap(a.ListMessages))
		authorized.POST("/messages", context.Wrap(a.CreateMessage))

		// 内容管理
		authorized.GET("/banners", context.Wrap(a.ListBanners))
		authorized.POST("/banners", context.Wrap(a.CreateBanner))
		authorized.PUT("/banners/sort", context.Wrap(a.SortBanners))
		authorized.PUT("/banners/:id", context.Wrap(a.UpdateBanner))
		authorized.DELETE("/banners/:id", context.Wrap(a.DeleteBanner))

		// 平台设置
		authorized.GET("/settings", context.Wrap(a.GetSettings))
		authorized.PUT("/settings", context.Wrap(a.UpdateSettings))

		// 数据概览
		authorized.GET("/dashboard", context.Wrap(a.GetDashboard))
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
	if err := a.AdminService.UpdateAdmin(c.Request.Context(), adminParamID(c), req); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) DeleteAdmin(c *gin.Context) error {
	if err := a.AdminService.DeleteAdmin(c.Request.Context(), adminParamID(c)); err != nil {
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
	resp, err := a.AdminService.ListOperationLogs(c.Request.Context(), adminPage(c), adminPageSize(c), adminQueryInt64(c, "admin_id"), c.Query("keyword"))
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
	var req types.AdminSimpleStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
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

func (a *Admin) ListPointLogs(c *gin.Context) error {
	resp, err := a.AdminService.ListPointLogs(c.Request.Context(), adminPage(c), adminPageSize(c), adminQueryInt64(c, "user_id"))
	if err != nil {
		return err
	}
	response.Success(c, resp)
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
	if err := a.AdminService.AuditWithdraw(c.Request.Context(), adminParamID(c), req); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (a *Admin) ListMessages(c *gin.Context) error {
	resp, err := a.AdminService.ListMessages(c.Request.Context(), adminPage(c), adminPageSize(c))
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
	id, err := a.AdminService.CreateMessage(c.Request.Context(), req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": id})
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
	if err := a.AdminService.AuditOrganizer(c.Request.Context(), id, req); err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, "审核完成")
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

	result, err := a.AdminService.GetActivityList(c.Request.Context(), page, pageSize, status, keyword, organizerID)
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
	if err := a.AdminService.AuditActivity(c.Request.Context(), id, req); err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, "审核完成")
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

	result, err := a.AdminService.GetTicketOrderList(c.Request.Context(), page, pageSize, activityID, status, keyword)
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

func (a *Admin) ApproveOrderRefund(c *gin.Context) error {
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

func (a *Admin) GetFinanceSettlements(c *gin.Context) error {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	resp, err := a.AdminService.GetFinanceSettlements(c.Request.Context(), page, pageSize, 0)
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
	if err := a.AdminService.UpdateSettings(c.Request.Context(), req.Settings); err != nil {
		return response.NewError(500, err.Error())
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
