package handlers_test

import (
	"bytes"
	"context"
	"errors"

	// "log"
	"notification-service/internal/common/api/dtos"
	"notification-service/internal/common/api/handlers"
	config "notification-service/internal/configs"
	"notification-service/pkg/logger"
	"notification-service/proto"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) SaveConfig(ctx context.Context, configs []dtos.ChannelConfig) error {
	args := m.Called(ctx, configs)
	return args.Error(0)
}

func (m *MockNotificationService) GetConfigure() (*proto.GetConfigureResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*proto.GetConfigureResponse), args.Error(1)
}

func setupTest() {
	LOKI_URL := config.GetEnv("LOKI_URL", "http://localhost:5000")
	// Initialize logger for tests
	logger.InitLogger(LOKI_URL)
}

func TestConfigure_Success(t *testing.T) {
	setupTest()
	mockSvc := new(MockNotificationService)
	server := handlers.NewGRPCServer(mockSvc)

	configs := []*proto.ChannelConfig{
		{
			Service:  "email",
			Primary:  "smtp",
			Fallback: "sendgrid",
		},
	}

	req := &proto.ConfigureRequest{
		Configs: configs,
	}

	mockSvc.On("SaveConfig", mock.Anything, mock.AnythingOfType("[]dtos.ChannelConfig")).Return(nil)

	resp, err := server.Configure(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, int32(200), resp.StatusCode)
	assert.Equal(t, "Configuration updated", resp.Message)
	assert.Equal(t, "Channel configuration saved successfully", resp.Data)
	mockSvc.AssertExpectations(t)
}

func TestConfigure_ValidationError(t *testing.T) {
	setupTest()
	mockSvc := new(MockNotificationService)
	server := handlers.NewGRPCServer(mockSvc)

	configs := []*proto.ChannelConfig{
		{
			Service:  "", // Missing required field
			Primary:  "smtp",
			Fallback: "sendgrid",
		},
	}

	req := &proto.ConfigureRequest{
		Configs: configs,
	}

	resp, err := server.Configure(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, int32(400), resp.StatusCode)
	assert.Equal(t, "Validation failed", resp.Message)
	assert.NotNil(t, resp.FieldError)
}

func TestConfigure_ServiceError(t *testing.T) {
	setupTest()
	mockSvc := new(MockNotificationService)
	server := handlers.NewGRPCServer(mockSvc)

	configs := []*proto.ChannelConfig{
		{
			Service:  "email",
			Primary:  "smtp",
			Fallback: "sendgrid",
		},
	}

	req := &proto.ConfigureRequest{
		Configs: configs,
	}

	mockSvc.On("SaveConfig", mock.Anything, mock.AnythingOfType("[]dtos.ChannelConfig")).Return(errors.New("database error"))

	resp, err := server.Configure(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, int32(500), resp.StatusCode)
	assert.Equal(t, "Failed to save configuration", resp.Message)
	assert.Equal(t, "SRV001", resp.FieldError.ErrorCode)
	mockSvc.AssertExpectations(t)
}

func TestGetConfigure_Success(t *testing.T) {
	setupTest()
	mockSvc := new(MockNotificationService)
	server := handlers.NewGRPCServer(mockSvc)

	expectedResp := &proto.GetConfigureResponse{
		Status:     "success",
		StatusCode: 200,
		Message:    "Configurations fetched successfully",
		Configs: []*proto.ChannelConfig{
			{
				Service:  "email",
				Primary:  "smtp",
				Fallback: "sendgrid",
			},
		},
	}

	mockSvc.On("GetConfigure").Return(expectedResp, nil)

	resp, err := server.GetConfigure(context.Background(), &proto.ConfigureRequest{})

	assert.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, int32(200), resp.StatusCode)
	assert.Equal(t, "Configurations fetched successfully", resp.Message)
	assert.Len(t, resp.Configs, 1)
	assert.Equal(t, "email", resp.Configs[0].Service)
	mockSvc.AssertExpectations(t)
}

func TestGetConfigure_ServiceError(t *testing.T) {
	setupTest()
	mockSvc := new(MockNotificationService)
	server := handlers.NewGRPCServer(mockSvc)

	// Create a buffer to capture logrus output
	var buf bytes.Buffer
	// Store the original logger output and formatter
	originalOutput := logger.Log.Out
	originalFormatter := logger.Log.Formatter
	// Set logger output to our buffer and use JSON formatter
	logger.Log.SetOutput(&buf)
	logger.Log.SetFormatter(&logrus.JSONFormatter{})
	// Restore original output and formatter after test
	defer func() {
		logger.Log.SetOutput(originalOutput)
		logger.Log.SetFormatter(originalFormatter)
	}()

	mockSvc.On("GetConfigure").Return(nil, errors.New("database error"))

	resp, err := server.GetConfigure(context.Background(), &proto.ConfigureRequest{})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, buf.String(), "Error in GetConfigure service call")
	mockSvc.AssertExpectations(t)
}

// func TestNewGRPCServer_NilService(t *testing.T) {
// 	setupTest()

// 	// Create a custom logger that doesn't terminate
// 	oldLogger := log.Default()
// 	defer func() { log.SetOutput(oldLogger.Writer()) }()

// 	// Create a buffer to capture log output
// 	var buf bytes.Buffer
// 	log.SetOutput(&buf)

// 	// Test that creating a server with nil service panics
// 	assert.Panics(t, func() {
// 		handlers.NewGRPCServer(nil)
// 	}, "Expected panic when creating server with nil service")

// 	// Verify the log message
// 	assert.Contains(t, buf.String(), "NotificationService cannot be nil")
// }
