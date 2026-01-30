# PRD: Unlink by Target Support

## Introduction

Add `--target` flag support to all unlink commands to provide feature parity with the link command. Currently, the link command allows users to specify which target (opencode, claudecode, custom targets, or all) to link components to, but the unlink command always operates on ALL detected targets. This creates an inconsistent user experience and prevents users from selectively unlinking components from specific targets while maintaining links to others.

## Goals

- Provide feature parity between link and unlink commands
- Enable selective unlinking from specific targets
- Maintain backward compatibility with current default behavior
- Improve user workflow flexibility when managing component links
- Deliver consistent CLI experience across all linking operations

## User Stories

- [x] Story-001: As a developer, I want to specify which target to unlink from so that I can maintain links to some targets while removing others.

  **Acceptance Criteria:**
  - Add --target flag to all unlink commands (skill/agent/command, skills/agents/commands, all)
  - Flag accepts same values as link command (opencode, claudecode, all, custom target names)
  - Default behavior unlinks from all targets when flag is omitted for backward compatibility

- [x] Story-002: As a user running unlink with --target flag, I want clear error messages when the target doesn't exist so I know what went wrong.

  **Acceptance Criteria:**
  - Validate specified target exists before attempting unlink
  - Show helpful error message with list of available targets when invalid target specified
  - Exit with non-zero status code on validation failure

- [x] Story-003: As a developer, I want the unlink command to display which target(s) were affected so I can verify the operation worked correctly.

  **Acceptance Criteria:**
  - Show target name in success messages for each unlink operation
  - Display skip messages when component not found in specified target
  - Maintain current output format for unlink operations

- [x] Story-004: As a developer, I want `unlink skill <name> --target <target>` to work so I can unlink individual skills from specific targets.

  **Acceptance Criteria:**
  - Add --target flag to unlinkSkillCmd in cmd/root.go
  - Pass targetFilter parameter to handleUnlink function
  - Update handler to use NewComponentLinkerWithFilter instead of NewComponentLinker

- [x] Story-005: As a developer, I want `unlink agent <name> --target <target>` to work so I can unlink individual agents from specific targets.

  **Acceptance Criteria:**
  - Add --target flag to unlinkAgentCmd in cmd/root.go
  - Pass targetFilter parameter to handleUnlink function
  - Update handler to use NewComponentLinkerWithFilter

- [x] Story-006: As a developer, I want `unlink command <name> --target <target>` to work so I can unlink individual commands from specific targets.

  **Acceptance Criteria:**
  - Add --target flag to unlinkCommandCmd in cmd/root.go
  - Pass targetFilter parameter to handleUnlink function
  - Update handler to use NewComponentLinkerWithFilter

- [x] Story-007: As a developer, I want `unlink skills --target <target>` to work so I can unlink all skills from a specific target.

  **Acceptance Criteria:**
  - Add --target flag to unlinkSkillsCmd in cmd/root.go
  - Pass targetFilter parameter to handleUnlinkType function
  - Update handler to use NewComponentLinkerWithFilter

- [x] Story-008: As a developer, I want `unlink agents --target <target>` to work so I can unlink all agents from a specific target.

  **Acceptance Criteria:**
  - Add --target flag to unlinkAgentsCmd in cmd/root.go
  - Pass targetFilter parameter to handleUnlinkType function
  - Update handler to use NewComponentLinkerWithFilter

- [x] Story-009: As a developer, I want `unlink commands --target <target>` to work so I can unlink all commands from a specific target.

  **Acceptance Criteria:**
  - Add --target flag to unlinkCommandsCmd in cmd/root.go
  - Pass targetFilter parameter to handleUnlinkType function
  - Update handler to use NewComponentLinkerWithFilter

- [x] Story-010: As a developer, I want `unlink all --target <target>` to work so I can unlink all components from a specific target.

  **Acceptance Criteria:**
  - Add --target flag to unlinkAllCmd in cmd/root.go
  - Pass targetFilter parameter to handleUnlinkAll function
  - Update handler to use NewComponentLinkerWithFilter

