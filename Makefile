# EVM TVL Aggregator Makefile

.PHONY: help setup build test clean docker-up docker-down run-api run-indexer run-tui

# Default target
help: ## Show available commands
	@echo "EVM TVL Aggregator - Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Development setup
setup: ## Install dependencies and setup environment
	@echo "Setting up development environment..."
	go mod tidy
	@if [ ! -f .env ]; then cp .env.example .env; echo "Created .env from .env.example"; fi
	@echo "Development environment ready"

# Build targets
build: ## Build all services
	@echo "Building all services..."
	mkdir -p bin
	go build -o bin/api ./cmd/api
	go build -o bin/indexer ./cmd/indexer
	go build -o bin/aggregator ./cmd/aggregator
	go build -o bin/tui ./cmd/tui
	@echo "Build completed"

build-api: ## Build API server only
	@echo "Building API server..."
	mkdir -p bin
	go build -o bin/api ./cmd/api

build-tui: ## Build terminal interface only
	@echo "Building terminal interface..."
	mkdir -p bin
	go build -o bin/tui ./cmd/tui

# Run commands
run-api: ## Run API server
	@echo "Starting API server..."
	go run cmd/api/main.go

run-indexer: ## Run blockchain indexer
	@echo "Starting indexer..."
	go run cmd/indexer/main.go

run-aggregator: ## Run TVL aggregator
	@echo "Starting aggregator..."
	go run cmd/aggregator/main.go

run-tui: ## Run terminal interface
	@echo "Starting terminal interface..."
	go run cmd/tui/main.go

# Docker commands
docker-up: ## Start services with Docker Compose
	@echo "Starting Docker services..."
	docker compose up -d
	@echo "Services started"
	@echo "PostgreSQL: localhost:5432"
	@echo "Redis: localhost:6379"

docker-down: ## Stop Docker services
	@echo "Stopping Docker services..."
	docker compose down
	@echo "Services stopped"

docker-logs: ## View Docker logs
	docker compose logs -f

docker-clean: ## Remove containers and volumes
	@echo "Cleaning Docker resources..."
	docker compose down -v --remove-orphans
	docker system prune -f
	@echo "Cleanup completed"

# Database commands
test-db: ## Test database connection
	@echo "Testing database connection..."
	@if [ ! -f .env ]; then echo "Error: .env file not found. Run 'make setup' first."; exit 1; fi
	go run scripts/test-db-connection.go

db-shell: ## Connect to PostgreSQL shell
	@echo "Connecting to PostgreSQL..."
	docker exec -it tvl-postgres psql -U postgres -d tvl_aggregator

# Testing
test: ## Run all tests
	@echo "Running tests..."
	go test ./...

test-verbose: ## Run tests with verbose output
	@echo "Running tests (verbose)..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Code quality
lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

# Cleanup
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache
	@echo "Cleanup completed"

# Development workflow
dev: setup build ## Setup and build for development
	@echo "Development environment ready"

# Production build
prod-build: ## Build for production
	@echo "Building for production..."
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/api ./cmd/api
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/indexer ./cmd/indexer
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/aggregator ./cmd/aggregator
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/tui ./cmd/tui
	@echo "Production build completed"

# Install tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed"

# All-in-one commands
all: clean setup build test ## Clean, setup, build and test
	@echo "All tasks completed"

demo: ## Run terminal demo
	@echo "Starting terminal demo..."
	./run_terminal.sh