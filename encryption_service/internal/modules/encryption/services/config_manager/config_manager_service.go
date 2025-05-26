package service

import (
	"context"
	"encoding/json"
	"encryption_microservice/internal/config"
	cacheclient "encryption_microservice/internal/modules/encryption/client/cache_client"
	client "encryption_microservice/internal/modules/encryption/client/config_client"
	"encryption_microservice/pkg/observability"
	"fmt"
	"time"

	"github.com/google/uuid"
	otelcodes "go.opentelemetry.io/otel/codes"
)

type ConfigManagerService struct {
	configClient client.ConfigClient
	cacheclient  cacheclient.RedisClient
	env          *config.EnvConfig
	obs          *observability.ObservabilityStack
}

func NewConfigManager(configClient client.ConfigClient, cacheclient cacheclient.RedisClient, obs *observability.ObservabilityStack, env *config.EnvConfig) (*ConfigManagerService, error) {
	// Validate input arguments
	if configClient == nil {
		return nil, fmt.Errorf("configClient cannot be nil in NewConfigManager")
	}
	if cacheclient == nil {
		return nil, fmt.Errorf("cacheClient cannot be nil in NewConfigManager")
	}
	if obs == nil {
		return nil, fmt.Errorf("observability stack cannot be nil in NewConfigManager")
	}

	return &ConfigManagerService{
		configClient: configClient,
		cacheclient:  cacheclient,
		obs:          obs,
		env:          env,
	}, nil
}

// SetEnvironment sets the environment configuration
func (s *ConfigManagerService) SetEnvironment(env *config.EnvConfig) {
	s.env = env
}

func (s *ConfigManagerService) GetFromApiConfiguration(ctx context.Context) (*config.Config, error) {
	functionName := "GetFromApiConfiguration"

	// Start tracing
	ctx, span := s.obs.TracerService.StartTracer(ctx, functionName)
	defer s.obs.TracerService.StopSpan(span)

	// Set trace attributes
	s.obs.TracerService.SetAttributes(span, map[string]string{
		"service": config.SERVICE_NAME,
	})

	// Increment metrics counter
	s.obs.MetricsService.IncrementCounter(ctx, functionName, 1, nil)
	startTime := time.Now()

	var data *config.Config
	var err error
	var cacheErr error

	// Try to fetch config from API
	s.obs.LoggerService.Debug(ctx, "Attempting to fetch configuration from API")

	// Create a child span for API call
	apiCtx, apiSpan := s.obs.TracerService.StartTracer(ctx, "FetchConfig")
	data, err = s.configClient.FetchConfig(apiCtx)
	if err != nil {
		s.obs.TracerService.SetStatus(apiSpan, otelcodes.Error, "API fetch failed")
		s.obs.TracerService.SetAttributes(apiSpan, map[string]string{
			"error": err.Error(),
		})
		s.obs.LoggerService.Error(ctx, "API fetch failed, falling back to cache", "error", err)

		// If API fetch fails, try to get from cache
		cacheCtx, cacheSpan := s.obs.TracerService.StartTracer(ctx, "GetDataFromCache")
		data, cacheErr = s.GetDataToCache(cacheCtx, "config")
		if cacheErr != nil {
			s.obs.TracerService.SetStatus(cacheSpan, otelcodes.Error, "Cache retrieval failed")
			s.obs.TracerService.SetAttributes(cacheSpan, map[string]string{
				"error": cacheErr.Error(),
			})
			s.obs.LoggerService.Error(ctx, "Failed to retrieve from cache after API failure", "error", cacheErr)
			s.obs.TracerService.StopSpan(cacheSpan)
			s.obs.TracerService.StopSpan(apiSpan)

			// Set the main span status to error
			s.obs.TracerService.SetStatus(span, otelcodes.Error, "Both API and cache retrieval failed")
			s.obs.MetricsService.IncrementCounter(ctx, functionName+"_failed", 1, nil)

			return nil, cacheErr
		}
		s.obs.TracerService.SetStatus(cacheSpan, otelcodes.Ok, "Successfully retrieved from cache")
		s.obs.TracerService.StopSpan(cacheSpan)
		s.obs.LoggerService.Debug(ctx, "Successfully retrieved config from cache after API failure")
	} else {
		s.obs.TracerService.SetStatus(apiSpan, otelcodes.Ok, "Successfully fetched from API")
		s.obs.LoggerService.Debug(ctx, "Successfully fetched config from API")
	}
	s.obs.TracerService.StopSpan(apiSpan)

	// Validate and set the global config
	_, validateSpan := s.obs.TracerService.StartTracer(ctx, "ValidateAndSetConfig")
	if err := config.SetConfig(data); err != nil {
		s.obs.TracerService.SetStatus(validateSpan, otelcodes.Error, "Error setting config")
		s.obs.TracerService.SetAttributes(validateSpan, map[string]string{
			"error": err.Error(),
		})
		s.obs.LoggerService.Error(ctx, "Error setting config", "error", err)
	} else {
		s.obs.TracerService.SetStatus(validateSpan, otelcodes.Ok, "Successfully set global config")
		s.obs.LoggerService.Debug(ctx, "Successfully set global config")
	}
	s.obs.TracerService.StopSpan(validateSpan)

	// Store in cache
	s.obs.LoggerService.Debug(ctx, "Storing config in cache", "key", config.SERVICE_NAME)
	cacheStoreCtx, cacheStoreSpan := s.obs.TracerService.StartTracer(ctx, "StoreDataToCache")
	err = s.SetDataToCache(cacheStoreCtx, "config", data)
	if err != nil {
		s.obs.TracerService.SetStatus(cacheStoreSpan, otelcodes.Error, "Error saving data in cache")
		s.obs.TracerService.SetAttributes(cacheStoreSpan, map[string]string{
			"error": err.Error(),
		})
		s.obs.LoggerService.Error(ctx, "Error saving data in cache service ", "key", config.SERVICE_NAME, "error", err)
	} else {
		s.obs.TracerService.SetStatus(cacheStoreSpan, otelcodes.Ok, "Successfully stored in cache")
		s.obs.LoggerService.Debug(ctx, "Successfully stored config in cache", "key", config.SERVICE_NAME)
	}
	s.obs.TracerService.StopSpan(cacheStoreSpan)

	// Record metrics for the overall operation
	processingTime := time.Since(startTime).Milliseconds()
	s.obs.MetricsService.RecordHistogram(ctx, "config_manager_processing_time_ms", float64(processingTime), nil)

	// Set overall span status
	s.obs.TracerService.SetStatus(span, otelcodes.Ok, "Configuration retrieval completed")

	return data, nil
}

