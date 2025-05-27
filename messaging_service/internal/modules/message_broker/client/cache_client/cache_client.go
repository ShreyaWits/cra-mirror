package cacheclient

import (
	"context"
	pb "cra-protos/redis_service"
	"fmt"
	"messaging_service/pkg/errors"
	"messaging_service/pkg/observability"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	// Metric names
	metricCacheSetTotal          = "redis_set_total"
	metricCacheSetSuccess        = "redis_set_success"
	metricCacheSetFailure        = "redis_set_failure"
	metricCacheGetTotal          = "redis_get_total"
	metricCacheGetSuccess        = "redis_get_success"
	metricCacheGetFailure        = "redis_get_failure"
	metricCacheInvalidateTotal   = "redis_invalidate_total"
	metricCacheInvalidateSuccess = "redis_invalidate_success"
	metricCacheInvalidateFailure = "redis_invalidate_failure"

	// Error types
	errorTypeConnection = "connection_error"
	errorTypeOperation  = "operation_error"
	errorTypeNotFound   = "not_found_error"
)

// DialerFunc defines the signature for a function that creates a gRPC connection
type DialerFunc func(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error)

// DefaultDialer implements the default gRPC connection creation
func DefaultDialer(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.NewClient(target, opts...)
}

// CacheServiceClientFactory defines the signature for a function that creates a CacheServiceClient
type CacheServiceClientFactory func(conn *grpc.ClientConn) pb.CacheServiceClient

// DefaultCacheServiceClientFactory is the default factory for creating a CacheServiceClient
func DefaultCacheServiceClientFactory(conn *grpc.ClientConn) pb.CacheServiceClient {
	return pb.NewCacheServiceClient(conn)
}

type RedisClient interface {
	Close() error
	SetCache(ctx context.Context, namespace, key, value string, ttl time.Duration, trackingID string) error
	GetCache(ctx context.Context, namespace, key, trackingID string) (string, bool, error)
	InvalidateCache(ctx context.Context, namespace, key, trackingID string) error
}

// RedisClientStruct represents a Redis cache client
type RedisClientStruct struct {
	conn   *grpc.ClientConn
	client pb.CacheServiceClient
	obs    *observability.ObservabilityStack
}

// NewRedisClient creates a new Redis cache client
func NewRedisClient(redisServiceAddr string, obs *observability.ObservabilityStack) (*RedisClientStruct, error) {
	return NewRedisClientWithOptions(redisServiceAddr, DefaultDialer, DefaultCacheServiceClientFactory, obs)
}

// NewRedisClientWithDialer creates a new Redis cache client using the provided dialer function
func NewRedisClientWithDialer(redisServiceAddr string, dialer DialerFunc, obs *observability.ObservabilityStack) (*RedisClientStruct, error) {
	return NewRedisClientWithOptions(redisServiceAddr, dialer, DefaultCacheServiceClientFactory, obs)
}

// NewRedisClientWithOptions creates a new Redis cache client with custom options
func NewRedisClientWithOptions(redisServiceAddr string, dialer DialerFunc, clientFactory CacheServiceClientFactory, obs *observability.ObservabilityStack) (*RedisClientStruct, error) {
	// Validate input arguments
	if redisServiceAddr == "" {
		return nil, fmt.Errorf("redis service address cannot be empty")
	}

	if obs == nil {
		return nil, fmt.Errorf("observability stack cannot be nil")
	}

	conn, err := dialer(
		redisServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, errors.NewCustomError(errors.CACErrConnectionFailed,
			fmt.Errorf("failed to connect to Redis service: %w", err))
	}

	// Validate connection
	if conn == nil {
		return nil, errors.NewCustomError(errors.CACErrConnectionFailed,
			fmt.Errorf("failed to establish connection to Redis service"))
	}

	client := clientFactory(conn)

	// Validate client creation
	if client == nil {
		conn.Close() // Clean up connection if client creation failed
		return nil, errors.NewCustomError(errors.CACErrConnectionFailed,
			fmt.Errorf("failed to create Redis service client"))
	}

	return &RedisClientStruct{
		conn:   conn,
		client: client,
		obs:    obs,
	}, nil
}

// Close closes the gRPC connection
func (c *RedisClientStruct) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// SetCache sets a value in the cache
func (c *RedisClientStruct) SetCache(ctx context.Context, namespace, key, value string, ttl time.Duration, trackingID string) error {
	// Start tracing
	functionName := "RedisSetCache"
	tCtx, span := c.obs.TracerService.StartTracer(ctx, functionName)
	defer c.obs.TracerService.StopSpan(span)

	// Set tracing attributes
	c.obs.TracerService.SetAttributes(span, map[string]string{
		"operation":   functionName,
		"namespace":   namespace,
		"key":         key,
		"tracking_id": trackingID,
	})

	// Increment total metric
	c.obs.MetricsService.IncrementCounter(tCtx, metricCacheSetTotal, 1, nil)

	c.obs.LoggerService.Debug(tCtx, fmt.Sprintf("[%s] Setting cache value for namespace: %s, key: %s",
		trackingID, namespace, key))

	req := &pb.SetCacheRequest{
		Namespace: namespace,
		Key:       key,
		Value:     value,
		Ttl:       int64(ttl.Seconds()),
	}

	// Use traced context to propagate the trace
	resp, err := c.client.SetCache(tCtx, req)
	if err != nil {
		c.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Failed to set cache value: %s", trackingID, err.Error()))
		c.obs.MetricsService.IncrementCounter(tCtx, metricCacheSetFailure, 1, map[string]string{
			"error_type": errorTypeOperation,
		})
		return errors.NewCustomError(errors.CACErrSetFailed, err)
	}

	if !resp.Success {
		c.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Cache service reported failure: %s", trackingID, resp.Message))
		c.obs.MetricsService.IncrementCounter(tCtx, metricCacheSetFailure, 1, map[string]string{
			"error_type": errorTypeOperation,
		})
		return errors.NewCustomError(errors.CACErrSetFailed,
			fmt.Errorf("failed to set cache: %s", resp.Message))
	}

	c.obs.LoggerService.Debug(tCtx, fmt.Sprintf("[%s] Successfully set cache value", trackingID))
	c.obs.MetricsService.IncrementCounter(tCtx, metricCacheSetSuccess, 1, nil)
	return nil
}

