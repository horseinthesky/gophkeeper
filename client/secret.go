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
	Creds SecretKind = iota
	Text
	Bytes
	BankCard
)

var secretStrings = [...]string{
	"Creds",
	"Text",
	"Bytes",
	"BankCard",
}

func (k SecretKind) String() string {
	return secretStrings[k]
}

func (c *Client) SetSecret(ctx context.Context, kind SecretKind, name string, payload []byte) (db.Secret, error) {
	localSecret, err := c.storage.GetSecret(
		ctx,
		db.GetSecretParams{
			Owner: sql.NullString{
				String: c.config.User,
				Valid:  true,
			},
			Kind: sql.NullInt32{
				Int32: int32(kind),
				Valid: true,
			},
			Name: sql.NullString{
				String: name,
				Valid:  true,
			},
		},
	)
	if errors.Is(err, sql.ErrNoRows) {
		newSecret, err := c.storage.CreateSecret(
			ctx,
			db.CreateSecretParams{
				Owner: sql.NullString{
					String: c.config.User,
					Valid:  true,
				},
				Kind: sql.NullInt32{
					Int32: int32(kind),
					Valid: true,
				},
				Name: sql.NullString{
					String: name,
					Valid:  true,
				},
				Value: payload,
			},
		)
		if err != nil {
			c.log.Error().Err(err).Msgf("failed to save user '%s' new secret '%s'", c.config.User, name)
			return db.Secret{}, err
		}

		return newSecret, nil
	}
	if err != nil {
		c.log.Error().Err(err).Msgf("failed got user '%s' local secret '%s'", c.config.User, name)
		return db.Secret{}, err
	}

	updateSecret, err := c.storage.UpdateSecret(
		ctx,
		db.UpdateSecretParams{
			Owner: sql.NullString{
				String: c.config.User,
				Valid:  true,
			},
			Kind: sql.NullInt32{
				Int32: int32(kind),
				Valid: true,
			},
			Name: sql.NullString{
				String: name,
				Valid:  true,
			},
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

	return updateSecret, nil
}

func (c *Client) GetSecret(ctx context.Context, kind SecretKind, name string) (db.Secret, error) {
	return c.storage.GetSecret(
		ctx,
		db.GetSecretParams{
			Owner: sql.NullString{
				String: c.config.User,
				Valid:  true,
			},
			Kind: sql.NullInt32{
				Int32: int32(kind),
				Valid: true,
			},
			Name: sql.NullString{
				String: name,
				Valid:  true,
			},
		},
	)
}
