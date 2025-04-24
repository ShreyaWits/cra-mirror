package app

import (
	handler "nps-config-service/internal/modules/config-manager/apis/handlers"
	"nps-config-service/internal/modules/config-manager/apis/routes"

	"nps-config-service/internal/modules/config-manager/di"
	"nps-config-service/internal/modules/config-manager/services"

	"github.com/gofiber/fiber/v2"
)

func RegisterModules(app *fiber.App) {
	// Dependency injection
	api := app.Group("/api/v1/")
	protected := api.Group("/config")

	// Initialize shared dependencies only once
	deps, err := di.InitSharedDependencies()
	if err != nil {
		panic("Failed to initialize shared dependencies: " + err.Error())
	}

	// Webhook handler using shared webhook service
	webhookHandler := handler.NewWebhookHandler(deps.WebhookService)
	routes.RegisterWebHookRoutes(protected, webhookHandler)

	// Config handler using shared repo and webhook service
	configService := services.NewConfigService(deps.Repo, deps.WebhookService)
	configHandler := handler.NewConfigHandler(configService)

	routes.RegisterAdminRoutes(api, configHandler)
	routes.RegisterConfigRoutes(protected, configHandler)
}
