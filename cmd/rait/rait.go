package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"gitlab.com/NickCao/RAIT/rait"
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
				Name:  "load",
				Usage: "same as up, but from json and stdin (for programmatic invocations)",
				Action: func(c *cli.Context) error {
					return rait.RAITLoad(os.Stdin)
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
		fmt.Println(err)
	}
}
