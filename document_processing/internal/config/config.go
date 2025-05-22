//go:build !test
// build +test
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
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
	api := fmt.Sprintf("%s/%s/document-processing", strings.TrimRight(os.Getenv("CONFIG_API"), "/"), os.Getenv("ENVIRONMENT"))
	log.Println(api)
	req, err := http.NewRequest("GET", api, nil)
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
	if err != nil {
		return nil, err
	}

	var configResp ConfigResponse
	if err := json.Unmarshal(body, &configResp); err != nil {
		return nil, err
	}

	return configResp.Data, nil
}

// NewConfig validates and constructs Config from raw API data
func NewConfig(data map[string]string) (*Config, error) {
	requiredKeys := []string{
		"MINIO_ENDPOINT",
		"MINIO_ACCESS_KEY",
		"MINIO_SECRET_KEY",
		"MINIO_BUCKET_NAME",
		"GEMINI_API_KEY",
		"SERVER_PORT",
		"YUGABYTE_HOST",
		"YUGABYTE_USER",
		"YUGABYTE_PASSWORD",
		"YUGABYTE_NAME",
		"YUGABYTE_PORT",
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

	return &Config{
		ServerPort:       port,
		MinioEndpoint:    data["MINIO_ENDPOINT"],
		MinioAccessKey:   data["MINIO_ACCESS_KEY"],
		MinioSecretKey:   data["MINIO_SECRET_KEY"],
		MinioBucketName:  data["MINIO_BUCKET_NAME"],
		GeminiAPIKey:     data["GEMINI_API_KEY"],
		YugabyteHost:     data["YUGABYTE_HOST"],
		YugabyteUser:     data["YUGABYTE_USER"],
		YugabytePassWord: data["YUGABYTE_PASSWORD"],
		YugabyteName:     data["YUGABYTE_NAME"],
		YugabytePort:     data["YUGABYTE_PORT"],
	}, nil
}
