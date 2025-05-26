package v1

import (
	"log"
	"net/http"
	"thirdparty_service/internal/app"
	"thirdparty_service/internal/config"
	"thirdparty_service/internal/dtos"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(v1 fiber.Router) error {

	// webhook route route
	webhookRoute := v1.Group("/webhook")
	webhookRoute.Post("/", HandleConfig)

	return nil
}

func HandleConfig(c *fiber.Ctx) error {
	configHandler := app.GetConfigHandler()
	dynamicConfig, err := configHandler.GetDynamicConfig()
	if err != nil {
		return dtos.Response{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
		}
	}
	config.AppConfig.SetEnv(dynamicConfig)
	// Set Latest config to cache
	configManagerService := app.GetCacheManagerService()
	ctx := c.UserContext() // Get context from Fiber Ctx
	errorx := configManagerService.SetDataToCache(ctx, "config", dynamicConfig)
	if errorx != nil {
		log.Println("Error occured in SET CACHE DATA: ", err)
	}
	data, err := configManagerService.GetDataToCache(ctx, "config")
	if err != nil {
		log.Println("Error occured in GET CACHE DATA: ", err)
	}
	log.Println("the CACHE DATA: ", data)

	return dtos.Response{
		Code: http.StatusOK,
		Msg:  "Webhook Received Successfully",
	}
}

func HandleHealthCheck(c *fiber.Ctx) error {
	return dtos.Response{
		Code: http.StatusOK,
		Msg:  "Service OK",
	}
}
