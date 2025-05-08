package service

import (
	"context"
	"encoding/json"
	"fmt"
	"notification-service/internal/common/api/dtos"
	"notification-service/internal/common/models"
	"notification-service/pkg/logger"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

// Mock logger
type mockLogger struct{ mock.Mock }

func (m *mockLogger) Write(p []byte) (n int, err error) {
	m.Called(p)
	return len(p), nil
}

func (m *mockLogger) Println(args ...interface{}) {
	m.Called(args)
}

func (m *mockLogger) Printf(format string, args ...interface{}) {
	m.Called(format, args)
}

// Mock implementations
type mockRedis struct {
	mock.Mock
}

func (m *mockRedis) GetNotificationConfig(key string) ([]dtos.ChannelConfig, error) {
	args := m.Called(key)
	if configs, ok := args.Get(0).([]dtos.ChannelConfig); ok {
		return configs, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRedis) SetNotificationConfig(key string, config []dtos.ChannelConfig) error {
	args := m.Called(key, config)
	return args.Error(0)
}

type mockConfigRepo struct {
	mock.Mock
}

func (m *mockConfigRepo) GetConfig() ([]dtos.ChannelConfig, error) {
	args := m.Called()
	if configs, ok := args.Get(0).([]dtos.ChannelConfig); ok {
		return configs, args.Error(1)
	}
	return nil, args.Error(1)
}

type mockNotificationRepo struct {
	mock.Mock
}

func (m *mockNotificationRepo) SaveNotification(n *models.Notification) (*string, error) {
	args := m.Called(n)
	if id, ok := args.Get(0).(*string); ok {
		return id, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockNotificationRepo) UpdateNotificationByID(id string, n *models.Notification) error {
	args := m.Called(id, n)
	return args.Error(0)
}

type mockTemporalClient struct {
	mock.Mock
}

func (m *mockTemporalClient) ExecuteEmailWorkflow(data map[string]interface{}) map[string]interface{} {
	args := m.Called(data)
	if result, ok := args.Get(0).(map[string]interface{}); ok {
		return result
	}
	return nil
}

func (m *mockTemporalClient) ExecuteSmsWorkflow(data map[string]interface{}) map[string]interface{} {
	args := m.Called(data)
	if result, ok := args.Get(0).(map[string]interface{}); ok {
		return result
	}
	return nil
}

func (m *mockTemporalClient) ExecutePushNotificationWorkflow(data map[string]interface{}) map[string]interface{} {
	args := m.Called(data)
	if result, ok := args.Get(0).(map[string]interface{}); ok {
		return result
	}
	return nil
}

func (m *mockTemporalClient) ExecuteWhatsappWorkflow(data map[string]interface{}) map[string]interface{} {
	args := m.Called(data)
	if result, ok := args.Get(0).(map[string]interface{}); ok {
		return result
	}
	return nil
}

type mockKafkaClient struct {
	mock.Mock
}

func (m *mockKafkaClient) Publish(ctx context.Context, topic string, key, value []byte) error {
	args := m.Called(ctx, topic, key, value)
	return args.Error(0)
}

// Test helper functions
func setupTestLogger() *mockLogger {
	logger.Log = logrus.New()
	mockLogger := new(mockLogger)
	mockLogger.On("Write", mock.Anything).Return(0, nil).Maybe()
	mockLogger.On("Println", mock.Anything).Return().Maybe()
	mockLogger.On("Printf", mock.Anything, mock.Anything).Return().Maybe()
	logger.Log.SetOutput(mockLogger)
	return mockLogger
}

func setupTestMocks() (*mockRedis, *mockConfigRepo, *mockTemporalClient, *mockNotificationRepo, *mockKafkaClient) {
	return &mockRedis{}, &mockConfigRepo{}, &mockTemporalClient{}, &mockNotificationRepo{}, &mockKafkaClient{}
}

func createTestHandler(redis *mockRedis, configRepo *mockConfigRepo, temporalClient *mockTemporalClient,
	notificationRepo *mockNotificationRepo, kafkaClient *mockKafkaClient) *QueueHandlers {
	return NewQueueHandlers(redis, configRepo, temporalClient, notificationRepo, kafkaClient)
}

func TestHandleSendNotification_ValidPayload(t *testing.T) {
	defaultConfig := []dtos.ChannelConfig{
		{Service: "email", Primary: "sendgrid", Fallback: "smtp"},
		{Service: "sms", Primary: "twilio", Fallback: "fast2sms"},
		{Service: "whatsapp", Primary: "twilio", Fallback: "fast2sms"},
	}

	tests := []struct {
		name          string
		channel       string
		recipient     dtos.Recipient
		config        []dtos.ChannelConfig
		expectedError bool
		expectedCalls func(*mockRedis, *mockConfigRepo, *mockTemporalClient, *mockNotificationRepo, *mockKafkaClient)
	}{
		{
			name:    "Valid Email Channel",
			channel: "email",
			recipient: dtos.Recipient{
				Email:  "test@example.com",
				UserID: "u1",
				Data: map[string]string{
					"username": "testuser",
					"otp":      "123456",
				},
			},
			config: []dtos.ChannelConfig{
				{Service: "email", Primary: "sendgrid", Fallback: "smtp"},
			},
			expectedError: false,
			expectedCalls: func(redis *mockRedis, configRepo *mockConfigRepo, temporalClient *mockTemporalClient,
				notificationRepo *mockNotificationRepo, kafkaClient *mockKafkaClient) {
				redis.On("GetNotificationConfig", "notification-config").Return([]dtos.ChannelConfig{}, nil)
				configRepo.On("GetConfig").Return(defaultConfig, nil)
				redis.On("SetNotificationConfig", "notification-config", defaultConfig).Return(nil)
				notificationID := "notif-id"
				notificationRepo.On("SaveNotification", mock.MatchedBy(func(n *models.Notification) bool {
					return n.Channel == "email" && n.PrimaryService == "sendgrid"
				})).Return(&notificationID, nil)
				temporalClient.On("ExecuteEmailWorkflow", mock.MatchedBy(func(data map[string]interface{}) bool {
					return data["id"] == "notif-id"
				})).Return(map[string]interface{}{})
			},
		},
		{
			name:    "Valid SMS Channel",
			channel: "sms",
			recipient: dtos.Recipient{
				Phone:  "+911234567890",
				UserID: "u1",
				Data: map[string]string{
					"username": "testuser",
					"otp":      "123456",
				},
			},
			config: []dtos.ChannelConfig{
				{Service: "sms", Primary: "twilio", Fallback: "fast2sms"},
			},
			expectedError: false,
			expectedCalls: func(redis *mockRedis, configRepo *mockConfigRepo, temporalClient *mockTemporalClient,
				notificationRepo *mockNotificationRepo, kafkaClient *mockKafkaClient) {
				redis.On("GetNotificationConfig", "notification-config").Return([]dtos.ChannelConfig{}, nil).Run(func(args mock.Arguments) {
					t.Log("GetNotificationConfig called with:", args[0])
				})
				configRepo.On("GetConfig").Return(defaultConfig, nil).Run(func(args mock.Arguments) {
					t.Log("GetConfig called")
				})
				redis.On("SetNotificationConfig", "notification-config", defaultConfig).Return(nil).Run(func(args mock.Arguments) {
					t.Log("SetNotificationConfig called with:", args[0], args[1])
				})
				notificationID := "notif-id"
				notificationRepo.On("SaveNotification", mock.MatchedBy(func(n *models.Notification) bool {
					match := n.Channel == "sms" &&
						n.PrimaryService == "twilio" &&
						n.FallbackService == "fast2sms" &&
						n.Recipient["phone"] == "+911234567890"
					t.Log("SaveNotification called with:", n, "Match:", match)
					return match
				})).Return(&notificationID, nil)
				temporalClient.On("ExecuteSmsWorkflow", mock.MatchedBy(func(data map[string]interface{}) bool {
					match := data["id"] == "notif-id"
					t.Log("ExecuteSmsWorkflow called with:", data, "Match:", match)
					return match
				})).Return(map[string]interface{}{})
			},
		},
		{
			name:    "Valid WhatsApp Channel",
			channel: "whatsapp",
			recipient: dtos.Recipient{
				WhatsappNumber: "+911234567890",
				UserID:         "u1",
				Data: map[string]string{
					"username": "testuser",
					"otp":      "123456",
				},
			},
			config: []dtos.ChannelConfig{
				{Service: "whatsapp", Primary: "twilio", Fallback: "fast2sms"},
			},
			expectedError: false,
			expectedCalls: func(redis *mockRedis, configRepo *mockConfigRepo, temporalClient *mockTemporalClient,
				notificationRepo *mockNotificationRepo, kafkaClient *mockKafkaClient) {
				redis.On("GetNotificationConfig", "notification-config").Return([]dtos.ChannelConfig{}, nil)
				configRepo.On("GetConfig").Return(defaultConfig, nil)
				redis.On("SetNotificationConfig", "notification-config", defaultConfig).Return(nil)
				notificationID := "notif-id"
				notificationRepo.On("SaveNotification", mock.MatchedBy(func(n *models.Notification) bool {
					return n.Channel == "whatsapp" &&
						n.PrimaryService == "twilio" &&
						n.FallbackService == "fast2sms" &&
						n.Recipient["whatsapp_number"] == "+911234567890"
				})).Return(&notificationID, nil)
				temporalClient.On("ExecuteWhatsappWorkflow", mock.MatchedBy(func(data map[string]interface{}) bool {
					return data["id"] == "notif-id"
				})).Return(map[string]interface{}{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			setupTestLogger()
			redis, configRepo, temporalClient, notificationRepo, kafkaClient := setupTestMocks()
			h := createTestHandler(redis, configRepo, temporalClient, notificationRepo, kafkaClient)

			// Setup expectations
			tt.expectedCalls(redis, configRepo, temporalClient, notificationRepo, kafkaClient)

			// Create test payload
			notificationJSON := dtos.NotificationRequest{
				TrackingID: "track123",
				TemplateID: "otp_verification",
				Recipients: []dtos.Recipient{tt.recipient},
				Channels:   []string{tt.channel},
				Meta:       map[string]string{"foo": "bar"},
				Tags:       []string{"t1"},
			}

			b, _ := json.Marshal(notificationJSON)

			// Execute
			h.HandleSendNotification(kafka.Message{Value: b})

			// Verify
			redis.AssertExpectations(t)
			configRepo.AssertExpectations(t)
			temporalClient.AssertExpectations(t)
			notificationRepo.AssertExpectations(t)
			kafkaClient.AssertExpectations(t)
		})
	}
}

func TestHandleSendNotification_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name          string
		message       kafka.Message
		expectedError bool
		expectedCalls func(*mockRedis, *mockConfigRepo, *mockTemporalClient, *mockNotificationRepo, *mockKafkaClient)
	}{
		{
			name:          "Invalid JSON",
			message:       kafka.Message{Value: []byte("invalid json")},
			expectedError: true,
			expectedCalls: func(redis *mockRedis, configRepo *mockConfigRepo, temporalClient *mockTemporalClient,
				notificationRepo *mockNotificationRepo, kafkaClient *mockKafkaClient) {
				// No expectations as the error should be logged and return early
			},
		},
		{
			name: "Invalid Payload",
			message: kafka.Message{Value: func() []byte {
				invalidPayload := dtos.NotificationRequest{
					TrackingID: "track123",
					// Missing TemplateID
					Recipients: []dtos.Recipient{{
						Email:  "test@example.com",
						UserID: "u1",
					}},
					Channels: []string{"email"},
				}
				b, _ := json.Marshal(invalidPayload)
				return b
			}()},
			expectedError: true,
			expectedCalls: func(redis *mockRedis, configRepo *mockConfigRepo, temporalClient *mockTemporalClient,
				notificationRepo *mockNotificationRepo, kafkaClient *mockKafkaClient) {
				// No expectations as validation should fail
			},
		},
		{
			name: "Redis Error",
			message: kafka.Message{Value: func() []byte {
				payload := dtos.NotificationRequest{
					TrackingID: "track123",
					TemplateID: "otp_verification",
					Recipients: []dtos.Recipient{{
						Email:  "test@example.com",
						UserID: "u1",
						Data: map[string]string{
							"username": "testuser",
							"otp":      "123456",
						},
					}},
					Channels: []string{"email"},
					Meta:     map[string]string{"foo": "bar"},
					Tags:     []string{"t1"},
				}
				b, _ := json.Marshal(payload)
				return b
			}()},
			expectedError: true,
			expectedCalls: func(redis *mockRedis, configRepo *mockConfigRepo, temporalClient *mockTemporalClient,
				notificationRepo *mockNotificationRepo, kafkaClient *mockKafkaClient) {
				redis.On("GetNotificationConfig", "notification-config").Return([]dtos.ChannelConfig{}, fmt.Errorf("redis error"))
				configRepo.On("GetConfig").Return([]dtos.ChannelConfig{
					{Service: "email", Primary: "sendgrid", Fallback: "smtp"},
				}, nil)
				redis.On("SetNotificationConfig", "notification-config", mock.Anything).Return(nil)
				notificationID := "notif-id"
				notificationRepo.On("SaveNotification", mock.Anything).Return(&notificationID, nil)
				temporalClient.On("ExecuteEmailWorkflow", mock.Anything).Return(map[string]interface{}{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			setupTestLogger()
			redis, configRepo, temporalClient, notificationRepo, kafkaClient := setupTestMocks()
			h := createTestHandler(redis, configRepo, temporalClient, notificationRepo, kafkaClient)

			// Setup expectations
			tt.expectedCalls(redis, configRepo, temporalClient, notificationRepo, kafkaClient)

			// Execute
			h.HandleSendNotification(tt.message)

			// Verify
			redis.AssertExpectations(t)
			configRepo.AssertExpectations(t)
			temporalClient.AssertExpectations(t)
			notificationRepo.AssertExpectations(t)
			kafkaClient.AssertExpectations(t)
		})
	}
}

func TestChannelConfigSliceToMap(t *testing.T) {
	configs := []dtos.ChannelConfig{
		{Service: "email", Primary: "sendgrid", Fallback: "smtp"},
		{Service: "sms", Primary: "twilio", Fallback: "fast2sms"},
	}
	result := ChannelConfigSliceToMap(configs)
	if len(result) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result))
	}
	if result["email"].Primary != "sendgrid" || result["email"].Fallback != "smtp" {
		t.Errorf("unexpected mapping for email: %+v", result["email"])
	}
	if result["sms"].Primary != "twilio" || result["sms"].Fallback != "fast2sms" {
		t.Errorf("unexpected mapping for sms: %+v", result["sms"])
	}

	// Test with empty input
	emptyResult := ChannelConfigSliceToMap([]dtos.ChannelConfig{})
	if len(emptyResult) != 0 {
		t.Errorf("expected 0 entries for empty input, got %d", len(emptyResult))
	}
}
