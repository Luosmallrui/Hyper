package config

type App struct {
	Env                      string `json:"env" yaml:"env"`
	Debug                    bool   `json:"debug" yaml:"debug"`
	AppID                    string `json:"appid" yaml:"app_id"`
	AppSecret                string `json:"appsecret" yaml:"app_secret"`
	OrganizerApplyTemplateID string `json:"organizer_apply_template_id" yaml:"organizer_apply_template_id"`
	// QRCodeEnvVersion 小程序码 env_version（develop/trial/release）。
	// 可选字段：缺省或为空时按 "trial" 处理，以保持线上现状（小程序未正式发布前 release 会生成失败）。
	QRCodeEnvVersion string `json:"qr_code_env_version" yaml:"qr_code_env_version"`
}
