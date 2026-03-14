# PRD: Add Copilot Target Support for Materialize and Link Commands

**Created**: 2026-02-03 15:30 UTC

---

## Introduction

Add GitHub Copilot as a new supported target for both `materialize` and `link` commands. This enables users to materialize components to `.github/` directory for project-local use and link components to `~/.copilot/` for global Copilot integration.

**Problem Statement**: Currently, agent-smith only supports OpenCode (`.opencode/`) and Claude Code (`.claude/`) targets. Users working with GitHub Copilot need a way to materialize and link components for Copilot integration.

**Solution**: Extend the existing target architecture to support `copilot` as a third target, following the same patterns used for OpenCode and Claude Code targets.

---

## Goals

- Add `copilot` as a valid target option alongside `opencode` and `claudecode`
- Support materialization to `.github/` directory (project-local)
- Support linking to `~/.copilot/` directory (global)
- Maintain backward compatibility with existing OpenCode and Claude Code functionality
- Follow existing architectural patterns (Target interface implementation)
- Include copilot in `--target all` operations
- Provide complete test coverage for new functionality

---

## User Stories

- [ ] **Story-001**: As a developer using GitHub Copilot, I want to materialize components to `.github/` so that they're available in my project.

  **Acceptance Criteria:**
  - Command `agent-smith materialize skill <name> --target copilot` copies skill to `.github/skills/`
  - Command `agent-smith materialize agent <name> --target copilot` copies agent to `.github/agents/`
  - Command `agent-smith materialize command <name> --target copilot` copies command to `.github/commands/`
  - Directory structure (skills/, agents/, commands/) is created under `.github/`
  - Provenance metadata is tracked in `.github/.materializations.json`
  - `--target all` includes copilot alongside opencode and claudecode
  
  **Testing Criteria:**
  **Unit Tests:**
  - `GetTargetDirectory()` returns `.github/` for copilot target
  - Target validation accepts "copilot" as valid
  - All target list includes copilot
  
  **Integration Tests:**
  - Materialize skill to copilot target end-to-end
  - Materialize agent to copilot target end-to-end
  - Materialize command to copilot target end-to-end
  - Materialize all includes copilot target
  - Metadata file creation in `.github/.materializations.json`
  - Directory structure creation verification
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

- [ ] **Story-002**: As a developer, I want to link components to `~/.copilot/` so that they're available globally for Copilot.

  **Acceptance Criteria:**
  - Command `agent-smith link skill <name> --target copilot` creates symlink in `~/.copilot/skills/`
  - Command `agent-smith link agent <name> --target copilot` creates symlink in `~/.copilot/agents/`
  - Command `agent-smith link command <name> --target copilot` creates symlink in `~/.copilot/commands/`
  - `~/.copilot/` directory is created if it doesn't exist
  - Link status commands show copilot links correctly
  - `--target all` includes copilot in linking operations
  
  **Testing Criteria:**
  **Unit Tests:**
  - CopilotTarget implements Target interface correctly
  - GetBaseDir() returns `~/.copilot/` path
  - GetSkillsDir(), GetAgentsDir(), GetCommandsDir() return correct subdirectories
  - GetAllTargets() includes copilot target
  
  **Integration Tests:**
  - Link skill to copilot target end-to-end
  - Link all components to copilot target
  - Link status shows copilot links
  - Unlink from copilot target works correctly
  - Directory creation for `~/.copilot/` and subdirectories
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

- [ ] **Story-003**: As a developer, I want copilot to be included in all multi-target operations so that I don't have to specify it separately.

  **Acceptance Criteria:**
  - `agent-smith materialize skill <name> --target all` materializes to opencode, claudecode, AND copilot
  - `agent-smith link skill <name> --target all` links to opencode, claudecode, AND copilot
  - `agent-smith materialize all --target all` materializes all components to all three targets
  - `agent-smith link all --target all` links all components to all three targets
  - Status and list commands show copilot components alongside others
  
  **Testing Criteria:**
  **Unit Tests:**
  - Target filtering logic includes copilot in "all" results
  - GetAllTargets() returns three targets
  
  **Integration Tests:**
  - Materialize with --target all creates three copies
  - Link with --target all creates three sets of links
  - Status commands show all three targets
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

