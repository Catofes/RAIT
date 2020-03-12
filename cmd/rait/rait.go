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
		Commands: []*cli.Command{
			{
				Name:  "up",
				Usage: "time to take flight",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "config",
						Aliases:  []string{"c"},
						Usage:    "Load rait configuration from `FILE`",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "peers",
						Aliases:  []string{"d"},
						Usage:    "Load peers from `DIR`",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					return rait.RAITUp(c.String("config"), c.String("peers"))
				},
			},
			{
				Name:  "down",
				Usage: "the big red button",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "config",
						Aliases:  []string{"c"},
						Usage:    "Load rait configuration from `FILE`",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					return rait.RAITDown(c.String("config"))
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Println(err)
	}
}
