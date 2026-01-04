package chat

import (
	"Hyper/pkg/socket"
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

type handle func(ctx context.Context, client socket.IClient, data []byte)

var handlers map[string]handle

type Handler struct {
	Redis *redis.Client
	//Source        *dao.Source
	//MemberService service.IGroupMemberService
	//PushMessage   *business.PushMessage
}

func (h *Handler) init() {
	handlers = make(map[string]handle)
	// 注册自定义绑定事件
	handlers["im.message.keyboard"] = h.onKeyboardMessage
}

func (h *Handler) Call(ctx context.Context, client socket.IClient, event string, data []byte) {

	if handlers == nil {
		h.init()
	}

	if call, ok := handlers[event]; ok {
		call(ctx, client, data)
	} else {
		log.Printf("Chat Event: [%s]未注册回调事件\n", event)
	}
}
