package client

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"

	"gophkeeper/crypto"
	"gophkeeper/db/db"
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

func buildCredsPayload(inputs []textinput.Model) ([]byte, error) {
	secretPayload := CredsPayload{
		Login:    inputs[1].Value(),
		Password: inputs[2].Value(),
		Notes:    inputs[3].Value(),
	}

	return json.Marshal(secretPayload)
}

func buildTextPayload(inputs []textinput.Model) ([]byte, error) {
	secretPayload := TextPayload{
		Text:  inputs[1].Value(),
		Notes: inputs[2].Value(),
	}

	return json.Marshal(secretPayload)
}

func buildBytesPayload(inputs []textinput.Model) ([]byte, error) {
	filePath := inputs[1].Value()

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	if fileInfo.Size() > fileSizeLimit {
		return nil, fmt.Errorf(
			"file %s is too big, pls use files less than %v bytes",
			filePath,
			fileSizeLimit,
		)
	}

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	secretPayload := BytesPayload{
		Filename: fileInfo.Name(),
		Bytes:    fileBytes,
		Notes:    inputs[2].Value(),
	}

	return json.Marshal(secretPayload)
}

func buildCardPayload(inputs []textinput.Model) ([]byte, error) {
	secretPayload := CardPayload{
		Number: inputs[1].Value(),
		Owner:  inputs[2].Value(),
		EXP:    inputs[3].Value(),
		CVV:    inputs[4].Value(),
		PIN:    inputs[5].Value(),
		Notes:  inputs[6].Value(),
	}

	return json.Marshal(secretPayload)
}

var payloaderMap = map[SecretKind]func([]textinput.Model) ([]byte, error){
	SecretCreds: buildCredsPayload,
	SecretText:  buildTextPayload,
	SecretBytes: buildBytesPayload,
	SecretCard:  buildCardPayload,
}

func (c *Client) storeSecretFromEntry(kind SecretKind, inputs []textinput.Model) (db.Secret, error) {
	secretName := inputs[0].Value()

	payloader, ok := payloaderMap[kind]
	if !ok {
		return db.Secret{}, fmt.Errorf("unsupported secret kind: %s", secretKindToString[kind])
	}

	payloadBytes, err := payloader(inputs)
	if err != nil {
		return db.Secret{}, fmt.Errorf("failed to build secret '%s' payload: %w", secretName, err)
	}

	if c.config.Encrypt {
		payloadBytes, err = crypto.Encrypt(payloadBytes, []byte(c.config.Key))
		if err != nil {
			return db.Secret{}, fmt.Errorf("failed to encrypt secret '%s' payload: %w", secretName, err)
		}
	}

	dbSecret, err := c.SetSecret(kind, secretName, payloadBytes)
	if err != nil {
		return db.Secret{}, err
	}

	return dbSecret, nil
}

func (c *Client) loadSecretContentFromEntry(secret db.Secret) (string, error) {
	switch SecretKind(secret.Kind) {
	case SecretCreds:
		var secretPayload CredsPayload
		err := json.Unmarshal(secret.Value, &secretPayload)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(
			` Secret: %s
 Created: %s
 Modified: %s

 Login: %s
 Password: %s
 Notes: %s
`,
			secret.Name,
			secret.Created,
			secret.Modified,
			secretPayload.Login,
			secretPayload.Password,
			secretPayload.Notes,
		), nil
	case SecretText:
		var secretPayload TextPayload
		err := json.Unmarshal(secret.Value, &secretPayload)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(
			` Secret: %s
 Created: %s
 Modified: %s

 Text: %s
 Notes: %s
`,
			secret.Name,
			secret.Created,
			secret.Modified,
			secretPayload.Text,
			secretPayload.Notes,
		), nil
	case SecretBytes:
		var secretPayload BytesPayload
		err := json.Unmarshal(secret.Value, &secretPayload)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(
			` Secret: %s
 Created: %s
 Modified: %s

 Filename: %s
 Notes: %s

 Press "s" to save the file to your local drive.
`,
			secret.Name,
			secret.Created,
			secret.Modified,
			secretPayload.Filename,
			secretPayload.Notes,
		), nil
	case SecretCard:
		var secretPayload CardPayload
		err := json.Unmarshal(secret.Value, &secretPayload)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(
			` Secret: %s
 Created: %s
 Modified: %s

 Number: %s
 Owner: %s
 EXP: %s
 CVV: %s
 PIN: %s
 Notes: %s
`,
			secret.Name,
			secret.Created,
			secret.Modified,
			secretPayload.Number,
			secretPayload.Owner,
			secretPayload.EXP,
			secretPayload.CVV,
			secretPayload.PIN,
			secretPayload.Notes,
		), nil
	}

	return "", fmt.Errorf("unsupported secret kind: %s", secretKindToString[SecretKind(secret.Kind)])
}
