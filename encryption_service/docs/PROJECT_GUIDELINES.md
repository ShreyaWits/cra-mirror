
# Project Structure Guideline

This document outlines the recommended project structure for new projects based on this boilerplate.

## Root Directory Structure

```
project-name/
├── .air.toml                 # Air configuration for live reloading
├── .env.example              # Example environment variables
├── .gitignore                # Git ignore rules
├── .gitlab-ci.yml            # GitLab CI/CD configuration
├── Dockerfile                # Main Dockerfile for the application
├── Dockerfile.otel           # OpenTelemetry collector Dockerfile
├── README.md                 # Project documentation
├── docker-compose.yml        # Docker Compose configuration
├── docker-compose.local.yml  # Local development Docker Compose config
├── go.mod                    # Go module file
├── go.sum                    # Go module checksums
├── otel-config.yaml          # OpenTelemetry configuration
├── cmd/                      # Application entry points
├── internal/                 # Private application code
├── pkg/                      # Shared, exportable packages
└── scripts/                  # Utility scripts
```

## Command Directory Structure (cmd/)

The `cmd` directory contains entry points for the application:

```
cmd/
├── README.md                 # Directory documentation
├── client/                   # Client application entrypoint
└── server/                   # Server application entrypoint
    └── main.go               # Main server entry point
```

## Internal Directory Structure (internal/)

The `internal` directory contains code that's specific to this application and not meant to be imported by other projects:

```
internal/
├── app/                      # Application core
│   ├── README.md             # Directory documentation
│   ├── modules.go            # Module registration and wiring
│   └── router.go             # API router configuration
├── common/                   # Common utilities and shared code
├── configs/                  # Configuration handlers
└── modules/                  # Feature modules
    ├── README.md             # Module documentation
    └── [module-name]/        # Individual feature modules
        ├── README.md         # Module documentation
        ├── apis/             # HTTP handlers and routes
        ├── constants/        # Module-specific constants
        ├── di/               # Dependency injection setup
        ├── models/           # Data models/entities
        ├── repositories/     # Data access layer
        ├── services/         # Business logic
        └── utils/            # Module-specific utilities
```

## Package Directory Structure (pkg/)

The `pkg` directory contains reusable packages that can be imported by other projects:

```
pkg/
├── README.md                 # Directory documentation
├── pgsql/                    # PostgreSQL client library
├── redis/                    # Redis client library
└── tracer/                   # Distributed tracing utilities
```

## Module Structure Guidelines

Each module should be organized according to clean architecture principles:

1. **APIs Layer**: HTTP handlers, request validation, and response formatting
2. **Services Layer**: Business logic implementation
3. **Repository Layer**: Data access and persistence
4. **Models**: Data structures and entities
5. **Constants**: Module-specific constants and configurations
6. **Utils**: Helper functions specific to the module

## Dependency Injection

Modules should use dependency injection to manage their dependencies:

- The `di` directory in each module should contain factories and wiring code
- Dependencies should flow from outer layers (APIs) to inner layers (repositories)

## Module Registration

New modules should be registered in the `internal/app/modules.go` file:

1. Create the module directory following the structure above
2. Implement the required interface for the module
3. Register the module in `modules.go`
4. Configure routes in the appropriate location

## Best Practices

1. **Separation of Concerns**: Each package and module should have a single responsibility
2. **Dependency Injection**: Use DI for better testability and loose coupling
3. **Clean Architecture**: Keep business logic separate from frameworks and external dependencies
4. **Consistent Naming**: Follow Go naming conventions and maintain consistency across the project
5. **Documentation**: Each module and package should have a README explaining its purpose and usage
6. **Testing**: Include tests for critical functionality

## Getting Started with a New Module

To create a new module:

1. Create a directory in `internal/modules/[module-name]`
2. Set up the standard directory structure (apis, services, repositories, etc.)
3. Define your models and interfaces
4. Implement repositories and services
5. Create HTTP handlers in the APIs layer
6. Register the module in `internal/app/modules.go`
7. Add routes to the router

## Configuration Management

Store configuration in:

- Environment variables (with defaults in `.env.example`)
- Configuration files in `internal/configs/`
- Use a configuration object to pass settings to modules

## Deployment

The project includes:

- Docker and Docker Compose files for containerization
- GitLab CI/CD configuration for automated pipelines
- OpenTelemetry integration for observability

## Observability

Built-in support for:

- Logging
- Metrics
- Distributed tracing through OpenTelemetry
apps-fileview.texmex_20250403.00_p0
PROJECT_STRUCTURE_GUIDELINE.md
Displaying PROJECT_STRUCTURE_GUIDELINE.md.