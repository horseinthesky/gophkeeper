package client

import (
	"context"
	"database/sql"
	"errors"
	"gophkeeper/converter"
	"gophkeeper/db/db"
	"gophkeeper/pb"
	"time"
)

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
				remoteSecret.Owner.String,
				remoteSecret.Name.String,
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
					remoteSecret.Owner.String,
					remoteSecret.Name.String,
				)
				continue
			}

			c.log.Info().Msgf(
				"successfully synced new user '%s' secret '%s'",
				remoteSecret.Owner.String,
				remoteSecret.Name.String,
			)
			continue
		}

		if remoteSecret.Deleted.Bool {
			err := c.storage.MarkSecretDeleted(
				ctx,
				db.MarkSecretDeletedParams{
					Owner: remoteSecret.Owner,
					Kind:  remoteSecret.Kind,
					Name:  remoteSecret.Name,
				},
			)
			if err != nil {
				c.log.Error().Err(err).Msgf(
					"failed to mark user '%s' secret '%s' as deleted",
					remoteSecret.Owner.String,
					remoteSecret.Name.String,
				)
				continue
			}

			c.log.Info().Msgf(
				"successfully mark user '%s' secret '%s' for deletion",
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
					"failed to update user '%s' secret '%s'",
					remoteSecret.Owner.String,
					remoteSecret.Name.String,
				)
				continue
			}

			c.log.Info().Msgf(
				"successfully synced update of user '%s' secret '%s'",
				remoteSecret.Owner.String,
				remoteSecret.Name.String,
			)
		}
	}

	// Push local
	localSecrets, err := c.storage.GetSecretsByUser(
		ctx,
		sql.NullString{
			String: c.config.User,
			Valid:  true,
		},
	)
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

	c.log.Info().Msg("sync successfull")
}
