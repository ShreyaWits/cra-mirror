package main

import (
	pb "cra-protos/messaging_service"
	"log"
	"messaging_service/internal/di"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

func main() {
	// Initialize dependency container
	container, err := di.NewContainer()
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}

	// Set up a TCP listener
	listener, err := net.Listen("tcp", ":"+container.Config.GrpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create and configure gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterMessagingServiceServer(grpcServer, container.MessagingHandler)

	// Start the gRPC server
	go func() {
		log.Println("Starting gRPC server on:", container.Config.GrpcPort)
		if err = grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Set up graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Println("Shutting down gRPC server...")
	grpcServer.GracefulStop()

	// Allow time for ongoing requests to complete
	time.Sleep(1 * time.Second)
	log.Println("gRPC server stopped")
}
