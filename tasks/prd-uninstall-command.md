# Product Requirements Document: Uninstall Command

## Overview

Add `uninstall` command functionality to agent-smith CLI to enable removal of installed components (skills, agents, commands) from the system. This provides symmetry with the existing `install` command and completes the component lifecycle management.

## Problem Statement

Currently, agent-smith only supports installing components via `install` commands. Users have no way to cleanly remove components except by manually deleting directories and editing lock files. This leads to:
- Orphaned component directories
- Stale lock file entries
- Broken symlinks in target directories
- Manual cleanup burden

## Goals

1. Provide clean removal of individual components
2. Enable bulk removal of all components from a repository
3. Automatically handle unlinking from targets (OpenCode, Claude Code, etc.)
4. Support profile-based uninstallation
5. Maintain lock file integrity
6. Provide symmetrical command structure with `install`

## Non-Goals

- Backup/restore functionality
- Component migration between profiles
- Selective file removal within components
- Undo/rollback of uninstall operations

## User Stories

### Story 1: Remove Individual Component
**As a** developer
**I want to** uninstall a specific skill I no longer need
**So that** I can keep my system clean and organized

```bash
agent-smith uninstall skill mcp-builder
```

### Story 2: Remove All Components from Repository
**As a** developer
**I want to** remove all components I installed from a specific repository
**So that** I can cleanly uninstall an entire component collection

```bash
agent-smith install all https://github.com/anthropics/skills
# Later...
agent-smith uninstall all https://github.com/anthropics/skills
```

### Story 3: Remove Component from Profile
**As a** developer with multiple profiles
**I want to** uninstall components from a specific profile
**So that** I can manage different environments independently

```bash
agent-smith uninstall skill test-skill --profile work
```

## Command Structure

### Individual Component Uninstall

```bash
agent-smith uninstall skill <name> [--profile <name>]
agent-smith uninstall agent <name> [--profile <name>]
agent-smith uninstall command <name> [--profile <name>]
```

**Arguments:**
- `<name>`: Component name (required)

**Flags:**
- `--profile, -p <name>`: Target specific profile instead of ~/.agents/

### Bulk Uninstall

```bash
agent-smith uninstall all <repository-url> [--force]
```

**Arguments:**
- `<repository-url>`: Git repository URL (GitHub shorthand, full URL, or SSH)

**Flags:**
- `--force, -f`: Skip confirmation prompt

## Functional Requirements

### FR1: Individual Component Uninstall

**Behavior:**
1. Validate component exists in lock file
2. Auto-unlink component from all detected targets
3. Remove component directory from filesystem
4. Remove entry from appropriate lock file
5. Display success message

**Error Handling:**
- Component not found: Display error "Component 'X' not installed" and exit
- Directory removal fails: Display error and exit
- Lock file update fails: Display warning but continue

**Output:**
```
Unlinking from 2 target(s)...
✓ Removed skill: mcp-builder
```

### FR2: Bulk Uninstall by Source URL

**Behavior:**
1. Parse and normalize repository URL
2. Scan all lock files (.skill-lock.json, .agent-lock.json, .command-lock.json)
3. Find all components matching source URL
4. Display list of components to be removed
5. Prompt for confirmation (unless --force)
6. For each component:
   - Auto-unlink from targets
   - Remove directory
   - Remove lock file entry
7. Display summary

