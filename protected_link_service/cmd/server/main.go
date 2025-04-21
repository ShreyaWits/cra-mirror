package main

import (
	"fmt"
	"log"
	router "protected_link/internal/app"
	"protected_link/internal/common/api/middlewares"
	configEnv "protected_link/internal/configs"
	database "protected_link/pkg/redis"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	println("Hello World", app)

	app.Use(middlewares.RecoveryMiddleware())

	cfg, err := configEnv.LoadConfig()
	db, eror := database.ConnectRedis(cfg)

	if err != nil {
		println("Error loading config:", err)
	}

	if eror != nil {
		println("Error db loading config:", err)
	}
	router.SetupRoutes(db, app)

	port := cfg.ServerPort
	if port == "" {
		port = "9000"
	}

	// Start the Fiber server
	log.Printf("🚀 Server starting on port %s...", port)
	err = app.Listen(fmt.Sprintf(":%s", "8080"))
	if err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
