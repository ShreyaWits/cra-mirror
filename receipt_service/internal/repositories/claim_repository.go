package repositories

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// ClaimRepository defines the interface for claim data operations
type ClaimRepository interface {
	IncrementClaimSequence(ctx context.Context, datePart string) (int64, error)
	SetClaimSequenceExpiry(ctx context.Context, datePart string, expiry time.Duration) error
}

// RedisClaimRepository implements ClaimRepository using Redis
type RedisClaimRepository struct {
	redisClient *redis.Client
}

// NewRedisClaimRepository creates a new RedisClaimRepository
func NewRedisClaimRepository(client *redis.Client) *RedisClaimRepository {
	return &RedisClaimRepository{
		redisClient: client,
	}
}

// IncrementClaimSequence increments the claim sequence for a given date part
func (r *RedisClaimRepository) IncrementClaimSequence(ctx context.Context, datePart string) (int64, error) {
	key := "claim:sequence:" + datePart
	return r.redisClient.Incr(ctx, key).Result()
}

// SetClaimSequenceExpiry sets the expiry for the claim sequence key
func (r *RedisClaimRepository) SetClaimSequenceExpiry(ctx context.Context, datePart string, expiry time.Duration) error {
	key := "claim:sequence:" + datePart
	return r.redisClient.Expire(ctx, key, expiry).Err()
}