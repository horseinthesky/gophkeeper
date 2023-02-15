package client

import (
	"database/sql"
	"os"
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
		config:  Config{User: expectedSecret.Owner},
		storage: mockStorage,
	}

	secret, err := client.GetSecret(SecretKind(expectedSecret.Kind), expectedSecret.Name)
	require.NoError(t, err)
	require.Equal(t, secret.Value, expectedSecret.Value)
}

func TestSetNewSecret(t *testing.T) {
	controller := gomock.NewController(t)
	mockStorage := mock.NewMockQuerier(controller)

	testOwner := random.RandomOwner()

	newSecret := db.Secret{
		Owner: testOwner,
		Kind:  random.RandomSecretKind(),
		Name:  random.RandomString(10),
		Value: []byte(random.RandomString(100)),
	}

	mockStorage.EXPECT().
		GetSecret(
			gomock.Any(),
			db.GetSecretParams{
				Owner: newSecret.Owner,
				Kind:  newSecret.Kind,
				Name:  newSecret.Name,
			},
		).
		Times(1).
		Return(db.Secret{}, sql.ErrNoRows)

	mockStorage.EXPECT().
		CreateSecret(
			gomock.Any(),
			gomock.Any(),
		).
		Times(1).
		Return(newSecret, nil)

	client := Client{
		config:  Config{User: testOwner},
		storage: mockStorage,
	}

	secret, err := client.SetSecret(SecretKind(newSecret.Kind), newSecret.Name, newSecret.Value)
	require.NoError(t, err)
	require.Equal(t, secret.Value, newSecret.Value)
}

func TestSetExistingSecret(t *testing.T) {
	controller := gomock.NewController(t)
	mockStorage := mock.NewMockQuerier(controller)

	testOwner := random.RandomOwner()

	existingSecret := db.Secret{
		Owner: testOwner,
		Kind:  random.RandomSecretKind(),
		Name:  random.RandomString(10),
		Value: []byte(random.RandomString(100)),
	}

	mockStorage.EXPECT().
		GetSecret(
			gomock.Any(),
			db.GetSecretParams{
				Owner: existingSecret.Owner,
				Kind:  existingSecret.Kind,
				Name:  existingSecret.Name,
			},
		).
		Times(1).
		Return(existingSecret, nil)

	mockStorage.EXPECT().
		UpdateSecret(
			gomock.Any(),
			gomock.Any(),
		).
		Times(1).
		Return(existingSecret, nil)

	client := Client{
		config:  Config{User: testOwner},
		storage: mockStorage,
	}

	secret, err := client.SetSecret(SecretKind(existingSecret.Kind), existingSecret.Name, existingSecret.Value)
	require.NoError(t, err)
	require.Equal(t, secret.Value, existingSecret.Value)
}

func TestDeleteSecret(t *testing.T) {
	controller := gomock.NewController(t)
	mockStorage := mock.NewMockQuerier(controller)

	mockStorage.EXPECT().
		MarkSecretDeleted(
			gomock.Any(),
			gomock.Any(),
		).
		Times(1).
		Return(nil)

	client := Client{
		config:  Config{},
		storage: mockStorage,
	}

	err := client.DeleteSecret(SecretKind(0), "somename")
	require.NoError(t, err)
}

func TestSaveOnDisk(t *testing.T) {
	err := saveOnDisk("/tmp/savesecret", []byte("somevalue"))
	require.NoError(t, err)

	err = os.Remove("/tmp/savesecret")
	require.NoError(t, err)
}
