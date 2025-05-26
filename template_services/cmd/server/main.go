package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	configEnv "template-services/internal/configs"
	"template-services/internal/di"
	"template-services/internal/template/routes"
	"template-services/pkg/observability"
	pb "template-services/proto"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"google.golang.org/grpc"
)

func InitApp(app *fiber.App, tracerShutdownFuncs func(context.Context) error, appConfigs *configEnv.Config, iConfig *configEnv.ImmutableConfig, firstStart bool) {
	configEnv.ClearPreRestartHooks()
	if tracerShutdownFuncs != nil {
		if err := tracerShutdownFuncs(context.Background()); err != nil {
			fmt.Println("Error shutting down OpenTelemetry SDK:", err)
		}
	}

	err := di.GlobalContainer.ConfigService.RegisterWebhook(context.TODO())
	if err != nil {
		fmt.Println("Error registering webhook:", err)
	}

	tracerShutdownFuncs, err = observability.SetupOTelSDK(context.Background(), appConfigs)
	configEnv.RegisterPreRestartHook(func() {
		fmt.Println("Shutting down tracer...")
		tracerShutdownFuncs(context.Background())
		fmt.Println("Shutting down server...")
		app.Shutdown()
	})
	if err != nil {
		log.Fatalf("❌ Failed to initialize OpenTelemetry SDK: %v", err)
	}

	// Initialize DI container
	container, err := di.NewContainer()
	if err != nil {
		log.Fatalf("❌ Failed to initialize container: %v", err)
	}

	// Get handlers from container
	templateHandler, templateGRPCHandler, err := di.InitHandlers(container)
	if err != nil {
		log.Fatalf("❌ Failed to initialize handlers: %v", err)
	}

	app = fiber.New()
	// Middleware
	// app.Use(recover.New())
	app.Use(logger.New())

	// Setup routes
	routes.SetupTemplateRoutes(app, templateHandler)

	// Start gRPC Server
	go func() {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("GRPC_PORT")))
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}
		grpcServer := grpc.NewServer()
		pb.RegisterTemplateServiceServer(grpcServer, templateGRPCHandler)

		configEnv.RegisterPreRestartHook(func() {
			fmt.Println("Shutting down gRPC server...")
			grpcServer.GracefulStop()
			listener.Close()
			fmt.Println("gRPC server stopped")
		})
		fmt.Println("gRPC server listening on :", fmt.Sprintf(":%s", appConfigs.GRPCPort))
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Start HTTP Server
	fmt.Println("HTTP server listening on :", fmt.Sprintf(":%s", appConfigs.HttpPort))
	err = app.Listen(fmt.Sprintf(":%s", appConfigs.HttpPort))
	for err != nil {
		err = app.Listen(fmt.Sprintf(":%s", appConfigs.HttpPort))
		fmt.Println(err)
		time.Sleep(5 * time.Second)
	}
}

func main() {
	err := configEnv.LoadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}
	di.InitCacheConfig()

	var fiberApp *fiber.App = nil
	var tracerShutdownFuncs func(context.Context) error = nil

	firstStart := false
	go configEnv.ListenForConfigChanges(func(configs *configEnv.Config, iConfig *configEnv.ImmutableConfig) {
		InitApp(fiberApp, tracerShutdownFuncs, configs, iConfig, firstStart)
	})

	time.Sleep(2 * time.Second) // Wait for the server to start

	config, err := di.GlobalContainer.ConfigService.GetCurrentConfig(context.Background(), "config")
	if err != nil {
		fmt.Println("Error fetching config:", err)
		os.Exit(1)
	}

	configEnv.RefreshConfig(config)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-stop
	fmt.Println("Shutting down gracefully...")
}
