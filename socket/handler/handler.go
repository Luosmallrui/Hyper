package handler

import (
	"Hyper/config"
	"Hyper/pkg/socket"
)

type Handler struct {
	Chat        *ChatChannel
	Config      *config.Config
	RoomStorage *socket.RoomStorage
}
