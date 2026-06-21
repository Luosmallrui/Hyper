package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/response"
	"Hyper/pkg/snowflake"
	"Hyper/service"
	"Hyper/types"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Ticketing struct {
	Config           *config.Config
	TicketingService service.ITicketingService
	OssService       service.IOssService
}

func (h *Ticketing) RegisterRouter(r gin.IRouter) {
	auth := middleware.Auth([]byte(h.Config.Jwt.Secret))
	v1 := r.Group("/v1")

	organizer := v1.Group("/organizer", auth)
	{
		organizer.GET("/info", h.wrap(h.GetOrganizerInfo))
		organizer.PUT("/basic", h.wrap(h.UpdateOrganizerBasic))
		organizer.GET("/withdraw-info", h.wrap(h.GetWithdrawInfo))
		organizer.PUT("/withdraw-info", h.wrap(h.UpdateWithdrawInfo))
		organizer.GET("/withdraws", h.wrap(h.ListOrganizerWithdraws))
		organizer.POST("/withdraws", h.wrap(h.CreateOrganizerWithdraw))
		organizer.GET("/profile", h.wrap(h.GetOrganizerProfile))
		organizer.PUT("/profile", h.wrap(h.UpdateOrganizerProfile))
		organizer.GET("/subscription/summary", h.wrap(h.GetOrganizerSubscriptionSummary))
		organizer.GET("/users/lookup", h.wrap(h.LookupOrganizerUser))
		organizer.GET("/collections", h.wrap(h.ListOrganizerCollections))
		organizer.GET("/collections/:id", h.wrap(h.GetOrganizerCollection))
		organizer.POST("/collections", h.wrap(h.CreateOrganizerCollection))
		organizer.PUT("/collections/:id", h.wrap(h.UpdateOrganizerCollection))
		organizer.DELETE("/collections/:id", h.wrap(h.DeleteOrganizerCollection))
		organizer.GET("/messages", h.wrap(h.ListOrganizerMessages))
		organizer.POST("/messages/read-all", h.wrap(h.MarkAllOrganizerMessagesRead))
		organizer.POST("/messages/:id/read", h.wrap(h.MarkOrganizerMessageRead))
		organizer.GET("/finance/summary", h.wrap(h.GetOrganizerFinanceSummary))
		organizer.GET("/finance/flows", h.wrap(h.ListOrganizerFinanceFlows))
		organizer.GET("/level-rules", h.wrap(h.ListOrganizerLevelRules))
		organizer.GET("/roles", h.wrap(h.ListOrganizerRoles))
		organizer.POST("/roles", h.wrap(h.CreateOrganizerRole))
		organizer.PUT("/roles/:id", h.wrap(h.UpdateOrganizerRole))
		organizer.DELETE("/roles/:id", h.wrap(h.DeleteOrganizerRole))
		organizer.GET("/staff", h.wrap(h.ListOrganizerStaff))
		organizer.POST("/staff", h.wrap(h.CreateOrganizerStaff))
		organizer.PUT("/staff/:id", h.wrap(h.UpdateOrganizerStaff))
		organizer.DELETE("/staff/:id", h.wrap(h.DeleteOrganizerStaff))
		organizer.GET("/operation-logs", h.wrap(h.ListOrganizerOperationLogs))
		organizer.GET("/posts", h.wrap(h.ListOrganizerPosts))
		organizer.POST("/posts", h.wrap(h.CreateOrganizerPost))
		organizer.PUT("/posts/:id", h.wrap(h.UpdateOrganizerPost))
		organizer.PATCH("/posts/:id/visibility", h.wrap(h.UpdateOrganizerPostVisibility))
		organizer.DELETE("/posts/:id", h.wrap(h.DeleteOrganizerPost))
		organizer.POST("/apply", h.wrap(h.ApplyOrganizer))
		organizer.GET("/audit-status", h.wrap(h.GetOrganizerAuditStatus))
		organizer.GET("/orders", h.wrap(h.ListOrganizerOrders))
		organizer.GET("/refunds", h.wrap(h.ListOrganizerRefunds))
		organizer.GET("/verifiers", h.wrap(h.ListVerifiers))
		organizer.POST("/verifier", h.wrap(h.AddVerifier))
		organizer.PATCH("/verifier/:id/status", h.wrap(h.UpdateVerifierStatus))
		organizer.DELETE("/verifier/:id", h.wrap(h.DeleteVerifier))
		organizer.GET("/verifier/:id/activation-qr", h.wrap(h.GetVerifierActivationQR))
		organizer.GET("/stores", h.wrap(h.ListStores))
		organizer.POST("/stores", h.wrap(h.CreateStore))
		organizer.PUT("/stores/:id", h.wrap(h.UpdateStore))
		organizer.DELETE("/stores/:id", h.wrap(h.DeleteStore))
	}

	activity := v1.Group("/activity", auth)
	{
		activity.POST("/create", h.wrap(h.SaveActivityStep))
		activity.GET("/my-list", h.wrap(h.GetMyActivities))
		activity.GET("/search", h.wrap(h.SearchActivities))
		activity.GET("/subscriptions", h.wrap(h.ListSubscribedActivities))
		activity.GET("/:id/statistics", h.wrap(h.GetActivityStatistics))
		activity.GET("/:id/statistics/daily", h.wrap(h.GetActivityDailyStatistics))
		activity.GET("/:id", h.wrap(h.GetActivity))
		activity.POST("/:id/subscribe", h.wrap(h.SubscribeActivity))
		activity.POST("/:id/unsubscribe", h.wrap(h.UnsubscribeActivity))
		activity.DELETE("/:id", h.wrap(h.DeleteActivity))
		activity.POST("/:id/submit-audit", h.wrap(h.SubmitActivityAudit))
		activity.GET("/:id/ticket-specs", h.wrap(h.GetTicketSpecs))
		activity.POST("/:id/ticket-specs", h.wrap(h.SaveTicketSpecs))
	}

	v1.DELETE("/ticket-spec/:id", auth, h.wrap(h.DeleteTicketSpec))

	order := v1.Group("/order", auth)
	{
		order.POST("/create", h.wrap(h.CreateTicketOrder))
		order.GET("/cancel-reasons", h.wrap(h.ListCancelReasons))
		order.GET("/:order_no", h.wrap(h.GetTicketOrderDetail))
		order.POST("/:order_no/cancel", h.wrap(h.CancelTicketOrder))
		order.DELETE("/:order_no", h.wrap(h.DeleteTicketOrder))
	}

	refund := v1.Group("/refund", auth)
	{
		refund.GET("/reasons", h.wrap(h.ListRefundReasons))
		refund.GET("/list", h.wrap(h.ListUserRefunds))
		refund.POST("/apply", h.wrap(h.ApplyRefund))
		refund.GET("/:refund_no", h.wrap(h.GetRefundDetail))
		refund.POST("/:refund_no/cancel", h.wrap(h.CancelRefund))
		refund.POST("/:refund_no/reject", h.wrap(h.RejectRefund))
	}

	verifier := v1.Group("/verifier")
	{
		verifier.GET("/activation-info", h.wrap(h.GetVerifierActivationInfo))
		verifier.POST("/activate", auth, h.wrap(h.ActivateVerifier))
		verifier.POST("/scan", h.wrap(h.ScanOrder))
		verifier.POST("/confirm", h.wrap(h.ConfirmVerify))
		verifier.GET("/verified-list", h.wrap(h.ListVerified))
	}

	viewers := v1.Group("/viewers", auth)
	{
		viewers.GET("", h.wrap(h.ListViewers))
		viewers.POST("", h.wrap(h.CreateViewer))
		viewers.PUT("/:id", h.wrap(h.UpdateViewer))
		viewers.DELETE("/:id", h.wrap(h.DeleteViewer))
	}

	v1.POST("/upload", auth, h.wrap(h.UploadFile))
}

