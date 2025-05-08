package grpc

import (
	"log"
	"net"

	"google.golang.org/grpc"
)

type GRPCServerInstance struct {
	server   *grpc.Server
	listener net.Listener
}

var listenFunc = net.Listen
var newGRPCServerFunc = grpc.NewServer

func NewGRPCServer(port string) *GRPCServerInstance {
	// start the gRPC server
	lis, err := listenFunc("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := newGRPCServerFunc()

	return &GRPCServerInstance{
		server:   s,
		listener: lis,
	}
}

func (s *GRPCServerInstance) Start() error {
	if err := s.server.Serve(s.listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
		return err
	}
	return nil
}

func (s *GRPCServerInstance) Stop() {
	s.server.GracefulStop()
	log.Println("gRPC server stopped")
}

func (s *GRPCServerInstance) GetServer() *grpc.Server {
	return s.server
}

// For testing: allow injection of server and listener
func newGRPCServerWith(server *grpc.Server, listener net.Listener) *GRPCServerInstance {
	return &GRPCServerInstance{
		server:   server,
		listener: listener,
	}
}
