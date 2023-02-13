package client

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"gophkeeper/db/db"
	"gophkeeper/db/mock"
	"gophkeeper/random"
)

func TestGetSecret(t *testing.T) {
	controller := gomock.NewController(t)
	mockStorage := mock.NewMockQuerier(controller)

	expectedSecret := db.Secret{
		Owner: random.RandomOwner(),
		Kind:  random.RandomSecretKind(),
		Name:  random.RandomString(10),
		Value: []byte(random.RandomString(100)),
	}

	mockStorage.EXPECT().
		GetSecret(
			gomock.Any(),
			gomock.Eq(db.GetSecretParams{
				Owner: expectedSecret.Owner,
				Kind:  expectedSecret.Kind,
				Name:  expectedSecret.Name,
			}),
		).
		Times(1).
		Return(expectedSecret, nil)

	client := Client{
		config: Config{User: expectedSecret.Owner},
		storage: mockStorage,
	}

	secret, err := client.GetSecret(SecretKind(expectedSecret.Kind), expectedSecret.Name)
	require.NoError(t, err)
	require.Equal(t, secret.Value, expectedSecret.Value)

}
