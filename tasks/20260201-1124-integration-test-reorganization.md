# PRD: Reorganize Integration Tests into /test/integration Directory

**Created**: 2026-02-01 11:24 UTC

---

## Introduction

Reorganize integration tests from the root directory into a dedicated `/test/integration` directory to improve project organization, reduce root clutter, and establish a scalable test structure for future growth.

The agent-smith project currently has 3 integration test files at the root level which creates clutter and doesn't scale well as more tests are added. Moving these to a dedicated directory follows Go best practices and makes the project structure clearer.

## Goals

- Move all integration tests to `/test/integration` directory structure
- Maintain all existing test functionality without changes to test logic
- Update documentation to reflect new structure and test execution patterns
- Ensure CI/CD compatibility with new test location
- Preserve build tags and test execution patterns
- Establish foundation for future test organization (fixtures, e2e, acceptance)

## User Stories

- [x] Story-001: As a developer, I want integration tests organized in a dedicated directory so that the root directory is less cluttered and tests are easier to find.

  **Acceptance Criteria:**
  - `/test/integration/` directory structure exists
  - All 3 integration test files moved from root to `/test/integration/`
  - Files renamed without `_integration` suffix (redundant with directory name)
  - `component_download_integration_test.go` → `test/integration/component_download_test.go`
  - `e2e_workflow_integration_test.go` → `test/integration/e2e_workflow_test.go`
  - `profile_add_lock_preservation_test.go` → `test/integration/profile_add_lock_preservation_test.go`
  - Build tags remain at top of each file unchanged
  - Package declaration remains `package main`
  - Import paths remain unchanged
  - Test logic remains completely unchanged
  
  **Testing Criteria:**
  **Integration Tests:**
  - All 3 integration tests run successfully from new location
  - `go test -tags=integration ./test/integration/...` executes all integration tests
  - Each test file can be run individually
  - Test output shows correct file paths

- [x] Story-002: As a developer, I want unit tests to remain separate from integration tests so that I can run fast unit tests during development without running slower integration tests.

  **Acceptance Criteria:**
  - `go test ./...` runs only unit tests (excludes integration tests)
  - `go test -tags=integration ./...` runs both unit and integration tests
  - `go test -tags=integration ./test/integration/...` runs only integration tests
  - Build tags continue to properly gate integration test execution
  - Unit test locations remain unchanged (co-located with source)
  - Test execution time for unit tests remains the same
  
  **Testing Criteria:**
  **Integration Tests:**
  - Verify `go test ./...` does not execute integration tests
  - Verify `go test -tags=integration ./...` executes all tests
  - Verify `go test -tags=integration ./test/integration/...` executes only integration tests
  - Verify specific test can be targeted with `-run` flag
  - Time execution to confirm unit-only runs are fast

- [x] Story-003: As a developer, I want updated documentation so that I know where integration tests are located and how to run them.

  **Acceptance Criteria:**
  - TESTING.md "Integration Tests" section updated with new location (line 27-37)
  - TESTING.md "Run Integration Tests" section updated with new commands (line 48-58)
  - TESTING.md "Adding an Integration Test" section updated with new path (line 165-199)
  - TESTING.md "Test Categories" table updated with new location (line 135-138)
  - All test command examples reference correct paths
  - Test count remains accurate (3 integration test files)
  - Example code shows correct directory structure
  
  **Testing Criteria:**
  **Integration Tests:**
  - Verify each documented command executes successfully
  - Verify file paths in documentation match actual file locations
  - Verify example code uses correct import paths

- [x] Story-004: As a developer, I want old test files removed from root so that there is no confusion about where integration tests are located.

  **Acceptance Criteria:**
  - `component_download_integration_test.go` deleted from root
  - `e2e_workflow_integration_test.go` deleted from root
  - `profile_add_lock_preservation_test.go` deleted from root
  - No integration test files remain in root directory
  - Only moved files in `/test/integration/` execute when tests run
  - Git history preserved for moved files
  
  **Testing Criteria:**
  **Integration Tests:**
  - Verify `ls *_integration_test.go` in root returns no results
  - Verify `ls profile_add_lock_preservation_test.go` in root returns no results
  - Verify all 3 tests run from new location
  - Final verification: `go test -tags=integration ./...` passes all tests

