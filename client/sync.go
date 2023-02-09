package client

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	"gophkeeper/converter"
	"gophkeeper/db/db"
	"gophkeeper/pb"
)

func (c *Client) syncJob(ctx context.Context) {
	ticker := time.NewTicker(c.config.Sync)

	c.log.Info().Msg("started periodic syncing")

	for {
		select {
		case <-ctx.Done():
			c.log.Info().Msg("periodic syncing stopped")
			return
		case <-ticker.C:
			_, err := c.g.Ping(ctx, &emptypb.Empty{})
			if err != nil {
				c.log.Warn().Msg("server unavailable...working offline")
				continue
			}

			if c.token == "" {
				c.log.Warn().Msg("not authorized...working offline")
				continue
			}

			c.sync(ctx)
			c.log.Info().Msg("sync job successfull")
		}
	}
}

func (c *Client) sync(ctx context.Context) {
	c.log.Info().Msg("secrets sync started...")

	// Provide token
	md := metadata.Pairs("token", c.token)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Pull remote
	remotePBSecrets, err := c.g.GetSecrets(ctx, &pb.SecretsRequest{
		Owner: c.config.User,
	})
	if err != nil {
		c.log.Error().Err(err).Msgf("failed to pull user '%s' remote secrets", c.config.User)
		return
	}

	c.log.Info().Msgf("sync got %v secrets", len(remotePBSecrets.Secrets))

	for _, pbSecret := range remotePBSecrets.Secrets {
		remoteSecret := converter.PBSecretToDBSecret(pbSecret)

		localSecret, err := c.storage.GetSecret(
			ctx,
			db.GetSecretParams{
				Owner: remoteSecret.Owner,
				Kind:  remoteSecret.Kind,
				Name:  remoteSecret.Name,
			},
		)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			c.log.Error().Err(err).Msgf(
				"failed to get user '%s' secret '%s' from local db",
				remoteSecret.Owner,
				remoteSecret.Name,
			)
			continue
		}
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
					"failed to sync new user '%s' secret '%s'",
					remoteSecret.Owner,
					remoteSecret.Name,
				)
				continue
			}

			c.log.Info().Msgf(
				"successfully synced new user '%s' secret '%s'",
				remoteSecret.Owner,
				remoteSecret.Name,
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
					"failed to delete user '%s' secret '%s'",
					remoteSecret.Owner,
					remoteSecret.Name,
				)
				continue
			}

			c.log.Info().Msgf(
				"successfully deleted user '%s' secret '%s'",
				remoteSecret.Owner,
				remoteSecret.Name,
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
					"failed to update user '%s' secret '%s'",
					remoteSecret.Owner,
					remoteSecret.Name,
				)
				continue
			}

			c.log.Info().Msgf(
				"successfully synced update of user '%s' secret '%s'",
				remoteSecret.Owner,
				remoteSecret.Name,
			)
		}
	}

	// Push local
	localSecrets, err := c.storage.GetSecretsByUser(ctx, c.config.User)
	if err != nil {
		c.log.Error().Err(err).Msgf(
			"failed to get user '%s' local secrets",
			c.config.User,
		)
		return
	}

	localPBSecrets := []*pb.Secret{}
	for _, secret := range localSecrets {
		localPBSecrets = append(localPBSecrets, converter.DBSecretToPBSecret(secret))
	}

	_, err = c.g.SetSecrets(ctx, &pb.Secrets{Secrets: localPBSecrets})
	if err != nil {
		c.log.Error().Err(err).Msgf(
			"failed to push user '%s' local secrets",
			c.config.User,
		)
		return
	}

	c.log.Info().Msg("secrets sync finished")
}
