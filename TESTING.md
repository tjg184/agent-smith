# Testing Guide

This document describes the test organization and how to run different types of tests in the agent-smith project.

## Quick Reference

**Where are the tests located?**
- **Unit tests**: Co-located with source files (e.g., `internal/detector/*_test.go`)
- **Integration tests**: `tests/integration/` directory

**How do I run the tests?**
```bash
# Run unit tests only (fast, for development)
just test

# Run integration tests only
just test-integration

# Run all tests (unit + integration)
just test-all
```

**Using Go commands directly:**
```bash
# Unit tests only
go test ./...

# Integration tests only
go test -tags=integration ./tests/integration/...

# All tests
go test -tags=integration ./...
```

## Test Organization

### Unit Tests
Unit tests are co-located with source files following Go conventions:
- `*_test.go` files in the same directory as the code they test
- Test individual packages and functions in isolation
- Run with standard `go test` command
- Fast execution, suitable for frequent testing during development

**Packages with unit tests:**
- `internal/detector/`: Component detection, extraction, and pattern matching (6 test files)
- `internal/downloader/`: Download operations and error handling (2 test files)
- `internal/fileutil/`: File operations and utilities (2 test files)
- `internal/formatter/`: Output formatting (4 test files)
- `internal/git/`: Git operations and URL normalization (3 test files)
- `internal/linker/`: Component linking and profile collision handling (6 test files)
- `internal/materializer/`: File copy and sync-detection logic (1 test file)
- `internal/metadata/`: Lock file read/write operations (1 test file)
- `internal/testutil/`: Shared test utilities and helpers (1 test file)
- `internal/uninstaller/`: Component uninstall logic (1 test file)
- `internal/updater/`: Update operations and logic (1 test file)
- `pkg/colors/`: Terminal color helpers (1 test file)
- `pkg/config/`: Configuration and target management (6 test files)
- `pkg/errors/`: Error helpers (1 test file)
- `pkg/help/`: Help formatter (1 test file)
- `pkg/logger/`: Logging functionality (1 test file)
- `pkg/paths/`: Path utilities (1 test file)
- `pkg/profiles/`: Profile management, reuse, and activation (5 test files)
- `pkg/project/`: Project detection and materialization (2 test files)
- `pkg/services/find/`: Find service (1 test file)
- `pkg/services/lock/`: Lock service (1 test file)
- `pkg/services/materialize/`: Materialization post-processors (2 test files)
- `pkg/styles/`: UI styles (1 test file)

### Integration Tests
Integration tests verify end-to-end functionality and are distinguished by:
- Build tag `//go:build integration` at the top of the file
- Standard `_test.go` naming suffix
- Test complete workflows involving multiple components
- **Located in `tests/integration/` directory** for better organization

**Location:** All integration tests are located in the `tests/integration/` directory at the project root.

**Current integration tests (18 files):**
- `tests/integration/e2e_workflow_test.go`: End-to-end workflows (install → link → update → uninstall)
- `tests/integration/install_profile_switch_test.go`: Profile switching during install
- `tests/integration/link_auto_profile_test.go`: Auto-profile linking behavior
- `tests/integration/link_mixed_profiles_test.go`: Linking across multiple profiles
- `tests/integration/link_profile_collision_test.go`: Profile collision handling during link
- `tests/integration/link_status_default_behavior_test.go`: Link status command defaults
- `tests/integration/materialize_all_components_test.go`: Materializing all component types
- `tests/integration/materialize_commands_test.go`: Command component materialization
- `tests/integration/materialize_flat_agents_commands_test.go`: Flat agent/command materialization
- `tests/integration/materialize_flatten_copilot_test.go`: Copilot target flattening
- `tests/integration/materialize_nested_skills_test.go`: Nested skill materialization
- `tests/integration/profile_add_lock_preservation_test.go`: Profile addition and lock file preservation
- `tests/integration/profile_error_messages_test.go`: Profile command error message quality
- `tests/integration/profile_remove_lock_cleanup_test.go`: Lock file cleanup on profile removal
- `tests/integration/profile_rename_test.go`: Profile rename command behavior
- `tests/integration/profile_share_active_test.go`: Sharing components across profiles
- `tests/integration/uninstall_test.go`: Component uninstall workflows
- `tests/integration/unlink_test.go`: Component unlink workflows

## Running Tests

### Quick Start with justfile (Recommended)

The project includes a justfile that provides convenient commands for running different types of tests:

```bash
# Run unit tests only (fast, for development)
just test

# Run integration tests only
just test-integration

# Run all tests (unit + integration)
just test-all

# Run with verbose output
just test-verbose
just test-integration-verbose

# Run with coverage
just coverage
just coverage-integration

# See all available commands
just
```

**Why use the justfile?**
- **Faster development**: `just test` runs only fast unit tests, allowing rapid iteration
- **Clear separation**: Easy to run unit tests separately from slower integration tests
- **Convenience**: Shorter commands with sensible defaults
- **Consistency**: Same commands work for all team members

### Using Go Commands Directly

You can also run tests directly with Go commands:

#### Run Unit Tests Only (Default)
```bash
go test ./...
```

This runs all unit tests but skips integration tests due to build tags.

#### Run Integration Tests Only
```bash
go test -tags=integration ./tests/integration/...
```

This runs only the integration tests in the `tests/integration/` directory.

#### Run All Tests (Unit + Integration)
```bash
go test -tags=integration ./...
```

This runs both unit tests and integration tests across the entire codebase.

#### Run Tests in Specific Package
```bash
# Unit tests only
go test ./internal/detector

# With integration tests
go test -tags=integration ./internal/detector
```

#### Run Specific Test
```bash
# Unit test
go test -run TestComponentExtraction

# Integration test
go test -tags=integration -run TestPluginMirroringEndToEnd ./tests/integration/...
```

#### Run Tests with Coverage
```bash
# Unit tests
go test -cover ./...

# Integration tests
go test -tags=integration -cover ./tests/integration/...
```

#### Run Tests with Verbose Output
```bash
go test -v ./...
go test -tags=integration -v ./tests/integration/...
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
| Unit Tests | None | Co-located with source | 48 files | Test individual functions and packages |
| Integration Tests | `integration` | `tests/integration/` directory | 18 files | Test end-to-end workflows |

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
1. Create a file named `<feature>_test.go` in the `tests/integration/` directory
2. Add the build tag at the top:
   ```go
   //go:build integration
   
   package main
   ```
3. Write test functions starting with `Test`
4. Use `internal/testutil.NewTestHelper()` for creating test environments
5. Focus on critical user workflows and end-to-end scenarios
6. When building the agent-smith binary, set `cmd.Dir` to the repository root:
   ```go
   repoRoot := filepath.Join("..", "..")
   cmd := exec.Command("go", "build", "-o", binaryPath, ".")
   cmd.Dir = repoRoot
   ```

Example:
```go
//go:build integration

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
just test
# or: go test ./...

# Full test suite (suitable for PRs and releases)
just test-all
# or: go test -tags=integration ./...
```

### Recommended Workflow

**During Development:**
```bash
# Run fast unit tests frequently while coding
just test
```

**Before Committing:**
```bash
# Run all tests to ensure nothing is broken
just test-all
```

**In CI/CD:**
- Run `just test` on every push for fast feedback
- Run `just test-all` for pull requests and before merging
- Use `just coverage` to track test coverage metrics
