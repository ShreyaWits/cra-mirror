package workflows

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"nps-config-service/internal/modules/config-manager/apis/dtos"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type WebhookInput struct {
	Hook dtos.RegisterWebhookRequest
	Data map[string]interface{}
}

func WebhookWorkflow(ctx workflow.Context, input WebhookInput) error {
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second * 5,
		BackoffCoefficient: 2.0,
		MaximumAttempts:    5,
	}

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 10,
		RetryPolicy:         retryPolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	return workflow.ExecuteActivity(ctx, SendWebhookActivity, input).Get(ctx, nil)
}

func SendWebhookActivity(ctx context.Context, input WebhookInput) error {
	body, err := json.Marshal(map[string]interface{}{
		"values":      input.Data,
		"method":      input.Hook.Method,
		"environment": input.Hook.Environment,
		"serviceName": input.Hook.ServiceName,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	fullURL := fmt.Sprintf("%s/%s/%s", strings.TrimRight(input.Hook.URL, "/"), input.Hook.Environment, input.Hook.ServiceName)
	if !strings.HasPrefix(fullURL, "http://") && !strings.HasPrefix(fullURL, "https://") {
		fullURL = "http://" + fullURL
	}

	resp, err := http.Post(fullURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("non-2xx status: %s", resp.Status)
	}

	log.Printf("Webhook to %s succeeded", fullURL)
	return nil
}