- [ ] **Story-004**: As a developer, I want environment variable support for copilot target so that I can set a default.

  **Acceptance Criteria:**
  - `AGENT_SMITH_TARGET=copilot` sets copilot as default target
  - Commands work without --target flag when env var is set
  - Env var works for both materialize and link commands
  
  **Testing Criteria:**
  **Unit Tests:**
  - Config target detection reads AGENT_SMITH_TARGET correctly
  
  **Integration Tests:**
  - Test with AGENT_SMITH_TARGET=copilot environment variable
  - Verify materialize uses copilot as default
  - Verify link uses copilot as default
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

- [ ] **Story-005**: As a developer, I want status and info commands to work with copilot target so that I can inspect materialized components.

  **Acceptance Criteria:**
  - `agent-smith materialize status --target copilot` shows copilot status
  - `agent-smith materialize list` includes copilot components
  - `agent-smith materialize info skill <name> --target copilot` shows copilot metadata
  - `agent-smith link status` shows copilot links
  - Update commands work with copilot target
  
  **Testing Criteria:**
  **Unit Tests:**
  - Status filtering handles copilot target
  - Metadata loading from `.github/.materializations.json`
  
  **Integration Tests:**
  - Status command with copilot target
  - List command includes copilot
  - Info command with copilot target
  - Update command with copilot target
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

---

## Technical Design

### Architecture

The implementation follows the existing Target interface pattern:

```
Target Interface (pkg/config/target.go)
    ├── OpencodeTarget (existing)
    ├── ClaudeCodeTarget (existing)
    └── CopilotTarget (NEW)
```

### Component Structure

**File Changes Required:**

1. **New Files:**
   - `pkg/config/copilot_target.go` - CopilotTarget implementation
   - `pkg/config/copilot_target_test.go` - Unit tests for CopilotTarget
   - `tests/integration/materialize_copilot_test.go` - Integration tests for materialize
   - `tests/integration/link_copilot_test.go` - Integration tests for link

2. **Modified Files:**
   - `pkg/paths/paths.go` - Add CopilotDir constant and GetCopilotDir()
   - `pkg/project/detection.go` - Add .github to ProjectMarkers, handle copilot in GetTargetDirectory()
   - `pkg/config/target_manager.go` - Add copilot case to NewTarget() and GetAllTargets()
   - `pkg/services/materialize/service.go` - Include copilot in target lists
   - `cmd/root.go` - Update help text to mention copilot
   - `pkg/errors/helpers.go` - Update error messages to include copilot

### Target Directory Mapping

```
Target Name    | Materialize Directory | Link Directory
---------------|----------------------|----------------
opencode       | .opencode/           | ~/.config/opencode/
claudecode     | .claude/             | ~/.claude/
copilot        | .github/             | ~/.copilot/
```

### Metadata Storage

Each target maintains its own metadata file:
- `.opencode/.materializations.json`
- `.claude/.materializations.json`
- `.github/.materializations.json` (NEW)

### Directory Structure

All targets use the same subdirectory structure:
```
.github/
├── .materializations.json
├── skills/
├── agents/
└── commands/
```

---

## Implementation Plan

### Phase 1: Core Infrastructure
- [ ] Add `CopilotDir` constant to `pkg/paths/paths.go`
- [ ] Add `GetCopilotDir()` function to `pkg/paths/paths.go`
- [ ] Update `ProjectMarkers` in `pkg/project/detection.go` to include `.github`
- [ ] Update `GetTargetDirectory()` in `pkg/project/detection.go` to handle copilot case

### Phase 2: Target Implementation
- [ ] Create `pkg/config/copilot_target.go` implementing Target interface
- [ ] Create `pkg/config/copilot_target_test.go` with unit tests
- [ ] Update `NewTarget()` in `pkg/config/target_manager.go` to handle copilot case
- [ ] Update `GetAllTargets()` in `pkg/config/target_manager.go` to include copilot

### Phase 3: Materialize Command Support
- [ ] Update target list in `MaterializeComponent()` to include copilot
- [ ] Update target validation to accept copilot
- [ ] Update error messages to mention copilot
- [ ] Test materialize commands with copilot target

### Phase 4: Link Command Support
- [ ] Verify linker works with copilot target (should work automatically via Target interface)
- [ ] Test link commands with copilot target
- [ ] Test link status with copilot

### Phase 5: Help Text and Documentation
- [ ] Update all `--target` flag help text in `cmd/root.go`
- [ ] Update error messages in `pkg/errors/helpers.go`
- [ ] Update examples in command help text

