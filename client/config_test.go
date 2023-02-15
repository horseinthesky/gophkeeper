package client

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig("testdata/client_config.yml")
	require.Error(t, err)
	require.Equal(t, "user password cannot be empty", err.Error())

	os.Setenv("GOPHKEEPER_ENV", "dev")
	os.Setenv("GOPHKEEPER_PASSWORD", "password")

	config, err = LoadConfig("testdata/client_config.yml")
	require.NoError(t, err)
	require.Equal(t, config.Environment, "dev")
	require.Equal(t, config.Address, defaultAddress)
	require.Equal(t, config.User, "someguy")
	require.Equal(t, config.Password, "password")
}
