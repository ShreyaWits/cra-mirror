package main

import (
	"log"
	"os"
	"os/signal"
	"redis-service/internal/interface/grpc"
	"redis-service/internal/interface/http"
	"redis-service/pkg/config"
	server "redis-service/pkg/grpc"
	"redis-service/proto"
	"syscall"

	"github.com/gofiber/fiber/v2"
)

func main() {

	log.Println("Loading env variables...")
	// load environment variables
	config.LoadEnv()

	app := fiber.New()

	// create a new grpc server instance
	log.Println("Creating gRPC server instance...")
	grpcServer, err := server.NewGRPCServer(config.GRPC_PORT)

	if err != nil {
		log.Fatalf("gRPC server failed to listen: %v", err)
	}

	grpcHandler := grpc.NewGRPCHandler()
	log.Println("Registering gRPC services...")
	proto.RegisterCacheServiceServer(grpcServer.GetServer(), grpcHandler)

	app.Get("/health", http.HealthCheck)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, syscall.SIGTERM)

	go grpcServer.Start()
	go func() {
		err := app.Listen(config.PORT)
		if err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
		log.Printf("Starting HTTP server on port %s...", config.PORT)
	}()

	<-stop
	log.Println("Stopping gRPC server...")
	grpcServer.Stop()

}
