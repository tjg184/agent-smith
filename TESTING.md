# Testing Guide

This document describes the test organization and how to run different types of tests in the agent-smith project.

## Test Organization

### Unit Tests
Unit tests are co-located with source files following Go conventions:
- `*_test.go` files in the same directory as the code they test
- Test the `main` package and internal packages
- Run with standard `go test` command

**Location:**
- Root level: `component_extraction_test.go`, `duplicate_warning_test.go`, etc.
- `internal/detector/`: Detector, component, and pattern matching tests
- `internal/fileutil/`: File utility tests
- `internal/git/`: Git operations tests
- `pkg/paths/`: Path utility tests

### Integration Tests
Integration tests verify end-to-end functionality and are distinguished by:
- Build tag `//go:build integration` at the top of the file
- Suffix `_integration_test.go` in the filename
- Test complex workflows involving multiple components

**Location:**
- Root level: `plugin_mirroring_integration_test.go`

## Running Tests

### Run Unit Tests Only (Default)
```bash
go test ./...
```

This runs all unit tests but skips integration tests due to build tags.

### Run Integration Tests
```bash
go test -tags=integration ./...
```

This runs both unit tests and integration tests.

### Run All Tests
```bash
go test -tags=integration ./...
```

### Run Tests in Specific Package
```bash
# Unit tests only
go test ./internal/detector

# With integration tests
go test -tags=integration ./internal/detector
```

### Run Specific Test
```bash
# Unit test
go test -run TestComponentExtraction

# Integration test
go test -tags=integration -run TestPluginMirroringEndToEnd
```

### Run Tests with Coverage
```bash
# Unit tests
go test -cover ./...

# Integration tests
go test -tags=integration -cover ./...
```

### Run Tests with Verbose Output
```bash
go test -v ./...
go test -tags=integration -v ./...
```

## Test Categories

| Category | Build Tag | Location | Purpose |
|----------|-----------|----------|---------|
| Unit Tests | None | Co-located with source | Test individual functions and components |
| Integration Tests | `integration` | Root level `*_integration_test.go` | Test end-to-end workflows |

## Adding New Tests

### Adding a Unit Test
1. Create a file named `<source>_test.go` in the same directory as the source file
2. Use `package <name>` (same as source) or `package <name>_test` for black-box testing
3. Write test functions starting with `Test`

### Adding an Integration Test
1. Create a file named `<feature>_integration_test.go` in the root directory
2. Add build tags at the top:
   ```go
   //go:build integration
   // +build integration
   
   package main
   ```
3. Write test functions starting with `Test`
4. Use the `TestHelper` utility for creating mock repositories

## CI/CD Integration

To run tests in CI/CD pipelines:

```bash
# Fast unit tests (suitable for every commit)
go test ./...

# Full test suite (suitable for PRs and releases)
go test -tags=integration ./...
```
