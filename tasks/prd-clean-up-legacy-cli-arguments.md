# PRD: Clean Up Legacy CLI Arguments

## Introduction

The agent-smith CLI has accumulated legacy commands that maintain backward compatibility with older command syntax. These legacy commands include `add-skill`, `add-agent`, `add-command`, `add-all`, `link-legacy`, `auto-link`, `list-links`, and `link-status`. Additionally, legacy metadata file support exists for `.{type}-metadata.json` files alongside the modern lock file format.

This PRD outlines the complete removal of these legacy components to simplify the codebase, reduce maintenance burden, and enforce the modern command structure.

## Goals

- Remove all legacy CLI commands from the codebase
- Eliminate support for legacy metadata file formats
- Simplify cmd/root.go by removing backward compatibility code
- Standardize on modern command structure (e.g., `add skill` instead of `add-skill`)
- Update documentation to reflect only modern command syntax
- Create comprehensive tests to validate legacy code removal

## User Stories

- [ ] Story-001: As a developer, I want to remove the add-skill legacy command so that the codebase only supports the modern 'add skill' syntax.

  **Acceptance Criteria:**
  - Remove add-skill command definition from cmd/root.go (lines 152-184)
  - Update any internal references or documentation
  - Ensure tests are updated to use modern syntax

  **Testing Criteria:**
  **Unit Tests:**
  - Verify command registry no longer includes add-skill
  - Test that add-skill command returns command not found error
  
  **Integration Tests:**
  - Validate 'add skill' command works correctly as replacement
  - Test error messaging when legacy command is attempted
  
  **Component Browser Tests:**
  - N/A (CLI only)

- [ ] Story-002: As a developer, I want to remove the add-agent legacy command so that the codebase only supports the modern 'add agent' syntax.

  **Acceptance Criteria:**
  - Remove add-agent command definition from cmd/root.go (lines 186-217)
  - Update any internal references or documentation
  - Ensure tests are updated to use modern syntax

  **Testing Criteria:**
  **Unit Tests:**
  - Verify command registry no longer includes add-agent
  - Test that add-agent command returns command not found error
  
  **Integration Tests:**
  - Validate 'add agent' command works correctly as replacement
  - Test error messaging when legacy command is attempted
  
  **Component Browser Tests:**
  - N/A (CLI only)

- [ ] Story-003: As a developer, I want to remove the add-command legacy command so that the codebase only supports the modern 'add command' syntax.

  **Acceptance Criteria:**
  - Remove add-command command definition from cmd/root.go (lines 219-250)
  - Update any internal references or documentation
  - Ensure tests are updated to use modern syntax

  **Testing Criteria:**
  **Unit Tests:**
  - Verify command registry no longer includes add-command
  - Test that add-command command returns command not found error
  
  **Integration Tests:**
  - Validate 'add command' command works correctly as replacement
  - Test error messaging when legacy command is attempted
  
  **Component Browser Tests:**
  - N/A (CLI only)

- [ ] Story-004: As a developer, I want to remove the add-all legacy command so that the codebase only supports the modern 'add all' syntax.

  **Acceptance Criteria:**
  - Remove add-all command definition from cmd/root.go (lines 252-283)
  - Update any internal references or documentation
  - Ensure tests are updated to use modern syntax

  **Testing Criteria:**
  **Unit Tests:**
  - Verify command registry no longer includes add-all
  - Test that add-all command returns command not found error
  
  **Integration Tests:**
  - Validate 'add all' command works correctly as replacement
  - Test error messaging when legacy command is attempted
  
  **Component Browser Tests:**
  - N/A (CLI only)

