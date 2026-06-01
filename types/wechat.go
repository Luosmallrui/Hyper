package types

type WeChatSubscribeMessageRequest struct {
	ToUser           string                       `json:"touser"`
	TemplateID       string                       `json:"template_id"`
	Page             string                       `json:"page,omitempty"`
	Data             map[string]WeChatMessageItem `json:"data"`
	MiniprogramState string                       `json:"miniprogram_state,omitempty"`
	Lang             string                       `json:"lang,omitempty"`
}

type WeChatMessageItem struct {
	Value string `json:"value"`
}
