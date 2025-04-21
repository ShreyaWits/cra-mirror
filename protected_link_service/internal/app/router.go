package router

import (
	"protected_link/internal/common/api/routes"
	database "protected_link/pkg/redis"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberOtel "github.com/psmarcin/fiber-opentelemetry/pkg/fiber-otel"
)

func SetupRoutes(radis *database.RedisConfig, server *fiber.App) {

	// Allow CORS
	server.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Fiber OpenTelemetry
	server.Use(fiberOtel.New(fiberOtel.Config{
		SpanName:     "HTTP Api Call",
		LocalKeyName: "otel-context",
	}))

	// Initialize routes
	routes.InitializeServerRoutes(radis, server)
}
