package redis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

type RedisServiceInterface interface {
	Get(key string, ctx context.Context) (string, error)
	Set(key string, value []byte, expiration time.Duration, ctx context.Context) error
}

func RedisClient(redisHost, redisPort, redisUsername, redisPassword string) (*redis.Client, error) {
	redisURL := fmt.Sprintf("%s:%s", redisHost, redisPort)

	log.Println("redisURL", redisURL)

	client := redis.NewClient(&redis.Options{
		Addr: redisURL,
		// Username: redisUsername,
		// Password: redisPassword,
		DB: 0,
	})

	if os.Getenv("ENABLE_TRACING") == "true" {
		// Instrument the Redis client with OpenTelemetry
		if err := redisotel.InstrumentTracing(client); err != nil {
			return nil, fmt.Errorf("failed to instrument Redis client with tracing: %w", err)
		}
	}

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("✨ Connected to Redis")
	return client, nil
}

type Config struct {
	Address string
	DB      int
}

func (c *Config) Validate() error {
	if c.Address == "" {
		return errors.New("address cannot be empty")
	}
	if c.DB < 0 {
		return errors.New("DB index cannot be negative")
	}
	return nil
}
