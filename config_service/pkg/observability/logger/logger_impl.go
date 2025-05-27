package logger

import (
	"context"
	"log/slog"
)

// LoggerStruct is a wrapper around slog.LoggerStruct that enforces context-based logging
type LoggerStruct struct {
	*slog.Logger
}

// NewLogger creates a new LoggerStruct instance
func NewLogger(logger *slog.Logger) Logger {
	return &LoggerStruct{Logger: logger}
}

// DebugContext logs a debug message with context and structured fields
func (l *LoggerStruct) DebugContext(ctx context.Context, msg string, args ...any) {
	l.Logger.DebugContext(ctx, msg, args...)
}

// InfoContext logs an info message with context and structured fields
func (l *LoggerStruct) InfoContext(ctx context.Context, msg string, args ...any) {
	l.Logger.InfoContext(ctx, msg, args...)
}

// WarnContext logs a warning message with context and structured fields
func (l *LoggerStruct) WarnContext(ctx context.Context, msg string, args ...any) {
	l.Logger.WarnContext(ctx, msg, args...)
}

// ErrorContext logs an error message with context and structured fields
func (l *LoggerStruct) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.Logger.ErrorContext(ctx, msg, args...)
}

// With returns a new LoggerStruct with the given attributes added to all log messages
func (l *LoggerStruct) With(args ...any) Logger {
	return &LoggerStruct{Logger: l.Logger.With(args...)}
}

// WithGroup returns a new LoggerStruct that starts a group with the given name
func (l *LoggerStruct) WithGroup(name string) Logger {
	return &LoggerStruct{Logger: l.Logger.WithGroup(name)}
}

// Enabled reports whether the logger handles records at the given level
func (l *LoggerStruct) Enabled(ctx context.Context, level slog.Level) bool {
	return l.Logger.Enabled(ctx, level)
}

// Handler returns the underlying slog.Handler
func (l *LoggerStruct) Handler() slog.Handler {
	return l.Logger.Handler()
}

// WithContext returns a new LoggerStruct with the given context
func (l *LoggerStruct) WithContext(ctx context.Context) *ContextLogger {
	return &ContextLogger{
		Logger: l,
		ctx:    ctx,
	}
}

// ContextLogger is a logger that has a context attached to it
type ContextLogger struct {
	Logger
	ctx context.Context
}

// Debug logs a debug message with the attached context
func (l *ContextLogger) Debug(msg string, args ...any) {
	l.Logger.DebugContext(l.ctx, msg, args...)
}

// Info logs an info message with the attached context
func (l *ContextLogger) Info(msg string, args ...any) {
	l.Logger.InfoContext(l.ctx, msg, args...)
}

// Warn logs a warning message with the attached context
func (l *ContextLogger) Warn(msg string, args ...any) {
	l.Logger.WarnContext(l.ctx, msg, args...)
}

// Error logs an error message with the attached context
func (l *ContextLogger) Error(msg string, args ...any) {
	l.Logger.ErrorContext(l.ctx, msg, args...)
}

// With returns a new ContextLogger with the given attributes added to all log messages
func (l *ContextLogger) With(args ...any) *ContextLogger {
	return &ContextLogger{
		Logger: l.Logger.With(args...),
		ctx:    l.ctx,
	}
}

// WithGroup returns a new ContextLogger that starts a group with the given name
func (l *ContextLogger) WithGroup(name string) *ContextLogger {
	return &ContextLogger{
		Logger: l.Logger.WithGroup(name),
		ctx:    l.ctx,
	}
}
