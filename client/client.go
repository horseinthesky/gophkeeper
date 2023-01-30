package client

import (
	"context"
	"database/sql"
	"gophkeeper/db/db"
	"gophkeeper/pb"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	config    Config
	storage   *db.Queries
	g         pb.GophKeeperClient
	log       zerolog.Logger
	workGroup sync.WaitGroup
}

func NewClient(cfg Config, logger zerolog.Logger) (*Client, error) {
	pool, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, err
	}

	err = pool.Ping()
	if err != nil {
		return nil, err
	}

	queries := db.New(pool)

	conn, err := grpc.Dial(cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := pb.NewGophKeeperClient(conn)

	return &Client{
		cfg,
		queries,
		client,
		logger,
		sync.WaitGroup{},
	}, nil
}

func (c *Client) Run() {
	ctx, cancel := context.WithCancel(context.Background())

	c.workGroup.Add(1)
	go func() {
		defer c.workGroup.Done()
		c.syncJob(ctx)
	}()

	c.workGroup.Add(1)
	go func() {
		defer c.workGroup.Done()
		c.cleanJob(ctx)
	}()

	c.workGroup.Add(1)
	go func() {
		defer c.workGroup.Done()
		c.Shell()
	}()

	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	sig := <-term
	c.log.Info().Msgf("signal received: %v; terminating...\n", sig)

	cancel()

	c.workGroup.Wait()
	c.log.Info().Msg("successfully shut down")
}
