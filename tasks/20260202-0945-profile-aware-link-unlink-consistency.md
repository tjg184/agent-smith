# PRD: Profile-Aware Link/Unlink Consistency

**Created**: 2026-02-02 09:45 UTC

---

## Introduction

Currently, `agent-smith link all` only links components from the currently active profile, but `agent-smith unlink all` unlinks ALL symlinks found in target directories regardless of which profile they came from. This creates an inconsistency where users can accidentally unlink components from profiles they're not currently working with.

This PRD addresses the inconsistency by making `unlink all` profile-aware to match `link all` behavior, and adds an `--all-profiles` flag for both commands to provide flexibility when users need to work across all profiles.

## Goals

- Make `unlink all` behavior consistent with `link all` (both profile-aware by default)
- Prevent accidental unlinking of components from inactive profiles
- Provide explicit `--all-profiles` flag for bulk operations when needed
- Maintain clear user messaging about which profiles are affected
- Ensure backward compatibility for users without profiles

## User Stories

- [x] Story-001: As a user with an active profile, I want `unlink all` to only unlink components from my current profile so that I don't accidentally remove links from other profiles.

  **Acceptance Criteria:**
  - When a profile is active, `unlink all` only removes symlinks pointing to that profile's components
  - Symlinks pointing to other profiles are skipped and preserved
  - User sees clear messaging about which profile's components are being unlinked
  - User sees count of components skipped (from other profiles)
  - Without an active profile, behavior matches current (unlinks from base ~/.agent-smith/)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile path detection logic for symlink targets
  - Symlink filtering based on profile ownership
  
  **Integration Tests:**
  - Multi-profile scenario where profile A is active and both profile A and B components are linked
  - Verify only profile A components are unlinked
  - Verify profile B symlinks remain intact
  
  **Component Browser Tests:**
  - Not applicable (CLI command)

- [x] Story-002: As a user managing multiple profiles, I want an `--all-profiles` flag for `unlink all` so that I can bulk unlink components from all profiles when needed.

  **Acceptance Criteria:**
  - `unlink all --all-profiles` unlinks components from all profiles, not just active one
  - Confirmation prompt clearly states all profiles will be affected
  - Summary message shows total count across all profiles
  - Works regardless of whether a profile is active
  - Requires explicit flag (not default behavior)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Flag parsing and validation
  
  **Integration Tests:**
  - Multi-profile scenario with --all-profiles flag
  - Verify all symlinks from all profiles are removed
  - Verify confirmation prompt mentions "all profiles"
  
  **Component Browser Tests:**
  - Not applicable (CLI command)

- [x] Story-003: As a user with an active profile, I want `link all` to support `--all-profiles` flag so that I can link components from all profiles simultaneously when needed.

  **Acceptance Criteria:**
  - `link all --all-profiles` links components from all profiles to targets
  - Handles potential naming conflicts gracefully (first profile wins, shows warning)
  - Shows clear messaging about multiple profiles being linked
  - Summary indicates components linked from each profile
  - Requires explicit flag (not default behavior)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile iteration logic
  - Conflict detection and resolution
  
  **Integration Tests:**
  - Multi-profile linking with no conflicts
  - Multi-profile linking with naming conflicts (same component in multiple profiles)
  - Verify all profiles' components are linked when no conflicts
  - Verify conflict resolution follows first-wins strategy
  
  **Component Browser Tests:**
  - Not applicable (CLI command)

- [x] Story-004: As a developer, I want a reusable method to determine which profile a symlink belongs to so that profile-aware operations can share this logic.

  **Acceptance Criteria:**
  - New method `isSymlinkFromProfile(symlinkPath, profilePath)` returns boolean and profile name
  - Handles absolute and relative symlink paths correctly
  - Identifies symlinks from base installation (no profile)
  - Identifies symlinks from specific profiles
  - Returns clear error for broken or invalid symlinks
  
  **Testing Criteria:**
  **Unit Tests:**
  - Symlink path resolution (absolute and relative)
  - Profile path matching logic
  - Base installation detection
  - Profile name extraction from path
  - Error handling for broken symlinks
  
  **Integration Tests:**
  - Not needed (unit tests sufficient for path manipulation logic)
  
  **Component Browser Tests:**
  - Not applicable (internal utility method)

