package main

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/urfave/cli/v2"
	"gitlab.com/NickCao/RAIT/v3/pkg/misc"
	"gitlab.com/NickCao/RAIT/v3/pkg/rait"
	"go.uber.org/zap"
	"os"
	"strings"
)

var Version string
var r *rait.RAIT

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
	&cli.BoolFlag{
		Name:    "bind",
		Usage:   "enable wireguard bind support",
		Aliases: []string{"b"},
		Value:   false,
	},
}

var commonBeforeFunc = func(ctx *cli.Context) error {
	misc.Bind = ctx.Bool("bind")

	config := zap.NewDevelopmentConfig()
	config.DisableStacktrace = true
	switch ctx.Bool("debug") {
	case true:
		config.Level.SetLevel(zap.DebugLevel)
	case false:
		config.Level.SetLevel(zap.InfoLevel)
	}

	logger, err := config.Build()
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(logger)

	r, err = rait.NewRAIT(ctx.String("config"))
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
				return r.Sync(true)
			},
		}, {
			Name:      "down",
			Aliases:   []string{"d"},
			Usage:     "destroy the tunnels",
			UsageText: "rait down [options]",
			Flags:     commonFlags,
			Before:    commonBeforeFunc,
			Action: func(ctx *cli.Context) error {
				return r.Sync(false)
			},
		}, {
			Name:      "list",
			Aliases:   []string{"l"},
			Usage:     "list the tunnels",
			UsageText: "rait list [options]",
			Flags:     commonFlags,
			Before:    commonBeforeFunc,
			Action: func(ctx *cli.Context) error {
				list, err := r.List()
				if err != nil {
					return err
				}
				_, err = fmt.Println(strings.Join(misc.LinkString(list), " "))
				return err
			},
		}, {
			Name:      "pub",
			Aliases:   []string{"p"},
			Usage:     "generate public metadata",
			UsageText: "rait pub [options] DEST",
			Flags:     commonFlags,
			Before:    commonBeforeFunc,
			Action: func(ctx *cli.Context) error {
				if ctx.Args().Len() != 1 {
					return fmt.Errorf("expecting 1 argument: DEST")
				}
				return r.PublicConf(ctx.Args().First())
			},
		}, {
			Name:      "remarks",
			Aliases:   []string{"r"},
			Usage:     "query remarks from rait.conf",
			UsageText: "rait query [options] QUERY",
			Flags:     commonFlags,
			Before:    commonBeforeFunc,
			Action: func(ctx *cli.Context) error {
				if ctx.Args().Len() != 1 {
					return fmt.Errorf("expecting 1 argument: QUERY")
				}
				bc, _ := r.Remarks.Content(&hcl.BodySchema{Blocks: []hcl.BlockHeaderSchema{{Type: "remarks"}}})
				if len(bc.Blocks.OfType("remarks")) == 0 {
					return nil
				}
				attrs, _ := bc.Blocks.OfType("remarks")[0].Body.JustAttributes()
				attr, _ := attrs[ctx.Args().First()]
				if attr == nil {
					return nil
				}
				value, _ := attr.Expr.Value(nil)
				fmt.Println(value.AsString())
				return nil
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
				Flags:     commonFlags,
				Before:    commonBeforeFunc,
				Action: func(context *cli.Context) error {
					links, err := r.Babeld.LinkList()
					if err != nil {
						return err
					}
					_, err = fmt.Println(strings.Join(links, " "))
					return err
				},
			}, {
				Name:      "sync",
				Aliases:   []string{"s"},
				Usage:     "sync babeld interfaces",
				UsageText: "rait babeld sync [options]",
				Flags:     commonFlags,
				Before:    commonBeforeFunc,
				Action: func(context *cli.Context) error {
					links, err := r.List()
					if err != nil {
						return err
					}
					return r.Babeld.LinkSync(misc.LinkString(links))
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
