# PRD: Universal Target Support for Materialize

**Created**: 2026-02-03 13:34 UTC

---

## Introduction

Add a `universal` target that materializes components to a `.agents/` directory in the project root. This provides a target-agnostic location for materialized components that can be used by any AI coding assistant, without tying components to a specific tool's directory structure.

This feature enables developers to maintain a single materialized copy of components that multiple AI assistants can reference, reducing duplication and simplifying multi-tool workflows.

## Goals

- Provide a target-agnostic `.agents/` directory for materialized components
- Support materialize operations to the universal target (link support deferred to future work)
- Follow the same subdirectory structure as other targets (skills/, agents/, commands/)
- Keep universal as an opt-in feature (not included in `--target all`)
- Maintain consistency with existing target patterns and CLI conventions

## User Stories

- [ ] Story-001: As a developer, I want to materialize components to a universal `.agents/` directory so that multiple AI coding assistants can access them without duplication.

  **Acceptance Criteria:**
  - Universal target creates `.agents/` directory in project root
  - Directory structure matches other targets: `.agents/skills/`, `.agents/agents/`, `.agents/commands/`
  - User can specify `--target universal` on materialize commands
  - Components materialize successfully to `.agents/` subdirectories
  
  **Testing Criteria:**
  **Unit Tests:**
  - UniversalTarget returns correct directory paths for all component types
  - GetName() returns "universal"
  - GetComponentDir() properly switches on component types
  
  **Integration Tests:**
  - Materialize skill to universal target creates correct directory structure
  - Materialize multiple component types to universal target

- [ ] Story-002: As a developer, I want the universal target to be opt-in only so that it doesn't interfere with my existing multi-target workflows.

  **Acceptance Criteria:**
  - `--target all` does NOT include universal target
  - Universal target only used when explicitly specified with `--target universal`
  - `.agents` directory not added to project detection markers
  - Existing workflows with opencode/claudecode/copilot unchanged
  
  **Testing Criteria:**
  **Unit Tests:**
  - Verify `--target all` resolves to opencode, claudecode, copilot only
  
  **Integration Tests:**
  - Materialize with `--target all` does not create `.agents/` directory
  - Materialize with `--target universal` creates only `.agents/` directory

- [ ] Story-003: As a developer, I want to see the universal target in help text and error messages so that I know it's available.

  **Acceptance Criteria:**
  - Help text for materialize commands includes "universal" in target options
  - Example commands show usage of `--target universal`
  - Error messages for invalid targets mention universal as a valid option
  - Help text includes a note explaining universal is for target-agnostic storage
  
  **Testing Criteria:**
  **Unit Tests:**
  - Error message helpers include "universal" in valid targets list
  
  **Integration Tests:**
  - Run `agent-smith materialize --help` and verify universal appears in target options

- [ ] Story-004: As a developer, I want to see universal target components in status and list commands so that I can track what's materialized.

  **Acceptance Criteria:**
  - `agent-smith materialize status` shows universal target components with "Universal (.agents/)" label
  - `agent-smith materialize list` includes universal target in output
  - `agent-smith materialize info` displays universal target information
  - Status output uses consistent formatting with other targets
  
  **Testing Criteria:**
  **Integration Tests:**
  - Materialize to universal, run status command, verify output shows Universal (.agents/)
  - Materialize to universal, run list command, verify component appears

- [ ] Story-005: As a developer, I want `.agents/` mentioned in project detection error messages so that I'm aware of it as an option.

  **Acceptance Criteria:**
  - Project detection error message includes `.agents/` in list of project markers
  - Error message describes universal as target-agnostic option
  - Documentation note explains when to use universal vs specific targets
  
  **Testing Criteria:**
  **Integration Tests:**
  - Trigger project detection error and verify `.agents/` appears in output

## Functional Requirements

### Core Target Implementation

- FR-1: The system SHALL implement a `UniversalTarget` struct in `pkg/config/universal_target.go` that implements the `Target` interface
- FR-2: The `UniversalTarget` SHALL return `.agents/` as the base directory within a project root
- FR-3: The `UniversalTarget` SHALL support subdirectories: `skills/`, `agents/`, `commands/`
- FR-4: The `GetName()` method SHALL return "universal" for the target name

### Target Manager Integration

- FR-5: The `target_manager.go` SHALL recognize "universal" as a valid target type in `NewTarget()`
- FR-6: The `GetTargetDirectory()` function in `detection.go` SHALL return `<project-root>/.agents` for universal target
- FR-7: The universal target SHALL NOT be included in `DetectAllTargets()` (used by link commands)
- FR-8: The universal target SHALL NOT be included in target lists for `--target all` behavior

