package client

import (
	"context"
	"encoding/json"
	"fmt"
	"messaging_service/internal/config"
	"messaging_service/internal/modules/message_broker/models"
	"messaging_service/pkg/errors"
	httpclient "messaging_service/pkg/http"
	"messaging_service/pkg/observability"
	"net/http"
	"time"
)

const (
	// Metric names
	metricConfigFetchTotal   = "config_fetch_total"
	metricConfigFetchSuccess = "config_fetch_success"
	metricConfigFetchFailure = "config_fetch_failure"

	// Error types
	errorTypeHttp       = "http_error"
	errorTypeDecode     = "decode_error"
	errorTypeValidation = "validation_error"
)

type ConfigClient interface {
	FetchConfig(ctx context.Context) (*models.MessaggingConfigResponse, error)
}

type ConfigClientImpl struct {
	httpClient httpclient.HTTPClient
	cfg        *config.Env
	obs        *observability.ObservabilityStack
}

func NewConfigClient(httpClient httpclient.HTTPClient, cfg *config.Env, obs *observability.ObservabilityStack) (ConfigClient, error) {
	// Validate input arguments
	if httpClient == nil {
		return nil, fmt.Errorf("httpClient cannot be nil in NewConfigClient")
	}
	if cfg == nil {
		return nil, fmt.Errorf("cfg cannot be nil in NewConfigClient")
	}
	if obs == nil {
		return nil, fmt.Errorf("observability stack cannot be nil in NewConfigClient")
	}

	// Validate required configuration fields
	if cfg.ConfigServiceUrl == "" {
		return nil, fmt.Errorf("config service URL is required in environment configuration")
	}
	if cfg.ConfigServiceToken == "" {
		return nil, fmt.Errorf("config service token is required in environment configuration")
	}
	if cfg.Environment == "" {
		return nil, fmt.Errorf("environment is required in environment configuration")
	}
	if config.SERVICE_NAME == "" {
		return nil, fmt.Errorf("service name is required in environment configuration")
	}

	return &ConfigClientImpl{
		httpClient: httpClient,
		cfg:        cfg,
		obs:        obs,
	}, nil
}

// GetCurrentConfig returns the config for the given key.
func (c *ConfigClientImpl) FetchConfig(ctx context.Context) (*models.MessaggingConfigResponse, error) {
	// Start tracing
	functionName := "FetchConfig"
	tCtx, span := c.obs.TracerService.StartTracer(ctx, functionName)
	defer c.obs.TracerService.StopSpan(span)

	// Generate request ID for correlation
	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())

	// Set tracing attributes
	c.obs.TracerService.SetAttributes(span, map[string]string{
		"request_id":  requestID,
		"operation":   functionName,
		"service":     config.SERVICE_NAME,
		"environment": c.cfg.Environment,
	})

	// Increment total metric
	c.obs.MetricsService.IncrementCounter(tCtx, metricConfigFetchTotal, 1, nil)

	fullURL := fmt.Sprintf(
		"%s/config/%s/%s",
		c.cfg.ConfigServiceUrl,
		c.cfg.Environment,
		config.SERVICE_NAME,
	)

	c.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Fetching config from service for environment: %s",
		requestID, c.cfg.Environment))

	req := httpclient.Request{
		Method: httpclient.GET,
		URL:    fullURL,
		Headers: map[string]string{
			"Authorization": "Bearer " + c.cfg.ConfigServiceToken, // We need the actual token for the request
		},
	}

	var configResponse models.ConfigServiceResponse
	resp, err := c.httpClient.Do(tCtx, req) // Pass traced context to maintain trace propagation
	if err != nil {
		c.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Config fetch request failed", requestID))
		c.obs.MetricsService.IncrementCounter(tCtx, metricConfigFetchFailure, 1, map[string]string{
			"error_type": errorTypeHttp,
		})
		return nil, errors.NewCustomError(errors.CFGErrFetchFailed, err)
	}
	defer resp.Body.Close()

	c.obs.LoggerService.Debug(tCtx, fmt.Sprintf("[%s] Received response with status code: %d",
		requestID, resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		c.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Config fetch failed with status code: %d",
			requestID, resp.StatusCode))
		c.obs.MetricsService.IncrementCounter(tCtx, metricConfigFetchFailure, 1, map[string]string{
			"error_type":  errorTypeHttp,
			"status_code": fmt.Sprintf("%d", resp.StatusCode),
		})
		return nil, errors.NewCustomError(errors.CFGErrFetchFailed,
			fmt.Errorf("config fetch failed: status=%d", resp.StatusCode))
	}

	if err := json.NewDecoder(resp.Body).Decode(&configResponse); err != nil {
		c.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Failed to decode config response", requestID))
		c.obs.MetricsService.IncrementCounter(tCtx, metricConfigFetchFailure, 1, map[string]string{
			"error_type": errorTypeDecode,
		})
		return nil, errors.NewCustomError(errors.CFGErrUnmarshalFailed, err)
	}

	// Basic validation of the response
	if configResponse.Data.KafkaBrokers == nil || len(configResponse.Data.KafkaBrokers) == 0 {
		c.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Invalid config received: KafkaBrokers is missing or empty", requestID))
		c.obs.MetricsService.IncrementCounter(tCtx, metricConfigFetchFailure, 1, map[string]string{
			"error_type": errorTypeValidation,
		})
		return nil, errors.NewCustomError(errors.CFGErrValidationFailed,
			fmt.Errorf("invalid config received: KafkaBrokers is required but missing or empty"))
	}

	c.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Successfully fetched and validated config", requestID))
	c.obs.MetricsService.IncrementCounter(tCtx, metricConfigFetchSuccess, 1, nil)

	return &configResponse.Data, nil
}
