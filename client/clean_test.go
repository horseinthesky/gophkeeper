package client

import (
	"context"
	"gophkeeper/db/db"
	"gophkeeper/db/mock"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

func TestClean(t *testing.T) {
	controller := gomock.NewController(t)
	mockStorage := mock.NewMockQuerier(controller)

	mockStorage.EXPECT().
		CleanSecrets(
			gomock.Any(),
		).
		Times(2).
		Return([]db.Secret{}, nil)

	client := &Client{
		config:  Config{Clean: time.Second},
		storage: mockStorage,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
	defer cancel()

	client.cleanJob(ctx)
}
