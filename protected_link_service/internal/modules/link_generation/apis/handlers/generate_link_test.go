package handlers

import (
	"context"
	"testing"

	"github.com/gocql/gocql"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	commonDtos "protected_link/internal/common/api/dtos"
	"protected_link/internal/common/constants"
	messageUtility "protected_link/internal/common/utils"
	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"
	"protected_link/internal/modules/link_generation/models"
	"protected_link/internal/modules/link_generation/repositories"
	"protected_link/internal/modules/link_generation/services"
	pb "protected_link/pkg/grpc/proto"
	database "protected_link/pkg/redis"
	"protected_link/pkg/validation"
)

func init() {
	validation.InitValidator()
}

// MockGeneratedRepository is a mock implementation of repositories.GeneratedRepository
type MockGeneratedRepository struct {
	repositories.GeneratedRepository
	mock.Mock
}

func (m *MockGeneratedRepository) SaveGeneratedLink(dto *apiDtos.GenerateUrlRequest) (*commonDtos.ApiResponseDto, error) {
	args := m.Called(dto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*commonDtos.ApiResponseDto), args.Error(1)
}

func (m *MockGeneratedRepository) GetOriginalToken(shortCode string) (string, error) {
	args := m.Called(shortCode)
	return args.String(0), args.Error(1)
}

func (m *MockGeneratedRepository) GetTokenData(link *string) (*apiDtos.SecurePayload, error) {
	args := m.Called(link)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*apiDtos.SecurePayload), args.Error(1)
}

func (m *MockGeneratedRepository) DeleteShortCode(shortCode string) (*commonDtos.ApiResponseDto, error) {
	args := m.Called(shortCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*commonDtos.ApiResponseDto), args.Error(1)
}

// MockCassandraRepository is a mock implementation of ICassandraRepository
type MockCassandraRepository struct {
	mock.Mock
}

func (m *MockCassandraRepository) SaveData(dto *apiDtos.GenerateUrlRequest) (gocql.UUID, error) {
	args := m.Called(dto)
	if args.Get(0) == nil {
		return gocql.UUID{}, args.Error(1)
	}
	return args.Get(0).(gocql.UUID), args.Error(1)
}

func (m *MockCassandraRepository) GetDataByID(id string) (*apiDtos.GenerateUrlRequest, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*apiDtos.GenerateUrlRequest), args.Error(1)
}

func (m *MockCassandraRepository) DeleteById(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockRedisConfig implements database.RedisConfig interface
type MockRedisConfig struct {
	mock.Mock
	client *redis.Client
	ctx    context.Context
}

func NewMockRedisConfig() *database.RedisConfig {
	mockRedis := &MockRedisConfig{
		client: redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			DB:       0,
			Password: "", // No password for testing
		}),
		ctx: context.Background(),
	}
	return &database.RedisConfig{
		Client: mockRedis.client,
		Ctx:    mockRedis.ctx,
	}
}

func (m *MockRedisConfig) Set(ctx context.Context, key string, value interface{}, expiration int) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockRedisConfig) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockRedisConfig) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func NewMockGenerateLinkService() *services.GenerateLinkService {
	repo := &MockGeneratedRepository{}
	mockRedis := &MockRedisConfig{
		client: redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			DB:       0,
			Password: "", // No password for testing
		}),
		ctx: context.Background(),
	}
	redis := &database.RedisConfig{
		Client: mockRedis.client,
		Ctx:    mockRedis.ctx,
	}
	cassandra := &MockCassandraRepository{}

	// Set up mock responses
	repo.On("SaveGeneratedLink", mock.Anything).Return(&commonDtos.ApiResponseDto{
		Success: true,
		Message: messageUtility.GetMessage(string(constants.LinkGeneratedSuccessfully)),
		Data: &models.ProtectedLinkResponse{
			URL: "http://example.com/test-link",
		},
	}, nil)

	repo.On("DeleteShortCode", mock.Anything).Return(&commonDtos.ApiResponseDto{
		Success: true,
		Message: messageUtility.GetMessage(string(constants.ProtectedLinkDeletedSuccessfully)),
		Data: &models.ProtectedLinkResponse{
			URL: "test-link",
		},
	}, nil)

	repo.On("GetTokenData", mock.Anything).Return(&apiDtos.SecurePayload{
		Data: map[string]interface{}{
			"key": "value",
		},
		ModelType: "jwt",
	}, nil)

	repo.On("GetOriginalToken", mock.Anything).Return("test-token", nil)

	// Set up Redis mock responses
	mockRedis.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockRedis.On("Get", mock.Anything, mock.Anything).Return("test-token", nil)
	mockRedis.On("Delete", mock.Anything, mock.Anything).Return(nil)

	// Create a new repository instance
	generatedRepo, err := repositories.NewGeneratedRepository(redis, cassandra)
	if err != nil {
		panic(err)
	}

	return services.NewGenerateLinkService(generatedRepo, redis, cassandra)
}

