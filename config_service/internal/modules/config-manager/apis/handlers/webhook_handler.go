package handler

import (
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/utils"
	"nps-config-service/pkg/observability"

	common "nps-config-service/internal/common/errors"
	"nps-config-service/internal/modules/config-manager/services"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type WebhookHandler struct {
	Service            services.IWebhookService
	ObservabilityStack *observability.ObservabilityStack
}

func NewWebhookHandler(service services.IWebhookService, observabilityStack *observability.ObservabilityStack) *WebhookHandler {
	if observabilityStack == nil {
		panic("ObservabilityStack cannot be nil")
	}
	return &WebhookHandler{Service: service, ObservabilityStack: observabilityStack}
}

func (h *WebhookHandler) RegisterWebhook(c *fiber.Ctx) error {
	ctx := c.UserContext()
	ctx, span := h.ObservabilityStack.TracerService.Start(ctx, "WebhookHandler.RegisterWebhook")
	defer span.End()

	h.ObservabilityStack.Logger.InfoContext(ctx, "Processing webhook registration request")

	configData, ok := c.Locals("contextData").(*dtos.RegisterWebhookRequest)
	if !ok {
		span.SetStatus(codes.Error, "Invalid request body")
		span.SetAttributes(attribute.String("error.code", common.Errors["CNF004"]))
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Invalid request body for webhook registration",
			"error_code", common.Errors["CNF004"])
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ApiResponseDto{
			Success: false,
			Error: &dtos.ErrorResponseDto{
				Code:    common.Errors["CNF004"],
				Message: "Invalid request body. Please make sure it's a valid JSON.",
			},
		})
	}

	span.SetAttributes(
		attribute.String("webhook.url", configData.URL),
		attribute.String("webhook.method", configData.Method),
	)

	res, err := h.Service.RegisterWebhookService(ctx, *configData)
	if err != nil {
		span.SetStatus(codes.Error, err.ErrorMessage)
		span.SetAttributes(
			attribute.String("error.code", err.ErrorCode),
			attribute.Int("error.status_code", int(err.StatusCode)),
		)
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to register webhook",
			"error", err.ErrorMessage,
			"error_code", err.ErrorCode,
			"status_code", err.StatusCode)
		utils.SendError(c, int(err.StatusCode), err.ErrorCode, err.ErrorMessage)
		return nil
	}

	span.SetStatus(codes.Ok, "Webhook registered successfully")
	h.ObservabilityStack.Logger.InfoContext(ctx, "Successfully registered webhook",
		"response", res)
	utils.SendSuccess(c, res.StatusCode, "Webhook registered successfully", res.Data)
	return nil
}

func (h *WebhookHandler) GetWebhooks(c *fiber.Ctx) error {
	ctx := c.UserContext()
	ctx, span := h.ObservabilityStack.TracerService.Start(ctx, "WebhookHandler.GetWebhooks")
	defer span.End()

	env := c.Params("environment")
	service := c.Params("service")

	span.SetAttributes(
		attribute.String("environment", env),
		attribute.String("service", service),
	)

	h.ObservabilityStack.Logger.InfoContext(ctx, "Retrieving webhooks",
		"environment", env,
		"service", service)

	res, err := h.Service.GetWebhooks(ctx, env, service)
	if err != nil {
		span.SetStatus(codes.Error, err.ErrorMessage)
		span.SetAttributes(
			attribute.String("error.code", err.ErrorCode),
			attribute.Int("error.status_code", int(err.StatusCode)),
		)
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to retrieve webhooks",
			"environment", env,
			"service", service,
			"error", err.ErrorMessage,
			"error_code", err.ErrorCode,
			"status_code", err.StatusCode)
		utils.SendError(c, int(err.StatusCode), err.ErrorCode, err.ErrorMessage)
		return nil
	}

	span.SetStatus(codes.Ok, "Webhooks retrieved successfully")
	h.ObservabilityStack.Logger.InfoContext(ctx, "Successfully retrieved webhooks",
		"environment", env,
		"service", service,
		"count", len(res))
	utils.SendSuccess(c, fiber.StatusOK, "Webhooks retrieved successfully", res)
	return nil
}

func (h *WebhookHandler) DeleteWebhook(c *fiber.Ctx) error {
	ctx := c.UserContext()
	ctx, span := h.ObservabilityStack.TracerService.Start(ctx, "WebhookHandler.DeleteWebhook")
	defer span.End()

	env := c.Params("environment")
	service := c.Params("service")

	span.SetAttributes(
		attribute.String("environment", env),
		attribute.String("service", service),
	)

	h.ObservabilityStack.Logger.InfoContext(ctx, "Processing webhook deletion request",
		"environment", env,
		"service", service)

	contextData := c.Locals("contextData")
	bodyMap, ok := contextData.(*map[string]any)
	if !ok {
		span.SetStatus(codes.Error, "Invalid context data")
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Invalid context data for delete request",
			"context_data", contextData)
		utils.SendError(c, fiber.StatusBadRequest, "Invalid delete request", "Context data missing or invalid")
		return nil
	}

	var deleteReq dtos.RegisterWebhookRequest
	if err := utils.MapToStruct(*bodyMap, &deleteReq); err != nil {
		span.SetStatus(codes.Error, err.Error())
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to map delete request data",
			"error", err.Error())
		utils.SendError(c, fiber.StatusBadRequest, "Invalid delete structure", err.Error())
		return nil
	}

	span.SetAttributes(
		attribute.String("webhook.url", deleteReq.URL),
		attribute.String("webhook.method", deleteReq.Method),
	)

	res, err := h.Service.DeleteWebhook(ctx, env, service, deleteReq.URL, deleteReq.Method)
	if err != nil {
		span.SetStatus(codes.Error, err.ErrorMessage)
		span.SetAttributes(
			attribute.String("error.code", err.ErrorCode),
			attribute.Int("error.status_code", int(err.StatusCode)),
		)
		h.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to delete webhook",
			"environment", env,
			"service", service,
			"error", err.ErrorMessage,
			"error_code", err.ErrorCode,
			"status_code", err.StatusCode)
		utils.SendError(c, int(err.StatusCode), err.ErrorCode, err.ErrorMessage)
		return nil
	}

	span.SetStatus(codes.Ok, "Webhook deleted successfully")
	h.ObservabilityStack.Logger.InfoContext(ctx, "Successfully deleted webhook",
		"environment", env,
		"service", service,
		"response", res)
	utils.SendSuccess(c, fiber.StatusOK, "Webhook deleted successfully", res)
	return nil
}
