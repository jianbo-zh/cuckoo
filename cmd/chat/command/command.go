package command

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func Execute() {
	app := cli.App{
		Name: "dchat",
		Commands: []*cli.Command{
			&DaemonCmd,
			&InitCmd,
		},
		Action: func(ctx *cli.Context) error {
			fmt.Printf("chat -> args: %v", ctx.Args())
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
