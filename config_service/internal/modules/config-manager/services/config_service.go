package services

import (
	"context"
	"encoding/json"
	"fmt"
	"nps-config-service/internal/constants"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/repositories"
	"nps-config-service/pkg/observability"
	workflows "nps-config-service/pkg/temporal"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	temporalClient "go.temporal.io/sdk/client"
)

type ConfigService struct {
	Repo               repositories.IConfigRepo
	WebhookService     IWebhookService
	temporalClient     temporalClient.Client
	ObservabilityStack *observability.ObservabilityStack
}

type IConfigService interface {
	StoreConfigService(ctx context.Context, env string, service string, req map[string]interface{}) (*dtos.SuccessResponse, *dtos.ServiceErrorResponse)
	GetConfigService(ctx context.Context, service string, env string) (interface{}, error)
	GetConfigValueService(ctx context.Context, serviceName string, env string, key string) (interface{}, error)
}

func NewConfigService(repo repositories.IConfigRepo, webHook IWebhookService, temporal temporalClient.Client, ObservabilityStack *observability.ObservabilityStack) IConfigService {
	if ObservabilityStack == nil {
		panic("ObservabilityStack cannot be nil")
	}
	return &ConfigService{Repo: repo, WebhookService: webHook, temporalClient: temporal, ObservabilityStack: ObservabilityStack}
}

func (s *ConfigService) StoreConfigService(ctx context.Context, env string, service string, req map[string]interface{}) (*dtos.SuccessResponse, *dtos.ServiceErrorResponse) {
	ctx, span := s.ObservabilityStack.TracerService.Start(ctx, "ConfigService.StoreConfig")
	defer span.End()

	span.SetAttributes(
		attribute.String("environment", env),
		attribute.String("service", service),
	)

	s.ObservabilityStack.Logger.InfoContext(ctx, "Storing config",
		"environment", env,
		"service", service)

	response, err := s.Repo.StoreConfig(ctx, service, env, req)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error", "failed to store config"))
		s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to store config",
			"error", err,
			"environment", env,
			"service", service)
		return nil, &dtos.ServiceErrorResponse{
			StatusCode:   500,
			ErrorCode:    "Failed to store config",
			ErrorMessage: err.Error(),
		}
	}

	// Check for webhooks
	key := fmt.Sprintf("/webhooks/%s/%s", env, service)
	webHook, err := s.Repo.GetEtcdKey(ctx, key)
	if err != nil {
		s.ObservabilityStack.Logger.WarnContext(ctx, "No webhooks found",
			"key", key,
			"error", err)
		span.SetAttributes(attribute.String("webhook.status", "not_found"))
		return &dtos.SuccessResponse{
			StatusCode: 201,
			Message:    "Config stored successfully: NOTE - No webhooks found",
			Data:       response,
		}, nil
	}

	var hooks []dtos.RegisterWebhookRequest
	if err := json.Unmarshal([]byte(webHook), &hooks); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error", "failed to parse webhook data"))
		s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to parse webhook data",
			"error", err,
			"key", key)
		return nil, &dtos.ServiceErrorResponse{
			StatusCode:   500,
			ErrorCode:    "Failed to parse webhook data",
			ErrorMessage: err.Error(),
		}
	}

	span.SetAttributes(attribute.Int("webhook.count", len(hooks)))
	s.ObservabilityStack.Logger.InfoContext(ctx, "Processing webhooks",
		"count", len(hooks),
		"environment", env,
		"service", service)

	// Process webhooks
	for _, hook := range hooks {
		ctx, span := s.ObservabilityStack.TracerService.Start(ctx, "ConfigService.ProcessWebhook")
		span.SetAttributes(
			attribute.String("webhook.url", hook.URL),
			attribute.String("webhook.method", hook.Method),
		)

		input := workflows.WebhookInput{
			Hook: hook,
			Data: req,
		}

		s.ObservabilityStack.Logger.InfoContext(ctx, "Starting webhook workflow",
			"url", hook.URL,
			"environment", env,
			"service", service)

		_, err := s.temporalClient.ExecuteWorkflow(ctx,
			temporalClient.StartWorkflowOptions{
				ID:        fmt.Sprintf("webhook-%s-%s url: %s", hook.ServiceName, hook.Environment, hook.URL),
				TaskQueue: constants.SendWebhookTaskQueueName,
			},
			workflows.WebhookWorkflow,
			input,
		)

		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.SetAttributes(attribute.String("error", "failed to start webhook workflow"))
			s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to start webhook workflow",
				"error", err,
				"url", hook.URL)
		} else {
			span.SetStatus(codes.Ok, "Webhook workflow started successfully")
			s.ObservabilityStack.Logger.InfoContext(ctx, "Webhook workflow started successfully",
				"url", hook.URL)
		}
		span.End()
	}

	span.SetStatus(codes.Ok, "Config stored successfully")
	s.ObservabilityStack.Logger.InfoContext(ctx, "Config stored successfully",
		"environment", env,
		"service", service)

	return &dtos.SuccessResponse{
		StatusCode: 201,
		Message:    "Config stored successfully",
		Data:       response,
	}, nil
}

func (s *ConfigService) GetConfigService(ctx context.Context, service string, env string) (any, error) {
	ctx, span := s.ObservabilityStack.TracerService.Start(ctx, "ConfigService.GetConfig")
	defer span.End()

	span.SetAttributes(
		attribute.String("environment", env),
		attribute.String("service", service),
	)

	s.ObservabilityStack.Logger.InfoContext(ctx, "Getting config",
		"environment", env,
		"service", service)

	response, err := s.Repo.GetConfig(ctx, service, env)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error", "failed to get config"))
		s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to get config",
			"error", err,
			"environment", env,
			"service", service)
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	if response == nil {
		span.SetStatus(codes.Ok, "Config not found")
		s.ObservabilityStack.Logger.InfoContext(ctx, "Config not found",
			"environment", env,
			"service", service)
		return nil, nil
	}

	span.SetStatus(codes.Ok, "Config retrieved successfully")
	s.ObservabilityStack.Logger.InfoContext(ctx, "Config retrieved successfully",
		"environment", env,
		"service", service)

	return response, nil
}

func (s *ConfigService) GetConfigValueService(ctx context.Context, serviceName string, env string, key string) (any, error) {
	ctx, span := s.ObservabilityStack.TracerService.Start(ctx, "ConfigService.GetConfigValue")
	defer span.End()

	span.SetAttributes(
		attribute.String("environment", env),
		attribute.String("service", serviceName),
		attribute.String("key", key),
	)

	s.ObservabilityStack.Logger.InfoContext(ctx, "Getting config value",
		"environment", env,
		"service", serviceName,
		"key", key)

	response, err := s.Repo.GetConfigValue(ctx, serviceName, env, key)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error", "failed to get config value"))
		s.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to get config value",
			"error", err,
			"environment", env,
			"service", serviceName,
			"key", key)
		return nil, fmt.Errorf("failed to get config value: %w", err)
	}

	if response == nil {
		span.SetStatus(codes.Ok, "Config value not found")
		s.ObservabilityStack.Logger.InfoContext(ctx, "Config value not found",
			"environment", env,
			"service", serviceName,
			"key", key)
		return nil, nil
	}

	span.SetStatus(codes.Ok, "Config value retrieved successfully")
	s.ObservabilityStack.Logger.InfoContext(ctx, "Config value retrieved successfully",
		"environment", env,
		"service", serviceName,
		"key", key)

	return response, nil
}
