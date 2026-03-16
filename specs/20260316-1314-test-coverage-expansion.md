# PRD: Test Coverage Expansion

**Created**: 2026-03-16 13:14 UTC

---

## Introduction

The agent-smith codebase has significant gaps in unit test coverage. Three integration test files are missing required build tags (causing them to run unintentionally as unit tests), several core packages have zero test coverage, and integration test file naming is inconsistent. This PRD captures all planned remediation work.

## Goals

- Fix correctness bugs: add missing `//go:build integration` tags to 3 files
- Normalize integration test file naming to `_test.go` suffix
- Add lightweight unit tests (using `os.TempDir()`) for all untested core packages
- Mirror existing opencode target tests for claudecode target
- Extend partial unit test files with missing function coverage

## User Stories

- [ ] Story-001: As a developer, I want integration test files to be correctly gated by the `integration` build tag so that `go test ./...` does not unintentionally run them.

  **Acceptance Criteria:**
  - `profile_error_messages_test.go` has `//go:build integration` as its first line
  - `profile_rename_test.go` has `//go:build integration` as its first line
  - `profile_share_active_test.go` has `//go:build integration` as its first line
  - `go test ./...` completes without attempting to run integration test binary setup

  **Testing Criteria:**
  **Unit Tests:**
  - `go test ./...` passes without the `integration` tag and does not reference `AgentSmithBinary`

- [ ] Story-002: As a developer, I want integration test file names to be consistent so that the test suite is easy to navigate.

  **Acceptance Criteria:**
  - `e2e_workflow_integration_test.go` → `e2e_workflow_test.go`
  - `uninstall_integration_test.go` → `uninstall_test.go`
  - `unlink_integration_test.go` → `unlink_test.go`
  - All integration tests still run correctly with `-tags=integration`

  **Testing Criteria:**
  **Unit Tests:**
  - `go test -tags=integration ./tests/integration/...` passes after renaming

- [ ] Story-003: As a developer, I want TESTING.md to accurately document the current test structure so I can quickly understand how to run tests.

  **Acceptance Criteria:**
  - TESTING.md lists the correct number of integration test files
  - All filenames referenced in TESTING.md exist on disk
  - Instructions for running unit and integration tests are accurate

  **Testing Criteria:**
  - Manual review: all file paths in TESTING.md resolve to existing files

- [ ] Story-004: As a developer, I want unit tests for `internal/uninstaller` so that uninstall logic is verified without requiring a live git repo.

  **Acceptance Criteria:**
  - `UninstallComponent` rejects invalid component types with a clear error
  - `UninstallComponent` removes the component directory from disk when not shared
  - `UninstallComponent` keeps the directory when shared by another source
  - `UninstallComponent` updates the lock file after removal
  - `UninstallAllFromSource` finds and removes all components from a given source URL (force mode)
  - `isDirectorySharedByOtherSource` returns true when another source maps to the same filesystem name
  - `normalizeURLForComparison` strips trailing slashes, `.git` suffix, and lowercases
  - `matchesSourceURL` returns true for equivalent URL variants

  **Testing Criteria:**
  **Unit Tests:**
  - Tests use `os.MkdirTemp` + real filesystem; no network access required
  - Lock files created programmatically using `metadata.SaveComponentEntry`

- [ ] Story-005: As a developer, I want unit tests for `internal/materializer` so that file copy and sync-detection logic is verified in isolation.

  **Acceptance Criteria:**
  - `CopyDirectory` copies all files and subdirectories recursively
  - `CopyDirectory` returns an error when source does not exist
  - `CopyFlatMdFiles` copies only non-ignored `.md` files into destDir (no wrapper subdir)
  - `CopyFlatMdFiles` ignores README.md, LICENSE.md, DOCS.md, CHANGELOG.md
  - `FlatMdFilesMatch` returns true when dest has identical content for all src `.md` files
  - `FlatMdFilesMatch` returns false when a file is missing from dest
  - `FlatMdFilesMatch` returns false when content differs
  - `FlatMdFilesMatch` returns false (not an error) when dest file is missing
  - `DirectoriesMatch` returns true for identical directories
  - `DirectoriesMatch` returns false when content differs
  - `CalculateDirectoryHash` produces the same hash for identical directory trees

  **Testing Criteria:**
  **Unit Tests:**
  - Tests use `os.MkdirTemp`; no external dependencies

