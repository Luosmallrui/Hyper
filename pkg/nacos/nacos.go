package nacos

import (
	"Hyper/pkg/log"

	"github.com/cloudwego/kitex/pkg/registry"
	nacosreg "github.com/kitex-contrib/registry-nacos/v2/registry"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"go.uber.org/zap"
)

func NewRegistry() registry.Registry {
	sc := []constant.ServerConfig{
		*constant.NewServerConfig("47.109.188.142", 9848),
	}
	cc := constant.ClientConfig{
		NamespaceId:         "public",
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "info",
	}

	cli, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	_, err = cli.GetService(vo.GetServiceParam{
		ServiceName: "DUMMY-CHECK",
	})
	if err != nil {
		panic("nacos not reachable: " + err.Error())
	}
	r := nacosreg.NewNacosRegistry(cli)
	if err != nil {
		log.L.Fatal("err", zap.Error(err))
	}
	log.L.Info("nacos registry created")
	return r
}