func (s *ConfigManagerService) SetDataToCache(ctx context.Context, key string, cfg *config.Config) error {
	functionName := "SetDataToCache"

	// Start tracing
	ctx, span := s.obs.TracerService.StartTracer(ctx, functionName)
	defer s.obs.TracerService.StopSpan(span)

	// Set trace attributes
	s.obs.TracerService.SetAttributes(span, map[string]string{
		"namespace": config.SERVICE_NAME,
		"key":       key,
	})

	// Add debug logs to show cache operation details
	s.obs.LoggerService.Debug(ctx, "Setting data to cache ", "namespace ", config.SERVICE_NAME, "key", key)
	startTime := time.Now()

	// Marshal config to JSON in a separate span
	_, marshalSpan := s.obs.TracerService.StartTracer(ctx, "MarshalConfig")
	data, err := json.Marshal(cfg)
	if err != nil {
		s.obs.TracerService.SetStatus(marshalSpan, otelcodes.Error, "Failed to marshal config")
		s.obs.TracerService.SetAttributes(marshalSpan, map[string]string{
			"error": err.Error(),
		})
		s.obs.LoggerService.Error(ctx, "Failed to marshal config ", "error", err)
		s.obs.TracerService.StopSpan(marshalSpan)

		// Set parent span status
		s.obs.TracerService.SetStatus(span, otelcodes.Error, "Failed to marshal config")

		return fmt.Errorf("failed to marshal config: %w", err)
	}
	s.obs.TracerService.SetStatus(marshalSpan, otelcodes.Ok, "Successfully marshaled config")
	s.obs.TracerService.StopSpan(marshalSpan)

	trackingID := uuid.NewString()
	ttl := time.Duration(s.env.CacheTTL) * time.Hour
	s.obs.LoggerService.Debug(ctx, "Cache parameters ", "ttl_hours ", s.env.CacheTTL, "tracking_id", trackingID)

	// Set cache in a separate span
	cacheCtx, cacheSpan := s.obs.TracerService.StartTracer(ctx, "SetCache")
	s.obs.TracerService.SetAttributes(cacheSpan, map[string]string{
		"tracking_id": trackingID,
		"ttl_hours":   fmt.Sprintf("%d", s.env.CacheTTL),
	})

	err = s.cacheclient.SetCache(cacheCtx, config.SERVICE_NAME, key, string(data), ttl, trackingID)
	if err != nil {
		s.obs.TracerService.SetStatus(cacheSpan, otelcodes.Error, "Failed to set cache")
		s.obs.TracerService.SetAttributes(cacheSpan, map[string]string{
			"error": err.Error(),
		})
		s.obs.TracerService.StopSpan(cacheSpan)

		// Set parent span status
		s.obs.TracerService.SetStatus(span, otelcodes.Error, "Failed to set cache")

		return err
	}
	s.obs.TracerService.SetStatus(cacheSpan, otelcodes.Ok, "Successfully set cache")
	s.obs.TracerService.StopSpan(cacheSpan)

	// Record metrics for the operation
	processingTime := time.Since(startTime).Milliseconds()
	s.obs.MetricsService.RecordHistogram(ctx, "cache_set_processing_time_ms", float64(processingTime), nil)

	// Set overall span status
	s.obs.TracerService.SetStatus(span, otelcodes.Ok, "Successfully set data to cache")

	return nil
}

