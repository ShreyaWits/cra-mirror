package temporal

import (
	"context"
	"log"
	"notification-service/internal/common/repositories"
	"notification-service/pkg/kafka"

	"go.temporal.io/sdk/client"
)

type TemporalClient struct {
	TemporalUrl string
	TaskQueue   string
	ID          string
	client      client.Client
	workflows   *TemporalWorkflow
}

type NotificationTemporalPayload struct {
	Recipient        Recipient         `json:"recipient"`
	Channel          string            `json:"channel"`
	Service          string            `json:"service"`
	Meta             map[string]string `json:"meta"`
	Tags             []string          `json:"tags,omitempty"`
	TemplateID       string            `json:"templateID"`
	RetryCount       int               `json:"retryCount"`
	ExecutionTimeout int               `json:"executionTimeout"`
}

type Recipient struct {
	UserID         string            `json:"userID,omitempty"`
	Email          string            `json:"email,omitempty" validate:"omitempty,email"`
	Phone          string            `json:"phone,omitempty" validate:"omitempty,phone"`
	WhatsappNumber string            `json:"whatsapp_number,omitempty"`
	Data           map[string]string `json:"data" validate:"required,min=1"`
}

func InitTemporal(TemporalUrl string, ID string, notificationRepo repositories.NotificationRepositoryInterface, kafkaClient *kafka.KafkaPublisher) (*TemporalClient, error) {

	log.Println("TemporalUrl", TemporalUrl)

	// Create a Temporal client
	c, err := NewTemporalClient(TemporalUrl)

	// default value
	TIMEOUT := 1000
	RETRY_COUNT := 3

	temporalWorkflows := NewTemporalWorkflow(&TIMEOUT, &RETRY_COUNT, notificationRepo, kafkaClient)

	if err != nil {
		log.Println("Err in temporal init", err)
		return nil, err
	}

	return &TemporalClient{
		TemporalUrl: TemporalUrl,
		TaskQueue:   "notification-service-queue",
		ID:          ID,
		client:      c,
		workflows:   temporalWorkflows,
	}, nil
}

func (tc *TemporalClient) ExecuteEmailWorkflow(payload map[string]interface{}) map[string]interface{} {
	log.Println("Executing the workflow via Temporal", payload)
	ctx := context.Background()
	// Prepare parameters for ExecuteRuleWorkflow
	options := client.StartWorkflowOptions{
		ID:        tc.ID,
		TaskQueue: tc.TaskQueue,
	}

	// Start ExecuteJourneyWorkflow with a map payload
	run, err := tc.client.ExecuteWorkflow(ctx, options, tc.workflows.ExecuteEmailWorkflow, payload)
	if err != nil {
		return nil
	}

	// Retrieve workflow result
	var responseData map[string]interface{}
	err = run.Get(context.Background(), &responseData)
	if err != nil {
		return nil
	}

	return responseData
}
func (tc *TemporalClient) ExecuteSmsWorkflow(payload map[string]interface{}) map[string]interface{} {
	log.Println("Executing the workflow via Temporal for sms", payload)
	ctx := context.Background()
	// Prepare parameters for ExecuteRuleWorkflow
	options := client.StartWorkflowOptions{
		ID:        tc.ID,
		TaskQueue: tc.TaskQueue,
	}

	// Start ExecuteJourneyWorkflow with a map payload
	run, err := tc.client.ExecuteWorkflow(ctx, options, tc.workflows.ExecuteSMSWorkflow, payload)
	if err != nil {
		return nil
	}

	// Retrieve workflow result
	var responseData map[string]interface{}
	err = run.Get(context.Background(), &responseData)
	if err != nil {
		return nil
	}

	return responseData
}
func (tc *TemporalClient) ExecuteWhatsappWorkflow(payload map[string]interface{}) map[string]interface{} {
	log.Println("Executing the workflow via Temporal whatsap", payload)
	ctx := context.Background()
	// Prepare parameters for ExecuteRuleWorkflow
	options := client.StartWorkflowOptions{
		ID:        tc.ID,
		TaskQueue: tc.TaskQueue,
	}

	// Start ExecuteJourneyWorkflow with a map payload
	run, err := tc.client.ExecuteWorkflow(ctx, options, tc.workflows.ExecuteWhatsappWorkflow, payload)
	if err != nil {
		return nil
	}

	// Retrieve workflow result
	var responseData map[string]interface{}
	err = run.Get(context.Background(), &responseData)
	if err != nil {
		return nil
	}

	return responseData
}
func (tc *TemporalClient) ExecutePushNotificationWorkflow(payload map[string]interface{}) map[string]interface{} {
	log.Println("Executing the workflow via Temporal", payload)
	ctx := context.Background()
	// Prepare parameters for ExecuteRuleWorkflow
	options := client.StartWorkflowOptions{
		ID:        tc.ID,
		TaskQueue: tc.TaskQueue,
	}

	// Start ExecuteJourneyWorkflow with a map payload
	run, err := tc.client.ExecuteWorkflow(ctx, options, tc.workflows.ExecutePushNotificationWorkflow, payload)
	if err != nil {
		return nil
	}

	// Retrieve workflow result
	var responseData map[string]interface{}
	err = run.Get(context.Background(), &responseData)
	if err != nil {
		return nil
	}

	return responseData
}
