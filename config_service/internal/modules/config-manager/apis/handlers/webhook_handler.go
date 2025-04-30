package handler

import (
	"log"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/utils"

	common "nps-config-service/internal/common/errors"
	"nps-config-service/internal/modules/config-manager/services"

	"github.com/gofiber/fiber/v2"
)

type WebhookHandler struct {
	Service services.IWebhookService
}

func NewWebhookHandler(service services.IWebhookService) *WebhookHandler {
	return &WebhookHandler{Service: service}
}

func (h *WebhookHandler) RegisterWebhook(c *fiber.Ctx) error {
	configData, ok := c.Locals("contextData").(*dtos.RegisterWebhookRequest)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ApiResponseDto{
			Success: false,
			Error: &dtos.ErrorResponseDto{
				Code:    common.Errors["CNF004"],
				Message: "Invalid request body. Please make sure it's a valid JSON.",
			},
		})
	}
	res, err := h.Service.RegisterWebhookService(*configData)
	if err != nil {
		utils.SendError(c, int(err.StatusCode), err.ErrorCode, err.ErrorMessage)
		return nil
	}
	log.Printf("Successfully registered webhook: %v", res)
	utils.SendSuccess(c, res.StatusCode, "Webhook registered successfully", res.Data)
	return nil
}

func (h *WebhookHandler) GetWebhooks(c *fiber.Ctx) error {
	env := c.Params("environment")
	service := c.Params("service")
	res, err := h.Service.GetWebhooks(env, service)
	if err != nil {
		log.Printf("Error retrieving webhooks for environment %s and service %s: %v", env, service, err)

		utils.SendError(c, int(err.StatusCode), err.ErrorCode, err.ErrorMessage)

		return nil
	}

	log.Printf("Successfully retrieved webhooks for environment %s and service %s: %v", env, service, res)
	utils.SendSuccess(c, fiber.StatusOK, "Webhooks retrieved successfully", res)
	return nil
}

func (h *WebhookHandler) DeleteWebhook(c *fiber.Ctx) error {
	env := c.Params("environment")
	service := c.Params("service")

	contextData := c.Locals("contextData")
	bodyMap, ok := contextData.(*map[string]any)
	if !ok {
		log.Printf("Invalid context data for delete request: %v", contextData)
		utils.SendError(c, fiber.StatusBadRequest, "Invalid delete request", "Context data missing or invalid")
		return nil
	}

	var deleteReq dtos.RegisterWebhookRequest
	if err := utils.MapToStruct(*bodyMap, &deleteReq); err != nil {
		log.Printf("Error mapping context data to delete request struct: %v", err)
		utils.SendError(c, fiber.StatusBadRequest, "Invalid delete structure", err.Error())
		return nil
	}

	res, err := h.Service.DeleteWebhook(env, service, deleteReq.URL, deleteReq.Method)
	if err != nil {
		log.Printf("Error deleting webhook for environment %s and service %s: %v", env, service, err)
		utils.SendError(c, int(err.StatusCode), err.ErrorCode, err.ErrorMessage)
		return nil
	}

	log.Printf("Successfully deleted webhook: %v", res)
	utils.SendSuccess(c, fiber.StatusOK, "Webhook deleted successfully", res)
	return nil
}
