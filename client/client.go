package client

import (
	"context"
	"database/sql"
	"sync"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"gophkeeper/certs"
	"gophkeeper/db/db"
	migrate "gophkeeper/db"
	"gophkeeper/pb"
	"gophkeeper/token"
)

type Client struct {
	config    Config
	storage   db.Querier
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

	err = migrate.RunDBMigration(cfg.DSN)
	if err != nil {
		return nil, err
	}

	queries := db.New(pool)

	creds, err := certs.LoadClientCreds()
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(cfg.Address, grpc.WithTransportCredentials(creds))
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

	// Try to load token from cache
	err := c.loadCachedToken(tokenCachedDir + tokenCachedFileName)
	if err != nil {
		c.log.Error().Err(err).Msg("failed to load token")
	} else {
		c.log.Info().Msg("successfully loaded cached token")
	}

	// Check server connection and run initial sync
	_, err = c.g.Ping(ctx, &emptypb.Empty{})
	if err != nil {
		c.log.Warn().Msg("server unavailable...working offline")
	} else {
		if c.token == "" {
			c.login(ctx)
		}

		if c.token == "" {
			c.log.Warn().Msg("not authorized...working offline")
		} else {
			c.sync(ctx)
		}
	}

	// Run periodic login job to refresh token
	c.workGroup.Add(1)
	go func() {
		defer c.workGroup.Done()
		c.loginJob(ctx)
	}()

	// Run periodic sync job
	c.workGroup.Add(1)
	go func() {
		defer c.workGroup.Done()
		c.syncJob(ctx)
	}()

	// Run periodic DB clead job
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
