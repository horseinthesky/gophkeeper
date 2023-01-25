package server

import (
	"context"
	"database/sql"
	"gophkeeper/converter"
	"gophkeeper/db/db"
	"gophkeeper/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) SetSecret(ctx context.Context, in *pb.Secret) (*emptypb.Empty, error) {
	s.log.Info().Msgf("user %s sent his %s secret %s", in.Owner, in.Kind, in.Name)

	_, err := s.storage.CreateSecret(
		ctx,
		db.CreateSecretParams{
			Owner: sql.NullString{
				String: in.Owner,
				Valid:  true,
			},
			Kind: sql.NullInt32{
				Int32: in.Kind,
				Valid: true,
			},
			Name: sql.NullString{
				String: in.Name,
				Valid:  true,
			},
			Value: in.Value,
			Created: sql.NullTime{
				Time:  in.Created.AsTime(),
				Valid: true,
			},
			Modified: sql.NullTime{
				Time:  in.Modified.AsTime(),
				Valid: true,
			},
		},
	)
	if err != nil {
		s.log.Error().Err(err).Msgf("failed to save user %s %s secret %s", in.Owner, in.Kind, in.Name)
		return nil, status.Error(codes.Internal, "failed to save secret to db")
	}

	s.log.Info().Msgf("successfully saved user %s %s secret %s", in.Owner, in.Kind, in.Name)

	return &emptypb.Empty{}, nil
}

func (s *Server) GetSecret(ctx context.Context, in *pb.SecretRequest) (*pb.Secret, error) {
	s.log.Info().Msgf("user %s requested his %s secret %s", in.Owner, in.Kind, in.Name)

	secret, err := s.storage.GetSecret(
		ctx,
		db.GetSecretParams{
			Owner: sql.NullString{
				String: in.Owner,
				Valid:  true,
			},
			Kind: sql.NullInt32{
				Int32: in.Kind,
				Valid: true,
			},
			Name: sql.NullString{
				String: in.Name,
				Valid:  true,
			},
		},
	)
	if err != nil {
		s.log.Error().Err(err).Msgf("failed to get user %s %s secret %s", in.Owner, in.Kind, in.Name)
		return nil, status.Error(codes.Internal, "failed to get secret from db")
	}

	s.log.Info().Msgf("successfully got user %s %s secret %s", in.Owner, in.Kind, in.Name)

	return converter.DBSecretToPBSecret(secret), nil
}

func (s *Server) GetSecrets(ctx context.Context, in *pb.SecretsRequest) (*pb.Secrets, error) {
	s.log.Info().Msgf("user %s requested his secrets", in.Owner)

	secrets, err := s.storage.GetSecretsByUser(
		ctx,
		sql.NullString{
			String: in.Owner,
			Valid:  true,
		},
	)
	if err != nil {
		s.log.Error().Err(err).Msgf("failed to get user %s secrets", in.Owner)
		return nil, status.Error(codes.Internal, "failed to get secrets from db")
	}

	s.log.Info().Msgf("successfully got user %s secrets", in.Owner)

	pbSecrets := []*pb.Secret{}
	for _, secret := range secrets {
		pbSecrets = append(pbSecrets, converter.DBSecretToPBSecret(secret))
	}

	return &pb.Secrets{Secrets: pbSecrets}, nil
}
