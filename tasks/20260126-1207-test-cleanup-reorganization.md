# PRD: Test Cleanup and Reorganization

## Introduction

The agent-smith repository currently has test files scattered at the root level that should be organized according to Go testing conventions. There are 7 test files at the root (6 active + 1 skipped), most of which are unit tests that should be colocated with the packages they test. This PRD outlines the systematic reorganization of these tests to improve code organization, maintainability, and follow Go best practices.

## Goals

- Move unit tests from root to their corresponding internal packages
- Organize integration tests in a dedicated tests/ directory
- Remove obsolete skipped test files
- Ensure all tests pass after reorganization
- Follow Go testing conventions (colocated tests in same package)
- Improve test discoverability and maintainability

## User Stories

- [x] Story-001: As a developer, I want unit tests colocated with the detector package so that detector functionality is easier to test and maintain.

  **Acceptance Criteria:**
  - `component_extraction_test.go` moved to `internal/detector/component_extraction_test.go`
  - `component_frontmatter_priority_test.go` moved to `internal/detector/component_frontmatter_priority_test.go`
  - `duplicate_warning_test.go` moved to `internal/detector/duplicate_warning_test.go`
  - All test imports updated to reference correct packages
  - Tests use `package detector` instead of `package main`
  - All tests pass with `go test ./internal/detector/...`
  
  **Testing Criteria:**
  **Unit Tests:**
  - Verify detector package tests run independently
  - Confirm test coverage remains the same or improves
  
  **Integration Tests:**
  - Ensure detector tests don't break integration test workflows
  
  **Component Browser Tests:**
  - N/A (backend Go tests)

- [x] Story-002: As a developer, I want unit tests colocated with the fileutil package so that file utility functions are easier to test and maintain.

  **Acceptance Criteria:**
  - `directory_copy_test.go` moved to `internal/fileutil/directory_copy_test.go`
  - `copy_component_files_test.go` moved to `internal/fileutil/copy_component_files_test.go`
  - All test imports updated to reference correct packages
  - Tests use `package fileutil` instead of `package main`
  - All tests pass with `go test ./internal/fileutil/...`
  
  **Testing Criteria:**
  **Unit Tests:**
  - Verify fileutil package tests run independently
  - Confirm file copying logic is thoroughly tested
  
  **Integration Tests:**
  - Ensure fileutil tests work with detector integration tests
  
  **Component Browser Tests:**
  - N/A (backend Go tests)

- [x] Story-003: As a developer, I want integration tests organized in a dedicated directory so I can easily distinguish them from unit tests.

  **Acceptance Criteria:**
  - Create `tests/` directory at repository root
  - Move `plugin_mirroring_integration_test.go` to `tests/plugin_mirroring_integration_test.go`
  - Test remains in `package main` since it tests the full application
  - Update test imports if needed
  - Test passes with `go test ./tests/...`
  - Add `tests/README.md` explaining that this directory contains integration tests
  
  **Testing Criteria:**
  **Unit Tests:**
  - N/A (this is integration test organization)
  
  **Integration Tests:**
  - Verify plugin mirroring integration test passes in new location
  - Confirm test can access main package functionality
  
  **Component Browser Tests:**
  - N/A (backend Go tests)

- [x] Story-004: As a developer, I want obsolete skipped test files removed so the repository stays clean and maintainable.

  **Acceptance Criteria:**
  - Delete `plugin_path_helpers_test.go.skip` from repository root
  - Verify file is not referenced anywhere in the codebase
  - Commit deletion with clear message explaining removal
  
  **Testing Criteria:**
  **Unit Tests:**
  - Verify no other tests depend on the skipped file
  
  **Integration Tests:**
  - Confirm removal doesn't break test workflows
  
  **Component Browser Tests:**
  - N/A (backend Go tests)

- [x] Story-005: As a developer, I want to verify all tests pass after reorganization so I can be confident nothing broke during the move.

  **Acceptance Criteria:**
  - Run `go test ./...` to execute all tests in repository
  - All tests pass without errors
  - Test coverage reports show same or better coverage
  - No import errors or missing dependencies
  - CI/CD pipelines (if any) continue to work
  
  **Testing Criteria:**
  **Unit Tests:**
  - All unit tests in internal/detector/ pass
  - All unit tests in internal/fileutil/ pass
  - All unit tests in internal/git/ pass (pre-existing)
  - All unit tests in pkg/paths/ pass (pre-existing)
  
  **Integration Tests:**
  - All integration tests in tests/ directory pass
  
  **Component Browser Tests:**
  - N/A (backend Go tests)

## Functional Requirements

- FR-1: The system must move detector-related tests to `internal/detector/` package
- FR-2: The system must move fileutil-related tests to `internal/fileutil/` package  
- FR-3: The system must move integration tests to `tests/` directory
- FR-4: The system must update all test package declarations from `package main` to appropriate package names
- FR-5: The system must update all test imports to reference correct package paths
- FR-6: The system must delete the skipped test file `plugin_path_helpers_test.go.skip`
- FR-7: The system must verify all tests pass after each move (sequential approach)
- FR-8: The system must create a `tests/README.md` documenting integration test organization

## Non-Goals (Out of Scope)

- Creating new unit tests for untested packages (detector, downloader, executor, etc.)
- Implementing test coverage reporting tools
- Setting up continuous integration test runners
- Refactoring test code or improving test quality
- Adding benchmark tests
- Creating test fixtures or mock frameworks
- Updating documentation beyond tests/README.md
- Performance optimization of existing tests

## Test File Mapping

### Current State (Root Level)
```
/
├── component_extraction_test.go (1,137 lines) → internal/detector/
├── component_frontmatter_priority_test.go (168 lines) → internal/detector/
├── duplicate_warning_test.go (381 lines) → internal/detector/
├── directory_copy_test.go (296 lines) → internal/fileutil/
├── copy_component_files_test.go (123 lines) → internal/fileutil/
├── plugin_mirroring_integration_test.go (740 lines) → tests/
└── plugin_path_helpers_test.go.skip → DELETE
```

### Target State
```
/
├── tests/
│   ├── README.md (NEW)
│   └── plugin_mirroring_integration_test.go (MOVED)
├── internal/
│   ├── detector/
│   │   ├── component_extraction_test.go (MOVED)
│   │   ├── component_frontmatter_priority_test.go (MOVED)
│   │   └── duplicate_warning_test.go (MOVED)
│   └── fileutil/
│       ├── directory_copy_test.go (MOVED)
│       └── copy_component_files_test.go (MOVED)
```

## Implementation Notes

### Sequential Approach Rationale
Tests will be moved one at a time with verification after each move to ensure:
1. Import paths are correctly updated
2. Package declarations are accurate
3. Tests still pass
4. No unexpected dependencies are broken

### Package Declaration Updates
When moving tests from root (`package main`) to internal packages:
- Detector tests: Change to `package detector`
- Fileutil tests: Change to `package fileutil`
- Integration tests: Keep as `package main` (tests main package)

### Import Path Updates
Tests may need import adjustments:
- Root tests can import `internal/*` packages directly
- Tests in internal packages may need to adjust relative imports
- Watch for circular import issues

### Success Criteria Summary
The reorganization is complete when:
1. All 5 unit test files are in their appropriate packages
2. 1 integration test is in `tests/` directory
3. 1 skipped test file is deleted
4. `go test ./...` passes completely
5. No files remain at root matching `*test*.go` pattern
