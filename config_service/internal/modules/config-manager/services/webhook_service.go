package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"nps-config-service/internal/modules/config-manager/repositories"

	customErr "nps-config-service/internal/common/errors"
	"strings"
)

type WebhookService struct {
	Repo repositories.IConfigRepo
}

type IWebhookService interface {
	RegisterWebhookService(req dtos.RegisterWebhookRequest) (interface{}, error)
	GetWebhooks(env, service string) ([]dtos.RegisterWebhookRequest, error)
	DeleteWebhook(env, service, url, method string) (string, error)
	DeleteAllWebhooks(env, service string) error
	NotifyWebhook(hook dtos.RegisterWebhookRequest, data map[string]interface{})
}

func NewWebhookService(repo repositories.IConfigRepo) IWebhookService {
	return &WebhookService{Repo: repo}
}

func (s *WebhookService) RegisterWebhookService(req dtos.RegisterWebhookRequest) (interface{}, error) {
	key := fmt.Sprintf("/webhooks/%s/%s", req.Environment, req.ServiceName)
	ctx := context.Background()
	var hooks []dtos.RegisterWebhookRequest

	webHook, err := s.Repo.Get(ctx, key)
	if err == nil && len(webHook) > 0 {
		if err := json.Unmarshal([]byte(webHook), &hooks); err != nil {
			return nil, fmt.Errorf("invalid webhook data stored for %s: %v", key, err)
		}
	}

	for _, existing := range hooks {
		if existing.URL == req.URL && existing.Method == req.Method {
			return nil, customErr.NewConflictError(fmt.Sprintf("webhook with URL '%s' and method '%s' already exists", req.URL, req.Method))
		}
	}

	hooks = append(hooks, req)
	data, err := json.Marshal(hooks)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal webhook: %v", err)
	}
	if err := s.Repo.Set(ctx, key, string(data), -1); err != nil {
		return nil, err
	}

	res := fmt.Sprintf("Webhook registered successfully for service: %s in environment: %s", req.ServiceName, req.Environment)
	return res, nil
}
func (s *WebhookService) GetWebhooks(env, service string) ([]dtos.RegisterWebhookRequest, error) {
	key := fmt.Sprintf("/webhooks/%s/%s", env, service)
	ctx := context.Background()

	webHook, err := s.Repo.Get(ctx, key)
	if err != nil || len(webHook) == 0 {
		return nil, fmt.Errorf("no webhooks found for %s/%s", env, service)
	}

	var hooks []dtos.RegisterWebhookRequest
	if err := json.Unmarshal([]byte(webHook), &hooks); err != nil {
		return nil, fmt.Errorf("failed to parse webhook data: %v", err)
	}
	return hooks, nil
}
func (s *WebhookService) DeleteWebhook(env, service, url, method string) (string, error) {
	key := fmt.Sprintf("/webhooks/%s/%s", env, service)
	ctx := context.Background()

	webHook, err := s.Repo.Get(ctx, key)
	if err != nil {
		return "", fmt.Errorf("webhook not found for %s/%s", env, service)
	}

	var hooks []dtos.RegisterWebhookRequest
	if err := json.Unmarshal([]byte(webHook), &hooks); err != nil {
		return "", fmt.Errorf("failed to parse existing webhooks: %v", err)
	}

	updated := make([]dtos.RegisterWebhookRequest, 0)
	found := false
	for _, h := range hooks {
		if h.URL == url && h.Method == method {
			found = true
			continue // skip the one to delete
		}
		updated = append(updated, h)
	}
	if !found {
		return "", fmt.Errorf("webhook with URL '%s' and method '%s' not found", url, method)
	}

	// Re-save the updated list
	data, err := json.Marshal(updated)
	if err != nil {
		return "", fmt.Errorf("failed to marshal updated webhooks: %v", err)
	}
	if err := s.Repo.Set(ctx, key, string(data), -1); err != nil {
		return "", err
	}
	return "Webhook deleted successfully", nil
}
func (s *WebhookService) DeleteAllWebhooks(env, service string) error {
	key := fmt.Sprintf("/webhooks/%s/%s", env, service)
	ctx := context.Background()
	return s.Repo.Delete(ctx, key)
}
func (s *WebhookService) NotifyWebhook(hook dtos.RegisterWebhookRequest, data map[string]interface{}) {
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
