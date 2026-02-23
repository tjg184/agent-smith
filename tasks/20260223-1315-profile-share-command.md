# PRD: Profile Share Command

**Created**: 2026-02-23 13:15 UTC

---

## Introduction

Implement a `profile share` command that generates plain text commands to recreate an agent-smith profile. This enables users to share their profile configurations with teammates, back up their profiles, or commit profile setups to version control. The output is a simple, transparent text file containing the exact `agent-smith` commands needed to recreate the profile from scratch.

## Goals

- Enable easy sharing of profile configurations between users
- Generate human-readable, copy-pasteable command scripts
- Support team standardization through shareable profile setups
- Provide transparent, auditable profile recreation process
- Integrate seamlessly with version control and dotfiles workflows

## User Stories

- [ ] Story-001: As a user, I want to generate commands to recreate my profile so that I can share it with teammates.

  **Acceptance Criteria:**
  - Command `agent-smith profile share <profile-name>` generates install commands
  - Output includes all components from the specified profile
  - Commands are in correct execution order (create, activate, install)
  - Output includes helpful comments explaining each section
  - Components from local file paths are skipped with warning
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile validation logic
  - Command generation for empty profiles
  - Command generation for single component
  - Command generation for multiple component types
  - Local path filtering logic
  
  **Integration Tests:**
  - Full workflow: create profile, install components, generate share commands
  - Verify generated commands are syntactically valid
  - Test with repo-sourced profiles
  - Test with user-created profiles

- [ ] Story-002: As a user, I want to save generated commands to a file so that I can commit them to version control.

  **Acceptance Criteria:**
  - `--output` flag saves commands to specified file path
  - File is created with appropriate permissions (0644)
  - Success message shows the output file path
  - Error handling for write permission issues
  - Without `--output` flag, commands print to stdout
  
  **Testing Criteria:**
  **Unit Tests:**
  - File writing logic with various paths
  - Permission error handling
  
  **Integration Tests:**
  - Save to file and verify contents match stdout output
  - Test with relative and absolute file paths
  - Test error handling for invalid paths

- [ ] Story-003: As a user, I want the output to include clear instructions so that recipients know how to use it.

  **Acceptance Criteria:**
  - Header includes profile name and generation date
  - Instructions at top explain how to use the commands
  - Comments group commands by component type with counts
  - Footer includes optional link command
  - Summary line shows total component count breakdown
  
  **Testing Criteria:**
  **Unit Tests:**
  - Header generation format
  - Comment generation for different component counts
  - Summary line formatting

- [ ] Story-004: As a user, I want components to be organized by type so that the output is easy to understand.

  **Acceptance Criteria:**
  - Skills grouped together with comment header
  - Agents grouped together with comment header
  - Commands grouped together with comment header
  - Each group shows count in comment
  - Empty component types are omitted from output
  
  **Testing Criteria:**
  **Unit Tests:**
  - Component grouping logic
  - Count calculation per type
  - Empty group filtering

- [ ] Story-005: As a developer, I want to add the ShareProfile method to the profile service so that it integrates with existing architecture.

  **Acceptance Criteria:**
  - `ShareProfile(name, outputPath string) error` method added to Service
  - Method validates profile exists before generation
  - Method delegates to helper functions for command generation
  - Method handles both file output and stdout
  - Method uses existing formatter for success messages
  
  **Testing Criteria:**
  **Unit Tests:**
  - Service method with valid profile
  - Service method with invalid profile
  - Service method with output file
  - Service method with stdout

- [ ] Story-006: As a developer, I want helper functions to generate commands from lock files so that the code is maintainable.

  **Acceptance Criteria:**
  - `generateProfileCommands` creates full output string
  - `generateComponentCommands` reads lock file and generates install commands
  - Lock file reading handles missing files gracefully
  - Source URL parsing skips local paths
  - Component names and URLs are properly formatted in commands
  
  **Testing Criteria:**
  **Unit Tests:**
  - Lock file parsing with various formats
  - Local path detection and filtering
  - Command string formatting
  - Empty lock file handling

- [ ] Story-007: As a user, I want to share the base installation profile so that I can backup my base components.

  **Acceptance Criteria:**
  - `agent-smith profile share base` generates commands for base installation
  - Output omits profile creation/activation for base
  - Base components use standard install commands without `--profile` flag
  - All other functionality works identically to named profiles
  
  **Testing Criteria:**
  **Unit Tests:**
  - Base profile detection logic
  - Command generation without profile flags
  
  **Integration Tests:**
  - Share base profile and verify output format

- [ ] Story-008: As a user, I want helpful error messages when sharing fails so that I can fix issues quickly.

  **Acceptance Criteria:**
  - Clear error when profile does not exist
  - Helpful message when profile is empty
  - Error message for file write failures includes path
  - Validation errors show expected format
  
  **Testing Criteria:**
  **Unit Tests:**
  - Error message formatting
  - Non-existent profile handling
  
  **Integration Tests:**
  - Attempt to share non-existent profile
  - Attempt to share empty profile
  - Attempt to write to invalid path

## Functional Requirements

- FR-1: The system SHALL generate install commands from profile lock files
- FR-2: The system SHALL output commands in correct execution order (create, activate, install, link)
- FR-3: The system SHALL skip components installed from local file paths
- FR-4: The system SHALL include profile creation and activation commands for named profiles
- FR-5: The system SHALL omit profile-specific commands when sharing base installation
- FR-6: The system SHALL support output to both stdout and file via `--output` flag
- FR-7: The system SHALL include comment headers with component type and counts
- FR-8: The system SHALL validate profile exists before generating commands
- FR-9: The system SHALL use singular component types in install commands (skill, agent, command)
- FR-10: The system SHALL preserve source URLs exactly as stored in lock files

## Non-Goals

- No automatic execution of generated commands (user must run manually)
- No profile import command in initial implementation
- No version pinning or commit hash references in initial version
- No JSON or YAML output format options
- No clipboard integration
- No QR code generation
- No remote sharing service or registry
- No automatic detection of which components were manually added vs installed
- No validation that generated commands will actually work (URLs might be invalid/changed)
- No filtering or selection of which components to include (all or nothing)
