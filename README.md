# Agent Smith

Agent Smith is a powerful CLI tool for downloading, managing, and executing AI agents, skills, and commands from git repositories.

It provides npm-like functionality for AI components, allowing you to:
- Download and install agents, skills, and commands from any git repository
- Link components to supported targets (OpenCode, Claude Code, etc.)
- Manage multiple profiles for context switching
- Update and maintain installed components
- Remove components cleanly when no longer needed

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
agent-smith install skill owner/repo

# Install a specific agent
agent-smith install agent owner/repo

# Install a specific command
agent-smith install command owner/repo

# Install all components from a repository
agent-smith install all owner/repo
```

**Repository URL formats:**
- GitHub shorthand: `owner/repo`
- Full GitHub URL: `https://github.com/owner/repo`
- GitLab URL: `https://gitlab.com/owner/repo`
- SSH URL: `git@github.com:owner/repo.git`
- Local path: `/path/to/local/repo`

### Link

Link installed components to detected targets (OpenCode, Claude Code, etc.).

```bash
# Link a specific skill
agent-smith link skill mcp-builder

# Link a specific agent
agent-smith link agent coding-assistant

# Link a specific command
agent-smith link command format-json

# Link all components
agent-smith link all

# Link to specific target
agent-smith link skill mcp-builder --target opencode

# Show link status
agent-smith link status

# List all linked components
agent-smith link list
```

**Profile awareness:**
When a profile is active, link commands automatically use components from the active profile directory.

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

Agent Smith stores components in the following locations:

```
~/.agents/
├── skills/              # Installed skills
├── agents/              # Installed agents
├── commands/            # Installed commands
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

## Testing

See [TESTING.md](TESTING.md) for information about running tests.

## Contributing

Contributions are welcome! Please ensure:
- All tests pass: `go test -tags=integration ./...`
- Code follows Go conventions
- New features include tests

## License

[Add license information here]
