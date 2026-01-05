package oss

import (
	"Hyper/config"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

func GetOssClient(conf *config.Config) (*oss.Client, error) {
	provider := credentials.NewEnvironmentVariableCredentialsProvider()
	cfg := oss.LoadDefaultConfig().WithCredentialsProvider(provider).
		WithEndpoint(conf.Oss.Endpoint).WithRegion(conf.Oss.Region)
	client := oss.NewClient(cfg)
	return client, nil
}
