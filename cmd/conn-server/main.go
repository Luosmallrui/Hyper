package main

import (
	"Hyper/config"
	"Hyper/pkg/log"
	"Hyper/rpc"
	"Hyper/rpc/kitex_gen/im/push/pushservice"
	s "Hyper/socket"
	"fmt"

	"net"
	"os"

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

	cliApp := &cli.App{
		Name: "conn-server",

		// 👇 默认启动行为
		Action: func(ctx *cli.Context) error {
			rpcPort := cfg.Server.Rpc
			go startKitexRPC(rpcPort)
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
func startKitexRPC(rpcPort int) {
	h := &handler.PushServiceImpl{}

	svr := pushservice.NewServer(
		h,
		server.WithServiceAddr(&net.TCPAddr{Port: rpcPort}),
	)
	log.L.Info("[RPC] Kitex Push Server is running on ", zap.Int("port", rpcPort))
	if err := svr.Run(); err != nil {
		log.L.Fatal("failed to start  rpc server", zap.Error(err))
	}
}
