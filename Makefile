.PHONY: help run dev test lint build migrate generate clean install-tools docker-build docker-up docker-down

# Variables
APP_NAME := go-fiber
BINARY_NAME := $(APP_NAME)
DOCKER_IMAGE := $(APP_NAME):latest
POSTGRES_URL := postgresql://postgres:password@localhost:5432/go_fiber?sslmode=disable

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development commands
run: ## Run the application in development mode
	@echo "Starting $(APP_NAME)..."
	@go run main.go

dev: ## Run the application with hot reload (requires air)
	@echo "Starting $(APP_NAME) with hot reload..."
	@if command -v air > /dev/null 2>&1; then \
		air; \
	else \
		echo "Air not found. Installing air..."; \
		go install github.com/air-verse/air@latest; \
		air; \
	fi

test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

lint-fix: ## Run linter with auto-fix
	@echo "Running linter with auto-fix..."
	@golangci-lint run --fix

# Build commands
build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@go build -o bin/$(BINARY_NAME) cmd/$(APP_NAME)/main.go

build-linux: ## Build for Linux
	@echo "Building $(APP_NAME) for Linux..."
	@GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)-linux cmd/$(APP_NAME)/main.go

build-windows: ## Build for Windows
	@echo "Building $(APP_NAME) for Windows..."
	@GOOS=windows GOARCH=amd64 go build -o bin/$(BINARY_NAME)-windows.exe cmd/$(APP_NAME)/main.go

build-mac: ## Build for macOS
	@echo "Building $(APP_NAME) for macOS..."
	@GOOS=darwin GOARCH=amd64 go build -o bin/$(BINARY_NAME)-mac cmd/$(APP_NAME)/main.go

# Database commands
migrate: ## Run database migrations
	@echo "Running database migrations..."
	@goose -dir migrations/postgres postgres "$(POSTGRES_URL)" up

migrate-down: ## Rollback database migrations
	@echo "Rolling back database migrations..."
	@goose -dir migrations/postgres postgres "$(POSTGRES_URL)" down

migrate-status: ## Check migration status
	@echo "Checking migration status..."
	@goose -dir migrations/postgres postgres "$(POSTGRES_URL)" status

migrate-create: ## Create a new migration (usage: make migrate-create NAME=migration_name)
	@echo "Creating new migration: $(NAME)"
	@goose -dir migrations/postgres create $(NAME) sql

# Code generation commands
generate: ## Generate code (SQLC, mocks, etc.)
	@echo "Generating code..."
	@sqlc generate
	@go generate ./...

generate-mocks: ## Generate mock files
	@echo "Generating mocks..."
	@mockgen -source=internal/repository/user.go -destination=internal/repository/mocks/user_mock.go
	@mockgen -source=internal/repository/todo.go -destination=internal/repository/mocks/todo_mock.go

# Dependency management
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

deps-verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	@go mod verify

# Development tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@go install github.com/golang/mock/mockgen@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/air-verse/air@latest

# Docker commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -f docker/Dockerfile -t $(DOCKER_IMAGE) .

docker-up: ## Start services with Docker Compose
	@echo "Starting services with Docker Compose..."
	@docker-compose up -d

docker-down: ## Stop services with Docker Compose
	@echo "Stopping services with Docker Compose..."
	@docker-compose down

docker-logs: ## Show Docker Compose logs
	@docker-compose logs -f

docker-clean: ## Clean Docker images and containers
	@echo "Cleaning Docker images and containers..."
	@docker-compose down -v
	@docker system prune -f

# Documentation
docs: ## Generate API documentation
	@echo "Generating API documentation..."
	@swag init -g cmd/$(APP_NAME)/main.go -o docs

docs-serve: ## Serve documentation locally
	@echo "Serving documentation at http://localhost:9000/swagger/"
	@make run

# Cleanup
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf coverage.out coverage.html
	@rm -rf docs/swagger.json docs/swagger.yaml

clean-all: clean ## Clean everything including dependencies
	@echo "Cleaning everything..."
	@go clean -modcache

# Development workflow
dev-setup: install-tools deps ## Setup development environment
	@echo "Setting up development environment..."
	@cp .env.example .env
	@echo "Development environment setup complete!"
	@echo "Please edit .env file with your configuration"

dev-reset: clean deps migrate generate ## Reset development environment
	@echo "Resetting development environment..."

# Production
prod-build: ## Build for production
	@echo "Building for production..."
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o bin/$(BINARY_NAME) cmd/$(APP_NAME)/main.go

# Health check
health: ## Check application health
	@echo "Checking application health..."
	@curl -f http://localhost:9000/health || echo "Application is not running"

# Quick commands for common workflows
quick-start: dev-setup docker-up migrate generate dev ## Quick start for new developers with hot reload

quick-test: lint test ## Quick test (lint + test)

quick-deploy: clean build docker-build ## Quick deploy preparation