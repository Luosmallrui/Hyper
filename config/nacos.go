package config

type NacosConfig struct {
	Address   string `yaml:"address"`
	Port      uint64 `yaml:"port"`
	Namespace string `yaml:"namespace"`
	User      string `yaml:"user"`
	Password  string `yaml:"password"`
	TimeoutMs uint64 `yaml:"timeout_ms"`
	LogLevel  string `yaml:"log_level"`
}