## Functional Requirements

- FR-1: The system SHALL create `/test/integration/` directory structure to house integration tests
- FR-2: The system SHALL move all 3 integration test files from root to `/test/integration/`
- FR-3: The system SHALL rename test files to remove `_integration` suffix where present
- FR-4: The system SHALL preserve all build tags, package declarations, and test logic during move
- FR-5: The system SHALL maintain proper build tag gating so `go test ./...` excludes integration tests
- FR-6: The system SHALL update TESTING.md documentation to reflect new structure
- FR-7: The system SHALL ensure all test execution commands work with new file locations
- FR-8: The system SHALL remove old test files from root after successful verification

## Non-Goals

- Modifying test logic or test cases
- Changing unit test locations (they remain co-located with source files)
- Creating new tests (only reorganizing existing tests)
- Modifying test helper utilities in `internal/testutil`
- Changing test naming conventions beyond removing redundant `_integration` suffix
- Adding new test categories (e2e, acceptance) at this time
- Modifying CI/CD pipeline configuration

## Technical Design

### Directory Structure

```
agent-smith/
├── cmd/
├── internal/
├── pkg/
├── test/
│   └── integration/
│       ├── component_download_test.go      (moved from root)
│       ├── e2e_workflow_test.go           (moved from root)
│       └── profile_add_lock_preservation_test.go (moved from root)
├── main.go
├── TESTING.md                              (updated)
└── go.mod
```

### Test File Format

Each test file maintains:

```go
//go:build integration
// +build integration

package main

import (
    "testing"
    // All existing imports remain the same
)

func TestFeature(t *testing.T) {
    // Test logic remains unchanged
}
```

### Test Execution Commands

```bash
# Unit tests only (default)
go test ./...

# All integration tests
go test -tags=integration ./test/integration/...

# All tests (unit + integration)
go test -tags=integration ./...

# Specific integration test
go test -tags=integration -run TestProfileAddPreservesLockFileEntries ./test/integration/

# With verbose output
go test -tags=integration -v ./test/integration/...

# With coverage
go test -tags=integration -cover ./test/integration/...
```

## Success Metrics

- All integration tests run successfully from new location: 3/3 passing
- Unit test execution excludes integration tests: Verified via `go test ./...`
- Documentation accurately reflects new structure: 4 sections updated in TESTING.md
- Zero test failures introduced by reorganization
- Root directory has 3 fewer files (reduced clutter)

## Risks and Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Import path issues | High | Low | Tests are in `package main` so imports should be transparent |
| Build tag not recognized | High | Low | Build tags remain unchanged; test during move |
| CI/CD pipeline breaks | Medium | Low | Test commands remain compatible; `./...` pattern works |
| Test fixtures not found | Medium | Low | Tests use absolute paths via testutil; verify during testing |
| Documentation out of sync | Low | Medium | Update docs in same commit; verify all examples |

## Future Enhancements

Once this structure is in place, consider:
- Add `/test/testdata/` for shared test fixtures
- Add `/test/e2e/` for separate end-to-end tests
- Add `/test/acceptance/` for acceptance tests
- Create integration-test-specific helper utilities
- Add test documentation in `/test/README.md`

## File Mappings

| Old Location | New Location |
|-------------|--------------|
| `/component_download_integration_test.go` | `/test/integration/component_download_test.go` |
| `/e2e_workflow_integration_test.go` | `/test/integration/e2e_workflow_test.go` |
| `/profile_add_lock_preservation_test.go` | `/test/integration/profile_add_lock_preservation_test.go` |

## References

- Go Testing Best Practices: https://go.dev/doc/tutorial/add-a-test
- Build Constraints: https://pkg.go.dev/cmd/go#hdr-Build_constraints
- Standard Go Project Layout: https://github.com/golang-standards/project-layout
