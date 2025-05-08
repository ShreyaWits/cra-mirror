package grpc

import (
	"fmt"
	"log"
	"net"

	pb "Document-Processing/proto"

	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	port       int
}

func NewServer(port int) *Server {
	return &Server{
		grpcServer: grpc.NewServer(),
		port:       port,
	}
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	log.Printf("Starting gRPC server on port %d", s.port)
	return s.grpcServer.Serve(lis)
}

func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
}

func (s *Server) RegisterServices(documentService pb.DocumentProcessingServiceServer, healthHandler pb.HealthServiceServer) {
	// Register document processing service
	pb.RegisterDocumentProcessingServiceServer(s.grpcServer, documentService)

	// Register health check service
	pb.RegisterHealthServiceServer(s.grpcServer, healthHandler)
}
