package client

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gophkeeper/token"
)

const (
	testUser = "testUser"
)

func TestLoadCachedToken(t *testing.T) {
	client := Client{
		config: Config{User: testUser},
		tm:     token.NewPasetoMaker(),
	}

	// Test load token from missing file
	err := client.loadCachedToken("/tmp/doesnotexist")
	require.Error(t, err)

	// Test load token from empty file
	emptyFile, err := os.CreateTemp("", "empty")
	require.NoError(t, err)
	defer os.Remove(emptyFile.Name())

	err = client.loadCachedToken(emptyFile.Name())
	require.Error(t, err)
	require.Equal(t, err, token.ErrInvalidToken)

	// Test load expired token
	tm := token.NewPasetoMaker()
	expiredToken, err := tm.CreateToken(testUser, -time.Minute)

	expiredTokenFile, err := os.CreateTemp("", "testToken")
	require.NoError(t, err, "failed to create temp file for test token")
	defer os.Remove(expiredTokenFile.Name())

	_, err = expiredTokenFile.Write([]byte(expiredToken))
	require.NoError(t, err, "failed to write test token to temp file")

	err = client.loadCachedToken(expiredTokenFile.Name())
	require.Error(t, err)
	require.Equal(t, err, token.ErrExpiredToken)

	// Test load valid token
	testToken, err := tm.CreateToken(testUser, time.Minute)

	validTokenFile, err := os.CreateTemp("", "testToken")
	require.NoError(t, err, "failed to create temp file for test token")
	defer os.Remove(validTokenFile.Name())

	_, err = validTokenFile.Write([]byte(testToken))
	require.NoError(t, err, "failed to write test token to temp file")

	err = client.loadCachedToken(validTokenFile.Name())
	require.NoError(t, err)
}

func TestSaveToken(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tokenTmpDir")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	client := Client{}
	err = client.saveToken(tmpDir, "tmpToken", "testToken")
}
