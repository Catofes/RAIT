package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"gitlab.com/NickCao/RAIT/v2/pkg/babeld"
	"log"
	"os"
)

func main() {
	commonFlags := []cli.Flag{
		&cli.StringFlag{
			Name:    "control",
			Usage:   "path to babeld control socket",
			Aliases: []string{"c"},
			Value:   "/run/babeld.ctl",
		},
		&cli.StringFlag{
			Name:    "network",
			Usage:   "type of babeld control socket",
			Aliases: []string{"n"},
			Value:   "unix",
		},
	}

	app := &cli.App{
		Name:      "babelctl",
		UsageText: "babelctl [command] [options]",
		Commands: []*cli.Command{{
			Name:      "list",
			Aliases:   []string{"l"},
			Usage:     "list interfaces",
			UsageText: "babelctl list [options]",
			Flags:     commonFlags,
			Action: func(context *cli.Context) error {
				links, err := (&babeld.Babeld{
					Network: context.String("network"),
					Address: context.String("control"),
				}).ListList()

				if err != nil {
					return err
				}

				for _, link := range links {
					fmt.Println(link)
				}
				return nil
			},
		},
			{
				Name:      "sync",
				Aliases:   []string{"s"},
				Usage:     "sync interfaces",
				UsageText: "babelctl sync [options] [interfaces]...",
				Flags:     commonFlags,
				Action: func(context *cli.Context) error {
					err := (&babeld.Babeld{
						Network: context.String("network"),
						Address: context.String("control"),
					}).LinkSync(context.Args().Slice())

					if err != nil {
						return err
					}
					return nil
				},
			}},
		HideHelpCommand: true,
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
