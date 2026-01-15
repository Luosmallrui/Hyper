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
	"time"

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
	g := r.Group("/v1/user")
	g.POST("/:user_id/follow", authorize, context.Wrap(f.FollowUser))
	g.DELETE("/:user_id/follow", authorize, context.Wrap(f.UnfollowUser))
	g.GET("/:user_id/follow", authorize, context.Wrap(f.GetFollowStatus))
	g.GET("/:user_id/followers/count", context.Wrap(f.GetFollowerCount))
	g.GET("/:user_id/following/count", context.Wrap(f.GetFollowingCount))
	g.GET("/:user_id/following/list", authorize, context.Wrap(f.GetFollowingList))
	g.GET("/test")
}

// func (f *Follow) TestMq(c *gin.Context) error {
// 	body := []byte{"test yige"}
// 	mqMsg := &primitive.Message{
// 		Topic: "IM_CHAT_MSGS",
// 		Body:  body,
// 	}
// 	_, err := f.MqProducer.SendSync(c.Request.Context(), mqMsg)
// 	return err

// }

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
	_, err := context.GetUserID(c)
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

	var req types.GetFollowingListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数错误: "+err.Error())
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = types.DefaultPage
	}
	if req.PageSize == 0 {
		req.PageSize = types.DefaultPageSize
	}

	limit := req.PageSize
	offset := (req.Page - 1) * req.PageSize

	follows, total, err := f.FollowService.GetFollowingList(c.Request.Context(), targetUserID, limit, offset)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}

	// 转换为响应格式
	result := make([]*types.FollowingUser, 0, len(follows))
	for _, follow := range follows {
		userID, _ := follow["user_id"].(int64)
		if userID == 0 {
			// 用户被删除，跳过
			continue
		}
		nickname, _ := follow["nickname"].(string)
		avatar, _ := follow["avatar"].(string)
		var updatedAt time.Time
		if t, ok := follow["updated_at"].(time.Time); ok {
			updatedAt = t
		}

		result = append(result, &types.FollowingUser{
			UserID:    userID,
			Nickname:  nickname,
			Avatar:    avatar,
			UpdatedAt: updatedAt,
		})
	}

	response.Success(c, types.GetFollowingListResponse{
		List:  result,
		Total: total,
	})
	return nil
}
