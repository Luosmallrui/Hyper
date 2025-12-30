package server

import (
	"Hyper/handler"
)

type Handlers struct {
	Auth    *handler.Auth
	Map     *handler.Map
	Message *handler.MessageHandler
	WS      *handler.WSHandler
	Note    *handler.Note

}
