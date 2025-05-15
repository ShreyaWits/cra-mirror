package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "cra-protos/messaging_service"
	routes "messaging_service/internal/app"
	"messaging_service/internal/modules/message_broker/di"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const shutdownTimeout = 5 * time.Second

func main() {
	// 1. Initialize Logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	// 2. Initialize Dependency Injection Container
	container, err := di.NewContainer()
	if err != nil {
		sugar.Fatalw("Failed to initialize container", "error", err)
	}

	// 3. gRPC Server Setup
	grpcServer := grpc.NewServer()
	pb.RegisterMessagingServiceServer(grpcServer, container.MessagingHandler)

	grpcListener, err := net.Listen("tcp", ":"+container.Config.GrpcPort)
	if err != nil {
		sugar.Fatalw("Failed to bind to port", "port", container.Config.GrpcPort, "error", err)
	}

	// Start gRPC server in goroutine
	go func() {
		sugar.Infow("Starting gRPC server", "port", container.Config.GrpcPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
			sugar.Fatalw("gRPC server failed", "error", err)
		}
	}()

	// 4. Fiber HTTP App Setup
	app := fiber.New()
	routes.RegisterConfigRoutes(app, container.ConfigHandler)

	go func() {
		httpPort := container.Config.HttpPort
		sugar.Infow("Starting Fiber HTTP server", "port", httpPort)
		if err := app.Listen(":" + httpPort); err != nil {
			sugar.Fatalw("Fiber HTTP server failed", "error", err)
		}
	}()

	// 5. Graceful Shutdown Handling
	waitForShutdown(sugar, grpcServer, app)
}

// waitForShutdown handles graceful termination of both gRPC and Fiber servers
func waitForShutdown(logger *zap.SugaredLogger, grpcServer *grpc.Server, app *fiber.App) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-quit
	logger.Infow("Shutdown signal received")

	// Context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Graceful gRPC shutdown
	grpcDone := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(grpcDone)
	}()

	// Graceful Fiber shutdown
	fiberDone := make(chan struct{})
	go func() {
		_ = app.Shutdown()
		close(fiberDone)
	}()

	select {
	case <-grpcDone:
		logger.Info("gRPC server shut down gracefully")
	case <-ctx.Done():
		logger.Warn("gRPC shutdown timed out, forcing stop")
		grpcServer.Stop()
	}

	select {
	case <-fiberDone:
		logger.Info("Fiber server shut down gracefully")
	case <-ctx.Done():
		logger.Warn("Fiber shutdown timed out")
	}

	logger.Info("Service shutdown complete")
}
