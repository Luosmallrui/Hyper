package config

type OssConfig struct {
	Endpoint         string `json:"endpoint" yaml:"endpoint"`
	InternalEndpoint string `json:"internal_endpoint" yaml:"internal_endpoint"`
	Region           string `json:"region" yaml:"region"`
	Bucket           string `json:"bucket" yaml:"bucket"`
	AccessKeyID      string `json:"ak" yaml:"ak"`
	AccessKeySecret  string `json:"sk" yaml:"sk"`
	// CDNBaseURL 可选：图片访问的 CDN 域名前缀，缺省保持线上现状
	CDNBaseURL string `json:"cdn_base_url" yaml:"cdn_base_url"`
}

func ProvideOssConfig(cfg *Config) *OssConfig {
	return cfg.Oss
}
