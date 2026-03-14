# Batch Repository Updates Optimization

**Status:** In Progress  
**Created:** 2026-02-03  
**Type:** Performance Optimization  
**Priority:** High

## Overview

Optimize repository cloning operations in `update` and `materialization sync` commands by batching components from the same source repository. Apply the same optimization pattern used in `install-all` to reduce redundant git clone operations.

## Problem Statement

### Current Performance Issues

**UpdateAll() Process:**
- For each component, performs two git clones:
  1. Shallow clone to check current HEAD SHA
  2. Full clone to re-download component
- **Impact:** For N components from same repo = 2N clone operations

**Materialization Sync Checks:**
- For each materialized component, performs one git clone to check current SHA
- **Impact:** For N components from same repo = N clone operations

### Example Scenario

**Current behavior with 5 skills from same repository:**
- 5 clones to check SHA
- 5 clones to download
- **Total: 10 git clone operations**

**Optimized behavior:**
- 1 clone (reused for all operations)
- **Total: 1 git clone operation**
- **Result: 90% reduction in git operations**

## Success Criteria

### Functional Requirements
- ✅ All existing functionality works identically
- ✅ Same progress display format (transparent to users)
- ✅ Same error handling and messages
- ✅ Same summary statistics
- ✅ Profile support preserved
- ✅ Auth support preserved

### Performance Requirements
- ✅ Reduce git clone operations for same-repo components by 80-90%
- ✅ No performance regression for single-component updates
- ✅ No additional memory overhead (temp dirs cleaned immediately)

### Quality Requirements
- ✅ No breaking changes to public APIs
- ✅ Code follows existing patterns from `bulk.go`
- ✅ Proper cleanup of temporary directories

## Technical Design

### Architecture Pattern

The optimization uses the same batching pattern already implemented in `bulk.go`:

**Current Flow (Per-Component):**
```
For each component:
  ├─ Clone repo to temp dir
  ├─ Check SHA / detect components
  └─ Download component
```

**Optimized Flow (Batched by Repository):**
```
Group components by source repository
For each unique repository:
  ├─ Clone repo once to temp dir
  ├─ Get HEAD SHA once
  └─ For each component from this repo:
      ├─ Check if update needed
      └─ Download using *WithRepo() method (reuses clone)
```

### Component 1: UpdateAll() Optimization

**File:** `internal/updater/updater.go`

**Changes:**
1. Add helper to group components by repository
2. Refactor `UpdateAll()` to process by repository batch
3. Reuse existing `Download*WithRepo()` methods
4. Preserve all existing display/error handling

**Key Implementation Details:**
```go
// Group components by source repository
type ComponentUpdateInfo struct {
    Type     string
    Name     string
    Metadata *models.ComponentMetadata
}

// Map: repoURL -> []ComponentUpdateInfo
componentsByRepo := make(map[string][]ComponentUpdateInfo)

// For each unique repository:
for repoURL, components := range componentsByRepo {
    // Clone once
    tempDir, _ := os.MkdirTemp("", "agent-smith-update-batch-*")
    defer os.RemoveAll(tempDir)
    
    // Get current SHA once
    currentSHA := cloneAndGetSHA(repoURL, tempDir)
    
    // Check and update all components from this repo
    for _, comp := range components {
        // Same display format as before
        fmt.Print(styles.ComponentProgressFormat(...))
        
        if comp.Metadata.Commit != currentSHA {
            // Update using *WithRepo() method
            downloader.Download*WithRepo(fullURL, name, repoURL, tempDir, ...)
        }
    }
}
```

