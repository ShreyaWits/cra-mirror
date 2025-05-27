package service

import (
	"context"
	"encoding/json"
	"fmt"
	"messaging_service/internal/config"
	cacheclient "messaging_service/internal/modules/message_broker/client/cache_client"
	client "messaging_service/internal/modules/message_broker/client/config_client"
	"messaging_service/pkg/errors"
	"messaging_service/pkg/observability"
	"time"

	"github.com/google/uuid"
)

const (
	// Metric names
	metricConfigApiTotal   = "config_api_fetch_total"
	metricConfigApiSuccess = "config_api_fetch_success"
	metricConfigApiFailure = "config_api_fetch_failure"
	metricCacheGetTotal    = "cache_get_total"
	metricCacheGetSuccess  = "cache_get_success"
	metricCacheGetFailure  = "cache_get_failure"
	metricCacheSetTotal    = "cache_set_total"
	metricCacheSetSuccess  = "cache_set_success"
	metricCacheSetFailure  = "cache_set_failure"

	// Error types
	errorTypeApi        = "api_error"
	errorTypeCache      = "cache_error"
	errorTypeMarshal    = "marshal_error"
	errorTypeUnmarshal  = "unmarshal_error"
	errorTypeNotFound   = "not_found_error"
	errorTypeValidation = "validation_error"
)

type ConfigManagerService struct {
	configClient client.ConfigClient
	cacheclient  cacheclient.RedisClient
	env          *config.Env
	obs          *observability.ObservabilityStack
}

