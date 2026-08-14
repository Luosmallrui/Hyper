package config

type LlmConfig struct {
	APIKey  string `json:"api_key" yaml:"api_key"`
	BaseURL string `json:"base_url" yaml:"base_url"`
	// TagModel 图片标签生成模型（视觉）
	TagModel string `json:"tag_model" yaml:"tag_model"`
	// ClassifyModel 笔记频道分类模型（视觉）
	ClassifyModel string `json:"classify_model" yaml:"classify_model"`
}
