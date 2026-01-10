package config

type Nacos struct {
	Host        string `json:"host" yaml:"host"`
	Port        uint64 `json:"port" yaml:"port"`
	Username    string `json:"username" yaml:"username"`
	Password    string `json:"password" yaml:"password"`
	NamespaceId string `json:"namespace_id" yaml:"namespace_id"`
	GroupName   string `json:"group_name" yaml:"group_name"`
}
