package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"nps-config-service/internal/constants"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/repositories"
	"nps-config-service/pkg/observability"
	workflows "nps-config-service/pkg/temporal"

	temporalClient "go.temporal.io/sdk/client"
)

type ConfigService struct {
	Repo               repositories.IConfigRepo
	WebhookService     IWebhookService
	temporalClient     temporalClient.Client
	ObservabilityStack *observability.ObservabilityStack
}

type IConfigService interface {
	StoreConfigService(ctx context.Context,env string, service string, req map[string]interface{}) (*dtos.SuccessResponse, *dtos.ServiceErrorResponse)
	GetConfigService(ctx context.Context,service string, env string) (interface{}, error)
	GetConfigValueService(ctx context.Context,serviceName string, env string, key string) (interface{}, error)
}

func NewConfigService(repo repositories.IConfigRepo, webHook IWebhookService, temporal temporalClient.Client, ObservabilityStack *observability.ObservabilityStack) IConfigService {
	return &ConfigService{Repo: repo, WebhookService: webHook, temporalClient: temporal, ObservabilityStack: ObservabilityStack}
}

func (s *ConfigService) StoreConfigService(ctx context.Context,env string, service string, req map[string]interface{}) (*dtos.SuccessResponse, *dtos.ServiceErrorResponse) {

	response, err := s.Repo.StoreConfig(ctx, service, env, req)

	if err != nil {
		fmt.Printf("failed to store %s: %v", env, err)
		return nil, &dtos.ServiceErrorResponse{
			StatusCode:   500,
			ErrorCode:    "Failed to store config",
			ErrorMessage: err.Error(),
		}
	}
	key := fmt.Sprintf("/webhooks/%s/%s", env, service)
	webHook, err := s.Repo.GetEtcdKey(ctx, key)
	if err != nil {
		fmt.Printf("No webhook found for %s: %v", key, err)
		return &dtos.SuccessResponse{
			StatusCode: 201,
			Message:    "Config stored successfully: NOTE - No webhooks found",
			Data:       response,
		}, nil
	}

	var hooks []dtos.RegisterWebhookRequest
	if err := json.Unmarshal([]byte(webHook), &hooks); err != nil {
		fmt.Printf("Failed to parse webhook data for %s: %v", key, err)
		return nil, &dtos.ServiceErrorResponse{
			StatusCode:   500,
			ErrorCode:    "Failed to parse webhook data",
			ErrorMessage: err.Error(),
		}
	}

	for _, hook := range hooks {
		input := workflows.WebhookInput{
			Hook: hook,
			Data: req,
		}
		log.Println("Starting webhook workflow for URL:", hook.URL)
		_, err := s.temporalClient.ExecuteWorkflow(context.Background(),
			temporalClient.StartWorkflowOptions{
				ID:        fmt.Sprintf("webhook-%s-%s url: %s", hook.ServiceName, hook.Environment, hook.URL),
				TaskQueue: constants.SendWebhookTaskQueueName,
			},
			workflows.WebhookWorkflow,
			input,
		)

		if err != nil {
			log.Printf("Failed to start webhook workflow for %s: %v", hook.URL, err)
		}
	}

	return &dtos.SuccessResponse{
		StatusCode: 201,
		Message:    "Config stored successfully",
		Data:       response,
	}, nil

}

func (s *ConfigService) GetConfigService(ctx context.Context,service string, env string) (any, error) {

	response, err := s.Repo.GetConfig(ctx, service, env)

	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, nil // Explicitly return nil if response is nil
	}

	return response, nil
}

func (s *ConfigService) GetConfigValueService(ctx context.Context,serviceName string, env string, key string) (any, error) {
	response, err := s.Repo.GetConfigValue(ctx, serviceName, env, key)

	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, nil // Explicitly return nil if response is nil
	}

	return response, nil
}