func (h *Ticketing) wrap(fn func(*gin.Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := fn(c); err != nil {
			response.Abort(c, http.StatusInternalServerError, err.Error())
			return
		}
	}
}

func (h *Ticketing) GetOrganizerInfo(c *gin.Context) error {
	resp, err := h.TicketingService.GetOrganizerInfo(c.Request.Context(), currentUserID(c))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) ApplyOrganizer(c *gin.Context) error {
	var req types.OrganizerApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	resp, err := h.TicketingService.ApplyOrganizer(c.Request.Context(), currentUserID(c), req)
	if err != nil {
		if isOrganizerApplyConflict(err) {
			response.Abort(c, http.StatusConflict, err.Error())
			return nil
		}
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) UpdateOrganizerBasic(c *gin.Context) error {
	var req types.OrganizerBasicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	if err := h.TicketingService.UpdateOrganizerBasic(c.Request.Context(), currentUserID(c), req); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) GetWithdrawInfo(c *gin.Context) error {
	resp, err := h.TicketingService.GetWithdrawInfo(c.Request.Context(), currentUserID(c))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) UpdateWithdrawInfo(c *gin.Context) error {
	var req types.OrganizerWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	if err := h.TicketingService.UpdateWithdrawInfo(c.Request.Context(), currentUserID(c), req); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) ListOrganizerWithdraws(c *gin.Context) error {
	var status *int8
	if raw := c.Query("status"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 8)
		if err != nil {
			return err
		}
		v := int8(value)
		status = &v
	}
	resp, err := h.TicketingService.ListOrganizerWithdraws(c.Request.Context(), currentUserID(c), status, queryInt(c, "page"), queryInt(c, "size"))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) CreateOrganizerWithdraw(c *gin.Context) error {
	var req types.CreateOrganizerWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	id, err := h.TicketingService.CreateOrganizerWithdraw(c.Request.Context(), currentUserID(c), req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": id})
	return nil
}

func (h *Ticketing) ListOrganizerCollections(c *gin.Context) error {
	resp, err := h.TicketingService.ListOrganizerCollections(c.Request.Context(), currentUserID(c), page(c), size(c), c.Query("keyword"))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) GetOrganizerCollection(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	resp, err := h.TicketingService.GetOrganizerCollection(c.Request.Context(), currentUserID(c), id)
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) CreateOrganizerCollection(c *gin.Context) error {
	var req types.OrganizerCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	id, err := h.TicketingService.SaveOrganizerCollection(c.Request.Context(), currentUserID(c), 0, req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": id})
	return nil
}

func (h *Ticketing) UpdateOrganizerCollection(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	var req types.OrganizerCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	savedID, err := h.TicketingService.SaveOrganizerCollection(c.Request.Context(), currentUserID(c), id, req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": savedID})
	return nil
}

func (h *Ticketing) DeleteOrganizerCollection(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	if err := h.TicketingService.DeleteOrganizerCollection(c.Request.Context(), currentUserID(c), id); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) ListOrganizerMessages(c *gin.Context) error {
	unreadOnly := c.Query("unread_only") == "1" || c.Query("unread_only") == "true"
	resp, err := h.TicketingService.ListOrganizerMessages(c.Request.Context(), currentUserID(c), page(c), size(c), unreadOnly)
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) MarkOrganizerMessageRead(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	if err := h.TicketingService.MarkOrganizerMessageRead(c.Request.Context(), currentUserID(c), id); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) MarkAllOrganizerMessagesRead(c *gin.Context) error {
	resp, err := h.TicketingService.MarkAllOrganizerMessagesRead(c.Request.Context(), currentUserID(c))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) GetOrganizerSubscriptionSummary(c *gin.Context) error {
	resp, err := h.TicketingService.GetOrganizerSubscriptionSummary(c.Request.Context(), currentUserID(c))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) GetOrganizerProfile(c *gin.Context) error {
	resp, err := h.TicketingService.GetOrganizerProfile(c.Request.Context(), currentUserID(c))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) UpdateOrganizerProfile(c *gin.Context) error {
	var req types.OrganizerProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	if err := h.TicketingService.UpdateOrganizerProfile(c.Request.Context(), currentUserID(c), req); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) LookupOrganizerUser(c *gin.Context) error {
	resp, err := h.TicketingService.LookupOrganizerUser(c.Request.Context(), currentUserID(c), c.Query("phone"))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) GetOrganizerFinanceSummary(c *gin.Context) error {
	resp, err := h.TicketingService.GetOrganizerFinanceSummary(c.Request.Context(), currentUserID(c))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) ListOrganizerFinanceFlows(c *gin.Context) error {
	resp, err := h.TicketingService.ListOrganizerFinanceFlows(c.Request.Context(), currentUserID(c), page(c), size(c), c.Query("type"))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) ListOrganizerLevelRules(c *gin.Context) error {
	resp, err := h.TicketingService.ListOrganizerLevelRules(c.Request.Context(), currentUserID(c))
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"list": resp})
	return nil
}

func (h *Ticketing) CreateOrganizerLevelRule(c *gin.Context) error {
	var req types.OrganizerLevelRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	id, err := h.TicketingService.SaveOrganizerLevelRule(c.Request.Context(), currentUserID(c), 0, req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": id})
	return nil
}

func (h *Ticketing) UpdateOrganizerLevelRule(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	var req types.OrganizerLevelRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	savedID, err := h.TicketingService.SaveOrganizerLevelRule(c.Request.Context(), currentUserID(c), id, req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": savedID})
	return nil
}

func (h *Ticketing) DeleteOrganizerLevelRule(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	if err := h.TicketingService.DeleteOrganizerLevelRule(c.Request.Context(), currentUserID(c), id); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) ListOrganizerRoles(c *gin.Context) error {
	resp, err := h.TicketingService.ListOrganizerRoles(c.Request.Context(), currentUserID(c), page(c), size(c))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) CreateOrganizerRole(c *gin.Context) error {
	var req types.OrganizerRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	id, err := h.TicketingService.SaveOrganizerRole(c.Request.Context(), currentUserID(c), 0, req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": id})
	return nil
}

func (h *Ticketing) UpdateOrganizerRole(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	var req types.OrganizerRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	savedID, err := h.TicketingService.SaveOrganizerRole(c.Request.Context(), currentUserID(c), id, req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": savedID})
	return nil
}

func (h *Ticketing) DeleteOrganizerRole(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	if err := h.TicketingService.DeleteOrganizerRole(c.Request.Context(), currentUserID(c), id); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) ListOrganizerStaff(c *gin.Context) error {
	resp, err := h.TicketingService.ListOrganizerStaff(c.Request.Context(), currentUserID(c), page(c), size(c))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) CreateOrganizerStaff(c *gin.Context) error {
	var req types.OrganizerStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	id, err := h.TicketingService.SaveOrganizerStaff(c.Request.Context(), currentUserID(c), 0, req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": id})
	return nil
}

func (h *Ticketing) UpdateOrganizerStaff(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	var req types.OrganizerStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	savedID, err := h.TicketingService.SaveOrganizerStaff(c.Request.Context(), currentUserID(c), id, req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": savedID})
	return nil
}

func (h *Ticketing) DeleteOrganizerStaff(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	if err := h.TicketingService.DeleteOrganizerStaff(c.Request.Context(), currentUserID(c), id); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) ListOrganizerOperationLogs(c *gin.Context) error {
	resp, err := h.TicketingService.ListOrganizerOperationLogs(c.Request.Context(), currentUserID(c), page(c), size(c), c.Query("keyword"))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) ListOrganizerPosts(c *gin.Context) error {
	var status *int
	if raw := c.Query("status"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil {
			return err
		}
		status = &value
	}
	activityID := 0
	if raw := c.Query("activity_id"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil {
			return err
		}
		activityID = value
	}
	var storeID int64
	if raw := c.Query("store_id"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return err
		}
		storeID = value
	}
	resp, err := h.TicketingService.ListOrganizerPosts(c.Request.Context(), currentUserID(c), page(c), size(c), c.Query("keyword"), status, activityID, storeID)
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) CreateOrganizerPost(c *gin.Context) error {
	var req types.OrganizerPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	id, err := h.TicketingService.SaveOrganizerPost(c.Request.Context(), currentUserID(c), 0, req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": id})
	return nil
}

func (h *Ticketing) UpdateOrganizerPost(c *gin.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return err
	}
	var req types.OrganizerPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	savedID, err := h.TicketingService.SaveOrganizerPost(c.Request.Context(), currentUserID(c), id, req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"id": savedID})
	return nil
}

func (h *Ticketing) UpdateOrganizerPostVisibility(c *gin.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return err
	}
	var req types.OrganizerPostVisibilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	if err := h.TicketingService.UpdateOrganizerPostVisibility(c.Request.Context(), currentUserID(c), id, req); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) DeleteOrganizerPost(c *gin.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return err
	}
	if err := h.TicketingService.DeleteOrganizerPost(c.Request.Context(), currentUserID(c), id); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) GetOrganizerAuditStatus(c *gin.Context) error {
	resp, err := h.TicketingService.GetOrganizerAuditStatus(c.Request.Context(), currentUserID(c))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) SaveActivityStep(c *gin.Context) error {
	var req types.ActivityCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	id, err := h.TicketingService.SaveActivityStep(c.Request.Context(), currentUserID(c), req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"activity_id": id})
	return nil
}

func (h *Ticketing) GetActivity(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	resp, err := h.TicketingService.GetActivity(c.Request.Context(), currentUserID(c), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewError(http.StatusNotFound, "活动不存在")
		}
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) SubscribeActivity(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	if err := h.TicketingService.SubscribeActivity(c.Request.Context(), currentUserID(c), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewError(http.StatusNotFound, "活动不存在")
		}
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) UnsubscribeActivity(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	if err := h.TicketingService.UnsubscribeActivity(c.Request.Context(), currentUserID(c), id); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) ListSubscribedActivities(c *gin.Context) error {
	resp, err := h.TicketingService.ListSubscribedActivities(c.Request.Context(), currentUserID(c), page(c), size(c))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) GetMyActivities(c *gin.Context) error {
	resp, err := h.TicketingService.GetMyActivities(c.Request.Context(), currentUserID(c), page(c), size(c))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) SearchActivities(c *gin.Context) error {
	list, err := h.TicketingService.SearchActivities(c.Request.Context(), c.Query("keyword"))
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"list": list})
	return nil
}

