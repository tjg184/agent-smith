# PRD: Fix Empty Profile Creation on Failed Install

**Created**: 2026-02-07 18:50 UTC

---

## Introduction

When running `agent-smith install all some_invalid_url`, the current implementation creates the profile directory first, then attempts to download from the repository. If the URL is invalid or the download fails for any reason, the user is left with an empty profile that they must manually clean up. This creates a poor user experience and leaves orphaned state.

This PRD addresses the issue by validating the repository accessibility and detecting components **before** creating the profile, ensuring profiles are only created for successful installations.

## Goals

- Prevent creation of empty profiles when repository download fails
- Maintain single clone operation (no double-cloning)
- Preserve existing profile reuse behavior for repeated installs
- Keep the fix localized to the `install all` command path

## User Stories

- [ ] Story-001: As a user, I want failed `install all` commands to not create empty profiles so that I don't have to manually clean up failed attempts.

  **Acceptance Criteria:**
  - When `install all` is called with an invalid URL, no profile directory is created
  - When `install all` is called with a valid URL that fails during download, no profile directory is created
  - When `install all` succeeds, the profile is created and populated with components
  - The error message clearly indicates why the install failed

  **Testing Criteria:**
  **Unit Tests:**
  - Test `ValidateRepo()` returns error for invalid URLs
  - Test `ValidateRepo()` returns component list for valid repos
  - Test profile creation only occurs after successful validation

  **Integration Tests:**
  - Test full `install all` flow with invalid URL (no profile created)
  - Test full `install all` flow with valid URL (profile created with components)

- [ ] Story-002: As a user, I want repository validation to reuse the temporary clone so that installs complete faster.

  **Acceptance Criteria:**
  - Repository is cloned only once during the `install all` process
  - The temporary clone used for validation is reused for component installation
  - No performance regression in successful installs

  **Testing Criteria:**
  **Unit Tests:**
  - Test that validation returns the temp directory path for reuse
  - Test that installation accepts and uses the pre-cloned temp directory

  **Integration Tests:**
  - Verify single clone operation in successful install flow

## Functional Requirements

- **FR-1**: The system SHALL validate repository accessibility before creating any profile directory
- **FR-2**: The system SHALL detect available components during validation to ensure the repository contains installable content
- **FR-3**: The system SHALL reuse the validation clone for component installation (single clone operation)
- **FR-4**: The system SHALL maintain existing profile reuse behavior (detect and update existing profiles from the same source)
- **FR-5**: The system SHALL provide clear error messages when validation fails (invalid URL, no components found, network issues)
- **FR-6**: The system SHALL only create the profile directory after successful validation AND successful component installation

## Non-Goals

- Cleanup of partially created profiles (out of scope - we're preventing creation instead)
- Changes to individual `install skill|agent|command` commands (out of scope)
- Changes to profile activation behavior (out of scope)
- Retry logic for failed downloads (out of scope)
- Validation of individual component contents (out of scope)

## Implementation Approach

### Current Flow
```
1. Create profile directory
2. Clone repo to temp directory
3. Detect components
4. Install components to profile
5. (If step 2-4 fails, profile remains empty)
```

### New Flow
```
1. Clone repo to temp directory (validation)
2. Detect components
3. If valid:
   a. Create profile directory
   b. Install components from temp directory
4. If invalid:
   a. Return error, no profile created
```

### Key Changes

**1. Add `ValidateRepo()` method to `BulkDownloader`**
- Extract the clone + detect logic from `AddAll()` into a separate method
- Returns: `(tempDir string, components []DetectedComponent, err error)`
- Cleans up temp dir on error, preserves it on success for reuse

**2. Modify `installBulkToProfile()` in install service**
- Call `ValidateRepo()` first
- Only create profile if validation succeeds
- Pass temp directory to installation to avoid re-cloning

**3. Add `AddAllFromTemp()` method to `BulkDownloader`**
- Accepts pre-cloned temp directory and component list
- Skips clone, proceeds directly to installation
- Uses existing component downloaders with shared repo

### Files to Modify

1. `/Users/tgaines/dev/git/agent-smith/internal/downloader/bulk.go`
   - Add `ValidateRepo(repoURL string) (tempDir string, components []models.DetectedComponent, err error)`
   - Add `AddAllFromTemp(repoURL, tempDir string, components []models.DetectedComponent) error`
   - Refactor `AddAll()` to use `ValidateRepo()` + `AddAllFromTemp()`

2. `/Users/tgaines/dev/git/agent-smith/pkg/services/install/service.go`
   - Modify `installBulkToProfile()` to validate before profile creation
   - Use `ValidateRepo()` then `AddAllFromTemp()` for installation

## Risk Assessment

**Risk**: Profile reuse detection happens before validation in current flow
**Mitigation**: Move profile existence check after validation, or check if profile would be created vs reused before validation

**Risk**: Changes to `AddAll()` might affect other callers
**Mitigation**: Keep `AddAll()` public interface unchanged, internally refactor to use new methods

**Risk**: Temp directory cleanup on partial failures
**Mitigation**: Ensure `ValidateRepo()` cleans up temp dir on any error, caller responsible for cleanup after successful validation

## Success Criteria

1. Running `agent-smith install all invalid_url` does not create any profile directory
2. Running `agent-smith install all valid_url` creates a profile with components
3. The clone operation happens exactly once per `install all` command
4. Existing tests continue to pass
5. New tests cover validation-first behavior
