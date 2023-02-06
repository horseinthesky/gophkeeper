package main

import (
	"gophkeeper/logger"
	"gophkeeper/client"
)

func main() {
	configPath := "client_config.yml"

	config, err := client.LoadConfig(configPath)
	if err != nil {
		panic(err)
	}

	logger, err := logger.NewFileLogger(config.Environment, "gophkeeper.log")
	if err != nil {
		panic(err)
	}

	client, err := client.NewClient(config, logger)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create new client")
		return
	}

	client.Run()
}
