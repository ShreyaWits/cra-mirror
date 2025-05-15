package handler

import (
	"fmt"
	"messaging_service/internal/config"
	constants "messaging_service/internal/modules/message_broker/constant"
	"messaging_service/internal/modules/message_broker/service"

	"github.com/gofiber/fiber/v2"
)

type ConfigHandler struct {
	configManagerService *service.ConfigManagerService
}

// NewConfigHandler creates a new instance of ConfigHandler
func NewConfigHandler(configManagerService *service.ConfigManagerService) *ConfigHandler {
	return &ConfigHandler{
		configManagerService: configManagerService,
	}
}

func (h *ConfigHandler) UpdateConfigurations(ctx *fiber.Ctx) error {
	fmt.Println("Config Updated Calling successfully API")
	var configurations *config.Config

	if err := ctx.BodyParser(&configurations); err != nil {
		fmt.Println("Config Updated Calling Error API", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	config.SetConfig(configurations)
	
	err := h.configManagerService.SetDataToCache(ctx.Context(), constants.SERVICE_NAME, configurations)
	if err != nil {
		fmt.Println("error saving data in cache service", err)
	}

	// TODO: update logic here, e.g. h.configManagerService.Apply(configurations)

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Configuration updated successfully",
	})
}

func (h *ConfigHandler) Health(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "OK",
		"success": true,
	})
}
