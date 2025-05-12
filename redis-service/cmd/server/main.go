package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	app_module "redis-service/internal/app"
	handler "redis-service/internal/interface/grpc"
	"redis-service/internal/interface/http"
	"redis-service/pkg/config"
	server "redis-service/pkg/grpc"
	opentelemetry "redis-service/pkg/otel"
	"redis-service/proto"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
)

func main() {
	// load environment variables
	config.LoadEnv()

	// init otel sdk
	shutdown, err := opentelemetry.SetupOTelSDK(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize OpenTelemetry: %v", err)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			log.Fatalf("Failed to shutdown OpenTelemetry: %v", err)
		}
	}()

	// init container
	app_module.InitContainer()

	// Create Fiber app with tracing middleware
	app := fiber.New()
	app.Use(http.TraceMiddleware())

	// create a new grpc server instance
	app_module.Di.Logger.Info("Creating gRPC server instance...")

	// create a new grpc server instance with tracing interceptor
	lis, err := net.Listen("tcp", config.GRPC_PORT)
	if err != nil {
		app_module.Di.Logger.Error("Failed to listen", "error", err)
		os.Exit(1)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(server.TraceInterceptor()),
	)

	grpcServer, err := server.NewGRPCServer(s, lis)
	if err != nil {
		app_module.Di.Logger.Error("gRPC server failed to listen", "error", err)
		os.Exit(1)
	}

	grpcHandler := handler.NewGRPCHandler()
	app_module.Di.Logger.Info("Registering gRPC services...")
	proto.RegisterCacheServiceServer(grpcServer.GetServer(), grpcHandler)

	app.Get("/health", http.HealthCheck)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, syscall.SIGTERM)

	go grpcServer.Start()
	go func() {
		err := app.Listen(config.PORT)
		if err != nil {
			app_module.Di.Logger.Error("Failed to start HTTP server", "error", err)
			os.Exit(1)
		}
		app_module.Di.Logger.Info("Starting HTTP server", "port", config.PORT)
	}()

	<-stop
	app_module.Di.Logger.Info("Stopping gRPC server...")
	grpcServer.Stop()
}
