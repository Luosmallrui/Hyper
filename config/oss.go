package config

type OssConfig struct {
	Endpoint         string `json:"endpoint" yaml:"endpoint"`
	InternalEndpoint string `json:"internal_endpoint" yaml:"internal_endpoint"`
	Region           string `json:"region" yaml:"region"`
	Bucket           string `json:"bucket" yaml:"bucket"`
	AccessKeyID      string `json:"ak" yaml:"ak"`
	AccessKeySecret  string `json:"sk" yaml:"sk"`
}

func ProvideOssConfig(cfg *Config) *OssConfig {
	return cfg.Oss
}
