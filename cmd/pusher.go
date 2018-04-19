package cmd

import (
	"github.com/kochman/runner"
	"github.com/spf13/cobra"

	"github.com/kochman/hotshots/config"
	"github.com/kochman/hotshots/log"
	"github.com/kochman/hotshots/pusher"
)

func init() {
	rootCmd.AddCommand(pusherCmd)
}

var pusherCmd = &cobra.Command{
	Use:   "pusher",
	Short: "Run the Hotshots pusher",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Hotshots pusher starting...")
		log.SetLevel("debug")

		config, err := config.New()
		if err != nil {
			log.WithError(err).Error("unable to create config")
			return
		}

		pusher, err := pusher.New(config)
		if err != nil {
			log.WithError(err).Error("Unable to create pusher")
			return
		}

		runner := runner.New()
		runner.Add(pusher)
		runner.Run()
	},
}
