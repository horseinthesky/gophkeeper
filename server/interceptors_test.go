package server

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"gophkeeper/db/db"
	"gophkeeper/db/mock"
	"gophkeeper/pb"
	"gophkeeper/random"
	"gophkeeper/token"
)

var testUsername3 = random.RandomOwner()

func TestCheckAuth(t *testing.T) {
	// Create mock storage
	controller := gomock.NewController(t)
	mockStorage := mock.NewMockQuerier(controller)

	mockStorage.EXPECT().
		GetSecretsByUser(
			gomock.Any(),
			gomock.Any(),
		).
		Times(1).
		Return([]db.Secret{}, nil)

	// Create server
	testServer := &Server{
		config:  Config{},
		storage: mockStorage,
		tm:      token.NewPasetoMaker(),
	}

	// Run test gRPC server
	client, closer := runTestServer(testServer, grpc.UnaryInterceptor(testServer.checkAuth))
	defer closer()

	// Generate token
	tm := token.NewPasetoMaker()
	token, err := tm.CreateToken(testUsername3, time.Hour)
	require.NoError(t, err)

	// Add token to metadata
	md := metadata.Pairs("token", token)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// Run rpc
	pbSecrets, err := client.GetSecrets(ctx, &pb.SecretsRequest{Owner: testUsername3})
	require.NoError(t, err)
	require.Equal(t, len(pbSecrets.Secrets), 0)
}

func TestCheckAuthMissingToken(t *testing.T) {
	// Create mock storage
	controller := gomock.NewController(t)
	mockStorage := mock.NewMockQuerier(controller)

	mockStorage.EXPECT().
		GetSecretsByUser(
			gomock.Any(),
			gomock.Any(),
		).
		Times(0).
		Return([]db.Secret{}, nil)

	// Create server
	testServer := &Server{
		config:  Config{},
		storage: mockStorage,
		tm:      token.NewPasetoMaker(),
	}

	// Run test gRPC server
	client, closer := runTestServer(testServer, grpc.UnaryInterceptor(testServer.checkAuth))
	defer closer()

	// Run rpc with no token
	_, err := client.GetSecrets(context.Background(), &pb.SecretsRequest{Owner: testUsername3})
	require.Error(t, err)
	e, _ := status.FromError(err)
	require.Equal(t, codes.Unauthenticated, e.Code())
	require.Equal(t, e.Message(), "authorization token is missing")

	// Add token to metadata
	md := metadata.Pairs("token", "someinvalidtoken")
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// Run rpc with invalid token
	_, err = client.GetSecrets(ctx, &pb.SecretsRequest{Owner: testUsername3})
	require.Error(t, err)
	e, _ = status.FromError(err)
	require.Equal(t, codes.Unauthenticated, e.Code())
	require.Equal(t, e.Message(), "token is invalid")
}
