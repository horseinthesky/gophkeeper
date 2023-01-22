package main

import (
	"gophkeeper/server"
)

func main() {
	configPath := "server_config.yml"

	config, err := server.LoadConfig(configPath)
	if err != nil {
		panic(err)
	}

	server := server.NewServer(config)
	server.Run()
}
