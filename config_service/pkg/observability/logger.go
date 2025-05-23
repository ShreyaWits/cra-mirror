package observability

import (
	"context"
	"log/slog"
)

// Logger is a wrapper around slog.Logger that enforces context-based logging
type Logger struct {
	*slog.Logger
}

// NewLogger creates a new Logger instance
func NewLogger(logger *slog.Logger) *Logger {
	return &Logger{Logger: logger}
}

// DebugContext logs a debug message with context and structured fields
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.Logger.DebugContext(ctx, msg, args...)
}

// InfoContext logs an info message with context and structured fields
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.Logger.InfoContext(ctx, msg, args...)
}

// WarnContext logs a warning message with context and structured fields
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.Logger.WarnContext(ctx, msg, args...)
}

// ErrorContext logs an error message with context and structured fields
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.Logger.ErrorContext(ctx, msg, args...)
}

// With returns a new Logger with the given attributes added to all log messages
func (l *Logger) With(args ...any) *Logger {
	return &Logger{Logger: l.Logger.With(args...)}
}

// WithGroup returns a new Logger that starts a group with the given name
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{Logger: l.Logger.WithGroup(name)}
}

// Enabled reports whether the logger handles records at the given level
func (l *Logger) Enabled(ctx context.Context, level slog.Level) bool {
	return l.Logger.Enabled(ctx, level)
}

// Handler returns the underlying slog.Handler
func (l *Logger) Handler() slog.Handler {
	return l.Logger.Handler()
}

// WithContext returns a new Logger with the given context
func (l *Logger) WithContext(ctx context.Context) *ContextLogger {
	return &ContextLogger{
		Logger: l,
		ctx:    ctx,
	}
}

// ContextLogger is a logger that has a context attached to it
type ContextLogger struct {
	*Logger
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
