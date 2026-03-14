# PRD: Fix Summary Table Alignment with ANSI Color Codes

## Problem Statement

The `SummaryTableFormat` in `pkg/styles/styles.go` produces misaligned table rows when colored symbols are used because ANSI escape codes are being counted as visible characters in padding calculations.

**Current Broken Output:**
```
┌──────────────────────────────────────────────────────────────────────────────┐
│                                Update Summary                                │
├──────────────────────────────────────────────────────────────────────────────┤
│ Total components checked:      236                                         │
│ ✓ Already up to date: 236                                         │  ← misaligned
│ ✓ Successfully updated: 0                                           │  ← misaligned
└──────────────────────────────────────────────────────────────────────────────┘
```

**Expected Fixed Output:**
```
┌──────────────────────────────────────────────────────────────────────────────┐
│                                Update Summary                                │
├──────────────────────────────────────────────────────────────────────────────┤
│ Total components checked: 236                                                │
│ ✓ Already up to date: 236                                                    │
│ ✓ Successfully updated: 0                                                    │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Root Cause

In `pkg/styles/styles.go` lines 122 and 134:
- Using `fmt.Sprintf("│ %-*s %-*v│", ...)` which counts ANSI codes as visible characters
- Not using the established `formatter.VisibleLength()` pattern used everywhere else in the codebase

## Goals

1. Fix alignment issues in summary tables when using colored symbols
2. Use the established pattern from `internal/formatter/box_table.go` and `boxes.go`
3. Add tests to prevent regression
4. Ensure consistency with existing codebase formatting patterns

## Non-Goals

- Changing the visual style or layout of the tables
- Refactoring other formatting code
- Adding new formatting features

## Solution Design

### Established Pattern (from box_table.go and boxes.go)

```go
// 1. Calculate visible length excluding ANSI codes
visLen := VisibleLength(content)

// 2. Calculate padding needed
padding := targetWidth - visLen

// 3. Manually build string with proper padding
result := content + strings.Repeat(" ", padding)
```

### Implementation

**File 1: `pkg/styles/styles.go`**

1. Add import: `"github.com/tgaines/agent-smith/internal/formatter"`
2. Update `AddRow` method to use ANSI-aware padding
3. Update `AddRowWithSymbol` method to use ANSI-aware padding

**File 2: `pkg/styles/styles_test.go`**

1. Add `TestSummaryTableBuilderANSIAlignment` test
2. Verify all rows have consistent visible width
3. Verify rows end with proper border pattern

## Technical Details

### Before (Broken)
```go
func (b *SummaryTableBuilder) AddRowWithSymbol(symbol, label string, value interface{}) *SummaryTableBuilder {
    labelWidth := 30
    valueWidth := b.width - labelWidth - 6
    labelWithSymbol := fmt.Sprintf("%s %s", symbol, label)
    row := fmt.Sprintf("│ %-*s %-*v│", labelWidth, labelWithSymbol, valueWidth, value)
    b.rows = append(b.rows, row)
    return b
}
```

### After (Fixed)
```go
func (b *SummaryTableBuilder) AddRowWithSymbol(symbol, label string, value interface{}) *SummaryTableBuilder {
    // Format the complete content
    content := fmt.Sprintf("%s %s %v", symbol, label, value)
    
    // Calculate visible length and padding (following box_table.go pattern)
    innerWidth := b.width - 2  // Subtract borders
    contentWidth := innerWidth - 2  // Subtract space padding on each side
    visLen := formatter.VisibleLength(content)
    padding := contentWidth - visLen
    
    // Build row with proper ANSI-aware padding (following boxes.go line 135 pattern)
    row := fmt.Sprintf("│ %s%s │", content, strings.Repeat(" ", padding))
    b.rows = append(b.rows, row)
    return b
}
```

## Testing Strategy

### Unit Tests
- Test with colors enabled to verify ANSI code handling
- Test with colors disabled to verify plain text
- Verify all rows have same visible width (80 chars)
- Verify rows end with proper border pattern

### Manual Testing
```bash
# Test with colors
./agent-smith update all

# Test without colors
NO_COLOR=1 ./agent-smith update all
```

### Visual Verification
- All vertical borders `│` should align perfectly
- Values should have consistent spacing
- Table should be exactly 80 characters wide (visible)

## Success Criteria

- [x] Unit tests pass with colored symbols
- [x] Manual test shows proper alignment in `update all` command
- [x] Visual inspection confirms borders align
- [x] Works with both colors enabled and disabled
- [x] No regression in other table formatting

## Implementation Complete ✅

**Date:** 2025-02-02

**Changes Made:**
1. Updated `pkg/styles/styles.go`:
   - Fixed `AddRow` method to use `formatter.VisibleLength()` for ANSI-aware padding
   - Fixed `AddRowWithSymbol` method to use `formatter.VisibleLength()` for ANSI-aware padding
   - Applied established pattern from `internal/formatter/box_table.go` and `boxes.go`

2. Updated `pkg/styles/styles_test.go`:
   - Added `TestSummaryTableBuilderANSIAlignment` test
   - Verifies all rows have same visible width (80 chars)
   - Verifies proper border alignment with ANSI color codes

**Test Results:**
```
PASS: TestSummaryTableBuilder
PASS: TestSummaryTableBuilderWithSymbol
PASS: TestSummaryTableBuilderANSIAlignment
```

**Visual Verification:**
- Tested with `./agent-smith update all` - borders perfectly aligned ✓
- Tested with `NO_COLOR=1 ./agent-smith update all` - works without colors ✓
- All vertical borders `│` align perfectly ✓

## Impact

**Users Affected:** All users running `agent-smith update all` command

**Severity:** Low (cosmetic issue, doesn't affect functionality)

**Frequency:** Every time update command is run with results

## Rollout Plan

1. Implement fix in `pkg/styles/styles.go`
2. Add tests in `pkg/styles/styles_test.go`
3. Run unit tests to verify
4. Manual testing with real update command
5. Merge to main branch

## Alternatives Considered

1. **Strip ANSI codes before formatting** - Rejected because we want colored output
2. **Use a table library** - Rejected because we already have the right pattern established
3. **Fixed-width columns** - Rejected because current approach is more flexible

## References

- Existing pattern: `internal/formatter/box_table.go` lines 110-113, 125-128
- Existing pattern: `internal/formatter/boxes.go` line 135
- VisibleLength function: `internal/formatter/box_table.go` lines 30-64
