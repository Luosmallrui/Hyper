package types

const (
	ImTopicChat = "IM_CHAT_MSGS"

	SingleChat = 1 //私聊
	GroupChat  = 2 //群聊
)

type Message struct {
	Id          int64  `json:"msg_id,string"`    // 全局唯一雪花ID
	ClientMsgID string `json:"client_msg_id"`    // 前端生成的UUID，用于幂等去重
	SenderID    int64  `json:"sender_id,string"` // 发送者ID
	TargetID    int64  `json:"target_id,string"` // 接收者ID或群ID
	SessionType int    `json:"session_type"`     // 1-单聊, 2-群聊

	SessionHash int64  `json:"session_hash"` // 数据库索引专用，不需要返回给前端
	SessionID   string `json:"session_id"`   // 原始会话ID，用于碰撞校验和展示

	MsgType     int    `json:"msg_type"` // 1-文本, 2-图片等
	Content     string `json:"content"`  // 消息内容
	ParentMsgID int64  `json:"parent_msg_id,string"`
	Timestamp   int64  `json:"timestamp"`     // 服务端生成的时间戳
	Status      int    `json:"status"`        // 0-发送中, 1-成功, 2-已读, 3-撤回
	Ext         string `json:"ext,omitempty"` // 扩展字段 (JSON字符串)
	Channel     string `json:"channel"`
}
