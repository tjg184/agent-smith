# Agent Smith

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

# Link all components from a specific profile (bypasses active profile)
agent-smith link all --profile work

# Link all components from all profiles simultaneously
agent-smith link all --all-profiles

# Show link status
agent-smith link status

# List all linked components
agent-smith link list
```

**Profile awareness:**
When a profile is active, link commands automatically use components from the active profile directory. You can also use the `--profile` flag to link components from a specific profile without switching to it first, or use `--all-profiles` to link components from all profiles simultaneously.

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

# Unlink all components from a specific profile (bypasses active profile)
agent-smith unlink all --profile work

# Unlink all components from all profiles
agent-smith unlink all --all-profiles
```

**Profile awareness:**
By default, unlink commands work with components from the currently active profile. You can use the `--profile` flag to unlink components from a specific profile without switching to it first, or use `--all-profiles` to unlink components from all profiles.

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

#### Basic Profile Management

```bash
# Create a new profile
agent-smith profile create work

# Activate a profile
agent-smith profile activate work

# Deactivate current profile
agent-smith profile deactivate

# List all profiles
agent-smith profile list

# Show details about a specific profile
agent-smith profile show work

# Delete a profile
agent-smith profile delete work
```

When a profile is active:
- Install commands save components to the profile directory
- Link commands use components from the active profile
- Uninstall commands remove components from the active profile

#### Advanced Profile Operations

```bash
# Add an existing component from ~/.agent-smith/ to a profile
agent-smith profile add skills work-profile api-design
agent-smith profile add agents work-profile code-reviewer
agent-smith profile add commands dev-profile test-runner

# Copy a component from one profile to another
agent-smith profile copy skills work-profile personal-profile api-design
agent-smith profile copy agents team-profile solo-profile code-reviewer

# Remove a component from a profile
agent-smith profile remove skills work-profile old-skill
agent-smith profile remove agents work-profile deprecated-agent

# Cherry-pick components interactively from existing profiles
agent-smith profile cherry-pick new-profile
agent-smith profile cherry-pick project-x --source work-profile
agent-smith profile cherry-pick custom --source work --source personal
```

**Profile operations:**
- `add` - Copy an existing component from `~/.agent-smith/` to a profile
- `copy` - Copy a component between profiles with independent lock entries
- `remove` - Remove a component from a profile directory
- `cherry-pick` - Interactive UI to select and copy components from existing profiles
- `show` - Display detailed information about a profile's components

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
Custom targets are stored in `~/.agent-smith/config.json`:

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

### Materialize

Materialize components from `~/.agent-smith/` to project directories for version control and team sharing.

```bash
# Materialize a skill to OpenCode
agent-smith materialize skill mcp-builder --target opencode

# Materialize an agent to Claude Code
agent-smith materialize agent coding-assistant --target claudecode

# Materialize a command to both targets
agent-smith materialize command format-json --target all

# Materialize all skills to a target
agent-smith materialize skills --target opencode

# Materialize all agents from a specific profile
agent-smith materialize agents --target claudecode --profile work

# Materialize all commands with preview
agent-smith materialize commands --target all --dry-run

# Materialize from a specific profile
agent-smith materialize skill api-design --target opencode --profile work

# Materialize from base ~/.agent-smith/ directory
agent-smith materialize skill mcp-builder --target opencode --profile base

# Force overwrite existing component
agent-smith materialize skill mcp-builder --target opencode --force

# Preview without making changes
agent-smith materialize skill mcp-builder --target opencode --dry-run

# Override project directory detection
agent-smith materialize skill mcp-builder --target opencode --project-dir ./my-project
```

**What is materialization?**
Materialization copies components from your global `~/.agent-smith/` directory (or a profile) to project-local directories like `.opencode/` or `.claude/`. This allows you to:
- **Version control**: Commit components alongside your project code
- **Team sharing**: Share components with your team via git
- **Project isolation**: Lock specific component versions per project
- **Offline work**: Bundle components with your project

**Materialization targets:**
- `opencode` - Materialize to `.opencode/` directory
- `claudecode` - Materialize to `.claude/` directory  
- `all` - Materialize to both directories

**Environment variable:**
Set `AGENT_SMITH_TARGET` to avoid repeating `--target` flag:
```bash
export AGENT_SMITH_TARGET=opencode
agent-smith materialize skill mcp-builder
```

**Provenance tracking:**
Materialized components are tracked in `.materializations.json` with full provenance including source path, git URL, commit hash, and timestamps. This enables updates and conflict detection.

### Status

Show current status and active profile.

```bash
agent-smith status
```

## Project Detection

Agent Smith uses intelligent project detection to determine where to materialize components. Understanding how this works helps you control where `.opencode/` or `.claude/` directories will be created.

### How Project Detection Works

When you run materialization commands (e.g., `agent-smith materialize skill api-design --target opencode`), agent-smith walks up the directory tree from your current location looking for project markers.

**Search order:**
1. **Preferred markers** (`.opencode/` or `.claude/` directories) - If found, this directory is immediately used as the project root
2. **Project boundary markers** - Common project indicators that define where your project begins
3. **Stop conditions** - Home directory or filesystem root (search stops, error returned)

### Supported Project Markers

**Preferred markers** (immediately recognized):
- `.opencode/` - Agent Smith project directory
- `.claude/` - Claude project directory

**Project boundary markers** (define project root):
- `.git/` - Git repository
- `go.mod` - Go projects
- `package.json` - Node.js projects
- `pyproject.toml` - Python projects
- `Cargo.toml` - Rust projects
- `composer.json` - PHP projects
- `pom.xml` - Java Maven projects
- `build.gradle` - Java Gradle projects
- `Gemfile` - Ruby projects
- `mix.exs` - Elixir projects

