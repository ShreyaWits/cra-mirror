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

	log.Println("gRPC server is running on port 50051...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	// // Initialize Fiber app and routes
	// app := fiber.New()
	// container.SetupRoutes(app)

	// // Channel to capture server errors
	// errChan := make(chan error, 1)
	// go func() {
	// 	errChan <- app.Listen(":" + container.Config.ServerPort)
	// }()

	// // Graceful shutdown on interrupt or re-throw server errors
	// quit := make(chan os.Signal, 1)
	// signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// select {
	// case sig := <-quit:
	// 	logger.LogEvent("", "Shutdown", "", "Info", fmt.Sprintf("received signal: %v, shutting down server", sig))
	// case err := <-errChan:
	// 	return fmt.Errorf("server error: %w", err)
	// }

	// // Shutdown server
	// if err := app.Shutdown(); err != nil {
	// 	return fmt.Errorf("failed to shutdown server: %w", err)
	// }
	// logger.LogEvent("", "Shutdown", "", "Info", "server shutdown completed")
	// return nil
}
