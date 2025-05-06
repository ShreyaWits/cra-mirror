package services

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gocql/gocql"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	repository "protected_link/internal/modules/cassandra/repository"
	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"
	"protected_link/internal/modules/link_generation/repositories"
	database "protected_link/pkg/redis"
)

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

func setupTest(t *testing.T) (*database.RedisConfig, repository.ICassandraRepository, *repositories.GeneratedRepository) {
	// Create a miniredis instance
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to create miniredis instance: %v", err)
	}

	// Create Redis client connected to miniredis
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Create Redis config
	redisConfig := &database.RedisConfig{
		Client: client,
		Ctx:    context.Background(),
	}

	// Create mock Cassandra repository
	cassandra := &MockCassandraRepository{}

	// Create repository with mocked dependencies
	repo, err := repositories.NewGeneratedRepository(redisConfig, cassandra)
	assert.NoError(t, err)

	// Cleanup function
	t.Cleanup(func() {
		client.Close()
		mr.Close()
	})

	return redisConfig, cassandra, repo
}

func TestNewGenerateLinkService(t *testing.T) {
	redis, cassandra, repo := setupTest(t)

	service := NewGenerateLinkService(repo, redis, cassandra)

	assert.NotNil(t, service)
	assert.Equal(t, repo, service.repo)
	assert.Equal(t, redis, service.redis)
	assert.Equal(t, cassandra, service.cassandra)
}

func TestSaveGeneratedLink(t *testing.T) {
	tests := []struct {
		name          string
		dto           *apiDtos.GenerateUrlRequest
		setupMocks    func(*MockCassandraRepository)
		expectedError error
	}{
		{
			name: "successful save",
			dto: &apiDtos.GenerateUrlRequest{
				UserID:      "test-user",
				Name:        "Test User",
				RequestType: "test",
				ModelType:   "standard",
				Email:       "test@example.com",
				ExpireIn:    "1h",
				Phone:       "1234567890",
				ChannelType: "email",
				Data:        apiDtos.JSONB{"key": "value"},
			},
			setupMocks: func(mockCassandra *MockCassandraRepository) {
				// No Cassandra interaction needed for standard model
			},
			expectedError: nil,
		},
		{
			name: "validation error",
			dto: &apiDtos.GenerateUrlRequest{
				UserID: "test-user",
			},
			setupMocks: func(mockCassandra *MockCassandraRepository) {
				// No mock setup needed for validation error case
			},
			expectedError: errors.New("validation error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redis, cassandra, repo := setupTest(t)
			mockCassandra := cassandra.(*MockCassandraRepository)

			if tt.setupMocks != nil {
				tt.setupMocks(mockCassandra)
			}

			service := NewGenerateLinkService(repo, redis, cassandra)
			response, err := service.SaveGeneratedLink(tt.dto)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.True(t, response.Success)
			}
		})
	}
}

func TestDeleteGeneratedLink(t *testing.T) {
	tests := []struct {
		name          string
		link          string
		setupMocks    func(*MockCassandraRepository)
		expectedError error
	}{
		{
			name: "successful delete standard model",
			link: "test-link",
			setupMocks: func(mockCassandra *MockCassandraRepository) {
				// Assuming no Cassandra interaction needed for standard model link
			},
			expectedError: nil,
		},
		{
			name: "link not found",
			link: "invalid-link",
			setupMocks: func(mockCassandra *MockCassandraRepository) {
				// Simulate Cassandra having no record of this link
				// You may want to set expectations on methods like:
				// mockCassandra.On("DeleteLink", "invalid-link").Return(errors.New("link not found"))
			},
			expectedError: errors.New("link not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redis, cassandra, repo := setupTest(t)
			mockCassandra, ok := cassandra.(*MockCassandraRepository)
			require.True(t, ok, "cassandra should be of type *MockCassandraRepository")

			if tt.setupMocks != nil {
				tt.setupMocks(mockCassandra)
			}

			service := NewGenerateLinkService(repo, redis, cassandra)
			response, err := service.DeleteGeneratedLink(tt.link)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Nil(t, response)
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.True(t, response.Success)
			}

			mockCassandra.AssertExpectations(t)
		})
	}
}

func TestGetExtractData(t *testing.T) {
	tests := []struct {
		name          string
		link          string
		setupMocks    func(*MockCassandraRepository)
		expectedError error
	}{
		{
			name: "successful get data standard model",
			link: "test-link",
			setupMocks: func(mockCassandra *MockCassandraRepository) {
				// No Cassandra interaction needed for standard model
			},
			expectedError: nil,
		},
		{
			name: "link not found",
			link: "invalid-link",
			setupMocks: func(mockCassandra *MockCassandraRepository) {
				// Simulate link not found scenario if needed
				// e.g., mockCassandra.On("GetLinkData", "invalid-link").Return(nil, errors.New("link not found"))
			},
			expectedError: errors.New("link not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redis, cassandra, repo := setupTest(t)

			mockCassandra, ok := cassandra.(*MockCassandraRepository)
			require.True(t, ok, "cassandra should be of type *MockCassandraRepository")

			if tt.setupMocks != nil {
				tt.setupMocks(mockCassandra)
			}

			service := NewGenerateLinkService(repo, redis, cassandra)

			response, err := service.GetExtractData(&tt.link)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Nil(t, response)
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.True(t, response.Success)
			}

			mockCassandra.AssertExpectations(t)
		})
	}
}