func TestNewGenerateLinkHandler(t *testing.T) {
	mockService := NewMockGenerateLinkService()
	handler := NewGenerateLinkHandler(mockService)

	assert.NotNil(t, handler)
	//assert.Equal(t, mockService, handler.services)
}

func TestSaveGeneratedLinkV1(t *testing.T) {
	tests := []struct {
		name          string
		req           *pb.GenerateUrlRequestV1
		mockResponse  *commonDtos.ApiResponseDto
		mockError     error
		expectedError error
		expectedResp  *pb.GenerateUrlResponseV1
	}{
		{
			name: "successful save",
			req: &pb.GenerateUrlRequestV1{
				UserId:      "test-user",
				Name:        "Test User",
				RequestType: "test",
				ModelType:   "jwt",
				Email:       "test@example.com",
				ExpireIn:    "1h",
				Phone:       "1234567890",
				ChannelType: "email",
				Data:        map[string]string{"key": "value"},
			},
			mockResponse: &commonDtos.ApiResponseDto{
				Success: true,
				Message: messageUtility.GetMessage(string(constants.LinkGeneratedSuccessfully)),
				Data: &models.ProtectedLinkResponse{
					URL: "http://example.com/test-link",
				},
			},
			mockError:     nil,
			expectedError: nil,
			expectedResp: &pb.GenerateUrlResponseV1{
				Success: true,
				Message: messageUtility.GetMessage(string(constants.LinkGeneratedSuccessfully)),
				Data: &pb.ProtectedLinkResponse{
					Url: "http://example.com/test-link",
				},
			},
		},
		{
			name:          "nil request",
			req:           nil,
			mockResponse:  nil,
			mockError:     nil,
			expectedError: assert.AnError,
			expectedResp:  nil,
		},
		{
			name: "validation error",
			req: &pb.GenerateUrlRequestV1{
				// Missing required fields
				UserId: "test-user",
				// Empty ModelType will cause validation error
			},
			mockResponse:  nil,
			mockError:     nil,
			expectedError: nil,
			expectedResp: &pb.GenerateUrlResponseV1{
				Success: false,
				Message: "Validation failed",
				Error:   map[string]string{"validation": "ModelType cannot be empty"},
			},
		},
		{
			name: "service error",
			req: &pb.GenerateUrlRequestV1{
				UserId:      "test-user",
				Name:        "Test User",
				RequestType: "test",
				ModelType:   "jwt",
				Email:       "test@example.com",
				ExpireIn:    "1h",
				Phone:       "1234567890",
				ChannelType: "email",
				Data:        map[string]string{"key": "value"},
			},
			mockResponse:  nil,
			mockError:     assert.AnError,
			expectedError: assert.AnError,
			expectedResp:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockGenerateLinkService{}
			handler := NewGenerateLinkHandler(mockService)

			if tt.req != nil && tt.name != "validation error" {
				mockService.On("SaveGeneratedLink", mock.Anything).Return(tt.mockResponse, tt.mockError)
			}

			resp, err := handler.SaveGeneratedLinkV1(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if tt.name != "validation error" {
					assert.Nil(t, resp)
				}
			} else {
				assert.NoError(t, err)
				if tt.name == "validation error" {
					assert.False(t, resp.Success)
					assert.Equal(t, "Validation failed", resp.Message)
					assert.NotEmpty(t, resp.Error)
				} else {
					assert.Equal(t, tt.expectedResp.Success, resp.Success)
					assert.Equal(t, tt.expectedResp.Message, resp.Message)
					if tt.expectedResp.Data != nil {
						assert.Equal(t, tt.expectedResp.Data.Url, resp.Data.Url)
					}
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestGetExtractDataV1(t *testing.T) {
	tests := []struct {
		name          string
		req           *pb.GetExtractDataRequestV1
		mockResponse  *commonDtos.ApiResponseDto
		mockError     error
		expectedError error
		expectedResp  *pb.GetExtractDataResponseV1
	}{
		{
			name: "successful get data",
			req: &pb.GetExtractDataRequestV1{
				Token: "test-token",
			},
			mockResponse: &commonDtos.ApiResponseDto{
				Success: true,
				Message: messageUtility.GetMessage(string(constants.DataFetchedSuccessfully)),
				Data: map[string]interface{}{
					"key": "value",
				},
			},
			mockError:     nil,
			expectedError: nil,
			expectedResp: &pb.GetExtractDataResponseV1{
				Success: true,
				Message: messageUtility.GetMessage(string(constants.DataFetchedSuccessfully)),
				Data: map[string]string{
					"key": "value",
				},
			},
		},
		{
			name:          "nil request",
			req:           nil,
			mockResponse:  nil,
			mockError:     nil,
			expectedError: nil,
			expectedResp: &pb.GetExtractDataResponseV1{
				Success: false,
				Message: "Validation failed",
				Data:    nil,
				Error:   `{"validation_errors":"Token cannot be empty"}`,
			},
		},
		{
			name: "empty token",
			req: &pb.GetExtractDataRequestV1{
				Token: "",
			},
			mockResponse:  nil,
			mockError:     nil,
			expectedError: nil,
			expectedResp: &pb.GetExtractDataResponseV1{
				Success: false,
				Message: "Validation failed",
				Data:    nil,
				Error:   `{"validation_errors":"Token cannot be empty"}`,
			},
		},
		{
			name: "service error",
			req: &pb.GetExtractDataRequestV1{
				Token: "test-token",
			},
			mockResponse:  nil,
			mockError:     assert.AnError,
			expectedError: assert.AnError,
			expectedResp:  nil,
		},
		{
			name: "non-map data type",
			req: &pb.GetExtractDataRequestV1{
				Token: "test-token",
			},
			mockResponse: &commonDtos.ApiResponseDto{
				Success: true,
				Message: "Success",
				Data:    "string data", // Non-map data
			},
			mockError:     nil,
			expectedError: nil,
			expectedResp: &pb.GetExtractDataResponseV1{
				Success: true,
				Message: messageUtility.GetMessage(string(constants.DataFetchedSuccessfully)),
				Data:    map[string]string{"message": "Success"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockGenerateLinkService{}
			handler := NewGenerateLinkHandler(mockService)

			if tt.req != nil && tt.req.Token != "" {
				mockService.On("GetExtractData", &tt.req.Token).Return(tt.mockResponse, tt.mockError)
			}

			resp, err := handler.GetExtractDataV1(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				if tt.name == "nil request" || tt.name == "empty token" {
					assert.False(t, resp.Success)
					assert.Equal(t, "Validation failed", resp.Message)
					assert.NotEmpty(t, resp.Error)
				} else {
					assert.Equal(t, tt.expectedResp.Success, resp.Success)
					assert.Equal(t, tt.expectedResp.Message, resp.Message)
					if tt.name == "non-map data type" {
						assert.Contains(t, resp.Data, "message")
					} else if tt.expectedResp.Data != nil {
						assert.Equal(t, tt.expectedResp.Data["key"], resp.Data["key"])
					}
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestDeleteGeneratedLinkV1(t *testing.T) {
	tests := []struct {
		name          string
		req           *pb.DeleteGeneratedLinkRequestV1
		mockResponse  *commonDtos.ApiResponseDto
		mockError     error
		expectedError error
		expectedResp  *pb.DeleteGeneratedLinkResponseV1
	}{
		{
			name: "successful delete",
			req: &pb.DeleteGeneratedLinkRequestV1{
				Token: "test-link",
			},
			mockResponse: &commonDtos.ApiResponseDto{
				Success: true,
				Message: messageUtility.GetMessage(string(constants.ProtectedLinkDeletedSuccessfully)),
				Data: &models.ProtectedLinkResponse{
					URL: "test-link",
				},
			},
			mockError:     nil,
			expectedError: nil,
			expectedResp: &pb.DeleteGeneratedLinkResponseV1{
				Success: true,
				Message: messageUtility.GetMessage(string(constants.ProtectedLinkDeletedSuccessfully)),
				Data: &pb.ProtectedLinkResponse{
					Url: "test-link",
				},
			},
		},
		{
			name:          "nil request",
			req:           nil,
			mockResponse:  nil,
			mockError:     nil,
			expectedError: nil,
			expectedResp: &pb.DeleteGeneratedLinkResponseV1{
				Success: false,
				Message: "Validation failed",
				Data:    nil,
				Error:   `{"validation_errors":"Link cannot be empty"}`,
			},
		},
		{
			name: "empty link",
			req: &pb.DeleteGeneratedLinkRequestV1{
				Token: "",
			},
			mockResponse:  nil,
			mockError:     nil,
			expectedError: nil,
			expectedResp: &pb.DeleteGeneratedLinkResponseV1{
				Success: false,
				Message: "Validation failed",
				Data:    nil,
				Error:   `{"validation_errors":"Link cannot be empty"}`,
			},
		},
		{
			name: "service error",
			req: &pb.DeleteGeneratedLinkRequestV1{
				Token: "test-link",
			},
			mockResponse:  nil,
			mockError:     assert.AnError,
			expectedError: assert.AnError,
			expectedResp:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockGenerateLinkService{}
			handler := NewGenerateLinkHandler(mockService)

			if tt.req != nil && tt.req.Token != "" {
				mockService.On("DeleteGeneratedLink", tt.req.Token).Return(tt.mockResponse, tt.mockError)
			}

			resp, err := handler.DeleteGeneratedLinkV1(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				if tt.name == "nil request" || tt.name == "empty link" {
					assert.False(t, resp.Success)
					assert.Equal(t, "Validation failed", resp.Message)
					assert.NotEmpty(t, resp.Error)
				} else {
					assert.Equal(t, tt.expectedResp.Success, resp.Success)
					assert.Equal(t, tt.expectedResp.Message, resp.Message)
					if tt.expectedResp.Data != nil {
						assert.Equal(t, tt.expectedResp.Data.Url, resp.Data.Url)
					}
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

type mockGenerateLinkService struct {
	mock.Mock
}

func (m *mockGenerateLinkService) SaveGeneratedLink(req *apiDtos.GenerateUrlRequest) (*commonDtos.ApiResponseDto, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*commonDtos.ApiResponseDto), args.Error(1)
}

func (m *mockGenerateLinkService) DeleteGeneratedLink(link string) (*commonDtos.ApiResponseDto, error) {
	args := m.Called(link)
	if args.Get(0) != nil {
		return args.Get(0).(*commonDtos.ApiResponseDto), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockGenerateLinkService) GetExtractData(token *string) (*commonDtos.ApiResponseDto, error) {
	args := m.Called(token)
	if args.Get(0) != nil {
		return args.Get(0).(*commonDtos.ApiResponseDto), args.Error(1)
	}
	return nil, args.Error(1)
}
