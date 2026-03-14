# PRD: Agent-Smith Profiles

## Introduction

The "Agent-Smith Profiles" feature introduces the ability to manage and switch between distinct sets of agents, commands, and skills. This feature is designed for users who work across different contexts (e.g., "coding", "writing", "devops") and need to swap their toolset efficiently. Profiles are strictly opt-in, stored locally, and apply globally to all targets (OpenCode, Claude Code, etc.).

## Goals

- **Context Switching**: Allow users to easily switch between different configurations of tools.
- **Opt-in Usage**: Ensure the feature does not interfere with the existing workflow for users who do not use profiles.
- **Global Application**: Profile activation applies to the entire system (all targets).
- **Clean State Management**: Ensure switching profiles cleanly removes the previous profile's links before adding the new ones.
- **Backward Compatibility**: Existing direct linking workflows must remain unchanged.

## User Stories

- [x] Story-001: As a user, I want to define profiles in a specific directory so that the system can recognize them.
  **Acceptance Criteria:**
  - Profiles are located in `~/.agent-smith/profiles/<profile_name>/`.
  - Each profile directory mirrors the standard structure (agents/, commands/, skills/).
  - The system ignores incomplete or malformed profile directories gracefully.
  
  **Testing Criteria:**
  **Unit Tests:**
  - Validate directory structure parsing logic.
  - Test handling of empty or missing profile directories.
  
  **Integration Tests:**
  - Verify filesystem reads from `~/.agent-smith/profiles/`.

- [x] Story-002: As a user, I want to list available profiles so that I can see what configurations are available.
  **Acceptance Criteria:**
  - New command `agent-smith profiles list` (or similar).
  - output shows the names of all valid profiles found in `~/.agent-smith/profiles/`.
  - Indicates which profile (if any) is currently active.
  
  **Testing Criteria:**
  **Unit Tests:**
  - formatting of the list output.
  - Identification of the active profile from state.
  
  **Integration Tests:**
  - CLI command execution returns correct list based on filesystem.

- [x] Story-003: As a user, I want to activate a specific profile so that its tools replace the currently active ones.
  **Acceptance Criteria:**
  - Command `agent-smith profiles activate <name>`.
  - If another profile is active, it is automatically deactivated (unlinked) first.
  - Validates that the target profile exists.
  - Symlinks the profile's contents (agents, skills, commands) to the source of truth locations.
  - Updates a state file (e.g., `~/.agent-smith/active_profile`) to reflect the new active profile.
  - **Crucial**: Activation is isolated; it does not merge with the previous profile.
  
  **Testing Criteria:**
  **Unit Tests:**
  - State update logic.
  - Validation of profile existence.
  
  **Integration Tests:**
  - Full cycle: Activate A -> Check Links -> Activate B -> Check Links (A gone, B present).
  - Verify `~/.agent-smith/active_profile` updates.

- [x] Story-004: As a user, I want to deactivate the current profile so that I can return to a base state.
  **Acceptance Criteria:**
  - Command `agent-smith profiles deactivate`.
  - Removes all symlinks associated with the currently active profile.
  - Clears the active profile state.
  - Does not affect manually installed (non-profile) tools if they exist (though design implies profiles might manage everything, "base state" assumes no profile active).
  
  **Testing Criteria:**
  **Unit Tests:**
  - State clearing logic.
  
  **Integration Tests:**
  - Verify all profile-specific symlinks are removed upon deactivation.

- [x] Story-005: As a user, I want to see the current status so that I know which profile is active.
  **Acceptance Criteria:**
  - Command `agent-smith profiles status`.
  - Displays the name of the active profile or "None".
  - Optionally lists the count of active agents/skills from the profile.
  
  **Testing Criteria:**
  **Unit Tests:**
  - Output formatting.
  
  **Integration Tests:**
  - Verify output matches the `active_profile` state file.

## Functional Requirements

- **FR-1 Storage**: Profile definitions must be stored in `~/.agent-smith/profiles/`.
- **FR-2 Global Scope**: Changing a profile affects the global environment for all consumers of the `~/.agent-smith` bin/source.
- **FR-3 Linking Mechanism**: The system must use symlinks to "activate" tools from a profile into the main execution path.
- **FR-4 Auto-Cleanup**: Before linking a new profile, the system must identify and unlink the artifacts of the previously active profile.
- **FR-5 Persistence**: The active profile selection must persist across sessions (via a state file).
- **FR-6 Isolation**: Activation must be exclusive (Profile A OR Profile B), not additive (Profile A + Profile B), unless "None" is selected.

## Non-Goals

- **Default Profile**: The system will not automatically create a "default" profile.
- **Additive Mode**: Users cannot activate multiple profiles simultaneously.
- **UI Management**: No graphical user interface for managing profiles; CLI only.
- **Migration**: No automatic migration of existing flat structures into profiles.

