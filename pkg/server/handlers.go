package server

import (
	"Hyper/handler"
)

type Handlers struct {
	Auth    *handler.Auth
	Map     *handler.Map
	Message *handler.Message
	Note    *handler.Note
	Follow  *handler.Follow
	User    *handler.User
	Session *handler.Session
	Group   *handler.GroupHandler
}
