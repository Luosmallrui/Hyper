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
		if claims.ExpiresAt != nil && time.Until(claims.ExpiresAt.Time) < 20*time.Minute {
			newToken, _ := jwt.GenerateToken(
				secret,
				claims.UserID,
				claims.OpenID,
				"access",
				60*time.Second,
			)
			c.Header("X-New-Access-Token", newToken)
		}
		log.L.Debug("claims", zap.Any("claims", claims))
		c.Set("user_id", int(claims.UserID))
		c.Set("openid", claims.OpenID)

		c.Next()
	}
}

// OptionalAuth resolves a valid access token when one is supplied while keeping
// public endpoints available to anonymous callers.
func OptionalAuth(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

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
		if claims.ExpiresAt != nil && time.Until(claims.ExpiresAt.Time) < 20*time.Minute {
			newToken, _ := jwt.GenerateToken(secret, claims.UserID, claims.OpenID, "access", 60*time.Second)
			c.Header("X-New-Access-Token", newToken)
		}
		c.Set("user_id", int(claims.UserID))
		c.Set("openid", claims.OpenID)
		c.Next()
	}
}
