package main

import (
	"gophkeeper/logger"
	"gophkeeper/server"
)

func main() {
	configPath := "server_config.yml"

	config, err := server.LoadConfig(configPath)
	if err != nil {
		panic(err)
	}

	logger := logger.New(config.Environment)

	server, err := server.NewServer(config, logger)
	if err != nil {
		logger.Err(err)
	}

	server.Run()
}