func (h *Ticketing) DeleteActivity(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	if err := h.TicketingService.DeleteActivity(c.Request.Context(), currentUserID(c), id); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) SubmitActivityAudit(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	if err := h.TicketingService.SubmitActivityAudit(c.Request.Context(), currentUserID(c), id); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) GetActivityStatistics(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	resp, err := h.TicketingService.GetActivityStatistics(c.Request.Context(), currentUserID(c), id)
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) GetActivityDailyStatistics(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	resp, err := h.TicketingService.GetActivityDailyStatistics(c.Request.Context(), currentUserID(c), id, days)
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) GetTicketSpecs(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	list, err := h.TicketingService.GetTicketSpecs(c.Request.Context(), id)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"list": list})
	return nil
}

func (h *Ticketing) SaveTicketSpecs(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	var req types.SaveTicketSpecsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	if err := h.TicketingService.SaveTicketSpecs(c.Request.Context(), currentUserID(c), id, req.Specs); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) DeleteTicketSpec(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	if err := h.TicketingService.DeleteTicketSpec(c.Request.Context(), currentUserID(c), id); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) CreateTicketOrder(c *gin.Context) error {
	var req types.CreateTicketOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	resp, err := h.TicketingService.CreateTicketOrder(c.Request.Context(), currentUserID(c), req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{
		"success":         true,
		"order_no":        resp.OrderNo,
		"status":          resp.Status,
		"total_price":     resp.TotalPrice,
		"points_amount":   resp.PointsAmount,
		"points_discount": resp.PointsDiscount,
		"actual_price":    resp.ActualPrice,
	})
	return nil
}

