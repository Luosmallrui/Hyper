package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Follow struct {
	Config        *config.Config
	FollowService service.IFollowService
}

func (f *Follow) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(f.Config.Jwt.Secret))
	g := r.Group("/v1/user")
	g.POST("/:user_id/follow", authorize, context.Wrap(f.FollowUser))
	g.DELETE("/:user_id/follow", authorize, context.Wrap(f.UnfollowUser))
	g.GET("/:user_id/follow", authorize, context.Wrap(f.GetFollowStatus))
	g.GET("/:user_id/followers/count", context.Wrap(f.GetFollowerCount))
	g.GET("/:user_id/following/count", context.Wrap(f.GetFollowingCount))
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
