package main

import (
	"Hyper/config"
	"Hyper/rpc"
	"Hyper/rpc/kitex_gen/im/push/pushservice"
	s "Hyper/socket"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/cloudwego/kitex/server"
	"github.com/urfave/cli/v2"
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
			go startKitexRPC()
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
		log.Fatal(err)
	}
}
func startKitexRPC() {
	h := &handler.PushServiceImpl{}

	// 创建 Kitex Server
	svr := pushservice.NewServer(
		h,
		server.WithServiceAddr(&net.TCPAddr{Port: 8083}),
	)

	log.Println("[RPC] Kitex Push Server is running on :8083...")
	if err := svr.Run(); err != nil {
		log.Printf("[RPC] Kitex server run failed: %v", err)
	}
}
