package server

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func NewGinEngine(h *Handlers) *gin.Engine {
	r := gin.Default()
	r.Use(CORSMiddleware())
	h.User.RegisterRoutes(r)
	// h.Order.RegisterRoutes(r)
	// h.Auth.RegisterRoutes(r)

	return r
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置 CORS 头
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // 允许所有来源
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Content-Length, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		// 对于 OPTIONS 请求，直接返回 204
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
