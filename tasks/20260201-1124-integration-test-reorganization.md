# PRD: Reorganize Integration Tests into /test/integration Directory

## Overview
Reorganize integration tests from the root directory into a dedicated `/test/integration` directory to improve project organization, reduce root clutter, and establish a scalable test structure.

## Problem Statement
The agent-smith project currently has 3 integration test files at the root level:
- `component_download_integration_test.go`
- `e2e_workflow_integration_test.go`
- `profile_add_lock_preservation_test.go`

This creates several issues:
1. **Root directory clutter**: 25+ items in root makes navigation harder
2. **Scalability**: As more integration tests are added, root becomes unwieldy
3. **Test organization**: No clear separation between production code and integration tests
4. **Fixture management**: No dedicated location for shared test data/fixtures

## Goals
1. Move all integration tests to `/test/integration` directory
2. Maintain all existing test functionality without changes to test logic
3. Update documentation to reflect new structure
4. Ensure CI/CD compatibility with new structure
5. Preserve build tags and test execution patterns

## Non-Goals
- Modifying test logic or test cases
- Changing unit test locations (they remain co-located with source)
- Creating new tests (only reorganizing existing ones)
- Modifying test helper utilities in `internal/testutil`

## Success Criteria
1. All integration tests run successfully from new location
2. `go test ./...` excludes integration tests (unit tests only)
3. `go test -tags=integration ./...` runs all tests including integration
4. `go test -tags=integration ./test/integration/...` runs only integration tests
5. TESTING.md documentation accurately reflects new structure
6. No test failures introduced by the move

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

### File Changes

#### 1. Create Directory Structure
- Create `/test/integration/` directory

#### 2. Move Integration Test Files
Move and rename the following files:
- `component_download_integration_test.go` → `/test/integration/component_download_test.go`
- `e2e_workflow_integration_test.go` → `/test/integration/e2e_workflow_test.go`
- `profile_add_lock_preservation_test.go` → `/test/integration/profile_add_lock_preservation_test.go`

**Note**: Remove `_integration` suffix from filenames since the directory name makes this redundant.

#### 3. Update File Contents
Each moved test file requires these changes:

**Build tags** (keep as-is):
```go
//go:build integration
// +build integration
```

**Package declaration** (keep as-is):
```go
package main
```

**Import paths** (verify, likely no changes needed):
```go
import (
    "testing"
    // All existing imports remain the same
)
```

**Test code** (no changes needed):
- Test logic remains identical
- Helper functions remain identical

#### 4. Update Documentation

**TESTING.md** - Update sections:

**Line 27-37** (Integration Tests section):
```markdown
### Integration Tests
Integration tests verify end-to-end functionality and are distinguished by:
- Build tag `//go:build integration` at the top of the file
- Located in `/test/integration/` directory
- Test complete workflows involving multiple components

**Current integration tests:**
- `test/integration/component_download_test.go`: Component downloading, repository detection, cross-platform paths
- `test/integration/e2e_workflow_test.go`: End-to-end workflows (install → link → update → uninstall)
- `test/integration/profile_add_lock_preservation_test.go`: Profile addition and lock file preservation
```

**Line 48-58** (Running Tests section):
```markdown
### Run Integration Tests
```bash
# All integration tests
go test -tags=integration ./test/integration/...

# Or run all tests (unit + integration) across entire project
go test -tags=integration ./...
```

### Run Specific Integration Test
```bash
go test -tags=integration -run TestPluginMirroringEndToEnd ./test/integration/
```
```

**Line 165-199** (Adding an Integration Test section):
```markdown
### Adding an Integration Test
1. Create a file named `<feature>_test.go` in `/test/integration/`
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
    "github.com/tgaines/agent-smith/internal/testutil"
)

func TestEndToEndWorkflow(t *testing.T) {
    helper := testutil.NewTestHelper(t)
    defer helper.Cleanup()
    
    // Create mock repo
    repoPath := helper.CreateMockRepo(testutil.MockRepoOptions{...})
    
    // Test complete workflow
}
```
```

**Line 135-138** (Test Categories table):
```markdown
| Category | Build Tag | Location | Test Count | Purpose |
|----------|-----------|----------|------------|---------|
| Unit Tests | None | Co-located with source | 29 files | Test individual functions and packages |
| Integration Tests | `integration` | `/test/integration/` | 3 files | Test end-to-end workflows |
```

