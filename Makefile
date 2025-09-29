.PHONY: build run test clean docker-build docker-run docker-stop migrate

# Build the application
build:
	go build -o bin/task-scheduler ./cmd/server

# Run the application locally
run:
	go run ./cmd/server

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

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
