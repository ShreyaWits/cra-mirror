package tracer

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TracerServiceStruct implements the TracerService interface
type TracerServiceStruct struct {
	tracer      trace.Tracer
	enableTrace bool
	serviceName string
}

// NewTracer initializes the TracerService
func NewTracer(serviceName string, enableTracing bool) TracerService {
	if enableTracing {
		// Get a tracer from the global provider
		tracer := otel.Tracer(serviceName)
		fmt.Printf("Created OpenTelemetry-enabled tracer for service: %s\n", serviceName)

		return &TracerServiceStruct{
			tracer:      tracer,
			enableTrace: true,
			serviceName: serviceName,
		}
	}

	fmt.Printf("Created disabled tracer for service: %s (OpenTelemetry disabled)\n", serviceName)
	return &TracerServiceStruct{
		enableTrace: false,
		serviceName: serviceName,
	}
}

// StartTracer starts a new span and returns the context with the span.
func (t *TracerServiceStruct) StartTracer(ctx context.Context, name string) (context.Context, trace.Span) {
	if !t.enableTrace {
		// When tracing is disabled, return the original context and a no-op span
		return ctx, trace.SpanFromContext(ctx)
	}

	// Qualify the span name with service name for better organization
	spanName := fmt.Sprintf("%s.%s", t.serviceName, name)

	// Start a new span with standard options
	ctx, span := t.tracer.Start(ctx, spanName,
		trace.WithAttributes(
			attribute.String("service", t.serviceName),
		),
	)

	// Store the request ID in the span if available in context
	if requestID, ok := ctx.Value("request_id").(string); ok {
		span.SetAttributes(attribute.String("request_id", requestID))
	}

	return ctx, span
}

// StopSpan ends the span if tracing is enabled.
func (t *TracerServiceStruct) StopSpan(span trace.Span) {
	if t.enableTrace && span != nil {
		span.End()
	}
}

// SetAttributes sets attributes on the current span if tracing is enabled.
func (t *TracerServiceStruct) SetAttributes(span trace.Span, attrs map[string]string) {
	if !t.enableTrace || span == nil {
		return
	}

	// Add attributes to the span
	for key, value := range attrs {
		span.SetAttributes(attribute.String(key, value))
	}
}

// SetStatus sets the status of the span if tracing is enabled.
func (t *TracerServiceStruct) SetStatus(span trace.Span, code codes.Code, message string) {
	if !t.enableTrace || span == nil {
		return
	}

	span.SetStatus(code, message)
}

// RecordError records an error on the span if tracing is enabled.
func (t *TracerServiceStruct) RecordError(span trace.Span, err error) {
	if !t.enableTrace || span == nil || err == nil {
		return
	}

	// Record the error and mark the span as error
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}
