package client

import (
	"context"
	"encoding/json"
	"fmt"
	"messaging_service/internal/config"
	"messaging_service/internal/modules/message_broker/models"
	httpclient "messaging_service/pkg/http"
	"net/http"
)

type ConfigClient interface {
	FetchConfig(ctx context.Context) (*models.MessaggingConfigResponse, error)
}

type ConfigClientImpl struct {
	httpClient httpclient.HTTPClient
	cfg        *config.Env
}

func NewConfigClient(httpClient httpclient.HTTPClient, cfg *config.Env) ConfigClient {
	return &ConfigClientImpl{
		httpClient: httpClient,
		cfg:        cfg,
	}
}

// GetCurrentConfig returns the config for the given key.
func (c *ConfigClientImpl) FetchConfig(ctx context.Context) (*models.MessaggingConfigResponse, error) {
	fullURL := fmt.Sprintf(
		"%s/config/%s/%s",
		c.cfg.ConfigServiceUrl,
		c.cfg.Environment,
		c.cfg.ServiceName,
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

	var configResponse models.ConfigServiceResponse
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

	fmt.Println("[ConfigClient] Config data unmarshalled successfully")
	return &configResponse.Data, nil
}