func (h *Ticketing) GetTicketOrderDetail(c *gin.Context) error {
	resp, err := h.TicketingService.GetTicketOrderDetail(c.Request.Context(), currentUserID(c), c.Param("order_no"))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) ListCancelReasons(c *gin.Context) error {
	list, err := h.TicketingService.ListCancelReasons(c.Request.Context())
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"list": list})
	return nil
}

func (h *Ticketing) CancelTicketOrder(c *gin.Context) error {
	var req types.CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	if err := h.TicketingService.CancelTicketOrder(c.Request.Context(), currentUserID(c), c.Param("order_no"), req.ReasonID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewError(http.StatusNotFound, "订单不存在、已删除或不属于当前用户")
		}
		if err.Error() == "当前订单不可取消" {
			return response.NewError(http.StatusBadRequest, err.Error())
		}
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) DeleteTicketOrder(c *gin.Context) error {
	if err := h.TicketingService.DeleteTicketOrder(c.Request.Context(), currentUserID(c), c.Param("order_no")); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewError(http.StatusNotFound, "订单不存在、已删除或不属于当前用户")
		}
		if err.Error() == "当前订单不可删除" {
			return response.NewError(http.StatusBadRequest, err.Error())
		}
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) ListStores(c *gin.Context) error {
	resp, err := h.TicketingService.ListStores(c.Request.Context(), currentUserID(c), c.Query("keyword"), page(c), size(c))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) CreateStore(c *gin.Context) error {
	var req types.StoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	id, err := h.TicketingService.CreateStore(c.Request.Context(), currentUserID(c), req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true, "id": id})
	return nil
}

