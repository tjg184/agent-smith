# PRD: Custom Directory Linking

## Introduction

Add support for linking agent-smith components (skills, agents, commands) to custom user-defined directories beyond the built-in OpenCode and Claude Code targets. This enables users to integrate agent-smith with any editor, tool, or custom workflow by defining their own target directories through a configuration file, while maintaining backward compatibility with existing built-in targets.

## Goals

- Allow users to define custom target directories for any editor/tool (Cursor, VS Code, custom tools, etc.)
- Provide CLI commands to manage custom targets (`target add/remove/list`)
- Maintain full backward compatibility with existing OpenCode and Claude Code built-in targets
- Keep implementation minimal and maintainable with config file + basic CLI commands
- Validate custom target configurations to prevent linking errors

## User Stories

- [ ] Story-001: As a developer, I want to define a custom target directory so that I can link agent-smith components to editors/tools beyond OpenCode and Claude Code.

  **Acceptance Criteria:**
  - Config file at `~/.agents/config.json` supports custom target definitions
  - Each custom target has a unique name, base directory, and component subdirectories
  - Config is validated on load with clear error messages for invalid entries
  - Custom targets coexist with built-in OpenCode and Claude Code targets
  
  **Testing Criteria:**
  **Unit Tests:**
  - Config file parsing and validation logic tests
  - Custom target struct validation tests
  - Path expansion and normalization tests
  
  **Integration Tests:**
  - Config file loading from disk tests
  - Multiple custom targets in config file tests
  - Invalid config file handling tests

- [ ] Story-002: As a developer, I want to use `agent-smith target add <name> <path>` to register a new custom target so that I don't have to manually edit the config file.

  **Acceptance Criteria:**
  - Command accepts target name and base directory path
  - Validates target name is unique and path is valid/accessible
  - Creates config file if it doesn't exist
  - Adds new target to config file in correct format
  - Prompts for component subdirectory names (skills/, agents/, commands/) with sensible defaults
  - Reports success with confirmation message
  
  **Testing Criteria:**
  **Unit Tests:**
  - Command argument parsing and validation tests
  - Target name uniqueness validation tests
  - Path validation logic tests
  
  **Integration Tests:**
  - Adding first custom target (config file creation) tests
  - Adding additional custom targets tests
  - Duplicate target name rejection tests
  - Invalid path rejection tests
  
  **Component Browser Tests:**
  - Command output format verification
  - Error message display for various failure scenarios
  - Success confirmation message verification

- [ ] Story-003: As a developer, I want to use `agent-smith target remove <name>` to unregister a custom target so that I can clean up targets I no longer use.

  **Acceptance Criteria:**
  - Command accepts target name to remove
  - Validates target exists and is a custom target (not built-in)
  - Removes target from config file
  - Prevents removal of built-in OpenCode/Claude Code targets with clear error message
  - Optionally unlinks components from removed target (with confirmation prompt)
  - Reports success with confirmation message
  
  **Testing Criteria:**
  **Unit Tests:**
  - Target existence validation tests
  - Built-in target protection tests
  - Config file update logic tests
  
  **Integration Tests:**
  - Custom target removal tests
  - Built-in target removal prevention tests
  - Config file state after removal tests
  
  **Component Browser Tests:**
  - Confirmation prompt display and handling
  - Error message for built-in target removal attempt
  - Success message verification

- [ ] Story-004: As a developer, I want to use `agent-smith target list` to see all available targets so that I can understand which targets are configured.

  **Acceptance Criteria:**
  - Lists all built-in targets (OpenCode, Claude Code) with [built-in] indicator
  - Lists all custom targets from config file with [custom] indicator
  - Shows target base directory path for each target
  - Indicates if target directory exists or is missing
  - Displays active profile targets separately if profile is active
  - Clear formatting with columns for name, type, path, and status
  
  **Testing Criteria:**
  **Unit Tests:**
  - Target listing logic tests
  - Built-in vs custom target differentiation tests
  - Directory existence check tests
  
  **Integration Tests:**
  - Listing with no custom targets tests
  - Listing with multiple custom targets tests
  - Listing with active profile tests
  
  **Component Browser Tests:**
  - Table formatting verification
  - Visual distinction between built-in and custom targets
  - Status indicator accuracy tests

