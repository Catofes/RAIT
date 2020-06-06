package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"gitlab.com/NickCao/RAIT/v2/pkg/rait"
	"log"
	"os"
)

var Version string

func main() {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Usage:   "path to configuration file",
			Aliases: []string{"c"},
			EnvVars: []string{"RAIT_CONFIG"},
			Value:   "/etc/rait/rait.conf",
		},
	}
	app := &cli.App{
		Name:      "rait",
		Usage:     "Redundant Array of Inexpensive Tunnels",
		UsageText: "rait [command] [options]",
		Version:   Version,
		Commands: []*cli.Command{{
			Name:      "up",
			Aliases:   []string{"u"},
			Usage:     "create or sync the tunnels",
			UsageText: "rait up [options]",
			Flags:     flags,
			Action: func(context *cli.Context) error {
				instance, err := rait.InstanceFromPath(context.String("config"))
				if err != nil {
					return err
				}
				return instance.SyncInterfaces(true)
			},
		}, {
			Name:      "down",
			Aliases:   []string{"d"},
			Usage:     "destroy the tunnels",
			UsageText: "rait down [options]",
			Flags:     flags,
			Action: func(context *cli.Context) error {
				instance, err := rait.InstanceFromPath(context.String("config"))
				if err != nil {
					return err
				}
				return instance.SyncInterfaces(false)
			},
		}, {
			Name:      "render",
			Aliases:   []string{"r"},
			Usage:     "render template based on the desired state of the tunnels",
			UsageText: "rait render [options] SRC DEST",
			Flags:     flags,
			Action: func(context *cli.Context) error {
				if context.Args().Len() != 2 {
					return fmt.Errorf("expecting two arguments")
				}
				instance, err := rait.InstanceFromPath(context.String("config"))
				if err != nil {
					return err
				}
				return instance.RenderTemplate(context.Args().Get(0), context.Args().Get(1))
			},
		}},
		HideHelpCommand: true,
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
