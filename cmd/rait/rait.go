package main

import (
	"github.com/urfave/cli/v2"
	"gitlab.com/NickCao/RAIT/rait"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Name: "RAIT",
		Usage: "Redundant Array of Inexpensive Tunnels",
		Commands: []*cli.Command{
			{
				Name:  "up",
				Usage: "time to take flight",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "Load configuration from `FILE`",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					r, err := rait.NewRAITFromFile(c.String("config"))
					if err != nil {
						return err
					}
					return r.SetupLinks()
				},
			},
			{
				Name:  "down",
				Usage: "the big red button",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "prefix",
						Aliases: []string{"p"},
						Usage:   "Delete all wireguard links with prefix `PREFIX`",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					return rait.DestroyLinks(c.String("prefix"))
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
