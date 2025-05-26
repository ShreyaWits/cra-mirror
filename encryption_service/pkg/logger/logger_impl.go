package logger

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type customLogger struct {
	instance    *slog.Logger
	srcIP       string
	serviceName string
	isEnabled   bool
}

// Public constructor
func NewLogger(serviceName string, isEnabled bool) Logger {
	srcIP := getOutboundIP()

	var logger *slog.Logger

	if isEnabled {
		// Use OpenTelemetry slog bridge if enabled
		logger = otelslog.NewLogger(serviceName)
		fmt.Printf("Created OpenTelemetry-enabled logger for service: %s\n", serviceName)
	} else {
		// Fall back to standard slog if OTel is disabled
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelInfo,
		})
		logger = slog.New(handler)
		fmt.Printf("Created standard logger for service: %s (OpenTelemetry disabled)\n", serviceName)
	}

	// Add service info and host IP to logger
	logger = logger.With(
		"service", serviceName,
		"host_ip", srcIP,
		"version", os.Getenv("SERVICE_VERSION"),
		"environment", os.Getenv("ENVIRONMENT"),
	)

	return &customLogger{
		instance:    logger,
		srcIP:       srcIP,
		serviceName: serviceName,
		isEnabled:   isEnabled,
	}
}

// Helper to get IP
func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "unknown"
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
	if fields == nil {
		return l
	}

	// Convert map to key-value pairs
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}

	// Create new logger with added fields
	newLogger := l.instance.With(attrs...)

	return &customLogger{
		instance:    newLogger,
		srcIP:       l.srcIP,
		serviceName: l.serviceName,
		isEnabled:   l.isEnabled,
	}
}

// Helper method to log with context and trace information
func (l *customLogger) logWithContext(ctx context.Context, level slog.Level, args ...interface{}) {
	if !l.isEnabled {
		return
	}

	// Get span from context for tracing
	span := trace.SpanFromContext(ctx)

	// Convert args to string message
	msg := fmt.Sprint(args...)

	// Default attributes for the log
	attrs := []any{
		"timestamp", time.Now().Format(time.RFC3339),
		"service", l.serviceName,
	}

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

	// Extract request ID from context if available
	if requestID, ok := ctx.Value("request_id").(string); ok {
		attrs = append(attrs, "request_id", requestID)
	}

	// Log with the appropriate level and all attributes
	switch level {
	case slog.LevelInfo:
		l.instance.InfoContext(ctx, msg, attrs...)
	case slog.LevelError:
		l.instance.ErrorContext(ctx, msg, attrs...)
	case slog.LevelDebug:
		l.instance.DebugContext(ctx, msg, attrs...)
	case slog.LevelWarn:
		l.instance.WarnContext(ctx, msg, attrs...)
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
