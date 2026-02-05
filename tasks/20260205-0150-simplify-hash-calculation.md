# PRD: Simplify Source Hash Calculation Using Local File Hash

**Created**: 2026-02-05 01:50 UTC

---

## Introduction

Currently, the system attempts to calculate `sourceHash` by making GitHub API calls to get tree SHAs for component folders. This approach:
- Fails for root-level components (no folder path to look up)
- Makes 2 API calls per component (main + master branches)
- Causes slowness during installations
- Shows misleading warnings: "failed to compute source hash: skill folder not found in GitHub API"
- Uses a different hash algorithm than `currentHash`, making comparison meaningless

This PRD simplifies the approach by using local file hashing for both `sourceHash` and `currentHash`, making them actually comparable while eliminating slow GitHub API calls.

## Goals

- Eliminate slow GitHub API calls during component installation
- Remove misleading "skill folder not found" warnings
- Make `sourceHash` and `currentHash` actually comparable (same algorithm)
- Support drift detection for all component types (root-level, nested, local, GitHub)
- Simplify the codebase by removing unnecessary GitHub tree API logic
- Improve installation performance (no network latency)

## User Stories

- [ ] Story-001: As a developer, I want source hash calculated from local files so that installations are faster and don't require GitHub API calls.

  **Acceptance Criteria:**
  - `sourceHash` calculated using `ComputeLocalFolderHash()` instead of `ComputeGitHubTreeSHA()`
  - No GitHub API calls made during hash calculation
  - Hash calculation works for all component types (skills, agents, commands)
  - Hash calculation works for all installation types (root-level, nested, local, GitHub)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Verify `sourceHash` equals `currentHash` immediately after installation
  - Verify hash calculation doesn't make network calls
  
  **Integration Tests:**
  - Install root-level component and verify hash is calculated
  - Install nested component and verify hash is calculated
  - Verify no API calls during installation

- [ ] Story-002: As a developer, I want source and current hashes to use the same algorithm so that drift detection actually works.

  **Acceptance Criteria:**
  - Both `sourceHash` and `currentHash` use `ComputeLocalFolderHash()` with SHA256
  - Hashes match when content is identical
  - Hashes differ when files are modified locally
  - Hash comparison is meaningful for drift detection
  
  **Testing Criteria:**
  **Unit Tests:**
  - Verify same directory produces identical hashes
  - Verify modified directory produces different hash
  
  **Integration Tests:**
  - Install component, modify file, verify `currentHash` changes but `sourceHash` stays same

- [ ] Story-003: As a developer, I want content-only hashing (no timestamps) so that hashes are stable across git operations.

  **Acceptance Criteria:**
  - Hash calculation includes file paths and content only
  - Hash calculation excludes file modification times
  - Hash is reproducible after git clone/checkout operations
  - Hash changes only when actual file content changes
  
  **Testing Criteria:**
  **Unit Tests:**
  - Verify hash unchanged after touching file (timestamp change only)
  - Verify hash changed after editing file content
  - Verify hash unchanged after git operations that preserve content

- [ ] Story-004: As a user, I want faster component installations without misleading warnings.

  **Acceptance Criteria:**
  - No "failed to compute source hash" warnings during installation
  - Installation completes without GitHub API rate limiting issues
  - Installation works offline (no network dependency for hashing)
  - Installation time reduced (no network latency)
  
  **Testing Criteria:**
  **Integration Tests:**
  - Install component and verify no hash-related warnings
  - Install multiple components and verify performance improvement
  - Install component without network access and verify hash calculated

- [ ] Story-005: As a developer, I want to remove unused GitHub tree API code to simplify the codebase.

  **Acceptance Criteria:**
  - `ComputeGitHubTreeSHA()` function removed from `internal/metadata/hash.go`
  - GitHub tree API imports removed
  - All references to GitHub tree SHA removed from downloaders
  - Code complexity reduced
  
  **Testing Criteria:**
  **Unit Tests:**
  - Verify all tests still pass after removing GitHub tree API code
  - Verify no references to `ComputeGitHubTreeSHA` remain in codebase

- [ ] Story-006: As a developer, I want materialize operations to use the same hash algorithm as install operations.

  **Acceptance Criteria:**
  - Verify `materializer.CalculateDirectoryHash()` uses same algorithm as `ComputeLocalFolderHash()`
  - If different, consolidate to single hash function
  - Both install and materialize operations produce comparable hashes
  - Drift detection works consistently across install and materialize contexts
  
  **Testing Criteria:**
  **Unit Tests:**
  - Verify same directory hashed by both functions produces same result
  
  **Integration Tests:**
  - Install component, materialize it, verify hashes match
  - Modify materialized component, verify drift detected

## Functional Requirements

- FR-1: The system SHALL calculate `sourceHash` using local file content hash (SHA256) immediately after installation
- FR-2: The system SHALL calculate `currentHash` using the same local file content hash algorithm
- FR-3: The system SHALL NOT make GitHub API calls for hash calculation
- FR-4: The system SHALL hash file paths and content only, excluding modification timestamps
- FR-5: The system SHALL set `sourceHash` equal to `currentHash` at installation time (no modifications yet)
- FR-6: The system SHALL recalculate `currentHash` during `materialize info` to detect drift
- FR-7: The system SHALL compare `sourceHash` with recalculated `currentHash` to determine sync status
- FR-8: The system SHALL display "In Sync" when hashes match, "Modified" when they differ
- FR-9: The system SHALL remove `ComputeGitHubTreeSHA()` function and related GitHub tree API code
- FR-10: The system SHALL use consistent hash algorithm across all component types and operations

## Non-Goals

- No changes to update detection logic (still uses `CommitHash` to detect upstream changes)
- No changes to uninstall operations
- No changes to linking/unlinking operations
- No automatic migration of existing lock files with old-style hashes
- No backward compatibility for comparing new hashes with old GitHub tree SHAs
- No UI changes to `materialize info` command output format

## Implementation Notes

### Files to Modify

1. **internal/metadata/hash.go**:
   - Remove `ComputeGitHubTreeSHA()` function
   - Update `ComputeLocalFolderHash()` to exclude timestamps (content-only)

2. **internal/downloader/skill.go**:
   - Remove GitHub tree SHA calculation code (lines 395-416)
   - Use `ComputeLocalFolderHash(skillDir)` for both `sourceHash` and `currentHash`

3. **internal/downloader/agent.go**:
   - Remove GitHub tree SHA calculation code
   - Use `ComputeLocalFolderHash(agentDir)` for both hashes

4. **internal/downloader/command.go**:
   - Remove GitHub tree SHA calculation code
   - Use `ComputeLocalFolderHash(commandDir)` for both hashes

5. **pkg/services/materialize/service.go**:
   - Verify `materializer.CalculateDirectoryHash()` uses same algorithm
   - Consolidate if different

### Testing Strategy

1. **Unit Tests**: Verify hash algorithm produces consistent results
2. **Integration Tests**: Verify end-to-end workflows (install → info → detect drift)
3. **Performance Tests**: Measure installation time improvement (before/after)

### Migration Considerations

Existing installations with GitHub tree SHA hashes will show "Modified" status in `materialize info` because the hash algorithms differ. This is acceptable because:
- Users can reinstall to get new-style hashes
- The issue is cosmetic (doesn't affect functionality)
- Future installations will work correctly
