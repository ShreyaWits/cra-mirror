package main

import (
	"context"
	pb "cra-protos/messaging_service"
	"log"
	"messaging_service/internal/di"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	shutdownTimeout = 5 * time.Second
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	// Initialize dependency container
	container, err := di.NewContainer()
	if err != nil {
		sugar.Fatalw("Failed to initialize container", "error", err)
	}

	// Set up a TCP listener
	listener, err := net.Listen("tcp", ":"+container.Config.GrpcPort)
	if err != nil {
		sugar.Fatalw("Failed to listen", "error", err, "port", container.Config.GrpcPort)
	}

	// Create and configure gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterMessagingServiceServer(grpcServer, container.MessagingHandler)

	// Start the gRPC server
	go func() {
		sugar.Infow("Starting gRPC server", "port", container.Config.GrpcPort)
		if err = grpcServer.Serve(listener); err != nil {
			sugar.Fatalw("Failed to serve", "error", err)
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit

	sugar.Info("Shutting down gRPC server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Graceful shutdown of gRPC server
	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	// Wait for either graceful shutdown or timeout
	select {
	case <-ctx.Done():
		sugar.Warn("Shutdown timeout reached, forcing stop")
		grpcServer.Stop()
	case <-done:
		sugar.Info("gRPC server stopped gracefully")
	}

	sugar.Info("Service shutdown complete")
}
