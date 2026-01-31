# Agent Smith - Product Snapshot

**Version**: 1.0 (Current State as of January 30, 2026)  
**Status**: Feature-Complete Production System

---

## Executive Summary

Agent Smith is a universal component management system for AI coding environments. It provides a comprehensive CLI tool for downloading, organizing, and linking AI components (skills, agents, and commands) from Git repositories to multiple AI coding assistants like OpenCode, Claude Code, Cursor, and custom tools.

The system implements a two-tier architecture: global canonical storage in `~/.agent-smith/` with automatic symlinking to environment-specific directories, enabling efficient component management across multiple projects and contexts.

---

## Core Capabilities

### 1. Component Installation & Management

**What it does**: Downloads and installs AI components from any Git repository to a centralized storage location.

**Key Features**:
- **Multi-source support**: GitHub shorthand (`owner/repo`), full URLs, GitLab, SSH URLs, local paths
- **Component types**: Skills (SKILL.md), Agents (AGENT.md), Commands (COMMAND.md)
- **Intelligent detection**: Recursive content-based detection supporting any repository structure
- **Change tracking**: GitHub tree SHA-based update detection (only re-downloads when content changes)
- **Bulk operations**: Install all components from a repository with `install all`
- **Custom directories**: Install to project-local directories with `--target-dir` flag

**Storage Architecture**:
```
~/.agent-smith/
├── skills/              # Installed skills
├── agents/              # Installed agents
├── commands/            # Installed commands
├── .skill-lock.json     # Skill metadata & versions
├── .agent-lock.json     # Agent metadata & versions
└── .command-lock.json   # Command metadata & versions
```

**Example Usage**:
```bash
# Install specific component
agent-smith install skill openai/cookbook gpt-skill

# Install all components from repository
agent-smith install all anthropics/skills

# Install to custom directory (project-local, isolated)
agent-smith install all github.com/org/tools --target-dir ./tools
```

---

### 2. Linking System

**What it does**: Creates symlinks from the centralized storage to AI editor/tool directories, making components available to AI assistants.

**Key Features**:
- **Multi-target support**: OpenCode, Claude Code, and unlimited custom targets
- **Profile-aware**: Automatically uses active profile's components when linking
- **Flexible targeting**: Link to specific targets or all targets simultaneously
- **Status tracking**: View link status across all targets in matrix format
- **Safety**: Handles conflicts gracefully, validates before linking

**Built-in Targets**:
- **OpenCode**: `~/.config/opencode/{skills,agents,commands}/`
- **Claude Code**: `~/.claude/{skills,agents,commands}/`
- **Custom**: User-defined via `target add` command

**Example Usage**:
```bash
# Link all components to all detected targets
agent-smith link all

# Link specific component to specific target
agent-smith link skill mcp-builder --target opencode

# View link status matrix
agent-smith link status

# List all linked components
agent-smith link list
```

---

### 3. Profile Management

**What it does**: Allows switching between different sets of components for different contexts (work, personal, project-specific).

**Key Features**:
- **Context switching**: Swap entire toolsets with profile activation
- **Isolated storage**: Each profile has independent component storage
- **Explicit control**: Two-step activation (activate profile, then link components)
- **Component management**: Add/remove components to/from profiles
- **Persistence**: Active profile persists across sessions

**Profile Structure**:
```
~/.agent-smith/
├── profiles/
│   ├── work/
│   │   ├── skills/
│   │   ├── agents/
│   │   ├── commands/
│   │   ├── .skill-lock.json
│   │   ├── .agent-lock.json
│   │   └── .command-lock.json
│   └── personal/
│       └── ...
└── .active-profile       # Current active profile
```

**Example Usage**:
```bash
# Create and activate profile
agent-smith profile create work
agent-smith profile activate work

# Install components to active profile
agent-smith install skill owner/repo skill-name --profile work

# Apply profile to editor
agent-smith link all

# Switch contexts
agent-smith profile activate personal
agent-smith link all
```

---

### 4. Custom Target System

**What it does**: Extends linking capabilities to any editor or tool beyond built-in OpenCode/Claude Code support.

**Key Features**:
- **User-defined targets**: Add any directory as a link target
- **Configuration-based**: Targets stored in `~/.agent-smith/config.json`
- **Subdirectory mapping**: Configure component subdirectory names per target
- **Management commands**: Add, remove, and list custom targets
- **Seamless integration**: Works with all link commands via `--target` flag

**Configuration Format**:
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

**Example Usage**:
```bash
# Add custom target
agent-smith target add cursor ~/.cursor

# List all targets (built-in + custom)
agent-smith target list

# Link to custom target
agent-smith link all --target cursor
```

---

### 5. Uninstall System

**What it does**: Cleanly removes components from the system with automatic unlinking.

