package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
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
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("Failed to flush logger: %v", err)
		}
	}()
	sugar := logger.Sugar()

	container, err := di.NewContainer()
	if err != nil {
		sugar.Fatalw("Failed to initialize container", "error", err)
	}

	// Setup servers
	grpcServer, grpcListener := setupGRPC(container, sugar)
	httpServer := setupHTTP(container, sugar)

	// Start servers
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		sugar.Infow("Starting gRPC server", "port", container.Env.GrpcPort)
		if err := grpcServer.Serve(grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			sugar.Errorw("gRPC server failed", "error", err)
		}
	}()

	go func() {
		defer wg.Done()
		sugar.Infow("Starting HTTP server", "port", container.Env.HttpPort)
		if err := httpServer.Listen(":" + container.Env.HttpPort); err != nil {
			sugar.Errorw("HTTP server failed", "error", err)
		}
	}()

	// Handle shutdown
	handleShutdown(sugar, grpcServer, httpServer)
	wg.Wait()
	sugar.Info("Service shutdown complete")
}

func setupGRPC(container *di.Container, logger *zap.SugaredLogger) (*grpc.Server, net.Listener) {
	listener, err := net.Listen("tcp", ":"+container.Env.GrpcPort)
	if err != nil {
		logger.Fatalw("Failed to bind gRPC port", "port", container.Env.GrpcPort, "error", err)
	}
	server := grpc.NewServer()
	pb.RegisterMessagingServiceServer(server, container.MessagingHandler)
	return server, listener
}

func setupHTTP(container *di.Container, logger *zap.SugaredLogger) *fiber.App {
	app := fiber.New()
	routes.RegisterConfigRoutes(app, container.ConfigHandler) // typo? should be just RegisterConfigRoutes
	return app
}

func handleShutdown(logger *zap.SugaredLogger, grpcServer *grpc.Server, fiberApp *fiber.App) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Infow("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// gRPC shutdown
	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("gRPC server shut down gracefully")
	case <-ctx.Done():
		logger.Warn("gRPC shutdown timed out, forcing stop")
		grpcServer.Stop()
	}

	// Fiber shutdown
	if err := fiberApp.Shutdown(); err != nil {
		logger.Warnw("Fiber shutdown failed", "error", err)
	} else {
		logger.Info("Fiber server shut down gracefully")
	}
}
