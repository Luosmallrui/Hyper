package types

import "encoding/json"

const SystemMessageTopic = "HYPER_SYSTEM_MSGS"

type FollowPayload struct {
	UserId    int    `json:"user_id"`
	TargetId  int    `json:"target_id"`
	Avatar    string `json:"avatar"`
	Nickname  string `json:"nickname"`
	CreatedAt string `json:"created_at"`
}

type SystemMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type PlatformMessagePayload struct {
	MessageID int64  `json:"message_id"`
	TargetID  int64  `json:"target_id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Type      string `json:"type"`
	Target    string `json:"target"`
	CreatedAt string `json:"created_at"`
}
