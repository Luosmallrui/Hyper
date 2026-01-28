package server

import (
	"Hyper/handler"
)

type Handlers struct {
	Auth            *handler.Auth
	Pay             *handler.Pay
	Map             *handler.Map
	Message         *handler.Message
	Note            *handler.Note
	Follow          *handler.Follow
	User            *handler.User
	Session         *handler.Session
	Group           *handler.GroupHandler
	GroupMember     *handler.GroupMemberHandler
	CommentsHandler *handler.CommentsHandler
	TopicHandler    *handler.TopicHandler
	ProductHandler  *handler.ProductHandler
	Party           *handler.Party
	Points          *handler.Point
	Order           *handler.Order
}
