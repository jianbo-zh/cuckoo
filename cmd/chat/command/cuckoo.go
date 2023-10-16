package command

import (
	"context"
	"fmt"
	"path/filepath"

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
				Name:  "datadir",
				Value: "./.storage/datadir",
				Usage: "set storage dir",
			},
			&cli.StringFlag{
				Name:  "resdir",
				Value: "./.storage/resdir",
				Usage: "set resource dir",
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

			dataDir, err := filepath.Abs(cCtx.String("datadir"))
			if err != nil {
				return fmt.Errorf("filepath.Abs error: %w", err)
			}

			resourceDir, err := filepath.Abs(cCtx.String("resdir"))
			if err != nil {
				return fmt.Errorf("filepath.Abs error: %w", err)
			}

			conf, err := cuckoo.LoadConfig(dataDir, resourceDir, resourceDir, resourceDir)
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
