# Agent Smith

![Agent Smith](./agent-smith.jpg)

Agent Smith is a package manager for AI agents, skills, and commands — think npm or pip, but for your AI editor ecosystem.

It lets you discover, install, update, and manage modular AI components from git repositories, then link or materialize them into your editor of choice.

- 🤖 **Agents** — Autonomous sub-agents with a defined role
- 🧩 **Skills** — Reusable prompt instructions (e.g. "write a PRD", "review architecture")
- ⚡ **Commands** — Slash commands available in your editor

Whether you're building, extending, or orchestrating automated agents, Agent Smith keeps your AI stack organized and up to date.

## Installation

**Homebrew (macOS/Linux):**
```bash
brew install tjg184/homebrew-tap/agent-smith
```

**Upgrade:**
```bash
brew upgrade agent-smith
```

<details>
<summary>Other install methods</summary>

**Curl installer:**
```bash
curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/install.sh | bash
```
After installation, add `~/.agent-smith/bin` to your PATH and restart your shell.

**Specific version:**
```bash
curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/install.sh | bash -s -- v1.1.0
```

**Go install:**
```bash
go install github.com/tjg184/agent-smith@latest
```

**Build from source:**
```bash
git clone https://github.com/tjg184/agent-smith.git
cd agent-smith && just build && just install
```

**Update (curl installer):**
```bash
curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/install.sh | bash --force
```

**Uninstall:**
```bash
# Homebrew
brew uninstall agent-smith

# Curl installer — binary only
curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/uninstall.sh | bash
# Curl installer — binary + all data (skills, agents, profiles)
curl -sSL https://raw.githubusercontent.com/tjg184/agent-smith/main/scripts/uninstall.sh | bash -s -- --purge
```
</details>

## Quick Start

```bash
# 1. Install all components from a repository
agent-smith install all owner/repo

# 2. Link them to your AI editor
agent-smith link all

# 3. Check what's installed and linked
agent-smith status
```

## Supported Targets

| Target | CLI Flag | Global Path |
|--------|----------|-------------|
| OpenCode | `opencode` | `~/.config/opencode/` |
| Claude Code | `claudecode` | `~/.claude/` |
| GitHub Copilot | `copilot` | `~/.copilot/` |
| Universal | `universal` | `~/.agents/` |

Use `--to <target>` with link commands or `--target <target>` with materialize commands.

## Commands

### Install

```bash
agent-smith install all owner/repo           # Install everything from a repo
agent-smith install skill owner/repo <name>  # Install a specific skill
agent-smith install agent owner/repo <name>  # Install a specific agent
agent-smith install command owner/repo <name>
agent-smith uninstall skill <name>
agent-smith uninstall all owner/repo
```

Components are always installed into a repository-sourced profile, keeping repos isolated from each other (prevents name collisions).

**URL formats:** `owner/repo`, `https://github.com/owner/repo`, `git@github.com:owner/repo.git`, `/path/to/local/repo`

### Link

Link components from `~/.agent-smith/` to your AI editor via symlink.

```bash
agent-smith link all                     # Link active profile to all editors
agent-smith link all owner/repo          # Link a specific repo's components
agent-smith link all --to opencode       # Link to a specific editor
agent-smith link skill <name>
agent-smith link skills                  # All skills
agent-smith link agent <name>
agent-smith link agents
agent-smith link command <name>
agent-smith link commands
agent-smith link auto                    # Auto-detect components in current dir
agent-smith link list                    # List linked components
agent-smith link status                  # Matrix view: all repos × editors
agent-smith link status --profile <name> # Scope to one repo
agent-smith unlink all                   # Unlink active profile
agent-smith unlink all owner/repo        # Unlink a specific repo
agent-smith unlink skill <name>
```

### Find

