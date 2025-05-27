# Observability Package

This package provides observability functionality for the messaging service using OpenTelemetry, including tracing, metrics, and logging.

## Testing Approach

Testing the observability package is challenging due to several factors:

1. **External Dependencies**: The observability package relies on OpenTelemetry, which interacts with external systems like OpenTelemetry Collector.

2. **Global State**: OpenTelemetry uses global state for its providers, which can cause interference between tests.

3. **Models Dependency**: The `SetupOTelSDK` function depends on `models.EnvConfig` which appears to have breaking changes or a mismatch between expected fields.

### Implemented Tests

We've created the following tests:

- Basic mock tests for the `ObservabilityStack` using mock implementations of the tracer, metrics, and logger services.
- Tests for utility functions like `formatEndpoint`
- Tests for the `StdoutExporter` to verify its functionality

### Testing Limitations

Current limitations in test coverage:

1. **Setup Functions**: The `SetupOTelSDK` function and its helper functions can't be fully tested without a real OpenTelemetry Collector.

2. **EnvConfig Issues**: There appears to be a mismatch between the expected fields in `models.EnvConfig` and how they're accessed in the code.

### Improving Test Coverage

To properly test this package:

1. **Use Dependency Injection**: Refactor the code to avoid global state and use dependency injection.

2. **Mock External Services**: Create mock implementations of the OpenTelemetry exporters and providers.

3. **Interface Alignment**: Ensure the `models.EnvConfig` interface aligns with how it's used in the code.

4. **Environment for Testing**: Use Docker containers with test OpenTelemetry collectors for integration tests.

## Example Test Implementation

For proper unit testing:

```go
func TestSetupOTelSDK(t *testing.T) {
    // Setup a mock collector using a test container
    // Use dependency injection to provide test exporters
    // Assert that the setup functions correctly with the mocks
}
```

For integration testing:

```go
func TestSetupOTelSDK_Integration(t *testing.T) {
    // Skip if not in integration test environment
    // Use docker-compose to set up a test collector
    // Test the real implementation
}
```

## Future Improvements

1. Refactor the observability setup to be more testable by avoiding global state
2. Add unit test support with mocked exporters
3. Add integration tests using containerized services 