# PRD: Integration Test Cleanup & Consolidation

**Created**: 2026-02-01 01:07 UTC

---

## Introduction

The agent-smith project currently has 10 integration test files (3,608 lines) at the repository root. These tests were created incrementally as features were developed, resulting in:

- **Duplication**: Multiple story tests verify overlapping functionality
- **Scope creep**: Story-specific acceptance tests testing business logic that should be in unit tests
- **Maintenance burden**: Integration tests are slower and harder to maintain than focused unit tests
- **Unclear purpose**: Some tests are temporary verification suites, others test core functionality

This PRD outlines a comprehensive cleanup to:
1. **Extract** reusable test utilities to `internal/testutil/` package
2. **Migrate** business logic tests from integration to focused unit tests
3. **Consolidate** integration tests to 3 focused end-to-end workflow tests
4. **Delete** 7 story-specific integration tests and temporary verification suites
5. **Verify** test coverage is maintained or improved through each phase

**Success metric**: Reduce integration test code by ~81% (from 3,608 to ~700 lines) while maintaining or improving coverage.

---

## Goals

- Reduce integration test lines of code from 3,608 to ~700 (81% reduction)
- Extract TestHelper and utilities to reusable `internal/testutil/` package
- Move business logic tests from CLI integration tests to focused package-level unit tests
- Consolidate to 2 clear-purpose integration tests: e2e workflows and component downloading
- Delete 9 integration tests (8 story tests + 1 debug flag test) and 1 feature demo test
- Maintain or improve test coverage (verified with go test -cover before/after)
- Execute as single cohesive PR suitable for Ralphy parallel execution

---

## User Stories

- [x] Story-001: As a developer, I want shared test utilities in internal/testutil so that I can write consistent tests across packages without duplication.

  **Acceptance Criteria:**
  - Create `internal/testutil/helpers.go` with TestHelper struct extracted from component_download_integration_test.go
  - TestHelper includes methods: NewTestHelper, Cleanup, CreateMockRepo, CreatePluginRepo, CreateFlatRepo, CreateMonorepo, CreateInstallDir
  - Create `internal/testutil/verification.go` with verification helpers: VerifyFileExists, VerifyDirExists, VerifyFileContent, CountFilesInDir
  - Create `internal/testutil/fixtures.go` with common test data: mock repo structures, metadata templates, lock file templates
  - All utilities properly documented with godoc comments
  - Package can be imported by both unit tests and integration tests
  
  **Testing Criteria:**
  **Unit Tests:**
  - TestHelper initialization and cleanup tests
  - Mock repo creation tests with various structures
  - Verification helper tests for file operations
  - Fixture generation tests
  
  **Integration Tests:**
  - Note: Will be validated by refactoring component_download_integration_test.go to use new package

- [x] Story-002: As a developer, I want profile management logic tested at the package level so that profile tests run faster and are easier to debug.

  **Acceptance Criteria:**
  - Add `pkg/profiles/reuse_test.go` covering profile reuse logic from Story-001 integration test
  - Tests cover: finding existing profile by normalized URL, updating existing profile metadata, preventing duplicate profiles for same source
  - Add `pkg/profiles/activation_test.go` covering profile activation/deactivation workflows
  - Tests use testutil helpers for creating mock profile structures
  - All tests can run without CLI invocation (pure package-level testing)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile reuse detection with various URL formats (HTTPS, SSH, shorthand)
  - Profile metadata creation and updates
  - Active profile tracking and state management
  - Profile directory structure validation
  
  **Integration Tests:**
  - Note: Core e2e profile workflows will be covered in Story-007 (e2e_workflow_integration_test.go)

