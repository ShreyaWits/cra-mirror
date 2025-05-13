package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"Document-Processing/internal/config"
	"Document-Processing/internal/di"
)

func main() {
	// Load environment variables
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwiZXhwIjoxNzQ3MTQxNzc5fQ.KcsFZ8YWHHcp41IZRhT5Z4mgDdzfOI9w7YJ4GbXxBg0"

	rawData, err := config.LoadConfigFromAPI(token)
	if err != nil {
		log.Fatalf("Failed to fetch config from API: %v", err)
	}

	cfg, err := config.NewConfig(rawData)
	if err != nil {
		log.Fatalf("Config validation failed: %v", err)
	}

	log.Printf("Config initialized: %+v\n", cfg)

	// Initialize application container
	container, err := di.NewContainer(cfg)
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
