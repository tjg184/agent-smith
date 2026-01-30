# Agent Smith

Agent Smith is a powerful CLI tool for downloading, managing, and executing AI agents, skills, and commands from git repositories.

Install, manage, and link AI components with ease:
- Download and install agents, skills, and commands from any git repository
- Link components to supported targets (OpenCode, Claude Code, etc.)
- Manage multiple profiles for context switching
- Update and maintain installed components
- Remove components cleanly when no longer needed


## Documentation

- [CONFIG.md](CONFIG.md) - Comprehensive configuration guide
- [TESTING.md](TESTING.md) - Testing guide

## Installation

```bash
# Build from source
go build -o agent-smith

# Install to PATH
cp agent-smith /usr/local/bin/
```

## Quick Start

```bash
# Install a skill from GitHub
agent-smith install skill owner/repo

# Link all components to detected targets
agent-smith link all

# Check status
agent-smith status

# Remove a component when done
agent-smith uninstall skill component-name
```

## Commands

### Install

Download and install components from git repositories.

```bash
# Install a specific skill
agent-smith install skill owner/repo skill-name

# Install a specific agent
agent-smith install agent owner/repo agent-name

# Install a specific command
agent-smith install command owner/repo command-name

# Install all components from a repository
agent-smith install all owner/repo

# Install to a custom directory (project-local, isolated from ~/.agent-smith/)
agent-smith install all owner/repo --target-dir ./tools
agent-smith install skill owner/repo skill-name --target-dir ./my-components
```

**Repository URL formats:**
- GitHub shorthand: `owner/repo`
- Full GitHub URL: `https://github.com/owner/repo`
- GitLab URL: `https://gitlab.com/owner/repo`
- SSH URL: `git@github.com:owner/repo.git`
- Local path: `/path/to/local/repo`

**Custom target directories (`--target-dir` flag):**

The `--target-dir` (or `-t`) flag allows installing components to a custom directory instead of the default `~/.agent-smith/`. This is useful for:
- **Project-local installations**: Keep components version-controlled with your project
- **Isolated testing**: Test components without affecting your main `~/.agent-smith/` installation
- **Offline distribution**: Package components for air-gapped systems

```bash
# Install to a project directory
agent-smith install all github.com/org/tools --target-dir ./tools

# Install with relative path
agent-smith install skill ./my-skill local-skill --target-dir ./test-components

# Install with absolute path
agent-smith install all github.com/org/tools --target-dir /opt/ai-components

# Install with tilde expansion
agent-smith install all github.com/org/tools --target-dir ~/my-project/agents
```

**Important notes about custom directories:**
- Custom directories are **standalone and isolated** from `~/.agent-smith/`
- They create their own subdirectories: `skills/`, `agents/`, `commands/`
- Lock files are stored in the target directory root
- Custom directories are **NOT managed** by `link`, `update`, or `profile` commands
- Use this for testing, distribution, or project-local installations

### Link

Link installed components to detected targets (OpenCode, Claude Code, or custom targets).

```bash
# Link a specific skill
agent-smith link skill mcp-builder

# Link a specific agent
agent-smith link agent coding-assistant

# Link a specific command
agent-smith link command format-json

# Link all components
agent-smith link all

# Link to specific target (built-in or custom)
agent-smith link skill mcp-builder --target opencode
agent-smith link all --target cursor

# Show link status
agent-smith link status

# List all linked components
agent-smith link list
```

**Profile awareness:**
When a profile is active, link commands automatically use components from the active profile directory.

**Custom targets:**
Link commands work seamlessly with custom targets defined via `agent-smith target add`. Use the `--target` flag to link to a specific custom target, or `link all` to link to all detected targets (including custom ones).

### Unlink

Remove linked components from targets.

```bash
# Unlink a specific skill
agent-smith unlink skill mcp-builder

# Unlink a specific agent
agent-smith unlink agent coding-assistant

# Unlink all components
agent-smith unlink all

# Unlink from specific target
agent-smith unlink skill mcp-builder --target opencode
```

### Uninstall

Remove installed components from the system.

```bash
# Remove a specific skill
agent-smith uninstall skill mcp-builder

# Remove a specific agent
agent-smith uninstall agent coding-assistant

# Remove a specific command
agent-smith uninstall command format-json

# Remove all components from a repository
agent-smith uninstall all owner/repo

# Remove without confirmation prompt
agent-smith uninstall all owner/repo --force

# Remove from a specific profile
agent-smith uninstall skill mcp-builder --profile work
```

**What happens during uninstall:**
1. Component is automatically unlinked from all detected targets
2. Component directory is removed from filesystem
3. Entry is removed from lock files

