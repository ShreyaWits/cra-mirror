package grpc

import (
	"context"
	"redis-service/internal/app"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TraceInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		tracer := app.Di.Tracer
		logger := app.Di.Logger
		meter := app.Di.Meter

		// Create a counter for RPC calls
		rpcCounter, _ := meter.Int64Counter("grpc.calls.total")
		rpcDuration, _ := meter.Float64Histogram("grpc.duration.seconds")

		// Start a new span only if we have a valid tracer
		var span trace.Span
		if tracer != nil {
			spanName := info.FullMethod
			ctx, span = tracer.Start(ctx, spanName)
			defer span.End()

			// Add request attributes
			span.SetAttributes(
				attribute.String("rpc.system", "grpc"),
				attribute.String("rpc.service", info.FullMethod),
			)
		}

		// Log incoming request
		logger.Info("incoming gRPC request",
			"method", info.FullMethod,
		)

		// Execute the handler
		startTime := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(startTime)

		// Record metrics
		if meter != nil {
			rpcCounter.Add(ctx, 1,
				metric.WithAttributes(
					attribute.String("method", info.FullMethod),
					attribute.String("status", statusFromError(err).String()),
				),
			)
			rpcDuration.Record(ctx, duration.Seconds(),
				metric.WithAttributes(
					attribute.String("method", info.FullMethod),
				),
			)
		}

		// Add response attributes if we have a span
		if span != nil {
			span.SetAttributes(
				attribute.Int64("rpc.duration_ms", duration.Milliseconds()),
			)

			if err != nil {
				s, _ := status.FromError(err)
				span.SetAttributes(
					attribute.Int("rpc.grpc.status_code", int(s.Code())),
				)
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				logger.Error("gRPC request failed",
					"method", info.FullMethod,
					"error", err.Error(),
					"duration_ms", duration.Milliseconds(),
				)
			} else {
				span.SetAttributes(
					attribute.Int("rpc.grpc.status_code", int(grpccodes.OK)),
				)
				logger.Info("gRPC request completed",
					"method", info.FullMethod,
					"duration_ms", duration.Milliseconds(),
				)
			}
		}

		return resp, err
	}
}

func statusFromError(err error) grpccodes.Code {
	if err == nil {
		return grpccodes.OK
	}
	s, _ := status.FromError(err)
	return s.Code()
}
