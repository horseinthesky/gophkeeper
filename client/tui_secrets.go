package client

import (
	"context"
	"encoding/json"
	"fmt"
	"gophkeeper/db/db"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
)

const fileSizeLimit = 500000 // bytes

type (
	CredsPayload struct {
		Login    string `json:"login"`
		Password string `json:"password"`
		Notes    string `json:"notes"`
	}

	TextPayload struct {
		Text  string `json:"text"`
		Notes string `json:"notes"`
	}

	BytesPayload struct {
		Filename string `json:"file"`
		Bytes    []byte `json:"bytes"`
		Notes    string `json:"notes"`
	}

	CardPayload struct {
		Number string `json:"number"`
		Owner  string `json:"owner"`
		EXP    string `json:"exp"`
		CVV    string `json:"cvv"`
		PIN    string `json:"pin"`
		Notes  string `json:"notes"`
	}
)

func (c *Client) storeSecretFromEntry(kind SecretKind, inputs []textinput.Model) (db.Secret, error) {
	secretName := inputs[0].Value()

	var payloadBytes []byte

	switch kind {
	case SecretCreds:
		secretPayload := CredsPayload{
			Login:    inputs[1].Value(),
			Password: inputs[2].Value(),
			Notes:    inputs[3].Value(),
		}

		payloadBytes, _ = json.Marshal(secretPayload)
	case SecretText:
		secretPayload := TextPayload{
			Text:  inputs[1].Value(),
			Notes: inputs[2].Value(),
		}

		payloadBytes, _ = json.Marshal(secretPayload)
	case SecretBytes:
		filePath := inputs[1].Value()
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return db.Secret{}, err
		}
		if fileInfo.Size() > fileSizeLimit {
			return db.Secret{}, fmt.Errorf(
				"file %s is too big, pls use files less than %v bytes",
				filePath,
				fileSizeLimit,
			)
		}

		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			return db.Secret{}, err
		}
		secretPayload := BytesPayload{
			Filename: fileInfo.Name(),
			Bytes:    fileBytes,
			Notes:    inputs[2].Value(),
		}

		payloadBytes, _ = json.Marshal(secretPayload)
	case SecretCard:
		secretPayload := CardPayload{
			Number: inputs[1].Value(),
			Owner:  inputs[2].Value(),
			EXP:    inputs[3].Value(),
			CVV:    inputs[4].Value(),
			PIN:    inputs[5].Value(),
			Notes:  inputs[6].Value(),
		}

		payloadBytes, _ = json.Marshal(secretPayload)
	}

	dbSecret, err := c.SetSecret(context.Background(), kind, secretName, payloadBytes)
	if err != nil {
		return db.Secret{}, err
	}

	return dbSecret, nil
}
