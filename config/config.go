package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 配置信息
type Config struct {
	App             *App             `json:"app" yaml:"app"`
	Redis           *Redis           `json:"redis" yaml:"redis"`
	MySQL           *MySQL           `json:"mysql" yaml:"mysql"`
	Jwt             *Jwt             `json:"jwt" yaml:"jwt"`
	Oss             *OssConfig       `json:"oss" yaml:"oss"`
	Nacos           *NacosConfig     `json:"nacos" yaml:"nacos"`
	Server          *Server          `json:"server" yaml:"server"`
	RocketMQ        *RocketMQConfig  `json:"rocketmq" yaml:"rocketmq"`
	WechatPayConfig *WechatPayConfig `json:"wechat_pay" yaml:"wechat_pay"`
}

type Server struct {
	Http      int `json:"http" yaml:"http"`
	Websocket int `json:"websocket" yaml:"websocket"`
	Tcp       int `json:"tcp" yaml:"tcp"`
	Rpc       int `json:"rpc" yaml:"rpc"`
}

func New(filename string) *Config {

	content, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var conf Config
	if yaml.Unmarshal(content, &conf) != nil {
		panic(fmt.Sprintf("解析 config.yaml 读取错误: %v", err))
	}

	return &conf
}

// Debug 调试模式
func (c *Config) Debug() bool {
	return c.App.Debug
}
