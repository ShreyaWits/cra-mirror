package logger

import (
	"context"
	"fmt"
	"log/slog"
	"messaging_service/internal/modules/message_broker/models"
	"net"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type customLogger struct {
	instance *slog.Logger
	srcIP    string
}

// Public constructor
func NewLogger(name string, isEnabled bool) Logger {
	srcIP := getOutboundIP()

	var logger *slog.Logger

	if isEnabled {
		// Use OpenTelemetry slog bridge if enabled
		logger = otelslog.NewLogger(name)
		fmt.Printf("Created OpenTelemetry-enabled logger for service: %s\n", name)
	} else {
		// Fall back to standard slog if OTel is disabled
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelInfo,
		})
		logger = slog.New(handler)
		fmt.Printf("Created standard logger for service: %s (OpenTelemetry disabled)\n", name)
	}

	// Add service info and host IP to logger
	logger = logger.With(
		"service", name,
		"host_ip", srcIP,
	)

	return &customLogger{
		instance: logger,
		srcIP:    srcIP,
	}
}

// Helper to get IP
func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		panic(fmt.Sprintf("failed to get outbound IP: %v", err))
	}
	defer conn.Close()

	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

// Logger methods with context support
func (l *customLogger) Info(ctx context.Context, args ...interface{}) {
	l.logWithContext(ctx, slog.LevelInfo, args...)
}

func (l *customLogger) Error(ctx context.Context, args ...interface{}) {
	l.logWithContext(ctx, slog.LevelError, args...)
}

func (l *customLogger) Debug(ctx context.Context, args ...interface{}) {
	l.logWithContext(ctx, slog.LevelDebug, args...)
}

func (l *customLogger) Warn(ctx context.Context, args ...interface{}) {
	l.logWithContext(ctx, slog.LevelWarn, args...)
}

func (l *customLogger) WithFields(fields map[string]interface{}) Logger {
	// Convert map to key-value pairs
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}

	// Create new logger with added fields
	newLogger := l.instance.With(attrs...)

	return &customLogger{
		instance: newLogger,
		srcIP:    l.srcIP,
	}
}

// Helper method to log with context and trace information
func (l *customLogger) logWithContext(ctx context.Context, level slog.Level, args ...interface{}) {
	// Get span from context for tracing
	span := trace.SpanFromContext(ctx)

	// Convert args to string message
	msg := fmt.Sprint(args...)

	// Create a new context for the logger to use
	logCtx := ctx

	// Default attributes for the log
	attrs := []any{}

	// Add trace information if we have a span
	if span != nil && span.SpanContext().IsValid() {
		// Add event to span
		span.AddEvent("log_event",
			trace.WithAttributes(
				attribute.String("level", level.String()),
				attribute.String("message", msg),
			),
		)

		// Add trace information to log
		attrs = append(attrs,
			"trace_id", span.SpanContext().TraceID().String(),
			"span_id", span.SpanContext().SpanID().String(),
		)
	}

	// Log with the appropriate level and all attributes
	switch level {
	case slog.LevelInfo:
		l.instance.InfoContext(logCtx, msg, attrs...)
	case slog.LevelError:
		l.instance.ErrorContext(logCtx, msg, attrs...)
	case slog.LevelDebug:
		l.instance.DebugContext(logCtx, msg, attrs...)
	case slog.LevelWarn:
		l.instance.WarnContext(logCtx, msg, attrs...)
	}
}

// Sync flushes any buffered log entries
func (l *customLogger) Sync() error {
	// slog doesn't have a Sync method, so this is a no-op
	return nil
}

// FormatEndpoint ensures an endpoint doesn't have a protocol prefix
// Exported for testing
func FormatEndpoint(endpoint string) string {
	// Remove any protocol prefix if present
	if strings.HasPrefix(endpoint, "http://") {
		endpoint = endpoint[7:]
	} else if strings.HasPrefix(endpoint, "https://") {
		endpoint = endpoint[8:]
	}
	return endpoint
}

// Check if observability backends are reachable
func ValidateObservabilityBackends(ctx context.Context, env *models.EnvConfig) error {
	// Try connecting to the OTLP endpoint
	conn, err := grpc.DialContext(
		ctx,
		FormatEndpoint(env.ObservabilityUrl),
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to observability backend: %w", err)
	}
	defer conn.Close()
	return nil
}
