# PRD: Component Detection Path Calculation Fix

## Introduction

Fix Agent Smith component detection logic that incorrectly returns parent directory instead of immediate containing directory for SKILL.md files during bulk download.

## Goals

- Fix component path calculation to prevent cross-plugin contamination
- Ensure each skill directory contains only its own components
- Maintain existing functionality for individual downloads

## User Stories

- [ ] Story-001: As a user running bulk download, I want each skill directory to contain only its own skills so that my storage isn't wasted with duplicate files. - When detecting plugins/X/skills/Y/SKILL.md, return plugins/X/skills/Y/ instead of plugins/X/skills/, This prevents copying all sibling skills into each skill directory, Add logic to split path and get immediate parent directory.

- [ ] Story-002: As a user downloading accessibility-compliance plugin, I want only 2 skills downloaded so that I get clean, isolated skill directories. - Test accessibility-compliance plugin download specifically, Verify only its own skills are copied, Check directory structure matches expected output, Validate metadata shows correct component count (2 skills).

- [ ] Story-003: As a developer maintaining Agent Smith, I want bulk download to process plugins independently so that the system works reliably across different repository structures. - Ensure each plugin is processed separately during bulk download, Prevent cross-plugin contamination, Confirm total skill count matches repository (127 across all plugins).

## Functional Requirements

- FR-1: Component detection must identify exact directory containing SKILL.md files
- FR-2: For plugins/X/skills/Y/SKILL.md, component path must be plugins/X/skills/Y/
- FR-3: Bulk download must not copy sibling directories during component copying
- FR-4: Individual skill downloads must continue working as before
- FR-5: Metadata generation must report component count per plugin, not total repository components

## Non-Goals

- No changes to individual skill download functionality
- No modifications to repository structure validation
- No changes to metadata file format or content

## Technical Solution

### Files to Change
- `main.go` - Fix `componentDir` calculation in `detectComponentForPattern()` function (around line 521)

### Current Issue
```go
// Current (incorrect)
componentDir := filepath.Dir(fullRelPath)
// Returns plugins/X/skills/ for plugins/X/skills/Y/SKILL.md
```

### Required Fix
```go
// Fixed (correct)
if rd.matchesExactFile(fileName, pattern.ExactFiles) {
    // For plugins/X/skills/Y/SKILL.md, we want plugins/X/skills/Y/
    // Not plugins/X/skills/
    parts := strings.Split(fullRelPath, string(filepath.Separator))
    if len(parts) >= 2 {
        componentDir = strings.Join(parts[:len(parts)-1], string(filepath.Separator))
    }
} else {
    componentDir = filepath.Dir(fullRelPath)
}
```

## Success Criteria

- accessibility-compliance plugin downloads only its 2 skills
- No cross-plugin contamination in any downloaded skill
- Bulk download completes with expected total of 127 skills across all plugins
- Individual skill downloads continue to work as before
- Metadata shows correct component count per plugin