### Phase 6: Testing
- [ ] Create `tests/integration/materialize_copilot_test.go`
- [ ] Create `tests/integration/link_copilot_test.go`
- [ ] Update `tests/integration/materialize_env_target_test.go` for copilot
- [ ] Run full test suite
- [ ] Manual testing of all commands

---

## Testing Strategy

### Unit Tests

**pkg/paths/paths.go:**
- Test `GetCopilotDir()` returns correct path
- Test tilde expansion works

**pkg/project/detection.go:**
- Test `GetTargetDirectory("copilot")` returns `.github/`
- Test `.github` is recognized as project marker

**pkg/config/copilot_target.go:**
- Test `GetBaseDir()` returns `~/.copilot/`
- Test `GetSkillsDir()` returns `~/.copilot/skills/`
- Test `GetAgentsDir()` returns `~/.copilot/agents/`
- Test `GetCommandsDir()` returns `~/.copilot/commands/`
- Test `GetName()` returns "copilot"

**pkg/config/target_manager.go:**
- Test `NewTarget("copilot")` creates CopilotTarget
- Test `GetAllTargets()` includes copilot
- Test target validation accepts copilot

### Integration Tests

**Materialize Commands:**
- Test `materialize skill <name> --target copilot`
- Test `materialize agent <name> --target copilot`
- Test `materialize command <name> --target copilot`
- Test `materialize all --target copilot`
- Test `materialize skill <name> --target all` (includes copilot)
- Test `.github/` directory structure creation
- Test `.github/.materializations.json` creation and content
- Test `materialize status --target copilot`
- Test `materialize list` includes copilot components
- Test `materialize info skill <name> --target copilot`

**Link Commands:**
- Test `link skill <name> --target copilot`
- Test `link agent <name> --target copilot`
- Test `link command <name> --target copilot`
- Test `link all --target copilot`
- Test `link skill <name> --target all` (includes copilot)
- Test `~/.copilot/` directory creation
- Test symlink creation in `~/.copilot/`
- Test `link status` shows copilot
- Test `unlink skill <name> --target copilot`

**Environment Variable:**
- Test `AGENT_SMITH_TARGET=copilot` sets default
- Test materialize without --target uses env var
- Test link without --target uses env var

---

## Acceptance Criteria

- [ ] All user stories completed and tested
- [ ] All unit tests passing
- [ ] All integration tests passing
- [ ] Help text updated to mention copilot
- [ ] Error messages include copilot as valid option
- [ ] `--target copilot` works for all materialize commands
- [ ] `--target copilot` works for all link commands
- [ ] `--target all` includes copilot for all operations
- [ ] Status and info commands work with copilot
- [ ] `AGENT_SMITH_TARGET=copilot` environment variable works
- [ ] No regression in existing opencode/claudecode functionality
- [ ] Manual testing confirms all workflows work end-to-end

---

## Dependencies

- No external dependencies
- Uses existing Target interface pattern
- Compatible with current materialize and link architecture

---

## Risks and Mitigations

**Risk**: Changes to target system could break existing functionality
**Mitigation**: Comprehensive test coverage, including existing opencode/claudecode tests

**Risk**: Users might be confused about which target to use
**Mitigation**: Clear help text and examples showing all three targets

**Risk**: `.github/` directory might conflict with GitHub Actions or other GitHub tooling
**Mitigation**: Subdirectory structure keeps agent-smith components isolated; `.github/` is a valid location for custom tooling

---

## Timeline Estimate

- **Phase 1 (Infrastructure)**: 1 hour
- **Phase 2 (Target Implementation)**: 2 hours
- **Phase 3 (Materialize Support)**: 1 hour
- **Phase 4 (Link Support)**: 1 hour
- **Phase 5 (Help Text)**: 1 hour
- **Phase 6 (Testing)**: 2 hours

**Total Estimate**: 8 hours

---

## Success Metrics

- All test suites passing (unit + integration)
- Zero regressions in existing functionality
- Copilot target works for all materialize operations
- Copilot target works for all link operations
- Help text accurately reflects new functionality
- Manual workflows complete successfully

---

## Future Enhancements

- Add support for other AI coding assistants as targets
- Add target-specific configuration options
- Add migration helpers for moving components between targets
- Add bulk operations for managing multiple targets simultaneously
