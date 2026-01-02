.PHONY: help build test run stop clean backup restore logs

# Default target
help:
	@echo "ST 2110 Monitoring Stack - Available Commands:"
	@echo ""
	@echo "  make build          - Build all exporters"
	@echo "  make up             - Start all services"
	@echo "  make down           - Stop all services"
	@echo "  make restart        - Restart all services"
	@echo "  make logs           - Show logs"
	@echo "  make test           - Run tests"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make backup         - Backup Prometheus data"
	@echo "  make restore        - Restore from backup"
	@echo "  make health-check   - Check service health"
	@echo ""

# Build all exporters
build:
	@echo "Building RTP Exporter..."
	cd exporters/rtp && go build -o ../../bin/st2110-rtp-exporter
	@echo "Building PTP Exporter..."
	cd exporters/ptp && go build -o ../../bin/st2110-ptp-exporter
	@echo "Building gNMI Collector..."
	cd exporters/gnmi && go build -o ../../bin/st2110-gnmi-collector
	@echo "✅ Build complete"

# Start services
up:
	@echo "Starting ST 2110 monitoring stack..."
	docker-compose up -d
	@echo "✅ Services started"
	@echo "   Grafana:      http://localhost:3000 (admin/admin)"
	@echo "   Prometheus:   http://localhost:9090"
	@echo "   Alertmanager: http://localhost:9093"

# Stop services
down:
	@echo "Stopping services..."
	docker-compose down
	@echo "✅ Services stopped"

# Restart services
restart: down up

# Show logs
logs:
	docker-compose logs -f

# Run tests
test:
	@echo "Running tests..."
	go test ./...
	@echo "✅ Tests passed"

# Clean
clean:
	@echo "Cleaning..."
	rm -rf bin/
	docker-compose down -v
	@echo "✅ Cleaned"

# Backup Prometheus data
backup:
	@mkdir -p backups
	@echo "Backing up Prometheus data..."
	docker run --rm \
		-v st2110-monitoring_prometheus_data:/data \
		-v $(PWD)/backups:/backup \
		alpine tar czf /backup/prometheus-backup-$$(date +%Y%m%d-%H%M%S).tar.gz -C /data .
	@echo "✅ Backup created in backups/"

# Restore from backup
restore:
	@echo "Available backups:"
	@ls -lh backups/
	@read -p "Enter backup filename: " backup; \
	docker run --rm \
		-v st2110-monitoring_prometheus_data:/data \
		-v $(PWD)/backups:/backup \
		alpine tar xzf /backup/$$backup -C /data
	@echo "✅ Backup restored"

# Health check
health-check:
	@echo "Checking service health..."
	@./scripts/health-check.sh

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	@echo "✅ Dependencies installed"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted"

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run
	@echo "✅ Lint passed"

# Generate mocks
generate:
	@echo "Generating mocks..."
	go generate ./...
	@echo "✅ Mocks generated"

# Deploy to Kubernetes
k8s-deploy:
	@echo "Deploying to Kubernetes..."
	kubectl apply -f kubernetes/
	@echo "✅ Deployed to Kubernetes"

# Remove from Kubernetes
k8s-delete:
	@echo "Removing from Kubernetes..."
	kubectl delete -f kubernetes/
	@echo "✅ Removed from Kubernetes"

# Create config from examples
config:
	@echo "Creating configuration files..."
	cp config/streams.yaml.example config/streams.yaml
	cp config/switches.yaml.example config/switches.yaml
	cp .env.example .env
	@echo "✅ Configuration files created"
	@echo "   Please edit config/*.yaml and .env with your settings"

# Quick start (create config + start services)
quickstart: config
	@echo "🚀 Quick starting ST 2110 monitoring..."
	@echo "   Please edit config/streams.yaml with your multicast addresses"
	@read -p "Press Enter when ready to start services..."
	@make up
	@echo "✅ Monitoring stack is running!"

