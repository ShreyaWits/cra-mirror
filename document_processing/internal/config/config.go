//go:build !test

// build +test
package config

import (
	"context"
	"document_processing/pkg/redis"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

var (
	configChangeChan = make(chan struct{})
	configMutex      sync.RWMutex
	currentConfig    *Config
	ImmutableConfigs *ImmutableConfig
	validate         *validator.Validate
)

type Config struct {
	ServerPort       int
	MinioEndpoint    string
	MinioAccessKey   string
	MinioSecretKey   string
	MinioBucketName  string
	GeminiAPIKey     string
	YugabyteHost     string
	YugabyteUser     string
	YugabytePassWord string
	YugabyteName     string
	YugabytePort     string
	LlamaModelName   string
	LlamaApiURL      string
	ObservabilityUrl string
}

func init() {
	// Initialize validator
	validate = validator.New()

	// Register field name function for better error messages
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return fld.Name
		}
		return name
	})
}

type ConfigResponse struct {
	StatusCode int               `json:"status_code"`
	Message    string            `json:"message"`
	Data       map[string]string `json:"data"`
}
type ImmutableConfig struct {
	ConfigServiceUrl   string
	ConfigServiceToken string
	Environment        string
	CacheSrvAddr       string
	RestPort           string
	CacheTtl           string
}

type ConfigWebhookData struct {
	Environment string            `json:"environment"`
	Method      string            `json:"method"`
	ServiceName string            `json:"serviceName"`
	Values      map[string]string `json:"values"`
}

// GetConfigChangeChan returns the channel that signals config changes
func GetConfigChangeChan() <-chan struct{} {
	return configChangeChan
}

// SetCurrentConfig sets the current config and notifies listeners
func SetCurrentConfig(cfg *Config) {
	configMutex.Lock()
	defer configMutex.Unlock()
	currentConfig = cfg
	select {
	case configChangeChan <- struct{}{}:
	default:
		// Channel is full, which means a notification is already pending
	}
}

// GetCurrentConfig returns the current config
func GetCurrentConfig() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return currentConfig
}

type ConfigWebhook struct {
	redisClient redis.RedisClient
	ttl         time.Duration
}

func NewConfigWebhook(redisClient redis.RedisClient, tl time.Duration) *ConfigWebhook {
	return &ConfigWebhook{
		redisClient: redisClient,
		ttl:         tl,
	}
}

// HandleConfigWebhook processes incoming webhook requests for config updates
func (c *ConfigWebhook) HandleConfigWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var webhookData ConfigWebhookData
	if err := json.NewDecoder(r.Body).Decode(&webhookData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate environment matches
	if webhookData.Environment != os.Getenv("ENVIRONMENT") {
		http.Error(w, "Environment mismatch", http.StatusBadRequest)
		return
	}

	// Create new config from webhook data
	newConfig, err := NewConfig(webhookData.Values)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid config data: %v", err), http.StatusBadRequest)
		return
	}

	// Update current config and notify listeners
	SetCurrentConfig(newConfig)

	data, err := json.Marshal(newConfig)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal config: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("Attempting to cache config from webhook with TTL: %v", c.ttl)
	if err := c.redisClient.SetCache(context.Background(), "document_processing", "config", string(data), c.ttl); err != nil {
		log.Printf("Failed to cache config from webhook: %v", err)
		http.Error(w, fmt.Sprintf("Failed to cache config: %v", err), http.StatusInternalServerError)
		return
	}
	log.Println("Successfully cached config from webhook in Redis")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "Config updated successfully"})
}

// LoadConfigFromAPI fetches config from the remote config service
func LoadConfigFromAPI(env *ImmutableConfig, redisClient redis.RedisClient) (map[string]string, error) {
	log.Printf("Attempting to fetch config from Redis with TTL: %s", env.CacheTtl)
	config, found, err := redisClient.GetCache(context.Background(), "document_processing", "config")

	if err != nil {
		log.Printf("Error fetching config from Redis: %v", err)
		return nil, err
	}

	if found {
		log.Println("Config found in Redis")
		var configResp ConfigResponse
		if err := json.Unmarshal([]byte(config), &configResp); err != nil {
			log.Printf("Error unmarshalling config data: %v", err)
			return nil, err
		}
		return configResp.Data, nil
	}

	log.Println("Config not found in Redis, fetching from API")
	client := &http.Client{}
	api := fmt.Sprintf("%s/%s/document-processing", strings.TrimRight(os.Getenv("CONFIG_SERVICE_URL"), "/"), os.Getenv("ENVIRONMENT"))
	log.Printf("Fetching config from API: %s", api)
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+env.ConfigServiceToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var configResp ConfigResponse
	if err := json.Unmarshal(body, &configResp); err != nil {
		return nil, err
	}

	// Convert all values to strings
	stringData := make(map[string]string)
	for k, v := range configResp.Data {
		stringData[k] = fmt.Sprintf("%v", v)
	}
	configResp.Data = stringData

	data, err := json.Marshal(configResp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cache data: %w", err)
	}

	tl, err := time.ParseDuration(env.CacheTtl)
	if err != nil {
		log.Printf("Error parsing cache TTL: %v, using default 50m", err)
		tl = 50 * time.Minute
	}
	log.Printf("Attempting to cache config with TTL: %v", tl)

	if err := redisClient.SetCache(context.Background(), "document_processing", "config", string(data), tl); err != nil {
		log.Printf("Failed to cache config: %v", err)
		// Don't return error here, just log it and continue
	} else {
		log.Println("Successfully cached config in Redis")
	}

	return configResp.Data, nil
}