### Test Execution Commands

All existing test commands will continue to work:

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

## Implementation Plan

### Phase 1: Directory Setup
**Task 1.1**: Create `/test/integration/` directory structure
- Create `/test` directory
- Create `/test/integration` subdirectory

### Phase 2: Move Test Files
**Task 2.1**: Move `component_download_integration_test.go`
- Move to `/test/integration/component_download_test.go`
- Verify file contents (no changes needed)

**Task 2.2**: Move `e2e_workflow_integration_test.go`
- Move to `/test/integration/e2e_workflow_test.go`
- Verify file contents (no changes needed)

**Task 2.3**: Move `profile_add_lock_preservation_test.go`
- Move to `/test/integration/profile_add_lock_preservation_test.go`
- Verify file contents (no changes needed)

### Phase 3: Verification
**Task 3.1**: Run unit tests
- Execute: `go test ./...`
- Verify: Only unit tests run, integration tests excluded

**Task 3.2**: Run integration tests
- Execute: `go test -tags=integration ./test/integration/...`
- Verify: All 3 integration tests pass

**Task 3.3**: Run all tests
- Execute: `go test -tags=integration ./...`
- Verify: Both unit and integration tests pass

**Task 3.4**: Run specific integration test
- Execute: `go test -tags=integration -run TestProfileAddPreservesLockFileEntries ./test/integration/`
- Verify: Specific test runs and passes

### Phase 4: Documentation
**Task 4.1**: Update TESTING.md
- Update "Integration Tests" section (line 27-37)
- Update "Run Integration Tests" section (line 48-58)
- Update "Adding an Integration Test" section (line 165-199)
- Update "Test Categories" table (line 135-138)

**Task 4.2**: Verify documentation accuracy
- Review all test command examples
- Ensure file paths are correct
- Verify test counts are accurate

### Phase 5: Cleanup
**Task 5.1**: Remove old test files from root
- Delete `component_download_integration_test.go` from root
- Delete `e2e_workflow_integration_test.go` from root
- Delete `profile_add_lock_preservation_test.go` from root

**Task 5.2**: Final verification
- Execute: `go test -tags=integration ./...`
- Verify: All tests pass with new structure

## Testing Strategy

### Pre-implementation Verification
1. Run all tests to establish baseline: `go test -tags=integration ./...`
2. Document current test count and pass/fail status

### During Implementation
1. After each file move, verify it runs correctly from new location
2. Keep old files in place until new location is verified
3. Run tests incrementally to catch issues early

### Post-implementation Verification
1. Run unit tests only: `go test ./...`
2. Run integration tests only: `go test -tags=integration ./test/integration/...`
3. Run all tests: `go test -tags=integration ./...`
4. Verify each specific test can run independently
5. Confirm test coverage reports work correctly

## Risks and Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Import path issues | High | Low | Verify imports work from new location; should be transparent |
| Build tag not recognized | High | Low | Verify build tags remain at top of files; test execution validates |
| CI/CD pipeline breaks | Medium | Low | Document new test commands; existing commands should still work |
| Test fixtures not found | Medium | Low | Tests use relative or absolute paths; verify during testing phase |
| Documentation out of sync | Low | Medium | Update docs in same commit; include in PR review checklist |

## Future Enhancements
Once this structure is in place, consider:
1. Add `/test/testdata/` for shared test fixtures
2. Add `/test/e2e/` for separate end-to-end tests if needed
3. Add `/test/acceptance/` for acceptance tests
4. Create test helper utilities specific to integration tests

## Metrics
- **Files moved**: 3
- **Directories created**: 2 (`/test`, `/test/integration`)
- **Documentation files updated**: 1 (TESTING.md)
- **Test execution time**: Should remain unchanged
- **Test coverage**: Should remain unchanged

## Appendix

### File Mappings
| Old Location | New Location |
|-------------|--------------|
| `/component_download_integration_test.go` | `/test/integration/component_download_test.go` |
| `/e2e_workflow_integration_test.go` | `/test/integration/e2e_workflow_test.go` |
| `/profile_add_lock_preservation_test.go` | `/test/integration/profile_add_lock_preservation_test.go` |

### References
- Go Testing Best Practices: https://go.dev/doc/tutorial/add-a-test
- Build Constraints: https://pkg.go.dev/cmd/go#hdr-Build_constraints
- Test Organization Patterns: Standard Go project layout conventions
