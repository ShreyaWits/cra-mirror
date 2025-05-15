package logger

import (
	"bytes"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestGetOutboundIP(t *testing.T) {
	InitLogger()
	getOutboundIP()

	assert.NotEmpty(t, srcIP, "Source IP should not be empty")
	ip := net.ParseIP(srcIP)
	assert.NotNil(t, ip, "Source IP should be a valid IP address")
}

func TestLogEvent(t *testing.T) {
	InitLogger()

	// Create a buffer to capture log output
	var buf bytes.Buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.InfoLevel,
	)
	Logger = zap.New(core)

	LogEvent("123", "test_event", "user1", "success", "This is a test log")
	logOutput := buf.String()

	assert.Contains(t, logOutput, "This is a test log", "Log should contain the message")
	assert.Contains(t, logOutput, "test_event", "Log should contain the event type")
	assert.Contains(t, logOutput, "user1", "Log should contain the user ID")
	assert.Contains(t, logOutput, "success", "Log should contain the status")
}

func TestLogErrorEvent(t *testing.T) {
	InitLogger()

	// Create a buffer to capture log output
	var buf bytes.Buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.ErrorLevel,
	)
	Logger = zap.New(core)

	LogErrorEvent("123", "error_event", "user1", "failed", "This is an error log")
	logOutput := buf.String()

	assert.Contains(t, logOutput, "This is an error log", "Log should contain the message")
	assert.Contains(t, logOutput, "error_event", "Log should contain the event type")
	assert.Contains(t, logOutput, "user1", "Log should contain the user ID")
	assert.Contains(t, logOutput, "failed", "Log should contain the status")
}

func TestLogWarnEvent(t *testing.T) {
	InitLogger()

	// Create a buffer to capture log output
	var buf bytes.Buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.WarnLevel,
	)
	Logger = zap.New(core)

	LogWarnEvent("123", "warn_event", "user1", "warning", "This is a warning log")
	logOutput := buf.String()

	assert.Contains(t, logOutput, "This is a warning log", "Log should contain the message")
	assert.Contains(t, logOutput, "warn_event", "Log should contain the event type")
	assert.Contains(t, logOutput, "user1", "Log should contain the user ID")
	assert.Contains(t, logOutput, "warning", "Log should contain the status")
}

func TestLogDebugEvent(t *testing.T) {
	InitLogger()

	// Create a buffer to capture log output
	var buf bytes.Buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.DebugLevel,
	)
	Logger = zap.New(core)

	LogDebugEvent("123", "debug_event", "user1", "debugging", "This is a debug log")
	logOutput := buf.String()

	assert.Contains(t, logOutput, "This is a debug log", "Log should contain the message")
	assert.Contains(t, logOutput, "debug_event", "Log should contain the event type")
	assert.Contains(t, logOutput, "user1", "Log should contain the user ID")
	assert.Contains(t, logOutput, "debugging", "Log should contain the status")
}

func TestError(t *testing.T) {
	InitLogger()

	// Create a buffer to capture log output
	var buf bytes.Buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.ErrorLevel,
	)
	Logger = zap.New(core)

	Error("Test error message", nil)
	logOutput := buf.String()

	assert.Contains(t, logOutput, "Test error message", "Log should contain the message")

	// Test with error
	Error("Test error with err", assert.AnError)
	logOutput = buf.String()
	assert.Contains(t, logOutput, "Test error with err", "Log should contain the message")
	assert.Contains(t, logOutput, assert.AnError.Error(), "Log should contain the error")
}

func TestLoggerNotInitialized(t *testing.T) {
	// Save original logger
	originalLogger := Logger

	// Set logger to nil
	Logger = nil

	// Test all logging functions
	LogEvent("123", "test_event", "user1", "success", "This is a test log")
	LogErrorEvent("123", "error_event", "user1", "failed", "This is an error log")
	LogWarnEvent("123", "warn_event", "user1", "warning", "This is a warning log")
	LogDebugEvent("123", "debug_event", "user1", "debugging", "This is a debug log")
	Error("Test error", nil)

	// Restore original logger
	Logger = originalLogger
}
