package handler

import (
	"context"
	common "nps-config-service/internal/common/errors"
	"nps-config-service/internal/configs"
	"nps-config-service/internal/constants"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/services"
	"nps-config-service/pkg/observability"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
}

func NewAdminHandler(service services.IAdminService, observabilityStack *observability.ObservabilityStack, configProvider ConfigProvider) *AdminHandler {
	if observabilityStack == nil {
		panic("ObservabilityStack cannot be nil")
	}
	if configProvider == nil {
		configProvider = &DefaultConfigProvider{}
	}

	return &AdminHandler{
		Service:            service,
		ObservabilityStack: observabilityStack,
		ConfigProvider:     configProvider,
	}
}

func (h *AdminHandler) recordMetrics(ctx context.Context, operation string, start time.Time, err error) {
	// Record request count
	h.ObservabilityStack.MetricsService.IncrementCounter(ctx, constants.AdminHandlerRequestCounterMetric, 1, map[string]string{
		"operation": operation,
		"service":   "admin_handler",
	})

	// Record request latency
	latency := time.Since(start).Seconds()
	h.ObservabilityStack.MetricsService.RecordHistogram(ctx, constants.AdminHandlerRequestLatencyMetric, latency, map[string]string{
		"operation": operation,
		"service":   "admin_handler",
	})

	// Record error if any
	if err != nil {
		h.ObservabilityStack.MetricsService.IncrementCounter(ctx, constants.AdminHandlerErrorCounterMetric, 1, map[string]string{
			"operation": operation,
			"service":   "admin_handler",
		})
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
		err := common.ThrowError(fiber.StatusBadRequest, "CNF004")
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error.code", "CNF004"))
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Invalid request body for admin creation",
			"error_code", "CNF004")
		h.recordMetrics(ctx, "create_admin", start, err)
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	// Validate request data
	if err := h.validateAdminSignupRequest(ctx, contextData); err != nil {
		span.SetStatus(codes.Error, err.Error())
		h.recordMetrics(ctx, "create_admin", start, err)
		if appErr, ok := err.(common.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(fiber.StatusBadRequest).JSON(err)
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
		err := common.ThrowError(fiber.StatusBadRequest, "CNF004")
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error.code", "CNF004"))
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Invalid request body for admin fetch",
			"error_code", "CNF004")
		h.recordMetrics(ctx, "fetch_admin", start, err)
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	// Validate request data
	if err := h.validateAdminLoginRequest(ctx, contextData); err != nil {
		span.SetStatus(codes.Error, err.Error())
		h.recordMetrics(ctx, "fetch_admin", start, err)
		if appErr, ok := err.(common.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	jwtSecret := h.ConfigProvider.GetJWTSecret()
	span.SetAttributes(
		attribute.String("username", contextData.Username),
	)

	responseData, responseError := h.Service.FetchAdminService(ctx, contextData, jwtSecret)
	if responseError != nil {
		span.SetStatus(codes.Error, responseError.Error())
		span.SetAttributes(attribute.String("error.code", "ADMIN002"))
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to fetch admin",
			"error", responseError,
			"error_code", "ADMIN002",
			"username", contextData.Username)
		h.recordMetrics(ctx, "fetch_admin", start, responseError)
		return c.Status(fiber.StatusUnauthorized).JSON(common.ThrowError(fiber.StatusUnauthorized, "ADMIN002"))
	}

	span.SetStatus(codes.Ok, "Admin fetched successfully")
	h.ObservabilityStack.Logger.InfoContext(ctx, "Admin fetched successfully",
		"username", contextData.Username)
	return c.Status(fiber.StatusOK).JSON(responseData)
}

// ListAdminsHandler returns a list of all admin users
func (h *AdminHandler) ListAdminsHandler(c *fiber.Ctx) error {
	start := time.Now()
	ctx := c.UserContext()
	ctx, span := h.ObservabilityStack.TracerService.Start(ctx, "AdminHandler.ListAdmins")
	defer span.End()
	defer func() {
		h.recordMetrics(ctx, "list_admins", start, nil)
	}()

	h.ObservabilityStack.Logger.InfoContext(ctx, "Processing list admins request")

	responseData, responseError := h.Service.ListAdminsService(ctx)
	if responseError != nil {
		span.SetStatus(codes.Error, responseError.Error())
		span.SetAttributes(attribute.String("error.code", "ADMIN007"))
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to list admin users",
			"error", responseError,
			"error_code", "ADMIN007")
		h.recordMetrics(ctx, "list_admins", start, responseError)
		return c.Status(fiber.StatusInternalServerError).JSON(common.ThrowError(fiber.StatusInternalServerError, "ADMIN007"))
	}

	span.SetStatus(codes.Ok, "Admin users listed successfully")
	h.ObservabilityStack.Logger.InfoContext(ctx, "Successfully listed admin users",
		"count", len(responseData.Admins))
	return c.Status(fiber.StatusOK).JSON(responseData)
}

// DeleteAdminHandler deletes an admin user
func (h *AdminHandler) DeleteAdminHandler(c *fiber.Ctx) error {
	start := time.Now()
	ctx := c.UserContext()
	ctx, span := h.ObservabilityStack.TracerService.Start(ctx, "AdminHandler.DeleteAdmin")
	defer span.End()
	defer func() {
		h.recordMetrics(ctx, "delete_admin", start, nil)
	}()

	username := c.Params("username")
	if username == "" {
		err := common.ThrowError(fiber.StatusBadRequest, "ADMIN008")
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error.code", "ADMIN008"))
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Username is required",
			"error_code", "ADMIN008")
		h.recordMetrics(ctx, "delete_admin", start, err)
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	// Prevent self-deletion
	currentUsername := c.Locals("username").(string)
	if username == currentUsername {
		err := common.ThrowError(fiber.StatusForbidden, "ADMIN009")
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error.code", "ADMIN009"))
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Cannot delete own account",
			"error_code", "ADMIN009",
			"username", username)
		h.recordMetrics(ctx, "delete_admin", start, err)
		return c.Status(fiber.StatusForbidden).JSON(err)
	}

	span.SetAttributes(
		attribute.String("username", username),
	)

	responseData, responseError := h.Service.DeleteAdminService(ctx, username)
	if responseError != nil {
		span.SetStatus(codes.Error, responseError.Error())
		span.SetAttributes(attribute.String("error.code", "ADMIN010"))
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to delete admin user",
			"error", responseError,
			"error_code", "ADMIN010",
			"username", username)
		h.recordMetrics(ctx, "delete_admin", start, responseError)
		if appErr, ok := responseError.(common.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusInternalServerError, "ADMIN010"))
	}

	span.SetStatus(codes.Ok, "Admin user deleted successfully")
	h.ObservabilityStack.Logger.InfoContext(ctx, "Successfully deleted admin user",
		"username", username)
	return c.Status(fiber.StatusOK).JSON(responseData)
}
