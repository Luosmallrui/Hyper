package middleware

import (
	"Hyper/pkg/log"
	"net/http"
	"strings"
	"time"

	"Hyper/pkg/jwt"
	"Hyper/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Auth(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Abort(c, http.StatusUnauthorized, "缺少 Authorization")
			return
		}
		log.L.Info("auth header", zap.String("authHeader", authHeader))

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Abort(c, http.StatusUnauthorized, "Authorization 格式错误")
			return
		}

		claims, err := jwt.ParseToken(secret, "access", parts[1])
		if err != nil {
			response.Abort(c, http.StatusUnauthorized, err.Error())
			return
		}
		if time.Until(claims.ExpiresAt.Time) < 20 {
			newToken, _ := jwt.GenerateToken(
				secret,
				claims.UserID,
				claims.OpenID,
				"access",
				60*time.Second,
			)
			c.Header("X-New-Access-Token", newToken)
		}
		log.L.Info("claims", zap.Any("claims", claims))
		c.Set("user_id", claims.UserID)
		c.Set("openid", claims.OpenID)

		c.Next()
	}
}
