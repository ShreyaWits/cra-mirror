package grpc

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TraceInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		tracer := otel.Tracer("grpc-server")
		spanName := info.FullMethod

		ctx, span := tracer.Start(ctx, spanName)
		defer span.End()

		// Add request attributes
		span.SetAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.service", info.FullMethod),
		)

		// Execute the handler
		startTime := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(startTime)

		// Add response attributes
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
		} else {
			span.SetAttributes(
				attribute.Int("rpc.grpc.status_code", int(grpccodes.OK)),
			)
		}

		return resp, err
	}
}
