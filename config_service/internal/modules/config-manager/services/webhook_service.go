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

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
	if ObservabilityStack == nil {
		panic("ObservabilityStack cannot be nil")
	}
	return &WebhookService{Repo: repo, ObservabilityStack: ObservabilityStack}
}

func (s *WebhookService) RegisterWebhookService(ctx context.Context, req dtos.RegisterWebhookRequest) (*dtos.SuccessResponse, *dtos.ServiceErrorResponse) {
	ctx, span := s.ObservabilityStack.TracerService.Start(ctx, "WebhookService.RegisterWebhook")
	defer span.End()

	span.SetAttributes(
		attribute.String("environment", req.Environment),
		attribute.String("service", req.ServiceName),
		attribute.String("webhook.url", req.URL),
		attribute.String("webhook.method", req.Method),
	)

	s.ObservabilityStack.Logger.InfoContext(ctx, "Registering webhook",
		"environment", req.Environment,
		"service", req.ServiceName,
		"url", req.URL,
		"method", req.Method)

	key := fmt.Sprintf("/webhooks/%s/%s", req.Environment, req.ServiceName)
	span.SetAttributes(attribute.String("etcd.key", key))

	var hooks []dtos.RegisterWebhookRequest

	webHook, err := s.Repo.GetEtcdKey(ctx, key)
	if err == nil && len(webHook) > 0 {
		if err := json.Unmarshal([]byte(webHook), &hooks); err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.SetAttributes(attribute.String("error", "failed to parse webhook data"))
			s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to parse webhook data",
				"error", err,
				"key", key)
			return nil, &dtos.ServiceErrorResponse{
				StatusCode:   500,
				ErrorCode:    "Failed to parse webhook data",
				ErrorMessage: fmt.Sprintf("invalid webhook data stored for %s: %v", key, err),
			}
		}
		span.SetAttributes(attribute.Int("existing_hooks_count", len(hooks)))
	}

	for _, existing := range hooks {
		if existing.URL == req.URL && existing.Method == req.Method {
			span.SetStatus(codes.Error, "Webhook already exists")
			span.SetAttributes(attribute.String("error", "webhook already exists"))
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
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error", "failed to marshal webhook data"))
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
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error", "failed to store webhook data"))
		s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to store webhook data",
			"error", err,
			"key", key)
		return nil, &dtos.ServiceErrorResponse{
			StatusCode:   500,
			ErrorCode:    "Failed to store webhook data",
			ErrorMessage: fmt.Sprintf("failed to store webhook data for %s: %v", key, err),
		}
	}

	span.SetStatus(codes.Ok, "Webhook registered successfully")
	s.ObservabilityStack.Logger.InfoContext(ctx, "Webhook registered successfully",
		"environment", req.Environment,
		"service", req.ServiceName)
	return &dtos.SuccessResponse{
		StatusCode: 201,
		Message:    "Webhook registered successfully",
		Data:       map[string]interface{}{"webhook": req},
	}, nil
}

func (s *WebhookService) GetWebhooks(ctx context.Context, env, service string) ([]dtos.RegisterWebhookRequest, *dtos.ServiceErrorResponse) {
	ctx, span := s.ObservabilityStack.TracerService.Start(ctx, "WebhookService.GetWebhooks")
	defer span.End()

	span.SetAttributes(
		attribute.String("environment", env),
		attribute.String("service", service),
	)

	s.ObservabilityStack.Logger.InfoContext(ctx, "Getting webhooks",
		"environment", env,
		"service", service)

	key := fmt.Sprintf("/webhooks/%s/%s", env, service)
	span.SetAttributes(attribute.String("etcd.key", key))

	webHook, err := s.Repo.GetEtcdKey(ctx, key)
	if err != nil || len(webHook) == 0 {
		span.SetStatus(codes.Error, "No webhooks found")
		span.SetAttributes(attribute.String("error", "no webhooks found"))
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
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error", "failed to parse webhook data"))
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

	span.SetAttributes(attribute.Int("webhooks_count", len(hooks)))
	span.SetStatus(codes.Ok, "Webhooks retrieved successfully")
	s.ObservabilityStack.Logger.InfoContext(ctx, "Successfully retrieved webhooks",
		"environment", env,
		"service", service,
		"count", len(hooks))
	return hooks, nil
}

func (s *WebhookService) DeleteWebhook(ctx context.Context, env, service, url, method string) (string, *dtos.ServiceErrorResponse) {
	ctx, span := s.ObservabilityStack.TracerService.Start(ctx, "WebhookService.DeleteWebhook")
	defer span.End()

	span.SetAttributes(
		attribute.String("environment", env),
		attribute.String("service", service),
		attribute.String("webhook.url", url),
		attribute.String("webhook.method", method),
	)

	s.ObservabilityStack.Logger.InfoContext(ctx, "Deleting webhook",
		"environment", env,
		"service", service,
		"url", url,
		"method", method)

	key := fmt.Sprintf("/webhooks/%s/%s", env, service)
	span.SetAttributes(attribute.String("etcd.key", key))

	hooks, errService := s.GetWebhooks(ctx, env, service)
	if errService != nil {
		span.SetStatus(codes.Error, errService.ErrorMessage)
		span.SetAttributes(attribute.String("error", "failed to get webhooks"))
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
		span.SetStatus(codes.Error, "Webhook not found")
		span.SetAttributes(attribute.String("error", "webhook not found"))
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
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error", "failed to marshal updated webhooks"))
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
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error", "failed to store updated webhooks"))
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

	span.SetAttributes(attribute.Int("remaining_hooks_count", len(updated)))
	span.SetStatus(codes.Ok, "Webhook deleted successfully")
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

	span.SetAttributes(
		attribute.String("environment", env),
		attribute.String("service", service),
	)

	s.ObservabilityStack.Logger.InfoContext(ctx, "Deleting all webhooks",
		"environment", env,
		"service", service)

	key := fmt.Sprintf("/webhooks/%s/%s", env, service)
	span.SetAttributes(attribute.String("etcd.key", key))

	err := s.Repo.DeleteEtcdKey(ctx, key)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error", "failed to delete all webhooks"))
		s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to delete all webhooks",
			"error", err,
			"environment", env,
			"service", service)
		return err
	}

	span.SetStatus(codes.Ok, "All webhooks deleted successfully")
	s.ObservabilityStack.Logger.InfoContext(ctx, "Successfully deleted all webhooks",
		"environment", env,
		"service", service)
	return nil
}

