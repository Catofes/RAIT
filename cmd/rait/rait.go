package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"gitlab.com/NickCao/RAIT/v2/pkg/babeld"
	"gitlab.com/NickCao/RAIT/v2/pkg/rait"
	"go.uber.org/zap"
	"os"
)

var Version string
var ra *rait.RAIT
var ba *babeld.Babeld

var commonFlags = []cli.Flag{
	&cli.StringSliceFlag{
		Name:    "config",
		Usage:   "path to configuration file, can be specified multiple times",
		Aliases: []string{"c"},
		Value:   cli.NewStringSlice("/etc/rait/rait.conf"),
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
		Name:    "babeld",
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
	ba = babeld.NewBabeld(ctx.String("network"), ctx.String("babeld"))

	var err error
	var logger *zap.Logger

	config := zap.NewDevelopmentConfig()
	config.DisableStacktrace = true
	switch ctx.Bool("debug") {
	case true:
		config.Level.SetLevel(zap.DebugLevel)
	case false:
		config.Level.SetLevel(zap.InfoLevel)
	}
	logger, err = config.Build()

	if err != nil {
		return err
	}
	zap.ReplaceGlobals(logger)

	ra, err = rait.RAITFromPaths(ctx.StringSlice("config"))
	if err != nil {
		return err
	}
	return nil
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
			Action: func(ctx *cli.Context) error {
				return ra.SyncInterfaces(true)
			},
		}, {
			Name:      "down",
			Aliases:   []string{"d"},
			Usage:     "destroy the tunnels",
			UsageText: "rait down [options]",
			Flags:     commonFlags,
			Before:    commonBeforeFunc,
			Action: func(ctx *cli.Context) error {
				return ra.SyncInterfaces(false)
			},
		}, {
			Name:      "render",
			Aliases:   []string{"r"},
			Usage:     "render template based on the desired state of the tunnels",
			UsageText: "rait render [options] SRC DEST",
			Flags:     commonFlags,
			Before:    commonBeforeFunc,
			Action: func(ctx *cli.Context) error {
				if ctx.Args().Len() != 2 {
					return fmt.Errorf("render expects two arguments, SRC and DEST")
				}
				return ra.RenderTemplate(ctx.Args().Get(0), ctx.Args().Get(1))
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
				Before:    commonBeforeFunc,
				Action: func(context *cli.Context) error {
					links, err := ba.LinkList()
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
				Before:    commonBeforeFunc,
				Action: func(context *cli.Context) error {
					links, err := ra.ListInterfaceName()
					if err != nil {
						return err
					}
					return ba.LinkSync(links)
				},
			}},
		}},
		HideHelpCommand: true,
	}

	if err := app.Run(os.Args); err != nil {
		zap.S().Error(err)
		os.Exit(1)
	}
}
