package cmd

import (
	"github.com/kochman/runner"
	"github.com/spf13/cobra"

	"github.com/kochman/hotshots/config"
	"github.com/kochman/hotshots/log"
	"github.com/kochman/hotshots/server"
)

func init() {
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the Hotshots server",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Hotshots server starting...")
		log.SetLevel("debug")

		config, err := config.New()
		if err != nil {
			log.WithError(err).Error("Unable to create config")
			return
		}
		runner := runner.New()

		server, err := server.New(config)
		if err != nil {
			log.WithError(err).Error("Unable to create server")
			return
		}
		runner.Add(server)

		runner.Run()
	},
}
