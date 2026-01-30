# PRD: Bulk Component Installation

**Created**: 2026-01-30 17:26 UTC

---

## Introduction

Enhance the `agent-smith install` command to support installing multiple skills, agents, or commands from a single repository in one operation. Currently, users can only install components one-by-one or use "all" to install everything. This feature allows selective bulk installation (e.g., `agent-smith install skills https://github.com/anthropics/skills` to install all skills from that repository).

## Goals

- Enable bulk installation of multiple components of the same type from a repository
- Support selective installation by component type (skills, agents, commands)
- Maintain compatibility with existing single-component installation workflow
- Provide clear feedback on what components were discovered and installed
- Handle partial failures gracefully (some components install, others fail)

## User Stories

- [ ] Story-001: As a user, I want to install all skills from a repository URL so that I can quickly add multiple related skills without installing each one individually.

  **Acceptance Criteria:**
  - Command syntax: `agent-smith install skills <repo-url>`
  - Clones the repository to a temporary location
  - Discovers all valid skills in the repository structure
  - Installs each discovered skill to `~/.agent-smith/skills/` (or active profile)
  - Displays list of discovered components before installation
  - Reports success/failure for each component installed
  
  **Testing Criteria:**
  **Unit Tests:**
  - Repository URL parsing and validation
  - Component type filtering logic
  - Installation path resolution with profiles
  
  **Integration Tests:**
  - Git clone operations with various repository formats
  - Component discovery across different directory structures
  - Installation to both default and profile directories
  
  **Component Browser Tests:**
  - CLI output formatting for component lists
  - Progress indicator display during installation
  - Error message display for failed installations

- [ ] Story-002: As a user, I want to install all agents from a repository URL so that I can deploy multiple related agents at once.

  **Acceptance Criteria:**
  - Command syntax: `agent-smith install agents <repo-url>`
  - Discovers all valid agents in the repository
  - Installs each agent to `~/.agent-smith/agents/` (or active profile)
  - Skips non-agent directories (skills, commands)
  - Handles monorepo containers appropriately
  
  **Testing Criteria:**
  **Unit Tests:**
  - Agent-specific metadata validation
  - Type filtering to exclude skills/commands
  
  **Integration Tests:**
  - Multi-agent repository installations
  - Monorepo structure handling
  - Profile-aware installation paths
  
  **Component Browser Tests:**
  - Agent list display formatting
  - Installation progress for multiple agents
  - Summary report display

- [ ] Story-003: As a user, I want to install all commands from a repository URL so that I can add multiple custom commands in bulk.

  **Acceptance Criteria:**
  - Command syntax: `agent-smith install commands <repo-url>`
  - Discovers all valid commands in the repository
  - Installs each command to `~/.agent-smith/commands/` (or active profile)
  - Validates command metadata before installation
  - Reports any invalid or malformed commands
  
  **Testing Criteria:**
  **Unit Tests:**
  - Command metadata validation logic
  - Installation path construction
  
  **Integration Tests:**
  - Multiple command installations from single repository
  - Command validation and error handling
  
  **Component Browser Tests:**
  - Command discovery feedback display
  - Validation error messages
  - Installation summary formatting

- [ ] Story-004: As a user, I want clear feedback during bulk installation so that I know which components were successfully installed and which failed.

  **Acceptance Criteria:**
  - Display count of discovered components before installation
  - Show progress indicator during installation (e.g., "Installing 3/5...")
  - Report success/failure for each component individually
  - Display summary at end (e.g., "5 installed, 2 skipped, 1 failed")
  - Exit with non-zero code if any installations fail
  
  **Testing Criteria:**
  **Unit Tests:**
  - Progress calculation logic
  - Summary statistics aggregation
  
  **Integration Tests:**
  - Partial failure handling (some succeed, some fail)
  - Exit code verification
  
  **Component Browser Tests:**
  - Progress indicator rendering
  - Summary table display
  - Color-coded status messages

- [ ] Story-005: As a user, I want bulk installation to respect my active profile so that components are installed to the correct profile directory.

  **Acceptance Criteria:**
  - Checks for active profile via ProfileManager
  - If profile active: installs to `~/.agent-smith/profiles/<profile>/<type>/`
  - If no profile: installs to `~/.agent-smith/<type>/`
  - Uses same profile logic as existing single-component installation
  - Displays target directory in output for transparency
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile detection logic
  - Installation path resolution with profiles
  
  **Integration Tests:**
  - Installation with active profile
  - Installation without active profile
  - Profile switching between installations
  
  **Component Browser Tests:**
  - Target directory display in output
  - Profile-aware path verification

- [ ] Story-006: As a user, I want bulk installation to handle existing components gracefully so that I don't accidentally overwrite my customizations.

  **Acceptance Criteria:**
  - Check if component already exists before installation
  - Skip existing components by default (with message)
  - Support `--force` flag to overwrite existing components
  - Support `--upgrade` flag to only update if source is newer
  - Count skipped components in final summary
  
  **Testing Criteria:**
  **Unit Tests:**
  - Existing component detection logic
  - Force/upgrade flag handling
  
  **Integration Tests:**
  - Skip behavior verification
  - Force overwrite behavior
  - Upgrade with version comparison
  
  **Component Browser Tests:**
  - Skip message display
  - Flag-based behavior variations
  - Warning messages for overwrites

- [ ] Story-007: As a user, I want the option to automatically link components after bulk installation so that they're immediately usable.

  **Acceptance Criteria:**
  - Support `--link` flag to auto-link after installation
  - Links to all configured targets (OpenCode, Claude Code, etc.)
  - Displays linking status after installation summary
  - If `--link` not provided, show hint about running `agent-smith link`
  - Handles linking failures gracefully (component installed but not linked)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Link flag parsing
  - Post-installation link invocation
  
  **Integration Tests:**
  - End-to-end install + link workflow
  - Link failure handling
  - Multi-target linking verification
  
  **Component Browser Tests:**
  - Link status display formatting
  - Hint message display
  - Link failure error messages

## Functional Requirements

- FR-1: The system SHALL support bulk installation syntax: `agent-smith install <type> <repo-url>` where type is "skills", "agents", or "commands"
- FR-2: The system SHALL clone the specified repository to a temporary directory for component discovery
- FR-3: The system SHALL discover all valid components of the specified type within the repository structure
- FR-4: The system SHALL validate each discovered component's metadata before installation
- FR-5: The system SHALL install valid components to `~/.agent-smith/<type>/<name>/` or `~/.agent-smith/profiles/<profile>/<type>/<name>/` if a profile is active
- FR-6: The system SHALL display discovered component count and installation progress during execution
- FR-7: The system SHALL report success, failure, or skip status for each component
- FR-8: The system SHALL provide a summary report showing total installed, skipped, and failed counts
- FR-9: The system SHALL skip existing components by default unless `--force` or `--upgrade` flags are provided
- FR-10: The system SHALL support `--link` flag to automatically link installed components to configured targets
- FR-11: The system SHALL exit with non-zero code if any component installations fail
- FR-12: The system SHALL clean up temporary clone directory after installation completes

## Non-Goals

- No automatic dependency resolution between components (manual installation required)
- No interactive component selection from repository (install all of specified type)
- No version pinning or lockfile generation (always installs latest from repository)
- No automatic updates or background syncing of installed components
- No registry/marketplace integration (direct Git repository URLs only)
- No support for installing multiple types in one command (must specify skills, agents, or commands)
- No automatic profile creation or switching during installation