### Materialize Command Integration

- FR-9: All materialize commands SHALL accept `--target universal` as a valid option
- FR-10: The materialize service SHALL correctly route universal target operations to `.agents/` directory
- FR-11: Target label display SHALL show "Universal (.agents/)" for universal target in status/list/info outputs
- FR-12: Example help text SHALL include at least one example using `--target universal`

### CLI and Error Messaging

- FR-13: All materialize command flag descriptions SHALL list "universal" as a valid target option
- FR-14: Error messages for invalid targets SHALL include "universal" in the list of valid targets
- FR-15: Help text SHALL include a note explaining universal is for target-agnostic component storage
- FR-16: Project detection error messages SHALL mention `.agents/` as an available project marker option

### Directory Structure

- FR-17: When materializing to universal, the system SHALL create `.agents/` in the project root if it doesn't exist
- FR-18: The system SHALL create component-type subdirectories (`skills/`, `agents/`, `commands/`) as needed
- FR-19: The directory structure SHALL follow the same metadata patterns as other targets

## Non-Goals

The following are explicitly OUT OF SCOPE for this initial implementation:

- No link command support for universal target (materialize only for now)
- No global `~/.agents` directory (project-local only)
- No auto-detection of `.agents` directory in project markers (opt-in only)
- No changes to `--target all` behavior (remains opencode/claudecode/copilot only)
- No changes to link commands or link service
- No custom detection configuration for universal target
- No profile-aware operations for universal target (profiles still use their configured targets)

## Technical Implementation Notes

### File Structure

**New Files:**
- `pkg/config/universal_target.go` - UniversalTarget implementation
- `pkg/config/universal_target_test.go` - Unit tests

**Modified Files:**
- `pkg/project/detection.go` - Add universal case to GetTargetDirectory(), update error message
- `pkg/config/target_manager.go` - Add TargetUniversal constant, update NewTarget()
- `pkg/services/materialize/service.go` - Update target labels and help text
- `cmd/root.go` - Update flag descriptions and examples
- `pkg/errors/helpers.go` - Update error messages

### Key Design Patterns

The UniversalTarget implementation follows the same pattern as CopilotTarget:
- Implements full Target interface
- Constructor pattern with NewUniversalTarget()
- Component directory resolution via GetComponentDir()
- Consistent error handling and return patterns

Unlike OpenCode/ClaudeCode targets, UniversalTarget:
- Has no global directory (project-local only)
- Is not included in DetectAllTargets() (materialize-only)
- Is opt-in explicit (not auto-detected)

## Testing Strategy

### Unit Tests (Required)
- UniversalTarget struct methods (GetBaseDir, GetSkillsDir, GetAgentsDir, GetCommandsDir, GetComponentDir, GetName)
- NewTarget("universal") returns UniversalTarget instance
- Error messages include "universal" in valid targets list

### Integration Tests (Required)
- Materialize skill to universal target creates `.agents/skills/<name>/`
- Materialize agent to universal target creates `.agents/agents/<name>/`
- Materialize command to universal target creates `.agents/commands/<name>/`
- `--target all` does NOT create `.agents/` directory
- `agent-smith materialize status` displays Universal (.agents/) label
- `agent-smith materialize list` includes universal target components
- Help text displays universal in target options

### Manual Testing
- Build and run `agent-smith materialize skill test-skill --target universal`
- Verify `.agents/skills/test-skill/` directory created
- Run `agent-smith materialize status` and verify output
- Run `agent-smith materialize --help` and verify universal appears
- Test invalid target error message includes universal

## Success Criteria

The feature is considered complete when:

1. ✅ Universal target successfully materializes components to `.agents/` directory
2. ✅ All unit tests pass for UniversalTarget implementation
3. ✅ Integration tests verify materialize operations work correctly
4. ✅ `--target universal` is recognized by all materialize commands
5. ✅ `--target all` behavior unchanged (does not include universal)
6. ✅ Help text and error messages include universal as an option
7. ✅ Status and list commands display universal target components
8. ✅ Project builds successfully with no test regressions
9. ✅ Help text includes explanatory note about universal target purpose

## Future Enhancements (Not in Scope)

- Link command support for universal target with global `~/.agents` directory
- Auto-detection of `.agents` in project markers
- Universal target included in `--target all` (based on user feedback)
- Profile-aware universal target operations
- Custom detection configuration for universal target