### Examples

**Example 1: Working in a nested directory**
```bash
# Directory structure:
# ~/projects/my-app/
# ├── .git/
# ├── src/
# │   └── components/  ← You are here
# └── tests/

cd ~/projects/my-app/src/components
agent-smith materialize skill api-design --target opencode

# Result: Creates ~/projects/my-app/.opencode/
# Detection: Found .git/ at ~/projects/my-app/, created .opencode/ there
```

**Example 2: Project with existing .opencode/**
```bash
# Directory structure:
# ~/projects/my-app/
# ├── .git/
# ├── .opencode/  ← Already exists
# └── src/
#     └── api/  ← You are here

cd ~/projects/my-app/src/api
agent-smith materialize skill api-design --target opencode

# Result: Uses existing ~/projects/my-app/.opencode/
# Detection: Found .opencode/ at ~/projects/my-app/, uses it immediately
```

**Example 3: No project markers**
```bash
# Directory structure:
# ~/random-scripts/  ← No .git, no project files
#     └── utils/  ← You are here

cd ~/random-scripts/utils
agent-smith materialize skill api-design --target opencode

# Result: Error - no project boundary detected
# Fix: Either run 'git init' or 'mkdir .opencode' or use --project-dir flag
```

**Example 4: Overriding detection**
```bash
# Use --project-dir to explicitly specify where to materialize
agent-smith materialize skill api-design --target opencode --project-dir ~/my-custom-location

# Result: Creates ~/my-custom-location/.opencode/
# Detection: Bypassed - explicit directory provided
```

### Important Notes

- **Home directory isolation**: If you have `~/.opencode/` in your home directory, it will NOT be used when you're inside a project boundary (e.g., a Git repository). Agent-smith respects project boundaries and won't cross them.

- **Project boundaries are respected**: Once a project boundary marker is found (like `.git/`), the search stops. Agent-smith will not look in parent directories beyond the project root.

- **Creating .opencode/ automatically**: If you're in a project (detected via boundary markers like `.git/`) but no `.opencode/` exists, it will be created at the project root automatically.

- **No project detected**: If you're in a directory without any project markers, materialization will fail with a helpful error message. You can fix this by:
  1. Running `git init` to create a Git repository
  2. Running `mkdir .opencode` to create the preferred marker
  3. Using `--project-dir` flag to explicitly specify the location

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

**Project directories (materialized components):**

```
~/my-project/
├── .opencode/
│   ├── skills/              # Materialized skills
│   ├── agents/              # Materialized agents
│   ├── commands/            # Materialized commands
│   └── .materializations.json  # Provenance tracking
└── .claude/
    ├── skills/              # Materialized skills
    ├── agents/              # Materialized agents
    ├── commands/            # Materialized commands
    └── .materializations.json  # Provenance tracking
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

### Work with profiles without switching context

```bash
# Link components from a specific profile without activating it
agent-smith link all --profile work

# Unlink components from a specific profile without switching
agent-smith unlink all --profile work

# Quickly switch between different profile contexts
agent-smith link all --profile personal
agent-smith link all --profile work

# Link components from all profiles at once
agent-smith link all --all-profiles
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

### Materialize components for team sharing

```bash
# Install and link components globally
agent-smith install all owner/team-components
agent-smith link all

# Materialize specific components to your project for version control
cd ~/my-project
agent-smith materialize skill api-design --target opencode
agent-smith materialize agent code-reviewer --target opencode
agent-smith materialize command test-runner --target opencode

# Or materialize all components of a specific type
agent-smith materialize skills --target opencode
agent-smith materialize agents --target opencode

# Commit to version control
git add .opencode/
git commit -m "Add team AI components"
git push

# Team members can now use the materialized components
# They're automatically available in .opencode/ directory
```

### Keep materialized components in sync

```bash
# Check if materialized components are up-to-date with GitHub
cd ~/my-project
agent-smith materialize status
# Shows:
#   ✓ api-design (in sync - abc1234)
#   ⚠ python-testing (out of sync - abc1234 → def5678)
#   ✗ old-skill (repository not found)

# Update out-of-sync components directly from GitHub
agent-smith materialize update

# Or force re-download everything from GitHub
agent-smith materialize update --force

# Preview changes first
agent-smith materialize update --dry-run

# Work with specific targets
agent-smith materialize status --target opencode
agent-smith materialize update --target claudecode

# For private repositories, set GitHub token
export GITHUB_TOKEN=ghp_your_token_here
agent-smith materialize status
```

**Note:** The `materialize status` and `materialize update` commands check components directly against their GitHub source repositories, not your local `~/.agent-smith/` library. This means:
- You don't need to run `agent-smith update all` first
- Components show their sync status with the upstream GitHub repository
- Updates download fresh from GitHub, bypassing the local library
- Works standalone in any project, even without agent-smith installed locally


### Build specialized profiles with cherry-pick

```bash
# Create profiles with different component sets
agent-smith profile create backend
agent-smith profile activate backend
agent-smith install all company/backend-tools

agent-smith profile create frontend
agent-smith profile activate frontend
agent-smith install all company/frontend-tools

# Create a full-stack profile by cherry-picking from both
agent-smith profile cherry-pick fullstack --source backend --source frontend
# Interactive UI lets you select specific components

# Activate the combined profile
agent-smith profile activate fullstack
agent-smith link all
```

### Share components between profiles

```bash
# Add a component from base installation to a profile
agent-smith profile add skills work-profile mcp-builder

# Copy a component from one profile to another
agent-smith profile copy skills personal-profile work-profile api-design

# Remove outdated components from a profile
agent-smith profile remove skills work-profile deprecated-skill

# View what's in a profile
agent-smith profile show work-profile
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
