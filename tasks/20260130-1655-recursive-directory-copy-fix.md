# PRD: Recursive Directory Copy for Component Installation

**Created**: 2026-01-30 16:55 UTC

---

## Introduction

The current component installation process fails to copy subdirectories when installing skills, agents, or commands from repositories. This causes skills like `skill-creator` from the anthropics/skills repository to be incomplete, missing essential subdirectories like `references/` and `scripts/` that are required for the skill to function properly.

The root cause is in the `CopyComponentFiles` function in `internal/fileutil/fileutil.go:91-128`, which explicitly skips all subdirectories and only copies files at the root level of the component directory. This was an intentional design choice to keep installations lightweight, but it breaks components that genuinely need subdirectory structures.

## Goals

- Enable complete component installation including all subdirectories and nested files
- Maintain reasonable installation sizes by filtering out unnecessary directories (`.git`, `node_modules`, etc.)
- Ensure backward compatibility with existing single-file and simple directory-based components
- Provide intuitive behavior that matches user expectations (all necessary files are copied)

## User Stories

- [ ] Story-001: As a user, I want all subdirectories and files to be copied when installing a component so that the component has all necessary assets to function properly.

  **Acceptance Criteria:**
  - Component installation recursively copies all subdirectories and their contents
  - The `skill-creator` skill installs with `references/` and `scripts/` subdirectories intact
  - Nested directory structures are preserved during installation
  - Single-file components continue to work as before (no regression)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Test recursive directory copy function with nested structures
  - Test exclusion filter logic for ignored patterns
  - Test single-file component copy (backward compatibility)
  
  **Integration Tests:**
  - Install skill-creator and verify references/ and scripts/ directories exist
  - Install a simple single-file component and verify it still works
  - Install a component with deeply nested directories and verify structure
  
  **Component Browser Tests:**
  - N/A (CLI tool, no browser components)

- [ ] Story-002: As a developer, I want common unwanted directories to be automatically excluded during installation so that installations remain clean and lightweight.

  **Acceptance Criteria:**
  - `.git/` directory is excluded from component installation
  - `node_modules/` directory is excluded from component installation
  - `__pycache__/` and `*.pyc` files are excluded from component installation
  - `.DS_Store` files are excluded from component installation
  - `.pytest_cache/` directory is excluded from component installation
  - Exclusion patterns can be easily extended in the future
  
  **Testing Criteria:**
  **Unit Tests:**
  - Test exclusion filter with each excluded pattern
  - Test that non-excluded files are still copied
  - Test exclusion filter with mixed directory structures
  
  **Integration Tests:**
  - Clone a repo with .git directory and verify it's not copied
  - Create test component with node_modules and verify exclusion
  - Create test component with Python cache files and verify exclusion
  
  **Component Browser Tests:**
  - N/A (CLI tool, no browser components)

- [ ] Story-003: As a user, I want the installation process to remain fast and efficient even with recursive copying so that bulk installations complete in reasonable time.

  **Acceptance Criteria:**
  - Recursive copy implementation uses efficient file system operations
  - Installation time for typical components increases by less than 20%
  - Bulk installation (`install all`) completes without significant performance degradation
  - Memory usage remains constant regardless of directory depth
  
  **Testing Criteria:**
  **Unit Tests:**
  - Benchmark recursive copy function with various directory sizes
  - Test memory usage with deeply nested directories
  - Verify no memory leaks during recursive operations
  
  **Integration Tests:**
  - Time bulk installation before and after changes
  - Install component with 100+ files and verify performance
  - Test installation with directory depth of 10+ levels
  
  **Component Browser Tests:**
  - N/A (CLI tool, no browser components)

## Functional Requirements

- FR-1: The system SHALL recursively copy all subdirectories and files within a component directory during installation
- FR-2: The system SHALL exclude the following patterns from component installation: `.git/`, `node_modules/`, `__pycache__/`, `*.pyc`, `.DS_Store`, `.pytest_cache/`
- FR-3: The system SHALL preserve the directory structure of the source component in the destination
- FR-4: The system SHALL maintain backward compatibility with single-file component installations
- FR-5: The system SHALL use the same recursive copy logic for skills, agents, and commands
- FR-6: The system SHALL handle symbolic links appropriately (either skip or copy the target)
- FR-7: The system SHALL provide clear error messages if directory copying fails
- FR-8: The system SHALL maintain cross-platform compatibility (Windows, macOS, Linux)

## Technical Implementation Details

### Files to Modify

**Primary File**: `internal/fileutil/fileutil.go`

**Changes Required**:

1. Create a new helper function `CopyDirectoryRecursive` that:
   - Accepts source path, destination path, and exclusion patterns
   - Walks the source directory recursively
   - Creates destination directories as needed
   - Copies files while respecting exclusion patterns
   - Handles errors gracefully

2. Create exclusion filter function `shouldExclude` that:
   - Accepts a file path and list of exclusion patterns
   - Returns boolean indicating if path should be excluded
   - Supports both exact matches and glob patterns

3. Update `CopyComponentFiles` function to:
   - Remove the `if entry.IsDir() { continue }` logic
   - Call `CopyDirectoryRecursive` for directory-based components
   - Maintain single-file component logic unchanged

