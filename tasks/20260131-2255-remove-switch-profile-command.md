# PRD: Remove Switch Profile Command

**Created**: 2026-01-31 22:55 UTC

---

## Introduction

Remove the redundant `switch` profile command from agent-smith. Currently, both `activate` and `switch` commands exist and perform identical operations - they both update the `.active-profile` state file without creating symlinks. This redundancy causes user confusion and unnecessary code complexity. This PRD focuses on removing `switch` while keeping `activate` as the single, clear command for making a profile active.

## Goals

- Eliminate confusion between `activate` and `switch` commands
- Reduce codebase complexity by removing duplicate functionality
- Standardize on `activate` as the single command for profile activation
- Maintain all existing functionality while simplifying the API
- Clean up documentation to reference only `activate`

## User Stories

- [ ] Story-001: As a developer, I want to remove the `profilesSwitchCmd` command definition so that users cannot invoke the switch command.

  **Acceptance Criteria:**
  - `profilesSwitchCmd` command removed from `cmd/root.go`
  - `handleProfilesSwitch` function removed from `cmd/root.go`
  - `handleProfilesSwitch` variable declaration removed
  - `handleProfilesSwitch` initialization removed
  - Command no longer appears in `--help` output
  - Running `agent-smith profile switch <name>` returns "unknown command" error
  
  **Testing Criteria:**
  **Unit Tests:**
  - No unit tests required (command removal)
  
  **Integration Tests:**
  - Verify `agent-smith profile switch` returns error
  - Verify `agent-smith profile --help` does not list switch
  
  **Component Browser Tests:**
  - N/A (CLI only)

- [ ] Story-002: As a developer, I want to remove the `SwitchProfile()` method from ProfileManager so that the redundant implementation is eliminated.

  **Acceptance Criteria:**
  - `SwitchProfile()` method removed from `pkg/profiles/manager.go`
  - All references to `SwitchProfile()` removed from `main.go`
  - No compilation errors after removal
  - `ActivateProfile()` method remains unchanged and functional
  
  **Testing Criteria:**
  **Unit Tests:**
  - Verify existing `ActivateProfile()` tests still pass
  
  **Integration Tests:**
  - Run full test suite to ensure no breaking changes
  
  **Component Browser Tests:**
  - N/A (backend only)

- [ ] Story-003: As a developer, I want to remove the switch profile test file so that obsolete tests are eliminated.

  **Acceptance Criteria:**
  - `pkg/profiles/switch_profile_test.go` file deleted
  - No test failures after deletion
  - Test suite runs successfully
  - `ActivateProfile()` tests in `manager_test.go` remain and pass
  
  **Testing Criteria:**
  **Unit Tests:**
  - Run `go test ./...` to verify no test failures
  
  **Integration Tests:**
  - Run integration test suite
  
  **Component Browser Tests:**
  - N/A (test cleanup)

- [ ] Story-004: As a user, I want documentation to reference only `activate` so that I'm not confused about which command to use.

  **Acceptance Criteria:**
  - `README.md` updated to remove all `switch` references
  - `PRODUCT_SNAPSHOT.md` updated to remove `switch` terminology
  - `skills/profile-builder/SKILL.md` references only `activate`
  - `skills/profile-builder/README.md` references only `activate`
  - Any help text or examples use only `activate`
  
  **Testing Criteria:**
  **Unit Tests:**
  - N/A (documentation only)
  
  **Integration Tests:**
  - Manual review of documentation
  - Search codebase for remaining `switch` references
  
  **Component Browser Tests:**
  - N/A (documentation only)

- [ ] Story-005: As a developer, I want to verify the changes work end-to-end so that I'm confident the removal was successful.

  **Acceptance Criteria:**
  - Compilation succeeds with no errors
  - All tests pass (`go test ./...`)
  - `agent-smith profile activate <name>` works correctly
  - `agent-smith profile deactivate` works correctly
  - `agent-smith profile --help` shows correct commands
  - No references to "switch" in user-facing output
  
  **Testing Criteria:**
  **Unit Tests:**
  - Full unit test suite passes
  
  **Integration Tests:**
  - Manual end-to-end workflow testing
  - Create profile → activate → verify state → deactivate
  
  **Component Browser Tests:**
  - N/A (CLI verification)

