package config

import (
	"context"
	"fmt"
	pb "nps-reciept-service/proto/redis_service"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
}

// NewRedisClient creates a new Redis cache client
func NewRedisClient(redisServiceAddr string) (*RedisClientStruct, error) {
	return NewRedisClientWithOptions(redisServiceAddr, DefaultDialer, DefaultCacheServiceClientFactory)
}

// NewRedisClientWithDialer creates a new Redis cache client using the provided dialer function
func NewRedisClientWithDialer(redisServiceAddr string, dialer DialerFunc) (*RedisClientStruct, error) {
	return NewRedisClientWithOptions(redisServiceAddr, dialer, DefaultCacheServiceClientFactory)
}

// NewRedisClientWithOptions creates a new Redis cache client with custom options
func NewRedisClientWithOptions(redisServiceAddr string, dialer DialerFunc, clientFactory CacheServiceClientFactory) (*RedisClientStruct, error) {
	// Validate input arguments
	if redisServiceAddr == "" {
		return nil, fmt.Errorf("redis service address cannot be empty")
	}

	conn, err := dialer(
		redisServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis service: %w", err)
	}

	// Validate connection
	if conn == nil {
		return nil, fmt.Errorf("failed to establish connection to Redis service")
	}

	client := clientFactory(conn)

	// Validate client creation
	if client == nil {
		conn.Close() // Clean up connection if client creation failed
		return nil, fmt.Errorf("failed to create Redis service client")
	}

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
func (c *RedisClientStruct) SetCache(ctx context.Context, namespace, key, value string, ttl time.Duration, trackingID string) error {
	req := &pb.SetCacheRequest{
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
func (c *RedisClientStruct) GetCache(ctx context.Context, namespace, key, trackingID string) (string, bool, error) {
	req := &pb.GetCacheRequest{
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
func (c *RedisClientStruct) InvalidateCache(ctx context.Context, namespace, key, trackingID string) error {
	req := &pb.InvalidateCacheRequest{
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