- [x] Story-011: As a developer, I want the handler function signatures updated to accept targetFilter so the flag values can be passed through.

  **Acceptance Criteria:**
  - Update handleUnlink signature in main.go to accept targetFilter parameter (line 1363)
  - Update handleUnlinkAll signature in main.go to accept targetFilter parameter (line 1364)
  - Update handleUnlinkType signature in main.go to accept targetFilter parameter (line 1365)

- [x] Story-012: As a developer, I want SetHandlers to work with the new signatures so the unlink commands integrate properly.

  **Acceptance Criteria:**
  - Update SetHandlers function in cmd/root.go to accept targetFilter in unlink function signature
  - Update SetHandlers function to accept targetFilter in unlinkAll function signature
  - Update SetHandlers function to accept targetFilter in unlinkType function signature

- [ ] Story-013: As a user, I want updated help text for all unlink commands so I know how to use the new --target flag.

  **Acceptance Criteria:**
  - Add --target flag documentation to unlinkSkillCmd help text with examples
  - Add --target flag documentation to unlinkAgentCmd help text with examples
  - Add --target flag documentation to unlinkCommandCmd help text with examples
  - Add --target flag documentation to plural unlink commands help text
  - Add --target flag documentation to unlink all command help text

- [ ] Story-014: As a developer, I want comprehensive test coverage for unlink by target so the feature is reliable.

  **Acceptance Criteria:**
  - Test unlinking from specific target leaves other targets untouched
  - Test invalid target name shows helpful error
  - Test default behavior (no flag) unlinks from all targets
  - Test --target all works correctly
  - Test unlinking from custom targets works
  - Test each unlink command variant with --target flag

## Functional Requirements

### CLI Flag Requirements

- FR-1: All unlink commands (skill/agent/command, skills/agents/commands, all) MUST accept a `--target` or `-t` flag
- FR-2: The --target flag MUST accept the same values as the link command: "opencode", "claudecode", "all", or any custom target name
- FR-3: When --target flag is omitted, unlink commands MUST operate on all detected targets (backward compatibility)
- FR-4: The --target flag MUST be optional for all unlink commands

### Validation Requirements

- FR-5: The system MUST validate that the specified target exists before attempting unlink operations
- FR-6: When an invalid target is specified, the system MUST display a clear error message listing available targets
- FR-7: Invalid target specification MUST result in a non-zero exit code
- FR-8: Custom target validation MUST check both built-in targets (opencode, claudecode) and user-configured custom targets

### Implementation Requirements

- FR-9: All singular unlink commands (skill/agent/command) MUST use NewComponentLinkerWithFilter when --target flag is provided
- FR-10: All plural unlink commands (skills/agents/commands) MUST use NewComponentLinkerWithFilter when --target flag is provided
- FR-11: The "unlink all" command MUST use NewComponentLinkerWithFilter when --target flag is provided
- FR-12: Handler function signatures (handleUnlink, handleUnlinkAll, handleUnlinkType) MUST be updated to accept targetFilter parameter
- FR-13: SetHandlers function in cmd/root.go MUST be updated to pass targetFilter to all unlink handlers

### Output Requirements

- FR-14: Success messages MUST include the target name(s) affected by the unlink operation
- FR-15: When a component is not found in the specified target, a skip message MUST be displayed
- FR-16: Output format MUST remain consistent with current unlink command output (maintain user familiarity)
- FR-17: When operating on multiple targets, each target's result MUST be reported separately

### Documentation Requirements

- FR-18: All unlink command help text MUST document the --target flag with usage examples
- FR-19: Help text MUST show examples of unlinking from specific targets
- FR-20: Help text MUST clarify default behavior when flag is omitted
- FR-21: Error messages MUST guide users toward correct usage when invalid targets are specified

## Technical Implementation Details

### Files to Modify

1. **cmd/root.go** (lines 679-876)
   - Add --target flag to all unlink commands: unlinkSkillCmd, unlinkAgentCmd, unlinkCommandCmd, unlinkSkillsCmd, unlinkAgentsCmd, unlinkCommandsCmd, unlinkAllCmd
   - Update Run functions to read targetFilter from flag
   - Pass targetFilter to handler functions
   - Update help text with flag documentation

