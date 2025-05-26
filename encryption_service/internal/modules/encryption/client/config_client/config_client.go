package client

import (
	"context"
	"encoding/json"
	"encryption_microservice/internal/config"
	httpclient "encryption_microservice/pkg/http"
	"fmt"
	"net/http"
)

// ConfigServiceResponse represents the response from the config service
type ConfigServiceResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Data    config.Config `json:"data"`
}

type ConfigClient interface {
	FetchConfig(ctx context.Context) (*config.Config, error)
}

type ConfigClientImpl struct {
	httpClient httpclient.HTTPClient
	cfg        *config.EnvConfig
}

func NewConfigClient(httpClient httpclient.HTTPClient, cfg *config.EnvConfig) (ConfigClient, error) {
	// Validate input arguments
	if httpClient == nil {
		return nil, fmt.Errorf("httpClient cannot be nil in NewConfigClient")
	}
	if cfg == nil {
		return nil, fmt.Errorf("cfg cannot be nil in NewConfigClient")
	}

	// Validate required configuration fields
	if cfg.ConfigServiceUrl == "" {
		return nil, fmt.Errorf("config service URL is required in environment configuration")
	}
	if cfg.ConfigServiceToken == "" {
		return nil, fmt.Errorf("config service token is required in environment configuration")
	}
	if cfg.Environment == "" {
		return nil, fmt.Errorf("environment is required in environment configuration")
	}
	if config.SERVICE_NAME == "" {
		return nil, fmt.Errorf("service name is required in environment configuration")
	}

	return &ConfigClientImpl{
		httpClient: httpClient,
		cfg:        cfg,
	}, nil
}

// FetchConfig returns the config for the encryption service
func (c *ConfigClientImpl) FetchConfig(ctx context.Context) (*config.Config, error) {
	fullURL := fmt.Sprintf(
		"%s/config/%s/%s",
		c.cfg.ConfigServiceUrl,
		c.cfg.Environment,
		config.SERVICE_NAME,
	)

	fmt.Println("[ConfigClient] Fetching config from URL:", fullURL)

	req := httpclient.Request{
		Method: httpclient.GET,
		URL:    fullURL,
		Headers: map[string]string{
			"Authorization": "Bearer " + c.cfg.ConfigServiceToken,
		},
	}

	fmt.Println("[ConfigClient] Sending request with headers:", req.Headers)

	var configResponse ConfigServiceResponse
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		fmt.Println("[ConfigClient] Request failed:", err)
		return nil, fmt.Errorf("httpclient.Do failed: %w", err)
	}
	defer resp.Body.Close()

	fmt.Println("[ConfigClient] Received response with status code:", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("config fetch failed: status=%d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&configResponse); err != nil {
		return nil, fmt.Errorf("failed to decode config response: %w", err)
	}

	// Basic validation of the response
	if configResponse.Data.VaultAddr == "" {
		return nil, fmt.Errorf("invalid config received: VaultAddr is required but missing")
	}

	fmt.Println("[ConfigClient] Config data unmarshalled successfully")
	return &configResponse.Data, nil
}
