package config

type WechatPayConfig struct {
	AppID                      string `yaml:"app_id"`                        // 应用ID
	MchID                      string `yaml:"mch_id"`                        // 商户号
	MchCertificateSerialNumber string `yaml:"mch_certificate_serial_number"` // 商户证书序列号
	MchAPIv3Key                string `yaml:"mch_apiv3_key"`                 // APIv3密钥
	WechatPayPublicKeyID       string `yaml:"wechat_pay_public_key_id"`      // 微信支付平台公钥ID
	MchPrivateKeyPath          string `yaml:"mch_private_key_path"`          // 商户私钥文件路径
	WechatPayPublicKeyPath     string `yaml:"wechat_pay_public_key_path"`    // 微信支付公钥文件路径
	NotifyURL                  string `yaml:"notify_url"`                    // 支付回调URL
}

func ProvideWechatPayConfig(cfg *Config) *WechatPayConfig {
	return cfg.WechatPayConfig
}
