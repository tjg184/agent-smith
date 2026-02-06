# Agent Smith

```
  ___                   _     _____           _ _   _
 / _ \                 | |   /  ___|         (_) | | |
/ /_\ \ __ _  ___ _ __ | |_  \ `--. _ __ ___  _| |_| |__
|  _  |/ _` |/ _ \ '_ \| __|  `--. \ '_ ` _ \| | __| '_ \
| | | | (_| |  __/ | | | |_  /\__/ / | | | | | | |_| | | |
\_| |_/\__, |\___|_| |_|\__| \____/|_| |_| |_|_|\__|_| |_|
        __/ |
       |___/
```

Agent Smith is a powerful CLI tool for downloading, managing, and executing AI agents, skills, and commands from git repositories.

Install, manage, and link AI components with ease:
- Download and install agents, skills, and commands from any git repository
- Link components to supported targets (OpenCode, Claude Code, etc.)
- Materialize components to project directories for version control and team sharing
- Manage multiple profiles for context switching
- Update and maintain installed components
- Remove components cleanly when no longer needed


## Documentation

- [TESTING.md](TESTING.md) - Testing guide

## Installation

```bash
# Build from source
just build

# Install to $GOPATH/bin
just install
```

## Quick Start

```bash
# Install all components from a repository
agent-smith install all owner/repo

# Link all components to detected targets
agent-smith link all

# Check status
agent-smith status

# Unlink all components when done
agent-smith unlink all
```

## Commands

### Install

Download and install components from git repositories.

```bash
# Install specific components
agent-smith install skill owner/repo skill-name
agent-smith install agent owner/repo agent-name
agent-smith install command owner/repo command-name

# Install all components from a repository
agent-smith install all owner/repo

# Install to custom directory
agent-smith install all owner/repo --install-dir ./tools
```

**URL formats:** `owner/repo`, `https://github.com/owner/repo`, `git@github.com:owner/repo.git`, `/path/to/local/repo`

### Link

Link installed components to detected targets (OpenCode, Claude Code, or custom).

```bash
# Link specific or all components
agent-smith link skill mcp-builder
agent-smith link agent coding-assistant
agent-smith link all

# Link to specific target
agent-smith link all --to opencode
agent-smith link all --to cursor

# Link with profile options
agent-smith link all --profile work
agent-smith link all --all-profiles

# Show status
agent-smith link status
agent-smith link list
```

### Unlink

Remove linked components from targets.

```bash
agent-smith unlink skill mcp-builder
agent-smith unlink agent coding-assistant
agent-smith unlink all
agent-smith unlink all --profile work
agent-smith unlink all --all-profiles
```

### Uninstall

Remove installed components from the system.

```bash
agent-smith uninstall skill mcp-builder
agent-smith uninstall agent coding-assistant
agent-smith uninstall command format-json
agent-smith uninstall all owner/repo
agent-smith uninstall all owner/repo --force
agent-smith uninstall skill mcp-builder --profile work
```

### Profile

Manage profiles for context switching.

```bash
# Basic management
agent-smith profile create work
agent-smith profile activate work
agent-smith profile deactivate
agent-smith profile list
agent-smith profile show work
agent-smith profile delete work

# Copy components between profiles
agent-smith profile add skills work-profile api-design
agent-smith profile copy skills work-profile personal-profile api-design
agent-smith profile remove skills work-profile old-skill
agent-smith profile cherry-pick new-profile --source work --source personal
```

### Target

Manage custom target directories for linking components.

```bash
agent-smith target add cursor ~/.cursor
agent-smith target list
agent-smith target remove cursor
```

### Update

```bash
agent-smith update skills mcp-builder
agent-smith update all
```

### Materialize

Copy components from `~/.agent-smith/` to project directories (`.opencode/`, `.claude/`) for version control and team sharing.

```bash
# Materialize specific components
agent-smith materialize skill api-design --target opencode
agent-smith materialize agent coding-assistant --target claudecode

# Materialize all components of a type
agent-smith materialize skills --target opencode
agent-smith materialize agents --target all

# Options
agent-smith materialize skill api-design --target opencode --profile work
agent-smith materialize skill api-design --target opencode --dry-run
agent-smith materialize skill api-design --target opencode --force
```

### Status

```bash
agent-smith status
```

## Project Detection

Agent Smith uses intelligent project detection to determine where to materialize components. When you run materialization commands, agent-smith walks up the directory tree looking for project markers (`.git/`, `.opencode/`, `.claude/`, `go.mod`, etc.).

The first detected marker determines the project root where `.opencode/` or `.claude/` will be created. Use `--project-dir` to explicitly specify a location:

```bash
agent-smith materialize skill api-design --target opencode --project-dir ~/my-project
```

## Directory Structure

```
~/.agent-smith/
├── skills/           # Installed skills
├── agents/           # Installed agents
├── commands/         # Installed commands
├── config.json       # Custom targets
├── .active-profile   # Active profile
├── .*-lock.json      # Lock files
└── profiles/         # Profile directories

~/my-project/
├── .opencode/        # Materialized components
└── .claude/          # Materialized components
```

## Common Workflows

### Install and use components

```bash
agent-smith install all owner/repo
agent-smith link all
agent-smith status
```

### Use profiles

```bash
agent-smith profile create work
agent-smith profile activate work
agent-smith install all company/internal-skills
agent-smith link all
```

### Materialize for team sharing

```bash
agent-smith materialize skill api-design --target opencode
agent-smith materialize agents --target opencode
```

### Update materialized components

```bash
agent-smith materialize status
agent-smith materialize update
```

## Testing

See [TESTING.md](TESTING.md) for information about running tests.

## Contributing

Contributions are welcome! Please ensure:
- All tests pass: `go test -tags=integration ./...`
- Code follows Go conventions
- New features include tests

## License

MIT
