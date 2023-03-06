package main

import (
	"flag"
	"fmt"

	"gophkeeper/client"
	"gophkeeper/logger"
)

var (
	buildTime string
	version   string = "0.0.1"
)

func main() {
	v := flag.Bool("v", false, "Gophkeeper version")
	configFilePath := flag.String("c", "", "Client config file path")
	flag.Parse()

	if *v {
		fmt.Printf("Gophkeeper client\n\nVersion: %s\nBuild date: %s\n", version, buildTime)
		return
	}

	config, err := client.LoadConfig(*configFilePath)
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
