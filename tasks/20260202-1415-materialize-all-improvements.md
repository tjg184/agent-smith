# PRD: Materialize All Command Improvements

**Created**: 2026-02-02 14:15 UTC

---

## Introduction

Fix two critical issues with the `agent-smith materialize all --target opencode` command that prevent users from successfully materializing components to new projects. Currently, the command fails silently when the target directory doesn't exist, and the summary output is hidden unless the user runs with `--verbose` flag, making it difficult to understand what happened.

## Goals

- Allow users to run `materialize all --target opencode` in directories without existing `.opencode` folders
- Automatically create target directory structure when it doesn't exist
- Always display command summary regardless of verbose mode
- Provide clear feedback when creating new project structures
- Maintain backward compatibility with existing project detection behavior

## User Stories

- [x] Story-001: As a user, I want to run `materialize all --target opencode` in a directory without an existing `.opencode` folder so that I can quickly set up a new project.

  **Acceptance Criteria:**
  - Command succeeds when no `.opencode` directory exists
  - `.opencode/` directory is created with proper subdirectories (skills/, agents/, commands/)
  - Components are materialized to the newly created structure
  - Helpful message indicates new project was created in current directory
  - Existing behavior unchanged when `.opencode` already exists

  **Testing Criteria:**
  **Unit Tests:**
  - No unit tests needed (directory creation covered by existing `EnsureTargetStructure` tests)
  
  **Integration Tests:**
  - Start in empty directory with no `.opencode/` folder
  - Run `materialize all --target opencode` with test components installed
  - Verify `.opencode/` directory is created
  - Verify components are materialized successfully
  - Verify output contains "creating new project" message

- [x] Story-002: As a user, I want to see the summary of what was materialized even when running without `--verbose` so that I can verify the command succeeded.

  **Acceptance Criteria:**
  - Summary section displays without `--verbose` flag
  - Summary shows total components count
  - Summary shows successful materialization count
  - Summary shows skipped count (when applicable)
  - Summary shows error count and messages (when applicable)
  - Progress messages during materialization still respect verbose flag

  **Testing Criteria:**
  **Unit Tests:**
  - No unit tests needed (logger.Print() already tested)
  
  **Integration Tests:**
  - Run `materialize all --target opencode` without `--verbose` flag
  - Capture stdout and verify summary section exists
  - Verify summary contains "Total components:" line
  - Verify summary contains "Materialized:" line with count
  - Verify summary is not empty or hidden

- [x] Story-003: As a developer, I want the project root detection to gracefully fallback to current directory so that new projects can be created without pre-existing markers.

  **Acceptance Criteria:**
  - When `FindProjectRoot()` fails to locate existing project markers, use current working directory
  - Log informational message: "No existing project found, creating new project in: [path]"
  - Do not exit with fatal error when no project markers found
  - Maintain existing behavior when project markers exist (finds correct project root)
  - Handle edge case where user cannot get current working directory

  **Testing Criteria:**
  **Unit Tests:**
  - No new unit tests needed (uses existing os.Getwd() which is stdlib)
  
  **Integration Tests:**
  - Test fallback behavior in directory with no project markers
  - Verify informational message is logged
  - Test existing behavior preserved when `.opencode` exists in parent directory
  - Verify command doesn't exit prematurely

## Functional Requirements

- FR-1: The `materialize all` command SHALL use the current working directory as project root when `FindProjectRoot()` fails to locate existing project markers (.opencode or .claude directories)

- FR-2: The `materialize all` command SHALL log an informational message "No existing project found, creating new project in: [path]" when falling back to current directory

- FR-3: The `materialize all` command SHALL call `EnsureTargetStructure()` to create target directory and subdirectories (skills/, agents/, commands/) when they don't exist

- FR-4: The summary section of `materialize all` command SHALL always display regardless of verbose mode setting

- FR-5: The summary section SHALL use `appLogger.Print()` method instead of `appLogger.Info()` to bypass log level filtering

- FR-6: The summary section SHALL display: total components count, successful materialization count, skipped count (if > 0), and error count with messages (if > 0)

- FR-7: Progress messages during component materialization SHALL continue to respect the verbose flag setting (not affected by summary changes)

- FR-8: The command SHALL maintain backward compatibility with existing project detection when .opencode or .claude directories already exist

## Technical Implementation Details

### Change 1: Project Root Detection Fallback

**File**: `main.go` (lines ~2476-2482)

Replace:
```go
// Auto-detect project root
projectRoot, err = project.FindProjectRoot()
if err != nil {
    log.Fatal(err)
}
```

With:
```go
// Auto-detect project root
projectRoot, err = project.FindProjectRoot()
if err != nil {
    // No existing project found, use current directory
    projectRoot, err = os.Getwd()
    if err != nil {
        log.Fatalf("Failed to get current directory: %v", err)
    }
    infoPrintf("No existing project found, creating new project in: %s\n", projectRoot)
}
```

### Change 2: Summary Output Always Visible

**File**: `main.go` (lines 2779-2797)

Replace all `infoPrintf()` calls in summary section with `appLogger.Print()`:

**Lines to modify**: 2779, 2781, 2783, 2787, 2789, 2793, 2795

Before:
```go
infoPrintf("  Total components: %d\n", totalComponents)
```

After:
```go
appLogger.Print("  Total components: %d", totalComponents)
```

**Note**: Keep `appFormatter.Section()` calls unchanged (lines 2777-2778) as they already output directly to stdout.

### Files Modified

1. **main.go**
   - Lines 2476-2482: Add fallback for `FindProjectRoot()` failure
   - Lines 2779, 2781, 2783, 2787, 2789, 2793, 2795: Replace `infoPrintf` with `appLogger.Print` in summary section

2. **tests/integration/materialize_all_without_project_test.go** (new file)
   - Test materialization in directory without existing .opencode folder
   - Test summary visibility without --verbose flag
   - Test existing project detection behavior unchanged

## Non-Goals (Out of Scope)

- No changes to individual `materialize skill/agent/command` commands (only `materialize all`)
- No changes to `--project-dir` flag behavior
- No changes to profile-based materialization logic
- No changes to conflict handling or force flag behavior
- No changes to dry-run mode functionality
- No new command flags or options
- No changes to materialization metadata format
- No changes to lock file structure or parsing
