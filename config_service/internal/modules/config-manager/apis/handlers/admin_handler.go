package handler

import (
	"context"
	"fmt"
	common "nps-config-service/internal/common/errors"
	"nps-config-service/internal/configs"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/services"
	"nps-config-service/pkg/observability"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

// ConfigProvider defines the interface for accessing configuration
type ConfigProvider interface {
	GetAdminSecret() string
	GetJWTSecret() string
}

// DefaultConfigProvider implements ConfigProvider using the global config
type DefaultConfigProvider struct{}

func (p *DefaultConfigProvider) GetAdminSecret() string {
	return configs.AppConfig.AdminSecret
}

func (p *DefaultConfigProvider) GetJWTSecret() string {
	return configs.AppConfig.JWTSecret
}

type AdminHandler struct {
	Service            services.IAdminService
	ObservabilityStack *observability.ObservabilityStack
	ConfigProvider     ConfigProvider
	// Metrics
	requestCounter   metric.Int64Counter
	requestLatency   metric.Float64Histogram
	errorCounter     metric.Int64Counter
	operationLatency metric.Float64Histogram
}

func NewAdminHandler(service services.IAdminService, observabilityStack *observability.ObservabilityStack, configProvider ConfigProvider) *AdminHandler {
	if observabilityStack == nil {
		panic("ObservabilityStack cannot be nil")
	}
	if configProvider == nil {
		configProvider = &DefaultConfigProvider{}
	}

	// Initialize metrics
	meter := observabilityStack.MetricsService
	requestCounter, _ := meter.Int64Counter(
		"admin_request_total",
		metric.WithDescription("Total number of admin requests"),
	)
	requestLatency, _ := meter.Float64Histogram(
		"admin_request_duration_seconds",
		metric.WithDescription("Admin request duration in seconds"),
	)
	errorCounter, _ := meter.Int64Counter(
		"admin_error_total",
		metric.WithDescription("Total number of admin errors"),
	)
	operationLatency, _ := meter.Float64Histogram(
		"admin_operation_duration_seconds",
		metric.WithDescription("Admin operation duration in seconds"),
	)

	return &AdminHandler{
		Service:            service,
		ObservabilityStack: observabilityStack,
		ConfigProvider:     configProvider,
		requestCounter:     requestCounter,
		requestLatency:     requestLatency,
		errorCounter:       errorCounter,
		operationLatency:   operationLatency,
	}
}

func (h *AdminHandler) recordMetrics(ctx context.Context, operation string, start time.Time, err error) {
	// Record request count
	h.requestCounter.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("operation", operation),
			attribute.String("service", "admin_handler"),
		),
	)

	// Record request latency
	latency := time.Since(start).Seconds()
	h.requestLatency.Record(ctx, latency,
		metric.WithAttributes(
			attribute.String("operation", operation),
			attribute.String("service", "admin_handler"),
		),
	)

	// Record operation latency
	h.operationLatency.Record(ctx, latency,
		metric.WithAttributes(
			attribute.String("operation", operation),
			attribute.String("service", "admin_handler"),
		),
	)

	// Record error if any
	if err != nil {
		h.errorCounter.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("operation", operation),
				attribute.String("service", "admin_handler"),
				attribute.String("error", err.Error()),
			),
		)
	}
}

func (h *AdminHandler) validateAdminSignupRequest(ctx context.Context, data *dtos.AdminSignupDto) error {
	if data.Username == "" {
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Username is required",
			"error_code", "CNF012")
		return common.ThrowError(fiber.StatusBadRequest, "CNF012")
	}
	if data.Password == "" {
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Password is required",
			"error_code", "CNF013")
		return common.ThrowError(fiber.StatusBadRequest, "CNF013")
	}
	if data.Secret == "" {
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Secret is required",
			"error_code", "CNF014")
		return common.ThrowError(fiber.StatusBadRequest, "CNF014")
	}

	adminSecret := h.ConfigProvider.GetAdminSecret()
	if adminSecret == "" {
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Admin secret not configured",
			"error_code", "CNF005")
		return common.ThrowError(fiber.StatusBadRequest, "CNF005")
	}

	if data.Secret != adminSecret {
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Invalid secret provided",
			"error_code", "CNF010")
		return common.ThrowError(fiber.StatusBadRequest, "CNF010")
	}

	return nil
}

