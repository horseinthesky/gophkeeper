package server

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *Server) checkAuth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if info.FullMethod == "/gophkeeper.GophKeeper/Register" || info.FullMethod == "/gophkeeper.GophKeeper/Login" || info.FullMethod == "/gophkeeper.GophKeeper/Ping" {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.log.Error().Msg("rpc failed due to interceptor failed to retrieve metadata")
		return nil, status.Errorf(codes.Internal, "failed to retrieve metadata")
	}

	authHeader, ok := md["token"]
	if !ok {
		s.log.Error().Msg("rpc failed due to authorization token is missing")
		return nil, status.Errorf(codes.Unauthenticated, "authorization token is missing")
	}

	token := authHeader[0]
	_, err := s.tm.VerifyToken(token)
	if err != nil {
		s.log.Error().Msgf("rpc failed due to %s", err.Error())
		return nil, status.Errorf(codes.Unauthenticated, err.Error())
	}

	return handler(ctx, req)
}
