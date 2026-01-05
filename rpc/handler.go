package handler

import (
	"Hyper/pkg/socket"
	"Hyper/rpc/kitex_gen/im/push"
	"context"
)

// PushServiceImpl implements the last service interface defined in the IDL.
type PushServiceImpl struct {
}

// PushToClient implements the PushServiceImpl interface.
func (s *PushServiceImpl) PushToClient(ctx context.Context, req *push.PushRequest) (resp *push.PushResponse, err error) {
	ch := socket.Session.Chat

	if ch == nil {
		return &push.PushResponse{Success: false, Msg: "chat channel not initialized"}, nil
	}

	// 2. 通过 Cid 查找 Client
	// 你在 socket.Channel 结构体中已经实现了 Client(cid int64) (*Client, bool)
	client, ok := ch.Client(req.Cid)
	if !ok {
		return &push.PushResponse{Success: false, Msg: "client offline on this node"}, nil
	}

	// 3. 直接调用 client.Write 发送消息
	// 这会把消息塞入 client.outChan，然后由 loopWrite 协程推送
	err = client.Write(&socket.ClientResponse{
		Event:   req.Event,
		Content: req.Payload,
	})

	if err != nil {
		return &push.PushResponse{Success: false, Msg: err.Error()}, nil
	}

	return &push.PushResponse{Success: true}, nil
}