func (h *Ticketing) UpdateStore(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	var req types.StoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	if err := h.TicketingService.UpdateStore(c.Request.Context(), currentUserID(c), id, req); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) DeleteStore(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	if err := h.TicketingService.DeleteStore(c.Request.Context(), currentUserID(c), id); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) ListViewers(c *gin.Context) error {
	resp, err := h.TicketingService.ListViewers(c.Request.Context(), currentUserID(c))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) CreateViewer(c *gin.Context) error {
	var req types.CreateViewerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	id, err := h.TicketingService.CreateViewer(c.Request.Context(), currentUserID(c), req)
	if err != nil {
		if isViewerConflict(err) {
			response.Abort(c, http.StatusConflict, err.Error())
			return nil
		}
		return err
	}
	response.Success(c, gin.H{"success": true, "id": id})
	return nil
}

func (h *Ticketing) UpdateViewer(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	var req types.UpdateViewerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	if err := h.TicketingService.UpdateViewer(c.Request.Context(), currentUserID(c), id, req); err != nil {
		if isViewerConflict(err) {
			response.Abort(c, http.StatusConflict, err.Error())
			return nil
		}
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) DeleteViewer(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	if err := h.TicketingService.DeleteViewer(c.Request.Context(), currentUserID(c), id); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) ListRefundReasons(c *gin.Context) error {
	list, err := h.TicketingService.ListRefundReasons(c.Request.Context())
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"list": list})
	return nil
}

