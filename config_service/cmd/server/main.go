package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nps-config-service/internal/app"
	"nps-config-service/internal/configs"
	"nps-config-service/internal/modules/config-manager/di"
	"nps-config-service/pkg/observability"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {
	// Create a context that will be canceled on SIGINT or SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Load configuration
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize OpenTelemetry
	shutdown, err := observability.SetupOTelSDK(ctx, config)
	if err != nil {
		log.Fatalf("Failed to initialize OpenTelemetry: %v", err)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down OpenTelemetry: %v", err)
		}
	}()

	// Initialize database
	// db.ConnectDatabase(config)

	// Initialize DI container
	container, err := di.InitContainer()
	if err != nil {
		log.Fatalf("Failed to initialize DI container: %v", err)
	}
	defer container.Close()

	// Create Fiber app
	appFiber := fiber.New(fiber.Config{
		AppName:      "NPS Config Service",
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
		IdleTimeout:  time.Second * 5,
	})

	// Add middleware
	appFiber.Use(cors.New())
	appFiber.Use(requestid.New())
	appFiber.Use(logger.New())

	// Setup routes with container
	appInstance := app.SetupRouter(appFiber, container)

	// Start server in a goroutine
	go func() {
		if err := appInstance.Listen(":" + config.Port); err != nil {
			log.Printf("Server error: %v", err)
			stop() // Signal shutdown on server error
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := appInstance.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	log.Println("Server shutdown complete")
}