- [ ] Story-005: As a developer, I want to remove the link-legacy deprecated command to eliminate hidden legacy code paths.

  **Acceptance Criteria:**
  - Remove link-legacy command definition from cmd/root.go (lines 645-706)
  - Verify modern link subcommands cover all functionality
  - Ensure no internal code references link-legacy

  **Testing Criteria:**
  **Unit Tests:**
  - Verify link-legacy is completely removed from command registry
  - Test that modern link subcommands provide equivalent functionality
  
  **Integration Tests:**
  - Validate all link operations work with modern syntax
  - Test --target and --all-targets flags on modern link command
  
  **Component Browser Tests:**
  - N/A (CLI only)

- [ ] Story-006: As a developer, I want to remove the auto-link standalone command so that users must use 'link auto' instead.

  **Acceptance Criteria:**
  - Remove auto-link command definition from cmd/root.go (lines 589-596)
  - Verify 'link auto' provides equivalent functionality
  - Update any documentation or examples

  **Testing Criteria:**
  **Unit Tests:**
  - Verify auto-link command is removed from registry
  - Test that auto-link returns command not found error
  
  **Integration Tests:**
  - Validate 'link auto' command provides identical functionality
  - Test auto-detection logic remains functional
  
  **Component Browser Tests:**
  - N/A (CLI only)

- [ ] Story-007: As a developer, I want to remove the list-links standalone command so that users must use 'link list' instead.

  **Acceptance Criteria:**
  - Remove list-links command definition from cmd/root.go (lines 598-618)
  - Verify 'link list' provides equivalent functionality
  - Update any documentation or examples

  **Testing Criteria:**
  **Unit Tests:**
  - Verify list-links command is removed from registry
  - Test that list-links returns command not found error
  
  **Integration Tests:**
  - Validate 'link list' command provides identical output
  - Test link status display remains functional
  
  **Component Browser Tests:**
  - N/A (CLI only)

- [ ] Story-008: As a developer, I want to remove the link-status standalone command so that users must use 'link status' instead.

  **Acceptance Criteria:**
  - Remove link-status command definition from cmd/root.go (lines 620-642)
  - Verify 'link status' provides equivalent functionality
  - Update any documentation or examples

  **Testing Criteria:**
  **Unit Tests:**
  - Verify link-status command is removed from registry
  - Test that link-status returns command not found error
  
  **Integration Tests:**
  - Validate 'link status' command provides identical output
  - Test status display for all target types
  
  **Component Browser Tests:**
  - N/A (CLI only)

- [ ] Story-009: As a developer, I want to remove legacy metadata file support to simplify the codebase and standardize on lock files.

  **Acceptance Criteria:**
  - Remove internal/metadata/legacy.go file entirely
  - Remove any imports or references to legacy metadata handling
  - Update migration documentation to indicate legacy format is no longer supported

  **Testing Criteria:**
  **Unit Tests:**
  - Verify legacy.go file and package are completely removed
  - Test that only lock file format is read and written
  
  **Integration Tests:**
  - Validate that components with legacy metadata files are not recognized
  - Test lock file reading and writing functionality
  
  **Component Browser Tests:**
  - N/A (CLI only)

- [ ] Story-010: As a maintainer, I want to update all documentation to remove references to legacy commands.

  **Acceptance Criteria:**
  - Update README.md to remove legacy command examples
  - Update any inline code comments referencing old commands
  - Verify no documentation references deprecated commands
  - Add migration notes to CHANGELOG.md

  **Testing Criteria:**
  **Unit Tests:**
  - N/A (documentation only)
  
  **Integration Tests:**
  - Verify all documented commands execute successfully
  - Test that documentation examples produce expected output
  
  **Component Browser Tests:**
  - N/A (CLI only)

- [ ] Story-011: As a maintainer, I want comprehensive tests validating all legacy commands are removed.

  **Acceptance Criteria:**
  - Add tests confirming removed commands return appropriate errors
  - Verify modern command alternatives work correctly
  - Test that lock file format is the only supported metadata format
  - Ensure no regression in modern command functionality

  **Testing Criteria:**
  **Unit Tests:**
  - Test suite for legacy command removal validation
  - Unit tests for modern command equivalents
  
  **Integration Tests:**
  - End-to-end tests for all modern CLI workflows
  - Integration tests for lock file operations
  
  **Component Browser Tests:**
  - N/A (CLI only)

