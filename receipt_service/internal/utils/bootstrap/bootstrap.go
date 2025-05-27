package bootstrap

import (
	"context"
	"log"
	"nps-reciept-service/internal/config"
	"nps-reciept-service/internal/grpc"
	"nps-reciept-service/pkg/logger"
	"nps-reciept-service/pkg/metrics"
	"nps-reciept-service/pkg/observability"
	"nps-reciept-service/pkg/tracer"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Bootstrap initializes env, Redis and returns the gRPC server
func BootstrapServices() *grpc.GrpcServer {
	// Load environment variables
	if os.Getenv("IS_DOCKER") != "true" {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: No .env file found. Proceeding without it. Error: %v AND IS_DOCKER not true", err)
		} else {
			log.Println("Loaded .env file")
		}
	}

	// Initialize observability
	ctx := context.Background()
	obs := initializeObservability(ctx)

	// Redis setup
	redisHost := os.Getenv("REDIS_HOST")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisUsername := os.Getenv("REDIS_USERNAME")
	redisTTL := os.Getenv("RECEIPT_SERVICE_REDIS_TTL")
	redisPort := os.Getenv("REDIS_PORT")

	if redisHost == "" {
		log.Fatal("REDIS_HOST is not set")
	}
	if redisPassword == "" {
		log.Fatal("REDIS_PASSWORD is not set")
	}
	if redisUsername == "" {
		log.Fatal("REDIS_USERNAME is not set")
	}
	if redisTTL == "" {
		log.Fatal("RECEIPT_SERVICE_REDIS_TTL is not set")
	}
	if redisPort == "" {
		log.Fatal("REDIS_PORT is not set")
	}

	// Validate Redis port is a valid number
	if _, err := strconv.Atoi(redisPort); err != nil {
		log.Fatalf("Invalid REDIS_PORT: %v", err)
	}

	config.InitRedis(redisHost, redisPort, redisUsername, redisPassword, redisTTL)

	// gRPC setup
	grpcAddr := os.Getenv("RECEIPT_SERVICE_GRPC_PORT")
	if grpcAddr == "" {
		log.Fatal("RECEIPT_SERVICE_GRPC_PORT is not set")
	}
	grpcPort, err := strconv.Atoi(grpcAddr)
	if err != nil {
		log.Fatalf("Invalid RECEIPT_SERVICE_GRPC_PORT: %v", err)
	}

	return grpc.NewGrpcServer(grpcPort, obs)
}

func initializeObservability(ctx context.Context) *observability.ObservabilityStack {
	// Get configuration from environment
	enableTracing := os.Getenv("ENABLE_TRACING") == "true"
	enableMetrics := os.Getenv("ENABLE_METRICS") == "true"
	serviceName := "nps-receipt-service"

	// Initialize tracer provider
	tracerEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")
	if enableTracing && tracerEndpoint != "" {
		tracerProvider, err := observability.NewTracerProvider(ctx, tracerEndpoint)
		if err != nil {
			log.Printf("Warning: Failed to initialize tracer: %v", err)
		} else {
			otel.SetTracerProvider(tracerProvider)
		}
	}

	// Initialize meter provider
	metricsEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT")
	if enableMetrics && metricsEndpoint != "" {
		meterProvider, err := initializeMeterProvider(ctx, metricsEndpoint, serviceName)
		if err != nil {
			log.Printf("Warning: Failed to initialize metrics: %v", err)
		} else {
			otel.SetMeterProvider(meterProvider)
		}
	}

	// Initialize logger provider
	logsEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT")
	if logsEndpoint != "" {
		loggerProvider, err := observability.NewLoggerProvider(logsEndpoint)
		if err != nil {
			log.Printf("Warning: LoggerProvider init failed: %v", err)
		}
		if loggerProvider != nil {
			// Use logger provider if needed
		}
	}

	// Create and return observability stack
	return &observability.ObservabilityStack{
		TracerService:  tracer.NewTracer(serviceName, enableTracing),
		MetricsService: metrics.NewMetricsService(serviceName, enableMetrics),
		LoggerService:  logger.NewLogger(),
	}
}

func initializeMeterProvider(ctx context.Context, endpoint, serviceName string) (*sdkmetric.MeterProvider, error) {
	// Create gRPC connection to collector
	conn, err := grpclib.DialContext(ctx, endpoint,
		grpclib.WithTransportCredentials(insecure.NewCredentials()),
		grpclib.WithBlock(),
	)
	if err != nil {
		return nil, err
	}

	// Create OTLP exporter
	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithGRPCConn(conn),
	)
	if err != nil {
		return nil, err
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create and return MeterProvider
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				exporter,
				sdkmetric.WithInterval(10*time.Second),
			),
		),
	)

	return mp, nil
}
