package config

<<<<<<< HEAD
type Nacos struct {
	Host        string `json:"host" yaml:"host"`
	Port        uint64 `json:"port" yaml:"port"`
	Username    string `json:"username" yaml:"username"`
	Password    string `json:"password" yaml:"password"`
	NamespaceId string `json:"namespace_id" yaml:"namespace_id"`
	GroupName   string `json:"group_name" yaml:"group_name"`
=======
type NacosConfig struct {
	Address         string `yaml:"address"`
	Port            uint64 `yaml:"port"`
	Namespace       string `yaml:"namespace"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	TimeoutMs       uint64 `yaml:"timeout_ms"`
	LogLevel        string `yaml:"log_level"`
	AccessKeyID     string `json:"ak" yaml:"ak"`
	AccessKeySecret string `json:"sk" yaml:"sk"`
>>>>>>> 7f36704970a7bb1dec9dc3c3a710f5cbec013f19
}
