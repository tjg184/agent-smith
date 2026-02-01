# PRD: Profile-Aware Update Command

**Created**: 2026-01-31 23:53 UTC

---

## Introduction

The `agent-smith update` command currently only updates components in the base `~/.agent-smith/` directory and is not aware of the profile system. This causes the update functionality to be completely non-functional when users have an active profile, as all components are stored in `~/.agent-smith/profiles/<profile-name>/` instead. This PRD addresses making the update command profile-aware, consistent with how the `link` command already operates.

**Problem Statement**: When a user has an active profile (e.g., `wshobson-agents`) with 236 components, running `./agent-smith update all` reports "Total components checked: 0" because it only looks in the empty base directory, not the profile directory where components actually exist.

## Goals

- Make `update all` command respect active profile configuration and update components from the correct location
- Add `--profile` flag support to allow updating specific profiles independent of active profile state
- Make `update <type> <name>` command profile-aware for single component updates
- Maintain backwards compatibility with existing behavior when no profile is active
- Display clear feedback about which profile/directory is being updated
- Align update command behavior with existing link command patterns for consistency

## User Stories

- [x] Story-001: As a user with an active profile, I want update all to check components in my active profile so that my profile components get updated.

  **Acceptance Criteria:**
  - When active profile exists, `update all` checks `~/.agent-smith/profiles/<profile-name>/` directory
  - Displays message "Using active profile: <profile-name>" before checking for updates
  - Updates all components (agents, skills, commands) found in the profile directory
  - Respects existing lock files in profile directory for version tracking
  - Shows update summary with correct component counts from profile
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile detection logic returns correct active profile name
  - Base directory resolution uses profile path when profile is active
  - Base directory resolution uses default path when no profile is active
  
  **Integration Tests:**
  - Update all with active profile processes profile components
  - Update all without active profile processes base directory components
  - Lock files are correctly read from profile directory
  - Update statistics reflect actual profile component counts

- [x] Story-002: As a user managing multiple profiles, I want to use --profile flag to update a specific profile's components so that I can update non-active profiles without switching.

  **Acceptance Criteria:**
  - `update all --profile <name>` updates specified profile's components
  - Bypasses active profile check when --profile flag is used
  - Works regardless of whether a profile is currently active
  - Validates that specified profile exists before attempting update
  - Displays clear message "Updating profile: <profile-name>" before checking
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile flag parsing extracts correct profile name
  - Profile validation checks profile directory exists
  - Profile flag overrides active profile detection
  
  **Integration Tests:**
  - Update with --profile flag updates correct profile directory
  - Update with --profile ignores active profile state
  - Error handling for non-existent profile names
  - Profile flag works with both "update all" and "update <type> <name>"

- [x] Story-003: As a user updating a single component, I want update <type> <name> to respect my active profile so that individual component updates work consistently with update all.

  **Acceptance Criteria:**
  - `update skills <name>` checks active profile if one exists
  - `update agents <name>` checks active profile if one exists
  - `update commands <name>` checks active profile if one exists
  - Supports --profile flag for explicit profile selection
  - Maintains backwards compatibility with base directory when no profile active
  
  **Testing Criteria:**
  **Unit Tests:**
  - Single component update uses correct base directory path
  - Component type validation works with profile paths
  - Metadata loading uses profile lock files when appropriate
  
  **Integration Tests:**
  - Update single skill from active profile succeeds
  - Update single agent from active profile succeeds
  - Update single command from active profile succeeds
  - Update single component with --profile flag uses specified profile

- [x] Story-004: As a user with no active profile, I want update commands to work on my base directory components so that I can still update components without using profiles.

  **Acceptance Criteria:**
  - When no profile is active, `update all` checks `~/.agent-smith/` directory
  - When no profile is active, `update <type> <name>` checks base directory
  - No profile-related messages displayed when no profile is active
  - Existing behavior is maintained for users not using profiles
  - Update summary shows components from base directory correctly
  
  **Testing Criteria:**
  **Unit Tests:**
  - Base directory path defaults to `~/.agent-smith/` when no profile active
  - Profile detection returns empty string when no active profile
  
  **Integration Tests:**
  - Update all without active profile processes base directory only
  - Update single component without active profile uses base directory
  - Lock files read from base directory when no profile active

