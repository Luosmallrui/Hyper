package middleware

import (
	"Hyper/pkg/log"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GinZap() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		cost := time.Since(start).Seconds()

		log.L.Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.Float64("cost", cost),
		)
	}
}
