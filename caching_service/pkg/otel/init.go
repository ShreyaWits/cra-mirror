package opentelemetry

import (
	"context"
	"errors"
	"redis-service/pkg/config"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/sdk/resource"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
)

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	grpcEndpoint := config.OTEL_COLLECTOR_GRPC_ENDPOINT

	// Create a resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.SERVICE_NAME),
			semconv.ServiceVersionKey.String(config.SERVICE_VERSION),
			semconv.DeploymentEnvironmentKey.String(config.ENVIRONMENT),
		),
	)
	if err != nil {
		return nil, err
	}

	// Set up trace provider with fallback to no-op
	tracerProvider, err := newTracerProvider(ctx, grpcEndpoint)
	if err != nil {
		// If trace provider setup fails, use no-op provider
		noopProvider := noop.NewTracerProvider()
		otel.SetTracerProvider(noopProvider)
	} else {
		shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
		otel.SetTracerProvider(tracerProvider)
	}

	// Set up meter provider with fallback to no-op
	if grpcEndpoint == "" {
		// Use no-op meter provider if no endpoint is configured
		otel.SetMeterProvider(metric.NewMeterProvider(
			metric.WithResource(res),
		))
	} else {
		meterProvider, err := newMeterProvider(grpcEndpoint)
		if err != nil {
			// If meter provider setup fails, use no-op provider
			otel.SetMeterProvider(metric.NewMeterProvider(
				metric.WithResource(res),
			))
		} else {
			shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
			otel.SetMeterProvider(meterProvider)
		}
	}

	// Set up logger provider with fallback to no-op
	if grpcEndpoint == "" {
		// Use no-op logger provider if no endpoint is configured
		global.SetLoggerProvider(log.NewLoggerProvider(
			log.WithResource(res),
		))
	} else {
		loggerProvider, err := newLoggerProvider(grpcEndpoint)
		if err != nil {
			// If logger provider setup fails, use no-op provider
			global.SetLoggerProvider(log.NewLoggerProvider(
				log.WithResource(res),
			))
		} else {
			shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
			global.SetLoggerProvider(loggerProvider)
		}
	}

	return shutdown, nil
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTracerProvider(ctx context.Context, endpoint string) (*trace.TracerProvider, error) {
	if endpoint == "" {
		return nil, errors.New("OTEL_COLLECTOR_GRPC_ENDPOINT is not set")
	}

	traceGrpcExporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)
	if err != nil {
		return nil, err
	}

	// Create a resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.SERVICE_NAME),
			semconv.ServiceVersionKey.String(config.SERVICE_VERSION),
			semconv.DeploymentEnvironmentKey.String(config.ENVIRONMENT),
		),
	)
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceGrpcExporter,
			trace.WithBatchTimeout(time.Second),
			trace.WithMaxExportBatchSize(512),
			trace.WithMaxQueueSize(2048),
		),
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()),
	)
	return tracerProvider, nil
}

func newMeterProvider(grpcEndpoint string) (*metric.MeterProvider, error) {
	metricExporter, err := otlpmetricgrpc.New(
		context.Background(),
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(grpcEndpoint),
	)
	if err != nil {
		return nil, err
	}
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			// Default is 1m. Set to 3s for demonstrative purposes.
			metric.WithInterval(3*time.Second))),
		metric.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(config.SERVICE_NAME),
			semconv.ServiceVersionKey.String(config.SERVICE_VERSION),
			semconv.DeploymentEnvironmentKey.String(config.ENVIRONMENT),
		),
		),
	)
	return meterProvider, nil
}

func newLoggerProvider(endpoint string) (*log.LoggerProvider, error) {
	logExporter, err := otlploggrpc.New(
		context.Background(),
		otlploggrpc.WithInsecure(),
		otlploggrpc.WithEndpoint(endpoint),
		otlploggrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)

	if err != nil {
		return nil, err
	}

	// Custom stdout exporter
	logStoutExporter, err := stdoutlog.New()
	if err != nil {
		return nil, err
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
		log.WithProcessor(log.NewBatchProcessor(logStoutExporter)),
		log.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(config.SERVICE_NAME),
			semconv.ServiceVersionKey.String(config.SERVICE_VERSION),
			semconv.DeploymentEnvironmentKey.String(config.ENVIRONMENT),
		),
		),
	)
	return loggerProvider, nil
}
