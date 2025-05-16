package grpc

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"redis-service/internal/app"

	"github.com/stretchr/testify/assert"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func setupTestContainer() {
	// Create a no-op tracer
	tracer := tracenoop.NewTracerProvider().Tracer("test")
	// Create a new logger that discards output
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	// Create a no-op meter
	meter := metricnoop.NewMeterProvider().Meter("test")

	app.Di = &app.Container{
		Tracer: tracer,
		Logger: logger,
		Meter:  meter,
	}
}

func TestTraceInterceptor_Success(t *testing.T) {
	setupTestContainer()

	interceptor := TraceInterceptor()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Test/Test",
	}

	resp, err := interceptor(context.Background(), "test", info, handler)
	assert.NoError(t, err)
	assert.Equal(t, "success", resp)
}

func TestTraceInterceptor_Error(t *testing.T) {
	setupTestContainer()

	interceptor := TraceInterceptor()
	expectedErr := status.Error(codes.Internal, "test error")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, expectedErr
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Test/Test",
	}

	resp, err := interceptor(context.Background(), "test", info, handler)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, expectedErr, err)
}

func TestTraceInterceptor_Panic(t *testing.T) {
	setupTestContainer()

	interceptor := TraceInterceptor()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		panic("test panic")
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Test/Test",
	}

	assert.Panics(t, func() {
		_, _ = interceptor(context.Background(), "test", info, handler)
	})
}
