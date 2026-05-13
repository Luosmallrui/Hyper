package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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
		// 派对/场地管理
		authorized.GET("/parties", context.Wrap(a.GetPartyList))
		authorized.GET("/parties/:id", context.Wrap(a.GetPartyDetail))
		authorized.PUT("/parties/:id/status", context.Wrap(a.UpdatePartyStatus))

		// 票券管理
		authorized.GET("/tickets", context.Wrap(a.GetAllTickets))
		authorized.GET("/events/:id/tickets", context.Wrap(a.GetEventTickets))

		// 订单管理
		authorized.GET("/orders", context.Wrap(a.GetOrderList))

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

	result, err := a.AdminService.GetOrderList(c.Request.Context(), page, pageSize, 0)
	if err != nil {
		return response.NewError(500, "查询失败: "+err.Error())
	}

	response.Success(c, result)
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
