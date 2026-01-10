package nacos

import (
	"Hyper/pkg/log"
	"os"
	"path/filepath"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"go.uber.org/zap"
)

type NacosClient struct {
	client naming_client.INamingClient
}

// NewNacosClient 创建 Nacos 客户端（gRPC 模式）
func NewNacosClient(host string, port uint64, username, password, namespaceId string) (*NacosClient, error) {
	// 使用系统临时目录（跨平台兼容）
	tempDir := os.TempDir()

	clientConfig := constant.ClientConfig{
		NamespaceId:         namespaceId,
		TimeoutMs:           10000,
		NotLoadCacheAtStart: true,
		LogDir:              filepath.Join(tempDir, "nacos", "log"),
		CacheDir:            filepath.Join(tempDir, "nacos", "cache"),
		Username:            username,
		Password:            password,
		LogLevel:            "error",
	}

	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      host,
			Port:        port,
			GrpcPort:    port + 1000, // gRPC 端口 = HTTP 端口 + 1000
			Scheme:      "http",
			ContextPath: "/nacos",
		},
	}

	// 使用 v2 API 创建客户端
	client, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		log.L.Error("failed to create nacos client", zap.Error(err))
		return nil, err
	}

	log.L.Info("nacos client created successfully")
	return &NacosClient{client: client}, nil
}

// RegisterService 注册服务到 Nacos
func (nc *NacosClient) RegisterService(serviceName, ip string, port uint64, groupName string, metadata map[string]string) error {
	success, err := nc.client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          ip,
		Port:        port,
		ServiceName: serviceName,
		GroupName:   groupName,
		Weight:      1,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    metadata,
	})

	if !success || err != nil {
		log.L.Error("failed to register service", zap.String("service", serviceName), zap.Error(err))
		return err
	}

	log.L.Info("service registered successfully", zap.String("service", serviceName))
	return nil
}

// DeregisterService 从 Nacos 注销服务
func (nc *NacosClient) DeregisterService(serviceName, ip string, port uint64, groupName string) error {
	success, err := nc.client.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          ip,
		Port:        port,
		ServiceName: serviceName,
		GroupName:   groupName,
		Ephemeral:   true,
	})

	if !success || err != nil {
		log.L.Error("failed to deregister service", zap.String("service", serviceName), zap.Error(err))
		return err
	}

	log.L.Info("service deregistered successfully", zap.String("service", serviceName))
	return nil
}

// GetServiceInstances 获取服务实例列表
func (nc *NacosClient) GetServiceInstances(serviceName, groupName string) ([]model.Instance, error) {
	instances, err := nc.client.SelectAllInstances(vo.SelectAllInstancesParam{
		ServiceName: serviceName,
		GroupName:   groupName,
	})

	if err != nil {
		log.L.Error("failed to get service instances", zap.String("service", serviceName), zap.Error(err))
		return nil, err
	}

	return instances, nil
}
