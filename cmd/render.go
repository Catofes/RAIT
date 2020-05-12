package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

// renderCmd represents the render command
var renderCmd = &cobra.Command{
	Use:   "render [template] [output]",
	Short: "render the given template",
	Example: "rait render /etc/rait/babeld.template /etc/babeld.conf",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		err := instance.RenderTemplate(args[0], args[1])
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(renderCmd)
}
