# PRD: Enhanced Component Detection for Agent Smith

## Introduction

The current component detection system in Agent Smith only recognizes specific filename patterns (SKILL.md, AGENT.md, COMMAND.md), which fails to detect components in repositories using different organizational patterns like the wshobson/agents repository. This enhancement will expand detection capabilities to support diverse repository structures while maintaining backward compatibility.

## Problem Analysis

### Current State
- **Skills Detection**: Works correctly for `SKILL.md` files
- **Agents Detection**: Only detects files named exactly `AGENT.md` 
- **Commands Detection**: Only detects files named exactly `COMMAND.md`
- **Result**: Only 129 skills detected in wshobson/agents, 0 agents/commands detected

### Repository Structure Analysis
The wshobson/agents repository uses this pattern:
- Skills: `plugins/category/skills/skill-name/SKILL.md` ✅ (working)
- Agents: `plugins/category/agents/agent-name.md` ❌ (not detected)
- Commands: `plugins/category/commands/command-name.md` ❌ (not detected)

## Goals

- Enable detection of skills, agents, and commands across different repository organizational patterns
- Support wshobson/agents repository structure specifically
- Maintain 100% backward compatibility with existing SKILL.md detection
- Improve reliability of `add-all` command for bulk component downloads

## User Stories

- [ ] Story-001: As a user, I want add-all command to detect all components in a repository so that I can bulk download skills, agents, and commands regardless of how they're organized.

  **Acceptance Criteria:**
  - System detects SKILL.md files in skills directories (existing behavior preserved)
  - System detects .md files in /agents/ directories as agent components
  - System detects .md files in /commands/ directories as command components
  - Bulk download processes all detected component types successfully
  - Existing functionality remains unchanged for repositories with current structure

- [ ] Story-002: As a user, I want correct component names so that downloaded components are properly labeled and organized.

  **Acceptance Criteria:**
  - Skills use directory name as component name (existing behavior)
  - Agents use filename without .md extension as component name
  - Commands use filename without .md extension as component name
  - Downloaded components have correct names in ~/.agents/ directory structure

- [ ] Story-003: As a developer, I want maintainable detection logic so that future repository patterns can be added easily.

  **Acceptance Criteria:**
  - Detection logic organized by component type with clear separation
  - Path-based patterns used instead of just filename matching
  - Comments explain detection patterns for future modifications
  - No hard-coded assumptions about directory depth or naming conventions

## Functional Requirements

- FR-1: The system must detect skills using existing SKILL.md filename pattern (unchanged)
- FR-2: The system must detect agents by finding .md files in paths containing "/agents/"
- FR-3: The system must detect commands by finding .md files in paths containing "/commands/"
- FR-4: The system must extract component names appropriately:
  - Skills: Use directory name (existing behavior)
  - Agents/Commands: Use filename without .md extension
- FR-5: The system must preserve all existing functionality for repositories using current patterns
- FR-6: The system must use same error handling approach as existing code

## Technical Implementation

### Location and Scope
- **File**: `main.go`
- **Function**: `detectComponentsInRepo()` (lines 352-393)
- **Approach**: Add new detection logic while preserving existing AGENT.md/COMMAND.md checks

### Enhanced Detection Logic

#### Add Agent Detection (after line 378)
```go
// Additional agent detection for .md files in /agents/ paths
if strings.HasSuffix(fileName, ".md") && strings.Contains(relPath, "/agents/") {
    componentName := strings.TrimSuffix(fileName, ".md")
    if componentName == "" || componentName == "." {
        componentName = "root-agent"
    }
    components = append(components, DetectedComponent{
        Type:       ComponentAgent,
        Name:       componentName,
        Path:       relPath,
        SourceFile: fileName,
    })
}
```

#### Add Command Detection (after line 392)
```go
// Additional command detection for .md files in /commands/ paths
if strings.HasSuffix(fileName, ".md") && strings.Contains(relPath, "/commands/") {
    componentName := strings.TrimSuffix(fileName, ".md")
    if componentName == "" || componentName == "." {
        componentName = "root-command"
    }
    components = append(components, DetectedComponent{
        Type:       ComponentCommand,
        Name:       componentName,
        Path:       relPath,
        SourceFile: fileName,
    })
}
```

### Implementation Strategy
1. **Preserve existing logic**: Keep all current detection code unchanged
2. **Add new patterns**: Implement path-based detection for agents/commands
3. **Consistent error handling**: Use same error handling patterns as existing code
4. **Component name extraction**: Follow appropriate naming convention for each type

## Non-Goals

- No changes to individual component download commands (add-skill, add-agent, add-command)
- No configuration file or user-configurable detection patterns in this version
- No support for custom component types beyond skills, agents, and commands
- No changes to metadata or lock file formats
- No modifications to linking or execution functionality
- No replacement of existing detection logic (only additive changes)

## Success Criteria

- Add-all command successfully detects and downloads all components from wshobson/agents repository
- Expected detection: ~129 skills, ~50+ agents, ~20+ commands
- Existing repositories with SKILL.md files continue to work unchanged
- Component names extracted correctly for all three types
- No breaking changes to existing functionality
- Code is maintainable and well-documented

## Testing Strategy

### Unit Tests
- Test path-based detection for agents in various directory structures
- Test path-based detection for commands in various directory structures
- Verify SKILL.md detection remains unchanged
- Test component name extraction for all three types
- Edge cases: empty directories, malformed paths, mixed patterns

### Integration Tests
- Test add-all command with wshobson/agents repository structure
- Verify all component types are detected and downloaded correctly
- Test with existing repositories to ensure no regression
- Confirm downloaded components have correct names and structure

### Validation Tests
- Confirm 129 skills detected (same as current)
- Verify agent detection finds expected number of .md files in /agents/ paths
- Verify command detection finds expected number of .md files in /commands/ paths
- Test bulk download processes all detected components successfully

## Risk Assessment

### Low Risk
- **Backward Compatibility**: Existing logic preserved unchanged
- **Scope**: Limited to adding new detection patterns only
- **Error Handling**: Uses established patterns from existing code

### Mitigation
- Comprehensive testing before deployment
- Fallback to existing behavior if issues detected
- Clear documentation of new detection patterns

## Implementation Checklist

- [ ] Modify detectComponentsInRepo function in main.go
- [ ] Add agent detection logic for .md files in /agents/ paths
- [ ] Add command detection logic for .md files in /commands/ paths
- [ ] Test component name extraction for all types
- [ ] Verify backward compatibility with existing repositories
- [ ] Run integration tests with wshobson/agents repository
- [ ] Confirm add-all command detects all component types
- [ ] Validate downloaded components have correct structure and names