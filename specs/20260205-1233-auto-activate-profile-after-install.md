# PRD: Auto-Activate Profile After Install All

**Created**: 2026-02-05 12:33 UTC

---

## Introduction

Currently, users must follow a 3-step workflow to install and use components: `install all` → `profile activate` → `link all`. This PRD streamlines the workflow by automatically activating the profile after `install all` completes, reducing it to 2 steps: `install all` → `link all`.

## Goals

- Reduce installation workflow from 3 steps to 2 steps
- Automatically activate profiles after `install all` command
- Switch to newly installed profile even if another profile is currently active
- Provide clear visual feedback about profile activation status
- Maintain user control over linking (do not auto-link)
- Ensure backward compatibility with existing workflows

## User Stories

- [ ] Story-001: As a user running `install all`, I want the profile to be automatically activated so that I don't need to manually run `profile activate`.

  **Acceptance Criteria:**
  - After successful `install all` execution, the profile is automatically activated
  - The `.active-profile` file is updated with the new profile name
  - If no profile was previously active, show: `✓ Profile activated: <profile-name>`
  - If switching from another profile, show: `✓ Switched profile: <old-profile> → <new-profile>`
  - Display next-step hint in color: `Next: Run 'agent-smith link all' to apply changes to your editor(s)`
  - Individual install commands (`install skill`, `install agent`, `install command`) are NOT affected
  
  **Testing Criteria:**
  **Unit Tests:**
  - ProfileManager returns previous profile information when activating new profile
  - Profile activation result contains both old and new profile names
  
  **Integration Tests:**
  - `install all` with no active profile sets active profile correctly
  - `install all` with existing active profile switches to new profile
  - `.active-profile` file content is updated correctly
  - Subsequent `link all` uses the newly activated profile as source

- [ ] Story-002: As a developer, I want the ProfileManager to return previous profile information so that calling code can display meaningful switch messages.

  **Acceptance Criteria:**
  - `ActivateProfile()` reads current active profile before changing it
  - Returns struct with `PreviousProfile` and `NewProfile` fields
  - `PreviousProfile` is empty string if no profile was previously active
  - All existing callers of `ActivateProfile()` are updated to handle new return type
  
  **Testing Criteria:**
  **Unit Tests:**
  - `ActivateProfile()` returns correct previous profile when switching
  - `ActivateProfile()` returns empty string for previous when none exists
  - Return struct correctly populated in both scenarios

- [ ] Story-003: As a user, I want clear feedback about profile activation so that I understand what happened and what to do next.

  **Acceptance Criteria:**
  - Success messages use ✓ symbol for visual clarity
  - "Switched profile" message shows old → new format with arrow
  - Next-step hint is displayed with colored text for visibility
  - Messages appear after installation summary, before command exits
  - No extra output added if activation fails (just warning)
  
  **Testing Criteria:**
  **Integration Tests:**
  - Output contains expected success messages after `install all`
  - Next-step hint appears with correct formatting
  - Messages display correctly in both scenarios (new activation vs. switch)

- [ ] Story-004: As a user, I want profile activation failures to not break installation so that I can manually activate if needed.

  **Acceptance Criteria:**
  - If activation fails, installation still succeeds
  - Warning message displayed: `⚠ Profile created but activation failed: <reason>`
  - User can manually run `profile activate <name>` as fallback
  - Installation components are successfully installed regardless of activation failure
  - Error is logged but does not return error from InstallBulk
  
  **Testing Criteria:**
  **Unit Tests:**
  - InstallBulk handles activation errors gracefully without propagating
  
  **Integration Tests:**
  - Installation succeeds even if activation fails
  - Warning message appears correctly in output
  - Profile is still created and components installed

- [ ] Story-005: As a developer, I want profile activation validated before execution so that we don't activate non-existent profiles.

  **Acceptance Criteria:**
  - Before calling `ActivateProfile()`, verify profile exists via ProfileManager
  - Only activate if profile directory and metadata are valid
  - Skip activation if profile validation fails (with warning)
  - Validation includes checking for `.profile-metadata` file
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile existence check validates directory structure
  - Profile existence check validates metadata file presence
  
  **Integration Tests:**
  - Activation only occurs for valid profiles after `install all`
  - Invalid profiles trigger appropriate warning messages

## Functional Requirements

- FR-1: The system SHALL automatically activate the profile after successful `install all` execution
- FR-2: The system SHALL switch to the newly installed profile even if another profile is currently active
- FR-3: The system SHALL display clear visual feedback including ✓ symbols and colored text for the next-step hint
- FR-4: The system SHALL NOT auto-link components (user maintains control over link command)
- FR-5: The system SHALL return previous profile information from `ActivateProfile()` for messaging purposes
- FR-6: The system SHALL validate profile existence before attempting activation
- FR-7: The system SHALL handle activation failures gracefully without breaking installation
- FR-8: The system SHALL NOT modify behavior of individual install commands (`install skill`, `install agent`, `install command`)

## Non-Goals

- No auto-linking after profile activation (user explicitly runs `link all`)
- No `--no-activate` flag to disable auto-activation
- No changes to individual install commands behavior
- No prompts or interactive confirmations during activation
- No changes to `profile activate` command behavior when run manually
- No changes to how profiles are created or managed

---

## Technical Implementation Notes

### Files to Modify

1. **`pkg/profiles/manager.go`**
   - Modify `ActivateProfile()` function (lines 519-567)
   - Return previous profile information for messaging

2. **`pkg/services/install/service.go`**
   - Modify `InstallBulk()` function (around lines 200-300)
   - Add profile activation after successful installation
   - Handle activation errors gracefully

3. **`cmd/root.go`**
   - Update `installAllCmd` execution (lines 270-315)
   - Display activation status messages
   - Add colored next-step hint

### Example Output

**Fresh install (no active profile):**
```bash
$ agent-smith install all anthropics/skills

Downloading from https://github.com/anthropics/skills...
✓ Installed skill: web-artifacts-builder
✓ Installed skill: docx
✓ Installed skill: pdf

✓ Profile activated: anthropics-skills

Next: Run 'agent-smith link all' to apply changes to your editor(s)
```

**Switching profiles:**
```bash
$ agent-smith install all other-repo/tools

Downloading from https://github.com/other-repo/tools...
✓ Installed skill: custom-tool

✓ Switched profile: anthropics-skills → other-repo-tools

Next: Run 'agent-smith link all' to apply changes to your editor(s)
```

### Activation Result Structure

```go
type ProfileActivationResult struct {
    PreviousProfile string // empty if no profile was active
    NewProfile      string
    Switched        bool   // true if switching from another profile
}
```
