# PRD: Optimize Add-All to Use Single Clone for All Operations

## Introduction

The install-all command currently performs repository cloning multiple times - once for component detection and once for each component type (skills, agents, commands) for metadata extraction. This creates 3-4 redundant git operations instead of the intended single clone optimization, causing slow performance and excessive network traffic.

## Goals

- Achieve true single-clone optimization where repository is cloned only once
- Eliminate redundant metadata cloning operations in individual component download methods
- Maintain backward compatibility for individual download commands
- Improve performance significantly especially for large repositories

## User Stories

- [x] Story-001: As a user running install-all, I want the repository cloned only once so that the command completes faster and uses less bandwidth.

  **Acceptance Criteria:**
  - Repository is cloned exactly one time during install-all execution
  - Component detection uses the initial cloned repository
  - Metadata extraction reuses the same repository object (no additional clones)
  - Individual install-skill/install-agent/add-install-command commands maintain current behavior
  - Performance improvement is measurable (75% reduction in git operations)
  - Large repositories process significantly faster than current implementation

- [ ] Story-002: As a developer maintaining this codebase, I want the optimization to be maintainable and safe so that I can understand and modify the logic easily.

  **Acceptance Criteria:**
  - Clear separation between optimized and non-optimized code paths
  - Backward compatibility is preserved through optional parameters
  - Error handling provides appropriate fallbacks
  - Code changes are minimal and focused on the clone optimization
  - Tests validate both optimized and legacy behavior

## Functional Requirements

- FR-1: The system must modify BulkDownloader.AddAll to extract git.Repository object from the initial clone and pass it to component download methods.
- FR-2: The system must update downloadSkillWithRepo, downloadAgentWithRepo, and downloadCommandWithRepo methods to accept an optional shared repository parameter.
- FR-3: The system must eliminate redundant git.PlainClone operations for metadata extraction when a shared repository is provided.
- FR-4: The system must maintain backward compatibility for individual install-skill, install-agent, and install-command commands by falling back to current cloning behavior when no shared repository is provided.
- FR-5: The system must preserve all existing functionality including metadata extraction, lock file generation, and file copying operations.

## Non-Goals

- No changes to component detection logic or file structure analysis
- No modifications to error handling or retry mechanisms
- No changes to user interface or command-line argument parsing
- No consideration for parallel processing (this optimization focuses on single-threaded sequential improvement)

## Implementation Strategy

### Phase 1: Modify Method Signatures
- Add optional `*git.Repository` parameter to all three `*WithRepo` methods
- Maintain existing parameters for backward compatibility
- Use nil checks to determine when to use shared repository vs individual cloning

### Phase 2: Update BulkDownloader Logic
- Extract git.Repository object from the initial clone in AddAll method
- Pass repository object to all component download methods
- Maintain existing error handling and fallback mechanisms

### Phase 3: Preserve Backward Compatibility
- Ensure individual install-skill/install-agent/add-install-command commands work exactly as before
- Only apply optimization when called from BulkDownloader.AddAll context
- Keep all existing error messages and metadata generation intact

### Phase 4: Validation and Testing
- Verify single clone behavior through testing
- Ensure metadata extraction works correctly with shared repository
- Test with various repository types (GitHub, GitLab, local)
- Performance benchmarking to measure improvement

## Expected Performance Improvement

**Current State:**
- 1 clone for component detection
- 3 additional clones for metadata extraction (skills, agents, commands)
- Total: 4 git clone operations per install-all execution

**Optimized State:**
- 1 clone for component detection AND metadata extraction
- 0 additional clones for individual component types
- Total: 1 git clone operation per install-all execution

**Improvement Metrics:**
- 75% reduction in git clone operations
- 75% reduction in network traffic
- Significant time savings for large repositories
- Reduced disk I/O from multiple clone operations