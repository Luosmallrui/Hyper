package chat

import (
	"Hyper/pkg/socket"
	"context"
	"encoding/json"
	"log"
)

type KeyboardMessage struct {
	Event   string `json:"event"`
	Payload struct {
		ToFromId int `json:"to_from_id"`
	} `json:"payload"`
}

// onKeyboardMessage 键盘输入事件
func (h *Handler) onKeyboardMessage(ctx context.Context, c socket.IClient, data []byte) {
	var in KeyboardMessage
	if err := json.Unmarshal(data, &in); err != nil {
		log.Println("Chat onKeyboardMessage Err: ", err)
		return
	}

	//_ = h.PushMessage.Push(ctx, types.ImTopicChat, &types.SubscribeMessage{
	//	Event: types.SubEventImMessageKeyboard,
	//	Payload: jsonutil.Encode(types.SubEventImMessageKeyboardPayload{
	//		FromId:   c.Uid(),
	//		ToFromId: in.Payload.ToFromId,
	//	}),
	//})
}
