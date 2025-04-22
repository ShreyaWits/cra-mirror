package main

import (
	"log"

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

	log.Printf("Starting server on port %s", container.Config.ServerPort)
	if err := app.Listen(":" + container.Config.ServerPort); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
