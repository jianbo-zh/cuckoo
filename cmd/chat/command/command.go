package command

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func Execute() {
	app := cli.App{
		Name: "dchat",
		Commands: []*cli.Command{
			// &DaemonCmd,
			&InitCmd,
			&CuckooCmd,
		},
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
