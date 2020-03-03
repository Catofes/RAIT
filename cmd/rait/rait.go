package main

import (
	"github.com/urfave/cli/v2"
	"gitlab.com/NickCao/RAIT/rait"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "up",
				Usage: "time to take flight",
				Action: func(c *cli.Context) error {
					r, err := rait.NewRAITFromFile(c.Args().First())
					if err != nil {
						return err
					}
					return r.SetupLinks()
				},
			},
			{
				Name:  "down",
				Usage: "the big red button",
				Action: func(c *cli.Context) error {
					return rait.DestroyLinks(c.Args().First())
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
