package cmd

import (
	"github.com/spf13/cobra"

	"github.com/kochman/hotshots/config"
	"github.com/kochman/hotshots/log"
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

		_, err := config.New()
		if err != nil {
			log.WithError(err).Error("unable to create config")
			return
		}
	},
}
