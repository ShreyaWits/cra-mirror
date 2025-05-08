package config

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestInitRedisAndGetRedisClient(t *testing.T) {
	// Start a miniredis server
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close() // Stop the server when the test finishes

	// Initialize Redis with the miniredis server address
	InitRedis(mr.Addr())

	// Get the Redis client
	client := GetRedisClient()

	// Verify that the client is not nil
	assert.NotNil(t, client)

	// Verify that the client is connected to the miniredis server by sending a PING command
	ctx := context.Background()
	pong, err := client.Ping(ctx).Result()
	assert.NoError(t, err)
	assert.Equal(t, "PONG", pong)
}