package configEnv

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Default values
// config := &Config{
// 	ServerPort:         getEnv("REST_PORT", "8080"),
// 	GRPCPort:           getEnv("GRPC_PORT", "50051"),
// 	RedisHost:          getEnv("REDIS_HOST", "localhost"),
// 	RedisPort:          getEnv("REDIS_PORT", "6379"),
// 	RedisUser:          getEnv("REDIS_USER", "templateuser"),
// 	RedisPassword:      getEnv("REDIS_PASSWORD", "templatepassword"),
// 	JWTSecret:          getEnv("JWT_SECRET", "myTemplateSecureKey1234567890@GoLan"),
// 	YugabyteDBHost:     getEnv("YUGABYTE_DATABASE_HOST", "yugabyte"),
// 	YugabyteDBPort:     getEnv("YUGABYTE_DATABASE_PORT", "5433"),
// 	YugabyteDBUser:     getEnv("YUGABYTE_DATABASE_USER", "yugabyte"),
// 	YugabyteDBPassword: getEnv("YUGABYTE_DATABASE_PASSWORD", "yugabyte"),
// 	YugabyteDBName:     getEnv("TEMPLATE_SERVICE_YUGABYTE_DATABASE_NAME", "yugabyte"),
// 	ServiceName:        getEnv("SERVICE_NAME", "template_service"),
// 	ServiceVersion:     getEnv("SERVICE_VERSION", "1.0.0"),
// 	CacheUrl:           getEnv("CACHE_URL", "localhost:50051"),
// 	CacheTtl:           getEnv("CACHE_TTL", "240"),
// 	ObservabilityUrl:   getEnv("OBSERVABILITY_URL", "http://localhost:4317"),
// 	Environment:        getEnv("ENVIRONMENT", "dev"),
// }

// LoadConfig loads configuration from environment variables
// func LoadConfig() (*Config, error) {
// 	// Load .env file if it exists
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Println(".env file not found, falling back to system env")
// 	}

// 	return config, nil
// }

var cbList []func(*Config, *ImmutableConfig) = []func(*Config, *ImmutableConfig){}
var preRestartHooks []func() = []func(){}

// var Configs *Config = nil
var Configs *Config = nil
var ImmutableConfigs *ImmutableConfig = nil

func RefreshConfig(config *Config) {
	Configs = config

	for _, cb := range preRestartHooks {
		cb()
	}

	for _, cb := range cbList {
		cb(Configs, ImmutableConfigs)
	}
}

func LoadConfig() error {
	if os.Getenv("IS_DOCKER") != "true" {
        if err := godotenv.Load(); err != nil {
            fmt.Printf("Warning: No .env file found. Proceeding without it. Error: %v AND IS_DOCKER not true", err)
            return err
        } else {
            fmt.Println("Loaded .env file")
        }
    }

	ImmutableConfigs = &ImmutableConfig{
		ConfigServiceUrl:   getEnv("CONFIG_SERVICE_URL", "http://localhost:4001/api/v1"),
		ConfigServiceToken: getEnv("CONFIG_SERVICE_TOKEN", "kjdasklfjklajdkfljkl"),
		Environment:        getEnv("ENVIRONMENT", "development"),
		CacheSrvAddr:       getEnv("CACHE_SERVICE_ADDR", "localhost:6379"),
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func ListenForConfigChanges(cb func(*Config, *ImmutableConfig)) {
	cbList = append(cbList, cb)
}

func RegisterPreRestartHook(cb func()) {
	preRestartHooks = append(preRestartHooks, cb)
}

func ClearPreRestartHooks() {
	preRestartHooks = []func(){}
}
