package repositories

import (
	"context"
	"notification-service/internal/common/api/dtos"
	"testing"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSession is a mock implementation of gocql.Session
type MockSession struct {
	mock.Mock
}

func (m *MockSession) Query(stmt string, values ...interface{}) QueryExecutor {
	args := m.Called(stmt, values)
	return args.Get(0).(QueryExecutor)
}

func (m *MockSession) ExecuteBatch(batch *gocql.Batch) error {
	args := m.Called(batch)
	return args.Error(0)
}

// MockConfigRepository implements ConfigRepositoryInterface for testing
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

func TestConfigRepository(t *testing.T) {
	t.Run("SaveConfig", func(t *testing.T) {
		mockRepo := new(MockConfigRepository)
		ctx := context.Background()
		configs := []dtos.ChannelConfig{
			{
				Service:  "email",
				Primary:  "sendgrid",
				Fallback: "smtp",
			},
		}

		mockRepo.On("SaveConfig", ctx, configs).Return(nil)
		err := mockRepo.SaveConfig(ctx, configs)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetConfig", func(t *testing.T) {
		mockRepo := new(MockConfigRepository)
		expectedConfigs := []dtos.ChannelConfig{
			{
				Service:  "email",
				Primary:  "sendgrid",
				Fallback: "smtp",
			},
		}

		mockRepo.On("GetConfig").Return(expectedConfigs, nil)
		configs, err := mockRepo.GetConfig()
		assert.NoError(t, err)
		assert.Equal(t, expectedConfigs, configs)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetConfigByChannel", func(t *testing.T) {
		mockRepo := new(MockConfigRepository)
		channel := "email"
		expectedConfig := &dtos.ChannelConfig{
			Service:  "email",
			Primary:  "sendgrid",
			Fallback: "smtp",
		}

		mockRepo.On("GetConfigByChannel", channel).Return(expectedConfig, nil)
		config, err := mockRepo.GetConfigByChannel(channel)
		assert.NoError(t, err)
		assert.Equal(t, expectedConfig, config)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetConfig Error", func(t *testing.T) {
		mockRepo := new(MockConfigRepository)
		mockRepo.On("GetConfig").Return(nil, assert.AnError)
		configs, err := mockRepo.GetConfig()
		assert.Error(t, err)
		assert.Nil(t, configs)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetConfigByChannel Error", func(t *testing.T) {
		mockRepo := new(MockConfigRepository)
		channel := "nonexistent"
		mockRepo.On("GetConfigByChannel", channel).Return(nil, assert.AnError)
		config, err := mockRepo.GetConfigByChannel(channel)
		assert.Error(t, err)
		assert.Nil(t, config)
		mockRepo.AssertExpectations(t)
	})

	t.Run("SaveConfig Error", func(t *testing.T) {
		mockRepo := new(MockConfigRepository)
		ctx := context.Background()
		configs := []dtos.ChannelConfig{}
		mockRepo.On("SaveConfig", ctx, configs).Return(assert.AnError)
		err := mockRepo.SaveConfig(ctx, configs)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestNewConfigRepository(t *testing.T) {
	t.Run("with nil session", func(t *testing.T) {
		repo, err := NewConfigRepository(nil)
		assert.Error(t, err)
		assert.Nil(t, repo)
		assert.Contains(t, err.Error(), "Cassandra session is nil")
	})
}

func TestSaveConfig(t *testing.T) {
	mockSession := new(MockSession)
	repo := &ConfigRepository{
		session: mockSession,
	}

	ctx := context.Background()
	testConfigs := []dtos.ChannelConfig{
		{
			Service:  "email",
			Primary:  "sendgrid",
			Fallback: "smtp",
		},
	}

	// Setup mock expectations
	mockSession.On("ExecuteBatch", mock.AnythingOfType("*gocql.Batch")).Return(nil)

	err := repo.SaveConfig(ctx, testConfigs)
	assert.NoError(t, err)
	mockSession.AssertExpectations(t)
}

// func TestGetConfig(t *testing.T) {
// 	mockSession := new(MockSession)
// 	repo := &ConfigRepository{
// 		session: mockSession,
// 	}

// 	expectedConfigs := []dtos.ChannelConfig{
// 		{
// 			Service:  "email",
// 			Primary:  "sendgrid",
// 			Fallback: "smtp",
// 		},
// 	}

// 	// Create a mock query result
// 	mockQuery := &gocql.Query{}
// 	mockIter := &gocql.Iter{}

// 	// Setup mock expectations
// 	mockSession.On("Query", "SELECT service, primary_provider, fallback_provider FROM channel_configs").Return(mockQuery)
// 	mockQuery.On("Iter").Return(mockIter)
// 	mockIter.On("Scan", mock.Anything, mock.Anything, mock.Anything).Return(true).Once()
// 	mockIter.On("Scan", mock.Anything, mock.Anything, mock.Anything).Return(false)

// 	configs, err := repo.GetConfig()
// 	assert.NoError(t, err)
// 	assert.Equal(t, expectedConfigs, configs)
// 	mockSession.AssertExpectations(t)
// }

// func TestGetConfigByChannel(t *testing.T) {
// 	mockSession := new(MockSession)
// 	repo := &ConfigRepository{
// 		session: mockSession,
// 	}

// 	channel := "email"
// 	expectedConfig := &dtos.ChannelConfig{
// 		Service:  "email",
// 		Primary:  "sendgrid",
// 		Fallback: "smtp",
// 	}

// 	// Create a mock query result
// 	mockQuery := &gocql.Query{}
// 	mockIter := &gocql.Iter{}

// 	// Setup mock expectations
// 	mockSession.On("Query", "SELECT service, primary_provider, fallback_provider FROM notifications").Return(mockQuery)
// 	mockQuery.On("Iter").Return(mockIter)
// 	mockIter.On("Scan", mock.Anything, mock.Anything, mock.Anything).Return(true).Once()
// 	mockIter.On("Scan", mock.Anything, mock.Anything, mock.Anything).Return(false)

// 	config, err := repo.GetConfigByChannel(channel)
// 	assert.NoError(t, err)
// 	assert.Equal(t, expectedConfig, config)
// 	mockSession.AssertExpectations(t)
// }

func TestConfigRepositoryNilChecks(t *testing.T) {
	repo := &ConfigRepository{
		session: nil,
	}

	t.Run("SaveConfig with nil session", func(t *testing.T) {
		err := repo.SaveConfig(context.Background(), nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository or session is nil")
	})

	t.Run("GetConfig with nil session", func(t *testing.T) {
		configs, err := repo.GetConfig()
		assert.Error(t, err)
		assert.Nil(t, configs)
		assert.Contains(t, err.Error(), "repository or session is nil")
	})

	t.Run("GetConfigByChannel with nil session", func(t *testing.T) {
		config, err := repo.GetConfigByChannel("email")
		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "repository or session is nil")
	})
}
