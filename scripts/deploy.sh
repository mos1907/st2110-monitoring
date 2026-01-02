#!/bin/bash
# deploy.sh - Deployment script for ST 2110 Monitoring Stack

set -e

echo "🚀 Deploying ST 2110 Monitoring Stack..."
echo

# Check if .env exists
if [ ! -f .env ]; then
    echo "⚠️  .env file not found. Copying from .env.example..."
    if [ -f .env.example ]; then
        cp .env.example .env
        echo "✅ Created .env file. Please edit it with your credentials."
        echo "   Then run this script again."
        exit 1
    else
        echo "❌ .env.example not found. Please create .env file manually."
        exit 1
    fi
fi

# Check if config files exist
if [ ! -f config/streams.yaml ]; then
    echo "⚠️  config/streams.yaml not found. Copying from example..."
    if [ -f config/streams.yaml.example ]; then
        cp config/streams.yaml.example config/streams.yaml
        echo "✅ Created config/streams.yaml. Please edit it with your stream definitions."
    else
        echo "❌ config/streams.yaml.example not found."
        exit 1
    fi
fi

if [ ! -f config/switches.yaml ]; then
    echo "⚠️  config/switches.yaml not found. Copying from example..."
    if [ -f config/switches.yaml.example ]; then
        cp config/switches.yaml.example config/switches.yaml
        echo "✅ Created config/switches.yaml. Please edit it with your switch configurations."
    else
        echo "❌ config/switches.yaml.example not found."
        exit 1
    fi
fi

# Build and start services
echo "🔨 Building Docker images..."
docker-compose build

echo "🚀 Starting services..."
docker-compose up -d

echo "⏳ Waiting for services to start..."
sleep 10

# Wait for services to be healthy
echo "🔍 Checking service health..."
MAX_RETRIES=30
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -sf http://localhost:9090/-/healthy > /dev/null 2>&1 && \
       curl -sf http://localhost:3000/api/health > /dev/null 2>&1; then
        echo "✅ Services are healthy!"
        break
    fi
    
    RETRY_COUNT=$((RETRY_COUNT + 1))
    echo "   Waiting... ($RETRY_COUNT/$MAX_RETRIES)"
    sleep 2
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo "⚠️  Services took longer than expected to start. Check logs with: docker-compose logs"
fi

echo
echo "✅ Deployment complete!"
echo
echo "📊 Access the monitoring stack:"
echo "   Grafana:      http://localhost:3000 (admin/admin)"
echo "   Prometheus:   http://localhost:9090"
echo "   Alertmanager: http://localhost:9093"
echo
echo "💡 Run health check: ./scripts/health-check.sh"
echo "💡 View logs: docker-compose logs -f"