2. **main.go** (lines 1363-1365)
   - Update handleUnlink signature: `func(componentType, componentName, targetFilter string)`
   - Update handleUnlinkAll signature: `func(targetFilter string)`
   - Update handleUnlinkType signature: `func(componentType, targetFilter string)`
   - Replace NewComponentLinker() calls with NewComponentLinkerWithFilter(targetFilter)

3. **cmd/root.go SetHandlers** (lines 1350-1407)
   - Update unlink function signature: `unlink func(componentType, componentName, targetFilter string)`
   - Update unlinkAll function signature: `unlinkAll func(targetFilter string)`
   - Update unlinkType function signature: `unlinkType func(componentType, targetFilter string)`

### Key Insight

The actual `UnlinkComponent()`, `UnlinkComponentsByType()`, and `UnlinkAllComponents()` methods in `internal/linker/linker.go` **do NOT need changes**. They already operate on whatever targets are configured in the ComponentLinker at creation time. Target filtering happens at the ComponentLinker instantiation level via `NewComponentLinkerWithFilter()`, which already exists and is used by link commands.

### Architecture Pattern

This implementation follows the existing pattern used by link commands:
1. CLI layer captures --target flag value
2. Handler layer passes targetFilter to NewComponentLinkerWithFilter()
3. ComponentLinker is created with filtered target list
4. Unlink methods operate on filtered targets automatically

This design provides clean separation of concerns and requires minimal code changes.

## Example Usage

### Before (Current Behavior)
```bash
# Always unlinks from ALL targets
agent-smith unlink skill my-skill

# No way to specify which target
agent-smith unlink agent my-agent
```

### After (Enhanced Behavior)
```bash
# Unlink from all targets (default, backward compatible)
agent-smith unlink skill my-skill

# Unlink from OpenCode only
agent-smith unlink skill my-skill --target opencode

# Unlink from specific custom target
agent-smith unlink agent my-agent --target my-custom-target

# Unlink all skills from ClaudeCode only
agent-smith unlink skills --target claudecode

# Unlink everything from all targets (explicit)
agent-smith unlink all --target all
```

## Success Criteria

- [ ] All unlink commands accept --target flag
- [ ] Default behavior (no flag) maintains backward compatibility
- [ ] Invalid targets show helpful error messages with available target list
- [ ] Unlinking from specific target leaves other targets untouched
- [ ] Output messages clearly indicate which target(s) were affected
- [ ] Help text documents the --target flag with examples
- [ ] All tests pass including new test cases for target filtering
- [ ] Feature parity achieved with link command behavior

## Non-Goals (Out of Scope)

- No changes to the link command (already supports --target)
- No changes to unlink logic in internal/linker/linker.go (already supports filtered targets)
- No changes to target detection or configuration system
- No new target types or target management features
- No changes to the status command output
- No interactive target selection prompts
- No batch operations across multiple targets in a single command
- No configuration file for default target preferences

## Implementation Notes

### Backward Compatibility

- Default behavior (no --target flag) continues to unlink from all targets
- Existing scripts and workflows continue to work without modification
- Exit codes remain consistent with current behavior

### Code Reuse

- Leverages existing `NewComponentLinkerWithFilter()` function from main.go
- Uses existing target validation in `config.NewTarget()` and `config.DetectAllTargets()`
- No new validation logic required

### Testing Strategy

- Add test cases for each unlink command with --target flag
- Verify target filtering works correctly (unlinks from specified target only)
- Verify default behavior (no flag) still unlinks from all targets
- Test invalid target error handling
- Test edge cases: custom targets, "all" keyword, non-existent targets

## Dependencies

- No external dependencies
- No new packages required
- Uses existing target system in pkg/config/
- Uses existing ComponentLinker filtering in main.go

## Timeline Estimate

- **Parallel Group 0** (Foundation): Target validation and core handler updates (~2-3 hours)
- **Parallel Group 1** (Independent Features): CLI flag additions for all commands (~2-3 hours)
- **Parallel Group 2** (Integration): Signature updates and documentation (~1-2 hours)

**Total Estimated Time:** 5-8 hours for full implementation and testing
