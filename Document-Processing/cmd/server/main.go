package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"Document-Processing/internal/di"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Initialize application container
	container, err := di.NewContainer()
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		if err := container.Server.Start(); err != nil {
			log.Printf("Server error: %v", err)
			sigChan <- syscall.SIGTERM // Trigger shutdown on error
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Received shutdown signal")

	// Cleanup resources
	if err := container.Close(); err != nil {
		log.Printf("Error during cleanup: %v", err)
	}

	log.Println("Server shutdown complete")
}
