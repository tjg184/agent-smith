# PRD: Profile-Aware Linking Refactor

## Introduction

This PRD refactors the profile activation/deactivation system to separate concerns between state management (activate/deactivate) and linking operations (link commands). Currently, profile activation immediately creates symlinks in `~/.agents/`, which causes issues when users have existing base installations. The new design makes activation/deactivation lightweight metadata operations, while linking commands become profile-aware and handle the actual symlinking.

## Problem Statement

### Current Issues

1. **Destructive Activation**: When activating a profile, `ActivateProfile()` attempts to remove all components from `~/.agents/` and replace them with symlinks to the profile. This fails with "directory not empty" errors when real directories exist.

2. **Loss of Base Installation**: Users who install components to `~/.agents/` and then create profiles lose access to their base installation when activating a profile, with no way to restore it.

3. **Coupling of Concerns**: Profile activation mixes state management (tracking which profile is active) with linking operations (creating symlinks), making the code complex and error-prone.

4. **Poor User Experience**: Users expect activation to be a lightweight operation that marks a profile as active, with a separate explicit step to apply changes to their editor.

### User Workflow Example

A user wants to:
1. Install 131 skills from a repository to `~/.agents/`
2. Create subset profiles (e.g., "engineer" with 10 skills, "writer" with 5 skills)
3. Switch between profiles without losing the base installation
4. Deactivate to restore access to all 131 skills

Currently this workflow fails at step 2-3 due to the issues above.

## Goals

- **Separate State from Action**: Make `activate` only set metadata, make `link` handle symlinking
- **Preserve Base Installation**: Never modify `~/.agents/` during activation/deactivation
- **Profile-Aware Linking**: Link commands automatically use active profile when present
- **Explicit Control**: Users explicitly run `link` commands to apply profile changes
- **Backward Compatibility**: Existing link commands work unchanged
- **Simplify Code**: Remove complex symlink creation/removal from profile manager

## User Stories

- [ ] Story-001: As a user, I want to activate a profile without immediately affecting my editor, so that I can control when changes are applied.

  **Acceptance Criteria:**
  - `agent-smith profiles activate <name>` only updates state file (`~/.agents/.active_profile`)
  - Does not create or remove any symlinks
  - Does not modify `~/.agents/` directory contents
  - Validates profile exists before updating state
  - Returns error if same profile already active
  - Prints guidance message: "Run 'agent-smith link all' to apply changes to your editor."

  **Testing Criteria:**
  **Unit Tests:**
  - Verify state file written correctly
  - Verify no symlinks created
  - Verify `~/.agents/` directory unchanged
  - Verify error on duplicate activation
  
  **Integration Tests:**
  - Activate profile, verify only state file changes
  - Verify `~/.agents/` contents identical before/after

- [ ] Story-002: As a user, I want to deactivate a profile without immediately affecting my editor, so that I can control when changes are applied.

  **Acceptance Criteria:**
  - `agent-smith profiles deactivate` only clears state file
  - Does not create or remove any symlinks
  - Does not modify `~/.agents/` directory contents
  - Returns error if no profile is active
  - Prints guidance message: "Run 'agent-smith link all' to restore base installation to your editor."

  **Testing Criteria:**
  **Unit Tests:**
  - Verify state file deleted
  - Verify no symlinks removed
  - Verify `~/.agents/` directory unchanged
  - Verify error when no active profile
  
  **Integration Tests:**
  - Deactivate profile, verify only state file removed
  - Verify `~/.agents/` contents identical before/after

- [ ] Story-003: As a user, I want link commands to automatically use my active profile, so that linking matches my current context.

  **Acceptance Criteria:**
  - When profile active, `link` commands source from `~/.agents/profiles/<profile>/` (ALREADY IMPLEMENTED)
  - When no profile active, `link` commands source from `~/.agents/`
  - Displays message "Using active profile: <name>" when profile active
  - Works for all link commands: `link skills`, `link agents`, `link commands`, `link all`
  - No breaking changes to existing link command syntax

  **Testing Criteria:**
  **Unit Tests:**
  - Verify source directory determination logic
  - Verify profile detection
  
  **Integration Tests:**
  - Activate profile, run `link all`, verify sources from profile
  - Deactivate profile, run `link all`, verify sources from base
  - Verify correct message displayed

- [ ] Story-004: As a user, I want to install components to `~/.agents/`, create profiles from them, and switch between profiles without losing my base installation.

  **Acceptance Criteria:**
  - Can install 100+ components to `~/.agents/`
  - Can create multiple profiles using `profiles add` to copy components
  - Can activate profile without affecting `~/.agents/` contents
  - Can run `link all` to link only profile components
  - Can deactivate profile without affecting `~/.agents/` contents
  - Can run `link all` after deactivate to restore all base components
  - Base installation in `~/.agents/` remains intact throughout

  **Testing Criteria:**
  **Integration Tests:**
  - Full workflow test with 10+ components
  - Verify `~/.agents/` unchanged after multiple activate/deactivate cycles
  - Verify linking works correctly at each step