Search the [skills.sh](https://skills.sh) registry:

```bash
agent-smith find skill prd
agent-smith find skill typescript --limit 10
agent-smith find skill api --json
```

### Update

```bash
agent-smith update all                   # Update all installed components
agent-smith update all owner/repo        # Update components from a specific repo (finds profile automatically)
agent-smith update skills <name>         # Update a specific skill
agent-smith update agents <name>
agent-smith update commands <name>
```

### Status

```bash
agent-smith status                       # System overview: profile, targets, counts
```

---

## Profiles

Profiles organize components by source repo (auto-created on install) or as custom curated collections. For most users profiles are invisible — you work in terms of repo URLs.

Two profile types:
- **📦 Repository-sourced** — auto-created by `install all`, tied to a repo, keeps components isolated
- **👤 User-created** — manually created for custom collections cherry-picked across repos

```bash
# List and inspect
agent-smith profile list
agent-smith profile list --type repo        # Filter by type (repo or user)
agent-smith profile status [name]           # Details of a profile

# Create and switch
agent-smith profile create work
agent-smith profile activate work
agent-smith profile deactivate

# Manage components in a profile
agent-smith profile add skills <profile> <component>
agent-smith profile remove skills <profile> <component>
agent-smith profile copy skills <src-profile> <dst-profile> <component>
agent-smith profile cherry-pick <new-profile>            # Interactive picker
agent-smith profile cherry-pick <new-profile> --source work

# Lifecycle
agent-smith profile rename <old> <new>
agent-smith profile delete <name>           # Must deactivate first
agent-smith profile share [name]            # Generate shareable setup commands
agent-smith profile share work | pbcopy
agent-smith profile share work --output setup-work.txt
```

---

## Materialize

Materialize copies components from `~/.agent-smith/` into your **project directory** for version control and team sharing. Provenance is tracked in `.component-lock.json`.

```bash
# Copy to project
agent-smith materialize skill <name> --target opencode
agent-smith materialize all --target opencode
agent-smith materialize all owner/repo --target opencode  # Only components from a specific repo
agent-smith materialize all --target all
agent-smith materialize all --target opencode --dry-run   # Preview

# Track sync status
agent-smith materialize status
agent-smith materialize status --target opencode
agent-smith materialize update                            # Re-sync changed components
agent-smith materialize update --force                    # Re-sync everything

# Inspect
agent-smith materialize list
agent-smith materialize info skills <name>
```

**Project detection:** agent-smith walks up from the current directory to find the project root. Override with `--project-dir <path>` or set `AGENT_SMITH_TARGET` env var.

**Project paths per target:**

| Target | Skills | Agents | Commands |
|--------|--------|--------|----------|
| OpenCode | `.opencode/skills/` | `.opencode/agents/` | `.opencode/commands/` |
| Claude Code | `.claude/skills/` | `.claude/agents/` | `.claude/commands/` |
| Copilot | `.github/skills/` | `.github/agents/` | `.github/commands/` |
| Universal | `.agents/skills/` | `.agents/agents/` | `.agents/commands/` |

---

## Custom Targets

Register directories for editors not natively supported:

```bash
agent-smith target add cursor ~/.cursor
agent-smith target list
agent-smith target remove cursor
```

---

## Common Workflows

**Install and use a repo:**
```bash
agent-smith install all owner/repo
agent-smith link all owner/repo
```

**Use components from multiple repos:**
```bash
agent-smith install all owner/repo-a
agent-smith install all owner/repo-b
agent-smith link all owner/repo-a
agent-smith link all owner/repo-b
```

**Remove a repo from your editor:**
```bash
agent-smith unlink all owner/repo        # Remove from editor, keep installed
agent-smith uninstall all owner/repo     # Remove entirely
```

**Switch to a work profile:**
```bash
agent-smith profile create work
agent-smith profile activate work
agent-smith install all company/internal-skills
agent-smith link all
```

**Share your setup with a teammate:**
```bash
agent-smith profile share work --output setup-work.txt
# Teammate runs the generated commands to replicate your setup
```

**Materialize for a project:**
```bash
agent-smith materialize all --target opencode
git add .opencode/ && git commit -m "add AI components"
# Later: re-sync after upstream updates
agent-smith update all
agent-smith materialize update
```

---

## Troubleshooting

**"command not found: agent-smith"** — If installed via Homebrew, run `brew doctor` to diagnose PATH issues. If installed via the curl installer, add `~/.agent-smith/bin` to your PATH and restart your shell.

**"permission denied"** — Curl installer only: `chmod +x ~/.agent-smith/bin/agent-smith`

**"checksum mismatch"** — Network issue; re-run the installer.

**Supported platforms:** macOS (Intel/Apple Silicon), Linux (x86_64/ARM64). Windows is not supported.

---

## Global Flags

```
--verbose    Show informational output
--debug      Enable debug output
-v           Show version
```

## Contributing

Contributions are welcome. Ensure all tests pass and new features include tests.

See [TESTING.md](TESTING.md) for the testing guide.

## License

MIT
