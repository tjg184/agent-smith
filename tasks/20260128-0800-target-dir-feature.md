# PRD: Add Custom Target Directory Support to `install all`

## Introduction

Add a `--target-dir` flag to the `agent-smith install all` command that allows installing components to a custom base directory instead of the default `~/.agents/`. Custom directories will be standalone and independent from the managed `~/.agents/` ecosystem, enabling project-local installations, isolated testing, and offline distribution packaging.

## Goals

- Enable installing components to arbitrary directories
- Support relative paths, absolute paths, and tilde expansion
- Auto-create directory structure (skills/, agents/, commands/)
- Store lock files in the target directory
- Maintain backward compatibility with existing behavior
- Keep custom directories isolated from link/update/profile commands

## User Stories

- [x] Story-001: As a developer working on a specific project, I want to install AI components directly into my project directory so they're version-controlled with my code.

  **Acceptance Criteria:**
  - Add `--target-dir` flag to `agent-smith install all` command with short form `-t`
  - Support relative paths (e.g., `./tools`), absolute paths (e.g., `/opt/components`), and tilde expansion (e.g., `~/project/tools`)
  - Resolve paths correctly and convert relative paths to absolute paths
  - Create target directory structure with subdirectories: `skills/`, `agents/`, `commands/`
  - Install components to appropriate subdirectories within target directory
  - Command example: `./agent-smith install all https://github.com/org/tools --target-dir ./tools` creates `./tools/skills/`, `./tools/agents/`, `./tools/commands/`

  **Testing Criteria:**
  **Unit Tests:**
  - Path resolution logic tests (relative, absolute, tilde expansion)
  - Directory creation validation tests
  - Path normalization and sanitization tests

  **Integration Tests:**
  - Full installation flow with custom target directory
  - Component placement in correct subdirectories
  - Multiple path format handling tests

  **Component Browser Tests:**
  - CLI flag parsing and validation
  - Help text display for new flag
  - Error message clarity for invalid paths

- [x] Story-002: As a component author, I want to test components in isolation without affecting my main `~/.agents/` installation.

  **Acceptance Criteria:**
  - Custom target directories are completely isolated from `~/.agents/` directory
  - No modifications to existing `~/.agents/` when using `--target-dir`
  - Lock files (`.skill-lock.json`, `.agent-lock.json`, `.command-lock.json`) stored in target directory root
  - Lock files contain source repository, commit hash, and metadata
  - Custom directories are NOT managed by `link`, `update`, or `profile` commands
  - Clear isolation documentation in help text and README

  **Testing Criteria:**
  **Unit Tests:**
  - Lock file creation and structure validation
  - Isolation from default directory tests
  - Lock file content validation tests

  **Integration Tests:**
  - Installation to temporary directory
  - Verification of no side effects to `~/.agents/`
  - Lock file persistence and content accuracy

  **Component Browser Tests:**
  - Complete isolation workflow testing
  - Clean installation and removal testing
  - Directory independence verification

- [ ] Story-003: As a systems administrator, I want to package components for offline distribution to air-gapped systems.

  **Acceptance Criteria:**
  - Components installed to custom directory are fully self-contained
  - All necessary files and subdirectories created in target directory
  - Target directory can be archived (tar/zip) and distributed
  - No external dependencies on `~/.agents/` or other system directories
  - Clear documentation on using custom directories for distribution
  - Support for very long paths and paths with special characters (spaces, unicode)

  **Testing Criteria:**
  **Unit Tests:**
  - Path validation for edge cases (long paths, special chars, unicode)
  - Directory structure completeness validation
  - Self-containment verification tests

  **Integration Tests:**
  - Full distribution workflow (install → archive → extract → use)
  - Cross-platform path handling tests
  - Edge case path handling (spaces, unicode, very long paths)

  **Component Browser Tests:**
  - Complete distribution packaging workflow
  - Archive and extraction validation
  - Portability verification across different systems