- [ ] Story-005: As a user, I want clear guidance on what to do after activation/deactivation, so that I understand the two-step workflow.

  **Acceptance Criteria:**
  - Activation prints: "Profile '<name>' activated. Run 'agent-smith link all' to apply changes to your editor."
  - Deactivation prints: "Profile deactivated. Run 'agent-smith link all' to restore base installation to your editor."
  - Help text for `profiles activate` explains two-step process
  - Help text for `profiles deactivate` explains two-step process
  - `link` command help mentions profile awareness

  **Testing Criteria:**
  **Manual Tests:**
  - Verify messages display correctly
  - Verify help text is clear and accurate

- [ ] Story-006: As a user, I want to see which profile is active and where components are sourced from, so that I understand the current state.

  **Acceptance Criteria:**
  - `agent-smith status` shows active profile (ALREADY IMPLEMENTED)
  - `agent-smith link status` optionally shows source directory
  - Status display makes it clear when profile is active vs base installation

  **Testing Criteria:**
  **Integration Tests:**
  - Verify status shows correct active profile
  - Verify status distinguishes profile vs base mode

## Functional Requirements

### Profile Activation
- **FR-001**: `ActivateProfile()` must only write to `~/.agents/.active_profile` state file
- **FR-002**: `ActivateProfile()` must NOT create symlinks in `~/.agents/`
- **FR-003**: `ActivateProfile()` must NOT modify `~/.agents/` directory contents
- **FR-004**: `ActivateProfile()` must validate profile exists before updating state
- **FR-005**: `ActivateProfile()` must return error if profile already active
- **FR-006**: `ActivateProfile()` must print guidance message about running `link all`

### Profile Deactivation
- **FR-007**: `DeactivateProfile()` must only delete `~/.agents/.active_profile` state file
- **FR-008**: `DeactivateProfile()` must NOT remove symlinks from `~/.agents/`
- **FR-009**: `DeactivateProfile()` must NOT modify `~/.agents/` directory contents
- **FR-010**: `DeactivateProfile()` must return error if no profile is active
- **FR-011**: `DeactivateProfile()` must print guidance message about running `link all`

### Profile-Aware Linking
- **FR-012**: Link commands must check for active profile before determining source
- **FR-013**: Link commands must use `~/.agents/profiles/<profile>/` when profile active (ALREADY IMPLEMENTED)
- **FR-014**: Link commands must use `~/.agents/` when no profile active (ALREADY IMPLEMENTED)
- **FR-015**: Link commands must display message when using active profile (ALREADY IMPLEMENTED)
- **FR-016**: Link behavior must be identical whether profile active or not (only source changes)

### Backward Compatibility
- **FR-017**: Existing `link` commands must work unchanged
- **FR-018**: Existing `unlink` commands must work unchanged
- **FR-019**: Existing `profiles add/remove/delete` commands must work unchanged
- **FR-020**: Profiles created before this change must continue to work

## Non-Goals

