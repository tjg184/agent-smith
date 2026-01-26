# PRD: Agent Smith CLI Command Restructuring

## Overview

Restructure the Agent Smith CLI from a flat command structure to a hierarchical, action-based command structure for improved usability and maintainability.

## Problem Statement

The current CLI has a flat command structure with commands like `add-skill`, `add-agent`, `add-command`, `add-all`, `auto-link`, `list-links`, `link-status`, etc. This creates:
- A cluttered root help menu with 15+ commands
- Inconsistent naming patterns
- Poor command discoverability
- Difficult mental model for users

## Goals

1. Group related commands under logical parent commands
2. Reduce root-level command count from 15+ to 5-6
3. Maintain backward compatibility where possible
4. Improve consistency in command naming and structure

## Proposed Structure

### Before (Current)
```
agent-smith add-skill <repo> <name>
agent-smith add-agent <repo> <name>
agent-smith add-command <repo> <name>
agent-smith add-all <repo>
agent-smith link <type|all> [name] [--target]
agent-smith auto-link
agent-smith list-links
agent-smith link-status
agent-smith unlink <type|all> [name] [--force]
agent-smith update <type|all> [name]
agent-smith npx <target> [args...]
agent-smith run <target> [args...]
```

### After (Proposed)
```
agent-smith add skill <repo> <name>
agent-smith add agent <repo> <name>
agent-smith add command <repo> <name>
agent-smith add all <repo>

agent-smith link skill <name> [--target]
agent-smith link agent <name> [--target]
agent-smith link command <name> [--target]
agent-smith link all [--target]
agent-smith link auto
agent-smith link list
agent-smith link status

agent-smith unlink skill <name>
agent-smith unlink agent <name>
agent-smith unlink command <name>
agent-smith unlink all [--force]

agent-smith update skill <name>
agent-smith update agent <name>
agent-smith update command <name>
agent-smith update all

agent-smith run <target> [args...]
```

## Technical Requirements

### 1. Command Group: `add`

**Parent Command**: `agent-smith add`
- Description: "Download and install components from git repositories"

**Subcommands**:
- `skill <repo> <name>` - Download a skill
- `agent <repo> <name>` - Download an agent
- `command <repo> <name>` - Download a command
- `all <repo>` - Download all components from a repository

**Implementation**:
- Create parent `addCmd` with no Run function
- Convert existing `add-skill`, `add-agent`, `add-command`, `add-all` to subcommands
- Preserve all existing flags, help text, and functionality
- Map to existing handlers: `handleAddSkill`, `handleAddAgent`, `handleAddCommand`, `handleAddAll`

### 2. Command Group: `link`

**Parent Command**: `agent-smith link`
- Description: "Link components to detected targets (OpenCode, Claude Code, etc.)"

**Subcommands**:
- `skill <name> [--target]` - Link a specific skill
- `agent <name> [--target]` - Link a specific agent
- `command <name> [--target]` - Link a specific command
- `all [--target]` - Link all components
- `auto` - Auto-detect and link components from current repository
- `list` - List all linked components
- `status` - Show link status matrix

**Flags**:
- `--target, -t <target>` - Specify target (opencode, claudecode, all)

**Implementation**:
- Refactor existing `linkCmd` to be parent command
- Create subcommands for each component type
- Migrate `auto-link` → `link auto`
- Migrate `list-links` → `link list`
- Migrate `link-status` → `link status`
- Map to existing handlers: `handleLink`, `handleLinkAll`, `handleLinkType`, `handleAutoLink`, `handleListLinks`, `handleLinkStatus`

### 3. Command Group: `unlink`

**Parent Command**: `agent-smith unlink`
- Description: "Remove linked components from targets"

**Subcommands**:
- `skill <name>` - Unlink a specific skill
- `agent <name>` - Unlink a specific agent
- `command <name>` - Unlink a specific command
- `all [--force]` - Unlink all components

**Flags**:
- `--force, -f` - Skip confirmation prompt

**Implementation**:
- Refactor existing `unlinkCmd` to be parent command
- Create subcommands for each component type
- Preserve safety confirmations for bulk operations
- Map to existing handlers: `handleUnlink`, `handleUnlinkAll`, `handleUnlinkType`

### 4. Command Group: `update`

