package grpc

import (
	"fmt"
	"net"
	"nps-reciept-service/internal/services"
	"nps-reciept-service/internal/utils"
	pb "nps-reciept-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GrpcServer represents the gRPC server
type GrpcServer struct {
	server *grpc.Server
	port   int
}

// NewGrpcServer creates a new gRPC server instance
func NewGrpcServer(port int) *GrpcServer {
	return &GrpcServer{
		server: grpc.NewServer(),
		port:   port,
	}
}

// Start initializes and starts the gRPC server
func (s *GrpcServer) Start() error {
	// Create TCP listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	// Register services
	claimServer := services.NewClaimServer()
	pb.RegisterClaimServiceServer(s.server, claimServer)

	// Enable reflection for tools like Postman
	reflection.Register(s.server)

	utils.LogInfo("Starting gRPC server", map[string]interface{}{
		"port": s.port,
	})

	// Start serving
	if err := s.server.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server
func (s *GrpcServer) Stop() {
	if s.server != nil {
		utils.LogInfo("Stopping gRPC server", nil)
		s.server.GracefulStop()
	}
}
