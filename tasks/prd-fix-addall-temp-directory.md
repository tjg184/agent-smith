# PRD: Fix AddAll temp Directory Parameter Passing

## Introduction

The `add-all` command in Agent Smith is experiencing performance degradation due to redundant repository cloning. Although the optimization was implemented to clone the repository only once, individual download methods are not properly using the provided temporary directory, causing each component type (skills, agents, commands) to re-clone the repository separately.

## Goals

- Eliminate redundant repository cloning in AddAll command
- Ensure proper usage of shared temporary directory across all component downloads
- Maintain existing functionality without breaking changes

## User Stories

- [ ] Story-001: As a user running add-all, I want the repository to be cloned only once so that the command completes faster and uses less bandwidth.

  **Acceptance Criteria:**
  - Repository is cloned exactly once during add-all execution
  - All component types (skills, agents, commands) use the same temporary directory
  - No breaking changes to existing CLI interface
  - Performance improvement measurable through timing

- [ ] Story-002: As a developer maintaining the codebase, I want the temp directory parameter passing to work correctly so that optimization functions as designed.

  **Acceptance Criteria:**
  - `providedRepoPath` parameter is correctly passed from AddAll to individual download methods
  - Individual download methods properly detect and use provided repository path
  - No redundant cloning occurs when valid temp directory is provided
  - Existing behavior preserved for direct download commands

## Functional Requirements

- FR-1: Fix the condition check `providedRepoPath[0] != ""` in individual download methods
- FR-2: Ensure AddAll passes temp directory correctly to downloadSkill, downloadAgent, and downloadCommand
- FR-3: Maintain backward compatibility for individual download commands without provided paths
- FR-4: Verify that all three component types (skills, agents, commands) use shared temp directory

## Non-Goals

- No changes to individual download command interfaces
- No new features or additional functionality
- No changes to component detection logic
- No modifications to error handling (keep existing approach)