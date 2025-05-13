package logger

import (
	"bytes"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newTestLogger(buf *bytes.Buffer, level zapcore.Level) *zap.Logger {
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(buf),
		level,
	)
	return zap.New(core)
}

func TestGetOutboundIP(t *testing.T) {
	InitLogger()
	getOutboundIP()

	assert.NotEmpty(t, srcIP, "Source IP should not be empty")
	ip := net.ParseIP(srcIP)
	assert.NotNil(t, ip, "Source IP should be a valid IP address")
}

func TestLogEvent(t *testing.T) {
	var buf bytes.Buffer
	Logger = newTestLogger(&buf, zapcore.InfoLevel)

	LogEvent("123", "test_event", "user1", "success", "This is a test log")
	logOutput := buf.String()

	assert.Contains(t, logOutput, "This is a test log")
	assert.Contains(t, logOutput, "test_event")
	assert.Contains(t, logOutput, "user1")
	assert.Contains(t, logOutput, "success")
}

func TestLogErrorEvent(t *testing.T) {
	var buf bytes.Buffer
	Logger = newTestLogger(&buf, zapcore.ErrorLevel)

	LogErrorEvent("123", "error_event", "user1", "failed", "This is an error log")
	logOutput := buf.String()

	assert.Contains(t, logOutput, "This is an error log")
	assert.Contains(t, logOutput, "error_event")
	assert.Contains(t, logOutput, "user1")
	assert.Contains(t, logOutput, "failed")
}

func TestLogWarnEvent(t *testing.T) {
	var buf bytes.Buffer
	Logger = newTestLogger(&buf, zapcore.WarnLevel)

	LogWarnEvent("123", "warn_event", "user1", "warning", "This is a warning log")
	logOutput := buf.String()

	assert.Contains(t, logOutput, "This is a warning log")
	assert.Contains(t, logOutput, "warn_event")
	assert.Contains(t, logOutput, "user1")
	assert.Contains(t, logOutput, "warning")
}

func TestLogDebugEvent(t *testing.T) {
	var buf bytes.Buffer
	Logger = newTestLogger(&buf, zapcore.DebugLevel)

	LogDebugEvent("123", "debug_event", "user1", "debugging", "This is a debug log")
	logOutput := buf.String()

	assert.Contains(t, logOutput, "This is a debug log")
	assert.Contains(t, logOutput, "debug_event")
	assert.Contains(t, logOutput, "user1")
	assert.Contains(t, logOutput, "debugging")
}

func TestError(t *testing.T) {
	var buf bytes.Buffer
	Logger = newTestLogger(&buf, zapcore.ErrorLevel)

	Error("Test error message", nil)
	logOutput := buf.String()
	assert.Contains(t, logOutput, "Test error message")

	// Test with error
	Error("Test error with err", assert.AnError)
	logOutput = buf.String()
	assert.Contains(t, logOutput, "Test error with err")
	assert.Contains(t, logOutput, assert.AnError.Error())
}

func TestLoggerNotInitialized(t *testing.T) {
	originalLogger := Logger
	Logger = nil

	// Should not panic or log
	LogEvent("123", "test_event", "user1", "success", "This is a test log")
	LogErrorEvent("123", "error_event", "user1", "failed", "This is an error log")
	LogWarnEvent("123", "warn_event", "user1", "warning", "This is a warning log")
	LogDebugEvent("123", "debug_event", "user1", "debugging", "This is a debug log")
	Error("Test error", nil)

	Logger = originalLogger
}
