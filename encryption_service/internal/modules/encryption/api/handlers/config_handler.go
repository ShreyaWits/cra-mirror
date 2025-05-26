package handlers

import (
	"context"
	"encryption_microservice/internal/config"
	"encryption_microservice/pkg/observability"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	defaultTimeoutConfig = 5 * time.Second
)

// ReinitializeCallback is a function type for handling reinitialization
type ReinitializeCallback func(configuration *config.Config) error

// ConfigHandler handles configuration-related API endpoints
type ConfigHandler struct {
	cacheClient interface {
		SetDataToCache(ctx context.Context, key string, value interface{}) error
	}
	env            *config.EnvConfig
	reinitCallback ReinitializeCallback
	obs            *observability.ObservabilityStack
}

// NewConfigHandler creates a new instance of ConfigHandler
func NewConfigHandler(cacheClient interface {
	SetDataToCache(ctx context.Context, key string, value interface{}) error
},
	env *config.EnvConfig,
	obs *observability.ObservabilityStack) (*ConfigHandler, error) {
	// Validate input arguments
	if cacheClient == nil {
		return nil, fmt.Errorf("cacheClient cannot be nil in NewConfigHandler")
	}
	if env == nil {
		return nil, fmt.Errorf("env cannot be nil in NewConfigHandler")
	}
	if obs == nil {
		return nil, fmt.Errorf("observability stack cannot be nil in NewConfigHandler")
	}

	return &ConfigHandler{
		cacheClient: cacheClient,
		env:         env,
		obs:         obs,
	}, nil
}

// SetReinitCallback sets the callback to be executed when configurations are updated
func (h *ConfigHandler) SetReinitCallback(callback ReinitializeCallback) {
	h.reinitCallback = callback
}

// UpdateConfigurations handles configuration update requests
func (h *ConfigHandler) UpdateConfigurations(ctx *fiber.Ctx) error {
	functionName := "UpdateConfigurations"
	functionFailed := "UpdateConfigurations_Failed"

	// Create context with timeout
	fiberCtx, cancel := context.WithTimeout(ctx.Context(), defaultTimeoutConfig)
	defer cancel()
	tCtx, span := h.obs.TracerService.StartTracer(fiberCtx, functionName)
	defer h.obs.TracerService.StopSpan(span)

	// Generate request ID for correlation
	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())

	h.obs.MetricsService.IncrementCounter(tCtx, functionName, 1, nil)

	// Set tracing attributes - only include request ID
	h.obs.TracerService.SetAttributes(span, map[string]string{
		"request_id": requestID,
	})

	h.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] STEP: Request received", requestID))

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
		h.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] STATUS: Bad request parsing: %v", requestID, err))
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{
			"error": "parse_error",
		})

		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	// Check if we have values
	if wrapper.Values == nil {
		h.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] STATUS: No values in request", requestID))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "No configuration values found",
		})
	}

	// Use the parsed config directly
	cfg := wrapper.Values

	h.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] STEP: Parsed configuration: %+v", requestID, cfg))

	// Set the new configuration
	if err := config.SetConfig(cfg); err != nil {
		h.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] STATUS: Configuration error: %v", requestID, err))
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{
			"error": "config_error",
		})

		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	h.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] STEP: Saving to cache", requestID))
	// Save to cache if available
	err := h.cacheClient.SetDataToCache(tCtx, "config", cfg)
	if err != nil {
		h.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] STATUS: Cache error", requestID))
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{
			"error": "cache_error",
		})
	}

	// Execute the reinitialization callback if it's been set
	if h.reinitCallback != nil {
		h.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] STEP: Reinitializing services", requestID))
		if err := h.reinitCallback(cfg); err != nil {
			h.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] STATUS: Reinitialization failed", requestID))
			h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{
				"error": "reinit_error",
			})

			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Error reinitializing services: " + err.Error(),
			})
		}
		h.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] STATUS: Reinitialization successful", requestID))
	} else {
		h.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] STATUS: No reinitialization needed", requestID))
	}

	h.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] STATUS: Success", requestID))
	h.obs.MetricsService.IncrementCounter(tCtx, "operation_success", 1, nil)

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Configuration updated successfully",
	})
}

// Health provides a health check endpoint
func (h *ConfigHandler) Health(ctx *fiber.Ctx) error {
	functionName := "Health"

	// Create context with timeout
	tCtx, cancel := context.WithTimeout(ctx.Context(), defaultTimeoutConfig)
	defer cancel()

	// Generate request ID for correlation
	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())

	// Start tracing
	tCtx, span := h.obs.TracerService.StartTracer(tCtx, functionName)
	defer h.obs.TracerService.StopSpan(span)

	// Set tracing attributes - only include request ID
	h.obs.TracerService.SetAttributes(span, map[string]string{
		"request_id": requestID,
	})

	h.obs.MetricsService.IncrementCounter(tCtx, functionName, 1, nil)
	h.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] STEP: Health check", requestID))
	h.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] STATUS: Healthy", requestID))
	h.obs.MetricsService.IncrementCounter(tCtx, "operation_success", 1, nil)

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "OK",
		"success": true,
	})
}