**Methods to Add/Modify:**
- Refactor `UpdateAll()` - batch processing
- Add `groupComponentsByRepository()` - helper
- Add `updateComponentsFromRepo()` - batch update helper
- Keep `UpdateComponent()` unchanged (single component doesn't benefit)

### Component 2: Materialization Sync Optimization

**File:** `pkg/project/materialization.go`

**Changes:**
1. Add batched sync check helper
2. Group materialized components by source repository
3. Clone each unique repository once
4. Apply sync status to all components from that repo

**Key Implementation Details:**
```go
// New helper function for batch sync checks
func CheckMultipleComponentsSyncStatusBatched(components []ComponentInfo) (map[string]SyncCheckResult, error)

type SyncCheckResult struct {
    Status SyncStatus
    Error  error
}

// Implementation:
// 1. Group components by metadata.Source
// 2. For each unique repo, clone once and get current SHA
// 3. Compare each component's stored SHA with current
// 4. Return map of component name -> sync result
```

**API Design:**
- Keep existing `CheckComponentSyncStatus()` for backward compatibility
- Add new `CheckMultipleComponentsSyncStatusBatched()` for batch operations
- Commands can opt into batched version for better performance

### Component 3: Command Integration

**Files to Update:**
- Status commands that check materialization sync
- Update commands to use batched checking where applicable

**Integration Pattern:**
```go
// Before (individual checks):
for _, comp := range components {
    status, err := CheckComponentSyncStatus(comp.Type, comp.Name, comp.Metadata)
    // Display status
}

// After (batched checks):
results, err := CheckMultipleComponentsSyncStatusBatched(components)
for name, result := range results {
    // Display same status format
}
```

## Implementation Plan

### Phase 1: UpdateAll() Batching
1. Extract component scanning and grouping logic
2. Implement batch processing by repository
3. Reuse existing `*WithRepo()` downloader methods
4. Preserve display format and error handling
5. Test with same-repo and mixed-repo scenarios

### Phase 2: Materialization Batching
1. Add batched sync check helper function
2. Implement grouping by source repository
3. Clone once per unique repository
4. Return results map for all components
5. Update commands to use batched version

### Phase 3: Testing & Validation
1. Test with multiple components from same repo
2. Test with mixed repositories
3. Test error scenarios (auth failures, missing repos)
4. Verify profile support works correctly
5. Verify progress display unchanged
6. Manual testing with real repositories

### Phase 4: Cleanup
1. Remove any temporary debug code
2. Ensure proper cleanup of temp directories
3. Add comments explaining batching logic
4. Update task documentation

## Expected Performance Improvements

### Benchmark Scenarios

| Scenario | Before | After | Improvement |
|----------|--------|-------|-------------|
| 5 skills from 1 repo | 10 clones | 1 clone | 90% faster |
| 10 components, 3 repos | 20 clones | 3 clones | 85% faster |
| 10 components, 10 repos | 20 clones | 10 clones | 50% faster |
| 15 components, 5 repos | 30 clones | 5 clones | 83% faster |

### Real-World Impact
- Most users have multiple skills from same library repo → **80-90% reduction**
- Even with diverse repos → **50%+ reduction**
- Materialization checks benefit most (typically all from same source) → **90-95% reduction**

## Testing Strategy

### Manual Testing Scenarios

**Test Case 1: Same Repository**
- Install 5 skills from same repo
- Run `update-all` command
- Verify only 1 clone operation occurs
- Verify all components update correctly

**Test Case 2: Mixed Repositories**
- Install components from 3 different repos
- Run `update-all` command
- Verify 3 clone operations (one per repo)
- Verify correct updates for each

**Test Case 3: Materialization Sync**
- Materialize 5 components from same repo
- Run sync check command
- Verify only 1 clone occurs
- Verify correct sync status for all

**Test Case 4: Error Handling**
- Components from repo that requires auth
- Components from non-existent repo
- Verify errors handled gracefully per batch

**Test Case 5: Profile Support**
- Update components in active profile
- Verify batching works with profile-aware downloaders

### Expected Results
- Same output format as before
- Faster execution time
- Same error messages
- Same statistics in summary

## Files Modified

| File | Purpose | Complexity |
|------|---------|------------|
| `internal/updater/updater.go` | Batch UpdateAll() by repository | Medium |
| `pkg/project/materialization.go` | Add batched sync check function | Medium |
| Status/sync commands | Use batched checking | Low |

## Dependencies

### Existing Infrastructure
- `bulk.go` - Reference implementation for batching pattern
- `Download*WithRepo()` methods - Already exist in all downloaders
- `GetCurrentRepoSHA()` - Already supports auth and profiles

### No Breaking Changes
- All existing APIs remain unchanged
- New batched functions are additions
- Existing single-component functions preserved

## Success Metrics

### Performance Metrics
- Number of unique repositories cloned vs total components
- Time saved per update operation
- Git operation reduction percentage

### Quality Metrics
- Zero breaking changes to existing APIs
- Zero regression in functionality
- Zero new error scenarios introduced

## Rollback Plan

Since this is a performance optimization with no API changes:
- If issues arise, can easily revert to original per-component logic
- All existing code paths remain intact
- No database or file format changes
- Low risk of breaking existing workflows

## Future Enhancements

**Potential Follow-ups:**
1. Add verbose flag to show which repos are being batched
2. Add performance metrics to summary output
3. Cache cloned repos between commands (more complex)
4. Parallel processing of different repositories

## References

- Original `install-all` optimization: `tasks/20260126-1207-single-clone-optimization.md`
- Bulk downloader implementation: `internal/downloader/bulk.go`
- Existing `*WithRepo()` methods in skill/agent/command downloaders

## Notes

- This optimization follows the exact same pattern as the successful `install-all` optimization
- No changes to user-facing behavior or output
- Pure performance improvement with no functional changes
- Already have infrastructure (`*WithRepo()` methods) to support batching