func (s *ConfigManagerService) GetDataToCache(ctx context.Context, key string) (*config.Config, error) {
	functionName := "GetDataToCache"

	// Start tracing
	ctx, span := s.obs.TracerService.StartTracer(ctx, functionName)
	defer s.obs.TracerService.StopSpan(span)

	// Set trace attributes
	s.obs.TracerService.SetAttributes(span, map[string]string{
		"namespace": config.SERVICE_NAME,
		"key":       key,
	})

	// Add debug logs to show cache lookup details
	s.obs.LoggerService.Debug(ctx, "Getting data from cache", "namespace", config.SERVICE_NAME, "key", key)
	startTime := time.Now()

	trackingID := uuid.NewString()
	s.obs.LoggerService.Debug(ctx, "Cache lookup with tracking_id", "tracking_id", trackingID)

	// Add tracking ID to span
	s.obs.TracerService.SetAttributes(span, map[string]string{
		"tracking_id": trackingID,
	})

	// Get from cache in a separate span
	cacheCtx, cacheSpan := s.obs.TracerService.StartTracer(ctx, "GetCache")
	val, found, err := s.cacheclient.GetCache(cacheCtx, config.SERVICE_NAME, key, trackingID)
	if err != nil {
		s.obs.TracerService.SetStatus(cacheSpan, otelcodes.Error, "Failed to get cache")
		s.obs.TracerService.SetAttributes(cacheSpan, map[string]string{
			"error": err.Error(),
		})
		s.obs.LoggerService.Error(ctx, "Failed to get cache", "error", err, "namespace", config.SERVICE_NAME, "key", key)
		s.obs.TracerService.StopSpan(cacheSpan)

		// Set parent span status
		s.obs.TracerService.SetStatus(span, otelcodes.Error, "Failed to get cache")

		return nil, fmt.Errorf("failed to get cache: %w", err)
	}
	if !found {
		s.obs.TracerService.SetStatus(cacheSpan, otelcodes.Error, "Config not found in cache")
		s.obs.LoggerService.Debug(ctx, "Config not found in cache", "namespace", config.SERVICE_NAME, "key", key)
		s.obs.TracerService.StopSpan(cacheSpan)

		// Set parent span status
		s.obs.TracerService.SetStatus(span, otelcodes.Error, "Config not found in cache")

		return nil, fmt.Errorf("config not found in cache")
	}
	s.obs.TracerService.SetStatus(cacheSpan, otelcodes.Ok, "Successfully retrieved from cache")
	s.obs.TracerService.StopSpan(cacheSpan)

	s.obs.LoggerService.Debug(ctx, "Successfully retrieved data from cache", "data_length", len(val))

	// Unmarshal in a separate span
	_, unmarshalSpan := s.obs.TracerService.StartTracer(ctx, "UnmarshalConfig")
	var cfg config.Config
	if err := json.Unmarshal([]byte(val), &cfg); err != nil {
		s.obs.TracerService.SetStatus(unmarshalSpan, otelcodes.Error, "Failed to unmarshal cached config")
		s.obs.TracerService.SetAttributes(unmarshalSpan, map[string]string{
			"error": err.Error(),
		})
		s.obs.LoggerService.Error(ctx, "Failed to unmarshal cached config", "error", err)
		s.obs.TracerService.StopSpan(unmarshalSpan)

		// Set parent span status
		s.obs.TracerService.SetStatus(span, otelcodes.Error, "Failed to unmarshal cached config")

		return nil, fmt.Errorf("failed to unmarshal cached config: %w", err)
	}
	s.obs.TracerService.SetStatus(unmarshalSpan, otelcodes.Ok, "Successfully unmarshaled config")
	s.obs.TracerService.StopSpan(unmarshalSpan)

	// Record metrics for the operation
	processingTime := time.Since(startTime).Milliseconds()
	s.obs.MetricsService.RecordHistogram(ctx, "cache_get_processing_time_ms", float64(processingTime), nil)

	// Set overall span status
	s.obs.TracerService.SetStatus(span, otelcodes.Ok, "Successfully retrieved data from cache")

	return &cfg, nil
}
