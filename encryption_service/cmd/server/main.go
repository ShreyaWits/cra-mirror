package main

import (
	"log"
	"net"

	"encryption_microservice/internal/modules/encryption/di"

	pb "encryption_microservice/internal/common/proto_gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {

	// Initialize encryption service
	container, err := di.NewContainer()
	
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}

	grpcServer := grpc.NewServer()

	lis, err := net.Listen("tcp", ":"+container.Config.GrpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	pb.RegisterEncryptionServiceServer(grpcServer, &container.EncryptionHandler)
	// Register reflection service on gRPC server (optional but useful for tools like grpcurl)
	reflection.Register(grpcServer)

	log.Printf("gRPC server is running on port %s...\n", container.Config.GrpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
