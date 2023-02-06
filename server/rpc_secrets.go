package server

import (
	"context"
	"database/sql"
	"errors"
	"gophkeeper/converter"
	"gophkeeper/db/db"
	"gophkeeper/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) SetSecret(ctx context.Context, in *pb.Secret) (*emptypb.Empty, error) {
	s.log.Info().Msgf("user '%s' sent his '%s' secret '%s'", in.Owner, in.Kind, in.Name)

	_, err := s.storage.CreateSecret(
		ctx,
		db.CreateSecretParams{
			Owner: in.Owner,
			Kind: in.Kind,
			Name: in.Name,
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
		s.log.Error().Err(err).Msgf("failed to save user '%s' '%s' secret '%s'", in.Owner, in.Kind, in.Name)
		return nil, status.Error(codes.Internal, "failed to save secret to db")
	}

	s.log.Info().Msgf("successfully saved user '%s' '%s' secret '%s'", in.Owner, in.Kind, in.Name)

	return &emptypb.Empty{}, nil
}

func (s *Server) GetSecret(ctx context.Context, in *pb.SecretRequest) (*pb.Secret, error) {
	s.log.Info().Msgf("user '%s' requested his '%s' secret '%s'", in.Owner, in.Kind, in.Name)

	secret, err := s.storage.GetSecret(
		ctx,
		db.GetSecretParams{
			Owner: in.Owner,
			Kind: in.Kind,
			Name: in.Name,
		},
	)
	if err != nil {
		s.log.Error().Err(err).Msgf("failed to get user '%s' '%s' secret '%s'", in.Owner, in.Kind, in.Name)
		return nil, status.Error(codes.Internal, "failed to get secret from db")
	}

	s.log.Info().Msgf("successfully got user '%s' '%s' secret '%s'", in.Owner, in.Kind, in.Name)

	return converter.DBSecretToPBSecret(secret), nil
}

func (s *Server) SetSecrets(ctx context.Context, in *pb.Secrets) (*emptypb.Empty, error) {
	for _, pbSecret := range in.Secrets {
		remoteSecret := converter.PBSecretToDBSecret(pbSecret)

		localSecret, err := s.storage.GetSecret(
			ctx,
			db.GetSecretParams{
				Owner: remoteSecret.Owner,
				Kind:  remoteSecret.Kind,
				Name:  remoteSecret.Name,
			},
		)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			s.log.Error().Err(err).Msgf(
				"failed to get user '%s' secret '%s' from local db",
				remoteSecret.Owner,
				remoteSecret.Name,
			)
			continue
		}
		if errors.Is(err, sql.ErrNoRows) && !remoteSecret.Deleted.Bool {
			_, err := s.storage.CreateSecret(
				ctx,
				db.CreateSecretParams{
					Owner:    remoteSecret.Owner,
					Kind:     remoteSecret.Kind,
					Name:     remoteSecret.Name,
					Value:    remoteSecret.Value,
					Created:  remoteSecret.Created,
					Modified: remoteSecret.Modified,
				},
			)
			if err != nil {
				s.log.Error().Err(err).Msgf(
					"failed to sync new user '%s' secret '%s'",
					remoteSecret.Owner,
					remoteSecret.Name,
				)
				continue
			}

			s.log.Info().Msgf(
				"successfully synced new user '%s' secret '%s'",
				remoteSecret.Owner,
				remoteSecret.Name,
			)
			continue
		}

		if remoteSecret.Deleted.Bool {
			err := s.storage.MarkSecretDeleted(
				ctx,
				db.MarkSecretDeletedParams{
					Owner: remoteSecret.Owner,
					Kind:  remoteSecret.Kind,
					Name:  remoteSecret.Name,
				},
			)
			if err != nil {
				s.log.Error().Err(err).Msgf(
					"failed to mark user '%s' secret '%s' as deleted",
					remoteSecret.Owner,
					remoteSecret.Name,
				)
				continue
			}

			s.log.Info().Msgf(
				"successfully marked user '%s' secret '%s' for deletion",
				remoteSecret.Owner,
				remoteSecret.Name,
			)
			continue
		}

		if remoteSecret.Modified.Time.After(localSecret.Modified.Time) {
			_, err := s.storage.UpdateSecret(
				ctx,
				db.UpdateSecretParams{
					Owner:    remoteSecret.Owner,
					Kind:     remoteSecret.Kind,
					Name:     remoteSecret.Name,
					Value:    remoteSecret.Value,
					Created:  remoteSecret.Created,
					Modified: remoteSecret.Modified,
				},
			)
			if err != nil {
				s.log.Error().Err(err).Msgf(
					"failed to update user '%s' secret '%s'",
					remoteSecret.Owner,
					remoteSecret.Name,
				)
				continue
			}

			s.log.Info().Msgf(
				"successfully synced update of user '%s' secret '%s'",
				remoteSecret.Owner,
				remoteSecret.Name,
			)
		}
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) GetSecrets(ctx context.Context, in *pb.SecretsRequest) (*pb.Secrets, error) {
	s.log.Info().Msgf("user '%s' requested his secrets", in.Owner)

	secrets, err := s.storage.GetSecretsByUser(ctx, in.Owner)
	if err != nil {
		s.log.Error().Err(err).Msgf("failed to get user %s secrets", in.Owner)
		return nil, status.Error(codes.Internal, "failed to get secrets from db")
	}

	s.log.Info().Msgf("successfully got user '%s' secrets from db", in.Owner)

	pbSecrets := []*pb.Secret{}
	for _, secret := range secrets {
		pbSecrets = append(pbSecrets, converter.DBSecretToPBSecret(secret))
	}

	return &pb.Secrets{Secrets: pbSecrets}, nil
}
