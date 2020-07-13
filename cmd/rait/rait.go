package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"gitlab.com/NickCao/RAIT/v2/pkg/babeld"
	"gitlab.com/NickCao/RAIT/v2/pkg/rait"
	"go.uber.org/zap"
	"log"
	"os"
)

var Version string
var instance *rait.Instance
var babel *babeld.Babeld

var commonFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "config",
		Usage:   "path to configuration file",
		Aliases: []string{"c"},
		Value:   "/etc/rait/rait.conf",
	},
	&cli.BoolFlag{
		Name:    "debug",
		Usage:   "enable debug log",
		Aliases: []string{"d"},
		Value:   false,
	},
}

var babeldFlags = append(commonFlags,
	&cli.StringFlag{
		Name:    "babel",
		Usage:   "path to babeld control socket",
		Aliases: []string{"b"},
		Value:   "/run/babeld.ctl",
	},
	&cli.StringFlag{
		Name:    "network",
		Usage:   "network type of babeld control socket, unix or tcp",
		Aliases: []string{"n"},
		Value:   "unix",
	})

var commonBeforeFunc = func(ctx *cli.Context) error {
	var err error
	var logger *zap.Logger

	if ctx.Bool("debug") {
		logger, err = zap.NewDevelopment()
		if err != nil {
			return err
		}
		zap.ReplaceGlobals(logger)
	}

	instance, err = rait.InstanceFromPath(ctx.String("config"))
	if err != nil {
		return err
	}
	return nil
}

var babeldBeforeFunc = func(ctx *cli.Context) error {
	babel = &babeld.Babeld{
		Network: ctx.String("network"),
		Address: ctx.String("babel"),
	}
	return commonBeforeFunc(ctx)
}

func main() {
	app := &cli.App{
		Name:      "rait",
		Usage:     "Redundant Array of Inexpensive Tunnels",
		UsageText: "rait [command] [options]",
		Version:   Version,
		Commands: []*cli.Command{{
			Name:      "up",
			Aliases:   []string{"u", "sync"},
			Usage:     "create or sync the tunnels",
			UsageText: "rait up [options]",
			Flags:     commonFlags,
			Before:    commonBeforeFunc,
			Action: func(context *cli.Context) error {
				return instance.SyncInterfaces(true)
			},
		}, {
			Name:      "down",
			Aliases:   []string{"d"},
			Usage:     "destroy the tunnels",
			UsageText: "rait down [options]",
			Flags:     commonFlags,
			Before:    commonBeforeFunc,
			Action: func(context *cli.Context) error {
				return instance.SyncInterfaces(false)
			},
		}, {
			Name:      "render",
			Aliases:   []string{"r"},
			Usage:     "render template based on the desired state of the tunnels",
			UsageText: "rait render [options] SRC DEST",
			Flags:     commonFlags,
			Before:    commonBeforeFunc,
			Action: func(context *cli.Context) error {
				if context.Args().Len() != 2 {
					return fmt.Errorf("expecting two arguments")
				}
				return instance.RenderTemplate(context.Args().Get(0), context.Args().Get(1))
			},
		}, {
			Name:      "babeld",
			Aliases:   []string{"b"},
			Usage:     "interaction with babeld",
			UsageText: "rait babeld [command] [options]",
			Subcommands: []*cli.Command{{
				Name:      "list",
				Aliases:   []string{"l"},
				Usage:     "list babeld interfaces",
				UsageText: "rait babeld list [options]",
				Flags:     babeldFlags,
				Before:    babeldBeforeFunc,
				Action: func(context *cli.Context) error {
					links, err := babel.LinkList()
					if err != nil {
						return err
					}

					for _, link := range links {
						fmt.Println(link)
					}
					return nil
				},
			}, {
				Name:      "sync",
				Aliases:   []string{"s"},
				Usage:     "sync babeld interfaces",
				UsageText: "rait babeld sync [options]",
				Flags:     babeldFlags,
				Before:    babeldBeforeFunc,
				Action: func(context *cli.Context) error {
					links, err := instance.ListInterface()
					if err != nil {
						return err
					}
					return babel.LinkSync(links)
				},
			}},
		}},
		HideHelpCommand: true,
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
