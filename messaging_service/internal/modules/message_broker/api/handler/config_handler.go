package handler

import (
	"context"
	"fmt"
	"messaging_service/internal/config"
	"messaging_service/internal/modules/message_broker/service"
	"messaging_service/pkg/observability"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	defaultTimeoutConfig = 5 * time.Second
)

// ReinitializeCallback is a function type for handling reinitialization
type ReinitializeCallback func(configuration *config.Config) error

type ConfigHandler struct {
	configManagerService service.ConfigManagerServiceInterface
	env                  *config.Env
	reinitCallback       ReinitializeCallback
	obs                  *observability.ObservabilityStack
}

// NewConfigHandler creates a new instance of ConfigHandler
func NewConfigHandler(configManagerService service.ConfigManagerServiceInterface, env *config.Env, obs *observability.ObservabilityStack) (*ConfigHandler, error) {
	// Validate input arguments
	if configManagerService == nil {
		return nil, fmt.Errorf("configManagerService cannot be nil in NewConfigHandler")
	}
	if env == nil {
		return nil, fmt.Errorf("env cannot be nil in NewConfigHandler")
	}
	if obs == nil {
		return nil, fmt.Errorf("observability stack cannot be nil in NewConfigHandler")
	}


	return &ConfigHandler{
		configManagerService: configManagerService,
		env:                  env,
		obs:                  obs,
	}, nil
}

// SetReinitCallback sets the callback to be executed when configurations are updated
func (h *ConfigHandler) SetReinitCallback(callback ReinitializeCallback) {
	h.reinitCallback = callback
}

func (h *ConfigHandler) UpdateConfigurations(ctx *fiber.Ctx) error {

	functionName := "UpdateConfigurations"
	functionFailed := "UpdateConfigurations_Failed"

	// Create context with timeout
	fiberCtx, cancel := context.WithTimeout(ctx.Context(), defaultTimeoutConfig)
	defer cancel()
	tCtx, span := h.obs.TracerService.StartTracer(fiberCtx, functionName)
	defer h.obs.TracerService.StopSpan(span)

	h.obs.MetricsService.IncrementCounter(tCtx, functionName, 1, map[string]string{})

	h.obs.LoggerService.Info(tCtx, "Configuration update requested")

	var configurations *config.Config

	if err := ctx.BodyParser(&configurations); err != nil {
		loggerdata := fmt.Sprintf("Failed to parse configuration update: %v", err)
		h.obs.LoggerService.Error(tCtx, loggerdata)
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{"error": loggerdata})

		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	config.SetConfig(configurations)

	err := h.configManagerService.SetDataToCache(tCtx, "config", configurations)
	if err != nil {
		loggerdata := fmt.Sprintf("Error saving data in cache service : %v", err)
		h.obs.LoggerService.Error(tCtx, loggerdata)
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{"error": loggerdata})
	}

	// Execute the reinitialization callback if it's been set
	if h.reinitCallback != nil {
		if err := h.reinitCallback(configurations); err != nil {
			loggerdata := fmt.Sprintf("Error during reinitialization: %v", err)
			h.obs.LoggerService.Error(tCtx, loggerdata)
			h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{"error": loggerdata})

			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Error reinitializing services: " + err.Error(),
			})
		}
		h.obs.LoggerService.Info(tCtx, "Services reinitialized successfully")
	} else {
		h.obs.LoggerService.Info(tCtx, "No reinitialization callback set")
	}

	h.obs.LoggerService.Info(tCtx, "Configuration updated successfully")
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Configuration updated successfully",
	})
}

func (h *ConfigHandler) Health(ctx *fiber.Ctx) error {
	functionName := "Health"

	// Create context with timeout
	tCtx, cancel := context.WithTimeout(ctx.Context(), defaultTimeoutConfig)
	defer cancel()

	// Start tracing
	tCtx, span := h.obs.TracerService.StartTracer(tCtx, functionName)
	defer h.obs.TracerService.StopSpan(span)

	h.obs.MetricsService.IncrementCounter(tCtx, functionName, 1, map[string]string{})
	h.obs.LoggerService.Info(tCtx, "Health check requested")

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "OK",
		"success": true,
	})
}