- [ ] Story-004: As a user, I want clear error handling when using invalid target directories so I understand what went wrong.

  **Acceptance Criteria:**
  - Clear error message when target path exists as a file (not directory)
  - Clear error message when lacking write permissions to target path
  - Helpful error message for invalid path formats
  - Automatic directory hierarchy creation for non-existent parent directories (e.g., `./a/b/c/components`)
  - Proper symlink resolution in paths
  - OS disk space errors propagated with clear messages

  **Testing Criteria:**
  **Unit Tests:**
  - Error handling logic for each error condition
  - Error message clarity and helpfulness tests
  - Permission checking validation

  **Integration Tests:**
  - Permission denied scenarios
  - Target exists as file scenarios
  - Parent directory creation validation
  - Symlink resolution tests

  **Component Browser Tests:**
  - User experience testing for error scenarios
  - Error message readability and actionability
  - Edge case error handling (disk full, invalid paths)

- [ ] Story-005: As a developer, I want backward compatibility with existing commands so my current workflows continue to work without changes.

  **Acceptance Criteria:**
  - `install all` without `--target-dir` flag installs to `~/.agents/` (current behavior unchanged)
  - Individual `install skill/agent/command` commands remain unchanged
  - `--profile` flag for individual installs continues to work as before
  - Link, update, and profile commands operate only on `~/.agents/` and profiles, not custom directories
  - All existing tests continue to pass
  - No breaking changes to existing CLI interface

  **Testing Criteria:**
  **Unit Tests:**
  - Default behavior validation tests
  - Backward compatibility regression tests
  - Flag interaction tests

  **Integration Tests:**
  - Full workflow tests with and without flag
  - Profile command isolation tests
  - Existing functionality preservation tests

  **Component Browser Tests:**
  - Existing user workflow validation
  - Regression testing for all existing features
  - Cross-feature compatibility verification

## Functional Requirements

- FR-1: The system MUST add a `--target-dir` (short form `-t`) flag to the `agent-smith install all` command
- FR-2: The system MUST support relative paths, absolute paths, and tilde expansion for target directory paths
- FR-3: The system MUST resolve paths correctly (tilde → home dir, relative → absolute, symlink resolution)
- FR-4: The system MUST auto-create target directory if it doesn't exist, including parent directories
- FR-5: The system MUST create subdirectories `skills/`, `agents/`, `commands/` within target directory
- FR-6: The system MUST install components to appropriate subdirectories: `<target-dir>/skills/<name>/`, `<target-dir>/agents/<name>/`, `<target-dir>/commands/<name>/`
- FR-7: The system MUST store lock files (`.skill-lock.json`, `.agent-lock.json`, `.command-lock.json`) in target directory root
- FR-8: Lock files MUST contain source repository URL, commit hash, and installation metadata
- FR-9: Custom target directories MUST be isolated from `~/.agents/` directory (no cross-contamination)
- FR-10: Custom target directories MUST NOT be managed by `link`, `update`, or `profile` commands
- FR-11: The system MUST provide clear error messages for invalid paths, permission errors, and target-is-file errors
- FR-12: The system MUST handle edge cases: empty string (use default), paths with spaces, unicode paths, very long paths
- FR-13: The system MUST maintain backward compatibility: `install all` without flag installs to `~/.agents/`
- FR-14: The system MUST update help text and CLI documentation to describe new flag and its behavior


## Non-Goals (Out of Scope)

- Integration with `link`, `update`, or `profile` commands (custom dirs are standalone)
- Adding `--target-dir` to individual install commands (skill/agent/command)
- Managing multiple custom directories from a central registry
- Auto-discovery of custom directory installations
- Adding `--target-dir` to individual install commands
- Support installing to profiles with `install all --profile <name>`
- Registry/discovery system for multiple custom directories
- Integration of custom dirs with link/update commands
- Validation that target directory is a valid agent-smith structure
- `agent-smith list-installations` to show all known installations
