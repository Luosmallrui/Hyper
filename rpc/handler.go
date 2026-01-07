package handler

import (
	"Hyper/pkg/socket"
	"Hyper/rpc/kitex_gen/im/push"
	"Hyper/types"
	"context"
	"encoding/json"
	"strconv"
)

// PushServiceImpl implements the last service interface defined in the IDL.
type PushServiceImpl struct {
}

// PushToClient implements the PushServiceImpl interface.
func (s *PushServiceImpl) PushToClient(ctx context.Context, req *push.PushRequest) (*push.PushResponse, error) {
	ch := socket.Session.Chat
	if ch == nil {
		return &push.PushResponse{Success: false, Msg: "chat channel not initialized"}, nil
	}

	client, ok := ch.Client(req.Cid)
	if !ok {
		return &push.PushResponse{Success: false, Msg: "client offline"}, nil
	}

	var m struct {
		Id          int64                  `json:"msg_id,string"`
		ClientMsgID string                 `json:"client_msg_id"`
		SenderID    int64                  `json:"sender_id,string"`
		TargetID    int64                  `json:"target_id,string"`
		SessionID   string                 `json:"session_id"`
		SessionType int                    `json:"session_type"`
		MsgType     int                    `json:"msg_type"`
		Content     string                 `json:"content"`
		ParentMsgID int64                  `json:"parent_msg_id,string"`
		Timestamp   int64                  `json:"timestamp"`
		Status      int                    `json:"status"`
		Ext         map[string]interface{} `json:"ext"`
	}

	if err := json.Unmarshal([]byte(req.Payload), &m); err != nil {
		return &push.PushResponse{Success: false, Msg: "invalid payload"}, err
	}
	extBytes, err := json.Marshal(m.Ext)
	if err != nil {
		extBytes = []byte("{}") // 序列化失败的兜底
	}
	dto := &types.MessageDTO{
		MsgID:       strconv.FormatInt(m.Id, 10),
		ClientMsgID: m.ClientMsgID,
		SenderID:    strconv.FormatInt(m.SenderID, 10),
		TargetID:    strconv.FormatInt(m.TargetID, 10),
		SessionID:   m.SessionID,
		SessionType: m.SessionType,
		MsgType:     m.MsgType,
		Content:     m.Content,
		Timestamp:   m.Timestamp,
		Status:      m.Status,
		Ext:         extBytes,
	}

	if m.ParentMsgID != 0 {
		dto.ParentMsgID = strconv.FormatInt(m.ParentMsgID, 10)
	}

	err = client.Write(&socket.ClientResponse{
		IsAck:   ok,
		Event:   req.Event,
		Content: dto,
	})
	if err != nil {
		return &push.PushResponse{Success: false, Msg: err.Error()}, nil
	}

	return &push.PushResponse{Success: true}, nil
}