**Key Features**:
- **Automatic unlinking**: Components are unlinked from all targets before removal
- **Individual & bulk**: Remove single components or all from a repository
- **Lock file management**: Updates lock files to maintain consistency
- **Safety confirmations**: Bulk operations prompt for confirmation (override with `--force`)
- **Profile support**: Remove from specific profiles or base installation

**Example Usage**:
```bash
# Remove single component (auto-unlinks first)
agent-smith uninstall skill mcp-builder

# Remove all components from repository
agent-smith uninstall all anthropics/skills

# Force removal without confirmation
agent-smith uninstall all owner/repo --force

# Remove from specific profile
agent-smith uninstall skill test-skill --profile work
```

---

### 6. Update System

**What it does**: Checks for and applies updates to installed components using intelligent change detection.

**Key Features**:
- **SHA-based detection**: Uses GitHub tree SHA for precise change detection
- **Selective updates**: Update individual components or all at once
- **Minimal downloads**: Only re-downloads when actual content changes
- **Lock file sync**: Updates metadata after successful updates

**Example Usage**:
```bash
# Update specific component
agent-smith update skills mcp-builder

# Update all components
agent-smith update all
```

---

## Technical Architecture

### Component Detection

Agent Smith uses **recursive content-based detection** to find components in any repository structure:

- **Skills**: Identifies directories containing `SKILL.md` files
- **Agents**: Identifies directories containing `AGENT.md` files  
- **Commands**: Identifies directories containing `COMMAND.md` files

This flexible approach supports:
- Standard layouts (`/skills/`, `/agents/`, `/commands/`)
- Monorepos with mixed component types
- Custom repository structures
- Nested component directories

### Lock File System

Three lock files track component metadata:

**`.skill-lock.json`**:
```json
{
  "version": 3,
  "skills": {
    "component-name": {
      "source": "github.com/owner/repo",
      "sourceUrl": "https://github.com/owner/repo",
      "commit": "abc123...",
      "treeSHA": "def456...",
      "installedAt": "2026-01-30T12:00:00Z"
    }
  }
}
```

Similar structure for `.agent-lock.json` and `.command-lock.json`.

### Symlink Strategy

- **macOS/Linux**: Native symlinks for zero-copy linking
- **Windows**: Junctions as fallback when symlinks unavailable
- **Graceful degradation**: Falls back to directory copy if symlinks fail
- **Relative paths**: Uses relative symlinks for portability

---

## Command Reference

### Install Commands

```bash
# Individual components
agent-smith install skill <repo-url> <skill-name>
agent-smith install agent <repo-url> <agent-name>
agent-smith install command <repo-url> <command-name>

# Bulk installation
agent-smith install all <repo-url>

# Flags
--profile, -p <name>      # Install to specific profile
--target-dir, -t <path>   # Install to custom directory (isolated)
```

### Link Commands

```bash
# Link specific components
agent-smith link skill <name>
agent-smith link agent <name>
agent-smith link command <name>

# Link by type
agent-smith link skills      # All skills
agent-smith link agents      # All agents
agent-smith link commands    # All commands

# Link everything
agent-smith link all

# Status & management
agent-smith link status      # Matrix view of all links
agent-smith link list        # List all linked components

# Flags
--target, -t <name>          # Link to specific target
--all-targets                # Explicitly link to all targets
```

### Unlink Commands

```bash
# Unlink specific components
agent-smith unlink skill <name>
agent-smith unlink agent <name>
agent-smith unlink command <name>

# Unlink by type
agent-smith unlink skills [name]
agent-smith unlink agents [name]
agent-smith unlink commands [name]

# Unlink everything
agent-smith unlink all

# Flags
--target, -t <name>          # Unlink from specific target
--force, -f                  # Skip confirmation prompts
```

### Uninstall Commands

```bash
# Remove individual components
agent-smith uninstall skill <name>
agent-smith uninstall agent <name>
agent-smith uninstall command <name>

# Remove all from repository
agent-smith uninstall all <repo-url>

# Flags
--profile, -p <name>         # Remove from specific profile
--force, -f                  # Skip confirmation prompts
```

### Profile Commands

```bash
# Profile management
agent-smith profile create <name>
agent-smith profile delete <name>
agent-smith profile activate <name>
agent-smith profile deactivate

# Profile inspection
agent-smith profile list
agent-smith profile show <name>

# Component management
agent-smith profile add <type> <profile> <component>
agent-smith profile remove <type> <profile> <component>
```

### Target Commands

```bash
# Manage custom targets
agent-smith target add <name> <path>
agent-smith target remove <name>
agent-smith target list
```

### Update Commands

```bash
# Update components
agent-smith update <type> <name>
agent-smith update all
```

### Status Command

```bash
# Show system status
agent-smith status
```

---

## Common Workflows

### Workflow 1: Basic Setup

