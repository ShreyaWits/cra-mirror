package handler

import (
	"log"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/utils"

	customErr "nps-config-service/internal/common/errors"
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
	contextData := c.Locals("contextData")
	bodyMap, ok := contextData.(*map[string]any)
	if !ok {
		log.Printf("Invalid context data: %v", contextData)
		utils.SendError(c, fiber.StatusBadRequest, "Invalid config data", "Context data missing or invalid")
		return nil
	}

	var configData dtos.RegisterWebhookRequest
	if err := utils.MapToStruct(*bodyMap, &configData); err != nil {
		log.Printf("Error mapping context data to struct: %v", err)
		utils.SendError(c, fiber.StatusBadRequest, "Invalid config structure", err.Error())
		return nil
	}

	res, err := h.Service.RegisterWebhookService(configData)
	if err != nil {
		if conflictErr, ok := err.(*customErr.ConflictError); ok {
			log.Printf("Conflict error while registering webhook: %v", conflictErr)
			utils.SendError(c, fiber.StatusConflict, "Webhook already exists", conflictErr.Error())
		} else {
			log.Printf("Internal error while registering webhook: %v", err)
			utils.SendError(c, fiber.StatusInternalServerError, "Failed to register webhook", err.Error())
		}
		return nil
	}

	log.Printf("Successfully registered webhook: %v", res)
	utils.SendSuccess(c, fiber.StatusOK, "Webhook registered successfully", res)
	return nil
}

func (h *WebhookHandler) GetWebhooks(c *fiber.Ctx) error {
	env := c.Params("environment")
	service := c.Params("service")
	res, err := h.Service.GetWebhooks(env, service)
	if err != nil {
		log.Printf("Error retrieving webhooks for environment %s and service %s: %v", env, service, err)
		utils.SendError(c, fiber.StatusNotFound, "Failed to retrieve webhooks", err.Error())
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
		if deleteErr, ok := err.(*customErr.ConflictError); ok {
			log.Printf("Webhook not found for delete request: %v", deleteErr)
			utils.SendError(c, fiber.StatusNotFound, "Webhook not found", deleteErr.Error())
		} else {
			log.Printf("Internal error while deleting webhook: %v", err)
			utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete webhook", err.Error())
		}
		return nil
	}

	log.Printf("Successfully deleted webhook: %v", res)
	utils.SendSuccess(c, fiber.StatusOK, "Webhook deleted successfully", res)
	return nil
}