func (h *Ticketing) ApplyRefund(c *gin.Context) error {
	var req types.ApplyRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	refundNo, err := h.TicketingService.ApplyRefund(c.Request.Context(), currentUserID(c), req)
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true, "refund_no": refundNo})
	return nil
}

func (h *Ticketing) ListUserRefunds(c *gin.Context) error {
	var status *int8
	if raw := c.Query("status"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 8)
		if err != nil {
			return err
		}
		v := int8(value)
		status = &v
	}
	resp, err := h.TicketingService.ListUserRefunds(c.Request.Context(), currentUserID(c), status, page(c), size(c))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) ListOrganizerOrders(c *gin.Context) error {
	var status *int8
	if raw := c.Query("status"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 8)
		if err != nil {
			return err
		}
		v := int8(value)
		status = &v
	}
	var activityID int64
	if raw := c.Query("activity_id"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return err
		}
		activityID = value
	}
	resp, err := h.TicketingService.ListOrganizerOrders(
		c.Request.Context(),
		currentUserID(c),
		activityID,
		status,
		c.Query("keyword"),
		c.Query("start_date"),
		c.Query("end_date"),
		page(c),
		size(c),
	)
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) ListOrganizerRefunds(c *gin.Context) error {
	var status *int8
	if raw := c.Query("status"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 8)
		if err != nil {
			return err
		}
		v := int8(value)
		status = &v
	}
	resp, err := h.TicketingService.ListOrganizerRefunds(c.Request.Context(), currentUserID(c), status, page(c), size(c))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) GetRefundDetail(c *gin.Context) error {
	resp, err := h.TicketingService.GetRefundDetail(c.Request.Context(), currentUserID(c), c.Param("refund_no"))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) RejectRefund(c *gin.Context) error {
	var req types.RejectRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	if err := h.TicketingService.RejectRefund(c.Request.Context(), currentUserID(c), c.Param("refund_no"), req); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) CancelRefund(c *gin.Context) error {
	if err := h.TicketingService.CancelRefund(c.Request.Context(), currentUserID(c), c.Param("refund_no")); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) ListVerifiers(c *gin.Context) error {
	resp, err := h.TicketingService.ListVerifiers(c.Request.Context(), currentUserID(c), page(c), size(c))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) AddVerifier(c *gin.Context) error {
	var req types.VerifierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	if err := h.TicketingService.AddVerifier(c.Request.Context(), currentUserID(c), req); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) UpdateVerifierStatus(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	var req struct {
		Status int8 `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	if err := h.TicketingService.UpdateVerifierStatus(c.Request.Context(), currentUserID(c), id, req.Status); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) DeleteVerifier(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	if err := h.TicketingService.DeleteVerifier(c.Request.Context(), currentUserID(c), id); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) GetVerifierActivationQR(c *gin.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}
	resp, err := h.TicketingService.GetVerifierActivationQR(c.Request.Context(), currentUserID(c), id)
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) GetVerifierActivationInfo(c *gin.Context) error {
	raw := c.Query("verifier_id")
	if raw == "" {
		raw = c.Query("v")
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return errors.New("verifier_id 参数错误")
	}
	resp, err := h.TicketingService.GetVerifierActivationInfo(c.Request.Context(), id)
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) ActivateVerifier(c *gin.Context) error {
	var req types.ActivateVerifierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	resp, err := h.TicketingService.ActivateVerifier(c.Request.Context(), currentUserID(c), req)
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) ScanOrder(c *gin.Context) error {
	var req types.ScanOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	resp, err := h.TicketingService.ScanOrder(c.Request.Context(), req)
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) ConfirmVerify(c *gin.Context) error {
	var req types.ConfirmVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	if err := h.TicketingService.ConfirmVerify(c.Request.Context(), verifierID(c), req); err != nil {
		return err
	}
	response.Success(c, gin.H{"success": true})
	return nil
}

func (h *Ticketing) ListVerified(c *gin.Context) error {
	resp, err := h.TicketingService.ListVerified(c.Request.Context(), verifierID(c), page(c), size(c))
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (h *Ticketing) UploadFile(c *gin.Context) error {
	header, err := c.FormFile("file")
	if err != nil {
		return err
	}
	fileType := c.PostForm("type")
	if fileType == "" {
		fileType = "misc"
	}
	key, err := buildUploadKey(fileType, header)
	if err != nil {
		return err
	}
	file, err := header.Open()
	if err != nil {
		return err
	}
	defer file.Close()
	if err := h.OssService.UploadRaw(c.Request.Context(), io.LimitReader(file, 20<<20), key); err != nil {
		return err
	}
	response.Success(c, gin.H{"url": "https://cdn.hypercn.cn/" + key})
	return nil
}

func currentUserID(c *gin.Context) int64 {
	return int64(c.GetInt("user_id"))
}

func verifierID(c *gin.Context) int64 {
	id, _ := strconv.ParseInt(c.GetHeader("X-Verifier-Id"), 10, 64)
	return id
}

func isViewerConflict(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "已存在") || strings.Contains(msg, "已被") || strings.Contains(msg, "Duplicate entry")
}

func isOrganizerApplyConflict(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "正在审核中") || strings.Contains(msg, "已通过")
}

func page(c *gin.Context) int {
	v, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	return v
}

func size(c *gin.Context) int {
	v, _ := strconv.Atoi(c.DefaultQuery("size", c.DefaultQuery("pageSize", "10")))
	return v
}

func parseID(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

func buildUploadKey(fileType string, header *multipart.FileHeader) (string, error) {
	allowed := map[string]bool{
		"poster_detail":     true,
		"poster_long":       true,
		"poster_list":       true,
		"poster_wechat":     true,
		"qualification_doc": true,
		"avatar":            true,
		"organizer_logo":    true,
		"misc":              true,
	}
	if !allowed[fileType] {
		return "", fmt.Errorf("unsupported upload type")
	}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = ".bin"
	}
	return fmt.Sprintf("ticketing/%s/%s/%d%s", fileType, time.Now().Format("2006/01/02"), snowflake.GenID(), ext), nil
}
