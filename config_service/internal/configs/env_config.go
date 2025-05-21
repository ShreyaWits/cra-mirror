package configs

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	EtcdEndpoint     string
	TemporalEndpoint string
	DatabaseHost     string
	DatabasePort     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
	AdminSecret      string
	JWTSecret        string
	APISecret        string
}

var AppConfig Config

// LoadConfig loads environment variables, validates them, and populates AppConfig
func LoadConfig() error {
	if os.Getenv("IS_DOCKER") != "true" {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: No .env file found. Proceeding without it. Error: %v", err)
			return err
		} else {
			log.Println("Loaded .env file")
		}
	}

	// Validate environment variables
	if err := ValidateEnv(); err != nil {
		log.Fatalf("❌ Environment validation failed: %v", err)
		return err
	}

	// Load values into AppConfig
	AppConfig = Config{
		Port:             os.Getenv("REST_PORT"),
		EtcdEndpoint:     os.Getenv("ETCD_SERVER_ENDPOINT"),
		TemporalEndpoint: os.Getenv("TEMPORAL_SERVER_ENDPOINT"),
		DatabaseHost:     os.Getenv("YUGABYTE_DATABASE_HOST"),
		DatabasePort:     os.Getenv("YUGABYTE_DATABASE_PORT"),
		DatabaseUser:     os.Getenv("YUGABYTE_DATABASE_USER"),
		DatabasePassword: os.Getenv("YUGABYTE_DATABASE_PASSWORD"),
		DatabaseName:     os.Getenv("CONFIG_SERVICE_YUGABYTE_DATABASE_NAME"),
		AdminSecret:      os.Getenv("CONFIG_SERVICE_ADMIN_SECRET"),
		JWTSecret:        os.Getenv("CONFIG_SERVICE_JWT_SECRET"),
	}
	return nil
}

// EnvRules contains validation functions for each environment variable
var EnvRules = map[string]func(string) error{
	"YUGABYTE_DATABASE_HOST": func(value string) error {
		if value == "" {
			return fmt.Errorf("YUGABYTE_DATABASE_HOST cannot be empty")
		}
		return nil
	},
	"YUGABYTE_DATABASE_PORT": func(value string) error {
		if value == "" {
			return fmt.Errorf("YUGABYTE_DATABASE_PORT cannot be empty")
		}
		return nil
	},
	"YUGABYTE_DATABASE_USER": func(value string) error {
		if value == "" {
			return fmt.Errorf("YUGABYTE_DATABASE_USER cannot be empty")
		}
		return nil
	},
	"YUGABYTE_DATABASE_PASSWORD": func(value string) error {
		if value == "" {
			return fmt.Errorf("YUGABYTE_DATABASE_PASSWORD cannot be empty")
		}
		return nil
	},
	"CONFIG_SERVICE_YUGABYTE_DATABASE_NAME": func(value string) error {
		if value == "" {
			return fmt.Errorf("CONFIG_SERVICE_YUGABYTE_DATABASE_NAME cannot be empty")
		}
		return nil
	},
	"ETCD_SERVER_ENDPOINT": func(value string) error {
		if value == "" {
			return fmt.Errorf("ETCD_SERVER_ENDPOINT cannot be empty")
		}
		return nil
	},
	"TEMPORAL_SERVER_ENDPOINT": func(value string) error {
		if value == "" {
			return fmt.Errorf("TEMPORAL_SERVER_ENDPOINT cannot be empty")
		}
		return nil
	},
	"REST_PORT": func(value string) error {
		if value == "" {
			return fmt.Errorf("REST_PORT cannot be empty")
		}
		port, err := strconv.Atoi(value)
		if err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("REST_PORT must be an integer between 1 and 65535")
		}
		return nil
	},
	"CONFIG_SERVICE_ADMIN_SECRET": func(value string) error {
		if value == "" {
			return fmt.Errorf("CONFIG_SERVICE_ADMIN_SECRET cannot be empty")
		}
		return nil
	},
	"CONFIG_SERVICE_JWT_SECRET": func(value string) error {
		if value == "" {
			return fmt.Errorf("CONFIG_SERVICE_JWT_SECRET cannot be empty")
		}
		return nil
	},
}


// ValidateEnv iterates over EnvRules and checks each variable
func ValidateEnv() error {
	for key, validate := range EnvRules {
		value := os.Getenv(key)
		if err := validate(value); err != nil {
			return fmt.Errorf("invalid value for %s: %v", key, err)
		}
	}
	return nil
}