## Functional Requirements

- **FR-1**: The system SHALL NOT include a `switch` command in the CLI interface
- **FR-2**: The system SHALL remove the `SwitchProfile()` method from ProfileManager
- **FR-3**: The system SHALL maintain full functionality of `ActivateProfile()` method
- **FR-4**: The system SHALL update all documentation to reference only `activate`
- **FR-5**: The system SHALL pass all existing tests after `switch` removal
- **FR-6**: The system SHALL return appropriate error when users try `agent-smith profile switch`
- **FR-7**: The system SHALL NOT break any existing profile workflows
- **FR-8**: The system SHALL remove all `switch`-related test files

## Non-Goals

- Not adding new profile functionality
- Not modifying the behavior of `activate` command
- Not changing how profiles work internally
- Not updating historical task files or git commit messages
- Not deprecating `switch` with warnings (clean removal)
- Not creating migration scripts (functionality is identical)
- Not adding aliases for `switch` → `activate`

## Implementation Notes

### Files to Modify

1. **cmd/root.go** (~1250 lines)
   - Remove `profilesSwitchCmd` definition (lines ~1224-1251)
   - Remove `handleProfilesSwitch` function variable
   - Remove `handleProfilesSwitch` initialization
   - Remove command registration

2. **pkg/profiles/manager.go** (~695 lines)
   - Remove `SwitchProfile()` method (lines ~650-695)

3. **main.go**
   - Remove any `pm.SwitchProfile()` calls
   - Keep all `pm.ActivateProfile()` calls

4. **Documentation files**
   - `README.md`
   - `PRODUCT_SNAPSHOT.md`
   - `skills/profile-builder/SKILL.md`
   - `skills/profile-builder/README.md`

5. **Test files**
   - Delete: `pkg/profiles/switch_profile_test.go`
   - Keep: `pkg/profiles/manager_test.go` (ActivateProfile tests)

### Files to Delete

- `pkg/profiles/switch_profile_test.go` (entire file)

### Search Terms for Cleanup

```bash
# Find remaining references
grep -r "switch" --include="*.go" --include="*.md" cmd/ pkg/ *.md skills/

# Verify no SwitchProfile calls remain
grep -r "SwitchProfile" --include="*.go" .

# Verify no handleProfilesSwitch references
grep -r "handleProfilesSwitch" --include="*.go" .
```

## Testing Strategy

1. **Pre-removal verification**
   - Run full test suite: `go test ./...`
   - Verify current behavior of both commands

2. **Post-removal verification**
   - Compilation check: `go build`
   - Run full test suite: `go test ./...`
   - Manual CLI testing:
     ```bash
     ./agent-smith profile activate test-profile
     ./agent-smith profile deactivate
     ./agent-smith profile switch test-profile  # Should error
     ./agent-smith profile --help  # Should not show switch
     ```

3. **Documentation verification**
   - Search for remaining "switch" references in docs
   - Verify all examples use "activate"

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Breaking existing user scripts | Low | Medium | Document in changelog; commands are identical |
| Incomplete removal leaves orphaned code | Low | Low | Use grep to verify all references removed |
| Documentation inconsistencies | Medium | Low | Thorough documentation review |
| Test failures after removal | Low | Low | Comprehensive test run after changes |

## Success Criteria

- [ ] Code compiles without errors
- [ ] All tests pass
- [ ] `agent-smith profile switch` returns error
- [ ] `agent-smith profile activate` works correctly
- [ ] Documentation references only `activate`
- [ ] No `SwitchProfile` references in codebase
- [ ] Help output shows only `activate` and `deactivate`
