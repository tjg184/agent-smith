# PRD: Add Claude Code Support

## Introduction
Expand `agent-smith` to support Claude Code in addition to OpenCode. This feature will allow developers using both environments to seamlessly manage and link their skills, agents, and commands to the appropriate global directories (`~/.claude/` for Claude Code and `~/.config/opencode/` for OpenCode).

## Goals
- Establish a target abstraction layer to handle differences between OpenCode and Claude Code.
- Auto-detect if Claude Code is installed on the system.
- Implement linking of skills, agents, and commands to the global Claude Code directories (`~/.claude/skills/`, `~/.claude/agents/`, `~/.claude/commands/`).
- Introduce new CLI flags `--target` and `--all-agents` to control linking behavior.
- Update the `status` command to visualize linking status across available targets.
- Ensure default behavior links to all detected agents (OpenCode and/or Claude Code).

## User Stories

- [ ] Story-001: As a developer, I want agent-smith to understand different target environments so that I can support both OpenCode and Claude Code.
  **Acceptance Criteria:**
  - Create an abstraction for "Target" (e.g., Enum or Class) encompassing OpenCode and Claude Code.
  - Define paths and configuration constants for Claude Code (e.g., `~/.claude/`).
  - Metadata tracking supports multiple targets for a single resource.

  **Testing Criteria:**
  **Unit Tests:**
  - Verify Target abstraction correctly identifies OpenCode and Claude Code properties.
  - Test path resolution logic for both targets.
  **Integration Tests:**
  - Verify configuration loading respects target definitions.

- [ ] Story-002: As a user, I want agent-smith to auto-detect if Claude Code is installed so that I don't have to configure it manually.
  **Acceptance Criteria:**
  - Check for the existence of `~/.claude` directory or specific configuration files.
  - Store the detected status of Claude Code availability.
  - Default operations should include Claude Code if detected.

  **Testing Criteria:**
  **Unit Tests:**
  - Mock filesystem to test detection logic (installed vs. not installed).
  **Integration Tests:**
  - run detection sequence on the actual environment and verify result matches expected state.

- [ ] Story-003: As a user, I want to link my tools to the global Claude Code directories so that I can use them within Claude Code.
  **Acceptance Criteria:**
  - Implement linking logic for Skills to `~/.claude/skills/`.
  - Implement linking logic for Agents to `~/.claude/agents/`.
  - Implement linking logic for Commands to `~/.claude/commands/`.
  - Ensure symlink creation handles existing files/links correctly (overwrite/skip prompting).
  - **Scope:** Global linking only (no project-local support).

  **Testing Criteria:**
  **Unit Tests:**
  - Verify link path generation for skills, agents, and commands.
  - Test conflict handling logic (what happens if link exists).
  **Integration Tests:**
  - Perform actual linking to a temporary/mock `~/.claude` directory and verify symlinks are created.

- [ ] Story-004: As a user, I want to control which target I am linking to using CLI flags so that I can be specific or broad in my operations.
  **Acceptance Criteria:**
  - Add `--target` flag accepting values like `opencode`, `claude`, or `all`.
  - Add `--all-agents` flag to explicitly request linking to all detected targets.
  - Default behavior (no flags): Link to ALL detected and available targets.

  **Testing Criteria:**
  **Unit Tests:**
  - Parse CLI arguments and verify correct target selection.
  - Verify default behavior selects all detected targets.
  **Component Browser Tests:**
  - N/A (CLI tool)

- [ ] Story-005: As a user, I want to see the status of my links across all targets so that I know what is installed where.
  **Acceptance Criteria:**
  - Update `status` command output to show columns or sections for each target.
  - clearly indicate if a resource is linked to OpenCode, Claude Code, or both.
  - Show "Not Installed" or similar status if a target is not available on the system.

  **Testing Criteria:**
  **Unit Tests:**
  - Format status output strings/tables with mock data for multiple targets.
  **Integration Tests:**
  - Run `status` command with mixed linking states and verify output correctness.

## Functional Requirements
- FR-1: The system must define a `Target` interface/type that includes name, description, and root configuration paths.
- FR-2: The system must automatically scan `~/.claude` to determine Claude Code presence.
- FR-3: The `link` command must accept a `--target` argument.
- FR-4: The `link` command must accept a `--all-agents` argument.
- FR-5: Symlinks for Claude Code must be created in `~/.claude/{type}/`.
- FR-6: The `status` command must display a matrix or list indicating link status per target.

## Non-Goals
- Project-local configuration or linking for Claude Code (Global only for now).
- Installing Claude Code itself (must be pre-installed by user).
- Managing Claude Code configuration files beyond simple linking of tools.