```bash
# Install components from repository
agent-smith install all openai/cookbook

# Link to all detected targets
agent-smith link all

# Check what's linked
agent-smith link status
```

### Workflow 2: Multi-Target Setup

```bash
# Add custom targets
agent-smith target add cursor ~/.cursor
agent-smith target add vscode ~/.vscode/agent-smith

# Install components
agent-smith install all anthropics/skills

# Link to all targets (built-in + custom)
agent-smith link all

# Or link to specific target
agent-smith link all --target cursor
```

### Workflow 3: Profile-Based Context Switching

```bash
# Create work profile
agent-smith profile create work
agent-smith profile activate work

# Install work-specific components
agent-smith install all company/internal-tools --profile work

# Apply to editor
agent-smith link all

# Switch to personal profile
agent-smith profile activate personal
agent-smith link all  # Switches editor to personal components
```

### Workflow 4: Project-Local Testing

```bash
# Install components to project directory (isolated)
agent-smith install all github.com/org/experimental --target-dir ./test-components

# Verify installation
ls -la ./test-components/skills/
ls -la ./test-components/agents/

# Clean up when done
rm -rf ./test-components
```

### Workflow 5: Component Lifecycle

```bash
# Install
agent-smith install skill owner/repo skill-name

# Link to editor
agent-smith link skill skill-name

# Update when changes available
agent-smith update skills skill-name

# Remove when no longer needed
agent-smith uninstall skill skill-name  # Auto-unlinks first
```

---

## Key Design Principles

### 1. **Explicit Control**
- Two-step profile activation (activate, then link) gives users control over when changes apply
- Confirmation prompts for bulk/destructive operations (override with `--force`)
- Clear status commands show current state before taking action

### 2. **Isolation & Independence**
- Custom target directories (`--target-dir`) are completely isolated from `~/.agent-smith/`
- Profiles maintain independent component sets
- Unmanaged directories won't be affected by agent-smith operations

### 3. **Flexibility**
- Works with any repository structure through content-based detection
- Supports any Git source (GitHub, GitLab, private repos, local paths)
- Extensible to any editor/tool through custom targets

### 4. **Safety**
- Automatic unlinking before uninstall prevents broken symlinks
- Lock files maintain consistency
- Graceful handling of conflicts and errors
- Non-destructive operations by default

### 5. **Efficiency**
- SHA-based change detection minimizes unnecessary downloads
- Symlinks for zero-copy component sharing
- Bulk operations for managing multiple components

---

## Integration Points

### Compatible AI Environments

**Built-in Support**:
- **OpenCode**: `~/.config/opencode/`
- **Claude Code**: `~/.claude/`

**Custom Target Support** (via `target add`):
- **Cursor**: `~/.cursor/`
- **VS Code**: Any custom path
- **Custom Tools**: Any directory structure

### Repository Compatibility

**Source Types**:
- GitHub (shorthand and full URLs)
- GitLab
- Bitbucket
- Private Git repositories
- Local file paths

**Repository Structures**:
- Standard layouts (`/skills/`, `/agents/`, `/commands/`)
- Monorepos with multiple component types
- Custom nested structures
- Any layout with proper marker files (SKILL.md, AGENT.md, COMMAND.md)

---

## npx install-skill Compatibility

Agent Smith maintains **full backward compatibility** with the existing `npx install-skill` ecosystem:

- **Shared storage**: Both tools use `~/.agent-smith/` directory
- **Lock file format**: Compatible `.skill-lock.json` format
- **Coexistence**: Tools can be used together without conflicts
- **No overwrites**: Respects existing installations from either tool

---

## Configuration Files

### User Configuration: `~/.agent-smith/config.json`

Stores custom target definitions:

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
    }
  ]
}
```

### Active Profile State: `~/.agent-smith/.active-profile`

Simple text file containing the name of the currently active profile:

```
work
```

### Lock Files

- `~/.agent-smith/.skill-lock.json`
- `~/.agent-smith/.agent-lock.json`
- `~/.agent-smith/.command-lock.json`

(Duplicated in each profile directory)

---

## Cross-Platform Support

### Supported Platforms

- **macOS**: Full support with native symlinks
- **Linux**: Full support with native symlinks
- **Windows**: Full support with junction fallback

### Path Handling

- **Tilde expansion**: `~/` automatically expands to home directory
- **Relative paths**: Converted to absolute paths internally
- **Path separators**: Normalized across platforms
- **Drive letters**: Properly handled on Windows

---

## Error Handling & User Experience

### Clear Error Messages

```
Component 'mcp-builder' not installed
Invalid component type 'skilz'

Valid component types:
  - skills
  - agents
  - commands
```

### Progress Feedback

```
Installing 5 components from repository...
✓ Installed skill: accessibility-compliance (1/5)
✓ Installed skill: api-design (2/5)
✓ Installed skill: code-review (3/5)
...