- [ ] Story-012: As a maintainer, I want to update the version number and CHANGELOG to reflect breaking changes.

  **Acceptance Criteria:**
  - Bump major version number to indicate breaking changes
  - Add detailed CHANGELOG entry documenting all removed commands
  - Include migration instructions in release notes
  - Update version constant in main.go

  **Testing Criteria:**
  **Unit Tests:**
  - Verify version number has been incremented
  - Test version display with --version flag
  
  **Integration Tests:**
  - N/A (version metadata)
  
  **Component Browser Tests:**
  - N/A (CLI only)

## Functional Requirements

- FR-1: The system must remove all legacy command definitions from cmd/root.go (add-skill, add-agent, add-command, add-all, link-legacy, auto-link, list-links, link-status)
- FR-2: The system must eliminate internal/metadata/legacy.go and all legacy metadata file support
- FR-3: The system must ensure modern command alternatives (add skill, add agent, link auto, link list, link status) provide equivalent functionality
- FR-4: The system must return appropriate "command not found" errors when legacy commands are attempted
- FR-5: The system must update all documentation to reflect only modern command syntax
- FR-6: The system must include comprehensive tests validating removal and modern command functionality
- FR-7: The system must bump the major version number and document breaking changes in CHANGELOG.md
- FR-8: The system must standardize on lock file format as the only supported metadata format

## Non-Goals (Out of Scope)

- Migration scripts or automatic translation of legacy commands
- Deprecation warnings or grace period for legacy commands
- Backward compatibility support for legacy command syntax
- Adding new CLI features or improvements beyond cleanup
- Command aliases or shell completion scripts
- Configuration file support (.agent-smith.yaml)
- Changes to the core functionality of modern commands

## Technical Notes

### Files to Modify

- **cmd/root.go**: Remove command definitions for all legacy commands (lines 152-283, 589-642, 645-706)
- **internal/metadata/legacy.go**: Delete entire file
- **main.go**: Update version constant for major version bump
- **README.md**: Update command examples to use modern syntax only
- **CHANGELOG.md**: Add comprehensive entry documenting breaking changes
- **Test files**: Update tests to remove legacy command usage and add validation tests

### Modern Command Equivalents

| Legacy Command | Modern Equivalent | Location in cmd/root.go |
|---------------|-------------------|------------------------|
| `add-skill` | `add skill` | Lines 36-149 (subcommands) |
| `add-agent` | `add agent` | Lines 36-149 (subcommands) |
| `add-command` | `add command` | Lines 36-149 (subcommands) |
| `add-all` | `add all` | Lines 36-149 (subcommands) |
| `auto-link` | `link auto` | Lines 333-587 (link subcommands) |
| `list-links` | `link list` | Lines 333-587 (link subcommands) |
| `link-status` | `link status` | Lines 333-587 (link subcommands) |
| `link-legacy` | `link <subcommand>` | Lines 333-587 (link subcommands) |

### Testing Strategy

1. **Removal Validation**: Test that all legacy commands return "command not found" errors
2. **Functional Equivalence**: Verify modern commands provide identical functionality
3. **Metadata Format**: Ensure only lock files are read/written
4. **Documentation Accuracy**: Validate all documented commands execute successfully
5. **Version Display**: Confirm version number reflects major version bump

## Success Criteria

- All legacy command definitions removed from cmd/root.go
- internal/metadata/legacy.go file deleted
- All tests pass with modern command syntax only
- Documentation updated to reflect modern commands exclusively
- Version bumped to next major version
- CHANGELOG.md includes comprehensive breaking change documentation
- No references to legacy commands remain in codebase
- Modern commands provide equivalent functionality to removed legacy commands
