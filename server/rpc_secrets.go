package server

import (
	"context"
	"database/sql"
	"gophkeeper/converter"
	"gophkeeper/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetSecrets(ctx context.Context, in *pb.SecretsRequest) (*pb.Secrets, error) {
	s.log.Info().Msgf("user %s requested his secrets", in.Owner)

	secrets, err := s.storage.GetSecretsByUser(
		ctx,
		sql.NullString{
			String: in.Owner,
			Valid: true,
		},
	)
	if err != nil {
		s.log.Error().Err(err).Msgf("user %s failed to get his secrets", in.Owner)
		return nil, status.Error(codes.Internal, "failed to get secrets from db")
	}

	s.log.Info().Msgf("user %s successfully got his secrets", in.Owner)

	pbSecrets := []*pb.Secret{}
	for _, secret := range secrets {
		pbSecrets = append(pbSecrets, converter.SecretToPB(secret))
	}

	return &pb.Secrets{Secrets: pbSecrets}, nil
}
