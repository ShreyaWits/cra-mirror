package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"log"
)

type Config struct {
	ServerPort      int
	MinioEndpoint   string
	MinioAccessKey  string
	MinioSecretKey  string
	MinioBucketName string
	GeminiAPIKey    string
}

type ConfigResponse struct {
	StatusCode int               `json:"status_code"`
	Message    string            `json:"message"`
	Data       map[string]string `json:"data"`
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
		ServerPort:      port,
		MinioEndpoint:   data["MINIO_ENDPOINT"],
		MinioAccessKey:  data["MINIO_ACCESS_KEY"],
		MinioSecretKey:  data["MINIO_SECRET_KEY"],
		MinioBucketName: data["MINIO_BUCKET_NAME"],
		GeminiAPIKey:    data["GEMINI_API_KEY"],
	}, nil
}
