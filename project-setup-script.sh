#!/bin/bash

# Set your project name
PROJECT_NAME=$1
if [ -z "$PROJECT_NAME" ]; then
  echo "Usage: $0 <project-name>"
  exit 1
fi
mkdir -p "$PROJECT_NAME"
cd "$PROJECT_NAME" || exit
# Initialize Go module
go mod init "$PROJECT_NAME"
# Root files
touch .air.toml .env.example .gitignore .gitlab-ci.yml Dockerfile Dockerfile.otel README.md docker-compose.yml docker-compose.local.yml otel-config.yaml

# Root directories
mkdir -p cmd/client
mkdir -p cmd/server
touch cmd/README.md cmd/server/main.go

mkdir -p internal/app
touch internal/app/README.md internal/app/modules.go internal/app/router.go

mkdir -p internal/common
mkdir -p internal/configs
mkdir -p internal/modules
touch internal/modules/README.md

# Example module
EXAMPLE_MODULE="example"
MODULE_PATH="internal/modules/$EXAMPLE_MODULE"
mkdir -p "$MODULE_PATH"/{apis,constants,di,models,repositories,services,utils}
touch "$MODULE_PATH"/README.md

# Package directory
mkdir -p pkg/{pgsql,redis,tracer}
touch pkg/README.md

# Scripts directory
mkdir -p scripts

echo "Project structure for '$PROJECT_NAME' has been created."
