package server

import (
	"Hyper/handler"
)

type Handlers struct {
	Auth *handler.Auth
	Map  *handler.Map
}
