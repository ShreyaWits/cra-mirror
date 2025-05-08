package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2" // Import miniredis
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisClaimRepository_IncrementClaimSequence(t *testing.T) {
	// Start a miniredis server
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close() // Stop the server when the test finishes

	// Create a redis client connected to the miniredis server
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(), // Use the miniredis server address
	})
	defer redisClient.Close()

	// Create the repository with the miniredis client
	repo := NewRedisClaimRepository(redisClient)
	ctx := context.Background()
	datePart := "20230101"
	key := "claim:sequence:" + datePart

	// Call the method
	value, err := repo.IncrementClaimSequence(ctx, datePart)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, int64(1), value) // The first increment should return 1

	// Verify the value in miniredis
	val, err := mr.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, "1", val)

	// Call again to check increment
	value, err = repo.IncrementClaimSequence(ctx, datePart)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), value)

	// Verify the value in miniredis again
	val, err = mr.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, "2", val)
}

func TestRedisClaimRepository_SetClaimSequenceExpiry(t *testing.T) {
	// Start a miniredis server
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close() // Stop the server when the test finishes

	// Create a redis client connected to the miniredis server
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(), // Use the miniredis server address
	})
	defer redisClient.Close()

	// Create the repository with the miniredis client
	repo := NewRedisClaimRepository(redisClient)
	ctx := context.Background()
	datePart := "20230101"
	key := "claim:sequence:" + datePart
	expiry := 24 * time.Hour

	// First, set a value so the key exists
	mr.Set(key, "10")

	// Call the method
	err = repo.SetClaimSequenceExpiry(ctx, datePart, expiry)

	// Assertions
	assert.NoError(t, err)

	// Verify the expiry in miniredis
	ttl := mr.TTL(key)
	// miniredis TTL might not be exactly the duration set, but should be close
	// Let's check if it's within a reasonable range, e.g., 23-24 hours
	assert.True(t, ttl > 23*time.Hour && ttl <= 24*time.Hour, "Expected TTL between 23h and 24h, got %s", ttl)

	// Test setting expiry on a non-existent key (should return true according to redis EXPIRE command)
	nonExistentKey := "non:existent:key"
	err = repo.SetClaimSequenceExpiry(ctx, nonExistentKey, expiry)
	assert.NoError(t, err) // Expire on non-existent key does not return an error in go-redis, just false for the bool cmd result. The repository method returns the error from Expire, which is nil.

	// Verify that the non-existent key still doesn't exist and has no TTL
	exists := mr.Exists(nonExistentKey)
	assert.False(t, exists)
}