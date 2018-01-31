package cmd

import (
	"github.com/kochman/runner"

	"github.com/kochman/hotshots/config"
	"github.com/kochman/hotshots/log"
	"github.com/kochman/hotshots/server"
)

func Run() {
	log.Info("Hotshots starting...")
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
}
