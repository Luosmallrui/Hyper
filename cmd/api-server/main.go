package main

import (
	"Hyper/config"
	"Hyper/pkg/log"
	"Hyper/pkg/server"
	"fmt"
	"os"

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
	appProvider := InitServer(cfg)
	cliApp := &cli.App{
		Name: "api-server",
		Commands: []*cli.Command{
			{
				Name:  "serve",
				Usage: "start http server",
				Action: func(ctx *cli.Context) error {
					return server.Run(ctx, appProvider)
				},
			},
		},
	}
	if err := cliApp.Run(os.Args); err != nil {
		log.L.Fatal("failed to start server", zap.Error(err))
	}
}
