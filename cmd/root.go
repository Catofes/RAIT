package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/NickCao/RAIT/rait"
	"log"
)

var instance = rait.DefaultInstance()

var RootCmd = &cobra.Command{
	Use:   "rait",
	Short: "rait - redundant array of inexpensive tunnels",
}

func init() {
	var configFile string
	cobra.OnInitialize(func() {
		viper.SetConfigFile(configFile)
		viper.SetConfigType("toml")
		err := viper.ReadInConfig()
		if err != nil {
			log.Fatal("Failed to load config file:", viper.ConfigFileUsed())
		}
		err = viper.Unmarshal(instance)
		if err != nil {
			log.Fatal("Failed to unmarshal config", err)
		}
	})
	RootCmd.PersistentFlags().StringVarP(&configFile, "config","c", "/etc/rait/rait.conf", "path to rait.conf")
}
