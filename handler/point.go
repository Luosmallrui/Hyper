package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PointHandler struct {
	Config       *config.Config
	PointService service.IPointService
}

func (p *PointHandler) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(p.Config.Jwt.Secret))
	pointGroup := r.Group("/v1/points")
	pointGroup.GET("/records", authorize, context.Wrap(p.GetRecordsItem))
	pointGroup.POST("/reward", context.Wrap(p.RewardPoints))
	pointGroup.POST("/consume", authorize, context.Wrap(p.ConsumePoint))
	pointGroup.GET("/balance", authorize, context.Wrap(p.GetAccountAllPoint))

}

func (p *PointHandler) RewardPoints(c *gin.Context) error {
	var req types.RewardPointsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "请求参数格式错误")
	}

	// 2. 调用 Service 执行业务逻辑
	account, err := p.PointService.RewardPoints(
		c.Request.Context(),
		req.UserID,
		req.Amount,
		req.ChangeType,
		req.SourceID,
		req.Remark,
		req.IsPending,
	)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, account)
	return nil
}
func (p *PointHandler) ConsumePoint(c *gin.Context) error {
	var req types.ConsumePointsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "请求参数格式错误")
	}
	var userID int
	authHeader := c.GetHeader("Authorization")
	if authHeader == "Bearer debug-mode" {
		userID = 6 // Debug 模式下使用固定用户ID
	} else {
		userID = c.GetInt("user_id")
	}

	account, err := p.PointService.ConsumePoints(
		c.Request.Context(),
		uint64(userID),
		req.Amount,
		req.ChangeType,
		req.SourceID,
		req.Remark,
	)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, account)
	return nil
}
func (p *PointHandler) GetAccountAllPoint(c *gin.Context) error {
	var userID int
	authHeader := c.GetHeader("Authorization")
	if authHeader == "Bearer debug-mode" {
		userID = 6 // Debug 模式下使用固定用户ID
	} else {
		userID = c.GetInt("user_id") // 从 Auth 中间件获取
	}

	data, err := p.PointService.GetAccountDashboard(c.Request.Context(), uint64(userID))
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}

	resp := types.PointsAccountResp{
		Balance:       int64(data.Balance),
		TotalEarned:   int64(data.TotalEarned),
		TotalUsed:     int64(data.TotalUsed),
		PendingCount:  data.PendingCount,
		PendingAmount: int64(data.PendingAmount),
	}

	response.Success(c, resp)
	return nil
}
func (p *PointHandler) GetRecordsItem(c *gin.Context) error {
	var req types.ListPointRecordsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "请求参数格式错误")
	}
	// 2. 获取用户 ID
	var userID int
	authHeader := c.GetHeader("Authorization")
	if authHeader == "Bearer debug-mode" {
		userID = 6
	} else {
		userID = c.GetInt("user_id")
	}
	actionStr := "all"
	if req.Action == 1 {
		actionStr = "income"
	} else if req.Action == 2 {
		actionStr = "expense"
	}
	resp, err := p.PointService.ListPointRecords(
		c.Request.Context(),
		uint64(userID),
		actionStr,
		req.Cursor,
		req.Limit)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, resp)
	return nil
}
