package main

import (
	"log"
	"os"
	"os/signal"
	"redis-service/internal/interface/grpc"
	"redis-service/pkg/config"
	server "redis-service/pkg/grpc"
	"redis-service/proto"
)

func main() {

	log.Println("Loading env variables...")
	// load environment variables
	config.LoadEnv()

	// create a new grpc server instance
	log.Println("Creating gRPC server instance...")
	grpcServer, err := server.NewGRPCServer(":50051")

	if err != nil {
		log.Fatalf("gRPC server failed to listen: %v", err)
	}

	grpcHandler := grpc.NewGRPCHandler()
	log.Println("Registering gRPC services...")
	proto.RegisterCacheServiceServer(grpcServer.GetServer(), grpcHandler)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, os.Kill)

	go grpcServer.Start()

	<-stop
	log.Println("Stopping gRPC server...")
	grpcServer.Stop()

}
