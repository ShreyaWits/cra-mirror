package main

import (
	pb "cra-protos/messaging_service"
	"log"
	"messaging_service/internal/messaging_service/handler"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

func main() {

	// container, err := di.NewContainer()
	// if err != nil {
	// 	log.Fatal("Error Loadin Container", err)
	// }
	// Set up a TCP listener on the specified port
	listener, err := net.Listen("tcp", ":"+"50051") // Change the port as needed
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	msgHandler := &handler.MessagingHandler{}
	// Register the messaging service
	pb.RegisterMessagingServiceServer(grpcServer, msgHandler)

	// Start the gRPC server in a goroutine
	go func() {
		log.Println("Starting gRPC server on :", "50051")
		if err = grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Graceful shutdown handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for interrupt signal to gracefully shutdown the server
	<-signalChan
	log.Println("Shutting down gRPC server...")

	// Gracefully stop the gRPC server
	grpcServer.GracefulStop()

	// Optionally, you can add a timeout for the shutdown process
	time.Sleep(1 * time.Second) // Allow some time for ongoing requests to complete
	log.Println("gRPC server stopped")
}
