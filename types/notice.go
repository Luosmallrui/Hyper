package types

import "encoding/json"

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
