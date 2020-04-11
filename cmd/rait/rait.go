package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"gitlab.com/NickCao/RAIT/rait"
	"os"
)

var (
	Version string
	Buildtime string
)

func main() {
	app := &cli.App{
		Name:  "RAIT",
		Usage: "Redundant Array of Inexpensive Tunnels",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Aliases:  []string{"c"},
				Usage:    "Load config from `URL`",
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
						Usage:    "Load peers from `URL`",
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
			{
				Name:  "render",
				Usage: "render templates to generate routing daemon configurations",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "from",
						Aliases:  []string{"f"},
						Usage:    "Load template from `URL`",
						Required: false,
						Value:    "stdin://",
					},
				},
				Action: func(c *cli.Context) error {
					return rait.EntryRender(c.String("config"), c.String("from"))
				},
			},
		},
	}
	app.Version = Version + " " + Buildtime
	err := app.Run(os.Args)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
