package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"

	"github.com/gin-gonic/gin"
)

type Session struct {
	SessionService service.ISessionService
	Config         *config.Config
}

func (s *Session) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(s.Config.Jwt.Secret))
	session := r.Group("/v1/session/")
	session.Use(authorize)
	session.GET("list", context.Wrap(s.ListSessions))
	session.POST("setting", context.Wrap(s.SessionSetting))
}
func (s *Session) ListSessions(c *gin.Context) error {
	userId, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(401, "未登录")
	}
	list, err := s.SessionService.ListUserSessions(
		c.Request.Context(),
		uint64(userId),
		50,
	)
	if err != nil {
		return response.NewError(500, "获取会话失败")
	}

	response.Success(c, gin.H{
		"list": list,
	})
	return nil
}
func (s *Session) SessionSetting(c *gin.Context) error {
	userId, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(401, "未登录")
	}

	in := &types.SessionSettingRequest{}
	if err := c.ShouldBindJSON(in); err != nil {
		return response.NewError(400, err.Error())
	}
	if in.SessionType != 1 && in.SessionType != 2 {
		return response.NewError(400, "session_type 必须是 1 或 2")
	}
	if in.PeerID == 0 {
		return response.NewError(400, "peer_id 不能为空")
	}
	if in.IsTop == nil || (*in.IsTop != 0 && *in.IsTop != 1) {
		return response.NewError(400, "is_top 只能是 0 或 1")
	}
	if in.IsMute == nil || (*in.IsMute != 0 && *in.IsMute != 1) {
		return response.NewError(400, "is_mute 只能是 0 或 1")
	}
	if err := s.SessionService.UpdateSessionSettings(c.Request.Context(), uint64(userId), in); err != nil {
		return response.NewError(500, "参数错误")
	}

	response.Success(c, "ok")
	return nil
}
