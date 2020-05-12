package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

// downCmd represents the down command
var downCmd = &cobra.Command{
	Use:   "down",
	Short: "remove links created by up",
	Run: func(cmd *cobra.Command, args []string) {
		err := instance.SyncInterfaces(false)
		if err != nil{
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(downCmd)
}
