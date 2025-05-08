package grpc

import (
	"log"
	"net"

	"google.golang.org/grpc"
)

type GRPCServerInstance struct {
	server   *grpc.Server
	listener GRPCListenerInterface
}

func NewGRPCServer(port string) (*GRPCServerInstance, error) {
	// start the gRPC server
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return nil, err
	}
	s := grpc.NewServer()

	return &GRPCServerInstance{
		server:   s,
		listener: lis,
	}, nil
}

func (s *GRPCServerInstance) Start() error {
	if err := s.server.Serve(s.listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
		return err
	}
	log.Println("gRPC server started on port", s.listener.Addr())
	return nil
}

func (s *GRPCServerInstance) Stop() {
	s.server.GracefulStop()
	log.Println("gRPC server stopped")
}

func (s *GRPCServerInstance) GetServer() *grpc.Server {
	return s.server
}
