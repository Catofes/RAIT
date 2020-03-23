package main

import (
	"github.com/urfave/cli/v2"
	"gitlab.com/NickCao/RAIT/rait"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Name:  "RAIT",
		Usage: "Redundant Array of Inexpensive Tunnels",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Aliases:  []string{"c"},
				Usage:    "Load config from `FILE`",
				Required: false,
				Value:    "/etc/rait/rait.conf",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "up",
				Usage: "bring up the wireguard mesh",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "peers",
						Aliases:  []string{"p"},
						Usage:    "Load peers from `DIR`",
						Required: false,
						Value:    "/etc/rait/peers",
					},
				},
				Action: func(c *cli.Context) error {
					return rait.EntryUp(c.String("config"), c.String("peers"))
				},
			},
			{
				Name:  "down",
				Usage: "destroy the wireguard mesh",
				Action: func(c *cli.Context) error {
					return rait.EntryDown(c.String("config"))
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
