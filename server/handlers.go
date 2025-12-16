package server

import (
	"Hyper/internal/module/user"
	// "Hyper/internal/module/order"
	// "Hyper/internal/module/auth"
)

type Handlers struct {
	User *user.Handler
	// Order *order.Handler
	// Auth  *auth.Handler
}

func NewHandlers(
	userHandler *user.Handler,
	// orderHandler *order.Handler,
	// authHandler  *auth.Handler,
) *Handlers {
	return &Handlers{
		User: userHandler,
		// Order: orderHandler,
		// Auth:  authHandler,
	}
}
