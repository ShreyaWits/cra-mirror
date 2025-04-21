package main

import (
	"log"
	"nps-config-service/internal/config-manager/apis/routes"
	"os"

	db "nps-config-service/pkg/etcd"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
        log.Println("No .env file found or error loading it")
    }
    
    app := fiber.New()
    routes.SetupRoutes(app)

    db.InitEtcdDB()
	defer db.Client.Close()

    port := os.Getenv("PORT")
    if port == "" {
        port = "3000" // Default port if not specified in env
    }

    log.Fatal(app.Listen(":" + port))
}