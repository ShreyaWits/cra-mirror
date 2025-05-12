package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// --- Mocks for OpenTelemetry ---

type mockSpan struct {
	trace.Span
	ended         bool
	attributes    []attribute.KeyValue
	recordedError error
	statusCode    codes.Code
	statusDesc    string
}

func (m *mockSpan) End(options ...trace.SpanEndOption) {
	m.ended = true
}
func (m *mockSpan) SetAttributes(kv ...attribute.KeyValue) {
	m.attributes = append(m.attributes, kv...)
}
func (m *mockSpan) RecordError(err error, opts ...trace.EventOption) {
	m.recordedError = err
}
func (m *mockSpan) SetStatus(code codes.Code, description string) {
	m.statusCode = code
	m.statusDesc = description
}

type mockTracer struct {
	trace.Tracer
	span *mockSpan
}

func (m *mockTracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	m.span = &mockSpan{}
	return ctx, m.span
}

// --- Test Setup ---

func setMockTracer() *mockTracer {
	mt := &mockTracer{}
	otel.SetTracerProvider(noop.NewTracerProvider())
	otel.SetTracerProvider(&mockTracerProvider{tracer: mt})
	return mt
}

type mockTracerProvider struct {
	trace.TracerProvider
	tracer *mockTracer
}

func (m *mockTracerProvider) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return m.tracer
}

// --- Tests ---

func TestTraceInterceptor_Success(t *testing.T) {
	mt := setMockTracer()
	interceptor := TraceInterceptor()

	ctx := context.Background()
	req := "test-request"
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		time.Sleep(10 * time.Millisecond) // Simulate some work
		return "test-response", nil
	}

	resp, err := interceptor(ctx, req, info, handler)

	assert.True(t, handlerCalled, "handler should be called")
	assert.NoError(t, err)
	assert.Equal(t, "test-response", resp)

	span := mt.span
	assert.True(t, span.ended, "span should be ended")
	assert.Contains(t, span.attributes, attribute.String("rpc.system", "grpc"))
	assert.Contains(t, span.attributes, attribute.String("rpc.service", info.FullMethod))
	assert.Contains(t, span.attributes, attribute.Int("rpc.grpc.status_code", int(grpccodes.OK)))
}

func TestTraceInterceptor_Error(t *testing.T) {
	mt := setMockTracer()
	interceptor := TraceInterceptor()

	ctx := context.Background()
	req := "test-request"
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return nil, status.Error(grpccodes.Internal, "internal error")
	}

	resp, err := interceptor(ctx, req, info, handler)

	assert.True(t, handlerCalled, "handler should be called")
	assert.Error(t, err)
	assert.Nil(t, resp)

	span := mt.span
	assert.True(t, span.ended, "span should be ended")
	assert.Contains(t, span.attributes, attribute.String("rpc.system", "grpc"))
	assert.Contains(t, span.attributes, attribute.String("rpc.service", info.FullMethod))
	assert.Contains(t, span.attributes, attribute.Int("rpc.grpc.status_code", int(grpccodes.Internal)))
	assert.Contains(t, span.statusDesc, "internal error")
	assert.Equal(t, codes.Error, span.statusCode)
	assert.EqualError(t, span.recordedError, "rpc error: code = Internal desc = internal error")
}
