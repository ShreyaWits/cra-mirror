package main

import (
	"context"
	"errors"
	"log"
	"nps-config-service/internal/app"
	"nps-config-service/internal/configs"
	"nps-config-service/internal/configs/db"
	"nps-config-service/pkg/observability"
	"os"
	"os/signal"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
)

func main() {

	//env file loading
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	appFiber := fiber.New()
	if err := run(appFiber, config); err != nil {
		log.Fatalln(err)
	}
	db.ConnectDatabase(config)

	//setup routes
	appInstance := app.SetupRouter()

	err = appInstance.Listen(":" + config.Port)
	log.Println("Server started on port: " + config.Port)
	if err != nil {

		log.Fatalf("Error starting server: %v", err)
	}

}

func run(app *fiber.App, config *configs.Config) (err error) {
	// Graceful shutdown support
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Setup OpenTelemetry
	otelShutdown, err := observability.SetupOTelSDK(ctx, config)
	if err != nil {
		return
	}
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	// Create Fiber app
	// app := fiber.New()
	type Option interface {
		// contains filtered or unexported methods
	}
	app.Use(otelfiber.Middleware())
	// Start server in background

	// Wait for interrupt signal
	// <-ctx.Done()

	// Gracefully shutdown Fiber
	// ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()
	// if err := app.ShutdownWithContext(ctxShutdown); err != nil {
	// 	log.Printf("Error shutting down server: %v", err)
	// }

	return
}