Installed 5 components from repository
```

### Status Display

```
Current Status:
  Active Profile: work
  Detected Targets: opencode, claudecode, cursor

Components in ~/.agent-smith/profiles/work/:
  Skills: 8
  Agents: 3
  Commands: 2
```

---

## Recent Enhancements (2026-01-26 to 2026-01-30)

### Major Features Added

1. **Bulk Component Installation** (2026-01-30)
   - Install all skills/agents/commands from repository by type
   - Commands: `install skills`, `install agents`, `install commands`

2. **Custom Directory Linking** (2026-01-29)
   - Add unlimited custom targets beyond OpenCode/Claude Code
   - Configuration-based target management

3. **Target Directory Feature** (2026-01-28)
   - Install to project-local directories with `--target-dir`
   - Isolated from main `~/.agent-smith/` installation

4. **Uninstall Command** (2026-01-28)
   - Clean removal with automatic unlinking
   - Bulk uninstall by repository URL

5. **Enhanced Profile Management** (2026-01-26 to 2026-01-27)
   - Profile-aware linking (two-step activation)
   - Component add/remove within profiles
   - Detailed profile inspection

6. **Singular/Plural CLI Commands** (2026-01-28)
   - Consistent command structure
   - `link skill` (one) vs `link skills` (all)

### Bug Fixes & Improvements

- Fixed recursive directory copy issues (2026-01-30)
- Fixed skill install name filtering (2026-01-30)
- Cleaned up debug output (2026-01-30)
- Fixed target filter bugs (2026-01-30)
- Enhanced unlink by target support (2026-01-29)
- Updated README for accuracy (2026-01-29)

---

## Testing & Quality

### Test Coverage

- **Unit Tests**: Core functionality, path handling, validation logic
- **Integration Tests**: End-to-end workflows, Git operations, file system operations
- **Component Browser Tests**: CLI output, user interactions, error messages

### Test Organization

See [TESTING.md](TESTING.md) for comprehensive testing guide.

---

## Documentation

### Available Documentation

- **README.md**: User guide with quick start and common workflows
- **CONFIG.md**: Comprehensive configuration reference
- **TESTING.md**: Testing guide for contributors
- **PRODUCT_SNAPSHOT.md**: This document (product overview)

### In-Code Documentation

- All public functions have GoDoc comments
- Command help text available via `--help` flag
- Examples provided in command descriptions

---

## Future Extensibility

### Designed for Growth

The architecture is designed with extensibility in mind:

1. **New AI Environments**: Easy to add support for new AI coding assistants
2. **New Component Types**: Flexible detection system can support new types
3. **New Source Types**: Provider system supports additional Git platforms
4. **Enhanced Profiles**: Foundation for more advanced profile features
5. **Workspace Support**: Architecture supports future workspace/team features

### Not Currently Planned (Non-Goals)

- Social media integration or authentication
- Web-based UI or dashboard
- Automatic background updates
- Enterprise team sharing
- Version constraints or semantic versioning
- Dependency management between components
- IDE/editor plugins beyond CLI

---

## Implementation Details

### Technology Stack

- **Language**: Go 1.23+
- **CLI Framework**: Cobra
- **Git Operations**: Native git commands via subprocess
- **File Operations**: Standard library `os`, `io`, `filepath` packages

### Code Organization

```
agent-smith/
├── cmd/                    # Cobra command definitions
├── internal/
│   ├── downloader/        # Component download logic
│   ├── linker/            # Symlink creation and management
│   ├── metadata/          # Lock file operations
│   ├── uninstaller/       # Component removal logic
│   └── git/               # Git operations
├── pkg/
│   ├── config/            # Configuration management
│   └── profile/           # Profile operations
├── main.go                # Application entry point
└── tests/                 # Test suites
```

---

## Success Metrics

### System Performance

- Individual component operations: <100ms
- Bulk installations: Efficient parallel processing
- Symlink creation: Near-instant
- Update checks: Only network-bound

### Reliability

- Lock files remain valid JSON after all operations
- No orphaned directories after uninstall
- No broken symlinks in target directories
- Zero manual cleanup required

---

## Conclusion

Agent Smith is a production-ready, feature-complete component management system for AI coding environments. It provides a comprehensive CLI tool that handles the entire lifecycle of AI components—from installation and organization to linking and removal—with a focus on flexibility, safety, and user control.

The system's two-tier architecture (centralized storage + symlinks), profile management, and custom target support make it suitable for individual developers, teams, and complex multi-project workflows. Its backward compatibility with the `npx install-skill` ecosystem ensures smooth adoption without disrupting existing setups.

---

**Document Version**: 1.0  
**Last Updated**: January 30, 2026  
**Project Repository**: https://github.com/tgaines/agent-smith
