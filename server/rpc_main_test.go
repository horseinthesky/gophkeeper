package server

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"gophkeeper/pb"
)

func runTestServer(server *Server, opts ...grpc.ServerOption) (pb.GophKeeperClient, func()) {
	lis := bufconn.Listen(1024)

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterGophKeeperServer(grpcServer, server)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("error serving server: %v", err)
		}
	}()

	conn, err := grpc.DialContext(context.Background(), "",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("error connecting to server: %v", err)
	}

	closer := func() {
		err := lis.Close()
		if err != nil {
			log.Printf("error closing listener: %v", err)
		}

		grpcServer.Stop()
	}

	client := pb.NewGophKeeperClient(conn)

	return client, closer
}
