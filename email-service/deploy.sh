#!/bin/bash
set -e

echo "🚀 Deploying Email Service..."

# Load environment variables
if [ -f .env ]; then
  export $(cat .env | grep -v '^#' | xargs)
fi

# Build Docker image
echo "📦 Building Docker image..."
docker-compose build

# Stop existing container
echo "🛑 Stopping existing container..."
docker-compose down || true

# Start new container
echo "▶️  Starting new container..."
docker-compose up -d

echo "✅ Deployment complete!"

