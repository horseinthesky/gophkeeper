package server

import (
	"context"
	"gophkeeper/db/db"
	"gophkeeper/db/mock"
	"gophkeeper/pb"
	"gophkeeper/random"
	"gophkeeper/token"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
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

	token, err := testServer.tm.CreateToken(testUsername, defaultTokenDuration)
	require.NoError(t, err)

	pbToken, err := client.Register(context.Background(), &pb.User{Name: testUsername, Password: testPassword})
	require.NoError(t, err)
	require.Equal(t, pbToken.Value, token)

}
