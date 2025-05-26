package observability

import (
	"encryption_microservice/internal/config"
	"encryption_microservice/pkg/logger"
	metrics "encryption_microservice/pkg/matrics"
	"encryption_microservice/pkg/tracer"
	"fmt"
	"os"
	"strings"
)

// ObservabilityStack holds all observability components
type ObservabilityStack struct {
	TracerService  tracer.TracerService
	MetricsService metrics.MetricsService
	LoggerService  logger.Logger
	ServiceName    string
	IsEnabled      bool
}

// NewObservabilityStack creates a new observability stack with all components
func NewObservabilityStack(env *config.EnvConfig) *ObservabilityStack {
	if env == nil {
		fmt.Println("Warning: Creating observability stack with nil environment config")
		// Create a minimal default stack with telemetry disabled
		return &ObservabilityStack{
			TracerService:  tracer.NewTracer(config.SERVICE_NAME, false),
			MetricsService: metrics.NewMetricsService(config.SERVICE_NAME, false),
			LoggerService:  logger.NewLogger(config.SERVICE_NAME, false),
			ServiceName:    config.SERVICE_NAME,
			IsEnabled:      false,
		}
	}

	// Determine if telemetry should be enabled - check environment or config
	isEnabled := true // Default to enabled
	if val := os.Getenv("ENABLE_TELEMETRY"); val == "false" {
		isEnabled = false
	}

	// Get service name from environment or use default
	serviceName := config.SERVICE_NAME

	// Create components with consistent configuration
	tracerService := tracer.NewTracer(serviceName, isEnabled)
	metricsService := metrics.NewMetricsService(serviceName, isEnabled)
	loggerService := logger.NewLogger(serviceName, isEnabled)

	// Log successful initialization
	if isEnabled {
		fmt.Printf("Observability stack initialized with telemetry ENABLED for service: %s\n", serviceName)
	} else {
		fmt.Printf("Observability stack initialized with telemetry DISABLED for service: %s\n", serviceName)
	}

	return &ObservabilityStack{
		TracerService:  tracerService,
		MetricsService: metricsService,
		LoggerService:  loggerService,
		ServiceName:    serviceName,
		IsEnabled:      isEnabled,
	}
}

// IsObservabilityEnabled returns whether telemetry is enabled for this stack
func (o *ObservabilityStack) IsObservabilityEnabled() bool {
	return o.IsEnabled
}

// GetServiceName returns the service name used by this stack
func (o *ObservabilityStack) GetServiceName() string {
	return o.ServiceName
}

// StripProtocol standardizes the endpoint format by removing protocol prefixes
// This helps with consistent configuration between components
func StripProtocol(endpoint string) string {
	// Remove any protocol prefix if present
	if strings.HasPrefix(endpoint, "http://") {
		endpoint = endpoint[7:]
	} else if strings.HasPrefix(endpoint, "https://") {
		endpoint = endpoint[8:]
	}
	return endpoint
}
