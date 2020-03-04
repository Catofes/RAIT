package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"gitlab.com/NickCao/RAIT/rait"
	"io/ioutil"
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
						Name:    "prefix",
						Aliases: []string{"p"},
						Usage:   "Wireguard links will be created with prefix `PREFIX`",
						Value:   "rait",
					},
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
					var r *rait.RAIT
					var p []*rait.Peer
					var err error
					r, err = rait.NewRAITFromToml(c.String("config"))
					if err != nil {
						return err
					}
					p, err = rait.LoadPeersFromTomls(c.String("peers"))
					if err != nil {
						return err
					}
					return r.SetupLinks(c.String("prefix"), p)
				},
			},
			{
				Name:  "load",
				Usage: "same as up, but from json and stdin (for programmatic invocations)",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "prefix",
						Aliases: []string{"p"},
						Usage:   "Wireguard links will be created with prefix `PREFIX`",
						Value:   "rait",
					},
				},
				Action: func(c *cli.Context) error {
					var r *rait.RAIT
					var p []*rait.Peer
					var data []byte
					var err error
					data, err = ioutil.ReadAll(os.Stdin)
					if err != nil {
						return nil
					}
					r, p, err = rait.LoadFromJSON(data)
					if err != nil {
						return nil
					}
					return r.SetupLinks(c.String("prefix"), p)
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
						Value:   "rait",
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
		fmt.Println(err)
	}
}
