package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/repositories"
	"time"

	"strings"
)

const (
	maxRetries     = 3
	retryDelay     = 2 * time.Second
	requestTimeout = 5 * time.Second
)

type WebhookService struct {
	Repo repositories.IConfigRepo
}

type IWebhookService interface {
	RegisterWebhookService(req dtos.RegisterWebhookRequest) (*dtos.SuccessResponse, *dtos.ServiceErrorResponse)
	GetWebhooks(env, service string) ([]dtos.RegisterWebhookRequest, *dtos.ServiceErrorResponse)
	DeleteWebhook(env, service, url, method string) (string, *dtos.ServiceErrorResponse)
	DeleteAllWebhooks(env, service string) error
	NotifyWebhook(hook dtos.RegisterWebhookRequest, data map[string]interface{})
}

func NewWebhookService(repo repositories.IConfigRepo) IWebhookService {
	return &WebhookService{Repo: repo}
}

func (s *WebhookService) RegisterWebhookService(req dtos.RegisterWebhookRequest) (*dtos.SuccessResponse, *dtos.ServiceErrorResponse) {
	key := fmt.Sprintf("/webhooks/%s/%s", req.Environment, req.ServiceName)
	ctx := context.Background()
	var hooks []dtos.RegisterWebhookRequest

	webHook, err := s.Repo.Get(ctx, key)
	if err == nil && len(webHook) > 0 {
		if err := json.Unmarshal([]byte(webHook), &hooks); err != nil {
			return nil, &dtos.ServiceErrorResponse{
				StatusCode:   500,
				ErrorCode:    "Failed to parse webhook data",
				ErrorMessage: fmt.Sprintf("invalid webhook data stored for %s: %v", key, err),
			}
		}
	}

	for _, existing := range hooks {
		if existing.URL == req.URL && existing.Method == req.Method {
			return nil, &dtos.ServiceErrorResponse{
				StatusCode:   409,
				ErrorCode:    "Webhook already exists",
				ErrorMessage: fmt.Sprintf("webhook with URL '%s' and method '%s' already exists", req.URL, req.Method),
			}
		}
	}

	hooks = append(hooks, req)
	data, err := json.Marshal(hooks)
	if err != nil {
		return nil, &dtos.ServiceErrorResponse{
			StatusCode:   500,
			ErrorCode:    "Failed to marshal webhook data",
			ErrorMessage: fmt.Sprintf("failed to marshal webhook data for %s: %v", key, err),
		}
	}
	if err := s.Repo.Set(ctx, key, string(data), -1); err != nil {
		return nil, &dtos.ServiceErrorResponse{
			StatusCode:   500,
			ErrorCode:    "Failed to store webhook data",
			ErrorMessage: fmt.Sprintf("failed to store webhook data for %s: %v", key, err),
		}
	}
	response := &dtos.SuccessResponse{
		StatusCode: 201,
		Message:    "Webhook registered successfully",
		Data:       map[string]interface{}{"webhook": req},
	}
	// res := fmt.Sprintf("Webhook registered successfully for service: %s in environment: %s", req.ServiceName, req.Environment)
	return response, nil
}
func (s *WebhookService) GetWebhooks(env, service string) ([]dtos.RegisterWebhookRequest, *dtos.ServiceErrorResponse) {
	key := fmt.Sprintf("/webhooks/%s/%s", env, service)
	ctx := context.Background()

	webHook, err := s.Repo.Get(ctx, key)
	if err != nil || len(webHook) == 0 {
		return nil, &dtos.ServiceErrorResponse{
			StatusCode:   404,
			ErrorCode:    "No webhooks found",
			ErrorMessage: fmt.Sprintf("no webhooks found for %s/%s: %v", env, service, err),
		}
	}

	var hooks []dtos.RegisterWebhookRequest
	if err := json.Unmarshal([]byte(webHook), &hooks); err != nil {
		return nil, &dtos.ServiceErrorResponse{
			StatusCode:   500,
			ErrorCode:    "Failed to parse webhook data",
			ErrorMessage: fmt.Sprintf("failed to parse webhook data for %s/%s: %v", env, service, err),
		}
	}
	return hooks, nil
}
func (s *WebhookService) DeleteWebhook(env, service, url, method string) (string, *dtos.ServiceErrorResponse) {
	key := fmt.Sprintf("/webhooks/%s/%s", env, service)
	ctx := context.Background()

	hooks, errService := s.GetWebhooks(env, service)
	if errService != nil {
		return "", &dtos.ServiceErrorResponse{
			StatusCode:   404,
			ErrorCode:    "No webhooks found",
			ErrorMessage: fmt.Sprintf("no webhooks found for %s/%s: %v", env, service, errService),
		}
	}
	updated := make([]dtos.RegisterWebhookRequest, 0)
	found := false
	for _, h := range hooks {
		if h.URL == url && h.Method == method {
			found = true
			continue // skip the one to delete
		}
		updated = append(updated, h)
	}
	if !found {
		return "", &dtos.ServiceErrorResponse{
			StatusCode:   404,
			ErrorCode:    "Webhook not found",
			ErrorMessage: fmt.Sprintf("webhook with URL '%s' and method '%s' not found", url, method),
		}
	}

	// Re-save the updated list
	data, err := json.Marshal(updated)
	if err != nil {
		return "", &dtos.ServiceErrorResponse{
			StatusCode:   500,
			ErrorCode:    "Failed to marshal updated webhooks",
			ErrorMessage: fmt.Sprintf("failed to marshal updated webhooks for %s/%s: %v", env, service, err),
		}
	}
	if err := s.Repo.Set(ctx, key, string(data), -1); err != nil {
		return "", &dtos.ServiceErrorResponse{
			StatusCode:   500,
			ErrorCode:    "Failed to store updated webhooks",
			ErrorMessage: fmt.Sprintf("failed to store updated webhooks for %s/%s: %v", env, service, err),
		}
	}
	return "Webhook deleted successfully", nil
}
func (s *WebhookService) DeleteAllWebhooks(env, service string) error {
	key := fmt.Sprintf("/webhooks/%s/%s", env, service)
	ctx := context.Background()
	return s.Repo.Delete(ctx, key)
}
func (s *WebhookService) NotifyWebhook(hook dtos.RegisterWebhookRequest, data map[string]interface{}) {
	body, err := json.Marshal(map[string]interface{}{
		"values":      data,
		"method":      hook.Method,
		"environment": hook.Environment,
		"serviceName": hook.ServiceName,
	})
	if err != nil {
		log.Printf("Failed to marshal webhook payload: %v", err)
		return
	}
	// fullURL := strings.TrimRight(hook.URL, "/")
	fullURL := fmt.Sprintf("%s/%s/%s", strings.TrimRight(hook.URL, "/"), hook.Environment, hook.ServiceName)
	if !strings.HasPrefix(fullURL, "http://") && !strings.HasPrefix(fullURL, "https://") {
		fullURL = "http://" + fullURL
	}
	for attempt := 1; attempt <= maxRetries; attempt++ {
		client := http.Client{Timeout: requestTimeout}
		resp, err := client.Post(fullURL, "application/json", bytes.NewBuffer(body))

		if err != nil {
			log.Printf("Attempt %d: Failed to notify %s: %v", attempt, fullURL, err)
		} else {
			defer resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				log.Printf("Webhook to %s succeeded with status: %s", fullURL, resp.Status)
				return
			}
			log.Printf("Attempt %d: Webhook to %s failed with status: %s", attempt, fullURL, resp.Status)
		}

		// Wait before retrying
		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	log.Printf("All attempts to notify webhook %s failed after %d retries", fullURL, maxRetries)
}
