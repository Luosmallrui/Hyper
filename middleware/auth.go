package middleware

import (
	"net/http"
	"strings"

	"Hyper/pkg/jwt"
	"Hyper/pkg/response"
	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
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

		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			response.Abort(c, http.StatusUnauthorized, "token 无效")
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("openid", claims.OpenID)

		c.Next()
	}
}