func LoadConfig() (*ImmutableConfig, error) {
	if os.Getenv("IS_DOCKER") != "true" {
		if err := godotenv.Load(); err != nil {
			log.Fatalf("error loading environment variables: %v\n", err)
		}
	}

	ImmutableConfigs = &ImmutableConfig{
		ConfigServiceUrl:   getEnv("CONFIG_SERVICE_URL", "http://localhost:4001/api/v1"),
		ConfigServiceToken: getEnv("DOCUMENT_CONFIG_SERVICE_TOKEN", "sadahgdhadbnabdja"),
		Environment:        getEnv("ENVIRONMENT", "dev"),
		CacheSrvAddr:       getEnv("CACHING_SERVICE_GRPC_URL", "localhost:6279"),
		RestPort:           getEnv("DOCUMENT_SERVICE_REST_PORT", "8080"),
		CacheTtl:           getEnv("DOCUMENT_SERVICE_REDIS_TTL", "50m"),
	}

	if err := validate.Struct(ImmutableConfigs); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if ok {
			// Format validation errors more nicely
			var errorMessages []string
			for _, e := range validationErrors {
				errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' failed validation: %s", e.Field(), e.Tag()))
			}
			return nil, fmt.Errorf("environment validation failed: %s", strings.Join(errorMessages, "; "))
		}
		return nil, fmt.Errorf("environment validation failed: %w", err)
	}
	return ImmutableConfigs, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// NewConfig validates and constructs Config from raw API data
func NewConfig(data map[string]string) (*Config, error) {
	requiredKeys := []string{
		"MINIO_ENDPOINT",
		"MINIO_ACCESS_KEY",
		"MINIO_SECRET_KEY",
		"MINIO_BUCKET_NAME",
		"GEMINI_API_KEY",
		"GRPC_PORT",
		"YUGABYTE_DATABASE_HOST",
		"YUGABYTE_DATABASE_USER",
		"YUGABYTE_DATABASE_PASSWORD",
		"YUGABYTE_DATABASE_NAME",
		"YUGABYTE_DATABASE_PORT",
		"LLAMA_MODEL_NAME",
		"LLAMA_API_URL",
		"OTEL_COLLECTOR_GRPC_ENDPOINT",
	}

	// Check all required keys
	for _, key := range requiredKeys {
		if data[key] == "" {
			log.Print("Config Service Token Expired")
			return nil, fmt.Errorf("missing required config key: %s", key)
		}
	}

	// Parse SERVER_PORT to int
	var port int
	_, err := fmt.Sscanf(data["GRPC_PORT"], "%d", &port)
	if err != nil || port <= 0 {
		return nil, errors.New("invalid GRPC_PORT value")
	}

	return &Config{
		ServerPort:       port,
		MinioEndpoint:    data["MINIO_ENDPOINT"],
		MinioAccessKey:   data["MINIO_ACCESS_KEY"],
		MinioSecretKey:   data["MINIO_SECRET_KEY"],
		MinioBucketName:  data["MINIO_BUCKET_NAME"],
		GeminiAPIKey:     data["GEMINI_API_KEY"],
		YugabyteHost:     data["YUGABYTE_DATABASE_HOST"],
		YugabyteUser:     data["YUGABYTE_DATABASE_USER"],
		YugabytePassWord: data["YUGABYTE_DATABASE_PASSWORD"],
		YugabyteName:     data["YUGABYTE_DATABASE_NAME"],
		YugabytePort:     data["YUGABYTE_DATABASE_PORT"],
		LlamaModelName:   data["LLAMA_MODEL_NAME"],
		LlamaApiURL:      data["LLAMA_API_URL"],
		ObservabilityUrl: data["OTEL_COLLECTOR_GRPC_ENDPOINT"],
	}, nil
}
