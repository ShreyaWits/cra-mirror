package tracer

import (
	"context"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TracerServiceInterface defines the contract for the tracer service.
type TracerService interface {
	StartTracer(ctx context.Context, name string) (context.Context, trace.Span)
	StopSpan(span trace.Span)
	SetAttributes(span trace.Span, attrs map[string]string)
	SetStatus(span trace.Span, code codes.Code, message string)
	RecordError(span trace.Span, err error)
}
