package client

import (
	"context"
	"database/sql"
	"gophkeeper/converter"
	"gophkeeper/db/db"
	"gophkeeper/pb"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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

	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	sig := <-term
	c.log.Info().Msgf("signal received: %v; terminating...\n", sig)

	cancel()

	c.workGroup.Wait()
	c.log.Info().Msg("successfully shut down")
}

func (c *Client) syncJob(ctx context.Context) {
	ticker := time.NewTicker(c.config.Sync)

	c.log.Info().Msg("started periodic syncing")

	for {
		select {
		case <-ctx.Done():
			c.log.Info().Msg("periodic syncing stopped")
			return
		case <-ticker.C:
			c.sync(ctx)
		}
	}
}

func (c *Client) sync(ctx context.Context) {
	pbSecrets, err := c.g.GetSecrets(ctx, &pb.SecretsRequest{
		Owner: c.config.User,
	})
	if err != nil {
		c.log.Error().Err(err).Msg("sync failed")
		return
	}

	c.log.Info().Msgf("sync got %v secrets", len(pbSecrets.Secrets))

	for _, pbSecret := range pbSecrets.Secrets {
		secret := converter.PBtoSecret(pbSecret)
		if pbSecret.Deleted {
			err := c.storage.DeleteSecret(
				ctx,
				db.DeleteSecretParams{
					Owner: secret.Owner,
					Kind:  secret.Kind,
					Name:  secret.Name,
				},
			)
			if err != nil {
				c.log.Error().Err(err).Msgf("failed to delete secret: %s", secret.Name)
			}

			continue
		}

		_, err := c.storage.CreateSecret(
			ctx,
			db.CreateSecretParams{
				Owner:    secret.Owner,
				Kind:     secret.Kind,
				Name:     secret.Name,
				Value:    secret.Value,
				Created:  secret.Created,
				Modified: secret.Modified,
			},
		)
		if err != nil {
			c.log.Error().Err(err).Msgf("failed to create secret: %s", secret.Name)
		}
	}
	c.log.Info().Msg("sync successfull")
}
