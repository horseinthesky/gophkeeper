package server

import (
	"context"
	"database/sql"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"gophkeeper/converter"
	"gophkeeper/db/db"
	"gophkeeper/pb"
)

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
		if errors.Is(err, sql.ErrNoRows) && remoteSecret.Deleted.Bool {
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
