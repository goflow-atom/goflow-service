# GoFlow Workflow Engine - Makefile
# This Makefile provides common development tasks for the GoFlow project

.PHONY: help build run test clean wire generate fmt lint vet tidy install-tools

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME=goflow-server
BINARY_WINDOWS=$(BINARY_NAME).exe
CMD_DIR=./cmd/server
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Help target - displays available commands
help: ## Display this help message
	@echo "GoFlow Workflow Engine - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""

# Build targets
build: ## Build the server binary
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_WINDOWS) $(CMD_DIR)
	@echo "Build complete: $(BINARY_WINDOWS)"

build-linux: ## Build the server binary for Linux
	@echo "Building $(BINARY_NAME) for Linux..."
	@GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME) $(CMD_DIR)
	@echo "Build complete: $(BINARY_NAME)"

# Run targets
run: ## Run the server
	@echo "Starting GoFlow server..."
	@go run $(CMD_DIR)

# Test targets
test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=$(COVERAGE_FILE) ./...
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

test-middleware: ## Run middleware tests only
	@echo "Running middleware tests..."
	@go test -v ./internal/api/middleware/...

test-short: ## Run short tests (skip integration tests)
	@echo "Running short tests..."
	@go test -v -short ./...

# Wire dependency injection
wire: ## Generate Wire dependency injection code
	@echo "Generating Wire code..."
	@cd $(CMD_DIR) && wire
	@echo "Wire generation complete"

generate: wire ## Alias for wire target
	@echo "Code generation complete"

# Code quality targets
fmt: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Formatting complete"

lint: ## Run golangci-lint
	@echo "Running linter..."
	@golangci-lint run ./...
	@echo "Linting complete"

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...
	@echo "Vet complete"

# Dependency management
tidy: ## Tidy and verify dependencies
	@echo "Tidying dependencies..."
	@go mod tidy
	@go mod verify
	@echo "Dependencies tidied"

vendor: ## Vendor dependencies
	@echo "Vendoring dependencies..."
	@go mod vendor
	@echo "Vendoring complete"

# Installation targets
install-tools: ## Install development tools (wire, golangci-lint)
	@echo "Installing development tools..."
	@go install github.com/google/wire/cmd/wire@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed"

# Clean targets
clean: ## Remove build artifacts and generated files
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME) $(BINARY_WINDOWS)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@echo "Clean complete"

clean-all: clean ## Remove all generated files including vendor
	@echo "Removing vendor directory..."
	@rm -rf vendor/
	@echo "Clean all complete"

# Development workflow
dev: fmt vet test ## Run development checks (format, vet, test)
	@echo "Development checks complete"

ci: fmt vet test-coverage ## Run CI checks (format, vet, test with coverage)
	@echo "CI checks complete"

# Docker targets (placeholder for future implementation)
docker-build: ## Build Docker image
	@echo "Docker build not yet implemented"

docker-run: ## Run Docker container
	@echo "Docker run not yet implemented"

# Database migration targets (placeholder for future implementation)
migrate-up: ## Run database migrations up
	@echo "Database migrations not yet implemented"

migrate-down: ## Rollback database migrations
	@echo "Database migrations not yet implemented"
