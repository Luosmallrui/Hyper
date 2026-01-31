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
	"github.com/redis/go-redis/v9"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}
	path := fmt.Sprintf("configs/config.%s.yaml", env)
	cfg := config.New(path)
	conn := InitSocketServer(cfg)

	cliApp := &cli.App{
		Name: "conn-server",
		Action: func(ctx *cli.Context) error {
			rpcPort := cfg.Server.Rpc
			go startKitexRPC(rpcPort, cfg.Nacos, conn.Db, conn.Redis)
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
func startKitexRPC(rpcPort int, cfg *config.NacosConfig, Db *gorm.DB, redis *redis.Client) {
	h := &handler.PushServiceImpl{Db: Db, Redis: redis}
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

	log.L.Info("[RPC] Kitex Server starting", zap.String("listen", listenAddr.String()), zap.String("registry", registryAddr.String()))

	if err := svr.Run(); err != nil {
		log.L.Fatal("failed to start rpc server", zap.Error(err))
	}
}
