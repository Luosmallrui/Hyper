package config

type RocketMQConfig struct {
	NameServer []string `yaml:"nameserver"`

	Producer Producer `yaml:"producer"`

	Consumer Consumer `yaml:"consumer"`
}

type Producer struct {
	Group string `yaml:"group"`
	Retry int    `yaml:"retry"`
}

type Consumer struct {
	Group string `yaml:"group"`
}

func ProvideRocketMQConfig(cfg *Config) *RocketMQConfig {
	return cfg.RocketMQ
}
