package main

import (
	"log"
	"os"
	"os/signal"
	"protected_link/cmd/initializer"
	router "protected_link/internal/app"
	"protected_link/internal/common/api/middlewares"
	"syscall"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	log.Println("🚀 Starting the server...")

	app.Use(middlewares.RecoveryMiddleware())

	// Initialize Redis and other dependencies
	cfg, db := initializer.InitialSetup()
	if db == nil {
		log.Println("❌ Failed to initialize Redis. Shutting down gracefully.")
		if err := app.Shutdown(); err != nil {
			log.Fatalf("❌ Error shutting down server: %v", err)
		}
		os.Exit(1)
	}

	router.SetupRoutes(db, app)

	// Start the server in a goroutine
	go func() {
		port := cfg.ServerPort
		if port == "" {
			port = "9000"
		}
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown on interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Println("🔻 Shutting down server...")

	// Shutdown Fiber app
	if err := app.Shutdown(); err != nil {
		log.Fatalf("❌ Error shutting down server: %v", err)
	}

	// Close Redis connection
	if db != nil && db.Client != nil {
		if err := db.Client.Close(); err != nil {
			log.Printf("❌ Error closing Redis connection: %v", err)
		} else {
			log.Println("✅ Redis connection closed.")
		}
	}

	log.Println("✅ Server shut down gracefully.")
}