func (h *AdminHandler) validateAdminLoginRequest(ctx context.Context, data *dtos.AdminLoginDto) error {
	if data.Username == "" {
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Username is required",
			"error_code", "CNF004")
		return common.ThrowError(fiber.StatusBadRequest, "CNF004")
	}
	if data.Password == "" {
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Password is required",
			"error_code", "CNF004")
		return common.ThrowError(fiber.StatusBadRequest, "CNF004")
	}

	jwtSecret := h.ConfigProvider.GetJWTSecret()
	if jwtSecret == "" {
		h.ObservabilityStack.Logger.ErrorContext(ctx, "JWT secret not configured",
			"error_code", "CNF007")
		return common.ThrowError(fiber.StatusBadRequest, "CNF007")
	}

	return nil
}

func (h *AdminHandler) CreateAdminHandler(c *fiber.Ctx) error {
	start := time.Now()
	ctx := c.UserContext()
	ctx, span := h.ObservabilityStack.TracerService.Start(ctx, "AdminHandler.CreateAdmin")
	defer span.End()
	defer func() {
		h.recordMetrics(ctx, "create_admin", start, nil)
	}()

	h.ObservabilityStack.Logger.InfoContext(ctx, "Processing admin creation request")

	contextData, ok := c.Locals("contextData").(*dtos.AdminSignupDto)
	if !ok {
		err := fmt.Errorf("invalid request body")
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error.code", "CNF004"))
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Invalid request body for admin creation",
			"error_code", "CNF004")
		h.recordMetrics(ctx, "create_admin", start, err)
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "CNF004"))
	}

	// Validate request data
	if err := h.validateAdminSignupRequest(ctx, contextData); err != nil {
		span.SetStatus(codes.Error, err.Error())
		h.recordMetrics(ctx, "create_admin", start, err)
		return err
	}

	span.SetAttributes(
		attribute.String("username", contextData.Username),
	)

	responseData, responseError := h.Service.CreateAdminService(ctx, contextData)
	if responseError != nil {
		span.SetStatus(codes.Error, responseError.Error())
		span.SetAttributes(attribute.String("error.code", "ADMIN001"))
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to create admin",
			"error", responseError,
			"error_code", "ADMIN001")
		h.recordMetrics(ctx, "create_admin", start, responseError)
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "ADMIN001"))
	}

	span.SetStatus(codes.Ok, "Admin created successfully")
	h.ObservabilityStack.Logger.InfoContext(ctx, "Admin created successfully",
		"username", contextData.Username,
		"admin_id", responseData.AdminId)
	return c.Status(fiber.StatusOK).JSON(responseData)
}

func (h *AdminHandler) FetchAdminHandler(c *fiber.Ctx) error {
	start := time.Now()
	ctx := c.UserContext()
	ctx, span := h.ObservabilityStack.TracerService.Start(ctx, "AdminHandler.FetchAdmin")
	defer span.End()
	defer func() {
		h.recordMetrics(ctx, "fetch_admin", start, nil)
	}()

	h.ObservabilityStack.Logger.InfoContext(ctx, "Processing admin fetch request")

	contextData, ok := c.Locals("contextData").(*dtos.AdminLoginDto)
	if !ok {
		err := fmt.Errorf("invalid request body")
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error.code", "CNF004"))
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Invalid request body for admin fetch",
			"error_code", "CNF004")
		h.recordMetrics(ctx, "fetch_admin", start, err)
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "CNF004"))
	}

	// Validate request data
	if err := h.validateAdminLoginRequest(ctx, contextData); err != nil {
		span.SetStatus(codes.Error, err.Error())
		h.recordMetrics(ctx, "fetch_admin", start, err)
		return err
	}

	jwtSecret := h.ConfigProvider.GetJWTSecret()
	span.SetAttributes(
		attribute.String("username", contextData.Username),
	)

	responseData, responseError := h.Service.FetchAdminService(ctx, contextData, jwtSecret)
	if responseError != nil {
		span.SetStatus(codes.Error, responseError.Error())
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to fetch admin",
			"error", responseError,
			"username", contextData.Username)
		h.recordMetrics(ctx, "fetch_admin", start, responseError)
		return responseError
	}

	span.SetStatus(codes.Ok, "Admin fetched successfully")
	h.ObservabilityStack.Logger.InfoContext(ctx, "Admin fetched successfully",
		"username", contextData.Username)
	return c.Status(fiber.StatusOK).JSON(responseData)
}
