package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

// InitRedis initializes the Redis client with the given configuration
func InitRedis(host, port, username, password, ttl string) {

	// Validate TTL format
	_, err := time.ParseDuration(ttl)
	if err != nil {
		log.Fatalf("Invalid RECEIPT_SERVICE_REDIS_TTL: %v", err)
	}
	//print the insctance of the redis connection
	fmt.Println("host", host)
	fmt.Println("port", port)
	fmt.Println("username", username)
	fmt.Println("password", password)
	fmt.Println("ttl", ttl)
	redisClient = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port),

		// Username: username,
		// Password: password,
		DB:       0,
	})

	// Test the connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Printf("Connected to Redis at %s", host)
}

// GetRedisClient returns the Redis client instance
func GetRedisClient() *redis.Client {
	return redisClient
}
