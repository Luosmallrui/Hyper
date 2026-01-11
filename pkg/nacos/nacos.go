package nacos

import (
	"Hyper/config"
	"Hyper/pkg/log"
	"github.com/cloudwego/kitex/pkg/registry"
	nacosreg "github.com/kitex-contrib/registry-nacos/v2/registry"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"go.uber.org/zap"
)

func NewRegistry(cfg *config.NacosConfig) registry.Registry {
	sc := []constant.ServerConfig{
		*constant.NewServerConfig(cfg.Address, cfg.Port),
	}
	cc := constant.ClientConfig{
		NamespaceId:         cfg.Namespace,
		TimeoutMs:           cfg.TimeoutMs,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            cfg.LogLevel,
		AccessKey:           cfg.AccessKeyID,
		SecretKey:           cfg.AccessKeySecret,
	}

	cli, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  &cc,
		ServerConfigs: sc,
	})
	if err != nil {
		log.L.Info("Nacos 无法连接: ", zap.Error(err))
	}
	r := nacosreg.NewNacosRegistry(cli)
	if err != nil {
		log.L.Fatal("err", zap.Error(err))
	}
	log.L.Info("nacos registry created")
	return r
}
