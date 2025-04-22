package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"nps-config-service/internal/modules/config-manager/apis/dtos"

	"strings"
)

func (s *ConfigService) RegisterWebhookService(req dtos.RegisterWebhookRequest) (interface{}, error) {
	key := fmt.Sprintf("/webhooks/%s/%s", req.Environment, req.ServiceName)
	ctx := context.Background()
	var hooks []dtos.RegisterWebhookRequest
	webHook, err := s.Repo.Get(ctx, key)
	if err == nil && len(webHook) > 0 {
		// Unmarshal existing ones if found
		if err := json.Unmarshal([]byte(webHook), &hooks); err != nil {
			fmt.Printf("Failed to parse webhook data for %s: %v\n", key, err)
			return nil, fmt.Errorf("invalid webhook data stored for %s", key)
		}
	}
	isDuplicate := false
	for _, existing := range hooks {
		if existing.URL == req.URL && existing.Method == req.Method {
			return nil, fmt.Errorf("webhook with URL '%s' and method '%s' already exists", req.URL, req.Method)
		}
	}
	if !isDuplicate {
		hooks = append(hooks, req)
		data, err := json.Marshal(hooks)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal webhook: %v", err)
		}
		err = s.Repo.Set(ctx, key, string(data), -1)
		if err != nil {
			return nil, err
		}
	}
	res := fmt.Sprintf("Webhook registered successfully for service: %s in environment: %s", req.ServiceName, req.Environment)
	return res, nil

}

func (s *ConfigService) NotifyWebhook(hook dtos.RegisterWebhookRequest, data map[string]interface{}) {
	body, _ := json.Marshal(map[string]interface{}{
		"values":      data,
		"method":      hook.Method,
		"environment": hook.Environment,
		"serviceName": hook.ServiceName,
	})

	// fullURL := strings.TrimRight(hook.URL, "/")
	fullURL := fmt.Sprintf("%s/%s/%s", strings.TrimRight(hook.URL, "/"), hook.Environment, hook.ServiceName)
	if !strings.HasPrefix(fullURL, "http://") && !strings.HasPrefix(fullURL, "https://") {
		fullURL = "http://" + fullURL
	}

	resp, err := http.Post(fullURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Failed to notify %s: %v", fullURL, err)
		return
	}
	defer resp.Body.Close()

	log.Printf("Webhook to %s responded with status: %s", fullURL, resp.Status)
}