// Add this function to allow mocking of JSON marshalling
var jsonMarshal = func(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// GetJsonMarshalFunc returns the current jsonMarshal function for testing
func GetJsonMarshalFunc() func(v interface{}) ([]byte, error) {
	return jsonMarshal
}

// SetJsonMarshalFunc sets the jsonMarshal function to the provided function for testing
func SetJsonMarshalFunc(fn func(v interface{}) ([]byte, error)) {
	jsonMarshal = fn
}

func NewConfigManager(configClient client.ConfigClient, cacheclient cacheclient.RedisClient, obs *observability.ObservabilityStack, env *config.Env) (*ConfigManagerService, error) {
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
func (s *ConfigManagerService) SetEnvironment(env *config.Env) {
	s.env = env
}

func (s *ConfigManagerService) GetFromApiConfiguration(ctx context.Context) (*config.Config, error) {
	// Start tracing
	functionName := "GetFromApiConfiguration"
	tCtx, span := s.obs.TracerService.StartTracer(ctx, functionName)
	defer s.obs.TracerService.StopSpan(span)

	// Generate request ID for correlation
	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())

	// Set tracing attributes
	s.obs.TracerService.SetAttributes(span, map[string]string{
		"request_id": requestID,
		"operation":  functionName,
		"service":    config.SERVICE_NAME,
	})

	// Increment API fetch attempts metric
	s.obs.MetricsService.IncrementCounter(tCtx, metricConfigApiTotal, 1, nil)

	var data *config.Config
	var err error
	var cacheErr error

	// Try to fetch config from API
	s.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Attempting to fetch configuration from API", requestID))
	data, err = s.configClient.FetchConfig(tCtx) // Pass traced context to maintain trace propagation
	if err != nil {
		s.obs.LoggerService.Warn(tCtx, fmt.Sprintf("[%s] API fetch failed, falling back to cache", requestID))
		s.obs.MetricsService.IncrementCounter(tCtx, metricConfigApiFailure, 1, map[string]string{
			"error_type": errorTypeApi,
		})

		// If API fetch fails, try to get from cache
		s.obs.MetricsService.IncrementCounter(tCtx, metricCacheGetTotal, 1, nil)
		data, cacheErr = s.GetDataToCache(tCtx, "config") // Pass traced context
		if cacheErr != nil {
			s.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Failed to retrieve from cache after API failure", requestID))
			s.obs.MetricsService.IncrementCounter(tCtx, metricCacheGetFailure, 1, map[string]string{
				"error_type": errorTypeCache,
			})
			// Return the original API error with proper error code
			return nil, errors.NewCustomError(errors.CFGErrFetchFailed, fmt.Errorf("API fetch failed and cache fallback failed: %w", err))
		}
		s.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Successfully retrieved config from cache after API failure", requestID))
		s.obs.MetricsService.IncrementCounter(tCtx, metricCacheGetSuccess, 1, nil)
	} else {
		s.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Successfully fetched config from API", requestID))
		s.obs.MetricsService.IncrementCounter(tCtx, metricConfigApiSuccess, 1, nil)
	}

	// Validate and set the global config
	if err := config.SetConfig(data); err != nil {
		s.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Error setting global config", requestID))
		s.obs.MetricsService.IncrementCounter(tCtx, metricConfigApiFailure, 1, map[string]string{
			"error_type": errorTypeValidation,
		})
		return data, errors.NewCustomError(errors.CFGErrValidationFailed, err)
	}

	// Store in cache
	s.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Storing config in cache", requestID))
	s.obs.MetricsService.IncrementCounter(tCtx, metricCacheSetTotal, 1, nil)
	err = s.SetDataToCache(tCtx, "config", data) // Pass traced context
	if err != nil {
		s.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Failed to save config in cache", requestID))
		s.obs.MetricsService.IncrementCounter(tCtx, metricCacheSetFailure, 1, map[string]string{
			"error_type": errorTypeCache,
		})
		// Return success but with a warning since we have the data
		s.obs.LoggerService.Warn(tCtx, fmt.Sprintf("[%s] Config fetched successfully but caching failed", requestID))
	} else {
		s.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Successfully stored config in cache", requestID))
		s.obs.MetricsService.IncrementCounter(tCtx, metricCacheSetSuccess, 1, nil)
	}

	return data, nil
}

func (s *ConfigManagerService) SetDataToCache(ctx context.Context, key string, cfg *config.Config) error {
	// Start tracing
	functionName := "SetDataToCache"
	tCtx, span := s.obs.TracerService.StartTracer(ctx, functionName)
	defer s.obs.TracerService.StopSpan(span)

	// Generate request ID and tracking ID for correlation
	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	trackingID := uuid.NewString()

	// Set tracing attributes
	s.obs.TracerService.SetAttributes(span, map[string]string{
		"request_id":  requestID,
		"tracking_id": trackingID,
		"operation":   functionName,
		"service":     config.SERVICE_NAME,
		"key":         key,
	})

	s.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Setting data to cache for key: %s", requestID, key))

	// Use the wrapper function to allow testing
	data, err := jsonMarshal(cfg)
	if err != nil {
		s.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Failed to marshal config for caching", requestID))
		s.obs.MetricsService.IncrementCounter(tCtx, metricCacheSetFailure, 1, map[string]string{
			"error_type": errorTypeMarshal,
		})
		return errors.NewCustomError(errors.CFGErrMarshalFailed, err)
	}

	ttl := time.Duration(s.env.MESSAGING_SERVICE_REDIS_TTL) * time.Hour
	s.obs.LoggerService.Debug(tCtx, fmt.Sprintf("[%s] Cache TTL set to %d hours", requestID, s.env.MESSAGING_SERVICE_REDIS_TTL))

	// Pass traced context to maintain trace propagation
	if err := s.cacheclient.SetCache(tCtx, config.SERVICE_NAME, key, string(data), ttl, trackingID); err != nil {
		s.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Cache set operation failed", requestID))
		return errors.NewCustomError(errors.CACErrSetFailed, err)
	}

	s.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Successfully set data in cache", requestID))
	return nil
}

func (s *ConfigManagerService) GetDataToCache(ctx context.Context, key string) (*config.Config, error) {
	// Start tracing
	functionName := "GetDataToCache"
	tCtx, span := s.obs.TracerService.StartTracer(ctx, functionName)
	defer s.obs.TracerService.StopSpan(span)

	// Generate request ID and tracking ID for correlation
	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	trackingID := uuid.NewString()

	// Set tracing attributes
	s.obs.TracerService.SetAttributes(span, map[string]string{
		"request_id":  requestID,
		"tracking_id": trackingID,
		"operation":   functionName,
		"service":     config.SERVICE_NAME,
		"key":         key,
	})

	s.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Getting data from cache for key: %s", requestID, key))

	// Pass traced context to maintain trace propagation
	val, found, err := s.cacheclient.GetCache(tCtx, config.SERVICE_NAME, key, trackingID)

	if err != nil {
		s.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Failed to get data from cache", requestID))
		s.obs.MetricsService.IncrementCounter(tCtx, metricCacheGetFailure, 1, map[string]string{
			"error_type": errorTypeCache,
		})
		return nil, errors.NewCustomError(errors.CACErrGetFailed, err)
	}
	if !found {
		s.obs.LoggerService.Warn(tCtx, fmt.Sprintf("[%s] Config not found in cache for key: %s", requestID, key))
		s.obs.MetricsService.IncrementCounter(tCtx, metricCacheGetFailure, 1, map[string]string{
			"error_type": errorTypeNotFound,
		})
		return nil, errors.NewCustomError(errors.CACErrNotFound, fmt.Errorf("config not found in cache for key: %s", key))
	}

	s.obs.LoggerService.Info(tCtx, fmt.Sprintf("[%s] Successfully retrieved data from cache", requestID))
	s.obs.MetricsService.IncrementCounter(tCtx, metricCacheGetSuccess, 1, nil)

	var cfg config.Config
	if err := json.Unmarshal([]byte(val), &cfg); err != nil {
		s.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Failed to unmarshal cached config", requestID))
		s.obs.MetricsService.IncrementCounter(tCtx, metricCacheGetFailure, 1, map[string]string{
			"error_type": errorTypeUnmarshal,
		})
		return nil, errors.NewCustomError(errors.CFGErrUnmarshalFailed, err)
	}

	return &cfg, nil
}
