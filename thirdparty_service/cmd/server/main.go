package main

import (
	"context"
	"log"
	"net"
	"os/signal"
	"sync"
	"syscall"
	"time"

	app_module "thirdparty_service/internal/app"
	"thirdparty_service/internal/config"
	rpc "thirdparty_service/internal/interface/grpc"
	rest "thirdparty_service/internal/interface/http"
	v1 "thirdparty_service/internal/interface/http/v1"
	"thirdparty_service/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app_module.InitDependencyInjection()

	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.FiberErrorHandler, // <-- Your custom handler
	})

	// Initialize Fiber REST server
	restServer, err := rest.InitializeServer(app)
	if err != nil {
		log.Fatalf("error initializing fiber app: %v", err)
	}

	if err := v1.SetupRoutes(restServer); err != nil {
		log.Fatalf("error setting application routes: %v", err)
	}

	// Initialize gRPC server
	grpcServer, err := rpc.InitializeGRPCServer()
	if err != nil {
		log.Fatalf("error initializing grpc server: %v", err)
	}

	var wg sync.WaitGroup

	// Run Fiber server
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Starting REST server...")
		if err := restServer.Listen(config.AppConfig.GetHTTPListenAddress()); err != nil {
			log.Printf("REST server stopped: %v", err)
		}
	}()

	// Run gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Starting gRPC server...")
		ln, err := net.Listen("tcp", config.AppConfig.GetGRPCListenAddress())
		if err != nil {
			log.Fatalf("error creating a listener for gRPC server: %v", err)
			return
		}

		if err := grpcServer.Serve(ln); err != nil {
			log.Fatalf("gRPC server stopped: %v", err)
		}
	}()

	// Graceful shutdown
	<-ctx.Done()
	log.Println("Shutting down servers...")

	// Stop accepting new connections
	stop()

	// Shutdown REST server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := restServer.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("REST shutdown error: %v", err)
	}

	// Shutdown gRPC server
	grpcServer.GracefulStop()

	wg.Wait()
	log.Println("Shutdown complete.")
}
