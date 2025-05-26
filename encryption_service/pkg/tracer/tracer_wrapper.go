package tracer

import (
	"context"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TracerService defines the contract for the tracer service
type TracerService interface {
	// StartTracer starts a new span and returns the context with the span
	StartTracer(ctx context.Context, name string) (context.Context, trace.Span)

	// StopSpan ends the span
	StopSpan(span trace.Span)

	// SetAttributes sets attributes on the current span
	SetAttributes(span trace.Span, attrs map[string]string)

	// SetStatus sets the status of the span
	SetStatus(span trace.Span, code codes.Code, message string)

	// RecordError records an error on the span
	RecordError(span trace.Span, err error)
}
