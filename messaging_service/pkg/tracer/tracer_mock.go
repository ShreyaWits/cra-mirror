package tracer

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// MockTracerService implements TracerService for testing
type MockTracerService struct {
	mu               sync.Mutex
	StartTracerCalls []StartTracerCall
	StopSpanCalls    []StopSpanCall
	AttributeCalls   []AttributeCall
	StatusCalls      []StatusCall
	ErrorCalls       []ErrorCall
}

// StartTracerCall records parameters from a call to StartTracer
type StartTracerCall struct {
	Ctx  context.Context
	Name string
}

// StopSpanCall records parameters from a call to StopSpan
type StopSpanCall struct {
	Span trace.Span
}

// AttributeCall records parameters from a call to SetAttributes
type AttributeCall struct {
	Span  trace.Span
	Attrs map[string]string
}

// StatusCall records parameters from a call to SetStatus
type StatusCall struct {
	Span    trace.Span
	Code    codes.Code
	Message string
}

// ErrorCall records parameters from a call to RecordError
type ErrorCall struct {
	Span trace.Span
	Err  error
}

// NewMockTracerService creates a new mock tracer service
func NewMockTracerService() *MockTracerService {
	return &MockTracerService{
		StartTracerCalls: []StartTracerCall{},
		StopSpanCalls:    []StopSpanCall{},
		AttributeCalls:   []AttributeCall{},
		StatusCalls:      []StatusCall{},
		ErrorCalls:       []ErrorCall{},
	}
}

// StartTracer records the call and returns a mock context and span
func (m *MockTracerService) StartTracer(ctx context.Context, name string) (context.Context, trace.Span) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.StartTracerCalls = append(m.StartTracerCalls, StartTracerCall{
		Ctx:  ctx,
		Name: name,
	})

	// Since we can't easily create a real span, we'll just return the original context
	// and the span from the context, or nil if there isn't one
	return ctx, trace.SpanFromContext(ctx)
}

// StopSpan records the call
func (m *MockTracerService) StopSpan(span trace.Span) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.StopSpanCalls = append(m.StopSpanCalls, StopSpanCall{
		Span: span,
	})
}

// SetAttributes records the call
func (m *MockTracerService) SetAttributes(span trace.Span, attrs map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.AttributeCalls = append(m.AttributeCalls, AttributeCall{
		Span:  span,
		Attrs: attrs,
	})
}

// SetStatus records the call
func (m *MockTracerService) SetStatus(span trace.Span, code codes.Code, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.StatusCalls = append(m.StatusCalls, StatusCall{
		Span:    span,
		Code:    code,
		Message: message,
	})
}

// RecordError records the call
func (m *MockTracerService) RecordError(span trace.Span, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ErrorCalls = append(m.ErrorCalls, ErrorCall{
		Span: span,
		Err:  err,
	})
}
