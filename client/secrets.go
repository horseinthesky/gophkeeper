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
	SecretBankCard
)

var secretKindToString = map[SecretKind]string{
	SecretCreds:    "Creds",
	SecretText:     "Text",
	SecretBytes:    "Bytes",
	SecretBankCard: "Card",
}

var stringToSecretKind = map[string]SecretKind{
	"Creds": SecretCreds,
	"Text":  SecretText,
	"Bytes": SecretBytes,
	"Card":  SecretBankCard,
}

func (k SecretKind) String() string {
	return secretKindToString[k]
}

type (
	SecretPayload struct {
		Notes string `json:"notes"`
	}

	CredsPayload struct {
		SecretPayload
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	TextPayload struct {
		SecretPayload
		Text string `json:"text"`
	}

	BytesPayload struct {
		SecretPayload
		File string `json:"file"`
	}

	CardPayload struct {
		SecretPayload
		Number  string `json:"number"`
		Owner   string `json:"owner"`
		Expires string `json:"expires"`
		CVV     string `json:"cvv"`
		PIN     string `json:"pin"`
	}
)

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

	c.log.Info().Msgf("successfully updated user '%s' secret '%s'", c.config.User, name)

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

func (c *Client) ListSecrets(ctx context.Context) ([]db.Secret, error) {
	return c.storage.GetSecretsByUser(
		ctx,
		sql.NullString{
			String: c.config.User,
			Valid:  true,
		},
	)
}

func (c *Client) DeleteSecret(ctx context.Context, kind SecretKind, name string) error {
	return c.storage.MarkSecretDeleted(
		ctx,
		db.MarkSecretDeletedParams{
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