- [x] Story-005: As a user, I want clear messaging during unlink operations so that I understand which profiles are being affected and what is being skipped.

  **Acceptance Criteria:**
  - Confirmation prompts show exact profile names being affected
  - Progress messages indicate current profile context
  - Summary displays unlinked count, skipped count, and which profiles were involved
  - When using --all-profiles, message clearly states "all profiles"
  - Skipped items show which profile they belong to
  
  **Testing Criteria:**
  **Unit Tests:**
  - Message formatting logic
  - Profile name display formatting
  
  **Integration Tests:**
  - Capture and verify output messages during unlink operations
  - Verify correct profile names in confirmation prompts
  - Verify summary counts match actual operations
  
  **Component Browser Tests:**
  - Not applicable (CLI output verification)

- [ ] Story-006: As a user, I want the `--all-profiles` flag to fail gracefully when no profile manager is available so that I get clear error messages.

  **Acceptance Criteria:**
  - Error message when --all-profiles used but profile manager not initialized
  - Suggests checking profile configuration or using regular link/unlink
  - Does not crash or create corrupt state
  - Error message includes actionable guidance
  
  **Testing Criteria:**
  **Unit Tests:**
  - Error handling for missing profile manager
  
  **Integration Tests:**
  - Attempt --all-profiles operation without profile manager
  - Verify error message content and exit code
  
  **Component Browser Tests:**
  - Not applicable (error handling)

- [ ] Story-007: As a user without profiles, I want existing link/unlink behavior to remain unchanged so that profile support doesn't break my workflow.

  **Acceptance Criteria:**
  - With no active profile, `link all` links from base ~/.agent-smith/
  - With no active profile, `unlink all` unlinks from base ~/.agent-smith/
  - No profile-related messages shown when no profiles exist
  - Performance unchanged for non-profile users
  - All existing flags and options continue to work
  
  **Testing Criteria:**
  **Unit Tests:**
  - No-profile code path validation
  
  **Integration Tests:**
  - Fresh installation with no profiles
  - Verify link all and unlink all work as before
  - Verify no profile-related output or warnings
  
  **Component Browser Tests:**
  - Not applicable (backward compatibility verification)

- [ ] Story-008: As a developer, I want comprehensive tests for mixed-profile scenarios so that edge cases are properly handled.

  **Acceptance Criteria:**
  - Tests cover active profile with other profiles' symlinks present
  - Tests cover broken symlinks pointing to deleted profiles
  - Tests cover manually created symlinks outside agent-smith
  - Tests cover empty profiles (no components)
  - Tests cover profile with only some component types
  
  **Testing Criteria:**
  **Unit Tests:**
  - Edge case handling in profile detection logic
  - Broken symlink detection and handling
  
  **Integration Tests:**
  - Mixed symlinks scenario (profile A active, profiles A/B/C linked)
  - Broken symlinks scenario (deleted profile)
  - Manual symlinks scenario (created outside agent-smith)
  - Empty profile scenario
  - Partial profile scenario (only skills, no agents)
  
  **Component Browser Tests:**
  - Not applicable (test infrastructure)

## Functional Requirements

