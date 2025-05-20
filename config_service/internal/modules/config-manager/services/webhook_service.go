package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/repositories"
	"nps-config-service/pkg/observability"
	"time"

	"strings"
)

const (
	maxRetries     = 3
	retryDelay     = 2 * time.Second
	requestTimeout = 5 * time.Second
)

type WebhookService struct {
	Repo               repositories.IConfigRepo
	ObservabilityStack *observability.ObservabilityStack
}

type IWebhookService interface {
	RegisterWebhookService(ctx context.Context, req dtos.RegisterWebhookRequest) (*dtos.SuccessResponse, *dtos.ServiceErrorResponse)
	GetWebhooks(ctx context.Context, env, service string) ([]dtos.RegisterWebhookRequest, *dtos.ServiceErrorResponse)
	DeleteWebhook(ctx context.Context, env, service, url, method string) (string, *dtos.ServiceErrorResponse)
	DeleteAllWebhooks(ctx context.Context, env, service string) error
	NotifyWebhook(ctx context.Context, hook dtos.RegisterWebhookRequest, data map[string]interface{})
}

func NewWebhookService(repo repositories.IConfigRepo, ObservabilityStack *observability.ObservabilityStack) IWebhookService {
	return &WebhookService{Repo: repo, ObservabilityStack: ObservabilityStack}
}

func (s *WebhookService) RegisterWebhookService(ctx context.Context, req dtos.RegisterWebhookRequest) (*dtos.SuccessResponse, *dtos.ServiceErrorResponse) {
	ctx, span := s.ObservabilityStack.TracerService.Start(ctx, "WebhookService.RegisterWebhookService")
	defer span.End()

	s.ObservabilityStack.Logger.InfoContext(ctx, "Registering webhook",
		"environment", req.Environment,
		"service", req.ServiceName,
		"url", req.URL,
		"method", req.Method)

	key := fmt.Sprintf("/webhooks/%s/%s", req.Environment, req.ServiceName)
	var hooks []dtos.RegisterWebhookRequest

	webHook, err := s.Repo.GetEtcdKey(ctx, key)
	if err == nil && len(webHook) > 0 {
		if err := json.Unmarshal([]byte(webHook), &hooks); err != nil {
			s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to parse webhook data",
				"error", err,
				"key", key)
			return nil, &dtos.ServiceErrorResponse{
				StatusCode:   500,
				ErrorCode:    "Failed to parse webhook data",
				ErrorMessage: fmt.Sprintf("invalid webhook data stored for %s: %v", key, err),
			}
		}
	}

	for _, existing := range hooks {
		if existing.URL == req.URL && existing.Method == req.Method {
			s.ObservabilityStack.Logger.WarnContext(ctx, "Webhook already exists",
				"url", req.URL,
				"method", req.Method)
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
		s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to marshal webhook data",
			"error", err,
			"key", key)
		return nil, &dtos.ServiceErrorResponse{
			StatusCode:   500,
			ErrorCode:    "Failed to marshal webhook data",
			ErrorMessage: fmt.Sprintf("failed to marshal webhook data for %s: %v", key, err),
		}
	}
	if err := s.Repo.SetEtcdKey(ctx, key, string(data), -1); err != nil {
		s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to store webhook data",
			"error", err,
			"key", key)
		return nil, &dtos.ServiceErrorResponse{
			StatusCode:   500,
			ErrorCode:    "Failed to store webhook data",
			ErrorMessage: fmt.Sprintf("failed to store webhook data for %s: %v", key, err),
		}
	}

	s.ObservabilityStack.Logger.InfoContext(ctx, "Webhook registered successfully",
		"environment", req.Environment,
		"service", req.ServiceName)
	response := &dtos.SuccessResponse{
		StatusCode: 201,
		Message:    "Webhook registered successfully",
		Data:       map[string]interface{}{"webhook": req},
	}
	return response, nil
}

func (s *WebhookService) GetWebhooks(ctx context.Context, env, service string) ([]dtos.RegisterWebhookRequest, *dtos.ServiceErrorResponse) {
	ctx, span := s.ObservabilityStack.TracerService.Start(ctx, "WebhookService.GetWebhooks")
	defer span.End()

	s.ObservabilityStack.Logger.InfoContext(ctx, "Getting webhooks",
		"environment", env,
		"service", service)

	key := fmt.Sprintf("/webhooks/%s/%s", env, service)

	webHook, err := s.Repo.GetEtcdKey(ctx, key)
	if err != nil || len(webHook) == 0 {
		s.ObservabilityStack.Logger.WarnContext(ctx, "No webhooks found",
			"environment", env,
			"service", service,
			"error", err)
		return nil, &dtos.ServiceErrorResponse{
			StatusCode:   404,
			ErrorCode:    "No webhooks found",
			ErrorMessage: fmt.Sprintf("no webhooks found for %s/%s: %v", env, service, err),
		}
	}

	var hooks []dtos.RegisterWebhookRequest
	if err := json.Unmarshal([]byte(webHook), &hooks); err != nil {
		s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to parse webhook data",
			"error", err,
			"environment", env,
			"service", service)
		return nil, &dtos.ServiceErrorResponse{
			StatusCode:   500,
			ErrorCode:    "Failed to parse webhook data",
			ErrorMessage: fmt.Sprintf("failed to parse webhook data for %s/%s: %v", env, service, err),
		}
	}

	s.ObservabilityStack.Logger.InfoContext(ctx, "Successfully retrieved webhooks",
		"environment", env,
		"service", service,
		"count", len(hooks))
	return hooks, nil
}

