package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	ServerPort int
	RedisHost  string
	RedisPort  string
	RedisUser  string
	RedisPass  string
	RedisTTL   string
	CacheUrl   string
}

type ConfigResponse struct {
	StatusCode int               `json:"status_code"`
	Message    string            `json:"message"`
	Data       map[string]string `json:"data"`
}

type ConfigWebhookData struct {
	Environment string            `json:"environment"`
	Method      string            `json:"method"`
	ServiceName string            `json:"serviceName"`
	Values      map[string]string `json:"values"`
}

var (
	configChangeChan = make(chan struct{})
	configMutex      sync.RWMutex
	currentConfig    *Config
	redisClient      *redis.Client
)

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

// HandleConfigWebhook processes incoming webhook requests for config updates
func HandleConfigWebhook(w http.ResponseWriter, r *http.Request) {
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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "Config updated successfully"})
}

// LoadConfigFromAPI fetches config from the remote config service
func LoadConfigFromAPI(token string) (map[string]string, error) {
	client := &http.Client{}
	api := fmt.Sprintf("%s/%s/receipt-service", strings.TrimRight(os.Getenv("CONFIG_API"), "/"), os.Getenv("ENVIRONMENT"))
	log.Println("this log", api)

	req, err := http.NewRequest("GET", api, nil)
	fmt.Println("Raw requesst from API:", req)

	fmt.Println("Raw requesst from API:", err)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	fmt.Println("Raw response from API:", string(body)) // <
	if err != nil {
		return nil, err
	}

	var configResp ConfigResponse
	if err := json.Unmarshal(body, &configResp); err != nil {
		return nil, err
	}
	fmt.Println("configResp", configResp)
	return configResp.Data, nil
}

// NewConfig validates and constructs Config from raw API data
func NewConfig(data map[string]string) (*Config, error) {
	requiredKeys := []string{
		"SERVER_PORT",
		"REDIS_HOST",
		"REDIS_PORT",
		"REDIS_USERNAME",
		"REDIS_PASSWORD",
		"RECEIPT_SERVICE_REDIS_TTL",
		"CACHING_SERVICE_REST_URL",
	}

	// Check all required keys
	for _, key := range requiredKeys {
		if data[key] == "" {
			return nil, fmt.Errorf("missing required config key: %s", key)
		}
	}

	// Parse SERVER_PORT to int
	var port int
	_, err := fmt.Sscanf(data["SERVER_PORT"], "%d", &port)
	if err != nil || port <= 0 {
		return nil, errors.New("invalid SERVER_PORT value")
	}

	// Validate TTL format
	_, err = time.ParseDuration(data["RECEIPT_SERVICE_REDIS_TTL"])
	if err != nil {
		return nil, fmt.Errorf("invalid RECEIPT_SERVICE_REDIS_TTL: %v", err)
	}

	return &Config{
		ServerPort: port,
		RedisHost:  data["REDIS_HOST"],
		RedisPort:  data["REDIS_PORT"],
		RedisUser:  data["REDIS_USERNAME"],
		RedisPass:  data["REDIS_PASSWORD"],
		RedisTTL:   data["RECEIPT_SERVICE_REDIS_TTL"],
		CacheUrl:   data["CACHING_SERVICE_REST_URL"],
	}, nil
}

// InitRedis initializes the Redis client with the given configuration
func InitRedis(cfg *Config) {
	//print the instance of the redis connection
	fmt.Println("host", cfg.RedisHost)
	fmt.Println("port", cfg.RedisPort)
	fmt.Println("username", cfg.RedisUser)
	fmt.Println("password", cfg.RedisPass)
	fmt.Println("ttl", cfg.RedisTTL)

	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Username: cfg.RedisUser,
		Password: cfg.RedisPass,
		DB:       0,
	})

	// Test the connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Printf("Connected to Redis at %s", cfg.RedisHost)
}

// GetRedisClient returns the Redis client instance
func GetRedisClient() *redis.Client {
	return redisClient
}