- [ ] Story-005: As a developer, I want custom targets to work seamlessly with existing link commands so that I can link components to custom targets using the same workflow as built-in targets.

  **Acceptance Criteria:**
  - `agent-smith link skill <name> --target <custom-target>` works for custom targets
  - `agent-smith link all` includes custom targets in detection
  - `agent-smith link status` shows link status for custom targets
  - `agent-smith link list` includes components linked to custom targets
  - Custom targets appear in target selection prompts
  - Error messages clearly distinguish between built-in and custom targets
  
  **Testing Criteria:**
  **Unit Tests:**
  - Custom target filtering logic tests
  - Target detection including custom targets tests
  - Link status tracking for custom targets tests
  
  **Integration Tests:**
  - Linking to custom target via --target flag tests
  - Auto-detection and linking to all targets tests
  - Link status matrix with custom targets tests
  - Unlinking from custom targets tests
  
  **Component Browser Tests:**
  - Link command output with custom targets
  - Status display with mixed built-in and custom targets
  - Error messages for non-existent custom targets

- [ ] Story-006: As a developer, I want the config file to have a clear, documented structure so that I can understand and modify it if needed.

  **Acceptance Criteria:**
  - Config file uses standard JSON format with clear schema
  - Includes version field for future compatibility
  - Custom targets defined in array with required fields (name, baseDir, skillsDir, agentsDir, commandsDir)
  - Supports both absolute paths and tilde expansion (~/)
  - Comments in documentation explain each field's purpose
  - Example config file provided in documentation
  
  **Testing Criteria:**
  **Unit Tests:**
  - JSON schema validation tests
  - Tilde path expansion tests
  - Relative vs absolute path handling tests
  - Missing field detection tests
  
  **Integration Tests:**
  - Loading valid config files with various path formats
  - Parsing config with multiple custom targets
  - Handling malformed JSON gracefully
  
  **Component Browser Tests:**
  - Documentation accuracy verification
  - Example config file validity tests

## Functional Requirements

- FR-1: The system shall support a configuration file at `~/.agents/config.json` that defines custom targets
- FR-2: The config file shall use JSON format with version field and custom targets array
- FR-3: Each custom target shall have required fields: name, baseDir, skillsDir, agentsDir, commandsDir
- FR-4: The system shall support tilde (~/) path expansion and both relative and absolute paths
- FR-5: The system shall validate the config file on load and report clear error messages for invalid configurations
- FR-6: The `agent-smith target add <name> <path>` command shall register a new custom target
- FR-7: The `agent-smith target remove <name>` command shall unregister a custom target (custom only, not built-in)
- FR-8: The `agent-smith target list` command shall display all built-in and custom targets with their paths and status
- FR-9: The link command shall detect and work with custom targets using the `--target` flag
- FR-10: The `link all`, `link status`, and `link list` commands shall include custom targets
- FR-11: Built-in OpenCode and Claude Code targets shall remain auto-detected and cannot be removed
- FR-12: Custom targets shall implement the existing Target interface for seamless integration
- FR-13: The system shall prevent duplicate target names across built-in and custom targets
- FR-14: The system shall validate that custom target paths are accessible before adding them

## Non-Goals (Out of Scope)

- No migration of built-in OpenCode/Claude Code to config file (they remain hardcoded)
- No per-component-type directory customization (all components use subdirectories under baseDir)
- No support for multiple installations of the same built-in target (OpenCode stable/beta)
- No interactive setup wizard or guided configuration
- No target templates or presets for common editors
- No automatic detection of custom targets (users must explicitly add them)
- No validation that target directories are actual editor/tool directories
- No override or customization of built-in target paths via config
- No environment variable configuration for custom targets
- No profile-specific custom targets (custom targets are global)

