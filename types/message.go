package types

type Message struct {
	Id          int64  `json:"msg_id,string"`
	ClientMsgID string `json:"client_msg_id"` // 客户端生成的UUID，用于重发去重
	SenderID    int64  `json:"sender_id"`
	TargetID    int64  `json:"target_id"`
	SessionType int    `json:"session_type"`
	MsgType     int    `json:"msg_type"`
	Content     string `json:"content"`
	ParentMsgID int64  `json:"parent_msg_id"` // 引用消息ID也建议用int64
	Timestamp   int64  `json:"timestamp"`
	Status      int    `json:"status"`
	Ext         string `json:"ext,omitempty"` // 建议直接存JSON字符串，性能更好
}
