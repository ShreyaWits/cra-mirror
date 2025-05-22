package handler

import (
	"fmt"
	common "nps-config-service/internal/common/errors"
	"nps-config-service/internal/configs"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/services"
	"nps-config-service/pkg/observability"
	"os"

	"github.com/gofiber/fiber/v2"
)

type AdminHandler struct {
	Service            services.IAdminService
	ObservabilityStack *observability.ObservabilityStack
}

func NewAdminHandler(service services.IAdminService, observabilityStack *observability.ObservabilityStack) *AdminHandler {
	return &AdminHandler{Service: service, ObservabilityStack: observabilityStack}
}

func (h *AdminHandler) CreateAdminHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()// Get the context.

	contextData, ok := c.Locals("contextData").(*dtos.AdminSignupDto)

	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "CNF004"))
	}
	if contextData.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "CNF012")) // Username is required
	}
	if contextData.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "CNF013")) // Password is required
	}
	if contextData.Secret == "" {
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "CNF014")) // Secret is required
	}
	// Load the admin secret from .env
	adminSecret := configs.AppConfig.AdminSecret
	if adminSecret == "" {
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "CNF005")) // Admin secret not configured
	}

	if contextData.Secret != adminSecret {
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "CNF010")) // Invalid secret
	}

	responseData, responseError := h.Service.CreateAdminService(ctx, contextData)
	if responseError != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "ADMIN001")) // Failed to create admin
	}
	return c.Status(fiber.StatusOK).JSON(responseData)
}

func (h *AdminHandler) FetchAdminHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()// Get the context.

	contextData, ok := c.Locals("contextData").(*dtos.AdminLoginDto)
	fmt.Print(contextData)

	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "CNF004"))
	}

	if contextData.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "CNF004")) // Username is required
	}

	if contextData.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "CNF004")) // Password is required
	}
	// Load the admin secret from .env
	// adminSecret := os.Getenv("ADMIN_SECRET")
	// if adminSecret == "" {
	// 	return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "CNF005")) // Admin secret not configured
	// }

	// if contextData.Secret != adminSecret {
	// 	return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "CNF010")) // Invalid secret
	// }

	// Load JWT secret
	jwtSecret := configs.AppConfig.JWTSecret
	if jwtSecret == "" {
		return c.Status(fiber.StatusBadRequest).JSON(common.ThrowError(fiber.StatusBadRequest, "CNF007"))
	}

	responseData, responseError := h.Service.FetchAdminService(ctx, contextData, jwtSecret)

	if responseError != nil {
		return responseError
	}
	return c.Status(fiber.StatusOK).JSON(responseData)
}
