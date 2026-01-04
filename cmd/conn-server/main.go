package main

import (
	"Hyper/config"
	"Hyper/socket"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	cfg := config.New("configs/config.dev.yaml")
	conn := InitSocketServer(cfg)

	cliApp := &cli.App{
		Name: "conn-server",

		// 👇 默认启动行为
		Action: func(ctx *cli.Context) error {
			return socket.Run(ctx, conn)
		},

		Commands: []*cli.Command{
			{
				Name: "serve",
				Action: func(ctx *cli.Context) error {
					return socket.Run(ctx, conn)
				},
			},
		},
	}

	if err := cliApp.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
