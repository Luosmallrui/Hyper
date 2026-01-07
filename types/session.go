package types

type SessionDTO struct {
	SessionType int    `json:"session_type"`
	PeerId      uint64 `json:"peer_id"`
	LastMsg     string `json:"last_msg"`
	LastMsgTime int64  `json:"last_msg_time"`
	Unread      uint32 `json:"unread"`
	IsTop       int    `json:"is_top"`
	IsMute      int    `json:"is_mute"`
}
