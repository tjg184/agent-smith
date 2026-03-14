# PRD: CLI Formatting Consistency Fixes

## Overview

Fix formatting inconsistencies in CLI output by standardizing on `formatter.NewBoxTable()` and proper Unicode width calculations. The primary issue is in `agent-smith profile list` which uses manual box-drawing with byte-length calculations instead of visual display width, causing misalignment with multi-byte UTF-8 characters like checkmarks.

## Problem Statement

### Current Issues

1. **Profile List Misalignment** (High Priority - User Facing)
   - The `agent-smith profile list` command has visibly misaligned columns
   - Uses `len()` which counts bytes, not visual character width
   - Checkmark indicator `✓` and multi-byte characters break padding calculations
   - Results in inconsistent spacing between profile names and component counts

2. **Inconsistent Infrastructure** (Medium Priority - Technical Debt)
   - `visibleLength()` function exists but is private (can't be used outside formatter package)
   - Some code paths use manual box-drawing, others use `BoxTable`
   - `boxes.go` helper functions have same width calculation bugs (not currently used in production)

### User Impact

Users see poorly formatted output:
```
┌────────────────────────────────────────────────────────────────────────────┐
│                            Available Profiles                              │
├────────────────────────────────────────────────────────────────────────────┤
│   paytient                                                          (empty) │
│ ✓ tjg184-agent-smith                                              (1 skill) │
│   tjg184-skills                                                   (1 skill) │
│   wshobson-agents                      (55 agents, 140 skills, 41 commands) │
└────────────────────────────────────────────────────────────────────────────┘
```

Notice the inconsistent spacing - profile names don't align properly, and the component counts have varying amounts of space before them.

## Goals

### Primary Goals

1. Fix `agent-smith profile list` alignment to match the quality of `agent-smith target list` and `agent-smith link status`
2. Export `VisibleLength()` function for use throughout the codebase
3. Establish consistent formatting patterns across all CLI commands

### Secondary Goals

1. Fix `boxes.go` helper functions to handle UTF-8 properly for future use
2. Reduce code complexity by replacing manual formatting with `BoxTable`
3. Improve maintainability with single source of truth for width calculations

## Non-Goals

- Changing the visual style or layout significantly (maintain current box-drawing aesthetic)
- Modifying commands that already work correctly (`target list`, `link status`)
- Adding new formatting features or capabilities
- Changing terminal color schemes or output verbosity

## Success Metrics

### Functional Requirements

- ✅ Profile list columns align perfectly regardless of profile name length
- ✅ Checkmark indicator (`✓`) doesn't break column alignment
- ✅ Long profile names are handled gracefully (truncation if needed)
- ✅ Visual consistency with `target list` and `link status` commands
- ✅ All existing tests pass
- ✅ No regression in other CLI commands

### Technical Requirements

- ✅ `VisibleLength()` function is exported and accessible
- ✅ `boxes.go` functions use proper width calculations
- ✅ Manual box-drawing code replaced with `BoxTable` implementation
- ✅ Code is cleaner (net reduction in lines of code)

## Technical Design

### Architecture Overview

**Existing Infrastructure:**
- ✅ `github.com/mattn/go-runewidth v0.0.19` already in dependencies
- ✅ `visibleLength()` function exists in `internal/formatter/box_table.go`
- ✅ `formatter.NewBoxTable()` used successfully in `target list` and `link status`

**Current Problem:**
```go
// main.go:1392-1393 - WRONG
nameLen := len(profile.Name)  // Counts bytes, not display width
countLen := len(componentStr)  // Counts bytes, not display width
padding := availableSpace - nameLen - countLen  // Incorrect padding
```

**Solution Pattern:**
```go
// Use BoxTable which internally uses visibleLength()
table := formatter.NewBoxTable(os.Stdout, []string{"Profile", "Components"})
for _, profile := range profiles {
    profileCell := fmt.Sprintf("%s %s", activeIndicator, profile.Name)
    table.AddRow([]string{profileCell, componentStr})
}
table.Render()
```

### Implementation Phases

#### Phase 1: Export VisibleLength() Function

**File:** `internal/formatter/box_table.go`

**Changes:**
1. Line 32: `func visibleLength(s string)` → `func VisibleLength(s string)`
2. Update internal calls (lines 78, 93, 111, 126) to use `VisibleLength`
3. Update tests in `box_table_test.go` to use exported function name

**Why:** Makes the function available for use in main.go and establishes single source of truth

**Risk:** Low - only visibility change, no behavioral changes

---

#### Phase 2: Fix Profile List Command

**File:** `main.go` (lines 1337-1425)

**Current Implementation:**
- Manual box-drawing with hardcoded borders
- Manual padding calculations using `len()`
- ~90 lines of complex formatting code

**New Implementation:**
```go
// Build headers for table
headers := []string{"Profile", "Components"}
table := formatter.NewBoxTable(os.Stdout, headers)

// Add rows
for _, profile := range filteredProfiles {
    // Count components
    agents, skills, commands := pm.CountComponents(profile)
    
    // Build component counts string
    var components []string
    if agents > 0 {
        components = append(components, fmt.Sprintf("%d agent%s", agents, plural(agents)))
    }
    if skills > 0 {
        components = append(components, fmt.Sprintf("%d skill%s", skills, plural(skills)))
    }
    if commands > 0 {
        components = append(components, fmt.Sprintf("%d command%s", commands, plural(commands)))
    }
    
    componentStr := ""
    if len(components) > 0 {
        componentStr = fmt.Sprintf("(%s)", strings.Join(components, ", "))
    } else {
        componentStr = "(empty)"
    }
    
    // Build profile cell with indicator
    activeIndicator := " "
    if profile.Name == activeProfile {
        activeIndicator = formatter.ColoredSuccess()
    }
    profileCell := fmt.Sprintf("%s %s", activeIndicator, profile.Name)
    
    table.AddRow([]string{profileCell, componentStr})
}

// Render the table
table.Render()
```

**Expected Output:**
```
┌────────────────────────┬────────────────────────────────────┐
│ Profile                │ Components                         │
├────────────────────────┼────────────────────────────────────┤
│   paytient             │ (empty)                            │
│ ✓ tjg184-agent-smith   │ (1 skill)                          │
│   tjg184-skills        │ (1 skill)                          │
│   wshobson-agents      │ (55 agents, 140 skills, 41 commands)│
└────────────────────────┴────────────────────────────────────┘
```

**Code Impact:**
- Remove: ~80 lines of manual formatting
- Add: ~30 lines using BoxTable
- Net: -50 lines (simpler code!)

**Risk:** Low - using proven BoxTable infrastructure

---

#### Phase 3: Fix boxes.go Helper Functions

**File:** `internal/formatter/boxes.go`

**Changes Needed:**

1. **DrawHeader() function** (line 56):
```go
// Before:
titleLen := len(titleWithSpaces)

// After:
titleLen := VisibleLength(titleWithSpaces)
```

2. **formatContentLine() function** (lines 110, 115):
```go
// Before:
if len(content) > contentWidth {
    content = content[:contentWidth-3] + "..."
}
padding := contentWidth - len(content)

// After:
if VisibleLength(content) > contentWidth {
    content = truncateToWidth(content, contentWidth-3) + "..."
}
padding := contentWidth - VisibleLength(content)
```

3. **Add helper function** for safe truncation:
```go
// truncateToWidth truncates a string to fit within maxWidth visual characters
func truncateToWidth(s string, maxWidth int) string {
    if VisibleLength(s) <= maxWidth {
        return s
    }
    
    // Truncate rune by rune until we fit
    runes := []rune(s)
    for i := len(runes); i > 0; i-- {
        candidate := string(runes[:i])
        if VisibleLength(candidate) <= maxWidth {
            return candidate
        }
    }
    return ""
}
```

**Functions Fixed:**
- `DrawBox()`
- `DrawHeader()`
- `formatContentLine()`
- `DrawMultilineBox()`
- `DrawBoxWithSections()`

**Current Usage:** These functions are only used in tests, not production. Fixing them prevents future bugs.

**Risk:** Low - not used in production yet

---

### Testing Strategy

#### Unit Tests

1. **VisibleLength Export Test**
   - Verify function is callable from outside formatter package
   - Test with various UTF-8 strings (emoji, checkmarks, multi-byte chars)

2. **Profile List Tests**
   - Test with short profile names
   - Test with long profile names
   - Test with active profile (checkmark indicator)
   - Test empty vs full profiles
   - Verify column alignment in all cases

3. **boxes.go UTF-8 Tests**
   - Add test cases in `boxes_test.go` with UTF-8 strings
   - Test DrawHeader with emoji titles
   - Test formatContentLine with multi-byte content
   - Verify truncation works correctly with UTF-8

#### Integration Tests

1. **Visual Inspection**
   ```bash
   ./agent-smith profile list
   ./agent-smith target list
   ./agent-smith link status
   ```
   - All columns should align perfectly
   - Checkmarks display correctly
   - No jagged edges

2. **Consistency Check**
   - Profile list should look visually consistent with target list
   - Same box-drawing characters and style
   - Same column alignment behavior

3. **Edge Cases**
   - Very long profile names (test truncation)
   - Many profiles (test scrolling behavior)
   - No profiles (empty state)
   - Profile with all component types
   - Profile with no components (empty)

#### Regression Testing

Run full test suite to ensure no breaking changes:
```bash
go test ./...
```

All existing tests must pass.

---

## Implementation Plan

### Files to Modify

| File | Lines Changed | Type | Priority |
|------|---------------|------|----------|
| `internal/formatter/box_table.go` | ~5 | Modify | High |
| `internal/formatter/box_table_test.go` | ~3 | Modify | High |
| `main.go` | -80, +30 | Replace | High |
| `internal/formatter/boxes.go` | ~20 | Modify | Medium |
| `internal/formatter/boxes_test.go` | +30 | Add | Medium |

### Estimated Code Changes

- **Total additions:** ~60 lines
- **Total deletions:** ~85 lines
- **Net change:** -25 lines (code reduction!)
- **Modified functions:** 8
- **New functions:** 1 (truncateToWidth helper)

### Implementation Order

1. **Phase 1:** Export VisibleLength() (enables everything else)
2. **Phase 2:** Fix profile list command (highest user impact)
3. **Phase 3:** Fix boxes.go functions (future-proofing)
4. **Testing:** Run full test suite and visual validation

### Time Estimate

- Phase 1: 15 minutes (straightforward rename + update calls)
- Phase 2: 45 minutes (refactor to use BoxTable + test)
- Phase 3: 30 minutes (fix width calculations + add tests)
- Testing: 20 minutes (run tests, visual validation)
- **Total:** ~2 hours

---

## Risk Assessment

### Technical Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| BoxTable breaks existing formatting | Low | Medium | Use proven pattern from target list |
| Tests fail after changes | Low | Low | Run tests after each phase |
| Visual regression in other commands | Very Low | Medium | Only modify profile list code path |
| Performance impact | Very Low | Low | VisibleLength already used in BoxTable |

### Rollback Plan

If issues arise:
1. Each phase is independent - can revert individual changes
2. Git history allows easy rollback
3. Tests catch issues before merge
4. Visual validation catches display issues

---

## Dependencies

### Required

- ✅ `github.com/mattn/go-runewidth v0.0.19` (already in go.mod)
- ✅ Existing `formatter.NewBoxTable()` implementation
- ✅ Existing test infrastructure

### No New Dependencies Needed

All required infrastructure already exists in the codebase.

---

## Open Questions

### Resolved

1. **Q:** Should we use BoxTable or fix manual formatting?
   - **A:** Use BoxTable for consistency with target list and link status

2. **Q:** Should boxes.go be fixed now or later?
   - **A:** Fix now for comprehensive solution (Phase 3)

3. **Q:** What visual style for profile list?
   - **A:** Two-column table with headers (Option A) for consistency

4. **Q:** What column header text?
   - **A:** Keep it simple: "Profile" and "Components"

### None Remaining

All design decisions have been made.

---

## Acceptance Criteria

### Must Have

- [ ] Profile list columns align perfectly
- [ ] Checkmark indicator doesn't break alignment
- [ ] VisibleLength() function is exported
- [ ] All existing tests pass
- [ ] Visual consistency with target list
- [ ] Code is cleaner (net reduction in LOC)

### Should Have

- [ ] boxes.go functions fixed for UTF-8
- [ ] New tests for UTF-8 edge cases
- [ ] Documentation updated if needed

### Nice to Have

- [ ] Performance benchmarks (if time permits)
- [ ] Additional edge case tests

---

## Future Considerations

### Long-term Improvements

1. **Deprecate manual box-drawing** - Establish BoxTable as standard
2. **Add linting rule** - Warn when `len()` used on display strings
3. **Documentation** - Document when to use `VisibleLength()` vs `len()`
4. **Style guide** - Document formatter package as single source of truth

### Related Work

- Consider audit of other CLI formatting for consistency
- Document box-drawing patterns for future contributors
- Add examples to formatter package documentation

---

## Appendix

### Current Commands Audit

| Command | Format Method | Status | Notes |
|---------|---------------|--------|-------|
| `profile list` | Manual box-drawing | ❌ Broken | This PRD fixes it |
| `target list` | BoxTable | ✅ Works | Reference implementation |
| `link status` | BoxTable | ✅ Works | Multi-column example |
| `profile show` | Bullet list | ✅ Works | Simple format, no issues |
| `materialize list` | Bullet list | ✅ Works | Simple format, no issues |
| `status` | Key-value pairs | ✅ Works | Simple format, no issues |

### Reference Implementation

**Good Example:** `target list` command (main.go:2100)
```go
table := formatter.NewBoxTable(os.Stdout, []string{"Status", "Target", "Type", "Location"})
for _, target := range allTargets {
    table.AddRow([]string{statusSymbol, target.name, targetType, target.baseDir})
}
table.Render()
```

This is the pattern we're adopting for profile list.

---

## Version History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-02-02 | Agent Smith | Initial PRD |

---

## Ralphy Task Breakdown

This PRD can be executed by autonomous agents with the following parallel task structure:

### Task 1: Export VisibleLength Function
**Depends on:** None  
**Files:** `internal/formatter/box_table.go`, `internal/formatter/box_table_test.go`  
**Estimated time:** 15 minutes  
**Validation:** Function callable from main.go, tests pass

### Task 2: Refactor Profile List Command
**Depends on:** Task 1 (needs exported VisibleLength)  
**Files:** `main.go`  
**Estimated time:** 45 minutes  
**Validation:** Profile list displays correctly, columns aligned

### Task 3: Fix boxes.go Width Calculations
**Depends on:** Task 1 (needs exported VisibleLength)  
**Files:** `internal/formatter/boxes.go`, `internal/formatter/boxes_test.go`  
**Estimated time:** 30 minutes  
**Validation:** UTF-8 tests pass, functions handle multi-byte chars

### Task 4: Integration Testing
**Depends on:** Tasks 1, 2, 3 (all phases complete)  
**Files:** None (validation only)  
**Estimated time:** 20 minutes  
**Validation:** All tests pass, visual inspection confirms alignment

---

**Total Estimated Delivery Time:** 2 hours  
**Complexity:** Medium  
**Priority:** High (user-facing issue)  
**Risk Level:** Low (using proven patterns)
