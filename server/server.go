package server

import (
	"context"
	"database/sql"
	"log"
	"net"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"gophkeeper/db/db"
	"gophkeeper/pb"
)

type Server struct {
	*pb.UnimplementedGophKeeperServer
	config  Config
	storage *db.Queries
	log     zerolog.Logger
}

func NewServer(config Config, logger zerolog.Logger) (*Server, error) {
	pool, err := sql.Open("postgres", config.DSN)
	if err != nil {
		return nil, err
	}

	err = pool.Ping()
	if err != nil {
		return nil, err
	}

	queries := db.New(pool)

	return &Server{
		&pb.UnimplementedGophKeeperServer{},
		config,
		queries,
		logger,
	}, nil
}

func (s *Server) Run() {
	listener, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to open socket")
		return
	}

	go s.cleanJob(context.Background())

	grpcServer := grpc.NewServer()
	pb.RegisterGophKeeperServer(grpcServer, s)
	reflection.Register(grpcServer)

	s.log.Info().Msgf("running gophkeeper server, listening on %s", s.config.Address)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("server crashed: %s", err)
	}

	s.log.Info().Msg("finished to serve gRPC requests")

	grpcServer.GracefulStop()
}