## Implementation Notes

### Config File Structure

```json
{
  "version": 1,
  "customTargets": [
    {
      "name": "cursor",
      "baseDir": "~/.cursor",
      "skillsDir": "skills",
      "agentsDir": "agents",
      "commandsDir": "commands"
    },
    {
      "name": "vscode",
      "baseDir": "~/.vscode/agent-smith",
      "skillsDir": "skills",
      "agentsDir": "agents",
      "commandsDir": "commands"
    }
  ]
}
```

### Architecture Changes

1. **New File**: `pkg/config/config.go` - Config file loading and validation
2. **New File**: `pkg/config/custom_target.go` - CustomTarget implementation of Target interface
3. **Modified**: `pkg/config/target_manager.go` - Add custom target detection from config
4. **New Commands**: `cmd/target.go` - CLI commands for target management
5. **Modified**: `internal/linker/linker.go` - Include custom targets in detection

### Target Detection Priority

1. Check `AGENT_SMITH_TARGET` environment variable
2. Load custom targets from config file
3. Auto-detect built-in targets (OpenCode, Claude Code)
4. Merge custom and built-in targets for `link all` operations

### Validation Rules

- Target name must be unique (case-insensitive)
- Target name must be alphanumeric with hyphens/underscores only
- Base directory must exist or be creatable
- Subdirectory names must be valid directory names (no slashes)
- Config file must be valid JSON with required fields

---

## Ralphy YAML Format (for parallel execution)

To execute this PRD using Ralphy's autonomous parallel execution system, use the following YAML format:

```yaml
tasks:
  - title: "Story-001: As a developer, I want to define a custom target directory so that I can link agent-smith components to editors/tools beyond OpenCode and Claude Code - Config file at ~/.agents/config.json supports custom target definitions, Each custom target has unique name/base directory/component subdirectories, Config validated on load with clear error messages, Custom targets coexist with built-in targets"
    completed: false
    parallel_group: 0
  - title: "Story-002: As a developer, I want to use 'agent-smith target add <name> <path>' to register a new custom target so that I don't have to manually edit the config file - Command accepts target name and base directory path, Validates uniqueness and path validity, Creates config file if needed, Prompts for subdirectory names with defaults, Reports success"
    completed: false
    parallel_group: 1
  - title: "Story-003: As a developer, I want to use 'agent-smith target remove <name>' to unregister a custom target so that I can clean up targets I no longer use - Validates target exists and is custom, Removes from config file, Prevents removal of built-in targets, Optionally unlinks components with confirmation, Reports success"
    completed: false
    parallel_group: 1
  - title: "Story-004: As a developer, I want to use 'agent-smith target list' to see all available targets so that I can understand which targets are configured - Lists built-in and custom targets with indicators, Shows base directory paths, Indicates directory existence status, Displays active profile targets separately, Clear column formatting"
    completed: false
    parallel_group: 1
  - title: "Story-005: As a developer, I want custom targets to work seamlessly with existing link commands so that I can link components to custom targets using the same workflow as built-in targets - Works with --target flag, Included in 'link all' detection, Shows in link status/list, Appears in target selection prompts, Clear error messages"
    completed: false
    parallel_group: 2
  - title: "Story-006: As a developer, I want the config file to have a clear, documented structure so that I can understand and modify it if needed - Standard JSON format with clear schema, Version field for compatibility, Required fields documented, Supports absolute paths and tilde expansion, Example config in documentation"
    completed: false
    parallel_group: 0
```

**Parallel Group Strategy:**
- **Group 0 (Foundation)**: Config file structure, validation logic, and documentation
- **Group 1 (Independent Features)**: CLI commands (add/remove/list) that depend on config foundation
- **Group 2 (Integration)**: Integration with existing link commands, requires both config and CLI commands to be complete