// GetCache retrieves a value from the cache
func (c *RedisClientStruct) GetCache(ctx context.Context, namespace, key, trackingID string) (string, bool, error) {
	// Start tracing
	functionName := "RedisGetCache"
	tCtx, span := c.obs.TracerService.StartTracer(ctx, functionName)
	defer c.obs.TracerService.StopSpan(span)

	// Set tracing attributes
	c.obs.TracerService.SetAttributes(span, map[string]string{
		"operation":   functionName,
		"namespace":   namespace,
		"key":         key,
		"tracking_id": trackingID,
	})

	// Increment total metric
	c.obs.MetricsService.IncrementCounter(tCtx, metricCacheGetTotal, 1, nil)

	c.obs.LoggerService.Debug(tCtx, fmt.Sprintf("[%s] Getting cache value for namespace: %s, key: %s",
		trackingID, namespace, key))

	req := &pb.GetCacheRequest{
		Namespace: namespace,
		Key:       key,
	}

	// Use traced context to propagate the trace
	resp, err := c.client.GetCache(tCtx, req)
	if err != nil {
		c.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Failed to get cache value: %s", trackingID, err.Error()))
		c.obs.MetricsService.IncrementCounter(tCtx, metricCacheGetFailure, 1, map[string]string{
			"error_type": errorTypeOperation,
		})
		return "", false, errors.NewCustomError(errors.CACErrGetFailed, err)
	}

	if !resp.Found {
		c.obs.LoggerService.Debug(tCtx, fmt.Sprintf("[%s] Cache value not found", trackingID))
		c.obs.MetricsService.IncrementCounter(tCtx, metricCacheGetFailure, 1, map[string]string{
			"error_type": errorTypeNotFound,
		})
		// Don't return an error, just indicate not found with the boolean
		return "", false, nil
	}

	c.obs.LoggerService.Debug(tCtx, fmt.Sprintf("[%s] Successfully retrieved cache value", trackingID))
	c.obs.MetricsService.IncrementCounter(tCtx, metricCacheGetSuccess, 1, nil)
	return resp.Value, true, nil
}

// InvalidateCache invalidates a cache entry
func (c *RedisClientStruct) InvalidateCache(ctx context.Context, namespace, key, trackingID string) error {
	// Start tracing
	functionName := "RedisInvalidateCache"
	tCtx, span := c.obs.TracerService.StartTracer(ctx, functionName)
	defer c.obs.TracerService.StopSpan(span)

	// Set tracing attributes
	c.obs.TracerService.SetAttributes(span, map[string]string{
		"operation":   functionName,
		"namespace":   namespace,
		"key":         key,
		"tracking_id": trackingID,
	})

	// Increment total metric
	c.obs.MetricsService.IncrementCounter(tCtx, metricCacheInvalidateTotal, 1, nil)

	c.obs.LoggerService.Debug(tCtx, fmt.Sprintf("[%s] Invalidating cache for namespace: %s, key: %s",
		trackingID, namespace, key))

	req := &pb.InvalidateCacheRequest{
		Namespace: namespace,
		Key:       key,
	}

	// Use traced context to propagate the trace
	resp, err := c.client.InvalidateCache(tCtx, req)
	if err != nil {
		c.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Failed to invalidate cache: %s", trackingID, err.Error()))
		c.obs.MetricsService.IncrementCounter(tCtx, metricCacheInvalidateFailure, 1, map[string]string{
			"error_type": errorTypeOperation,
		})
		return errors.NewCustomError(errors.CACErrInvalidateFailed, err)
	}

	if !resp.Success {
		c.obs.LoggerService.Error(tCtx, fmt.Sprintf("[%s] Cache invalidation reported failure: %s", trackingID, resp.Message))
		c.obs.MetricsService.IncrementCounter(tCtx, metricCacheInvalidateFailure, 1, map[string]string{
			"error_type": errorTypeOperation,
		})
		return errors.NewCustomError(errors.CACErrInvalidateFailed,
			fmt.Errorf("failed to invalidate cache: %s", resp.Message))
	}

	c.obs.LoggerService.Debug(tCtx, fmt.Sprintf("[%s] Successfully invalidated cache", trackingID))
	c.obs.MetricsService.IncrementCounter(tCtx, metricCacheInvalidateSuccess, 1, nil)
	return nil
}
