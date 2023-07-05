package command

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/jianbo-zh/dchat/cmd/chat/config"
	"github.com/urfave/cli/v2"
)

var InitCmd cli.Command

func init() {
	InitCmd = cli.Command{
		Name:  "init",
		Usage: "init dchat config",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "noprint",
				Usage: "no print config file",
				Value: false,
			},
			&cli.StringFlag{
				Name:  "root",
				Value: "",
				Usage: "set root dir",
			},
		},
		Action: func(cCtx *cli.Context) error {
			var conf *config.Config

			var rootPath = cCtx.String("root")
			var configFile = config.DefaultConfigFile

			cfile, err := config.Filename(rootPath, configFile)
			if err != nil {
				return err
			}

			bs, err := os.ReadFile(cfile)
			if err == nil {
				err = json.Unmarshal(bs, &conf)
				if err != nil {
					return err
				}

			} else if os.IsNotExist(err) {
				conf, err = config.Init()
				if err != nil {
					return err
				}

				bs, err = config.Marshal(conf)
				if err != nil {
					return err
				}

				if err = os.WriteFile(cfile, bs, 0600); err != nil {

					if !os.IsNotExist(err) {
						return err
					}

					if err = os.MkdirAll(path.Dir(cfile), 0755); err != nil {
						return err
					}

					if err = os.WriteFile(cfile, bs, 0600); err != nil {
						return err
					}
				}

			} else {
				return err
			}

			if !cCtx.Bool("noprint") {
				bs, err = config.HumanOutput(conf)
				if err != nil {
					return err
				}

				fmt.Printf("%s", bs)
			}

			return nil
		},
	}
}
