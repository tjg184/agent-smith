# PRD: Enhanced Component Detection for Agent Smith

## Introduction

The current component detection system in Agent Smith only recognizes specific filename patterns (SKILL.md, AGENT.md, COMMAND.md), which fails to detect components in repositories using different organizational patterns like the wshobson/agents repository. This enhancement will expand detection capabilities to support diverse repository structures while maintaining backward compatibility.

## Goals

- Enable detection of skills, agents, and commands across different repository organizational patterns
- Support the wshobson/agents repository structure (plugins/category/type/component-type files)
- Maintain backward compatibility with existing SKILL.md detection
- Improve reliability of the `add-all` command for bulk component downloads

## User Stories

- [x] Story-001: As a user, I want the add-all command to detect all components in a repository so that I can bulk download skills, agents, and commands regardless of how they're organized.

  **Acceptance Criteria:**
  - System detects SKILL.md files in skills directories (existing behavior preserved)
  - System detects .md files in /agents/ directories as agent components
  - System detects .md files in /commands/ directories as command components
  - Bulk download processes all detected component types successfully
  - Existing functionality remains unchanged for repositories with current structure

- [ ] Story-002: As a user, I want clear feedback during bulk download so that I know what components were found and their status.

  **Acceptance Criteria:**
  - Console output shows number of each component type detected
  - Progress indicators show which components are being downloaded
  - Success/failure status reported for each component
  - Summary provided with total counts and any errors encountered

- [ ] Story-003: As a developer, I want maintainable detection logic so that future repository patterns can be added easily.

  **Acceptance Criteria:**
  - Detection logic organized by component type with clear separation
  - Path-based patterns used instead of just filename matching
  - Commented code explains detection patterns for future modifications
  - No hard-coded assumptions about directory depth or naming conventions

## Functional Requirements

- FR-1: The system must detect skills using existing SKILL.md filename pattern
- FR-2: The system must detect agents by finding .md files in paths containing "/agents/"
- FR-3: The system must detect commands by finding .md files in paths containing "/commands/"
- FR-4: The system must extract component names appropriately (directory for skills, filename for agents/commands)
- FR-5: The system must preserve all existing functionality for repositories using current patterns
- FR-6: The system must provide clear console feedback during bulk download operations

## Non-Goals

- No changes to individual component download commands (add-skill, add-agent, add-command)
- No configuration file or user-configurable detection patterns in this version
- No support for custom component types beyond skills, agents, and commands
- No changes to metadata or lock file formats
- No modifications to linking or execution functionality

## Technical Approach

### Enhanced Detection Logic

Replace the current filename-only detection in `detectComponentsInRepo()` with path-based pattern matching:

1. **Skills**: Keep existing logic - detect files named exactly "SKILL.md"
2. **Agents**: Detect any `.md` file in a path containing "/agents/"
3. **Commands**: Detect any `.md` file in a path containing "/commands/"

### Implementation Details

- Modify `detectComponentsInRepo()` function in main.go (lines 352-393)
- Add path-based detection using `strings.Contains()` for pattern matching
- Extract component names based on component type:
  - Skills: Use directory name (existing behavior)
  - Agents/Commands: Use filename without .md extension
- Maintain existing `DetectedComponent` structure and return format
- Preserve all error handling and edge case logic

### Backward Compatibility

- Existing SKILL.md detection logic unchanged
- All current repository structures continue to work
- No breaking changes to APIs or function signatures
- Existing test cases should continue to pass

## Success Criteria

- Add-all command successfully detects and downloads all components from wshobson/agents repository
- Existing repositories with SKILL.md files continue to work unchanged
- Console output provides clear feedback on detection and download progress
- Code review confirms maintainable and well-documented detection logic
- Unit tests cover new detection patterns and edge cases

## Testing Strategy

### Unit Tests
- Test path-based detection for agents in various directory structures
- Test path-based detection for commands in various directory structures
- Verify SKILL.md detection remains unchanged
- Test component name extraction for all three types
- Edge cases: empty directories, malformed paths, mixed patterns

### Integration Tests
- Test add-all command with wshobson/agents repository structure
- Test with existing repositories using current patterns
- Verify bulk download processes all detected components correctly
- Confirm console output format and content

### Regression Tests
- Ensure all existing add-skill, add-agent, add-command functionality unchanged
- Verify metadata and lock file generation unchanged
- Test linking and execution functionality unaffected