- [x] Story-003: As a developer, I want Git URL normalization tested in internal/git so that URL handling tests are close to the implementation.

  **Acceptance Criteria:**
  - Add `internal/git/url_normalization_test.go` (or extend existing if present) covering Story-002 integration test cases
  - Tests cover: HTTPS URLs, SSH URLs (git@), SSH URLs (ssh://), HTTP → HTTPS upgrade, trailing slash removal, .git extension removal, case-insensitive domains, GitHub/GitLab/Bitbucket variations
  - Tests verify normalized output matches expected canonical format
  - All URL variations that should match resolve to same normalized form
  
  **Testing Criteria:**
  **Unit Tests:**
  - URL parsing for all common formats
  - Normalization rules applied correctly
  - Domain case-insensitivity
  - Path normalization (trailing slashes, .git extension)
  - Shorthand expansion (owner/repo → https://github.com/owner/repo)
  
  **Integration Tests:**
  - Note: Will verify URL normalization works end-to-end through profile reuse tests

- [x] Story-004: As a developer, I want updater logic tested in internal/updater so that update behavior is verified without full CLI integration.

  **Acceptance Criteria:**
  - Add `internal/updater/profile_update_test.go` covering Story-003, Story-006 update logic
  - Tests cover: detecting active profile, using specified profile via flag, falling back to base directory, providing location feedback messages
  - Add `internal/updater/base_directory_test.go` covering Story-004 base directory update logic
  - Tests cover: detecting components in base directory when no profile active, update summary for base directory components, backward compatibility for non-profile users
  - Tests use testutil helpers for mock directory structures
  - All update detection and messaging logic testable without CLI invocation
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile detection for update operations
  - Base directory fallback logic
  - Location message generation for different scenarios
  - Update summary generation
  - Component detection in profile vs base directory
  
  **Integration Tests:**
  - Note: End-to-end update workflows will be covered in Story-007

- [x] Story-005: As a developer, I want component downloading tests consolidated so that integration tests focus on real file system operations and cross-platform behavior.

  **Acceptance Criteria:**
  - Refactor `component_download_integration_test.go` to use `internal/testutil` helpers
  - Remove TestHelper struct and helper methods (now in testutil package)
  - Update all test functions to use testutil.NewTestHelper()
  - Verify tests still pass with same coverage
  - File size reduced from ~808 lines to ~400 lines (removing helper code)
  
  **Testing Criteria:**
  **Integration Tests:**
  - Grouped component download (plugins/ui-design/agents/)
  - Multiple components from same group
  - Flat repository structure backward compatibility
  - Monorepo structure support
  - Component linking via symlinks
  - Cross-platform path handling (Windows/Unix separators)
  - Error handling for missing components
  - Mixed structure detection (grouped + flat)
  - Git operations (commit hash tracking)
  - Skill not found error with available skills list
  
  **Unit Tests:**
  - Note: Component detection logic already covered in internal/detector/ unit tests

- [x] Story-006: As a developer, I want a single focused e2e workflow test so that critical user paths are verified without duplicating feature-level tests.

  **Acceptance Criteria:**
  - Create `e2e_workflow_integration_test.go` with 4 critical happy path workflows
  - Workflow 1: Install all → verify files → link all → verify symlinks → update all → verify updates
  - Workflow 2: Install single component → link single component → update single component
  - Workflow 3: Install with --profile flag → activate profile → verify active state → deactivate → verify inactive
  - Workflow 4: Install to custom --target-dir → verify isolation from ~/.agent-smith/
  - Each workflow is self-contained test function
  - Tests use testutil helpers for setup and verification
  - File size ~300 lines focused on end-to-end paths
  
  **Testing Criteria:**
  **Integration Tests:**
  - Full install → link → update → uninstall lifecycle
  - Profile-based workflows
  - Custom target directory isolation
  - Component type variations (skills, agents, commands)
  - Target detection and auto-linking
  
  **Unit Tests:**
  - Note: Individual command logic tested at package level, not duplicated here

- [ ] Story-007: As a developer, I want story-specific integration tests removed so that test suite is maintainable and focused on core functionality.

  **Acceptance Criteria:**
  - Delete `story_001_integration_test.go` (343 lines) - profile reuse now in pkg/profiles/reuse_test.go
  - Delete `story_002_integration_test.go` (413 lines) - URL normalization now in internal/git/url_normalization_test.go
  - Delete `story_003_update_single_component_test.go` (249 lines) - update logic now in internal/updater/profile_update_test.go
  - Delete `story_004_integration_test.go` (374 lines) - profile flag logic now in pkg/profiles/ and e2e_workflow_integration_test.go
  - Delete `story_004_update_base_directory_test.go` (203 lines) - base directory logic now in internal/updater/base_directory_test.go
  - Delete `story_005_verification_integration_test.go` (425 lines) - temporary acceptance test, functionality verified by other tests
  - Delete `story_005_feature_test.go` (211 lines) - feature demo for Story-005, profile workflows now in e2e_workflow_integration_test.go
  - Delete `story_006_update_location_feedback_test.go` (260 lines) - feedback logic now in internal/updater/profile_update_test.go
  - Total deletion: 2,478 lines of integration test code
  
  **Testing Criteria:**
  **Integration Tests:**
  - Verify remaining integration tests still pass: component_download_integration_test.go, debug_flag_integration_test.go, e2e_workflow_integration_test.go
  - Run go test -tags=integration ./... to confirm no regressions
  
  **Unit Tests:**
  - Verify new unit tests cover functionality from deleted story tests
  - Run go test ./... to confirm all unit tests pass

- [ ] Story-008: As a developer, I want coverage verified before and after cleanup so that I can confirm no test coverage was lost.

  **Acceptance Criteria:**
  - Run `go test -cover ./...` before cleanup and save baseline coverage report
  - Run `go test -tags=integration -cover ./...` before cleanup for integration coverage baseline
  - After each story completion, run coverage and compare to baseline
  - Document coverage changes in PR description (expected: same or better due to focused unit tests)
  - Final coverage report shows >= baseline coverage across all packages
  
  **Testing Criteria:**
  **Unit Tests:**
  - Compare unit test coverage percentage before/after for pkg/profiles/, internal/git/, internal/updater/
  - Verify new unit tests contribute to coverage increase
  
  **Integration Tests:**
  - Compare integration test coverage before/after for main package
  - Verify consolidated integration tests maintain critical path coverage

- [ ] Story-009: As a developer, I want documentation updated so that the testing guide reflects the new structure.

  **Acceptance Criteria:**
  - Update `TESTING.md` with new test organization: testutil package, focused integration tests, package-level unit tests
  - Remove references to story-specific tests
  - Add section on using testutil helpers in new tests
  - Update test file count and organization description
  - Delete `VERIFICATION.md` (temporary verification doc for Story-005, no longer relevant)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Note: Documentation changes verified through PR review
  
  **Integration Tests:**
  - Note: Documentation accuracy verified by following guide to run tests

- [ ] Story-010: As a developer, I want debug flag tests removed so that we don't test CLI framework functionality unnecessarily.

  **Acceptance Criteria:**
  - Delete `debug_flag_integration_test.go` (122 lines) - tests Cobra flag parsing, not application logic
  - Note: Flag parsing is Cobra's responsibility and doesn't require integration testing
  - Note: Debug/verbose output behavior is observable through normal CLI usage and doesn't need dedicated tests
  - No replacement test needed - Cobra framework handles flag parsing reliably
  
  **Testing Criteria:**
  **Integration Tests:**
  - Verify remaining integration tests still pass after deletion
  - Run go test -tags=integration ./... to confirm no dependencies on debug flag test

---

## Functional Requirements

- FR-1: The system SHALL extract TestHelper struct and helper methods from component_download_integration_test.go to internal/testutil/helpers.go
- FR-2: The system SHALL provide verification helpers (VerifyFileExists, VerifyDirExists, VerifyFileContent, CountFilesInDir) in internal/testutil/verification.go
- FR-3: The system SHALL provide test fixtures (mock repo structures, metadata templates) in internal/testutil/fixtures.go
- FR-4: The system SHALL add unit tests in pkg/profiles/ covering profile reuse and activation logic from Story-001 integration tests
- FR-5: The system SHALL add unit tests in internal/git/ covering URL normalization from Story-002 integration tests
- FR-6: The system SHALL add unit tests in internal/updater/ covering update logic from Story-003, Story-004, Story-006 integration tests
- FR-7: The system SHALL refactor component_download_integration_test.go to use internal/testutil package
- FR-8: The system SHALL create e2e_workflow_integration_test.go with 4 focused happy path workflows
- FR-9: The system SHALL delete 9 integration test files totaling 2,600 lines (8 story tests + 1 debug flag test)
- FR-10: The system SHALL delete debug_flag_integration_test.go as it tests Cobra framework functionality
- FR-11: The system SHALL verify test coverage before and after cleanup using go test -cover
- FR-12: The system SHALL update TESTING.md to reflect new test organization
- FR-13: The system SHALL delete VERIFICATION.md as it documents a temporary verification suite
- FR-14: The system SHALL maintain or improve overall test coverage after cleanup
- FR-15: The system SHALL reduce integration test code from 3,397 lines to approximately 700 lines (component_download ~400 + e2e_workflow ~300)
- FR-16: The system SHALL ensure all tests pass after each story completion: go test -tags=integration ./...

---

## Non-Goals

- No changes to existing unit tests in internal/detector/, internal/fileutil/, internal/linker/ (already well-tested)
- No rewrite of integration test approach or framework (keeping Go standard testing)
- No addition of new testing tools or libraries (using existing go test)
- No changes to CI/CD pipelines or test execution scripts (existing commands still work)
- No performance optimization of test execution time beyond removing redundancy
- No migration to different testing paradigms (e.g., BDD, table-driven tests) unless locally beneficial
- No consolidation of unit tests (only integration tests are being cleaned up)
- No addition of new test types (e.g., benchmarks, fuzzing) as part of this cleanup

---

## Test Migration Mapping

This section documents where each deleted integration test's functionality is migrated:

### Story-001 Integration Test → pkg/profiles/ Unit Tests
- Profile reuse detection → `pkg/profiles/reuse_test.go`
- Profile metadata handling → `pkg/profiles/reuse_test.go`
- Different repo creates different profiles → `pkg/profiles/reuse_test.go`
- Metadata integrity across updates → `pkg/profiles/reuse_test.go`

### Story-002 Integration Test → internal/git/ Unit Tests
- URL variations (HTTPS, SSH, HTTP, shorthand) → `internal/git/url_normalization_test.go`
- Case-insensitive domains → `internal/git/url_normalization_test.go`
- Trailing slash removal → `internal/git/url_normalization_test.go`
- .git extension handling → `internal/git/url_normalization_test.go`
- GitLab/Bitbucket support → `internal/git/url_normalization_test.go`

### Story-003 Integration Test → internal/updater/ Unit Tests
- Update single component with active profile → `internal/updater/profile_update_test.go`
- Update with explicit --profile flag → `internal/updater/profile_update_test.go`
- Update without active profile → `internal/updater/base_directory_test.go`

### Story-004 Integration Test → Multiple Locations
- Install with --profile flag → Covered by e2e_workflow_integration_test.go (Workflow 3)
- Profile name validation → `pkg/profiles/` existing tests
- Profile/target-dir flag conflict → CLI validation logic (not tested, simple error check)

### Story-004 Update Base Directory Test → internal/updater/ Unit Tests
- Update all checks base directory → `internal/updater/base_directory_test.go`
- Update single component in base directory → `internal/updater/base_directory_test.go`
- Update summary shows base directory components → `internal/updater/base_directory_test.go`
- Backward compatibility without profiles → `internal/updater/base_directory_test.go`

### Story-005 Verification Test → Deleted (Temporary Test)
- Command availability verification → Commands exist if code compiles
- Flag verification → Covered by debug_flag_integration_test.go and manual testing
- Profile workflow verification → Covered by e2e_workflow_integration_test.go
- No migration needed - this was a meta-test validating other tests

### Story-006 Integration Test → internal/updater/ Unit Tests
- Active profile location message → `internal/updater/profile_update_test.go`
- Explicit profile location message → `internal/updater/profile_update_test.go`
- Base directory has no location message → `internal/updater/base_directory_test.go`
- Single component update shows location → `internal/updater/profile_update_test.go`

---

## Coverage Verification Plan

### Baseline Coverage (Before Cleanup)

```bash
# Run before starting any changes
go test -cover ./... > coverage-before-unit.txt
go test -tags=integration -cover ./... > coverage-before-integration.txt
```

### Per-Story Coverage Checks

After completing each story:

```bash
# Story-001: After creating testutil package
go test -cover ./internal/testutil/... 

# Story-002: After adding profile unit tests
go test -cover ./pkg/profiles/...

# Story-003: After adding git unit tests
go test -cover ./internal/git/...

# Story-004: After adding updater unit tests
go test -cover ./internal/updater/...

# Story-005: After refactoring component_download test
go test -tags=integration -cover ./component_download_integration_test.go

# Story-006: After creating e2e_workflow test
go test -tags=integration -cover ./e2e_workflow_integration_test.go

# Story-007: After deleting story tests
go test -tags=integration -cover ./...

# Story-008: Final coverage comparison
go test -cover ./... > coverage-after-unit.txt
go test -tags=integration -cover ./... > coverage-after-integration.txt
diff coverage-before-unit.txt coverage-after-unit.txt
diff coverage-before-integration.txt coverage-after-integration.txt
```

### Success Criteria

- Unit test coverage in pkg/profiles/ increases (new tests added)
- Unit test coverage in internal/git/ maintained or increases
- Unit test coverage in internal/updater/ increases significantly (new package tests)
- Integration test coverage for main package maintained (consolidated tests cover same paths)
- No package shows coverage decrease > 2% (acceptable variation)

---

## Implementation Notes

### Execution Order (Sequential Dependencies)

This PRD is designed for single PR execution with these dependencies:

1. **Story-001 (testutil)** → Must complete first (foundation for all other stories)
2. **Story-002, Story-003, Story-004** → Can execute in parallel (independent unit test additions)
3. **Story-005 (refactor component_download)** → Depends on Story-001 (needs testutil)
4. **Story-006 (e2e_workflow)** → Depends on Story-001 (needs testutil)
5. **Story-007 (delete story tests)** → Depends on Story-002, Story-003, Story-004, Story-006 (needs replacement tests)
6. **Story-008 (coverage verification)** → Runs throughout all stories
7. **Story-009 (documentation)** → Can execute anytime, typically last
8. **Story-010 (delete debug flag test)** → Can execute anytime, typically with Story-007

### Parallel Execution Groups for Ralphy

- **Group 0 (Foundation)**: Story-001 (testutil package)
- **Group 1 (Independent Unit Tests)**: Story-002 (profiles), Story-003 (git), Story-004 (updater)
- **Group 2 (Dependent Tests)**: Story-005 (refactor), Story-006 (e2e)
- **Group 3 (Cleanup)**: Story-007 (delete story tests), Story-009 (docs), Story-010 (delete debug test)
- **Group 4 (Verification)**: Story-008 (coverage - runs throughout)

### File Impact Summary

**New Files Created (4 files, ~700 lines):**
- `internal/testutil/helpers.go` (~200 lines)
- `internal/testutil/verification.go` (~100 lines)
- `internal/testutil/fixtures.go` (~100 lines)
- `e2e_workflow_integration_test.go` (~300 lines)

**New Unit Test Files (6 files, ~600 lines estimated):**
- `pkg/profiles/reuse_test.go` (~150 lines)
- `pkg/profiles/activation_test.go` (~100 lines)
- `internal/git/url_normalization_test.go` (~150 lines)
- `internal/updater/profile_update_test.go` (~100 lines)
- `internal/updater/base_directory_test.go` (~100 lines)

**Modified Files:**
- `component_download_integration_test.go` (refactored: 808 → ~400 lines)
- `TESTING.md` (updated with new structure)

**Deleted Files (10 files, ~3,208 lines):**
- `story_001_integration_test.go` (343 lines)
- `story_002_integration_test.go` (413 lines)
- `story_003_update_single_component_test.go` (249 lines)
- `story_004_integration_test.go` (374 lines)
- `story_004_update_base_directory_test.go` (203 lines)
- `story_005_verification_integration_test.go` (425 lines)
- `story_005_feature_test.go` (211 lines)
- `story_006_update_location_feedback_test.go` (260 lines)
- `debug_flag_integration_test.go` (122 lines)
- `VERIFICATION.md` (101 lines)

**Net Change:**
- Lines added: ~1,300 (testutil + unit tests + e2e)
- Lines removed: ~3,208 (story tests + debug test + verification doc + helper refactor)
- Net reduction: ~1,908 lines
- Integration test reduction: 3,608 → ~700 lines (81% reduction - component_download ~400 + e2e_workflow ~300)

---

## Success Metrics

### Quantitative Metrics

- Integration test files reduced from 10 to 2 (80% reduction)
- Integration test lines reduced from 3,608 to ~700 (81% reduction)
- Unit test coverage maintained or improved (measured by go test -cover)
- All tests pass: `go test -tags=integration ./...` exits with 0

### Qualitative Metrics

- Test execution time for unit tests improves (unit tests faster than integration tests)
- New testutil package is reused across multiple test files
- Test purpose is clear (e2e workflows vs unit logic)
- Test maintenance easier (focused scope per test)
- PRD can be executed as single cohesive PR with clear dependencies

---

## Risks & Mitigations

### Risk: Coverage Loss During Migration

**Mitigation:**
- Run coverage reports before, during, and after each story
- Story-008 dedicated to coverage verification
- If coverage drops, add targeted unit tests before proceeding

### Risk: Edge Cases Missed in Migration

**Mitigation:**
- Manually audit each deleted test's test cases
- Document migration mapping (see "Test Migration Mapping" section)
- Create checklist of test cases from story tests before deletion

### Risk: Integration Tests Break During Refactor

**Mitigation:**
- Story-001 creates testutil package without breaking existing tests
- Story-005 refactors one test file at a time with verification
- Run `go test -tags=integration ./...` after each story completion

### Risk: New Unit Tests Don't Cover CLI Behavior

**Mitigation:**
- Story-006 creates e2e_workflow_integration_test.go to cover critical CLI paths
- Focus unit tests on business logic, integration tests on CLI orchestration
- Review deleted story tests to ensure CLI-specific behavior is in e2e test

### Risk: Parallel Execution Conflicts in Ralphy

**Mitigation:**
- Clear dependency groups defined (Group 0 → 1 → 2 → 3 → 4)
- Story-001 is blocking foundation for Group 1 stories
- Story-007 depends on all test creation stories (Group 0, 1, 2)
- Each story is self-contained with clear inputs/outputs

---

## Appendix: Current Test Inventory

### Integration Tests (Root Level)

| File | Lines | Created | Purpose | Status |
|------|-------|---------|---------|--------|
| component_download_integration_test.go | 808 | 2026-01-25 | Component downloading, repo detection, cross-platform paths | **KEEP (refactor to ~400 lines)** |
| debug_flag_integration_test.go | 122 | 2026-01-30 | Debug flag functionality (tests Cobra) | **DELETE (unnecessary)** |
| story_001_integration_test.go | 343 | 2026-01-31 | Profile reuse on repeated installs | **DELETE (move to unit tests)** |
| story_002_integration_test.go | 413 | 2026-01-31 | URL variation recognition | **DELETE (move to unit tests)** |
| story_003_update_single_component_test.go | 249 | 2026-01-31 | Single component update with active profile | **DELETE (move to unit tests)** |
| story_004_integration_test.go | 374 | 2026-01-31 | Force creation of new profiles with --profile flag | **DELETE (move to unit tests)** |
| story_004_update_base_directory_test.go | 203 | 2026-01-31 | Update commands on base directory | **DELETE (move to unit tests)** |
| story_005_verification_integration_test.go | 425 | 2026-01-31 | Comprehensive verification suite (temporary) | **DELETE (temporary test)** |
| story_005_feature_test.go | 211 | 2026-01-31 | Feature demo for Story-005 profile workflows | **DELETE (feature demo)** |
| story_006_update_location_feedback_test.go | 260 | 2026-01-31 | Location feedback in update commands | **DELETE (move to unit tests)** |
| **Total** | **3,608** | | | **2 files remain (~700 lines)** |

### Unit Tests (Package Level)

| Package | Test Files | Coverage Area |
|---------|-----------|---------------|
| internal/detector/ | 7 files | Component detection, patterns, frontmatter |
| internal/fileutil/ | 5 files | File operations, directory copy, error messages |
| internal/linker/ | 4 files | Auto-linking, profile collision, unlinking |
| internal/git/ | 1 file | Git operations |
| internal/downloader/ | 1 file | Error cleanup |
| internal/updater/ | 1 file | Update operations |
| pkg/profiles/ | Unknown | Profile management (needs audit) |

---

## Questions & Decisions

### Resolved Decisions

✅ **Test Migration Priority**: Extract testutil first, then migrate tests (safest approach)
✅ **Coverage Verification**: Run coverage reports before/after each phase
✅ **Documentation**: Delete VERIFICATION.md entirely (temporary verification doc)
✅ **E2E Test Scope**: Happy path workflows only (focused on critical user paths)
✅ **Timeline**: Single PR - all changes together (optimal for Ralphy parallel execution)
✅ **Cleanup Strategy**: Aggressive cleanup - remove story tests, rely on unit tests

### Open Questions

None - all clarifying questions answered during PRD generation.
