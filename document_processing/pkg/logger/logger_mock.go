package logger

import (
	"context"
	"sync"
)

// MockLogger implements the Logger interface for testing
type MockLogger struct {
	mu            sync.Mutex
	InfoMessages  []LogEntry
	ErrorMessages []LogEntry
	DebugMessages []LogEntry
	WarnMessages  []LogEntry
	Fields        map[string]interface{}
	SyncCalled    bool
}

// LogEntry represents a log entry with its arguments and context
type LogEntry struct {
	Args []interface{}
	Ctx  context.Context
}

// NewMockLogger creates a new mock logger for testing
func NewMockLogger() *MockLogger {
	return &MockLogger{
		InfoMessages:  []LogEntry{},
		ErrorMessages: []LogEntry{},
		DebugMessages: []LogEntry{},
		WarnMessages:  []LogEntry{},
		Fields:        make(map[string]interface{}),
	}
}

// Info logs info level messages
func (m *MockLogger) Info(ctx context.Context, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.InfoMessages = append(m.InfoMessages, LogEntry{Args: args, Ctx: ctx})
}

// Error logs error level messages
func (m *MockLogger) Error(ctx context.Context, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ErrorMessages = append(m.ErrorMessages, LogEntry{Args: args, Ctx: ctx})
}

// Debug logs debug level messages
func (m *MockLogger) Debug(ctx context.Context, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DebugMessages = append(m.DebugMessages, LogEntry{Args: args, Ctx: ctx})
}

// Warn logs warning level messages
func (m *MockLogger) Warn(ctx context.Context, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.WarnMessages = append(m.WarnMessages, LogEntry{Args: args, Ctx: ctx})
}

// WithFields returns a new logger with the given fields added
func (m *MockLogger) WithFields(fields map[string]interface{}) Logger {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a new mock logger
	newLogger := NewMockLogger()

	// Copy existing fields
	for k, v := range m.Fields {
		newLogger.Fields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newLogger.Fields[k] = v
	}

	return newLogger
}

// Sync simulates flushing buffered log entries
func (m *MockLogger) Sync() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.SyncCalled = true
	return nil
}
