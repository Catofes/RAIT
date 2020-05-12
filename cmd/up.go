package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "bring up wireguard links",
	Run: func(cmd *cobra.Command, args []string) {
		err := instance.SyncInterfaces(true)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(upCmd)
}
