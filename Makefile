.PHONY: build run test clean docker-build docker-run deps lint format help nats kafka dev prod

# Variables
APP_NAME := ads-metric-tracker
DOCKER_IMAGE := $(APP_NAME):latest
DOCKER_COMPOSE_NATS := docker-compose.nats.yaml
DOCKER_COMPOSE_FILE := $(DOCKER_COMPOSE_NATS)

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

deps: ## Install dependencies
	go mod download
	go mod tidy

run: ## Run the application locally
	go run ./cmd/main.go

.PHONY: swagger
swagger:
	@swag init -g main.go -d ./cmd


build: ## Build the application
	go build -o bin/$(APP_NAME) ./cmd/main.go

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -cover ./...

test-coverage-html: ## Generate HTML coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linter
	golangci-lint run

format: ## Format code
	go fmt ./...

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

##@ Docker

docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

docker-run-nats: ## Run application with NATS
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE)

docker-run-prod: ## Run application with Production setup
	docker-compose -f $(DOCKER_COMPOSE_PROD) up -d

docker-push: ## Push Docker image
	docker push $(DOCKER_IMAGE)

##@ Docker Compose

# Quick setup commands
nats: ## Start application with NATS (recommended)
	@echo "🚀 Starting Ads Metric Tracker with NATS..."
	docker-compose -f $(DOCKER_COMPOSE_NATS) up -d
	@echo "✅ All services started successfully!"
	@echo "📊 Access points:"
	@echo "   • API: http://localhost:8080"
	@echo "   • Health: http://localhost:8080/health"
	@echo "   • Metrics: http://localhost:8080/metrics"
	@echo "   • Prometheus: http://localhost:9090"
	@echo "   • Grafana: http://localhost:3000 (admin/admin123)"
	@echo "   • NATS Monitoring: http://localhost:8222"


dev: nats ## Alias for development (NATS setup)


up: nats ## Default: Start with NATS configuration

down: ## Stop all services
	@echo "🛑 Stopping all services..."
	docker-compose -f $(DOCKER_COMPOSE_NATS) down 2>/dev/null || true
	@echo "✅ All services stopped"

down-nats: ## Stop NATS services
	docker-compose -f $(DOCKER_COMPOSE_NATS) down

ps: ## Show running containers
	docker-compose -f $(DOCKER_COMPOSE_FILE) ps

logs: ## Show logs
	docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f

logs-app: ## Show application logs
	docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f ads-tracker

restart: ## Restart all services
	docker-compose -f $(DOCKER_COMPOSE_FILE) restart

rebuild: ## Rebuild and restart application
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d --build ads-tracker

##@ Database

db-connect: ## Connect to PostgreSQL database
	docker-compose -f $(DOCKER_COMPOSE_FILE) exec postgres psql -U adsuser -d adsmetrics

db-migrate: ## Run database migrations (if implemented)
	echo "Database migrations not implemented yet"

db-seed: ## Seed database with sample data
	echo "Database seeding happens automatically on startup"

##@ Monitoring

prometheus: ## Open Prometheus in browser
	open http://localhost:9090

grafana: ## Open Grafana in browser
	open http://localhost:3000

health: ## Check application health
	curl -s http://localhost:8080/health

metrics: ## Show Prometheus metrics
	curl -s http://localhost:8080/metrics

##@ Testing

test-api: ## Test API endpoints
	@echo "🧪 Testing API endpoints..."
	@echo "Testing GET /ads..."
	curl -s http://localhost:8080/ads | head -c 200
	@echo "\n\nTesting POST /ads/click..."
	curl -X POST http://localhost:8080/ads/click \
		-H "Content-Type: application/json" \
		-d '{"ad_id": "tech-001", "ip": "192.168.1.100", "video_play_time": 30, "timestamp": "'$$(date -u +%Y-%m-%dT%H:%M:%SZ)'"}'
	@echo "\n\nTesting GET /ads/analytics..."
	curl -s "http://localhost:8080/ads/analytics?ad_id=tech-001"
	@echo "\n"

test-comprehensive: ## Run comprehensive requirements test
	@echo "🧪 Running comprehensive requirements test..."
	./test_requirements.sh

load-test-basic: ## Run bash-based load test (full test suite)
	@echo "🚀 Running comprehensive load test using bash/curl..."
	@if [ -f test_api.sh ]; then \
		chmod +x test_api.sh && ./test_api.sh; \
	else \
		echo "❌ test_api.sh not found"; exit 1; \
	fi

load-test-quick: ## Run quick load test (clicks only)
	@echo "🚀 Running quick load test on /ads/click endpoint..."
	@echo "Sending 100 concurrent requests to /ads/click..."
	@for i in $$(seq 1 100); do \
		curl -s -X POST http://localhost:8080/ads/click \
			-H "Content-Type: application/json" \
			-d '{"ad_id":"tech-001","ip":"192.168.1.'$$i'","video_play_time":30}' & \
	done; wait
	@echo "✅ Load test completed!"

load-test: ## Run load test (100 requests)
	@echo "🚀 Running load test (100 requests)..."
	@start_time=$$(date +%s); \
	for i in $$(seq 1 100); do \
		curl -s -o /dev/null -X POST http://localhost:8080/ads/click \
			-H "Content-Type: application/json" \
			-d '{"ad_id":"tech-001","ip":"192.168.1.'$$((i%255+1))'","video_play_time":30}' & \
	done; wait; \
	end_time=$$(date +%s); \
	duration=$$((end_time - start_time)); \
	if [$$duration -eq 0];then duration=1; fi; \
	rps=$$((100 / duration)); \
	echo "✅ Load test completed in $${duration}s ($$rps RPS)"

##@ Maintenance

clean-docker: ## Clean Docker resources
	docker system prune -f
	docker volume prune -f

backup-db: ## Backup PostgreSQL database
	docker-compose -f $(DOCKER_COMPOSE_FILE) exec postgres pg_dump -U adsuser adsmetrics > backup_$$(date +%Y%m%d_%H%M%S).sql

restore-db: ## Restore PostgreSQL database (requires BACKUP_FILE variable)
	@if [ -z "$(BACKUP_FILE)" ]; then echo "Usage: make restore-db BACKUP_FILE=backup.sql"; exit 1; fi
	docker-compose -f $(DOCKER_COMPOSE_FILE) exec -T postgres psql -U adsuser adsmetrics < $(BACKUP_FILE)

tidy:
	@go mod tidy

remove:
	@docker stop $$(docker ps -q) 2>/dev/null || true
	@docker rm -f $$(docker ps -aq) 2>/dev/null || true

compose:
	@docker compose up --build