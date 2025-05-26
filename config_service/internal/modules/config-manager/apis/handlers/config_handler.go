package handler

import (
	"context"
	"fmt"
	"nps-config-service/internal/constants"
	"nps-config-service/internal/modules/config-manager/services"
	"nps-config-service/internal/modules/config-manager/utils"
	"nps-config-service/pkg/observability"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type ConfigHandler struct {
	Service            services.IConfigService
	ObservabilityStack *observability.ObservabilityStack
}

func NewConfigHandler(service services.IConfigService, observabilityStack *observability.ObservabilityStack) *ConfigHandler {
	if observabilityStack == nil {
		panic("ObservabilityStack cannot be nil")
	}

	// Initialize metrics

	return &ConfigHandler{
		Service:            service,
		ObservabilityStack: observabilityStack,
	}
}

func (h *ConfigHandler) recordMetrics(ctx context.Context, operation string, start time.Time, err error) {
	// Record request count
	h.ObservabilityStack.MetricsService.IncrementCounter(ctx, constants.ConfigHandlerRequestCounterMetric, 1, map[string]string{
		"operation": operation,
		"service":   "config_handler",
	})

	// Record request latency
	latency := time.Since(start).Seconds()
	h.ObservabilityStack.MetricsService.RecordHistogram(ctx, constants.ConfigHandlerRequestLatencyMetric, latency, map[string]string{
		"operation": operation,
		"service":   "config_handler",
	})

	// Record error if any
	if err != nil {
		h.ObservabilityStack.MetricsService.IncrementCounter(ctx, constants.ConfigHandlerErrorCounterMetric, 1, map[string]string{
			"operation": operation,
			"service":   "config_handler",
			"error":     err.Error(),
		})
	}
}

func (h *ConfigHandler) StoreConfigHandler(c *fiber.Ctx) error {
	start := time.Now()
	ctx := c.UserContext()
	ctx, span := h.ObservabilityStack.TracerService.Start(ctx, "ConfigHandler.StoreConfig")
	defer span.End()
	defer func() {
		h.recordMetrics(ctx, "store_config", start, nil)
	}()

	environment := c.Params("environment")
	serviceName := c.Params("service")

	span.SetAttributes(
		attribute.String("environment", environment),
		attribute.String("service", serviceName),
	)

	h.ObservabilityStack.Logger.InfoContext(ctx, "Processing store config request",
		"environment", environment,
		"service", serviceName)

	contextData := c.Locals("contextData")
	configData, ok := contextData.(*map[string]any)
	if !ok {
		err := fmt.Errorf("invalid config data")
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error.code", "CNF001"))
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Invalid config data",
			"error_code", "CNF001",
			"environment", environment,
			"service", serviceName)
		h.recordMetrics(ctx, "store_config", start, err)
		utils.SendError(c, fiber.StatusBadRequest, "Invalid config data", err.Error())
		return nil
	}

	responseData, responseError := h.Service.StoreConfigService(ctx, environment, serviceName, *configData)
	if responseError != nil {
		span.SetStatus(codes.Error, responseError.ErrorMessage)
		span.SetAttributes(
			attribute.String("error.code", responseError.ErrorCode),
			attribute.Int("error.status_code", int(responseError.StatusCode)),
		)
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to store config",
			"error", responseError.ErrorMessage,
			"error_code", responseError.ErrorCode,
			"status_code", responseError.StatusCode,
			"environment", environment,
			"service", serviceName)
		h.recordMetrics(ctx, "store_config", start, responseError)
		utils.SendError(c, fiber.StatusInternalServerError, "Failed to store config", responseError.ErrorMessage)
		return nil
	}

	span.SetStatus(codes.Ok, "Config stored successfully")
	h.ObservabilityStack.Logger.InfoContext(ctx, "Config stored successfully",
		"environment", environment,
		"service", serviceName)
	utils.SendSuccess(c, responseData.StatusCode, responseData.Message, responseData.Data)
	return nil
}

func (h *ConfigHandler) GetfullConfig(c *fiber.Ctx) error {
	start := time.Now()
	ctx := c.UserContext()
	ctx, span := h.ObservabilityStack.TracerService.Start(ctx, "ConfigHandler.GetFullConfig")
	defer span.End()
	defer func() {
		h.recordMetrics(ctx, "get_full_config", start, nil)
	}()

	environment := c.Params("environment")
	serviceName := c.Params("service")

	span.SetAttributes(
		attribute.String("environment", environment),
		attribute.String("service", serviceName),
	)

	h.ObservabilityStack.Logger.InfoContext(ctx, "Processing get full config request",
		"environment", environment,
		"service", serviceName)

	config, err := h.Service.GetConfigService(ctx, serviceName, environment)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to fetch config",
			"error", err.Error(),
			"environment", environment,
			"service", serviceName)
		h.recordMetrics(ctx, "get_full_config", start, err)
		utils.SendError(c, fiber.StatusInternalServerError, "Failed to fetch config", err.Error())
		return nil
	}

	span.SetStatus(codes.Ok, "Config fetched successfully")
	h.ObservabilityStack.Logger.InfoContext(ctx, "Config fetched successfully",
		"environment", environment,
		"service", serviceName)
	utils.SendSuccess(c, fiber.StatusOK, "Config fetched successfully", config)
	return nil
}

func (h *ConfigHandler) GetByValue(c *fiber.Ctx) error {
	start := time.Now()
	ctx := c.UserContext()
	ctx, span := h.ObservabilityStack.TracerService.Start(ctx, "ConfigHandler.GetByValue")
	defer span.End()
	defer func() {
		h.recordMetrics(ctx, "get_by_value", start, nil)
	}()

	environment := c.Params("environment")
	serviceName := c.Params("service")
	key := c.Params("key")

	span.SetAttributes(
		attribute.String("environment", environment),
		attribute.String("service", serviceName),
		attribute.String("key", key),
	)

	h.ObservabilityStack.Logger.InfoContext(ctx, "Processing get by value request",
		"environment", environment,
		"service", serviceName,
		"key", key)

	value, err := h.Service.GetConfigValueService(ctx, serviceName, environment, key)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to get config value",
			"error", err.Error(),
			"environment", environment,
			"service", serviceName,
			"key", key)
		h.recordMetrics(ctx, "get_by_value", start, err)
		utils.SendError(c, fiber.StatusInternalServerError, "Failed to get config value", err.Error())
		return nil
	}

	span.SetStatus(codes.Ok, "Value fetched successfully")
	h.ObservabilityStack.Logger.InfoContext(ctx, "Value fetched successfully",
		"environment", environment,
		"service", serviceName,
		"key", key)
	utils.SendSuccess(c, fiber.StatusOK, "Value fetched successfully", fiber.Map{key: value})
	return nil
}