**Safety features:**
- Individual uninstalls execute immediately (fast operation)
- Bulk uninstalls show a list and prompt for confirmation
- Use `--force` flag to skip confirmation prompts
- Components are automatically unlinked before removal to prevent broken symlinks

### Profile

Manage profiles for context switching between different component sets.

```bash
# Create a new profile
agent-smith profile create work

# Activate a profile
agent-smith profile activate work

# Deactivate current profile
agent-smith profile deactivate

# List all profiles
agent-smith profile list

# Delete a profile
agent-smith profile delete work
```

When a profile is active:
- Install commands save components to the profile directory
- Link commands use components from the active profile
- Uninstall commands remove components from the active profile

### Target

Manage custom target directories for linking components beyond built-in OpenCode and Claude Code targets.

```bash
# Add a custom target (e.g., Cursor, VS Code)
agent-smith target add cursor ~/.cursor

# List all targets (built-in and custom)
agent-smith target list

# Remove a custom target
agent-smith target remove cursor
```

**What are targets?**
Targets are directories where agent-smith links your components. Built-in targets (opencode, claudecode) are auto-detected. Custom targets let you integrate agent-smith with any editor or tool.

**Using custom targets with link commands:**
```bash
# Link all components to a custom target
agent-smith link all --target cursor

# Link specific component to custom target
agent-smith link skill my-skill --target cursor

# Check link status (shows all targets including custom)
agent-smith link status
```

**Custom target configuration:**
Custom targets are stored in `~/.agent-smith/config.json`. See [CONFIG.md](CONFIG.md) for detailed documentation.

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

### Update

Check for updates and update installed components.

```bash
# Update a specific component
agent-smith update skill mcp-builder

# Update all components
agent-smith update all
```

### Status

Show current status and active profile.

```bash
agent-smith status
```

## Directory Structure

Agent Smith stores components and configuration in the following locations:

```
~/.agent-smith/
├── skills/              # Installed skills
├── agents/              # Installed agents
├── commands/            # Installed commands
├── config.json          # User configuration (custom targets)
├── .active-profile      # Currently active profile name
├── .skill-lock.json     # Skill lock file
├── .agent-lock.json     # Agent lock file
├── .command-lock.json   # Command lock file
└── profiles/            # Profile directories
    └── work/
        ├── skills/
        ├── agents/
        ├── commands/
        ├── .skill-lock.json
        ├── .agent-lock.json
        └── .command-lock.json
```

## Common Workflows

### Install and link components

```bash
# Install all components from a repository
agent-smith install all owner/awesome-components

# Link them to your editor
agent-smith link all

# Check what's linked
agent-smith link status
```

### Use custom targets for additional editors

```bash
# Add Cursor as a custom target
agent-smith target add cursor ~/.cursor

# Link all components to Cursor
agent-smith link all --target cursor

# Verify the links
agent-smith link status

# List all targets
agent-smith target list
```

### Use profiles for different contexts

```bash
# Create and activate a work profile
agent-smith profile create work
agent-smith profile activate work

# Install work-specific components
agent-smith install all company/internal-skills

# Link work components
agent-smith link all

# Switch back to personal setup
agent-smith profile deactivate
agent-smith link all
```

### Clean up unwanted components

```bash
# Remove a single component
agent-smith uninstall skill old-skill

# Remove all components from a repository
agent-smith uninstall all owner/deprecated-repo

# Force removal without confirmation
agent-smith uninstall all owner/deprecated-repo --force
```

### Update components

```bash
# Update a specific skill
agent-smith update skill mcp-builder

# Update all installed components
agent-smith update all
```

### Test components in isolation

```bash
# Install components to a test directory without affecting ~/.agent-smith/
agent-smith install all github.com/org/experimental-tools --target-dir ./test-components

# Verify the installation
ls -la ./test-components/skills/
ls -la ./test-components/agents/
ls -la ./test-components/commands/

# Clean up when done
rm -rf ./test-components
```

### Package components for offline distribution

```bash
# Install components to a distribution directory
agent-smith install all github.com/org/ai-toolkit --target-dir ./dist/ai-components

# Archive for distribution
tar -czf ai-components.tar.gz -C ./dist ai-components

# On air-gapped system, extract and use
tar -xzf ai-components.tar.gz
ls -la ai-components/skills/
ls -la ai-components/agents/
```

## Testing

See [TESTING.md](TESTING.md) for information about running tests.

## Contributing

Contributions are welcome! Please ensure:
- All tests pass: `go test -tags=integration ./...`
- Code follows Go conventions
- New features include tests

## License

[Add license information here]
