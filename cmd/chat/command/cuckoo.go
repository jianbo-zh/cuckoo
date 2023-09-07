package command

import (
	"context"
	"fmt"
	"time"

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
				Value: "./.dchat3",
				Usage: "set storage dir",
			},
		},
		Action: func(cCtx *cli.Context) error {

			var storageDir = cCtx.String("storage")

			conf, err := cuckoo.LoadConfig(storageDir)
			if err != nil {
				return fmt.Errorf("cuckoo.LoadConfig error: %w", err)
			}

			cuckooInstance, err := cuckoo.NewCuckoo(context.Background(), conf)
			if err != nil {
				panic(err)
			}

			localhost, _ := cuckooInstance.GetHost()

			fmt.Println("PeerID: ", localhost.ID())

			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			for t := range ticker.C {
				fmt.Println()
				fmt.Printf("PeerInfo: %s", t)
				for _, addr := range localhost.Addrs() {
					fmt.Println(addr)
				}
			}

			select {}

			return nil
		},
	}
}
