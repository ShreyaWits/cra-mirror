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
	log.Println("Creating gRPC server instance...")

	// create a new grpc server instance with tracing interceptor
	lis, err := net.Listen("tcp", config.GRPC_PORT)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(server.TraceInterceptor()),
	)

	grpcServer, err := server.NewGRPCServer(s, lis)
	if err != nil {
		log.Fatalf("gRPC server failed to listen: %v", err)
	}

	grpcHandler := handler.NewGRPCHandler()
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
