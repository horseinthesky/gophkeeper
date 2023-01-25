package client

import (
	"context"
	"database/sql"
	"errors"
	"gophkeeper/converter"
	"gophkeeper/db/db"
	"gophkeeper/pb"
	"os"
	"os/signal"
	"reflect"
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
	c.sync(ctx)

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
	remotePBSecrets, err := c.g.GetSecrets(ctx, &pb.SecretsRequest{
		Owner: c.config.User,
	})
	if err != nil {
		c.log.Error().Err(err).Msgf("failed to pull user %s remote secrets", c.config.User)
		return
	}

	c.log.Info().Msgf("sync got %v secrets", len(remotePBSecrets.Secrets))

	for _, pbSecret := range remotePBSecrets.Secrets {
		// Pull remote
		remoteSecret := converter.PBSecretToDBSecret(pbSecret)

		localSecret, err := c.storage.GetSecret(
			ctx,
			db.GetSecretParams{
				Owner: remoteSecret.Owner,
				Kind:  remoteSecret.Kind,
				Name:  remoteSecret.Name,
			},
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) && !remoteSecret.Deleted.Bool {
				_, err := c.storage.CreateSecret(
					ctx,
					db.CreateSecretParams{
						Owner:    remoteSecret.Owner,
						Kind:     remoteSecret.Kind,
						Name:     remoteSecret.Name,
						Value:    remoteSecret.Value,
						Created:  remoteSecret.Created,
						Modified: remoteSecret.Modified,
					},
				)
				if err != nil {
					c.log.Error().Err(err).Msgf(
						"failed to sync new user %s secret %s",
						remoteSecret.Owner.String,
						remoteSecret.Name.String,
					)
					continue
				}

				c.log.Info().Msgf(
					"successfully synced new user %s secret %s",
					remoteSecret.Owner.String,
					remoteSecret.Name.String,
				)
				continue
			}

			c.log.Error().Err(err).Msgf(
				"failed to get user %s secret %s from local db",
				remoteSecret.Owner.String,
				remoteSecret.Name.String,
			)
			continue
		}

		if remoteSecret.Deleted.Bool {
			err := c.storage.DeleteSecret(
				ctx,
				db.DeleteSecretParams{
					Owner: remoteSecret.Owner,
					Kind:  remoteSecret.Kind,
					Name:  remoteSecret.Name,
				},
			)
			if err != nil {
				c.log.Error().Err(err).Msgf(
					"failed to delete user %s secret %s",
					remoteSecret.Owner.String,
					remoteSecret.Name.String,
				)
				continue
			}

			c.log.Info().Msgf(
				"successfully synced deletion of user %s secret %s",
				remoteSecret.Owner.String,
				remoteSecret.Name.String,
			)
			continue
		}

		if remoteSecret.Modified.Time.After(localSecret.Modified.Time) {
			_, err := c.storage.UpdateSecret(
				ctx,
				db.UpdateSecretParams{
					Owner:    remoteSecret.Owner,
					Kind:     remoteSecret.Kind,
					Name:     remoteSecret.Name,
					Value:    remoteSecret.Value,
					Created:  remoteSecret.Created,
					Modified: remoteSecret.Modified,
				},
			)
			if err != nil {
				c.log.Error().Err(err).Msgf(
					"failed to update user %s secret %s",
					remoteSecret.Owner.String,
					remoteSecret.Name.String,
				)
				continue
			}

			c.log.Info().Msgf(
				"successfully synced update of user %s secret %s",
				remoteSecret.Owner.String,
				remoteSecret.Name.String,
			)
		}
	}

	// Push local
	// localSecrets, err := c.storage.GetSecretsByUser(
	// 	ctx,
	// 	sql.NullString{
	// 		String: c.config.User,
	// 		Valid:  true,
	// 	},
	// )
	// if err != nil {
	// 	c.log.Error().Err(err).Msgf(
	// 		"failed to get user %s local secrets",
	// 		c.config.User,
	// 	)
	// 	return
	// }
	//
	// localPBSecrets := []*pb.Secret{}
	// for _, secret := range localSecrets {
	// 	localPBSecrets = append(localPBSecrets, converter.DBSecretToPBSecret(secret))
	// }
	//
	// _, err = c.g.SetSecrets(ctx, &pb.Secrets{Secrets: localPBSecrets})
	// if err != nil {
	// 	c.log.Error().Err(err).Msgf(
	// 		"failed to push user %s local secrets",
	// 		c.config.User,
	// 	)
	// 	return
	// }

	c.log.Info().Msg("sync successfull")
}
