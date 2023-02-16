package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGCM(t *testing.T) {
	text := []byte("answer to ultimate question of life the universe and everything")
	key := []byte("the-key-has-to-be-32-bytes-long!")

	ciphertext, err := Encrypt(text, key)
	require.NoError(t, err)

	plaintext, err := Decrypt(ciphertext, key)
	require.NoError(t, err)
	require.Equal(t, plaintext, text)
}
