package client

import (
	"database/sql"
	"gophkeeper/db/db"
	"gophkeeper/db/mock"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestSecretsFromToEntry(t *testing.T) {
	testOwner := "owner"

	credsInputs := newCreds()
	credsInputs[0].SetValue("testCredsName")
	credsInputs[1].SetValue("testCredsLogin")
	credsInputs[2].SetValue("testCredsPassword")
	credsInputs[3].SetValue("testCredsNotes")

	credsSecretPayload, err := buildCredsPayload(credsInputs)
	require.NoError(t, err)

	credsSecret := db.Secret{
		Owner: testOwner,
		Kind: int32(SecretCreds),
		Name: credsInputs[0].Value(),
		Value: credsSecretPayload,
	}

	controller := gomock.NewController(t)
	mockStorage := mock.NewMockQuerier(controller)

	mockStorage.EXPECT().
		GetSecret(
			gomock.Any(),
			gomock.Any(),
		).
		Times(1).
		Return(db.Secret{}, sql.ErrNoRows)

	mockStorage.EXPECT().
		CreateSecret(
			gomock.Any(),
			gomock.Any(),
		).
		Times(1).
		Return(credsSecret, nil)

	client := Client{
		config:  Config{User: "testOwner"},
		storage: mockStorage,
	}

	dbSecret, err := client.storeSecretFromEntry(SecretCreds, credsInputs)
	require.NoError(t, err)
	require.Equal(t, dbSecret.Owner, credsSecret.Owner)

	secretContent, err := client.loadSecretContentFromEntry(dbSecret)
	require.NoError(t, err)
	require.NotEmpty(t, secretContent)
}
