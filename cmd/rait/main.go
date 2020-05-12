package main

import (
	"gitlab.com/NickCao/RAIT/cmd"
	"os"
)

var Version string

func main() {
	cmd.RootCmd.Version = Version
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
