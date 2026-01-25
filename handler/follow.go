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

// GetFollowingList 查询用户已关注的用户列表
func (f *Follow) GetFollowingList(c *gin.Context) error {
	var req types.GetFollowingListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数错误: "+err.Error())
	}
	if req.PageSize == 0 {
		req.PageSize = types.DefaultPageSize
	}
	follows, err := f.FollowService.GetMyFollowingListWithStatus(c.Request.Context(), uint64(c.GetInt("user_id")), req.Cursor, req.PageSize)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	var nextCursor int64 = 0

	response.Success(c, types.GetFollowingFeedResponse{
		Following:  follows,
		NextCursor: nextCursor, // 下次请求带上这个值
		HasMore:    len(follows) == req.PageSize,
	})
	return nil
}
