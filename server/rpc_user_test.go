package server

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestRPCRegister(t *testing.T) {
	testServer, _ := NewServer(Config{}, zerolog.Logger{})

	_, closer := runTestServer(testServer)
	defer closer()
}