**Parent Command**: `agent-smith update`
- Description: "Check and update components from their source repositories"

**Subcommands**:
- `skill <name>` - Update a specific skill
- `agent <name>` - Update a specific agent
- `command <name>` - Update a specific command
- `all` - Update all components

**Implementation**:
- Refactor existing `updateCmd` to be parent command
- Create subcommands for each component type
- Map to existing handlers: `handleUpdate`, `handleUpdateAll`

### 5. Standalone Command: `run`

**Command**: `agent-smith run <target> [args...]`
- Description: "Execute a component without installing (npx-like)"
- Keep as-is, remove `npx` alias for simplicity

**Implementation**:
- Keep existing `run` command
- Remove `npx` command to reduce confusion
- Map to existing handler: `handleRun`

## File Changes Required

### `/Users/tgaines/dev/git/agent-smith/cmd/root.go`

**Changes**:
1. Remove flat commands: `add-skill`, `add-agent`, `add-command`, `add-all`, `auto-link`, `list-links`, `link-status`, `npx`
2. Create parent command groups: `addCmd`, refactor `linkCmd`, `unlinkCmd`, `updateCmd`
3. Add subcommands to each parent command group
4. Preserve all existing handlers and flags
5. Update help text to reflect new structure

**No changes needed**:
- `/Users/tgaines/dev/git/agent-smith/main.go` - All handlers remain unchanged
- Handler function signatures remain the same
- No changes to internal packages required

## Validation Criteria

### Functional Testing
- [x] `agent-smith add skill <repo> <name>` downloads skill correctly
- [x] `agent-smith add agent <repo> <name>` downloads agent correctly
- [x] `agent-smith add command <repo> <name>` downloads command correctly
- [ ] `agent-smith add all <repo>` downloads all components correctly
- [ ] `agent-smith link skill <name>` links skill to targets
- [ ] `agent-smith link agent <name>` links agent to targets
- [ ] `agent-smith link command <name>` links command to targets
- [ ] `agent-smith link all` links all components
- [ ] `agent-smith link auto` auto-detects and links components
- [ ] `agent-smith link list` lists linked components
- [ ] `agent-smith link status` shows link status matrix
- [ ] `agent-smith link skill <name> --target opencode` respects target flag
- [ ] `agent-smith unlink skill <name>` unlinks skill
- [ ] `agent-smith unlink all` prompts for confirmation
- [ ] `agent-smith unlink all --force` skips confirmation
- [ ] `agent-smith update skill <name>` updates skill
- [ ] `agent-smith update all` updates all components
- [ ] `agent-smith run <target> [args...]` executes component
- [ ] `agent-smith --help` shows clean root menu with 5-6 commands
- [ ] `agent-smith add --help` shows add subcommands
- [ ] `agent-smith link --help` shows link subcommands

### Build Testing
- [ ] `go build` completes without errors
- [ ] Binary runs successfully
- [ ] All help text renders correctly

## Migration Strategy

### Backward Compatibility
- Old commands are removed in favor of new structure
- Users will need to update scripts/workflows
- Consider adding deprecated command aliases in future if needed

### Documentation Updates
- Update README with new command structure
- Update examples in help text
- Update any tutorials or guides

## Non-Goals

- Changing handler implementations in `main.go`
- Modifying internal package logic
- Adding new features beyond restructuring
- Backward compatibility aliases (can be added later if needed)

## Success Metrics

1. Root help menu reduced from 15+ commands to ~5-6 command groups
2. All existing functionality preserved
3. Consistent command structure across all operations
4. Improved command discoverability through grouping
5. Clean build with no errors

## Timeline

**Phase 1**: Refactor `cmd/root.go` (1-2 hours)
- Create parent command groups
- Convert flat commands to subcommands
- Update help text

**Phase 2**: Testing (30 minutes)
- Build and run tests
- Manual testing of command groups
- Verify help text

**Phase 3**: Documentation (15 minutes)
- Update help text if needed
- Verify all examples work

## Open Questions

1. Should we keep `npx` as an alias for `run`? **Decision: Remove for simplicity**
2. Should we support both `link skill` and `link skills`? **Decision: Use singular form**
3. Should `update` support type-based bulk updates like `update skills`? **Decision: Yes, via subcommands**