- FR-1: The `UnlinkAllComponents` method SHALL filter symlinks by profile ownership before removal
- FR-2: The system SHALL provide `isSymlinkFromProfile(symlinkPath, profilePath)` method returning (isMatch bool, profileName string, error)
- FR-3: The `unlink all` command SHALL accept `--all-profiles` flag to override profile filtering
- FR-4: The `link all` command SHALL accept `--all-profiles` flag to link from all profiles
- FR-5: When `--all-profiles` is used, the system SHALL iterate through all profile directories
- FR-6: Profile-aware operations SHALL handle naming conflicts with first-wins strategy and warning
- FR-7: Confirmation prompts SHALL clearly indicate which profile(s) will be affected
- FR-8: Summary messages SHALL display counts broken down by: unlinked, skipped (other profiles), errors
- FR-9: Skipped components SHALL display which profile they belong to
- FR-10: Operations without active profile SHALL use base ~/.agent-smith/ directory (backward compatible)
- FR-11: The system SHALL gracefully handle broken symlinks pointing to deleted profiles
- FR-12: The system SHALL identify and preserve manually-created symlinks outside agent-smith
- FR-13: The `handleLinkAll` function signature SHALL be updated to `func(targetFilter, profile string, allProfiles bool)`
- FR-14: The `handleUnlinkAll` function signature SHALL be updated to `func(targetFilter string, force bool, allProfiles bool)`
- FR-15: The `LinkAllComponents` method SHALL accept `allProfiles bool` parameter
- FR-16: The `UnlinkAllComponents` method SHALL accept `allProfiles bool` parameter
- FR-17: Help text SHALL document `--all-profiles` flag for both link and unlink commands
- FR-18: Error messages for `--all-profiles` without profile manager SHALL be actionable
- FR-19: The system SHALL maintain performance for non-profile users (no extra overhead)

## Non-Goals

- Not implementing profile creation, deletion, or management features
- Not changing behavior of single-component link/unlink commands
- Not implementing automatic profile switching
- Not adding profile-based permissions or access control
- Not implementing profile merging or conflict resolution beyond first-wins
- Not adding profile analytics or usage tracking
- Not implementing profile import/export functionality
- Not changing `link auto` behavior (repository detection)
- Not modifying profile activation/deactivation commands
- Not implementing profile-specific configuration overrides

## Implementation Notes

### Key Files to Modify

1. **internal/linker/linker.go** (~line 1648)
   - Add `isSymlinkFromProfile` method
   - Modify `UnlinkAllComponents` to filter by profile
   - Modify `LinkAllComponents` to support --all-profiles
   - Update confirmation and summary messages

2. **cmd/root.go** (~lines 640, 968)
   - Add `--all-profiles` flag to `link all` command
   - Add `--all-profiles` flag to `unlink all` command
   - Update help text for both commands

3. **main.go** (~lines 915-926, 998-1006)
   - Update `handleLinkAll` signature and implementation
   - Update `handleUnlinkAll` signature and implementation
   - Update function declarations (~line 1776, 1782)

### Profile Detection Algorithm

```
For each symlink in target directory:
  1. Read symlink target path
  2. Make target path absolute if relative
  3. Check if target starts with current profile path (cl.agentsDir)
     - If yes: symlink belongs to current profile
     - If no: continue to step 4
  4. Check if target starts with profiles directory
     - Extract profile name from path (first component after profiles dir)
     - Return: isFromCurrentProfile=false, profileName=extracted
  5. If neither: symlink from base installation or unknown
     - Return: isFromCurrentProfile=baseInstall, profileName=""
```

### Edge Cases

- **Broken symlinks**: Attempt to read target, if broken note as "unknown profile" and skip in profile-aware mode
- **Manual symlinks**: If symlink doesn't point to agent-smith directories, preserve with warning
- **Naming conflicts**: First profile wins, subsequent conflicts logged as warnings
- **Empty targets**: If no symlinks found, show "No linked components" message
- **Mixed profile symlinks**: Show detailed breakdown of which profiles contributed components

### Backward Compatibility

- No active profile: Behavior identical to current implementation
- No `--all-profiles` flag: Default behavior unchanged (profile-aware)
- Existing flags (`--target`, `--force`, `--profile`): Continue to work as before
- Non-profile users: Zero performance impact, no profile-related messages

## Success Metrics

- User confusion around profile unlinking reduced (measured by GitHub issues/questions)
- Zero regression bugs for non-profile users
- Clear user feedback in testing showing improved understanding of profile scope
- All existing integration tests continue to pass
- New tests provide >90% coverage for profile-aware code paths

## Documentation Updates

Files requiring documentation updates:
- `CONFIG.md` - Add --all-profiles flag examples
- `README.md` - Update link/unlink examples with profile context
- Command help text (in cmd/root.go) - Document profile-aware behavior
