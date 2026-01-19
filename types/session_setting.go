package types

// SessionSettingRequest 更新会话设置请求
type SessionSettingRequest struct {
	SessionType int    `json:"session_type" binding:"required"` // 1=单聊 2=群聊
	PeerID      uint64 `json:"peer_id" binding:"required"`      // 单聊=对方uid 群聊=group_id
	IsTop       *int   `json:"is_top" binding:"required"`       // 0/1
	IsMute      *int   `json:"is_mute" binding:"required"`      // 0/1(这是会话免打扰，不是禁言，只不过同名了）
}
