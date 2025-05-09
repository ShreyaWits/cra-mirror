package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

var (
	AppConfig *Config
)

type Config struct {
	HTTPListenAddress string `validate:"required" env:"HTTP_LISTEN_ADDRESS"`
	HTTPListenPort    int    `validate:"required,min=1,max=65535" env:"HTTP_LISTEN_PORT"`

	GRPCListenAddress string `validate:"required" env:"GRPC_LISTEN_ADDRESS"`
	GRPCListenPort    int    `validate:"required,min=1,max=65535" env:"GRPC_LISTEN_PORT"`

	DatabaseHost     string `validate:"required" env:"DATABASE_HOST"`
	DatabasePort     int    `validate:"required,min=1,max=65535" env:"DATABASE_PORT"`
	DatabaseUser     string `validate:"required" env:"DATABASE_USER"`
	DatabasePassword string `validate:"required" env:"DATABASE_PASSWORD"`
	DatabaseName     string `validate:"required" env:"DATABASE_NAME"`
}

func LoadConfig() (*Config, error) {
	// Load .env file only if not running in Docker
	if os.Getenv("IS_DOCKER") != "true" {
		if err := godotenv.Load(); err != nil {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	}

	config := &Config{
		HTTPListenAddress: os.Getenv("HTTP_LISTEN_ADDRESS"),
		HTTPListenPort:    getEnvInt("HTTP_LISTEN_PORT"),

		GRPCListenAddress: os.Getenv("GRPC_LISTEN_ADDRESS"),
		GRPCListenPort:    getEnvInt("GRPC_LISTEN_PORT"),

		DatabaseHost:     os.Getenv("DATABASE_HOST"),
		DatabasePort:     getEnvInt("DATABASE_PORT"),
		DatabaseUser:     os.Getenv("DATABASE_USER"),
		DatabasePassword: os.Getenv("DATABASE_PASSWORD"),
		DatabaseName:     os.Getenv("DATABASE_NAME"),
	}

	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

func getEnvInt(key string) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return 0
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.DatabaseHost,
		c.DatabasePort,
		c.DatabaseUser,
		c.DatabasePassword,
		c.DatabaseName,
	)
}

func (c *Config) GetHTTPListenAddress() string {
	return fmt.Sprintf("%s:%d",
		AppConfig.HTTPListenAddress,
		AppConfig.HTTPListenPort,
	)
}

func (c *Config) GetGRPCListenAddress() string {
	return fmt.Sprintf("%s:%d",
		AppConfig.GRPCListenAddress,
		AppConfig.GRPCListenPort,
	)
}

func init() {
	var err error

	AppConfig, err = LoadConfig()
	if err != nil {
		log.Fatalf("error loading env variables: %v\n", err)
	}
}
