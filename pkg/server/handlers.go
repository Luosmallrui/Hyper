package server

import (
	"Hyper/handler"
)

type Handlers struct {
	Auth    *handler.Auth
	Map     *handler.Map
	Message *handler.MessageHandler
	Note    *handler.Note
	Follow  *handler.Follow
}
