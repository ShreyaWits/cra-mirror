package main

import (
	"log"

	"Document-Processing/internal/di"
)

func main() {
	// Initialize application container
	container := di.NewContainer()

	// Start server
	if err := container.Server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
