package logger

import (
	"context"
	"log/slog"
)

type Logger interface {
	DebugContext(ctx context.Context, msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	With(args ...any) Logger
	WithGroup(name string) Logger
	Enabled(ctx context.Context, level slog.Level) bool
	Handler() slog.Handler
	WithContext(ctx context.Context) *ContextLogger
}
