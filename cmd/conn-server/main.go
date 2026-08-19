package main

import (
	"Hyper/config"
	"Hyper/pkg/llm"
	"Hyper/pkg/log"
	"Hyper/rpc"
	"Hyper/rpc/kitex_gen/im/push/pushservice"
	s "Hyper/socket"
	"fmt"
	"net"
	"os"
	"syscall"

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
	llm.Init(cfg.Llm)
	conn := InitSocketServer(cfg)

	cliApp := &cli.App{
		Name: "conn-server",
		// 默认 Action 与 serve 子命令走同一条启动路径，保证两种方式都启动 Kitex RPC
		Action: func(ctx *cli.Context) error {
			return run(ctx, cfg, conn)
		},
		Commands: []*cli.Command{
			{
				Name: "serve",
				Action: func(ctx *cli.Context) error {
					return run(ctx, cfg, conn)
				},
			},
		},
	}

	if err := cliApp.Run(os.Args); err != nil {
		log.L.Fatal("failed to start server", zap.Error(err))
	}
}

func run(ctx *cli.Context, cfg *config.Config, conn *s.AppProvider) error {
	go func() {
		// RPC 服务退出属于致命故障：记日志后自我发送 SIGTERM，走主流程的优雅退出
		if err := startKitexRPC(cfg.Server.Rpc, cfg.Nacos, conn.Db, conn.Redis); err != nil {
			log.L.Error("kitex rpc server exited", zap.Error(err))
			if p, err := os.FindProcess(os.Getpid()); err == nil {
				_ = p.Signal(syscall.SIGTERM)
			}
		}
	}()
	return s.Run(ctx, conn)
}

func startKitexRPC(rpcPort int, cfg *config.NacosConfig, Db *gorm.DB, redis *redis.Client) error {
	h := &rpc.PushServiceImpl{Db: Db, Redis: redis}

	listenAddr := &net.TCPAddr{IP: net.IPv4zero, Port: rpcPort}

	//registryAddr := &net.TCPAddr{IP: net.ParseIP(cfg.Address), Port: rpcPort}

	svr := pushservice.NewServer(
		h,
		server.WithServiceAddr(listenAddr), // 指定端口
		server.WithServerBasicInfo(
			&rpcinfo.EndpointBasicInfo{
				ServiceName: "PushService",
			},
		),
	)

	log.L.Info("[RPC] Kitex Server starting", zap.String("listen", listenAddr.String()))

	if err := svr.Run(); err != nil {
		return fmt.Errorf("failed to start rpc server: %w", err)
	}
	return nil
}
