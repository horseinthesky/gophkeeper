package server

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gophkeeper/db/db"
	"gophkeeper/db/mock"
	"gophkeeper/pb"
	"gophkeeper/random"
	"gophkeeper/server/crypto"
	"gophkeeper/token"
)

var (
	testUsername = random.RandomOwner()
	testPassword = random.RandomString(20)
)

func TestRPCRegister(t *testing.T) {
	// Create mock storage
	controller := gomock.NewController(t)
	mockStorage := mock.NewMockQuerier(controller)

	mockStorage.EXPECT().
		CreateUser(
			gomock.Any(),
			gomock.Any(),
		).
		Times(1).
		Return(db.User{}, nil)

	// Create server
	testServer := &Server{
		config:  Config{},
		storage: mockStorage,
		tm:      token.NewPasetoMaker(),
	}

	// Run test gRPC server
	client, closer := runTestServer(testServer)
	defer closer()

	// Test valid username
	pbToken, err := client.Register(context.Background(), &pb.User{Name: testUsername, Password: testPassword})
	require.NoError(t, err)

	payload, err := testServer.tm.VerifyToken(pbToken.Value)
	require.NoError(t, err)

	require.Equal(t, payload.Username, testUsername)

	// Test invalid (too short) username and password
	pbToken, err = client.Register(context.Background(), &pb.User{Name: "bo", Password: "bla"})
	require.Error(t, err)

	e, _ := status.FromError(err)
	require.Equal(t, codes.InvalidArgument, e.Code())
}

func TestRPCLogin(t *testing.T) {
	testUserPasshash, err := crypto.HashPassword(testPassword)
	require.NoError(t, err)

	// Create mock storage
	controller := gomock.NewController(t)
	mockStorage := mock.NewMockQuerier(controller)

	mockStorage.EXPECT().
		GetUser(
			gomock.Any(),
			gomock.Any(),
		).
		Times(1).
		Return(db.User{Name: testUsername, Passhash: testUserPasshash}, nil)

	// Create server
	testServer := &Server{
		config:  Config{},
		storage: mockStorage,
		tm:      token.NewPasetoMaker(),
	}

	// Run test gRPC server
	client, closer := runTestServer(testServer)
	defer closer()

	// Test valid username
	pbToken, err := client.Login(context.Background(), &pb.User{Name: testUsername, Password: testPassword})
	require.NoError(t, err)

	payload, err := testServer.tm.VerifyToken(pbToken.Value)
	require.NoError(t, err)

	require.Equal(t, payload.Username, testUsername)

	// Test invalid (too short) username and password
	pbToken, err = client.Login(context.Background(), &pb.User{Name: "bo", Password: "bla"})
	require.Error(t, err)

	e, _ := status.FromError(err)
	require.Equal(t, codes.InvalidArgument, e.Code())
}
