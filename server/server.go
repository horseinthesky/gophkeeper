package server

import (
	"context"
	"database/sql"
	"log"
	"net"
	"sync"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"gophkeeper/certs"
	"gophkeeper/db/db"
	"gophkeeper/pb"
	"gophkeeper/token"
)

type Server struct {
	*pb.UnimplementedGophKeeperServer
	config  Config
	storage db.Querier
	tm      token.PasetoMaker
	log     zerolog.Logger
	wg      sync.WaitGroup
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
		token.NewPasetoMaker(),
		logger,
		sync.WaitGroup{},
	}, nil
}

func (s *Server) Run(ctx context.Context) {
	listener, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to open socket")
		return
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.cleanJob(ctx)
	}()

	creds, err := certs.LoadServerCreds()
	if err != nil {
		log.Fatalf("failed to run server: %s", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(s.checkAuth),
		grpc.Creds(creds),
	)
	pb.RegisterGophKeeperServer(grpcServer, s)
	reflection.Register(grpcServer)

	s.log.Info().Msgf("running gophkeeper server, listening on %s", s.config.Address)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("server crashed: %s", err)
	}

	s.log.Info().Msg("finished to serve gRPC requests")

	grpcServer.GracefulStop()
}

func (s *Server) Stop() {
	s.log.Info().Msg("shutting down...")

	s.wg.Wait()
	s.log.Info().Msg("successfully shut down")
}
