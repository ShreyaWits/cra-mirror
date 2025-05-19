package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	pb "cra-protos/messaging_service"
	routes "messaging_service/internal/app"
	"messaging_service/internal/modules/message_broker/di"
	"messaging_service/pkg/observability"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
)

const shutdownTimeout = 5 * time.Second

func main() {
	container, err := di.NewContainer()
	if err != nil {
		panic(err)
	}
	obs := container.Obs
	contx := context.Background()
	ctx, span := container.Obs.TracerService.StartTracer(contx, container.Env.ServiceName)
	defer container.Obs.TracerService.StopSpan(span)

	// Setup servers
	grpcServer, grpcListener := setupGRPC(ctx, container, obs)
	httpServer := setupHTTP(container)

	// Start servers
	var wg sync.WaitGroup
	wg.Add(2)
	defer container.Obs.TracerService.StopSpan(span)
	go func() {
		defer wg.Done()
		obs.LoggerService.Info(ctx, "Starting gRPC server", map[string]interface{}{
			"port": container.Env.GrpcPort,
		})
		if err := grpcServer.Serve(grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			obs.LoggerService.Error(ctx, "gRPC server failed", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	go func() {
		defer wg.Done()
		obs.LoggerService.Info(ctx, "Starting HTTP server", map[string]interface{}{
			"port": container.Env.HttpPort,
		})
		if err := httpServer.Listen(":" + container.Env.HttpPort); err != nil {
			obs.LoggerService.Error(ctx, "HTTP server failed", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	// Handle shutdown
	handleShutdown(ctx, obs, grpcServer, httpServer)
	wg.Wait()
	obs.LoggerService.Info(ctx, "Service shutdown complete")
}

func setupGRPC(ctx context.Context, container *di.Container, obs *observability.ObservabilityStack) (*grpc.Server, net.Listener) {

	listener, err := net.Listen("tcp", ":"+container.Env.GrpcPort)
	if err != nil {
		obs.LoggerService.Error(ctx, "Failed to bind gRPC port", map[string]interface{}{
			"port":  container.Env.GrpcPort,
			"error": err.Error(),
		})
		os.Exit(1)
	}
	server := grpc.NewServer()
	pb.RegisterMessagingServiceServer(server, container.MessagingHandler)
	return server, listener
}

func setupHTTP(container *di.Container) *fiber.App {
	app := fiber.New()
	routes.RegisterConfigRoutes(app, container.ConfigHandler)
	return app
}

func handleShutdown(ctx context.Context, obs *observability.ObservabilityStack, grpcServer *grpc.Server, fiberApp *fiber.App) {

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	obs.LoggerService.Info(ctx, "Shutdown signal received")

	ctx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	// gRPC shutdown
	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		obs.LoggerService.Info(ctx, "gRPC server shut down gracefully")
	case <-ctx.Done():
		obs.LoggerService.Warn(ctx, "gRPC shutdown timed out, forcing stop")
		grpcServer.Stop()
	}

	// Fiber shutdown
	if err := fiberApp.Shutdown(); err != nil {
		obs.LoggerService.Warn(ctx, "Fiber shutdown failed", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		obs.LoggerService.Info(ctx, "Fiber server shut down gracefully")
	}
}
