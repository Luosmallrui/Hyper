package context

import "github.com/gin-gonic/gin"

type HandlerFunc func(*gin.Context) error

func Wrap(h HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := h(c); err != nil {
			c.Error(err) // 交给 ErrorMiddleware
		}
	}
}
