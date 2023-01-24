package server

import (
	"context"

	"gophkeeper/pb"
)

func (s *Server) Register(ctx context.Context, in *pb.User) (*pb.Token, error) {
	return &pb.Token{Value: "bla"}, nil
}
