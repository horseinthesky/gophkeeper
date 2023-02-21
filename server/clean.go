package server

import (
	"context"
	"time"
)

func (s *Server) cleanJob(ctx context.Context) {
	ticker := time.NewTicker(s.config.Clean)

	s.log.Info().Msg("started periodic db cleaning")

	for {
		select {
		case <-ctx.Done():
			s.log.Info().Msg("finished periodic db cleaning")
			return
		case <-ticker.C:
			s.clean(ctx)
		}
	}
}

func (s *Server) clean(ctx context.Context) {
	deletedSecrets, err := s.storage.CleanSecrets(ctx)
	if err != nil {
		s.log.Error().Msg("failed to clean deleted secrets")
		return
	}

	s.log.Info().Msgf("cleaned up %v deleted secrets", len(deletedSecrets))
}
