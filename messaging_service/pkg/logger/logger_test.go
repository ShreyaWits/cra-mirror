package logger_test

import (
	"context"
	"messaging_service/pkg/logger"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockLogger(t *testing.T) {
	// Create a new mock logger
	mockLogger := logger.NewMockLogger()
	ctx := context.Background()

	// Test Info method
	mockLogger.Info(ctx, "test info message")
	assert.Len(t, mockLogger.InfoMessages, 1)
	assert.Equal(t, "test info message", mockLogger.InfoMessages[0].Args[0])
	assert.Equal(t, ctx, mockLogger.InfoMessages[0].Ctx)

	// Test Error method
	mockLogger.Error(ctx, "test error message")
	assert.Len(t, mockLogger.ErrorMessages, 1)
	assert.Equal(t, "test error message", mockLogger.ErrorMessages[0].Args[0])
	assert.Equal(t, ctx, mockLogger.ErrorMessages[0].Ctx)

	// Test Debug method
	mockLogger.Debug(ctx, "test debug message")
	assert.Len(t, mockLogger.DebugMessages, 1)
	assert.Equal(t, "test debug message", mockLogger.DebugMessages[0].Args[0])
	assert.Equal(t, ctx, mockLogger.DebugMessages[0].Ctx)

	// Test Warn method
	mockLogger.Warn(ctx, "test warn message")
	assert.Len(t, mockLogger.WarnMessages, 1)
	assert.Equal(t, "test warn message", mockLogger.WarnMessages[0].Args[0])
	assert.Equal(t, ctx, mockLogger.WarnMessages[0].Ctx)

	// Test multiple arguments
	mockLogger.Info(ctx, "test", "multiple", "args")
	assert.Len(t, mockLogger.InfoMessages, 2)
	assert.Equal(t, "test", mockLogger.InfoMessages[1].Args[0])
	assert.Equal(t, "multiple", mockLogger.InfoMessages[1].Args[1])
	assert.Equal(t, "args", mockLogger.InfoMessages[1].Args[2])
}

func TestMockLogger_WithFields(t *testing.T) {
	// Create a new mock logger
	mockLogger := logger.NewMockLogger()

	// Add fields
	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	// Get a new logger with fields
	loggerWithFields := mockLogger.WithFields(fields)

	// Check that it's a MockLogger
	mockLoggerWithFields, ok := loggerWithFields.(*logger.MockLogger)
	assert.True(t, ok)

	// Check that fields were added correctly
	assert.Equal(t, "value1", mockLoggerWithFields.Fields["key1"])
	assert.Equal(t, 123, mockLoggerWithFields.Fields["key2"])

	// Add more fields
	additionalFields := map[string]interface{}{
		"key3": true,
		"key1": "updated", // Override existing field
	}

	loggerWithMoreFields := mockLoggerWithFields.WithFields(additionalFields)
	mockLoggerWithMoreFields, ok := loggerWithMoreFields.(*logger.MockLogger)
	assert.True(t, ok)

	// Check fields were added and updated
	assert.Equal(t, "updated", mockLoggerWithMoreFields.Fields["key1"])
	assert.Equal(t, 123, mockLoggerWithMoreFields.Fields["key2"])
	assert.Equal(t, true, mockLoggerWithMoreFields.Fields["key3"])

	// Original logger fields should not be changed
	assert.Equal(t, "value1", mockLoggerWithFields.Fields["key1"])
	assert.NotContains(t, mockLoggerWithFields.Fields, "key3")
}

func TestMockLogger_Sync(t *testing.T) {
	// Create a new mock logger
	mockLogger := logger.NewMockLogger()

	// Check initial state
	assert.False(t, mockLogger.SyncCalled)

	// Call Sync
	err := mockLogger.Sync()

	// Verify Sync was called and returned no error
	assert.NoError(t, err)
	assert.True(t, mockLogger.SyncCalled)
}
