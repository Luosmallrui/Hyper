package types

type SessionDTO struct {
	SessionType int    `json:"session_type"`
	PeerId      uint64 `json:"peer_id"`
	LastMsg     string `json:"last_msg"`
	LastMsgTime int64  `json:"last_msg_time"`
	Unread      uint32 `json:"unread"`
	IsTop       int    `json:"is_top"`
	IsMute      int    `json:"is_mute"` //这是会话免打扰，不是禁言，只不过同名了
	PeerAvatar  string `json:"peer_avatar"`
	PeerName    string `json:"peer_name"`
}
type TalkSessionClearUnreadNumRequest struct {
	// 1=私聊 2=群聊
	SessionType int32 `json:"session_type" binding:"required,oneof=1 2"`

	// 单聊=对方uid 群聊=group_id
	PeerId int32 `json:"peer_id" binding:"required"`

	// 新增：客户端认为“已读到”的时间点（毫秒）
	// 可选字段：不加 binding，老客户端不传也能用
	ReadTime int64 `json:"read_time"`
}

type TalkSessionClearUnreadNumResponse struct {
}
