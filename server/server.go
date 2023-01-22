package server

import (
	"gophkeeper/pb"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	*pb.UnimplementedGophKeeperServer
	config Config
}

func NewServer(config Config) *Server {
	return &Server{
		&pb.UnimplementedGophKeeperServer{},
		config,
	}
}

func (s *Server) Run() {
	listener, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterGophKeeperServer(grpcServer, s)
	reflection.Register(grpcServer)

	log.Printf("running gophkeeper server, listening on %s", s.config.Address)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("server crashed: %s", err)
	}

	log.Printf("finished to serve gRPC requests")

	grpcServer.GracefulStop()
}
