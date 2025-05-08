package server

import (
	"fmt"
	"log"
	"net"

	"protected_link/pkg/grpc/proto"

	"google.golang.org/grpc"
)

type GRPCServer struct {
	server      *grpc.Server
	authHandler proto.AuthServiceServer
	linkHandler proto.LinkServiceServer
}

func NewGRPCServer(auth proto.AuthServiceServer, link proto.LinkServiceServer) *GRPCServer {
	return &GRPCServer{
		server:      grpc.NewServer(),
		authHandler: auth,
		linkHandler: link,
	}
}

func (g *GRPCServer) Start(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	proto.RegisterAuthServiceServer(g.server, g.authHandler)
	proto.RegisterLinkServiceServer(g.server, g.linkHandler)

	log.Printf("🚀 gRPC server running on port %s", port)
	return g.server.Serve(lis)
}
