# PRD: Target Filter Bug Fix for Link Command

**Created**: 2026-01-30 02:08 UTC

---

## Introduction

Fix a critical bug in the `agent-smith link` command where the `--target/-t` flag is ignored, causing components to be linked to all detected targets instead of the specified target. This breaks the user's ability to selectively link components to specific tools (OpenCode, Claude Code, etc.).

## Goals

- Ensure `--target` flag correctly filters targets before linking
- Prevent unintended cross-tool contamination when linking
- Maintain backward compatibility for "all" targets behavior
- Provide clear error messages when invalid targets are specified

## User Stories

- [ ] Story-001: As a developer, I want the `--target opencode` flag to link only to OpenCode so that Claude Code is not affected.

  **Acceptance Criteria:**
  - Running `./agent-smith link skills -t opencode` links only to `~/.config/opencode/skills/`
  - Claude Code directory `~/.claude/skills/` remains empty
  - Success message shows only opencode as linked target
  - Link status confirms skills are linked to OpenCode only
  
  **Testing Criteria:**
  **Unit Tests:**
  - Test `NewComponentLinkerWithFilter` correctly filters targets when given "opencode"
  - Test `NewComponentLinkerWithFilter` correctly filters targets when given "claudecode"
  - Test `NewComponentLinkerWithFilter` returns all targets when given empty string
  - Test `NewComponentLinkerWithFilter` returns all targets when given "all"
  - Test `getTargetNames` helper returns correct target name array
  
  **Integration Tests:**
  - Not required for this fix (unit tests are sufficient)
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

- [ ] Story-002: As a developer, I want the `--target claudecode` flag to link only to Claude Code so that OpenCode is not affected.

  **Acceptance Criteria:**
  - Running `./agent-smith link agents -t claudecode` links only to `~/.claude/agents/`
  - OpenCode directory `~/.config/opencode/agents/` remains empty
  - Success message shows only claudecode as linked target
  - Link status confirms agents are linked to Claude Code only
  
  **Testing Criteria:**
  **Unit Tests:**
  - Test filter logic with "claudecode" parameter
  - Test linker only receives claudecode target
  
  **Integration Tests:**
  - Not required for this fix
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

- [ ] Story-003: As a developer, I want clear error messages when I specify an invalid target so I know what targets are available.

  **Acceptance Criteria:**
  - Running `./agent-smith link skills -t invalid` returns error with available targets
  - Error message format: "target 'invalid' not found. Available targets: [opencode, claudecode]"
  - Command exits with non-zero status code
  - No linking occurs when invalid target specified
  
  **Testing Criteria:**
  **Unit Tests:**
  - Test `NewComponentLinkerWithFilter` returns error for non-existent target
  - Test error message includes list of available targets
  - Test `getTargetNames` helper is called in error path
  
  **Integration Tests:**
  - Not required for this fix
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

- [ ] Story-004: As a developer, I want the default behavior (no flag) to link to all targets so I can quickly set up all tools.

  **Acceptance Criteria:**
  - Running `./agent-smith link skills` (no `-t` flag) links to all detected targets
  - Both OpenCode and Claude Code receive skills
  - Success messages show all targets
  - Link status confirms skills are linked to all targets
  
  **Testing Criteria:**
  **Unit Tests:**
  - Test `NewComponentLinkerWithFilter` with empty string returns all targets
  - Test default behavior matches "all" behavior
  
  **Integration Tests:**
  - Not required for this fix
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

## Functional Requirements

- FR-1: The system SHALL filter targets in `NewComponentLinkerWithFilter` based on the `targetFilter` parameter before passing to `linker.NewComponentLinker`
- FR-2: The system SHALL return all detected targets when `targetFilter` is empty string or "all"
- FR-3: The system SHALL return only the matching target when `targetFilter` specifies a valid target name
- FR-4: The system SHALL return an error with available target names when `targetFilter` specifies an invalid target
- FR-5: The system SHALL implement a `getTargetNames` helper function to extract target names for error reporting
- FR-6: The system SHALL preserve existing behavior where omitting `--target` flag links to all detected targets

## Non-Goals

- No changes to the unlink command (already has filter logic)
- No changes to link status command (already displays all targets)
- No migration of existing links
- No changes to target detection logic
- No new flags or command options
- No documentation updates (fix is self-explanatory)

## Technical Implementation Notes

### Root Cause
The bug exists in `main.go:150-189` where `NewComponentLinkerWithFilter` receives `targetFilter` parameter but ignores it, passing all detected targets to the linker.

### Solution
Modify `NewComponentLinkerWithFilter` to:
1. Detect all available targets
2. Filter based on `targetFilter` parameter
3. Return error if specified target doesn't exist
4. Pass only filtered targets to `linker.NewComponentLinker`

### Code Changes
- **File**: `main.go`
- **Function**: `NewComponentLinkerWithFilter`
- **Lines**: 150-189
- **New Helper**: Add `getTargetNames` function for error messages

### Testing Strategy
- Unit tests for filter logic
- Unit tests for error handling
- Manual verification of CLI behavior
