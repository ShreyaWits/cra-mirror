package grpc

import (
	"log"
)

type GRPCServerInstance struct {
	server   GRPCServerInterface
	listener GRPCListenerInterface
}

func NewGRPCServer(server GRPCServerInterface, listener GRPCListenerInterface) (*GRPCServerInstance, error) {

	return &GRPCServerInstance{
		server:   server,
		listener: listener,
	}, nil
}

func (s *GRPCServerInstance) Start() error {
	if err := s.server.Serve(s.listener); err != nil {
		return err
	}
	log.Println("gRPC server started on port", s.listener.Addr())
	return nil
}

func (s *GRPCServerInstance) Stop() {
	s.server.GracefulStop()
	log.Println("gRPC server stopped")
}

func (s *GRPCServerInstance) GetServer() GRPCServerInterface {
	return s.server
}
