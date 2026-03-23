# Agent Smith

![Agent Smith](./agent-smith.jpg)

Agent Smith is a CLI tool for downloading, managing, and linking AI agents, skills, and commands from git repositories.

## Key Features

| Feature | Description |
|---------|-------------|
| Install | Download agents, skills, and commands from any git repository |
| Link | Connect components to your AI editor |
| Materialize | Copy components to project directories for version control |
| Profiles | Switch between different component sets for different contexts |
| Update | Keep installed components up to date |

See [Supported Targets](#supported-targets) for compatible AI editors.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Commands](#commands)
- [Supported Targets](#supported-targets)
- [Directory Structure](#directory-structure)
- [Common Workflows](#common-workflows)

## Installation

### Quick Install (Recommended)

Install the latest version:
```bash
curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/install.sh | bash
```

Install a specific version:
```bash
curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/install.sh | bash -s -- v1.1.0
```

After installation, add agent-smith to your PATH (the installer will show instructions).

### Alternative Installation Methods

<details>
<summary>Using Go</summary>

```bash
go install github.com/tjg184/agent-smith@latest
```
</details>

<details>
<summary>Manual Download</summary>

1. Download the appropriate binary for your platform from [Releases](https://github.com/tjg184/agent-smith/releases)
2. Extract: `tar -xzf agent-smith_*.tar.gz`
3. Move to PATH: `mv agent-smith ~/.agent-smith/bin/` (or `/usr/local/bin/`)
4. Make executable: `chmod +x ~/.agent-smith/bin/agent-smith`
5. Add `~/.agent-smith/bin` to your PATH
</details>

<details>
<summary>Build from Source</summary>

```bash
git clone https://github.com/tjg184/agent-smith.git
cd agent-smith
just build
just install
```
</details>

### Updating

To update to the latest version, re-run the installation script:
```bash
curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/install.sh | bash --force
```

### Uninstall

Remove the agent-smith binary:
```bash
curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/uninstall.sh | bash
```

Remove binary and all data (skills, agents, profiles):
```bash
curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/uninstall.sh | bash -s -- --purge
```

## Documentation

- [TESTING.md](TESTING.md) - Testing guide

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

### Installation Commands

| Command | Description |
|---------|-------------|
| `install` | Download components from git repositories |
| `uninstall` | Remove installed components from the system |
| `update` | Keep installed components up to date |

```bash
# Install components
agent-smith install skill owner/repo skill-name
agent-smith install agent owner/repo agent-name
agent-smith install command owner/repo command-name
agent-smith install all owner/repo

# Uninstall components
agent-smith uninstall skill mcp-builder
agent-smith uninstall all owner/repo

# Update components
agent-smith update skills mcp-builder
agent-smith update all
```

**URL formats:** `owner/repo`, `https://github.com/owner/repo`, `git@github.com:owner/repo.git`, `/path/to/local/repo`

### Discovery Commands

| Command | Description |
|---------|-------------|
| `find` | Search for skills in the skills.sh registry |
| `status` | Show installed components and their state |

```bash
agent-smith find skill prd
agent-smith find skill typescript --limit 10
agent-smith find skill prd --json
agent-smith status
```

### Integration Commands

| Command | Description |
|---------|-------------|
| `link` | Connect components to your AI editor |
| `unlink` | Remove components from your AI editor |
| `target` | Manage custom target directories |

```bash
# Link components to targets
agent-smith link skill mcp-builder
agent-smith link all --to opencode

# Unlink components
agent-smith unlink skill mcp-builder
agent-smith unlink all

# Manage custom targets
agent-smith target add cursor ~/.cursor
agent-smith target list
agent-smith target remove cursor
```

See [Supported Targets](#supported-targets) for available options.

### Distribution Commands

| Command | Description |
|---------|-------------|
| `materialize` | Copy components to project directories |
| `profile` | Manage component sets for different contexts |

```bash
# Materialize to project directories
agent-smith materialize skill api-design --target opencode
agent-smith materialize agents --target all

# Manage profiles
agent-smith profile create work
agent-smith profile activate work
agent-smith profile copy skills work-profile personal-profile
```

### Materialize

Copy components from `~/.agent-smith/` to project directories for version control and team sharing. See [Supported Targets](#supported-targets) for available options.

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

## Project Detection

Agent Smith uses intelligent project detection to determine where to materialize components. When you run materialization commands, agent-smith walks up the directory tree looking for project markers.

The first detected marker determines the project root where components will be created. Use `--project-dir` to explicitly specify a location:

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
├── .opencode/        # Materialized components for OpenCode
├── .claude/          # Materialized components for Claude Code
├── .github/          # Materialized components for GitHub Copilot
└── .agents/          # Materialized components for Universal
```

## Supported Targets

| Agent | CLI Flag | Component | Project Path | Global Path |
|-------|----------|----------|--------------|-------------|
| Claude Code | claudecode | skills | .claude/skills/ | ~/.claude/skills/ |
| Claude Code | claudecode | agents | .claude/agents/ | ~/.claude/agents/ |
| Claude Code | claudecode | commands | .claude/commands/ | ~/.claude/commands/ |
| GitHub Copilot | copilot | skills | .github/skills/ | ~/.copilot/skills/ |
| GitHub Copilot | copilot | agents | .github/agents/ | ~/.copilot/agents/ |
| GitHub Copilot | copilot | commands | .github/commands/ | ~/.copilot/commands/ |
| OpenCode | opencode | skills | .opencode/skills/ | ~/.config/opencode/skills/ |
| OpenCode | opencode | agents | .opencode/agents/ | ~/.config/opencode/agents/ |
| OpenCode | opencode | commands | .opencode/commands/ | ~/.config/opencode/commands/ |
| Universal | universal | skills | .agents/skills/ | ~/.agents/skills/ |
| Universal | universal | agents | .agents/agents/ | ~/.agents/agents/ |
| Universal | universal | commands | .agents/commands/ | ~/.agents/commands/ |

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

## Troubleshooting

### Installation Issues

**"command not found: agent-smith"**
- Make sure `~/.agent-smith/bin` is in your PATH
- Restart your shell after adding to PATH
- Verify with: `echo $PATH | grep agent-smith`

**"checksum mismatch"**
- Network issue during download
- Re-run the installer
- If persists, file an issue

**"permission denied"**
- Make sure the binary is executable: `chmod +x ~/.agent-smith/bin/agent-smith`

### Platform Support

Supported platforms:
- macOS (Intel): `darwin_amd64`
- macOS (Apple Silicon): `darwin_arm64`
- Linux (x86_64): `linux_amd64`
- Linux (ARM64): `linux_arm64`

Windows is not currently supported.

## Contributing

Contributions are welcome! Please ensure:
- All tests pass
- Code follows Go conventions
- New features include tests

## License

MIT
