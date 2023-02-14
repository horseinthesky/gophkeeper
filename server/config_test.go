package server

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	os.Setenv("GOPHKEEPER_ENV", "prod")

	config, err := LoadConfig("testdata/server_config.yml")
	require.NoError(t, err)

	require.Equal(t, config.Environment, "prod")
	require.Equal(t, config.Address, defaultAddress)
}
