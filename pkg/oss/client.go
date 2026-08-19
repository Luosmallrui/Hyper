package oss

import (
	"Hyper/config"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

func GetOssClient(conf *config.Config) (*oss.Client, error) {
	// 优先使用配置文件中的 AK/SK，未配置时回退到环境变量
	var provider credentials.CredentialsProvider
	if conf.Oss.AccessKeyID != "" && conf.Oss.AccessKeySecret != "" {
		provider = credentials.NewStaticCredentialsProvider(conf.Oss.AccessKeyID, conf.Oss.AccessKeySecret)
	} else {
		provider = credentials.NewEnvironmentVariableCredentialsProvider()
	}
	cfg := oss.LoadDefaultConfig().WithCredentialsProvider(provider).
		WithEndpoint(conf.Oss.Endpoint).WithRegion(conf.Oss.Region)
	client := oss.NewClient(cfg)
	return client, nil
}
