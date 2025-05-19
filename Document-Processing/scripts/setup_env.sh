#!/bin/bash

# Create .env file
cat > .env << EOL
SERVER_PORT=50051
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET_NAME=documents
GEMINI_API_KEY=your_gemini_api_key_here
EOL

echo "Environment variables have been set up in .env file"
echo "Please update GEMINI_API_KEY with your actual API key" 