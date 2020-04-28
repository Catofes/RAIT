package main

import (
	"github.com/urfave/cli/v2"
	"gitlab.com/NickCao/RAIT/rait"
	"log"
	"os"
)

var (
	Version string
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
						Value:    "",
					},
				},
				Action: func(c *cli.Context) error {
					var client *rait.Client
					var err error
					client, err = rait.LoadClientFromPath(c.String("config"))
					if err != nil {
						return err
					}
					if c.String("peers") != "" {
						client.Peers = c.String("peers")
					}
					var peers []*rait.Peer
					peers, err = rait.LoadPeersFromPath(client.Peers)
					if err != nil {
						return err
					}
					return client.SyncWireguardInterfaces(peers)
				},
			},
			{
				Name:  "down",
				Usage: "destroy the wireguard mesh",
				Action: func(c *cli.Context) error {
					var client *rait.Client
					var err error
					client, err = rait.LoadClientFromPath(c.String("config"))
					if err != nil {
						return err
					}
					return client.SyncWireguardInterfaces([]*rait.Peer{})
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
					var client *rait.Client
					var err error
					client, err = rait.LoadClientFromPath(c.String("config"))
					if err != nil {
						return err
					}
					var output []byte
					output, err = client.RenderTemplate(c.String("from"))
					if err != nil {
						return err
					}
					_, err = os.Stdout.Write(output)
					if err != nil {
						return err
					}
					return nil
				},
			},
		},
	}
	app.Version = Version
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
