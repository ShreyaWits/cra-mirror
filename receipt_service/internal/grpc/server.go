package grpc

import (
	"fmt"
	"net"
	"nps-reciept-service/internal/config"
	"nps-reciept-service/internal/middleware"
	"nps-reciept-service/internal/repositories"
	"nps-reciept-service/internal/services"
	"nps-reciept-service/internal/utils"
	"nps-reciept-service/pkg/observability"
	"nps-reciept-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GrpcServer struct {
	server        *grpc.Server
	port          int
	Observability *observability.ObservabilityStack
}

// NewGrpcServer creates a new gRPC server instance
func NewGrpcServer(port int, obs *observability.ObservabilityStack) *GrpcServer {
	return &GrpcServer{
		server:        grpc.NewServer(),
		port:          port,
		Observability: obs,
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
	redisClient := config.GetRedisClient()
	claimRepo := repositories.NewRedisClaimRepository(redisClient)
	claimServer := services.NewClaimServer(claimRepo, s.Observability)

	// Create a new server with the interceptor
	serverWithInterceptor := grpc.NewServer(grpc.UnaryInterceptor(middleware.ValidateClaimRequestInterceptor()))
	proto.RegisterClaimServiceServer(serverWithInterceptor, claimServer)

	s.server = serverWithInterceptor

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
