package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"gophkeeper/logger"
	"gophkeeper/server"
)

func main() {
	configFilePath := flag.String("c", "", "Server config file path")
	flag.Parse()

	config, err := server.LoadConfig(*configFilePath)
	if err != nil {
		panic(err)
	}

	logger := logger.New(config.Environment)

	server, err := server.NewServer(config, logger)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create new server")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	go server.Run(ctx)

	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	sig := <-term
	logger.Info().Msgf("signal received: %v; terminating...\n", sig)

	cancel()
	server.Stop()
}
