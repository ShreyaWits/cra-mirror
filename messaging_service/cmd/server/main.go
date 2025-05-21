package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	pb "cra-protos/messaging_service"
	routes "messaging_service/internal/app"
	"messaging_service/internal/config"
	"messaging_service/internal/modules/message_broker/di"
	"messaging_service/pkg/observability"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	shutdownTimeout     = 5 * time.Second
	grpcShutdownTimeout = 3 * time.Second
)

func main() {
	// Create a root context that can be cancelled
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	// Initialize container with better error handling
	container, err := di.NewContainer()
	if err != nil {
		fmt.Printf("ERROR: Failed to initialize container: %v\n", err)
		os.Exit(1)
	}

	// Create observability context
	obs := container.Obs
	ctx, span := obs.TracerService.StartTracer(rootCtx, config.SERVICE_NAME)
	defer obs.TracerService.StopSpan(span)

	// Setup servers with graceful shutdown
	grpcServer, grpcListener, err := setupGRPC(ctx, container, obs)
	if err != nil {
		obs.LoggerService.Error(ctx, "Failed to setup gRPC server", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}

	httpServer := setupHTTP(container)

	// Wait group for server goroutines
	var wg sync.WaitGroup
	wg.Add(2)

	// Start gRPC server
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

	// Start HTTP server
	go func() {
		defer wg.Done()
		obs.LoggerService.Info(ctx, "Starting HTTP server", map[string]interface{}{
			"port": container.Env.HttpPort,
		})
		if err := httpServer.Listen(":" + container.Env.HttpPort); err != nil {
			// For Fiber there's no standard error for server closed
			if err.Error() != "server closed" {
				obs.LoggerService.Error(ctx, "HTTP server failed", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}
	}()

	// Set up signal handling for graceful shutdown
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a shutdown signal
	sig := <-shutdownChan
	obs.LoggerService.Info(ctx, "Shutdown signal received", map[string]interface{}{
		"signal": sig.String(),
	})

	// Cancel the root context to signal shutdown to ongoing operations
	rootCancel()

	// Create a timeout context for the shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	// Perform the shutdown
	shutdownServices(shutdownCtx, obs, grpcServer, httpServer, container)

	// Wait for servers to exit
	wg.Wait()
	obs.LoggerService.Info(ctx, "Service shutdown complete")
}

func setupGRPC(ctx context.Context, container *di.Container, obs *observability.ObservabilityStack) (*grpc.Server, net.Listener, error) {
	listener, err := net.Listen("tcp", ":"+container.Env.GrpcPort)
	if err != nil {
		obs.LoggerService.Error(ctx, "Failed to bind gRPC port", map[string]interface{}{
			"port":  container.Env.GrpcPort,
			"error": err.Error(),
		})
		return nil, nil, err
	}

	// Create gRPC server with options
	server := grpc.NewServer(
		grpc.MaxConcurrentStreams(100), // Limit concurrent streams
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	)

	pb.RegisterMessagingServiceServer(server, container.MessagingHandler)
	return server, listener, nil
}

func setupHTTP(container *di.Container) *fiber.App {
	// Configure Fiber with sensible defaults
	app := fiber.New(fiber.Config{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	})

	routes.RegisterConfigRoutes(app, container.ConfigHandler)
	return app
}

func shutdownServices(ctx context.Context, obs *observability.ObservabilityStack,
	grpcServer *grpc.Server, fiberApp *fiber.App, container *di.Container) {

	var wg sync.WaitGroup
	wg.Add(2) // For gRPC and HTTP servers

	// Shutdown gRPC server
	go func() {
		defer wg.Done()

		grpcDone := make(chan struct{})
		go func() {
			obs.LoggerService.Info(ctx, "Stopping gRPC server")
			grpcServer.GracefulStop()
			close(grpcDone)
		}()

		select {
		case <-grpcDone:
			obs.LoggerService.Info(ctx, "gRPC server shut down gracefully")
		case <-time.After(grpcShutdownTimeout):
			obs.LoggerService.Warn(ctx, "gRPC shutdown timed out, forcing stop")
			grpcServer.Stop()
		}
	}()

	// Shutdown HTTP server
	go func() {
		defer wg.Done()

		obs.LoggerService.Info(ctx, "Stopping HTTP server")
		if err := fiberApp.Shutdown(); err != nil {
			obs.LoggerService.Warn(ctx, "Fiber shutdown failed", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			obs.LoggerService.Info(ctx, "Fiber server shut down gracefully")
		}
	}()

	// Wait for both servers to shut down or timeout
	wgDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(wgDone)
	}()

	select {
	case <-wgDone:
		obs.LoggerService.Info(ctx, "All servers shut down")
	case <-ctx.Done():
		obs.LoggerService.Warn(ctx, "Server shutdown timed out")
	}

	// Finally, close the container to clean up resources
	if err := container.Close(ctx); err != nil {
		obs.LoggerService.Error(ctx, "Error during container shutdown", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		obs.LoggerService.Info(ctx, "Container resources cleaned up")
	}
}
