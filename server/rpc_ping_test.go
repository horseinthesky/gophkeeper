package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestRPCPing(t *testing.T) {
	testServer := &Server{}

	client, closer := runTestServer(testServer)
	defer closer()

	_, err := client.Ping(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
}
