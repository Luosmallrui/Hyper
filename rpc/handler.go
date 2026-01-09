package handler

import (
	"Hyper/pkg/log"
	"Hyper/pkg/socket"
	"Hyper/rpc/kitex_gen/im/push"
	"Hyper/types"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"go.uber.org/zap"
)

// PushServiceImpl implements the last service interface defined in the IDL.
type PushServiceImpl struct {
}

// PushToClient implements the PushServiceImpl interface.
func (s *PushServiceImpl) PushToClient(
	ctx context.Context,
	req *push.PushRequest,
) (*push.PushResponse, error) {

	// trace 前缀：一条消息贯穿全链路
	trace := fmt.Sprintf(
		"[PUSH msg=%s uid=%d cid=%d event=%s]",
		req.Payload, //q 如果太长，下面会优化
		req.Uid,
		req.Cid,
		req.Event,
	)
	log.L.Info("enter PushToClient", zap.Any("trace", trace))
	//
	//log.Printf("%s enter PushToClient", trace)

	ch := socket.Session.Chat
	if ch == nil {
		log.L.Error("ch is nil")
		return &push.PushResponse{
			Success: false,
			Msg:     "chat channel not initialized",
		}, nil
	}

	client, ok := ch.Client(req.Cid)
	if !ok {
		log.L.Error("client  not found")
		return &push.PushResponse{
			Success: false,
			Msg:     "client offline",
		}, nil
	}

	// 解析 payload
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
		log.L.Error("unmarshal payload", zap.Error(err), zap.Any("payload", req.Payload))
		return &push.PushResponse{
			Success: false,
			Msg:     "invalid payload",
		}, err
	}

	extBytes, err := json.Marshal(m.Ext)
	if err != nil {
		log.L.Error("marshal ext", zap.Error(err), zap.Any("ext", m.Ext))
		extBytes = []byte("{}")
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

	res := &socket.ClientResponse{
		IsAck:   ok,
		Event:   req.Event,
		Content: dto,
		IsSelf:  false,
	}

	if m.TargetID == int64(client.Uid()) {
		res.IsSelf = true
	}
	if err := client.Write(res); err != nil {

		log.L.Error("write response", zap.Error(err))

		return &push.PushResponse{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &push.PushResponse{Success: true}, nil
}
