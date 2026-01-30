# PRD: Fix Skill Installation Name Filtering

**Created**: 2026-01-30 18:07 UTC

---

## Introduction

Fix a bug in the skill installation logic where specifying a skill name in `agent-smith install skill <repo-url> <skill-name>` installs all skills from a multi-skill repository instead of only the requested skill. When a user provides a specific component name, the system should install only that component or return an error if not found, never install all components.

## Goals

- Install only the specifically requested skill when a name is provided
- Return a clear error message if the requested skill is not found in the repository
- Match the existing correct behavior already implemented in agent and command installers
- Maintain backward compatibility for single-skill repositories and direct downloads
- Provide helpful feedback listing available skills when the requested name is not found

## User Stories

- [x] Story-001: As a user, when I run `agent-smith install skill <repo> <skill-name>` on a repository with multiple skills, I want only the named skill installed so that I don't get unwanted skills.

  **Acceptance Criteria:**
  - Command syntax: `agent-smith install skill <repo-url> <skill-name>`
  - System detects all skills in the repository
  - System searches for a skill matching the provided name
  - If match found: installs ONLY that skill to the target directory
  - If no match: returns error without installing any skills
  - Single-skill repositories continue to work as before (no regression)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Name matching logic with exact matches
  - Component filtering from detected components list
  - Error path when no match found
  
  **Integration Tests:**
  - Multi-skill repository with matching name
  - Multi-skill repository with non-matching name
  - Single-skill repository (regression test)
  - Installation to custom target directory with `-t` flag
  
  **Component Browser Tests:**
  - CLI output verification for successful single-skill install
  - Error message display for non-matching names
  - Directory structure verification (no monorepo nesting)

- [x] Story-002: As a user, when I specify a skill name that doesn't exist in the repository, I want a clear error message listing available skills so that I know what options are valid.

  **Acceptance Criteria:**
  - Error message format: "skill '<name>' not found in repository. Available skills: <list>"
  - Lists all detected skill names in comma-separated format
  - Does not install any skills when name doesn't match
  - Cleans up any created directories on error
  - Returns non-zero exit code
  
  **Testing Criteria:**
  **Unit Tests:**
  - Error message formatting
  - Available skills list generation
  - Directory cleanup on error
  
  **Integration Tests:**
  - End-to-end flow with non-matching name
  - Verify no files installed on error
  - Exit code verification
  
  **Component Browser Tests:**
  - Error message display format
  - Available skills list readability
  - No leftover directories after error

- [x] Story-003: As a user, I want the skill installer to behave consistently with agent and command installers so that I have a predictable experience across all component types.

  **Acceptance Criteria:**
  - Skill installation logic matches agent installation logic (agent.go:186-225)
  - Skill installation logic matches command installation logic (command.go:168-225)
  - Same name matching pattern used across all three component types
  - Same error handling pattern for non-matching names
  - Same single-component optimization logic
  
  **Testing Criteria:**
  **Unit Tests:**
  - Name matching logic consistency across downloaders
  - Error message format consistency
  
  **Integration Tests:**
  - Parallel testing of skill, agent, and command installers
  - Behavior verification across all component types
  
  **Component Browser Tests:**
  - User experience consistency verification
  - Error message format comparison

- [x] Story-004: As a user, I want the installation to maintain proper directory structure when installing a single skill from a multi-skill repository so that I don't get nested monorepo directories.

  **Acceptance Criteria:**
  - Single skill installed directly to target directory (no nesting)
  - Target directory structure: `~/.agent-smith/skills/<skill-name>/` contains skill files directly
  - No monorepo subdirectory structure created when single skill requested
  - Lock file created correctly for single-skill installation
  - Success message displays correct skill name
  
  **Testing Criteria:**
  **Unit Tests:**
  - Directory path construction logic
  - Lock file metadata generation
  
  **Integration Tests:**
  - Directory structure verification after installation
  - Lock file content verification
  - Multiple installations to verify consistency
  
  **Component Browser Tests:**
  - File system structure inspection
  - Success message verification
  - Target directory flag behavior with `-t`

- [x] Story-005: As a user, I want the installer to clean up properly on errors so that I don't have empty or partial directories left behind.

  **Acceptance Criteria:**
  - Call `os.RemoveAll(skillDir)` before returning error for non-matching names
  - No empty skill directories left after error
  - No partial installations when name doesn't match
  - Temporary clone directories cleaned up (existing behavior maintained)
  - Clear error message explains why installation failed
  
  **Testing Criteria:**
  **Unit Tests:**
  - Cleanup function invocation verification
  - Error path directory removal
  
  **Integration Tests:**
  - File system state verification after error
  - No leftover directories after failed install
  - Temporary directory cleanup verification
  
  **Component Browser Tests:**
  - Directory existence checks after error
  - Clean state verification for retry attempts

## Functional Requirements

- FR-1: The system SHALL search detected skill components for an exact name match with the provided skillName parameter
- FR-2: The system SHALL install ONLY the matching skill when exactly one name match is found, regardless of how many skills exist in the repository
- FR-3: The system SHALL return an error when the provided skillName does not match any detected skill components
- FR-4: The error message SHALL include the list of available skill names in comma-separated format
- FR-5: The system SHALL clean up any created directories when returning an error for non-matching names
- FR-6: The system SHALL maintain existing behavior for single-skill repositories (install without additional name checking)
- FR-7: The system SHALL maintain existing behavior for repositories with no detected skills (fall back to direct download)
- FR-8: The name matching logic SHALL use case-sensitive exact string comparison matching agent and command installer behavior
- FR-9: The system SHALL install skill files directly to the target directory without creating nested monorepo subdirectories when a single skill is requested
- FR-10: The system SHALL create lock files and success messages only after successful installation

## Non-Goals

- No interactive prompts or component selection menus (return error instead)
- No fuzzy name matching or typo suggestions (exact match only)
- No support for installing multiple skills in one command (use bulk install feature for that)
- No changes to agent or command installers (they already work correctly)
- No changes to the direct download fallback behavior
- No changes to profile or target directory handling logic
- No automatic installation of "all skills" when name doesn't match (explicitly error instead)