func (s *WebhookService) DeleteWebhook(ctx context.Context, env, service, url, method string) (string, *dtos.ServiceErrorResponse) {
	ctx, span := s.ObservabilityStack.TracerService.Start(ctx, "WebhookService.DeleteWebhook")
	defer span.End()

	s.ObservabilityStack.Logger.InfoContext(ctx, "Deleting webhook",
		"environment", env,
		"service", service,
		"url", url,
		"method", method)

	key := fmt.Sprintf("/webhooks/%s/%s", env, service)

	hooks, errService := s.GetWebhooks(ctx, env, service)
	if errService != nil {
		s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to get webhooks for deletion",
			"error", errService.ErrorMessage,
			"environment", env,
			"service", service)
		return "", errService
	}

	updated := make([]dtos.RegisterWebhookRequest, 0)
	found := false
	for _, h := range hooks {
		if h.URL == url && h.Method == method {
			found = true
			continue
		}
		updated = append(updated, h)
	}

	if !found {
		s.ObservabilityStack.Logger.WarnContext(ctx, "Webhook not found for deletion",
			"environment", env,
			"service", service,
			"url", url,
			"method", method)
		return "", &dtos.ServiceErrorResponse{
			StatusCode:   404,
			ErrorCode:    "Webhook not found",
			ErrorMessage: fmt.Sprintf("webhook with URL '%s' and method '%s' not found", url, method),
		}
	}

	data, err := json.Marshal(updated)
	if err != nil {
		s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to marshal updated webhooks",
			"error", err,
			"environment", env,
			"service", service)
		return "", &dtos.ServiceErrorResponse{
			StatusCode:   500,
			ErrorCode:    "Failed to marshal updated webhooks",
			ErrorMessage: fmt.Sprintf("failed to marshal updated webhooks for %s/%s: %v", env, service, err),
		}
	}

	if err := s.Repo.SetEtcdKey(ctx, key, string(data), -1); err != nil {
		s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to store updated webhooks",
			"error", err,
			"environment", env,
			"service", service)
		return "", &dtos.ServiceErrorResponse{
			StatusCode:   500,
			ErrorCode:    "Failed to store updated webhooks",
			ErrorMessage: fmt.Sprintf("failed to store updated webhooks for %s/%s: %v", env, service, err),
		}
	}

	s.ObservabilityStack.Logger.InfoContext(ctx, "Webhook deleted successfully",
		"environment", env,
		"service", service,
		"url", url,
		"method", method)
	return "Webhook deleted successfully", nil
}

func (s *WebhookService) DeleteAllWebhooks(ctx context.Context, env, service string) error {
	ctx, span := s.ObservabilityStack.TracerService.Start(ctx, "WebhookService.DeleteAllWebhooks")
	defer span.End()

	s.ObservabilityStack.Logger.InfoContext(ctx, "Deleting all webhooks",
		"environment", env,
		"service", service)

	key := fmt.Sprintf("/webhooks/%s/%s", env, service)
	err := s.Repo.DeleteEtcdKey(ctx, key)
	if err != nil {
		s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to delete all webhooks",
			"error", err,
			"environment", env,
			"service", service)
		return err
	}

	s.ObservabilityStack.Logger.InfoContext(ctx, "Successfully deleted all webhooks",
		"environment", env,
		"service", service)
	return nil
}

func (s *WebhookService) NotifyWebhook(ctx context.Context, hook dtos.RegisterWebhookRequest, data map[string]interface{}) {
	ctx, span := s.ObservabilityStack.TracerService.Start(ctx, "WebhookService.NotifyWebhook")
	defer span.End()

	s.ObservabilityStack.Logger.InfoContext(ctx, "Notifying webhook",
		"environment", hook.Environment,
		"service", hook.ServiceName,
		"url", hook.URL,
		"method", hook.Method)

	body, err := json.Marshal(map[string]interface{}{
		"values":      data,
		"method":      hook.Method,
		"environment": hook.Environment,
		"serviceName": hook.ServiceName,
	})
	if err != nil {
		s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to marshal webhook payload",
			"error", err)
		return
	}

	fullURL := fmt.Sprintf("%s/%s/%s", strings.TrimRight(hook.URL, "/"), hook.Environment, hook.ServiceName)
	if !strings.HasPrefix(fullURL, "http://") && !strings.HasPrefix(fullURL, "https://") {
		fullURL = "http://" + fullURL
	}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		client := http.Client{Timeout: requestTimeout}
		resp, err := client.Post(fullURL, "application/json", bytes.NewBuffer(body))

		if err != nil {
			s.ObservabilityStack.Logger.WarnContext(ctx, "Webhook notification attempt failed",
				"attempt", attempt,
				"url", fullURL,
				"error", err)
		} else {
			defer resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				s.ObservabilityStack.Logger.InfoContext(ctx, "Webhook notification succeeded",
					"url", fullURL,
					"status", resp.Status)
				return
			}
			s.ObservabilityStack.Logger.WarnContext(ctx, "Webhook notification failed with non-2xx status",
				"attempt", attempt,
				"url", fullURL,
				"status", resp.Status)
		}

		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	s.ObservabilityStack.Logger.ErrorContext(ctx, "All webhook notification attempts failed",
		"url", fullURL,
		"max_retries", maxRetries)
}