**URL Matching:**
- Normalize URLs (strip trailing slash, .git suffix)
- Match against `sourceUrl` and `source` fields in lock entries
- Support GitHub shorthand (owner/repo → https://github.com/owner/repo)

**Error Handling:**
- No components found: Display "No components found from this repository"
- Partial failures: Continue with remaining components, report failures at end
- User cancels prompt: Exit without changes

**Output:**
```
Found 12 components from https://github.com/anthropics/skills:
  Skills (8): accessibility-compliance, api-design, code-review, ...
  Agents (3): coding-assistant, debug-helper, test-writer
  Commands (1): format-json

Remove these components? (y/N): y
Unlinking from 2 target(s)...
✓ Removed skill: accessibility-compliance
✓ Removed skill: api-design
✓ Removed skill: code-review
✓ Removed skill: component-docs
✓ Removed skill: error-handling
✓ Removed skill: performance-audit
✓ Removed skill: refactoring-guide
✓ Removed skill: testing-strategy
✓ Removed agent: coding-assistant
✓ Removed agent: debug-helper
✓ Removed agent: test-writer
✓ Removed command: format-json

Removed 12 components from repository
```

### FR3: Auto-Unlinking

**Behavior:**
1. Before removing component directory, check if component is linked
2. Detect all targets (OpenCode, Claude Code, etc.)
3. For each target, remove symlink or copied directory
4. Display brief status: "Unlinking from N target(s)..."
5. Continue with removal even if unlinking fails (log warning)

**Integration:**
- Reuse existing `ComponentLinker.UnlinkComponent()` functionality
- No user interaction required
- Silent if component not linked

### FR4: Profile Support

**Behavior:**
- When `--profile` flag provided:
  - Validate profile exists
  - Use profile directory: `~/.agents/profiles/<profile>/`
  - Use profile-specific lock files
- When no profile flag:
  - Use base directory: `~/.agents/`
  - Use base lock files

**Error Handling:**
- Profile doesn't exist: Display error and exit
- Component not in profile: Display error and exit

### FR5: Lock File Management

**Operations:**
1. Load existing lock file
2. Remove component entry from appropriate map (skills/agents/commands)
3. Preserve remaining entries
4. Write updated JSON with formatting
5. Keep lock file even if empty (maintain structure)

**Lock File Structure:**
```json
{
  "version": 3,
  "skills": {},
  "agents": {},
  "commands": {}
}
```

## Technical Requirements

### TR1: File Structure

**New Files:**
- `internal/uninstaller/uninstaller.go` - Core uninstall logic

**Modified Files:**
- `cmd/root.go` - Add uninstall command structure
- `main.go` - Add handler functions
- `internal/metadata/lock.go` - Add removal functions

### TR2: Uninstaller Package API

```go
package uninstaller

type Uninstaller struct {
    baseDir string
    linker  *linker.ComponentLinker
}

// UninstallComponent removes a single component
func (u *Uninstaller) UninstallComponent(componentType, name string) error

// UninstallAllFromSource removes all components from a repository
func (u *Uninstaller) UninstallAllFromSource(repoURL string, force bool) error

// findComponentsBySource finds all components from a source URL
func (u *Uninstaller) findComponentsBySource(repoURL string) (map[string][]string, error)

// removeComponentDirectory removes component directory from filesystem
func (u *Uninstaller) removeComponentDirectory(componentType, name string) error

// autoUnlinkComponent unlinks component from all targets
func (u *Uninstaller) autoUnlinkComponent(componentType, name string) error
```

### TR3: Metadata Package Extensions

```go
// RemoveLockFileEntry removes a component entry from lock file
func RemoveLockFileEntry(baseDir, componentType, componentName string) error

// FindEntriesBySource finds all lock entries matching source URL
func FindEntriesBySource(baseDir, sourceURL string) (map[string][]string, error)
```

### TR4: URL Normalization

Implement URL normalization for consistent matching:
- Strip trailing slashes
- Remove .git suffix
- Convert GitHub shorthand to full URLs
- Handle case sensitivity

```go
func normalizeURL(url string) string {
    // Expand shorthand: owner/repo -> https://github.com/owner/repo
    // Strip trailing: https://github.com/owner/repo/ -> https://github.com/owner/repo
    // Remove .git: https://github.com/owner/repo.git -> https://github.com/owner/repo
}
```

## User Experience

### Safety & Confirmation

**Individual Components:**
- No confirmation prompt (quick operation)
- Display what's being removed
- Show success message

**Bulk Operations:**
- Show comprehensive list of components
- Single y/n confirmation prompt
- `--force` flag to skip prompt
- Clear summary at end

### Output Guidelines

- Use checkmark (✓) for successful operations
- Single line per component removed
- Group by type in bulk operations
- Brief status for unlinking ("Unlinking from N target(s)...")
- Clear summary line at end

### Error Messages

Keep errors concise and actionable:
- "Component 'X' not installed"
- "Profile 'Y' does not exist"
- "No components found from this repository"
- "Failed to remove component 'X': [reason]"

## Edge Cases

### EC1: Component Not Found
**Scenario:** User tries to uninstall non-existent component
**Behavior:** Display error and exit with code 1
**Message:** "Component 'mcp-builder' not installed"

### EC2: Already Unlinked Component
**Scenario:** Component directory exists but not linked to targets
**Behavior:** Skip unlink step silently, remove directory and lock entry
**Message:** Standard success message

### EC3: Partial Failures in Bulk Operation
**Scenario:** Some components fail to remove during bulk uninstall
**Behavior:** Continue with remaining components, collect errors, report at end
**Message:** 
```
✓ Removed skill: accessibility-compliance
✗ Failed to remove skill: api-design (permission denied)
✓ Removed skill: code-review
...
Removed 11 of 12 components (1 failed)
```

### EC4: Corrupted Lock File
**Scenario:** Lock file is malformed or unreadable
**Behavior:** Attempt directory removal anyway, warn user
**Message:** "Warning: Could not update lock file. Component directory removed."

### EC5: No Components from Repository
**Scenario:** User runs `uninstall all` for repo with no installed components
**Behavior:** Display informational message and exit
**Message:** "No components found from https://github.com/owner/repo"

### EC6: User Cancels Bulk Operation
**Scenario:** User responds 'n' to confirmation prompt
**Behavior:** Exit immediately without changes
**Message:** "Uninstall cancelled"

## Testing Requirements

### Unit Tests
- URL normalization
- Lock file entry removal
- Component discovery by source URL
- Error handling for missing components

### Integration Tests
- Install then uninstall (verify clean removal)
- Uninstall linked component (verify auto-unlink)
- Bulk uninstall from repository
- Profile-based uninstall
- Edge cases (not found, partial failures, etc.)

### Manual Testing Scenarios
1. Install multiple components from repo, uninstall all
2. Install to profile, uninstall from profile
3. Install and link components, verify unlinking during uninstall
4. Test with invalid component names, missing profiles
5. Test --force flag behavior
6. Verify lock file integrity after operations

## Success Metrics

- Command executes successfully in <100ms for individual components
- Lock files remain valid JSON after all operations
- No orphaned directories after uninstall
- No broken symlinks in target directories
- Zero manual cleanup required after uninstall

## Documentation Requirements

- Update README.md with uninstall command examples
- Add uninstall section to user guide
- Update command reference documentation
- Add examples for common workflows

## Future Enhancements (Out of Scope)

- `--dry-run` flag to preview removals
- Backup components before removal
- Undo last uninstall operation
- Uninstall with dependency checking
- Interactive selection of components to remove
- Archive components instead of deleting

## Appendix: Command Reference

### Complete Command Set

```bash
# Individual component uninstall
agent-smith uninstall skill <name> [--profile <name>]
agent-smith uninstall agent <name> [--profile <name>]
agent-smith uninstall command <name> [--profile <name>]

# Bulk uninstall
agent-smith uninstall all <repository-url> [--force]

# Examples
agent-smith uninstall skill mcp-builder
agent-smith uninstall agent coding-assistant --profile work
agent-smith uninstall all https://github.com/anthropics/skills
agent-smith uninstall all anthropics/skills --force
```

### Help Text

```
Usage: agent-smith uninstall [command]

Remove installed components from the system.

This command removes components (skills, agents, commands) from ~/.agents/
or from specific profiles. Components are automatically unlinked from all
detected targets before removal.

Available Commands:
  skill       Remove a specific skill
  agent       Remove a specific agent
  command     Remove a specific command
  all         Remove all components from a repository

Flags:
  -p, --profile string   Target specific profile
  -f, --force           Skip confirmation prompts (bulk operations only)
  -h, --help            Help for uninstall

Examples:
  # Remove a specific skill
  agent-smith uninstall skill mcp-builder

  # Remove from profile
  agent-smith uninstall skill test-skill --profile work

  # Remove all components from a repository
  agent-smith uninstall all https://github.com/anthropics/skills

  # Remove without confirmation
  agent-smith uninstall all anthropics/skills --force
```

## Acceptance Criteria

- [x] Individual component uninstall works for skills, agents, and commands
- [x] Bulk uninstall removes all components from specified repository
- [x] Components are automatically unlinked before removal
- [x] Profile flag works correctly for all uninstall operations
- [x] Lock files are updated correctly after removal
- [x] Component directories are completely removed
- [x] Confirmation prompt works for bulk operations
- [x] --force flag skips confirmation
- [x] Error messages are clear and actionable
- [x] Success messages indicate what was removed
- [x] No orphaned files or symlinks after uninstall
- [x] All edge cases are handled gracefully
- [x] Integration tests pass
- [x] Documentation is updated

## Open Questions

None - all design decisions confirmed with stakeholder.
