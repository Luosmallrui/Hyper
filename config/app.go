package config

type App struct {
	Env                      string `json:"env" yaml:"env"`
	Debug                    bool   `json:"debug" yaml:"debug"`
	AppID                    string `json:"appid" yaml:"app_id"`
	AppSecret                string `json:"appsecret" yaml:"app_secret"`
	OrganizerApplyTemplateID string `json:"organizer_apply_template_id" yaml:"organizer_apply_template_id"`
}
