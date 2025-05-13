package v1

import (
	"net/http"
	"thirdparty_service/internal/dtos"
	user_router "thirdparty_service/internal/modules/user/apis/router"
	webhook_router "thirdparty_service/internal/modules/webhooks/api/router"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app_router fiber.Router) error {
	v1 := app_router.Group("/api/v1")

	// health check route
	app_router.Get("/healthz", HandleHealthCheck)

	user_router.SetupUserRoutes(v1)
	webhook_router.SetupWebhookRoutes(v1)

	return nil
}

func HandleHealthCheck(c *fiber.Ctx) error {

	return dtos.Response{
		Code: http.StatusOK,
		Msg:  "Service OK",
	}
}
