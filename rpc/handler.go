package handler

import (
	"Hyper/pkg/socket"
	"Hyper/rpc/kitex_gen/im/push"
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// PushServiceImpl implements the last service interface defined in the IDL.
type PushServiceImpl struct {
}

// PushToClient implements the PushServiceImpl interface.
func (s *PushServiceImpl) PushToClient(ctx context.Context, req *push.PushRequest) (resp *push.PushResponse, err error) {
	ch := socket.Session.Chat

	fmt.Println(req.Payload, 55, req.Cid)
	fmt.Println("Okoko")
	if ch == nil {
		return &push.PushResponse{Success: false, Msg: "chat channel not initialized"}, nil
	}

	client, ok := ch.Client(req.Cid)
	if !ok {
		fmt.Println("client is nil")
		return &push.PushResponse{Success: false, Msg: "client offline on this node"}, nil
	}
	var payload any
	if err := json.Unmarshal([]byte(req.Payload), &payload); err != nil {
		log.Printf("payload 反序列化失败: %v", err)
		return &push.PushResponse{Success: false, Msg: "failed to get body"}, err
	}

	err = client.Write(&socket.ClientResponse{
		IsAck:   ok,
		Event:   req.Event,
		Content: payload,
	})

	if err != nil {
		fmt.Println("write error:", err)
		return &push.PushResponse{Success: false, Msg: err.Error()}, nil
	}

	return &push.PushResponse{Success: true}, nil
}
