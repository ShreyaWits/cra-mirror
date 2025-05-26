package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	configEnv "template-services/internal/configs"
	"template-services/internal/constants"
	"template-services/internal/template/dto"
	cacheclient "template-services/pkg/client/cache_client"
	"template-services/pkg/errors"
	httpclient "template-services/pkg/http"
	"template-services/pkg/observability"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ConfigServiceImpl struct {
	cache *cacheclient.RedisClientStruct
	obs   *observability.ObservabilityStack
}

func NewConfigService(cache *cacheclient.RedisClientStruct, obs *observability.ObservabilityStack) *ConfigServiceImpl {
	return &ConfigServiceImpl{
		cache: cache,
		obs:   obs,
	}
}

func (h *ConfigServiceImpl) HandleConfigWebhookChange(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	functionName := "HandleConfigWebhookChange"
	functionFailed := "HandleConfigWebhookChange_Failed"

	tCtx, span := h.obs.TracerService.StartTracer(ctx, functionName)
	defer h.obs.TracerService.StopSpan(span)
	h.obs.MetricsService.IncrementCounter(tCtx, functionName, 1, map[string]string{})

	webhookConfigData := configEnv.ConfigServiceWebhookData{}

	if err := c.BodyParser(&webhookConfigData); err != nil {
		h.obs.LoggerService.Error(tCtx, "Failed to parse webhook config data", map[string]interface{}{
			"error": err.Error(),
		})
		h.obs.MetricsService.IncrementCounter(tCtx, functionFailed, 1, map[string]string{
			"error": errors.GetAppErrorMessage(errors.TmpErrInvalidRequestBody),
		})
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success:      false,
			ErrorMessage: errors.GetAppErrorMessage(errors.TmpErrInvalidRequestBody),
			ErrorCode:    errors.TmpErrInvalidRequestBody,
			Data:         nil,
		})
	}

	h.UpdateConfig(tCtx, &webhookConfigData.Values)
	h.obs.LoggerService.Info(tCtx, "Webhook config data received", map[string]interface{}{
		"config_data": webhookConfigData,
	})
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"Success": true,
		"Message": "Webhook config data processed successfully",
		"Data":    nil,
	})
}

func (s *ConfigServiceImpl) UpdateConfig(ctx context.Context, config *configEnv.Config) error {
	ttl := 240 * time.Hour
	_, span := s.obs.TracerService.StartTracer(ctx, "UpdateConfig")
	defer s.obs.TracerService.StopSpan(span)

	fmt.Println("called refresh config")
	go configEnv.RefreshConfig(config)

	// Marshal config to JSON string
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return s.cache.SetCache(ctx, constants.ServiceName, "config", string(data), ttl)
}

func (s *ConfigServiceImpl) GetCurrentConfig(ctx context.Context, key string) (*configEnv.Config, error) {
	// Try to get from cache
	val, found, err := s.cache.GetCache(ctx, constants.ServiceName, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}
	if found {
		fmt.Println("Cache miss, fetching from config service")

		var cfg configEnv.Config
		if err := json.Unmarshal([]byte(val), &cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cached config: %w", err)
		}

		return &cfg, nil
	}

	fmt.Println("Fetching config from config service...")

	fullURL := configEnv.ImmutableConfigs.ConfigServiceUrl + "/config/" + configEnv.ImmutableConfigs.Environment + "/" + constants.ServiceName
	fmt.Println("Full URL:", fullURL)

	httpClient := httpclient.New(time.Second * 5)
	resp, err := httpClient.Do(ctx, "GET", fullURL, map[string]string{
		"Authorization": "Bearer " + configEnv.ImmutableConfigs.ConfigServiceToken,
	}, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch config, status code: %d", resp.StatusCode)
	}

	// Read the response body into a byte slice
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var responseMap configEnv.ConfigServiceResponse
	if err := json.Unmarshal(body, &responseMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body into map: %w", err)
	}

	defer resp.Body.Close()
	return &responseMap.Data, nil
}

func (s *ConfigServiceImpl) RegisterWebhook(ctx context.Context) error {
	req := configEnv.RegisterWebHookRequest{
		URL:         fmt.Sprintf(constants.SelfServiceUrl, configEnv.Configs.HttpPort) + "/config",
		Environment: configEnv.ImmutableConfigs.Environment,
		ServiceName: constants.ServiceName,
		Method:      "POST",
	}

	// Convert the webhook to JSON
	webhookJSON, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook: %w", err)
	}

	httpClient := httpclient.New(time.Second * 5)
	// Send the webhook to the configured URL
	resp, err := httpClient.Do(ctx, "POST", configEnv.ImmutableConfigs.ConfigServiceUrl+"/config/webhook", map[string]string{
		"Authorization": "Bearer " + configEnv.ImmutableConfigs.ConfigServiceToken,
		"Content-Type":  "application/json",
	}, webhookJSON)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to register webhook, status code: %d", resp.StatusCode)
	}

	return nil
}
