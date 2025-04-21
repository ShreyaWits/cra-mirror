package main

import (
	"log"
	"os"

	"encryption_microservice/internal/modules/encryption/di"

	"github.com/gofiber/fiber/v2"
)

func main() {

	// Initialize encryption service
	container, err := di.NewContainer()
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}

	// Initialize Fiber app
	app := fiber.New()

	// Setup routes
	container.SetupRoutes(app)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
