package main

import (
	"Hyper/config"
	"Hyper/pkg/log"
	"Hyper/pkg/nacos"
	"Hyper/rpc"
	"Hyper/rpc/kitex_gen/im/push/pushservice"
	s "Hyper/socket"
	"fmt"
	"net"
	"os"

	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func main() {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}
	path := fmt.Sprintf("configs/config.%s.yaml", env)
	cfg := config.New(path)
	conn := InitSocketServer(cfg)

	// 初始化 Nacos 客户端并注册 conn-server
	nacosClient, err := nacos.NewNacosClient(
		cfg.Nacos.Host,
		cfg.Nacos.Port,
		cfg.Nacos.Username,
		cfg.Nacos.Password,
		cfg.Nacos.NamespaceId,
	)
	if err != nil {
		log.L.Fatal("failed to init nacos client", zap.Error(err))
	}

	if err := nacosClient.RegisterService(
		"Krito-Test",
		"10.20.9.18", //后续更改为获取本机IP
		uint64(cfg.Server.Websocket),
		cfg.Nacos.GroupName,
		map[string]string{"env": cfg.App.Env},
	); err != nil {
		log.L.Fatal("failed to register conn-server", zap.Error(err))
	}

	cliApp := &cli.App{
		Name: "conn-server",
<<<<<<< HEAD

		Action: func(ctx *cli.Context) error {
			rpcPort := cfg.Server.Rpc
			go startKitexRPC(rpcPort, cfg, nacosClient)
=======
		Action: func(ctx *cli.Context) error {
			rpcPort := cfg.Server.Rpc
			go startKitexRPC(rpcPort, cfg.Nacos)
>>>>>>> 7f36704970a7bb1dec9dc3c3a710f5cbec013f19
			return s.Run(ctx, conn)
		},
		Commands: []*cli.Command{
			{
				Name: "serve",
				Action: func(ctx *cli.Context) error {
					return s.Run(ctx, conn)
				},
			},
		},
	}

	if err := cliApp.Run(os.Args); err != nil {
		log.L.Fatal("failed to start server", zap.Error(err))
	}
}
<<<<<<< HEAD

func startKitexRPC(rpcPort int, cfg *config.Config, nacosClient *nacos.NacosClient) {
=======
func startKitexRPC(rpcPort int, cfg *config.NacosConfig) {
>>>>>>> 7f36704970a7bb1dec9dc3c3a710f5cbec013f19
	h := &handler.PushServiceImpl{}
	nacosRegistry := nacos.NewRegistry(cfg)

	listenAddr := &net.TCPAddr{IP: net.IPv4zero, Port: rpcPort}

	registryAddr := &net.TCPAddr{IP: net.ParseIP(cfg.Address), Port: rpcPort}

	svr := pushservice.NewServer(
		h,
		server.WithRegistry(nacosRegistry),
		server.WithServiceAddr(listenAddr),
		server.WithRegistryInfo(&registry.Info{
			ServiceName: "PushService",
			Addr:        registryAddr,
			Tags:        map[string]string{"node_id": "ws-01"},
		}),
		server.WithServerBasicInfo(
			&rpcinfo.EndpointBasicInfo{
				ServiceName: "PushService",
			},
		),
	)
<<<<<<< HEAD
	log.L.Info("[RPC] Kitex Push Server is running on ", zap.Int("port", rpcPort))

	// 注册 push-service 到 Nacos
	if err := nacosClient.RegisterService(
		"push-service",
		"10.20.9.18",
		uint64(rpcPort),
		cfg.Nacos.GroupName,
		map[string]string{"type": "rpc"},
	); err != nil {
		log.L.Fatal("failed to register push-service", zap.Error(err))
	}
=======

	log.L.Info("[RPC] Kitex Server starting", zap.String("listen", listenAddr.String()), zap.String("registry", registryAddr.String()))
>>>>>>> 7f36704970a7bb1dec9dc3c3a710f5cbec013f19

	if err := svr.Run(); err != nil {
		log.L.Fatal("failed to start rpc server", zap.Error(err))
	}
}
