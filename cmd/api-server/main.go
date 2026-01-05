package main

import (
	"Hyper/config"
	"Hyper/pkg/server"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	cfg := config.New("configs/config.dev.yaml")
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
		log.Fatal(err)
	}
}