- [ ] Story-006: As a developer, I want unit tests for `pkg/project/materialization.go` so that the collision-safe filesystem name resolution and entry update logic is verified.

  **Acceptance Criteria:**
  - `ResolveFilesystemName` returns the base name when no collision exists
  - `ResolveFilesystemName` returns `name-2` when base name is taken on disk
  - `ResolveFilesystemName` returns `name-3` when both base name and `name-2` are taken
  - `ResolveFilesystemName` is idempotent: re-resolving the same sourceUrl+componentName returns the existing filesystem name
  - `ResolveFilesystemName` detects metadata-only conflicts (no disk directory required)
  - `UpdateMaterializationEntry` updates `SourceHash`, `CurrentHash`, and `MaterializedAt`
  - `UpdateMaterializationEntry` returns an error for unknown component types
  - `UpdateMaterializationEntry` returns an error when component is not in metadata
  - `AddMaterializationEntry` stores all fields correctly and is keyed by sourceUrl
  - `GetMaterializationComponentMap` returns correct map for each component type and nil for unknown type

  **Testing Criteria:**
  **Unit Tests:**
  - Tests use `os.MkdirTemp` for disk-collision scenarios; metadata-only tests use in-memory structs

- [ ] Story-007: As a developer, I want unit tests for `pkg/config/claudecode_target.go` that mirror the existing opencode target tests so that both target implementations are equally verified.

  **Acceptance Criteria:**
  - `NewClaudeCodeTargetWithDir` returns a target with the correct base directory
  - `GetGlobalBaseDir` returns the configured directory
  - `GetGlobalSkillsDir` returns `<base>/skills`
  - `GetGlobalAgentsDir` returns `<base>/agents`
  - `GetGlobalCommandsDir` returns `<base>/commands`
  - `GetGlobalComponentDir` returns correct paths for skills/agents/commands and errors for unknown types
  - `GetDetectionConfigPath` returns `<base>/detection-config.json`
  - `GetName` returns `"claudecode"`
  - `NewClaudeCodeTarget` succeeds and returns a target with a non-empty base directory

  **Testing Criteria:**
  **Unit Tests:**
  - Tests mirror `opencode_target_test.go` structure, adapted for claudecode paths

- [ ] Story-008: As a developer, I want the `internal/metadata/lock_test.go` extended with tests for `FindAllComponentInstances` and `LoadAllComponents` so their multi-source behavior is explicitly verified.

  **Acceptance Criteria:**
  - `FindAllComponentInstances` returns all instances of a component across sources
  - `FindAllComponentInstances` returns an empty slice when the component does not exist
  - `FindAllComponentInstances` returns an empty slice when the lock file does not exist
  - `LoadAllComponents` returns all components from a given type's lock file
  - `LoadAllComponents` returns an empty slice when the lock file does not exist
  - `LoadAllComponents` returns entries from multiple sources

  **Testing Criteria:**
  **Unit Tests:**
  - Tests use `t.TempDir()` consistent with existing tests in the file

## Functional Requirements

- FR-1: The system SHALL reject `go test ./...` running integration tests (they require the compiled binary)
- FR-2: All integration test files in `tests/integration/` SHALL use the `_test.go` naming suffix
- FR-3: New unit tests SHALL use real temp directories (`os.MkdirTemp` / `t.TempDir()`) and SHALL NOT require network access
- FR-4: New unit test files SHALL be placed adjacent to the source file they test
- FR-5: The `internal/uninstaller` tests SHALL exercise both the "shared directory kept" and "sole owner removed" code paths
- FR-6: The `internal/materializer` tests SHALL cover all exported functions including both `FlatMd` and `Directory` variants
- FR-7: The `pkg/project` materialization tests SHALL cover suffix collision logic up to at least `-3`
- FR-8: The `pkg/config/claudecode_target_test.go` SHALL mirror `opencode_target_test.go` test-for-test

## Non-Goals

- No unit tests for `pkg/services/` orchestration layer (covered indirectly by integration tests)
- No mocking infrastructure or mock generation tooling
- No new integration tests (existing integration test suite is adequate)
- No changes to production source code
- No test coverage for `internal/downloader/bulk.go` (requires network; covered by integration tests)
