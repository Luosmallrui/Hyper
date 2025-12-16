package server

import (
	"github.com/gin-gonic/gin"
)

func NewGinEngine(h *Handlers) *gin.Engine {
	r := gin.Default()

	h.User.RegisterRoutes(r)
	// h.Order.RegisterRoutes(r)
	// h.Auth.RegisterRoutes(r)

	return r
}
