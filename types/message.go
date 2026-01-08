package types

import "encoding/json"

const (
	ImTopicChat = "IM_CHAT_MSGS"

	SessionTypeSingle         = 1 //私聊
	GroupChatSessionTypeGroup = 2 // 群聊
	SessionTypeSystem         = 3 // 系统通知/服服务号
	MsgTypeText               = 1 // 文本消息
	MsgTypeImage              = 2 // 图片消息
	MsgTypeAudio              = 3 // 语音消息
	MsgTypeVideo              = 4 // 视频消息
	MsgTypeFile               = 5 // 文件消息
	MsgTypeLocation           = 6 // 位置消息
	MsgTypeInteraction        = 7 // 互动消息（点赞、关注提醒）
	MsgTypeCard               = 8 // 卡片/链接消息
	MsgStatusSending          = 0 // 发送中/待处理
	MsgStatusSuccess          = 1 // 发送成功
	MsgStatusRead             = 2 // 已读
	MsgStatusRevoked          = 3 // 已撤回
	MsgStatusFailed           = 4 // 发送失败
	MsgStatusDeleted          = 5 // 逻辑删除
)

const (
	ChannelChat         = "chat"         // 常规聊天
	ChannelSystem       = "system"       // 系统推送
	ChannelNotification = "notification" // 业务通知（如：关注、评论、收藏）
	ChannelControl      = "control"      // 控制指令（如：强制下线、多端同步）
	ChannelHeartbeat    = "heartbeat"    // 客户端心跳
)
const (
	ExtKeyDevice     = "device"      // 设备信息 (e.g., iPhone 15)
	ExtKeyAtUsers    = "at_users"    // @用户列表
	ExtKeyReplyCount = "reply_count" // 回复数
	ExtKeyBadge      = "badge"       // 角标数字
	ExtKeyIsSilent   = "is_silent"   // 是否静默消息
)

type Message struct {
	Id          int64  `json:"msg_id,string"`    // 全局唯一雪花ID
	ClientMsgID string `json:"client_msg_id"`    // 前端生成的UUID，用于幂等去重
	SenderID    int64  `json:"sender_id,string"` // 发送者ID
	TargetID    int64  `json:"target_id,string"` // 接收者ID或群ID
	SessionType int    `json:"session_type"`     // 1-单聊, 2-群聊

	SessionHash int64  `json:"session_hash"` // 数据库索引专用，不需要返回给前端
	SessionID   string `json:"session_id"`   // 原始会话ID，用于碰撞校验和展示

	MsgType     int                    `json:"msg_type"` // 1-文本, 2-图片等
	Content     string                 `json:"content"`  // 消息内容
	ParentMsgID int64                  `json:"parent_msg_id,string"`
	Timestamp   int64                  `json:"timestamp"` // 服务端生成的时间戳
	Status      int                    `json:"status"`    // 0-发送中, 1-成功, 2-已读, 3-撤回
	Ext         map[string]interface{} `json:"ext"`       // 扩展字段 (JSON字符串)
	Channel     string                 `json:"channel"`
}

// MessageDTO 最终推送到前端的消息结构
type MessageDTO struct {
	MsgID       string          `json:"msg_id"`        // 转为 string，避免 JS 精度丢失
	ClientMsgID string          `json:"client_msg_id"` // 前端生成的 ID
	SenderID    string          `json:"sender_id"`     // 转为 string
	TargetID    string          `json:"target_id"`     // 转为 string
	SessionID   string          `json:"session_id"`    // 会话 ID
	SessionType int             `json:"session_type"`  // 1-私聊, 2-群聊
	MsgType     int             `json:"msg_type"`      // 1-文本...
	Content     string          `json:"content"`       // 消息正文
	ParentMsgID string          `json:"parent_msg_id,omitempty"`
	Timestamp   int64           `json:"timestamp"`     // 毫秒时间戳
	Status      int             `json:"status"`        // 消息状态
	Ext         json.RawMessage `json:"ext,omitempty"` // 关键：不再是字符串，而是原始 JSON 对象
}

type ListMessageReq struct {
	Id       uint64                 `json:"id"`
	SenderId uint64                 `json:"sender_id"`
	Content  string                 `json:"content"`
	MsgType  int                    `json:"msg_type"`
	Ext      map[string]interface{} `json:"ext"`
	Time     int64                  `json:"time"`
}

type TalkSessionClearUnreadNumRequest struct {
	// TalkMode indicates the chat mode (e.g., 1 for private chat, 2 for group chat)
	SessionType int32 `json:"session_type" binding:"required,oneof=1 2"`

	// ToFromID represents the ID of the other participant in the chat
	PeerId int32 `json:"peer_id" binding:"required"`
}

type TalkSessionClearUnreadNumResponse struct {
}
