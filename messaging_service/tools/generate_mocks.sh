#!/bin/bash

# Install mockgen if not already installed
command -v mockgen >/dev/null 2>&1 || { echo "Installing mockgen..."; go install github.com/golang/mock/mockgen@latest; }

# Generate mocks for MessagingService interface
mockgen -source=internal/messaging_service/service/messaging_service.go -destination=internal/messaging_service/mock/messaging_service_mock.go -package=mock

echo "Mocks generated successfully!" 