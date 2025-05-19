package tracer

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TracerService defines a wrapper for OpenTelemetry tracing.
type TracerServiceStruct struct {
	tracer      trace.Tracer
	enableTrace bool
}

// NewTracerService initializes the TracerService based on the ENABLE_TRACING environment variable.
func NewTracer(pkgName string, enableTracing bool) *TracerServiceStruct {

	if enableTracing {
		tracer := otel.Tracer(pkgName)
		return &TracerServiceStruct{
			tracer:      tracer,
			enableTrace: true,
		}
	}

	return &TracerServiceStruct{
		enableTrace: false,
	}
}

// StartTracer starts a new span and returns the context with the span.
func (t *TracerServiceStruct) StartTracer(ctx context.Context, name string) (context.Context, trace.Span) {
	if t.enableTrace {
		return t.tracer.Start(ctx, name)
	}
	return ctx, trace.SpanFromContext(ctx)
}

// StopSpan ends the span if tracing is enabled.
func (t *TracerServiceStruct) StopSpan(span trace.Span) {
	if t.enableTrace && span != nil {
		span.End()
	}
}

// SetAttributes sets attributes on the current span if tracing is enabled.
func (t *TracerServiceStruct) SetAttributes(span trace.Span, attrs map[string]string) {
	if t.enableTrace && span != nil {
		for key, value := range attrs {
			span.SetAttributes(attribute.String(key, value))
		}
	}
}

// SetStatus sets the status of the span if tracing is enabled.
func (t *TracerServiceStruct) SetStatus(span trace.Span, code codes.Code, message string) {
	if t.enableTrace && span != nil {
		span.SetStatus(code, message)
	}
}

// RecordError records an error on the span if tracing is enabled.
func (t *TracerServiceStruct) RecordError(span trace.Span, err error) {
	if t.enableTrace && span != nil && err != nil {
		span.RecordError(err)
	}
}