func (s *WebhookService) NotifyWebhook(ctx context.Context, hook dtos.RegisterWebhookRequest, data map[string]interface{}) {
	ctx, span := s.ObservabilityStack.TracerService.Start(ctx, "WebhookService.NotifyWebhook")
	defer span.End()

	span.SetAttributes(
		attribute.String("environment", hook.Environment),
		attribute.String("service", hook.ServiceName),
		attribute.String("webhook.url", hook.URL),
		attribute.String("webhook.method", hook.Method),
		attribute.Int("max_retries", maxRetries),
		attribute.Float64("retry_delay_seconds", retryDelay.Seconds()),
		attribute.Float64("request_timeout_seconds", requestTimeout.Seconds()),
	)

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
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error", "failed to marshal webhook payload"))
		s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to marshal webhook payload",
			"error", err)
		return
	}

	fullURL := fmt.Sprintf("%s/%s/%s", strings.TrimRight(hook.URL, "/"), hook.Environment, hook.ServiceName)
	if !strings.HasPrefix(fullURL, "http://") && !strings.HasPrefix(fullURL, "https://") {
		fullURL = "http://" + fullURL
	}
	span.SetAttributes(attribute.String("webhook.full_url", fullURL))

	for attempt := 1; attempt <= maxRetries; attempt++ {
		retryCtx, retrySpan := s.ObservabilityStack.TracerService.Start(ctx, "WebhookService.NotifyWebhook.Attempt")
		retrySpan.SetAttributes(
			attribute.Int("attempt", attempt),
			attribute.String("url", fullURL),
		)

		client := http.Client{Timeout: requestTimeout}
		resp, err := client.Post(fullURL, "application/json", bytes.NewBuffer(body))

		if err != nil {
			retrySpan.SetStatus(codes.Error, err.Error())
			retrySpan.SetAttributes(attribute.String("error", "webhook notification attempt failed"))
			s.ObservabilityStack.Logger.WarnContext(retryCtx, "Webhook notification attempt failed",
				"attempt", attempt,
				"url", fullURL,
				"error", err)
		} else {
			defer resp.Body.Close()
			retrySpan.SetAttributes(attribute.Int("status_code", resp.StatusCode))
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				retrySpan.SetStatus(codes.Ok, "Webhook notification succeeded")
				s.ObservabilityStack.Logger.InfoContext(retryCtx, "Webhook notification succeeded",
					"url", fullURL,
					"status", resp.Status)
				retrySpan.End()
				span.SetStatus(codes.Ok, "Webhook notification succeeded")
				return
			}
			retrySpan.SetStatus(codes.Error, fmt.Sprintf("Webhook notification failed with status: %s", resp.Status))
			s.ObservabilityStack.Logger.WarnContext(retryCtx, "Webhook notification failed with non-2xx status",
				"attempt", attempt,
				"url", fullURL,
				"status", resp.Status)
		}

		retrySpan.End()
		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	span.SetStatus(codes.Error, "All webhook notification attempts failed")
	span.SetAttributes(attribute.String("error", "all attempts failed"))
	s.ObservabilityStack.Logger.ErrorContext(ctx, "All webhook notification attempts failed",
		"url", fullURL,
		"max_retries", maxRetries)
}
