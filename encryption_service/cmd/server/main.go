package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"encryption_microservice/internal/app"
	"encryption_microservice/internal/modules/encryption/di"

	pb "encryption_microservice/internal/common/proto_gen"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize encryption service
	container, err := di.NewContainer()
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}
	defer container.Close(ctx)

	// Set up signal handling for graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// Start the servers
	grpcServer := startGRPCServer(container)
	fiberApp := startFiberServer(container)

	// Wait for termination signal
	<-sig
	log.Println("Shutdown signal received")

	// Gracefully shut down both servers
	shutdownServers(ctx, grpcServer, fiberApp)
	log.Println("Servers stopped. Exiting")
}

func startGRPCServer(container *di.Container) *grpc.Server {
	grpcServer := grpc.NewServer()

	lis, err := net.Listen("tcp", ":"+container.Env.GrpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	pb.RegisterEncryptionServiceServer(grpcServer, &container.EncryptionHandler)
	reflection.Register(grpcServer)

	log.Printf("gRPC server is running on port %s...\n", container.Env.GrpcPort)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("Error serving gRPC: %v", err)
		}
	}()

	return grpcServer
}

func startFiberServer(container *di.Container) *fiber.App {
	fiberApp := fiber.New(fiber.Config{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	// Setup routes
	app.SetupRoutes(fiberApp, &container.EncryptionHandler, container.ConfigHandler)

	log.Printf("REST server is running on port %s...\n", container.Env.ServerPort)
	go func() {
		if err := fiberApp.Listen(":" + container.Env.ServerPort); err != nil {
			log.Printf("Error serving REST: %v", err)
		}
	}()

	return fiberApp
}

func shutdownServers(ctx context.Context, grpcServer *grpc.Server, fiberApp *fiber.App) {
	// Create a context with a timeout for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Gracefully stop the gRPC server
	grpcServer.GracefulStop()

	// Gracefully stop the Fiber server
	if err := fiberApp.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("Error shutting down Fiber server: %v", err)
	}
}
