package redis


import (
	"context"
	"fmt"
	"time"

	redisService "cra-protos/redis_service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// RedisClientStruct represents a Redis cache client
type RedisClientStruct struct {
	conn   *grpc.ClientConn
	client redisService.CacheServiceClient
}

// NewClient creates a new Redis cache client
func NewRedisClient(redisServiceAddr string) (*RedisClientStruct, error) {
	conn, err := grpc.NewClient(
		redisServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis service: %w", err)
	}

	client := redisService.NewCacheServiceClient(conn)

	return &RedisClientStruct{
		conn:   conn,
		client: client,
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
func (c *RedisClientStruct) SetCache(ctx context.Context, namespace, key, value string, ttl time.Duration) error {
	req := &redisService.SetCacheRequest{
		Namespace: namespace,
		Key:       key,
		Value:     value,
		Ttl:       int64(ttl.Seconds()),
	}

	resp, err := c.client.SetCache(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to set cache: %s", resp.Message)
	}

	return nil
}

// GetCache retrieves a value from the cache
func (c *RedisClientStruct) GetCache(ctx context.Context, namespace, key string) (string, bool, error) {
	req := &redisService.GetCacheRequest{
		Namespace: namespace,
		Key:       key,
	}

	resp, err := c.client.GetCache(ctx, req)
	if err != nil {
		return "", false, fmt.Errorf("failed to get cache: %w", err)
	}

	if !resp.Found {
		return "", false, nil
	}

	return resp.Value, true, nil
}

// InvalidateCache invalidates a cache entry
func (c *RedisClientStruct) InvalidateCache(ctx context.Context, namespace, key string) error {
	req := &redisService.InvalidateCacheRequest{
		Namespace: namespace,
		Key:       key,
	}

	resp, err := c.client.InvalidateCache(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to invalidate cache: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to invalidate cache: %s", resp.Message)
	}

	return nil
}