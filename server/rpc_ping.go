package server

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) Ping(ctx context.Context, in *empty.Empty) (*empty.Empty, error) {
	return &emptypb.Empty{}, nil
}