- No automatic linking after activation (explicit step required)
- No automatic unlinking before deactivation (user responsibility)
- No migration of existing active profiles (clean slate approach)
- No preservation of symlinks in `~/.agents/` (profiles don't manage them)
- No validation that user ran `link` after activation (trust user)
- No shorthand commands like `profiles activate --link` (can add later)

## Technical Implementation

### Files to Modify

#### 1. `pkg/profiles/manager.go`

**ActivateProfile() Changes (lines 223-331):**
- Remove lines 243-258: Deactivation of current profile
- Remove lines 260-321: Symlink creation loop
- Keep lines 226-235: Profile validation
- Keep lines 237-241: Get agents directory
- Simplify lines 244-253: Check if already active
- Keep lines 323-327: State file writing
- Update success message

**New implementation (~40 lines):**
```go
func (pm *ProfileManager) ActivateProfile(profileName string) error {
    // Validate profile name
    if err := validateProfileName(profileName); err != nil {
        return err
    }

    // Validate that the profile exists
    profile := pm.loadProfile(profileName)
    if !profile.IsValid() {
        return fmt.Errorf("profile '%s' does not exist or has no components", profileName)
    }

    // Get the agents directory
    agentsDir, err := paths.GetAgentsDir()
    if err != nil {
        return fmt.Errorf("failed to get agents directory: %w", err)
    }

    // Check if already active
    currentActive, err := pm.GetActiveProfile()
    if err != nil {
        return fmt.Errorf("failed to check current active profile: %w", err)
    }
    
    if currentActive == profileName {
        return fmt.Errorf("profile '%s' is already active", profileName)
    }

    // Update the active profile state file
    activeProfilePath := filepath.Join(agentsDir, ".active_profile")
    if err := os.WriteFile(activeProfilePath, []byte(profileName), 0644); err != nil {
        return fmt.Errorf("failed to write active profile state: %w", err)
    }

    fmt.Printf("Profile '%s' activated.\n", profileName)
    fmt.Println("Run 'agent-smith link all' to apply changes to your editor.")
    
    return nil
}
```

**DeactivateProfile() Changes (lines 479-513):**
- Remove lines 500-503: unlinkAllComponents call
- Keep lines 482-496: Profile existence check
- Keep lines 505-509: State file clearing
- Update success message

**New implementation (~25 lines):**
```go
func (pm *ProfileManager) DeactivateProfile() error {
    // Get the agents directory
    agentsDir, err := paths.GetAgentsDir()
    if err != nil {
        return fmt.Errorf("failed to get agents directory: %w", err)
    }

    // Check if a profile is currently active
    currentActive, err := pm.GetActiveProfile()
    if err != nil {
        return fmt.Errorf("failed to check current active profile: %w", err)
    }

    if currentActive == "" {
        return fmt.Errorf("no profile is currently active")
    }

    fmt.Printf("Deactivating profile: %s\n", currentActive)

    // Clear the active profile state file
    activeProfilePath := filepath.Join(agentsDir, ".active_profile")
    if err := os.Remove(activeProfilePath); err != nil && !os.IsNotExist(err) {
        return fmt.Errorf("failed to clear active profile state: %w", err)
    }

    fmt.Println("Profile deactivated.")
    fmt.Println("Run 'agent-smith link all' to restore base installation to your editor.")

    return nil
}
```

**unlinkAllComponents() (lines 594-634):**
- Keep function unchanged (used by `unlink` commands)
- Function no longer called by activate/deactivate

#### 2. `cmd/root.go`

**profilesActivateCmd (around line 529):**
Update help text:
```go
Long: `Activate a profile to mark it as the active configuration.

When a profile is active, the 'link' command will use components from
the profile directory instead of ~/.agents/.

After activating a profile, run 'agent-smith link all' to apply the
changes to your editor(s).

WORKFLOW:
  1. agent-smith profiles activate <name>
  2. agent-smith link all

This two-step process gives you explicit control over when changes
are applied to your editor.`,
```

**profilesDeactivateCmd (around line 542):**
Update help text:
```go
Long: `Deactivate the currently active profile.

After deactivating, the 'link' command will use components from 
~/.agents/ (base installation) instead of a profile.

Run 'agent-smith link all' after deactivating to restore the base
installation to your editor(s).

WORKFLOW:
  1. agent-smith profiles deactivate
  2. agent-smith link all

This ensures you can switch back to your full base installation.`,
```

#### 3. `internal/linker/linker.go` (Optional Enhancement)

**ShowLinkStatus() (line 577):**
Add source directory to header:
```go
fmt.Println("\n=== Link Status Across All Targets ===")

// Show source directory
sourceType := "base installation"
if strings.Contains(cl.agentsDir, "/profiles/") {
    parts := strings.Split(cl.agentsDir, "/profiles/")
    if len(parts) == 2 {
        profileName := strings.Split(parts[1], "/")[0]
        sourceType = fmt.Sprintf("profile '%s'", profileName)
    }
}
fmt.Printf("Source: %s (%s)\n", cl.agentsDir, sourceType)
fmt.Println()
```

#### 4. `main.go` (Optional Enhancement)

**Status command (line 751):**
Add source clarification:
```go
fmt.Println()
if activeProfile != "" {
    fmt.Printf("Note: Link commands will source from profile '%s'\n", activeProfile)
    fmt.Printf("      Run 'agent-smith profiles deactivate' to use base installation\n")
} else {
    fmt.Println("Note: Link commands will source from ~/.agents/ (base installation)")
    fmt.Println("      Run 'agent-smith profiles activate <name>' to use a profile")
}
```

#### 5. `pkg/profiles/manager_test.go`

**Update test expectations:**
- `TestActivateProfile`: Should NOT verify symlinks created
- `TestDeactivateProfile`: Should NOT verify symlinks removed
- Add `TestActivateProfileOnlyUpdatesState`: Verify only state file changes
- Add `TestDeactivateProfileOnlyClearsState`: Verify only state file removed
- Add `TestActivateProfileTwiceReturnsError`: Verify duplicate activation error

### Code to Remove

**Total lines removed: ~85 lines**
- Profile activation symlink creation loop: ~58 lines
- Profile deactivation unlinkAllComponents call: ~4 lines
- Related error handling and validation: ~23 lines

### Code to Add

**Total lines added: ~30 lines**
- Updated help text: ~20 lines
- Optional status enhancements: ~10 lines

### Net Change

**~55 lines removed** (simplification)

## Success Criteria

- User can activate/deactivate profiles without modifying `~/.agents/`
- User can run full workflow: install → create profile → activate → link → deactivate → link
- Base installation in `~/.agents/` preserved through all operations
- Link commands automatically use profile when active
- All existing tests pass (with updated expectations)
- Help text clearly guides users through two-step process
- No regression in existing link/unlink functionality

## Testing Strategy

### Unit Tests

**New tests:**
1. `TestActivateProfileOnlyUpdatesState` - Verify only `.active_profile` written
2. `TestDeactivateProfileOnlyClearsState` - Verify only `.active_profile` removed
3. `TestActivateProfileDoesNotModifyAgentsDir` - Verify `~/.agents/` unchanged
4. `TestDeactivateProfileDoesNotModifyAgentsDir` - Verify `~/.agents/` unchanged
5. `TestActivateProfileTwiceError` - Verify error on duplicate activation

**Modified tests:**
1. `TestActivateProfile` - Remove symlink verification, check only state
2. `TestDeactivateProfile` - Remove symlink verification, check only state

### Integration Tests

**New tests:**
1. `TestFullProfileWorkflow` - Install → create → activate → link → verify
2. `TestProfileSwitching` - Create 2 profiles, switch between them
3. `TestBasePreservation` - Verify `~/.agents/` unchanged through activate/deactivate
4. `TestDeactivateRestore` - Verify deactivate + link restores all components

### Manual Testing Checklist

- Install 10+ skills to `~/.agents/`
- Create profile with 2 skills
- Activate profile (verify no changes to `~/.agents/`)
- Run `link all` (verify only 2 skills linked)
- Create second profile with 3 skills
- Activate second profile (verify `~/.agents/` still has all 10 skills)
- Run `link all` (verify 3 skills linked, replacing previous 2)
- Deactivate profile (verify `~/.agents/` still has all 10 skills)
- Run `link all` (verify all 10 skills linked)
- Run `agent-smith status` at each step (verify correct profile shown)
- Run `agent-smith link status` at each step (verify correct links)

## Migration Plan

### Existing Users with Active Profiles

Users who have active profiles before this change will need to:
1. Run `agent-smith profiles deactivate` (clears state, removes old symlinks)
2. Run `agent-smith profiles activate <name>` (sets new state-only activation)
3. Run `agent-smith link all` (creates new profile-aware links)

**Note:** We could add a migration detection:
- Check if active profile exists but `~/.agents/` contains profile symlinks
- Print warning: "Old-style profile activation detected. Please deactivate and reactivate."
- This is outside scope of this PRD but could be added as follow-up

### Breaking Changes

**None for most users** - The refactor only changes the internal implementation of activate/deactivate.

**Potential impact:**
- Users who rely on activation immediately creating symlinks will need to add `link all` step
- This is mitigated by clear guidance messages printed by activate/deactivate

## Documentation Updates

### README/User Guide
- Update profile workflow section
- Add explicit two-step process: activate → link
- Add examples showing full workflow

### CLI Help Text
- Updated as described in Technical Implementation section

### Migration Guide
- Document for users with existing active profiles
- Explain new two-step workflow

## Risks and Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Users confused by two-step process | Medium | Low | Clear guidance messages, updated help text |
| Users forget to run `link` after activate | Medium | Low | Guidance message reminds them |
| Breaking existing automation scripts | Low | Medium | Version change, changelog, migration guide |
| State file corruption | Very Low | Low | Simple file I/O, same as current implementation |

## Timeline

- **Design Review**: 1 day
- **Implementation**: 4 hours
  - Code changes: 2 hours
  - Test updates: 1 hour
  - Manual testing: 1 hour
- **Documentation**: 1 hour
- **Total**: ~1 day

## Open Questions

1. Should we add automatic migration detection for old-style active profiles?
   - **Recommendation**: Add as follow-up PR, not blocking for this refactor

2. Should we add a `--link` flag to `profiles activate` for convenience?
   - **Recommendation**: Add as follow-up enhancement, keep explicit for now

3. Should we prevent `profiles add` when profile is active?
   - **Recommendation**: No, current validation handles this correctly

4. Should `link status` always show source directory or only optionally?
   - **Recommendation**: Always show, provides useful context

## Success Metrics

- Zero issues with "directory not empty" errors
- Base installation preserved across activate/deactivate cycles
- User workflow completion time unchanged (two-step is explicit, not slower)
- No regression in existing functionality
- Test coverage maintained or improved

## Conclusion

This refactor simplifies the profile system by separating state management (activate/deactivate) from linking operations (link commands). It solves the current "directory not empty" bug, preserves base installations, and provides users with explicit control over when changes are applied to their editors. The changes are mostly deletions (~85 lines removed), making the codebase simpler and more maintainable while improving the user experience.
