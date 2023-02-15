package client

import (
	"database/sql"
	"gophkeeper/db/db"
	"gophkeeper/db/mock"
	"os"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestSecretsFromToEntry(t *testing.T) {
	tests := []struct {
		name           string
		secretKind     SecretKind
		inputsLoader   func() []textinput.Model
		payloadBuilder func([]textinput.Model) ([]byte, error)
		cleaner        func() error
	}{
		{
			name:       "test creds secret",
			secretKind: SecretCreds,
			inputsLoader: func() []textinput.Model {
				inputs := newCreds()
				inputs[0].SetValue("testCredsName")
				inputs[1].SetValue("testCredsLogin")
				inputs[2].SetValue("testCredsPassword")
				inputs[3].SetValue("testCredsNotes")

				return inputs
			},
			payloadBuilder: func(inputs []textinput.Model) ([]byte, error) {
				return buildCredsPayload(inputs)
			},
		},
		{
			name:       "test text secret",
			secretKind: SecretText,
			inputsLoader: func() []textinput.Model {
				inputs := newText()
				inputs[0].SetValue("testTextName")
				inputs[1].SetValue("testTextText")
				inputs[2].SetValue("testTextNotes")

				return inputs
			},
			payloadBuilder: func(inputs []textinput.Model) ([]byte, error) {
				return buildTextPayload(inputs)
			},
		},
		{
			name:       "test bytes secret",
			secretKind: SecretBytes,
			inputsLoader: func() []textinput.Model {
				inputs := newBytes()
				inputs[0].SetValue("testFileName")
				inputs[1].SetValue("/tmp/testbytescontent")
				inputs[2].SetValue("testFileNotes")

				return inputs
			},
			payloadBuilder: func(inputs []textinput.Model) ([]byte, error) {
				file, err := os.Create("/tmp/testbytescontent")
				if err != nil {
					return nil, err
				}
				defer file.Close()

				return buildBytesPayload(inputs)
			},
			cleaner: func() error {
				return os.Remove("/tmp/testbytescontent")
			},
		},
		{
			name:       "test card secret",
			secretKind: SecretCard,
			inputsLoader: func() []textinput.Model {
				inputs := newCard()
				inputs[0].SetValue("testCardName")
				inputs[1].SetValue("testCardNumber")
				inputs[2].SetValue("testCardEXP")
				inputs[3].SetValue("testCardCVV")
				inputs[4].SetValue("testCardPIN")
				inputs[5].SetValue("testCardNotes")

				return inputs
			},
			payloadBuilder: func(inputs []textinput.Model) ([]byte, error) {
				return buildCardPayload(inputs)
			},
		},
	}

	testOwner := "owner"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputs := tt.inputsLoader()
			payload, err := tt.payloadBuilder(inputs)
			require.NoError(t, err)

			secret := db.Secret{
				Owner: testOwner,
				Kind:  int32(tt.secretKind),
				Name:  inputs[0].Value(),
				Value: payload,
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
				Return(secret, nil)

			client := Client{
				config:  Config{User: "testOwner"},
				storage: mockStorage,
			}

			mockedSecret, err := client.storeSecretFromEntry(tt.secretKind, inputs)
			require.NoError(t, err)
			require.Equal(t, mockedSecret.Owner, secret.Owner)
			require.Equal(t, mockedSecret.Name, secret.Name)
			require.Equal(t, mockedSecret.Kind, secret.Kind)
			require.Equal(t, mockedSecret.Value, secret.Value)

			secretContent, err := client.loadSecretContentFromEntry(mockedSecret)
			require.NoError(t, err)
			require.NotEmpty(t, secretContent)

			if tt.cleaner != nil {
				err := tt.cleaner()
				require.NoError(t, err)
			}
		})
	}
}
