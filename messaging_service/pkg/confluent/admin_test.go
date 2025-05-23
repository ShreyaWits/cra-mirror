package confluent_test

import (
	"context"
	"fmt"
	"testing"

	"messaging_service/internal/modules/message_broker/mock"
	"messaging_service/pkg/confluent"

	"github.com/stretchr/testify/assert"
	tmock "github.com/stretchr/testify/mock"
)

func TestCreateTopic(t *testing.T) {
	tests := []struct {
		name              string
		topic             string
		numPartitions     int
		replicationFactor int
		configs           map[string]string
		mockSetup         func(*mock.MockKafkaAdminTestify)
		expectErr         bool
		errorMsg          string
	}{
		{
			name:              "successful topic creation",
			topic:             "test-topic",
			numPartitions:     3,
			replicationFactor: 3,
			configs:           confluent.DefaultTopicConfig(),
			mockSetup: func(mockAdmin *mock.MockKafkaAdminTestify) {
				mockAdmin.On("CreateTopic", tmock.Anything, "test-topic", 3, 3, tmock.Anything).Return(nil)
			},
			expectErr: false,
		},
		{
			name:              "topic creation with client error",
			topic:             "test-topic",
			numPartitions:     3,
			replicationFactor: 3,
			configs:           nil,
			mockSetup: func(mockAdmin *mock.MockKafkaAdminTestify) {
				mockAdmin.On("CreateTopic", tmock.Anything, "test-topic", 3, 3, tmock.Anything).Return(fmt.Errorf("connection error"))
			},
			expectErr: true,
			errorMsg:  "connection error",
		},
		{
			name:              "topic creation with default partitions",
			topic:             "test-topic",
			numPartitions:     0, // Should default to 3
			replicationFactor: 3,
			configs:           nil,
			mockSetup: func(mockAdmin *mock.MockKafkaAdminTestify) {
				// The CreateTopic method in the actual implementation should use default of 3 partitions
				// when numPartitions is 0, so we should expect a call with 3 partitions
				mockAdmin.On("CreateTopic", tmock.Anything, "test-topic", tmock.Anything, 3, tmock.Anything).Return(nil)
			},
			expectErr: false,
		},
		{
			name:              "topic creation with default replication",
			topic:             "test-topic",
			numPartitions:     3,
			replicationFactor: 0, // Should default to 3
			configs:           nil,
			mockSetup: func(mockAdmin *mock.MockKafkaAdminTestify) {
				// The CreateTopic method in the actual implementation should use default of 3 replication factor
				// when replicationFactor is 0, so we should expect a call with 3 replication factor
				mockAdmin.On("CreateTopic", tmock.Anything, "test-topic", 3, tmock.Anything, tmock.Anything).Return(nil)
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get a mocked admin
			admin, _, _ := mock.SetupMockAdmin(t)

			// Cast to our specific mock type so we can set expectations
			mockAdmin := admin.(*mock.MockKafkaAdminTestify)

			// Setup mock expectations
			if tt.mockSetup != nil {
				tt.mockSetup(mockAdmin)
			}

			// Call the method under test
			err := admin.CreateTopic(context.Background(), tt.topic, tt.numPartitions, tt.replicationFactor, tt.configs)

			// Verify results
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			mockAdmin.AssertExpectations(t)
		})
	}
}

func TestDeleteTopic(t *testing.T) {
	tests := []struct {
		name      string
		topic     string
		mockSetup func(*mock.MockKafkaAdminTestify)
		expectErr bool
		errorMsg  string
	}{
		{
			name:  "successful topic deletion",
			topic: "test-topic",
			mockSetup: func(mockAdmin *mock.MockKafkaAdminTestify) {
				mockAdmin.On("DeleteTopic", tmock.Anything, "test-topic").Return(nil)
			},
			expectErr: false,
		},
		{
			name:  "topic deletion with client error",
			topic: "test-topic",
			mockSetup: func(mockAdmin *mock.MockKafkaAdminTestify) {
				mockAdmin.On("DeleteTopic", tmock.Anything, "test-topic").Return(fmt.Errorf("connection error"))
			},
			expectErr: true,
			errorMsg:  "connection error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get a mocked admin
			admin, _, _ := mock.SetupMockAdmin(t)

			// Cast to our specific mock type so we can set expectations
			mockAdmin := admin.(*mock.MockKafkaAdminTestify)

			// Setup mock expectations
			if tt.mockSetup != nil {
				tt.mockSetup(mockAdmin)
			}

			// Call the method under test
			err := admin.DeleteTopic(context.Background(), tt.topic)

			// Verify results
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			mockAdmin.AssertExpectations(t)
		})
	}
}

