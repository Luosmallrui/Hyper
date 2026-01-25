package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"fmt"
	"net/http"

	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/gin-gonic/gin"
)

type Follow struct {
	Config        *config.Config
	FollowService service.IFollowService
	MqProducer    rmq_client.Producer
}

func (f *Follow) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(f.Config.Jwt.Secret))
	g := r.Group("/v1/follow")
	g.POST("/:user_id/follow", authorize, context.Wrap(f.FollowUser))
	g.DELETE("/:user_id/follow", authorize, context.Wrap(f.UnfollowUser))
	g.GET("/:user_id/follow", authorize, context.Wrap(f.GetFollowStatus))
	g.GET("/:user_id/followers/count", context.Wrap(f.GetFollowerCount))
	g.GET("/:user_id/following/count", context.Wrap(f.GetFollowingCount))
	g.GET("/list", authorize, context.Wrap(f.GetFollowingList))
}

// FollowUser 关注用户
func (f *Follow) FollowUser(c *gin.Context) error {
	userID, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}

	targetUserIDParam := c.Param("user_id")
	if targetUserIDParam == "" {
		return response.NewError(http.StatusBadRequest, "缺少 user_id")
	}
	var targetUserID uint64
	_, err = fmt.Sscanf(targetUserIDParam, "%d", &targetUserID)
	if err != nil {
		return response.NewError(http.StatusBadRequest, "user_id 格式错误")
	}

	err = f.FollowService.Follow(c.Request.Context(), uint64(userID), targetUserID)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}

	response.Success(c, gin.H{"followed": true})
	return nil
}

// UnfollowUser 取消关注用户
func (f *Follow) UnfollowUser(c *gin.Context) error {
	userID, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}

	targetUserIDParam := c.Param("user_id")
	if targetUserIDParam == "" {
		return response.NewError(http.StatusBadRequest, "缺少 user_id")
	}
	var targetUserID uint64
	_, err = fmt.Sscanf(targetUserIDParam, "%d", &targetUserID)
	if err != nil {
		return response.NewError(http.StatusBadRequest, "user_id 格式错误")
	}

	err = f.FollowService.Unfollow(c.Request.Context(), uint64(userID), targetUserID)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}

	response.Success(c, gin.H{"followed": false})
	return nil
}

// GetFollowStatus 查询是否已关注
func (f *Follow) GetFollowStatus(c *gin.Context) error {
	userID, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}

	targetUserIDParam := c.Param("user_id")
	if targetUserIDParam == "" {
		return response.NewError(http.StatusBadRequest, "缺少 user_id")
	}
	var targetUserID uint64
	_, err = fmt.Sscanf(targetUserIDParam, "%d", &targetUserID)
	if err != nil {
		return response.NewError(http.StatusBadRequest, "user_id 格式错误")
	}

	isFollowing, err := f.FollowService.IsFollowing(c.Request.Context(), uint64(userID), targetUserID)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}

	response.Success(c, gin.H{"is_following": isFollowing})
	return nil
}

// GetFollowerCount 查询粉丝数
func (f *Follow) GetFollowerCount(c *gin.Context) error {
	targetUserIDParam := c.Param("user_id")
	if targetUserIDParam == "" {
		return response.NewError(http.StatusBadRequest, "缺少 user_id")
	}
	var targetUserID uint64
	_, err := fmt.Sscanf(targetUserIDParam, "%d", &targetUserID)
	if err != nil {
		return response.NewError(http.StatusBadRequest, "user_id 格式错误")
	}

	count, err := f.FollowService.GetFollowerCount(c.Request.Context(), targetUserID)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}

	response.Success(c, gin.H{"follower_count": count})
	return nil
}

// GetFollowingCount 查询关注数
func (f *Follow) GetFollowingCount(c *gin.Context) error {
	targetUserIDParam := c.Param("user_id")
	if targetUserIDParam == "" {
		return response.NewError(http.StatusBadRequest, "缺少 user_id")
	}
	var targetUserID uint64
	_, err := fmt.Sscanf(targetUserIDParam, "%d", &targetUserID)
	if err != nil {
		return response.NewError(http.StatusBadRequest, "user_id 格式错误")
	}

	count, err := f.FollowService.GetFollowingCount(c.Request.Context(), targetUserID)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}

	response.Success(c, gin.H{"following_count": count})
	return nil
}

// GetFollowingList 统一的关注/粉丝列表接口
func (f *Follow) GetFollowingList(c *gin.Context) error {
	var req types.GetFollowingListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数错误: "+err.Error())
	}

	if req.Type != "following" && req.Type != "follower" {
		return response.NewError(http.StatusBadRequest, "type 参数必须是 'following' 或 'follower'")
	}

	// 设置默认分页大小
	if req.PageSize == 0 {
		req.PageSize = types.DefaultPageSize
	}

	// ✅ 调用统一的 Service 方法
	myID := uint64(c.GetInt("user_id"))
	list, nextCursor, hasMore, err := f.FollowService.GetFollowList(
		c.Request.Context(),
		myID,
		req.Type, // ✅ 传递 type 参数
		req.Cursor,
		req.PageSize,
	)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}

	response.Success(c, types.GetFollowingFeedResponse{
		Following:  list,
		NextCursor: nextCursor, // ✅ 正确的 next_cursor
		HasMore:    hasMore,    // ✅ 正确的 has_more
	})
	return nil
}
