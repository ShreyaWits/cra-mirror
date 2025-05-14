package grpc

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "Document-Processing/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	grpcServer *grpc.Server
	port       int
}

func NewServer(port int) *Server {
	// Create a new gRPC server with default options
	server := grpc.NewServer()

	// Enable reflection for tools like grpcurl
	reflection.Register(server)

	return &Server{
		grpcServer: server,
		port:       port,
	}
}

func (s *Server) Start() error {
	// Create TCP listener
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Attempting to listen on %s", addr)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", addr, err)
	}

	// Set up graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Received shutdown signal")
		s.Stop()
	}()

	log.Printf("Starting gRPC server on %s", addr)

	// Start serving
	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve on %s: %v", addr, err)
	}

	return nil
}

func (s *Server) Stop() {
	if s.grpcServer != nil {
		log.Println("Stopping gRPC server gracefully")
		s.grpcServer.GracefulStop()
	}
}

func (s *Server) RegisterServices(documentService pb.DocumentProcessingServiceV1Server, healthHandler pb.HealthServiceServer) {
	// Register document processing service
	pb.RegisterDocumentProcessingServiceV1Server(s.grpcServer, documentService)
	
	// Log server info
	log.Printf("Server is ready to accept connections on port %d", s.port)

}
