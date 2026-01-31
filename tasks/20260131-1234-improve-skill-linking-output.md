# PRD: Improve Skill Linking Output Clarity

**Created**: 2026-01-31 12:34 UTC

---

## Introduction

Improve the clarity of the skill linking output messages to make it easier for users to understand what's being linked, where it's being linked to, and the source location. Currently, the output repeats "Successfully linked" for each target and shows the source path at the very end, creating confusion about the relationship between source and targets.

## Goals

- Group linking output by skill to reduce repetition and improve readability
- Show skill name once with all targets listed underneath
- Maintain minimal output verbosity while showing essential information
- Display errors inline with successes for each skill
- Make source path easily identifiable for each skill

## User Stories

- [x] Story-001: As a user installing skills, I want to see each skill grouped with all its targets so that I can quickly understand where each skill was linked.

  **Acceptance Criteria:**
  - Output shows skill name once at the top
  - All targets are listed underneath with arrow indicators (→)
  - Source path appears at the end of each skill's output group
  - Format: "Linked <type> '<name>':\n  → Target1: path\n  → Target2: path\n  Source: path"
  
  **Testing Criteria:**
  **Unit Tests:**
  - Test output formatting for single target scenarios
  - Test output formatting for multiple target scenarios
  - Test handling of linkedTargets slice
  
  **Integration Tests:**
  - Test linking to multiple targets (OpenCode + Claude Code)
  - Verify output format matches expected structure
  
  **Component Browser Tests:**
  - Not applicable (CLI output)

- [x] Story-002: As a user, I want to see linking errors inline with successful targets so that I can immediately identify which targets failed for each skill.

  **Acceptance Criteria:**
  - Errors appear in the same group as the skill being linked
  - Failed targets show with a clear error indicator (✗)
  - Successful targets show with arrow indicator (→)
  - Error messages appear directly after the skill group
  
  **Testing Criteria:**
  **Unit Tests:**
  - Test error tracking in linkedTargets structure
  - Test output formatting with mixed success/failure states
  
  **Integration Tests:**
  - Test partial failures (some targets succeed, others fail)
  - Verify error messages are grouped correctly
  
  **Component Browser Tests:**
  - Not applicable (CLI output)

- [x] Story-003: As a user linking to a single target, I want a simplified single-line output so that I don't see unnecessary grouping for just one target.

  **Acceptance Criteria:**
  - Single target uses compact format: "Linked <type> '<name>' → <target>"
  - Target path shown on next line
  - Source path shown on third line
  - No bullet points or grouping indicators for single target
  
  **Testing Criteria:**
  **Unit Tests:**
  - Test single target output formatting
  - Test path display for single target
  
  **Integration Tests:**
  - Test linking to only OpenCode
  - Test linking to only Claude Code
  
  **Component Browser Tests:**
  - Not applicable (CLI output)

## Functional Requirements

- FR-1: The system SHALL track linked targets during the linking loop including target name and destination path
- FR-2: The system SHALL determine output format based on number of successful targets (1 vs multiple)
- FR-3: For single target, the system SHALL display compact single-line format with paths underneath
- FR-4: For multiple targets, the system SHALL display grouped format with skill name as header
- FR-5: The system SHALL use arrow indicator (→) for successful target links
- FR-6: The system SHALL display source path at the end of each skill's output group
- FR-7: The system SHALL preserve existing error handling and display errors inline

## Non-Goals

- No changes to the actual linking logic or symlink creation
- No changes to target configuration or detection
- No colored output or ANSI formatting (keep plain text)
- No verbose/quiet flags or output level control (future enhancement)
- No changes to bulk linking operations output format (focus on single component linking)
- No internationalization of output messages

## Technical Implementation Notes

### Current Code Structure

The linking output is generated in `internal/linker/linker.go` in the `LinkComponent()` function around lines 156-164.

Current problematic pattern:
```go
for _, target := range cl.targets {
    // ... linking logic ...
    fmt.Printf("Successfully linked %s '%s' to %s\n", componentType, componentName, targetName)
    fmt.Printf("  Target: %s\n", dstDir)
    successCount++
}
if successCount > 0 {
    fmt.Printf("  Source: %s\n", srcDir)
}
```

### Proposed Changes

1. Track linked targets during the loop:
```go
linkedTargets := []struct {
    name string
    path string
}{}

for _, target := range cl.targets {
    // ... existing linking logic ...
    linkedTargets = append(linkedTargets, struct {
        name string
        path string
    }{target.GetName(), dstDir})
    successCount++
}
```

2. Display results after the loop based on count:
```go
if successCount > 0 {
    if successCount == 1 {
        // Single target format
        fmt.Printf("Linked %s '%s' → %s\n", componentType, componentName, linkedTargets[0].name)
        fmt.Printf("  Target: %s\n", linkedTargets[0].path)
        fmt.Printf("  Source: %s\n", srcDir)
    } else {
        // Multiple targets format
        fmt.Printf("Linked %s '%s':\n", componentType, componentName)
        for _, lt := range linkedTargets {
            fmt.Printf("  → %s: %s\n", lt.name, lt.path)
        }
        fmt.Printf("  Source: %s\n", srcDir)
    }
}
```

### Example Output

**Before (confusing):**
```
Successfully linked skills 'accessibility-compliance' to OpenCode
  Target: /Users/tgaines/.config/opencode/skills/accessibility-compliance
Successfully linked skills 'accessibility-compliance' to Claude Code
  Target: /Users/tgaines/.claude/skills/accessibility-compliance
  Source: /Users/tgaines/.agent-smith/skills/accessibility-compliance
```

**After (clear - multiple targets):**
```
Linked skills 'accessibility-compliance':
  → OpenCode: /Users/tgaines/.config/opencode/skills/accessibility-compliance
  → Claude Code: /Users/tgaines/.claude/skills/accessibility-compliance
  Source: /Users/tgaines/.agent-smith/skills/accessibility-compliance
```

**After (clear - single target):**
```
Linked skills 'accessibility-compliance' → OpenCode
  Target: /Users/tgaines/.config/opencode/skills/accessibility-compliance
  Source: /Users/tgaines/.agent-smith/skills/accessibility-compliance
```

## Files to Modify

- `internal/linker/linker.go` - Update `LinkComponent()` function output formatting (lines ~156-164)

## Testing Strategy

1. Test with single target configuration (OpenCode only)
2. Test with multiple targets configuration (OpenCode + Claude Code)
3. Test with partial failures (one target succeeds, one fails)
4. Verify no changes to linking logic itself
5. Manual testing of output readability
