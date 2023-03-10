package client

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"gophkeeper/crypto"
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

func (c *Client) GetSecret(kind SecretKind, name string) (db.Secret, error) {
	dbSecret, err := c.storage.GetSecret(
		context.Background(),
		db.GetSecretParams{
			Owner: c.config.User,
			Kind:  int32(kind),
			Name:  name,
		},
	)
	if err != nil {
		c.log.Error().Err(err).Msgf("failed to get secret '%s' from db", dbSecret.Name)
		return db.Secret{}, err
	}

	if c.config.Encrypt {
		decryptedPayload, err := crypto.Decrypt(dbSecret.Value, []byte(c.config.Key))
		if err != nil {
			c.log.Error().Err(err).Msgf("failed to decrypt secret '%s' payload", dbSecret.Name)
			return db.Secret{}, fmt.Errorf("failed to decrypt secret '%s' payload: %w", dbSecret.Name, err)
		}

		dbSecret.Value = decryptedPayload
		return dbSecret, nil
	}

	return dbSecret, nil
}

func (c *Client) SetSecret(kind SecretKind, name string, payload []byte) (db.Secret, error) {
	localSecret, err := c.storage.GetSecret(
		context.Background(),
		db.GetSecretParams{
			Owner: c.config.User,
			Kind:  int32(kind),
			Name:  name,
		},
	)
	if errors.Is(err, sql.ErrNoRows) {
		newSecret, err := c.storage.CreateSecret(
			context.Background(),
			db.CreateSecretParams{
				Owner: c.config.User,
				Kind:  int32(kind),
				Name:  name,
				Value: payload,
				Created: time.Now(),
				Modified: time.Now(),
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
		context.Background(),
		db.UpdateSecretParams{
			Owner: c.config.User,
			Kind:  int32(kind),
			Name:  name,
			Value: payload,
			Created: localSecret.Created,
			Modified: time.Now(),
		},
	)

	c.log.Info().Msgf("successfully updated user '%s' secret '%s'", c.config.User, name)

	return updateSecret, nil
}

func (c *Client) DeleteSecret(kind SecretKind, name string) error {
	return c.storage.MarkSecretDeleted(
		context.Background(),
		db.MarkSecretDeletedParams{
			Owner: c.config.User,
			Kind:  int32(kind),
			Name:  name,
		},
	)
}

func saveOnDisk(filename string, content []byte) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(content)

	return nil
}
