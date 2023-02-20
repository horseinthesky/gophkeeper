package server

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"gophkeeper/db/db"
	"gophkeeper/db/mock"
	"gophkeeper/pb"
	"gophkeeper/random"
)

var (
	testUsername2 = random.RandomOwner()
)

func TestRPCGetSecrets(t *testing.T) {
	// Create mock storage
	controller := gomock.NewController(t)
	mockStorage := mock.NewMockQuerier(controller)

	mockStorage.EXPECT().
		GetSecretsByUser(
			gomock.Any(),
			testUsername2,
		).
		Times(1).
		Return([]db.Secret{{Owner: testUsername2, Kind: 0, Name: "bla"}}, nil)

	// Create server
	testServer := &Server{
		config:  Config{},
		storage: mockStorage,
	}

	// Run test gRPC server
	client, closer := runTestServer(testServer)
	defer closer()

	pbSecrets, err := client.GetSecrets(context.Background(), &pb.SecretsRequest{Owner: testUsername2})
	require.NoError(t, err)
	require.Equal(t, len(pbSecrets.Secrets), 1)
}

func TestRPCSetSecrets(t *testing.T) {
	// Create mock storage
	controller := gomock.NewController(t)
	mockStorage := mock.NewMockQuerier(controller)

	// Mock secret to create
	mockStorage.EXPECT().
		GetSecret(
			gomock.Any(),
			db.GetSecretParams{
				Owner: testUsername2,
				Kind:  0,
				Name:  "testSecretToCreate",
			},
		).
		Times(1).
		Return(
			db.Secret{},
			sql.ErrNoRows,
		)

	mockStorage.EXPECT().
		CreateSecret(
			gomock.Any(),
			gomock.Any(),
		).
		Times(1).
		Return(
			db.Secret{},
			nil,
		)

	// Mock secret to delete
	mockStorage.EXPECT().
		GetSecret(
			gomock.Any(),
			db.GetSecretParams{
				Owner: testUsername2,
				Kind:  0,
				Name:  "testSecretToDelete",
			},
		).
		Times(1).
		Return(
			db.Secret{
				Owner: testUsername2,
				Kind:  0,
				Name:  "testSecretToDelete",
			},
			nil,
		)

	mockStorage.EXPECT().
		MarkSecretDeleted(
			gomock.Any(),
			gomock.Any(),
		).
		Times(1).
		Return(
			nil,
		)

	// Mock secret to update
	mockStorage.EXPECT().
		GetSecret(
			gomock.Any(),
			db.GetSecretParams{
				Owner: testUsername2,
				Kind:  0,
				Name:  "testSecretToUpdate",
			},
		).
		Times(1).
		Return(
			db.Secret{
				Owner:    testUsername2,
				Kind:     0,
				Name:     "testSecretToUpdate",
				Modified: time.Now().Add(-time.Minute),
			},
			nil,
		)

	mockStorage.EXPECT().
		UpdateSecret(
			gomock.Any(),
			gomock.Any(),
		).
		Times(1).
		Return(
			db.Secret{},
			nil,
		)

	// Create server
	testServer := &Server{
		config:  Config{},
		storage: mockStorage,
	}

	// Run test gRPC server
	client, closer := runTestServer(testServer)
	defer closer()

	// Test create secret
	_, err := client.SetSecrets(
		context.Background(),
		&pb.Secrets{
			Secrets: []*pb.Secret{
				{
					Owner: testUsername2,
					Kind:  0,
					Name:  "testSecretToCreate",
				},
			},
		},
	)
	require.NoError(t, err)

	// Test delete secret
	_, err = client.SetSecrets(
		context.Background(),
		&pb.Secrets{
			Secrets: []*pb.Secret{
				{
					Owner:   testUsername2,
					Kind:    0,
					Name:    "testSecretToDelete",
					Deleted: true,
				},
			},
		},
	)
	require.NoError(t, err)

	// Test update secret
	_, err = client.SetSecrets(
		context.Background(),
		&pb.Secrets{
			Secrets: []*pb.Secret{
				{
					Owner:    testUsername2,
					Kind:     0,
					Name:     "testSecretToUpdate",
					Modified: timestamppb.Now(),
				},
			},
		},
	)
	require.NoError(t, err)
}
