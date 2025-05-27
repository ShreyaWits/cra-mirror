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
	// Metric names
	metricConfigUpdateTotal   = "config_update_total"
	metricConfigUpdateSuccess = "config_update_success"
	metricConfigUpdateFailure = "config_update_failure"
	metricHealthCheckTotal    = "health_check_total"
	metricHealthCheckSuccess  = "health_check_success"
	metricHealthCheckFailure  = "health_check_failure"
	// Error types
	errorTypeParse      = "parse_error"
	errorTypeCache      = "cache_error"
	errorTypeReinit     = "reinit_error"
	errorTypeValidation = "validation_error"
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

	// Create context with timeout
	fiberCtx, cancel := context.WithTimeout(ctx.Context(), defaultTimeoutConfig)
	defer cancel()

	// Start tracing
	tCtx, span := h.obs.TracerService.StartTracer(fiberCtx, functionName)
	defer h.obs.TracerService.StopSpan(span)

	// Generate request ID for correlation
	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())

	// Set tracing attributes
	h.obs.TracerService.SetAttributes(span, map[string]string{
		"request_id": requestID,
		"operation":  functionName,
	})

	// Increment total requests metric
	h.obs.MetricsService.IncrementCounter(tCtx, metricConfigUpdateTotal, 1, nil)

	h.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Configuration update request received", requestID))

	// Define a custom wrapper struct to match the incoming JSON
	type ConfigWrapper struct {
		Environment string         `json:"environment"`
		Method      string         `json:"method"`
		ServiceName string         `json:"serviceName"`
		Values      *config.Config `json:"values"`
	}

	var wrapper ConfigWrapper

	// Parse the request body
	if err := ctx.BodyParser(&wrapper); err != nil {
		h.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Failed to parse request body", requestID))
		h.obs.MetricsService.IncrementCounter(tCtx, metricConfigUpdateFailure, 1, map[string]string{
			"error_type": errorTypeParse,
		})

		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request format",
		})
	}

	// Check if we have values
	if wrapper.Values == nil {
		h.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] No configuration values provided", requestID))
		h.obs.MetricsService.IncrementCounter(tCtx, metricConfigUpdateFailure, 1, map[string]string{
			"error_type": errorTypeValidation,
		})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "No configuration values found",
		})
	}

	// Use the parsed config directly
	cfg := wrapper.Values

	// Log only non-sensitive information
	h.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Processing configuration update for environment: %s, service: %s",
		requestID, wrapper.Environment, wrapper.ServiceName))

	// Save to cache if available
	err := h.configManagerService.SetDataToCache(tCtx, "config", cfg)
	if err != nil {
		h.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Failed to update configuration cache", requestID))
		h.obs.MetricsService.IncrementCounter(tCtx, metricConfigUpdateFailure, 1, map[string]string{
			"error_type": errorTypeCache,
		})
	}

	// Execute the reinitialization callback if it's been set
	if h.reinitCallback != nil {
		h.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Initiating service reinitialization", requestID))
		if err := h.reinitCallback(cfg); err != nil {
			h.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Service reinitialization failed", requestID))
			h.obs.MetricsService.IncrementCounter(tCtx, metricConfigUpdateFailure, 1, map[string]string{
				"error_type": errorTypeReinit,
			})

			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Error reinitializing services",
			})
		}
		h.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Service reinitialization completed successfully", requestID))
	}

	h.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Configuration update completed successfully", requestID))
	h.obs.MetricsService.IncrementCounter(tCtx, metricConfigUpdateSuccess, 1, nil)

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

	// Set tracing attributes
	h.obs.TracerService.SetAttributes(span, map[string]string{
		"operation": functionName,
	})

	// Increment metrics
	h.obs.MetricsService.IncrementCounter(tCtx, metricHealthCheckTotal, 1, nil)
	h.obs.MetricsService.IncrementCounter(tCtx, metricHealthCheckSuccess, 1, nil)

	h.obs.LoggerService.Info(tCtx, "Health check completed successfully")

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "OK",
		"success": true,
	})
}
