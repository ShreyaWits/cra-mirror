#!/bin/bash
# This script runs tests f and generates coverage report in tests/coverage.out and displays coverage report in html format

# Exit immediately if any command fails
set -e

# Run tests
go test ./... -coverprofile=tests/coverage.out

# Generate HTML coverage report
go tool cover -html=tests/coverage.out -o tests/coverage.html

# Display coverage report in html format
go tool cover -html=tests/coverage.out