### Exclusion Patterns

Define as a package-level variable for easy maintenance:

```go
var defaultExclusionPatterns = []string{
    ".git",
    ".gitignore", 
    "node_modules",
    "__pycache__",
    ".pytest_cache",
    ".DS_Store",
    ".vscode",    // Optional: IDE config
    ".idea",      // Optional: IDE config
}
```

### Error Handling

- Fail fast on permission errors
- Continue on individual file copy failures with warning
- Collect and report all errors at the end
- Maintain partial installation state for debugging

## Testing Strategy

### Unit Tests

Create `internal/fileutil/recursive_copy_test.go`:
- Test recursive copy with simple directory structure
- Test recursive copy with deeply nested directories
- Test exclusion patterns work correctly
- Test error handling (permission denied, disk full)
- Test symbolic link handling
- Test empty directories are preserved
- Test backward compatibility with single-file components

### Integration Tests

Update `component_download_integration_test.go`:
- Test installing anthropics/skills skill-creator
- Verify references/ directory exists and contains files
- Verify scripts/ directory exists and contains files
- Test bulk installation with multiple complex components
- Benchmark installation performance before/after

### Manual Testing Checklist

- [ ] Install skill-creator and verify subdirectories are present
- [ ] Check that .git directory is not copied
- [ ] Verify node_modules are not copied if present in source
- [ ] Test on macOS, Linux, and Windows (if available)
- [ ] Verify existing simple components still install correctly
- [ ] Test error messages when permissions are insufficient

## Non-Goals (Out of Scope)

- No configuration file or flags to control recursive vs non-recursive behavior
- No whitelist/blacklist configuration for specific subdirectories
- No compression or archival of installed components
- No modification of the component lockfile format
- No changes to the download progress bar or UI
- No optimization for large binary files (handled by Git LFS if needed)
- No automatic deduplication of common files across components
- No support for partial component updates (still full replacement)

## Dependencies

**External Dependencies**:
- Go standard library `os`, `path/filepath`, `io`
- Existing `go-git` library (no changes needed)

**Internal Dependencies**:
- `internal/fileutil/fileutil.go` (main changes)
- `internal/downloader/skill.go` (calls CopyComponentFiles)
- `internal/downloader/agent.go` (calls CopyComponentFiles)
- `internal/downloader/command.go` (calls CopyComponentFiles)

No changes needed to downloader files themselves, as they already call `CopyComponentFiles` which will be updated internally.

## Success Metrics

- All subdirectories and files are successfully copied during component installation
- Installation of anthropics/skills skill-creator includes references/ and scripts/ directories
- No regression in single-file or simple directory component installations
- Installation performance degrades by less than 20% for typical components
- Zero critical bugs reported related to file copying after release
- Test coverage for fileutil package increases to >80%

## Rollout Plan

**Phase 1: Implementation & Unit Tests**
- Implement `CopyDirectoryRecursive` helper function
- Implement `shouldExclude` filter function
- Update `CopyComponentFiles` to use recursive logic
- Write comprehensive unit tests

**Phase 2: Integration Testing**
- Add integration tests for skill-creator installation
- Performance benchmarking before/after
- Test on multiple platforms (macOS, Linux, Windows if available)

**Phase 3: Release**
- Merge changes to main branch
- Update CHANGELOG with fix description
- Document breaking changes (if any) in README
- Announce fix to users who reported the issue

## Open Questions

1. Should we make exclusion patterns configurable via a config file, or keep them hardcoded?
   - **Answer**: Keep hardcoded initially, can be made configurable in future if users request it

2. How should we handle symbolic links? Skip them, copy them as symlinks, or copy the target?
   - **Answer**: Skip symbolic links initially to avoid potential issues, can be enhanced later

3. Should we add a `--no-recursive` flag for users who want the old behavior?
   - **Answer**: No, recursive should be the default and only behavior for simplicity

4. Should we preserve empty directories, or only create directories that contain files?
   - **Answer**: Preserve empty directories to maintain exact structure from source

---

## Appendix: Current Code Analysis

### Current Implementation (Problematic)

Location: `internal/fileutil/fileutil.go:107-127`

```go
// Directory-based component - copy only files in component directory (non-recursive)
entries, err := os.ReadDir(componentDir)
if err != nil {
    return err
}

for _, entry := range entries {
    // Skip subdirectories - only copy files at the root level
    if entry.IsDir() {
        continue  // <-- THIS IS THE PROBLEM
    }
    
    srcPath := filepath.Join(componentDir, entry.Name())
    dstPath := filepath.Join(dst, entry.Name())
    
    if err := CopyFile(srcPath, dstPath); err != nil {
        return err
    }
}
```

### Affected Components

Real-world example from anthropics/skills repository:

**skill-creator directory structure**:
```
skill-creator/
├── SKILL.md            ✅ Currently copied
├── LICENSE.txt         ✅ Currently copied
├── references/         ❌ Currently skipped
│   └── (files)
└── scripts/            ❌ Currently skipped
    └── (files)
```

After this fix, the complete directory structure will be preserved.
