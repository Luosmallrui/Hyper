package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"

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
	session.GET("", context.Wrap(s.ListSessions))
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
