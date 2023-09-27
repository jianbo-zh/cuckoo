package command

import (
	"context"
	"fmt"

	"github.com/jianbo-zh/dchat/cmd/chat/httpapi"
	"github.com/jianbo-zh/dchat/cuckoo"
	"github.com/urfave/cli/v2"
)

var CuckooCmd cli.Command

func init() {
	CuckooCmd = cli.Command{
		Name:  "cuckoo",
		Usage: "run cuckoo",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "storage",
				Value: "./.datadir/storage",
				Usage: "set storage dir",
			},
			&cli.StringFlag{
				Name:  "avatar",
				Value: "./.datadir/avatar",
				Usage: "set avatar dir",
			},
			&cli.StringFlag{
				Name:    "host",
				Value:   "",
				Usage:   "set listen host",
				Aliases: []string{"H"},
			},
			&cli.UintFlag{
				Name:    "port",
				Value:   9000,
				Usage:   "set listen port",
				Aliases: []string{"P"},
			},
		},
		Action: func(cCtx *cli.Context) error {

			var storageDir = cCtx.String("storage")
			var avatarDir = cCtx.String("avatar")

			conf, err := cuckoo.LoadConfig(storageDir, avatarDir)
			if err != nil {
				return fmt.Errorf("cuckoo.LoadConfig error: %w", err)
			}

			cuckooInstance, err := cuckoo.NewCuckoo(context.Background(), conf)
			if err != nil {
				panic(err)
			}

			// 启动HTTP服务
			go func() {
				httpapi.Daemon(cuckooInstance, httpapi.Config{
					Host: cCtx.String("host"),
					Port: cCtx.Uint("port"),
				})
			}()

			localhost, _ := cuckooInstance.GetHost()
			fmt.Println("PeerID: ", localhost.ID())

			select {}
		},
	}
}
