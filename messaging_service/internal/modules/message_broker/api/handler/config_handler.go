package handler

import (
	"fmt"
	"messaging_service/internal/config"
	"messaging_service/internal/modules/message_broker/service"

	"github.com/gofiber/fiber/v2"
)

// ReinitializeCallback is a function type for handling reinitialization
type ReinitializeCallback func(configuration *config.Config) error

type ConfigHandler struct {
	configManagerService *service.ConfigManagerService
	env                  *config.Env
	reinitCallback       ReinitializeCallback
}

// NewConfigHandler creates a new instance of ConfigHandler
func NewConfigHandler(configManagerService *service.ConfigManagerService, env *config.Env) (*ConfigHandler, error) {
	// Validate input arguments
	if configManagerService == nil {
		return nil, fmt.Errorf("configManagerService cannot be nil in NewConfigHandler")
	}
	if env == nil {
		return nil, fmt.Errorf("env cannot be nil in NewConfigHandler")
	}

	// Validate required environment fields
	if env.ServiceName == "" {
		return nil, fmt.Errorf("service name is required in environment configuration")
	}

	return &ConfigHandler{
		configManagerService: configManagerService,
		env:                  env,
	}, nil
}

// SetReinitCallback sets the callback to be executed when configurations are updated
func (h *ConfigHandler) SetReinitCallback(callback ReinitializeCallback) {
	h.reinitCallback = callback
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

	err := h.configManagerService.SetDataToCache(ctx.Context(), h.env.ServiceName, configurations)
	if err != nil {
		fmt.Println("error saving data in cache service", err)
	}

	// Execute the reinitialization callback if it's been set
	if h.reinitCallback != nil {
		if err := h.reinitCallback(configurations); err != nil {
			fmt.Println("Error during reinitialization:", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Error reinitializing services: " + err.Error(),
			})
		}
		fmt.Println("Services reinitialized successfully")
	} else {
		fmt.Println("No reinitialization callback set")
	}

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
