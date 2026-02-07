# List available recipes
default:
    @just --list

# Run unit tests only (fast, for development)
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

# Build and install agent-smith to $GOPATH/bin
install:
    @echo "Installing agent-smith..."
    @go install .
    @echo "✓ agent-smith installed to $(go env GOPATH)/bin/agent-smith"

# Clean build artifacts and test files
clean:
    @echo "Cleaning build artifacts..."
    @rm -f agent-smith
    @rm -f coverage-*.txt
    @go clean -testcache

# Format code
fmt:
    @echo "Formatting all Go files..."
    @go fmt ./...

# Vet code
vet:
    @echo "Running go vet..."
    @go vet ./...

# Lint code
lint:
    @echo "Running golangci-lint..."
    @golangci-lint run || echo "Install with 'brew install golangci-lint' or 'go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest'"

# Tidy modules
tidy:
    @echo "Tidying Go modules..."
    @go mod tidy
