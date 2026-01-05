package models

type Message struct {
	MsgID       string `json:"msg_id"`        // 消息唯一ID (Snowflake)
	SenderID    string `json:"sender_id"`     // 发送者UID
	TargetID    string `json:"target_id"`     // 接收者UID或群ID
	SessionType int    `json:"session_type"`  // 1-单聊,2-群聊,3-系统通知
	MsgType     int    `json:"msg_type"`      // 消息类型
	Content     string `json:"content"`       // 消息正文
	ParentMsgID string `json:"parent_msg_id"` // 回复消息ID
	Timestamp   int64  `json:"timestamp"`     // 发送时间戳(毫秒)
	Status      int    `json:"status"`        // 状态
	Ext         string `json:"ext,omitempty"` // JSON 字符串 // 扩展字段
}

// TableName 指定表名
func (Message) TableName() string {
	return "im_message"
}
