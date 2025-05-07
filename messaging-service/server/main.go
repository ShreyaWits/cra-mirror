package main

import (
	"log"
	"messaging-service/internal/messaging_service/service"
	"net"

	pb "messaging-service/protos/messaging_service"

	"google.golang.org/grpc"
)

func main() {
	// Set up a TCP listener on the specified port
	listener, err := net.Listen("tcp", ":50051") // Change the port as needed
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Messaging service server instance
	msgServer := &service.MessagingServer{}

	// Register the messaging service
	pb.RegisterMessagingServiceServer(grpcServer, msgServer)

	// Start the gRPC server
	log.Println("Starting gRPC server on :50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