- [ ] Story-005: As a developer maintaining the codebase, I want UpdateDetector to accept a configurable base directory so that it can work with both base and profile directories.

  **Acceptance Criteria:**
  - `NewUpdateDetector()` constructor accepts optional baseDir parameter
  - When baseDir is provided, UpdateDetector uses it instead of default `~/.agent-smith/`
  - When baseDir is not provided (nil/empty), defaults to `~/.agent-smith/` for backwards compatibility
  - All metadata loading, component scanning, and update operations respect the configured baseDir
  - UpdateDetector initialization validates that baseDir exists
  
  **Testing Criteria:**
  **Unit Tests:**
  - NewUpdateDetector with nil baseDir defaults to `~/.agent-smith/`
  - NewUpdateDetector with explicit baseDir uses provided path
  - Base directory validation checks directory exists
  - Metadata loading constructs correct paths using baseDir
  
  **Integration Tests:**
  - UpdateDetector with profile baseDir updates profile components
  - UpdateDetector with default baseDir updates base components
  - Error handling for invalid base directory paths

- [ ] Story-006: As a user running update commands, I want clear feedback about which location is being updated so that I understand where components are being checked.

  **Acceptance Criteria:**
  - Display "Using active profile: <name>" when updating from active profile
  - Display "Updating profile: <name>" when using --profile flag
  - Display "Updating components in: <full-path>" to show exact directory
  - No extraneous messages when no profile is active (quiet operation on base dir)
  - Update summary clearly indicates source location for components
  
  **Testing Criteria:**
  **Unit Tests:**
  - Message formatting includes correct profile name
  - Path display shows expanded absolute path
  
  **Integration Tests:**
  - Active profile message appears before component checking starts
  - Profile flag message appears when --profile used
  - No profile messages when operating on base directory
  - Full directory path is accurate and matches actual scan location

## Functional Requirements

### Core Functionality

- **FR-1**: The UpdateDetector SHALL accept a configurable base directory parameter in its constructor
- **FR-2**: The update command handler SHALL check for an active profile before initializing UpdateDetector
- **FR-3**: When an active profile exists, the update command SHALL use `~/.agent-smith/profiles/<profile-name>/` as the base directory
- **FR-4**: When no active profile exists, the update command SHALL use `~/.agent-smith/` as the base directory
- **FR-5**: The update command SHALL support a `--profile` flag that bypasses active profile detection

### Profile Detection

- **FR-6**: Profile detection SHALL use the existing ProfileManager.GetActiveProfile() method
- **FR-7**: Profile validation SHALL verify the profile directory exists before attempting updates
- **FR-8**: Profile resolution SHALL construct the correct path using `paths.GetProfileDir(profileName)`

### Update Command Variants

- **FR-9**: The `update all` command SHALL respect active profile configuration
- **FR-10**: The `update <type> <name>` command SHALL respect active profile configuration
- **FR-11**: Both command variants SHALL support the `--profile` flag
- **FR-12**: When `--profile` flag is provided, it SHALL take precedence over active profile

### User Feedback

- **FR-13**: The system SHALL display "Using active profile: <name>" when updating from active profile
- **FR-14**: The system SHALL display "Updating profile: <name>" when --profile flag is used
- **FR-15**: The system SHALL display "Updating components in: <path>" showing the full directory path
- **FR-16**: The system SHALL NOT display profile messages when operating on base directory with no active profile

### Backwards Compatibility

- **FR-17**: The system SHALL maintain existing behavior for users without profiles
- **FR-18**: The system SHALL not break existing lock file formats or metadata structures
- **FR-19**: The system SHALL continue to support updating base directory components when no profile is active

### Error Handling

- **FR-20**: The system SHALL display clear error when --profile specifies non-existent profile
- **FR-21**: The system SHALL validate base directory exists before scanning for components
- **FR-22**: The system SHALL handle cases where profile directory exists but contains no components

## Non-Goals

- No migration of components between base directory and profiles
- No automatic profile creation during update operations
- No bulk updates across multiple profiles simultaneously (except explicit --profile usage)
- No changes to the actual update/download logic or git operations
- No modifications to lock file format or metadata structure
- No UI changes to status command output (already shows profile information correctly)
- No changes to how components are linked or unlinked
- No automatic profile activation based on update context
- No profile-aware updates for uninstall command (out of scope for this feature)
