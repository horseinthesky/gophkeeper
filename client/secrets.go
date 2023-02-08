package client

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gophkeeper/db/db"
)

type SecretKind int32

const (
	SecretCreds SecretKind = iota
	SecretText
	SecretBytes
	SecretCard
)

var secretKindToString = map[SecretKind]string{
	SecretCreds: "Creds",
	SecretText:  "Text",
	SecretBytes: "Bytes",
	SecretCard:  "Card",
}

var stringToSecretKind = map[string]SecretKind{
	"Creds": SecretCreds,
	"Text":  SecretText,
	"Bytes": SecretBytes,
	"Card":  SecretCard,
}

func (k SecretKind) String() string {
	return secretKindToString[k]
}

func (c *Client) SetSecret(ctx context.Context, kind SecretKind, name string, payload []byte) (db.Secret, error) {
	localSecret, err := c.storage.GetSecret(
		ctx,
		db.GetSecretParams{
			Owner: c.config.User,
			Kind:  int32(kind),
			Name:  name,
		},
	)
	if errors.Is(err, sql.ErrNoRows) {
		newSecret, err := c.storage.CreateSecret(
			ctx,
			db.CreateSecretParams{
				Owner: c.config.User,
				Kind:  int32(kind),
				Name:  name,
				Value: payload,
				Created: sql.NullTime{
					Time:  time.Now(),
					Valid: true,
				},
				Modified: sql.NullTime{
					Time:  time.Now(),
					Valid: true,
				},
			},
		)
		if err != nil {
			c.log.Error().Err(err).Msgf("failed to save user '%s' new secret '%s'", c.config.User, name)
			return db.Secret{}, err
		}

		c.log.Info().Msgf("successfully created user '%s' new secret '%s'", c.config.User, name)
		return newSecret, nil
	}
	if err != nil {
		c.log.Error().Err(err).Msgf("failed got user '%s' local secret '%s'", c.config.User, name)
		return db.Secret{}, err
	}

	updateSecret, err := c.storage.UpdateSecret(
		ctx,
		db.UpdateSecretParams{
			Owner: c.config.User,
			Kind:  int32(kind),
			Name:  name,
			Value: payload,
			Created: sql.NullTime{
				Time:  localSecret.Created.Time,
				Valid: true,
			},
			Modified: sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			},
		},
	)

	c.log.Info().Msgf("successfully updated user '%s' secret '%s'", c.config.User, name)

	return updateSecret, nil
}

func (c *Client) GetSecret(ctx context.Context, kind SecretKind, name string) (db.Secret, error) {
	return c.storage.GetSecret(
		ctx,
		db.GetSecretParams{
			Owner: c.config.User,
			Kind:  int32(kind),
			Name:  name,
		},
	)
}

// func (c *Client) ListSecrets(ctx context.Context) ([]db.Secret, error) {
// 	return c.storage.GetSecretsByUser(ctx, c.config.User)
// }

func (c *Client) DeleteSecret(ctx context.Context, kind SecretKind, name string) error {
	return c.storage.MarkSecretDeleted(
		ctx,
		db.MarkSecretDeletedParams{
			Owner: c.config.User,
			Kind:  int32(kind),
			Name:  name,
		},
	)
}
