.PHONY: help test test-unit test-integration test-all test-verbose test-unit-verbose test-integration-verbose coverage coverage-unit coverage-integration build clean

# Default target
help:
	@echo "Available targets:"
	@echo "  make test              - Run unit tests only (fast, for development)"
	@echo "  make test-unit         - Run unit tests only (alias for 'test')"
	@echo "  make test-integration  - Run integration tests only"
	@echo "  make test-all          - Run all tests (unit + integration)"
	@echo "  make test-verbose      - Run unit tests with verbose output"
	@echo "  make test-unit-verbose - Run unit tests with verbose output"
	@echo "  make test-integration-verbose - Run integration tests with verbose output"
	@echo "  make coverage          - Run unit tests with coverage report"
	@echo "  make coverage-unit     - Run unit tests with coverage report"
	@echo "  make coverage-integration - Run integration tests with coverage report"
	@echo "  make build             - Build the agent-smith binary"
	@echo "  make clean             - Remove built binary and test artifacts"

# Run unit tests only (default for development)
test: test-unit

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	@go test ./...

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	@go test -tags=integration ./tests/integration/...

# Run all tests (unit + integration)
test-all:
	@echo "Running all tests (unit + integration)..."
	@go test -tags=integration ./...

# Run unit tests with verbose output
test-verbose: test-unit-verbose

# Run unit tests with verbose output
test-unit-verbose:
	@echo "Running unit tests with verbose output..."
	@go test -v ./...

# Run integration tests with verbose output
test-integration-verbose:
	@echo "Running integration tests with verbose output..."
	@go test -tags=integration -v ./tests/integration/...

# Run unit tests with coverage
coverage: coverage-unit

# Run unit tests with coverage
coverage-unit:
	@echo "Running unit tests with coverage..."
	@go test -cover ./...

# Run integration tests with coverage
coverage-integration:
	@echo "Running integration tests with coverage..."
	@go test -tags=integration -cover ./tests/integration/...

# Build the agent-smith binary
build:
	@echo "Building agent-smith..."
	@go build -o agent-smith .

# Clean build artifacts and test files
clean:
	@echo "Cleaning build artifacts..."
	@rm -f agent-smith
	@rm -f coverage-*.txt
	@go clean -testcache
