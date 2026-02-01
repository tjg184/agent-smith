# Testing Guide

This document describes the test organization and how to run different types of tests in the agent-smith project.

## Test Organization

### Unit Tests
Unit tests are co-located with source files following Go conventions:
- `*_test.go` files in the same directory as the code they test
- Test individual packages and functions in isolation
- Run with standard `go test` command
- Fast execution, suitable for frequent testing during development

**Packages with unit tests:**
- `internal/detector/`: Component detection, extraction, and pattern matching (6 test files)
- `internal/fileutil/`: File operations and utilities (2 test files)
- `internal/git/`: Git operations and URL normalization (2 test files)
- `internal/linker/`: Component linking and profile collision handling (4 test files)
- `internal/updater/`: Update operations and logic (1 test file)
- `internal/downloader/`: Download operations and error handling (1 test file)
- `internal/testutil/`: Shared test utilities and helpers (1 test file)
- `pkg/profiles/`: Profile management, reuse, and activation (5 test files)
- `pkg/config/`: Configuration and target management (4 test files)
- `pkg/paths/`: Path utilities (1 test file)
- `pkg/logger/`: Logging functionality (1 test file)

### Integration Tests
Integration tests verify end-to-end functionality and are distinguished by:
- Build tag `//go:build integration` at the top of the file
- Suffix `_integration_test.go` in the filename
- Test complete workflows involving multiple components
- Located at repository root for full application access

**Current integration tests:**
- `component_download_integration_test.go`: Component downloading, repository detection, cross-platform paths
- `e2e_workflow_integration_test.go`: End-to-end workflows (install → link → update → uninstall)
- `profile_add_lock_preservation_test.go`: Profile addition and lock file preservation

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

## Test Utilities

### internal/testutil Package
The `internal/testutil` package provides shared utilities for writing tests:

**TestHelper**: Main test helper for creating isolated test environments
```go
import "github.com/yourusername/agent-smith/internal/testutil"

func TestYourFeature(t *testing.T) {
    helper := testutil.NewTestHelper(t)
    defer helper.Cleanup()
    
    // Create mock repositories
    repoPath := helper.CreateMockRepo(testutil.MockRepoOptions{
        Structure: testutil.StructureGrouped,
        Components: []testutil.MockComponent{
            {Type: "skill", Name: "test-skill", Content: "test content"},
        },
    })
    
    // Your test logic here
}
```

**Key TestHelper Methods:**
- `NewTestHelper(t)`: Creates a new test helper with temporary directory
- `Cleanup()`: Removes all temporary test files and directories
- `CreateMockRepo()`: Creates a mock Git repository with components
- `CreatePluginRepo()`: Creates a repository with plugin-style structure
- `CreateFlatRepo()`: Creates a flat structure repository
- `CreateMonorepo()`: Creates a monorepo-style repository
- `CreateInstallDir()`: Creates a mock install directory

**Verification Helpers:**
- `VerifyFileExists(path)`: Checks if a file exists
- `VerifyDirExists(path)`: Checks if a directory exists
- `VerifyFileContent(path, expected)`: Verifies file contains expected content
- `CountFilesInDir(path)`: Counts files in a directory

## Test Categories

| Category | Build Tag | Location | Test Count | Purpose |
|----------|-----------|----------|------------|---------|
| Unit Tests | None | Co-located with source | 29 files | Test individual functions and packages |
| Integration Tests | `integration` | Root level `*_integration_test.go` | 3 files | Test end-to-end workflows |

## Adding New Tests

### Adding a Unit Test
1. Create a file named `<source>_test.go` in the same directory as the source file
2. Use `package <name>` (same as source) or `package <name>_test` for black-box testing
3. Write test functions starting with `Test`
4. Use `internal/testutil` helpers when creating test fixtures or mock repositories

Example:
```go
package mypackage

import (
    "testing"
    "github.com/yourusername/agent-smith/internal/testutil"
)

func TestMyFunction(t *testing.T) {
    helper := testutil.NewTestHelper(t)
    defer helper.Cleanup()
    
    // Test implementation
}
```

### Adding an Integration Test
1. Create a file named `<feature>_integration_test.go` in the root directory
2. Add build tags at the top:
   ```go
   //go:build integration
   // +build integration
   
   package main
   ```
3. Write test functions starting with `Test`
4. Use `internal/testutil.NewTestHelper()` for creating test environments
5. Focus on critical user workflows and end-to-end scenarios

Example:
```go
//go:build integration
// +build integration

package main

import (
    "testing"
    "github.com/yourusername/agent-smith/internal/testutil"
)

func TestEndToEndWorkflow(t *testing.T) {
    helper := testutil.NewTestHelper(t)
    defer helper.Cleanup()
    
    // Create mock repo
    repoPath := helper.CreateMockRepo(testutil.MockRepoOptions{...})
    
    // Test complete workflow
}
```

## Best Practices

### Unit Tests
- Keep tests focused on a single function or behavior
- Use table-driven tests for multiple test cases
- Mock external dependencies (Git, file system where appropriate)
- Use `internal/testutil` for common test setup
- Prefer package-level tests over integration tests for business logic

### Integration Tests
- Focus on critical user workflows
- Test end-to-end scenarios that cross package boundaries
- Use descriptive test names that explain the workflow
- Clean up test artifacts with `defer helper.Cleanup()`
- Keep integration tests maintainable - avoid testing every edge case

### Test Coverage
The project aims to maintain high test coverage through focused unit tests:
- Business logic: Tested at package level with unit tests
- Integration workflows: Tested with focused integration tests
- Use `go test -cover ./...` to monitor coverage

## CI/CD Integration

To run tests in CI/CD pipelines:

```bash
# Fast unit tests (suitable for every commit)
go test ./...

# Full test suite (suitable for PRs and releases)
go test -tags=integration ./...
```
