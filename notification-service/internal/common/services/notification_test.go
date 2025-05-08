package services

import (
	"context"
	"notification-service/internal/common/api/dtos"
	"notification-service/proto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockConfigRepository is a mock implementation of ConfigRepositoryInterface
type MockConfigRepository struct {
	mock.Mock
}

func (m *MockConfigRepository) SaveConfig(ctx context.Context, configs []dtos.ChannelConfig) error {
	args := m.Called(ctx, configs)
	return args.Error(0)
}

func (m *MockConfigRepository) GetConfig() ([]dtos.ChannelConfig, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dtos.ChannelConfig), args.Error(1)
}

func (m *MockConfigRepository) GetConfigByChannel(channel string) (*dtos.ChannelConfig, error) {
	args := m.Called(channel)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.ChannelConfig), args.Error(1)
}

func TestSaveConfig(t *testing.T) {
	tests := []struct {
		name          string
		configs       []dtos.ChannelConfig
		mockError     error
		expectedError error
	}{
		{
			name: "successful save",
			configs: []dtos.ChannelConfig{
				{
					Service:  "email",
					Primary:  "sendgrid",
					Fallback: "smtp",
				},
			},
			mockError:     nil,
			expectedError: nil,
		},
		{
			name: "repository error",
			configs: []dtos.ChannelConfig{
				{
					Service:  "email",
					Primary:  "sendgrid",
					Fallback: "smtp",
				},
			},
			mockError:     assert.AnError,
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockConfigRepository)
			service := &NotificationService{
				configRepo: mockRepo,
			}

			ctx := context.Background()
			mockRepo.On("SaveConfig", ctx, tt.configs).Return(tt.mockError)

			err := service.SaveConfig(ctx, tt.configs)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetConfigure(t *testing.T) {
	tests := []struct {
		name           string
		mockConfigs    []dtos.ChannelConfig
		mockError      error
		expectedResp   *proto.GetConfigureResponse
		expectedError  error
		expectedStatus string
	}{
		{
			name: "successful get config",
			mockConfigs: []dtos.ChannelConfig{
				{
					Service:  "email",
					Primary:  "sendgrid",
					Fallback: "smtp",
				},
			},
			mockError: nil,
			expectedResp: &proto.GetConfigureResponse{
				Status:     "success",
				StatusCode: 200,
				Message:    "Configurations fetched successfully",
				Configs: []*proto.ChannelConfig{
					{
						Service:  "email",
						Primary:  "sendgrid",
						Fallback: "smtp",
					},
				},
			},
			expectedError:  nil,
			expectedStatus: "success",
		},
		{
			name:        "repository error",
			mockConfigs: nil,
			mockError:   assert.AnError,
			expectedResp: &proto.GetConfigureResponse{
				Status:     "error",
				StatusCode: 500,
				Message:    "Error fetching configurations",
			},
			expectedError:  assert.AnError,
			expectedStatus: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockConfigRepository)
			service := &NotificationService{
				configRepo: mockRepo,
			}

			mockRepo.On("GetConfig").Return(tt.mockConfigs, tt.mockError)

			resp, err := service.GetConfigure()

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectedResp.Status, resp.Status)
				assert.Equal(t, tt.expectedResp.StatusCode, resp.StatusCode)
				assert.Equal(t, tt.expectedResp.Message, resp.Message)
				assert.Equal(t, len(tt.expectedResp.Configs), len(resp.Configs))

				if len(tt.expectedResp.Configs) > 0 {
					assert.Equal(t, tt.expectedResp.Configs[0].Service, resp.Configs[0].Service)
					assert.Equal(t, tt.expectedResp.Configs[0].Primary, resp.Configs[0].Primary)
					assert.Equal(t, tt.expectedResp.Configs[0].Fallback, resp.Configs[0].Fallback)
				}
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
