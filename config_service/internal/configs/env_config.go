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
		Port:             os.Getenv("PORT"),
		EtcdEndpoint:     os.Getenv("ETCD_ENDPOINT"),
		TemporalEndpoint: os.Getenv("TEMPORAL_ENDPOINT"),
		DatabaseHost:     os.Getenv("DATABASE_HOST"),
		DatabasePort:     os.Getenv("DATABASE_PORT"),
		DatabaseUser:     os.Getenv("DATABASE_USER"),
		DatabasePassword: os.Getenv("DATABASE_PASSWORD"),
		DatabaseName:     os.Getenv("DATABASE_NAME"),
		AdminSecret:      os.Getenv("ADMIN_SECRET"),
		JWTSecret:        os.Getenv("JWT_SECRET"),
	}
	return nil
}

// EnvRules contains validation functions for each environment variable
var EnvRules = map[string]func(string) error{
	"DATABASE_HOST": func(value string) error {
		if value == "" {
			return fmt.Errorf("DATABASE_HOST cannot be empty")
		}
		return nil
	},
	"DATABASE_PORT": func(value string) error {
		if value == "" {
			return fmt.Errorf("DATABASE_PORT cannot be empty")
		}
		return nil
	},
	"DATABASE_USER": func(value string) error {
		if value == "" {
			return fmt.Errorf("DATABASE_USER cannot be empty")
		}
		return nil
	},
	"DATABASE_PASSWORD": func(value string) error {
		if value == "" {
			return fmt.Errorf("DATABASE_PASSWORD cannot be empty")
		}
		return nil
	},
	"DATABASE_NAME": func(value string) error {
		if value == "" {
			return fmt.Errorf("DATABASE_NAME cannot be empty")
		}
		return nil
	},
	"ETCD_ENDPOINT": func(value string) error {
		if value == "" {
			return fmt.Errorf("ETCD_ENDPOINT cannot be empty")
		}
		return nil
	},
	"TEMPORAL_ENDPOINT": func(value string) error {
		if value == "" {
			return fmt.Errorf("TEMPORAL_ENDPOINT cannot be empty")
		}
		return nil
	},
	"PORT": func(value string) error {
		if value == "" {
			return fmt.Errorf("PORT cannot be empty")
		}
		port, err := strconv.Atoi(value)
		if err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("PORT must be an integer between 1 and 65535")
		}
		return nil
	},
	"ADMIN_SECRET": func(value string) error {
		if value == "" {
			return fmt.Errorf("ADMIN_SECRET cannot be empty")
		}
		return nil
	},
	"JWT_SECRET": func(value string) error {
		if value == "" {
			return fmt.Errorf("JWT_SECRET cannot be empty")
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
