.PHONY: build run test clean docker-build docker-run docker-stop migrate test-unit test-integration test-e2e test-coverage test-race help

# Help target
help: ## Display available make targets
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build the application
build: ## Build the application binary
	go build -o bin/task-scheduler ./cmd/server

# Run the application locally
run: ## Run the application locally
	go run ./cmd/server

# Run tests
test: ## Run all tests (unit, integration, e2e)
	go test -v ./...

# Run unit tests only
test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	go test -v -timeout 5m ./tests/unit/...

# Run integration tests only
test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	go test -v -timeout 10m ./tests/integration/...

# Run end-to-end tests only
test-e2e: ## Run end-to-end tests only
	@echo "Running end-to-end tests..."
	go test -v -timeout 10m ./tests/e2e/...

# Run tests with coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -v -timeout 10m -coverprofile=coverage.out ./tests/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detection
test-race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	go test -v -timeout 10m -race ./tests/...

# Run specific test categories
test-models: ## Run model tests only
	go test -v -timeout 2m ./tests/unit/models/...

test-repository: ## Run repository tests only
	go test -v -timeout 5m ./tests/unit/repository/...

test-executor: ## Run executor tests only
	go test -v -timeout 5m ./tests/unit/executor/...

test-api: ## Run API tests only
	go test -v -timeout 5m ./tests/integration/api/...

test-workflow: ## Run workflow tests only
	go test -v -timeout 10m ./tests/e2e/...

# Run tests with different modes
test-short: ## Run tests with short flag (faster)
	go test -v -timeout 2m -short ./tests/...

test-parallel: ## Run tests in parallel
	go test -v -timeout 10m -parallel 4 ./tests/...

# Clean build artifacts
clean: ## Clean build artifacts and test cache
	rm -rf bin/
	go clean -testcache

# Docker commands
docker-build:
	docker-compose build

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

docker-logs:
	docker-compose logs -f task-scheduler

# Database migration (when running locally)
migrate:
	psql -h localhost -U postgres -d task_scheduler -f migrations/001_create_tasks_table.sql
	psql -h localhost -U postgres -d task_scheduler -f migrations/002_create_task_results_table.sql

# Development setup
dev-setup:
	cp .env.example .env
	docker-compose up -d postgres
	sleep 5
	make migrate

# Generate Swagger docs (requires swag CLI)
swagger:
	swag init -g cmd/server/main.go -o docs/

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Run with hot reload (requires air)
dev:
	air
