package middleware

import (
	"Hyper/pkg/jwt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AdminAuth 管理后台鉴权中间件
func AdminAuth(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			abortAdminAuth(c, http.StatusUnauthorized, "管理员登录已失效", "ADMIN_UNAUTHORIZED")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			abortAdminAuth(c, http.StatusUnauthorized, "管理员登录已失效", "ADMIN_UNAUTHORIZED")
			return
		}

		claims, err := jwt.ParseToken(secret, "access", parts[1])
		if err != nil {
			abortAdminAuth(c, http.StatusUnauthorized, "管理员登录已失效", "ADMIN_UNAUTHORIZED")
			return
		}

		if claims.OpenID != "admin" {
			abortAdminAuth(c, http.StatusForbidden, "非管理员账号", "ADMIN_FORBIDDEN")
			return
		}

		c.Set("admin_id", int(claims.UserID))
		c.Next()
	}
}

func abortAdminAuth(c *gin.Context, status int, message, errorCode string) {
	c.AbortWithStatusJSON(status, gin.H{
		"code": status, "msg": message, "message": message, "error_code": errorCode,
	})
}
