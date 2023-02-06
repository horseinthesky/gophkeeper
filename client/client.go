package client

import (
	"context"
	"database/sql"
	"sync"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"gophkeeper/db/db"
	"gophkeeper/pb"
	"gophkeeper/token"
)

type Client struct {
	config    Config
	storage   *db.Queries
	tm        token.PasetoMaker
	g         pb.GophKeeperClient
	log       zerolog.Logger
	token     string
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
		token.NewPasetoMaker(),
		client,
		logger,
		"",
		sync.WaitGroup{},
	}, nil
}

func (c *Client) Run() {
	c.log.Info().Msg("started gophkeeper client")

	ctx, cancel := context.WithCancel(context.Background())

	c.loadCachedToken()
	if c.token == "" {
		c.login(ctx)
	}

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

	c.runShell(ctx)

	cancel()

	c.workGroup.Wait()
	c.log.Info().Msg("successfully shut down")
}
