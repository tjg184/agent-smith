# PRD: Add-All Repository Cloning Optimization

## Introduction

The `add-all` command in Agent Smith currently clones the same repository multiple times when downloading skills, agents, and commands from a single repository. This inefficiency occurs because the `AddAll` method clones the repository once for component detection, then calls individual download methods (`downloadSkill`, `downloadAgent`, `downloadCommand`) which each clone the repository again for their own component detection. For a repository containing one skill, one agent, and one command, this results in 4 total clones of the same repository.

This optimization will eliminate redundant repository cloning while maintaining full backward compatibility and limiting scope to only the `add-all` command.

## Goals

- Eliminate redundant repository cloning in `add-all` command operations
- Reduce network bandwidth usage by avoiding duplicate downloads
- Improve `add-all` command performance by minimizing git operations
- Maintain full backward compatibility for existing APIs and commands
- Limit scope to `add-all` command optimization only

## User Stories

- [x] Story-001: As a user running `add-all`, I want the repository to be cloned only once so that the command completes faster and uses less bandwidth.

  **Acceptance Criteria:**
  - Repository is cloned exactly once during `add-all` execution
  - All component types (skills, agents, commands) are still detected and downloaded
  - No breaking changes to existing CLI interface
  - Performance improvement measurable through timing

- [ ] Story-002: As a developer, I want the optimization to maintain backward compatibility so that existing code continues to work without changes.

  **Acceptance Criteria:**
  - Individual download methods (`downloadSkill`, `downloadAgent`, `downloadCommand`) maintain existing signatures
  - No changes to existing API contracts or interfaces
  - Existing error handling and logging behaviors preserved
  - All existing unit tests continue to pass

- [ ] Story-003: As a user, I want consistent behavior between `add-all` and individual download commands so that the user experience remains predictable.

  **Acceptance Criteria:**
  - Metadata creation and lock file generation unchanged
  - Component detection results identical between optimized and original versions
  - Error messages and warnings remain consistent
  - File structure and permissions unchanged

## Functional Requirements

- FR-1: The system SHALL modify `BulkDownloader.AddAll()` to pass pre-cloned repository paths to individual download methods
- FR-2: The system SHALL add optional `repoPath` parameter to `downloadSkill`, `downloadAgent`, and `downloadCommand` methods
- FR-3: The system SHALL skip repository cloning when `repoPath` parameter is provided to individual download methods
- FR-4: The system SHALL maintain all existing method signatures for backward compatibility
- FR-5: The system SHALL preserve all existing functionality including metadata generation and lock file creation
- FR-6: The system SHALL ensure proper cleanup of temporary directories at the correct scope level
- FR-7: The system SHALL maintain existing error handling and logging patterns

## Non-Goals (Out of Scope)

- No optimization of individual download commands (`add-skill`, `add-agent`, `add-command`)
- No implementation of repository caching system beyond temporary operation scope
- No performance monitoring or metrics collection
- No changes to component detection logic or algorithms
- No modifications to file structure or metadata formats
- No changes to dependency injection or architecture patterns

## Technical Implementation Notes

Based on code analysis, the recommended approach is:

1. **Add Optional repoPath Parameter**: Modify individual download methods to accept optional `repoPath` parameter
2. **Conditional Cloning Logic**: Skip cloning step when `repoPath` is provided
3. **AddAll Integration**: Modify `AddAll` to pass the already-cloned temporary directory path
4. **Cleanup Management**: Ensure temporary directory cleanup happens at the correct level

This approach maintains the existing architecture while eliminating the 4x cloning issue (1 for AddAll detection + 3 for individual method detection).

## Success Criteria

- Repository cloning count reduced from N+1 to 1 (where N = number of component types found)
- No regression in existing functionality or user experience
- All existing tests pass without modification
- Performance improvement observable in real-world scenarios
- Zero breaking changes to public APIs or CLI interface