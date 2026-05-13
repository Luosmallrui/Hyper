package middleware

import (
	"Hyper/pkg/jwt"
	"Hyper/pkg/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AdminAuth 管理后台鉴权中间件
func AdminAuth(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Abort(c, http.StatusUnauthorized, "缺少 Authorization")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Abort(c, http.StatusUnauthorized, "Authorization 格式错误")
			return
		}

		claims, err := jwt.ParseToken(secret, "access", parts[1])
		if err != nil {
			response.Abort(c, http.StatusUnauthorized, "token 无效或已过期")
			return
		}

		if claims.OpenID != "admin" {
			response.Abort(c, http.StatusForbidden, "非管理员账号")
			return
		}

		c.Set("admin_id", int(claims.UserID))
		c.Next()
	}
}
