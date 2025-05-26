package logger

import (
	"context"
)

type Logger interface {
	// Basic logging methods with context support
	Info(ctx context.Context, args ...interface{})
	Error(ctx context.Context, args ...interface{})
	Debug(ctx context.Context, args ...interface{})
	Warn(ctx context.Context, args ...interface{})

	// Structured logging with fields
	WithFields(fields map[string]interface{}) Logger

	// Sync flushes any buffered log entries
	Sync() error
}
