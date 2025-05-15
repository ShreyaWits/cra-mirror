package config

import (
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	KAFKA_SERVER_URL string
	REDIS_HOST       string
	REDIS_PORT       string
	REDIS_USERNAME   string
	REDIS_PASSWORD   string
}

var (
	cfg  *Config
	once sync.Once
)

// LoadConfig loads environment variables into the Config struct (singleton style)
func LoadConfig() *Config {
	once.Do(func() {
		err := godotenv.Load()
		if err != nil {
			log.Println("No .env file found, falling back to system env vars")
		}

		cfg = &Config{
			KAFKA_SERVER_URL: GetEnv("KAFKA_SERVER_URL", "kafka:29092"),
			REDIS_HOST:       GetEnv("REDIS_HOST", "localhost"),
			REDIS_PORT:       GetEnv("REDIS_PORT", "6379"),
			REDIS_USERNAME:   GetEnv("REDIS_USERNAME", ""),
			REDIS_PASSWORD:   GetEnv("REDIS_PASSWORD", ""),
		}
	})
	return cfg
}

func GetEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
