package rest

import (
	v1 "thirdparty_service/internal/interface/http/v1"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// InitializeApp initializes the application with all dependencies
func InitializeServer(app *fiber.App) (*fiber.App, error) {
	// setup user route
	app.Use(recover.New())
	v1Group := app.Group("/api/v1")
	v1.SetupRoutes(v1Group)

	return app, nil
}
