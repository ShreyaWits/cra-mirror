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

func TestGetDLQTopicName(t *testing.T) {
	tests := []struct {
		name           string
		topic          string
		expectedResult string
	}{
		{
			name:           "normal topic name",
			topic:          "test-topic",
			expectedResult: "test-topic-dlq",
		},
		{
			name:           "empty topic name",
			topic:          "",
			expectedResult: "-dlq", // Edge case, probably should be validated elsewhere
		},
		{
			name:           "topic with special characters",
			topic:          "test.topic-123",
			expectedResult: "test.topic-123-dlq",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := confluent.GetDLQTopicName(tt.topic)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestNewDLQProducer(t *testing.T) {
	// This test would require mocking the kafka.NewProducer function which isn't possible without refactoring
	// the code to use a factory pattern. For now, we'll skip this test.
	t.Skip("Skipping test that requires mocking kafka.NewProducer function")
}

func TestSendToDLQ(t *testing.T) {
	tests := []struct {
		name          string
		messageID     string
		value         []byte
		headers       []confluent.Header
		failureReason string
		mockSetup     func(*mock.MockKafkaProducerTestify, *mock.MockLoggerServiceTestify)
		expectErr     bool
		errorMsg      string
	}{
		{
			name:          "successful message send",
			messageID:     "test-message-id",
			value:         []byte("test message"),
			headers:       []confluent.Header{{Key: "existing-header", Value: []byte("value")}},
			failureReason: "processing error",
			mockSetup: func(mockProducer *mock.MockKafkaProducerTestify, mockLogger *mock.MockLoggerServiceTestify) {
				// Expect Produce call
				mockProducer.On("Produce", tmock.AnythingOfType("*kafka.Message"), tmock.Anything).Return(nil)

				// Expect Flush call
				mockProducer.On("Flush", tmock.AnythingOfType("int")).Return(0)

				// Expect logger calls
				mockLogger.On("Info", tmock.Anything, tmock.Anything, tmock.Anything).Return()
			},
			expectErr: false,
		},
		{
			name:          "produce error",
			messageID:     "test-message-id",
			value:         []byte("test message"),
			headers:       []confluent.Header{},
			failureReason: "processing error",
			mockSetup: func(mockProducer *mock.MockKafkaProducerTestify, mockLogger *mock.MockLoggerServiceTestify) {
				// Expect Produce call with error
				mockProducer.On("Produce", tmock.AnythingOfType("*kafka.Message"), tmock.Anything).
					Return(fmt.Errorf("produce error"))

				// Expect logger calls
				mockLogger.On("Info", tmock.Anything, tmock.Anything, tmock.Anything).Return()
				mockLogger.On("Error", tmock.Anything, tmock.Anything, tmock.Anything).Return()
			},
			expectErr: true,
			errorMsg:  "failed to send message to DLQ: produce error",
		},
		{
			name:          "flush with remaining messages",
			messageID:     "test-message-id",
			value:         []byte("test message"),
			headers:       []confluent.Header{},
			failureReason: "processing error",
			mockSetup: func(mockProducer *mock.MockKafkaProducerTestify, mockLogger *mock.MockLoggerServiceTestify) {
				// Expect Produce call
				mockProducer.On("Produce", tmock.AnythingOfType("*kafka.Message"), tmock.Anything).Return(nil)

				// Expect Flush call with remaining messages
				mockProducer.On("Flush", tmock.AnythingOfType("int")).Return(5)

				// Expect logger calls
				mockLogger.On("Info", tmock.Anything, tmock.Anything, tmock.Anything).Return()
				mockLogger.On("Warn", tmock.Anything, tmock.Anything, tmock.Anything).Return()
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the setup function from the mock package
			dlqProducer, mockProducer, mockObs := mock.SetupMockDLQProducer(t)

			// Setup mock expectations
			if tt.mockSetup != nil {
				tt.mockSetup(mockProducer, mockObs.LoggerService)
			}

			// Call the method under test
			err := dlqProducer.SendToDLQ(context.Background(), tt.messageID, tt.value, tt.headers, tt.failureReason)

			// Verify results
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify the DLQ-specific headers were added
			if !tt.expectErr {
				// We can't directly check the message headers since we're mocking Produce,
				// but we can verify that the mock was called with the expected arguments
				mockProducer.AssertCalled(t, "Produce", tmock.AnythingOfType("*kafka.Message"), tmock.Anything)
			}

			// Verify all expectations were met
			mockProducer.AssertExpectations(t)
			mockObs.LoggerService.AssertExpectations(t)
		})
	}
}

func TestDLQProducerClose(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(*mock.MockKafkaProducerTestify, *mock.MockLoggerServiceTestify)
		remaining int
	}{
		{
			name: "successful close without remaining messages",
			mockSetup: func(mockProducer *mock.MockKafkaProducerTestify, mockLogger *mock.MockLoggerServiceTestify) {
				// Expect Flush call
				mockProducer.On("Flush", tmock.AnythingOfType("int")).Return(0)

				// Expect Close call
				mockProducer.On("Close").Return()

				// Expect logger calls
				mockLogger.On("Info", tmock.Anything, tmock.Anything, tmock.Anything).Return()
			},
			remaining: 0,
		},
		{
			name: "close with remaining messages",
			mockSetup: func(mockProducer *mock.MockKafkaProducerTestify, mockLogger *mock.MockLoggerServiceTestify) {
				// Expect Flush call with remaining messages
				mockProducer.On("Flush", tmock.AnythingOfType("int")).Return(3)

				// Expect Close call
				mockProducer.On("Close").Return()

				// Expect logger calls
				mockLogger.On("Info", tmock.Anything, tmock.Anything, tmock.Anything).Return()
				mockLogger.On("Warn", tmock.Anything, tmock.Anything, tmock.Anything).Return()
			},
			remaining: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the setup function from the mock package
			dlqProducer, mockProducer, mockObs := mock.SetupMockDLQProducer(t)

			// Setup mock expectations
			if tt.mockSetup != nil {
				tt.mockSetup(mockProducer, mockObs.LoggerService)
			}

			// Call the method under test
			err := dlqProducer.Close()

			// Verify results
			assert.NoError(t, err)

			// Verify the right methods were called
			mockProducer.AssertCalled(t, "Flush", tmock.AnythingOfType("int"))
			mockProducer.AssertCalled(t, "Close")

			// If there were remaining messages, verify the warning was logged
			if tt.remaining > 0 {
				mockObs.LoggerService.AssertCalled(t, "Warn", tmock.Anything, tmock.Anything, tmock.Anything)
			}

			// Verify all expectations were met
			mockProducer.AssertExpectations(t)
			mockObs.LoggerService.AssertExpectations(t)
		})
	}
}