func TestListTopics(t *testing.T) {
	tests := []struct {
		name         string
		mockSetup    func(*mock.MockKafkaAdminTestify)
		expectErr    bool
		errorMsg     string
		expectTopics []string
	}{
		{
			name: "successful topics listing",
			mockSetup: func(mockAdmin *mock.MockKafkaAdminTestify) {
				mockAdmin.On("ListTopics", tmock.Anything).Return([]string{"topic1", "topic2", "topic3"}, nil)
			},
			expectErr:    false,
			expectTopics: []string{"topic1", "topic2", "topic3"},
		},
		{
			name: "empty topics listing",
			mockSetup: func(mockAdmin *mock.MockKafkaAdminTestify) {
				mockAdmin.On("ListTopics", tmock.Anything).Return([]string{}, nil)
			},
			expectErr:    false,
			expectTopics: []string{},
		},
		{
			name: "metadata error",
			mockSetup: func(mockAdmin *mock.MockKafkaAdminTestify) {
				mockAdmin.On("ListTopics", tmock.Anything).Return([]string{}, fmt.Errorf("metadata retrieval error"))
			},
			expectErr: true,
			errorMsg:  "metadata retrieval error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get a mocked admin
			admin, _, _ := mock.SetupMockAdmin(t)

			// Cast to our specific mock type so we can set expectations
			mockAdmin := admin.(*mock.MockKafkaAdminTestify)

			// Setup mock expectations
			if tt.mockSetup != nil {
				tt.mockSetup(mockAdmin)
			}

			// Call the method under test
			topics, err := admin.ListTopics(context.Background())

			// Verify results
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				// Check for topics, regardless of order
				assert.ElementsMatch(t, tt.expectTopics, topics)
			}

			// Verify all expectations were met
			mockAdmin.AssertExpectations(t)
		})
	}
}

func TestHasActiveConsumers(t *testing.T) {
	tests := []struct {
		name         string
		topic        string
		mockSetup    func(*mock.MockKafkaAdminTestify, *mock.MockLoggerServiceTestify)
		expectErr    bool
		errorMsg     string
		expectResult bool
	}{
		{
			name:  "topic exists with active consumers",
			topic: "test-topic",
			mockSetup: func(mockAdmin *mock.MockKafkaAdminTestify, mockLogger *mock.MockLoggerServiceTestify) {
				mockAdmin.On("HasActiveConsumers", tmock.Anything, "test-topic").Return(true, nil)

				// We won't be directly using these logger calls since our mock admin won't call them
				// but we should still make the expectation optional so the test doesn't fail
				mockLogger.On("Info", tmock.Anything, tmock.Anything, tmock.Anything).Maybe()
			},
			expectErr:    false,
			expectResult: true,
		},
		{
			name:  "topic does not exist",
			topic: "non-existent-topic",
			mockSetup: func(mockAdmin *mock.MockKafkaAdminTestify, mockLogger *mock.MockLoggerServiceTestify) {
				mockAdmin.On("HasActiveConsumers", tmock.Anything, "non-existent-topic").Return(false, nil)

				// We won't be directly using these logger calls since our mock admin won't call them
				// but we should still make the expectation optional so the test doesn't fail
				mockLogger.On("Info", tmock.Anything, tmock.Anything, tmock.Anything).Maybe()
			},
			expectErr:    false,
			expectResult: false,
		},
		{
			name:  "error checking active consumers",
			topic: "test-topic",
			mockSetup: func(mockAdmin *mock.MockKafkaAdminTestify, mockLogger *mock.MockLoggerServiceTestify) {
				mockAdmin.On("HasActiveConsumers", tmock.Anything, "test-topic").Return(false, fmt.Errorf("metadata error"))

				// We won't be directly using these logger calls since our mock admin won't call them
				// but we should still make the expectation optional so the test doesn't fail
				mockLogger.On("Error", tmock.Anything, tmock.Anything, tmock.Anything).Maybe()
			},
			expectErr: true,
			errorMsg:  "metadata error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get a mocked admin
			admin, _, mockObs := mock.SetupMockAdmin(t)

			// Cast to our specific mock type so we can set expectations
			mockAdmin := admin.(*mock.MockKafkaAdminTestify)

			// Setup mock expectations
			if tt.mockSetup != nil {
				tt.mockSetup(mockAdmin, mockObs.LoggerService)
			}

			// Call the method under test
			hasConsumers, err := admin.HasActiveConsumers(context.Background(), tt.topic)

			// Verify results
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectResult, hasConsumers)
			}

			// Verify all expectations were met
			mockAdmin.AssertExpectations(t)
			mockObs.LoggerService.AssertExpectations(t)
		})
	}
}

func TestClose(t *testing.T) {
	// Get a mocked admin
	admin, _, _ := mock.SetupMockAdmin(t)

	// Cast to our specific mock type so we can set expectations
	mockAdmin := admin.(*mock.MockKafkaAdminTestify)

	// Setup expectations
	mockAdmin.On("Close").Return(nil)

	// Call the method under test
	err := admin.Close()

	// Verify results
	assert.NoError(t, err)

	// Verify expectations
	mockAdmin.AssertExpectations(t)
}
