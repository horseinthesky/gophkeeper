package main

import (
	"flag"
	"fmt"
	"gophkeeper/client"
	"gophkeeper/logger"
)

var (
	buildTime string
	version string = "0.0.1"
)

func main() {
	b := flag.Bool("b", false, "Build date")
	v := flag.Bool("v", false, "Gophkeeper version")
	flag.Parse()

	if *b {
		fmt.Printf("Gophkeeper client build date: %s\n", buildTime)
		return
	}

	if *v {
		fmt.Printf("Gophkeeper client version: %s\n", version)
		return
	}

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
