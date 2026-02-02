# PRD: Component Materialization for Project-Local Sharing

**Created**: 2026-02-02 02:41 UTC

---

## Introduction

Implement component materialization to enable copying skills, agents, and commands from `~/.agent-smith/` to project-local directories (`.opencode/`, `.claude/`) for version control and team sharing. This allows teams to commit components to git repositories so team members can use them immediately upon cloning without requiring agent-smith installation or manual component installation.

**Problem Statement**: Currently, components installed via agent-smith are only available in the user's global `~/.agent-smith/` directory. There's no way to share components with a team via version control, requiring each team member to manually install the same components.

**Solution**: Enable materialization (copying) of components to project-local directories that editors auto-discover, with full provenance tracking to know where components came from.

---

## Goals

- Enable teams to version-control AI components alongside project code
- Support both OpenCode (`.opencode/`) and Claude Code (`.claude/`) project discovery
- Track component provenance (source repo, commit hash, installation metadata, source profile)
- Provide seamless workflow similar to existing `link` command with `--target` flag
- Auto-create project structure on first materialize (no explicit init needed)
- Support offline/air-gapped development by committing components to project repos
- Maintain separation between global components and project-specific components
- Enable future sync detection through hash-based tracking
- Support materialization from profiles (active profile, specific profile, or base directory)

---

## User Stories

- [x] Story-001: As a project maintainer, I want to materialize a skill to my project so that my team can use it by cloning the repo.

  **Acceptance Criteria:**
  - Command `agent-smith materialize skill <name> --target <opencode|claudecode|all>` copies skill from `~/.agent-smith/skills/` to `.opencode/skills/` or `.claude/skills/`
  - Entire skill directory is copied recursively, preserving structure
  - Provenance metadata is recorded in `.opencode/.materializations.json` or `.claude/.materializations.json`
  - Clear output shows what was copied and where
  - If target directory doesn't exist, it's created automatically
  
  **Testing Criteria:**
  **Unit Tests:**
  - Directory copy logic with recursive subdirectories
  - Metadata extraction from lock files
  - Target directory resolution (opencode → .opencode, claudecode → .claude)
  
  **Integration Tests:**
  - End-to-end materialize skill command with file system verification
  - Auto-creation of project structure on first materialize
  - Metadata file creation and updates
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

- [x] Story-002: As a project maintainer, I want to materialize agents and commands to my project so that all component types are supported.

  **Acceptance Criteria:**
  - Command `agent-smith materialize agent <name> --target <type>` copies agent to project
  - Command `agent-smith materialize command <name> --target <type>` copies command to project
  - All component types use same materialization logic and metadata structure
  - Each component type has dedicated subdirectory (skills/, agents/, commands/)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Component type validation and routing
  - Directory path resolution for each component type
  
  **Integration Tests:**
  - Materialize agent command end-to-end
  - Materialize command command end-to-end
  - Metadata tracking for all component types
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

- [ ] Story-003: As a project maintainer, I want project directory auto-detection so that I can materialize from anywhere in my project.

  **Acceptance Criteria:**
  - System walks up directory tree from current working directory
  - Stops when `.opencode/` or `.claude/` directory is found
  - Stops at home directory or filesystem root if no project found
  - Clear error message if no project found with helpful next steps
  - `--project-dir <path>` flag allows overriding auto-detection
  
  **Testing Criteria:**
  **Unit Tests:**
  - Directory tree walking logic
  - Project root detection algorithm
  - Home directory and filesystem root boundary detection
  
  **Integration Tests:**
  - Auto-detection from nested subdirectories
  - Override with --project-dir flag
  - Error handling when no project found
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

- [ ] Story-004: As a project maintainer, I want provenance tracking so that I know where each materialized component came from.

  **Acceptance Criteria:**
  - `.materializations.json` file created in `.opencode/` or `.claude/` directory
  - Metadata includes: source repo URL, source type (github/gitlab/local), commit hash, original path, materialization timestamp
  - Metadata includes sourceHash and currentHash for future sync detection
  - Metadata loaded from existing `~/.agent-smith/.skill-lock.json` (or agent/command lock files)
  - JSON formatted with indentation for readability and git diffing
  
  **Testing Criteria:**
  **Unit Tests:**
  - Metadata extraction from lock files
  - Hash calculation for component directories
  - JSON serialization and deserialization
  
  **Integration Tests:**
  - Metadata file creation on first materialize
  - Metadata file updates on subsequent materializations
  - Multiple components tracked in same metadata file
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

- [ ] Story-005: As a project maintainer, I want to materialize to specific targets so that I can support both OpenCode and Claude Code users.

  **Acceptance Criteria:**
  - `--target opencode` materializes to `.opencode/` directory
  - `--target claudecode` materializes to `.claude/` directory
  - `--target all` materializes to both directories with separate metadata files
  - Target flag is required, error message shown if omitted
  - AGENT_SMITH_TARGET environment variable can set default target
  - Target resolution reuses existing logic from `link` command
  
  **Testing Criteria:**
  **Unit Tests:**
  - Target resolution logic (flag → directory mapping)
  - Environment variable handling for default target
  - "all" target expansion to multiple targets
  
  **Integration Tests:**
  - Materialize to opencode target
  - Materialize to claudecode target  
  - Materialize to all targets creates both directories
  - AGENT_SMITH_TARGET environment variable respected
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

- [ ] Story-006: As a project maintainer, I want automatic structure creation so that I don't need to manually set up directories.

  **Acceptance Criteria:**
  - First materialize command automatically creates `.opencode/` or `.claude/` directory
  - Subdirectories created: `skills/`, `agents/`, `commands/`
  - Empty `.materializations.json` created with proper structure
  - Clear output shows structure was created
  - No explicit `init` command required (though can be provided as optional convenience)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Directory structure creation logic
  - Empty metadata file initialization
  
  **Integration Tests:**
  - First materialize creates full structure
  - Subsequent materializations don't recreate existing structure
  - Structure creation for both opencode and claudecode targets
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

- [ ] Story-007: As a project maintainer, I want conflict handling so that I don't accidentally overwrite existing materialized components.

  **Acceptance Criteria:**
  - If component already exists in target directory, skip silently if files are identical (hash match)
  - If files differ, error with message: "Component exists and differs. Use --force to overwrite"
  - `--force` flag allows overwriting existing components
  - Hash comparison determines if files are identical
  - Clear output indicates when components are skipped vs copied
  
  **Testing Criteria:**
  **Unit Tests:**
  - File existence detection
  - Hash comparison logic
  - Force flag handling
  
  **Integration Tests:**
  - Materialize identical component twice (skip second time)
  - Materialize modified component (error without --force)
  - Force flag successfully overwrites existing component
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

- [ ] Story-008: As a project maintainer, I want to materialize all components at once so that I can quickly set up a project.

  **Acceptance Criteria:**
  - Command `agent-smith materialize all --target <type>` materializes all installed components
  - Processes all skills, agents, and commands
  - Shows progress for each component
  - Summary shows total components materialized, skipped, and any errors
  - Continues on error with individual components (doesn't abort entire operation)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Component enumeration from lock files
  - Batch processing logic
  - Error handling and continuation
  
  **Integration Tests:**
  - Materialize all command with multiple components
  - Partial failure handling (some succeed, some fail)
  - Progress output and final summary
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

- [ ] Story-009: As a developer, I want dry-run mode so that I can preview what will be materialized before doing it.

  **Acceptance Criteria:**
  - `--dry-run` flag shows what would be materialized without making changes
  - Output includes: component name, source path, destination path, file count
  - No files copied in dry-run mode
  - No metadata files created or modified in dry-run mode
  - Clear indication that it's a dry-run in output
  
  **Testing Criteria:**
  **Unit Tests:**
  - Dry-run flag detection and handling
  - Output formatting for dry-run mode
  
  **Integration Tests:**
  - Dry-run doesn't create files or directories
  - Dry-run output matches actual materialize output format
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

- [ ] Story-010: As a team member, I want to see which components are materialized in a project so that I know what's available.

  **Acceptance Criteria:**
  - Command `agent-smith materialize list` shows all materialized components in current project
  - Output grouped by target (opencode, claudecode) and component type
  - Shows component name and source repo for each
  - Works from any directory in project (auto-detects project root)
  - Clear message if no components materialized yet
  
  **Testing Criteria:**
  **Unit Tests:**
  - Metadata file parsing for list display
  - Output formatting logic
  
  **Integration Tests:**
  - List command in project with materialized components
  - List command in project without components
  - List command with multiple targets
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

- [ ] Story-011: As a team member, I want to see provenance information for a specific component so that I can understand its origin.

  **Acceptance Criteria:**
  - Command `agent-smith materialize info <type> <name>` shows detailed provenance
  - Output includes: source repo URL, commit hash, original path, materialization timestamp, target
  - Shows hash information for sync status
  - Clear error if component not materialized in current project
  - `--target <type>` flag optional to specify which target to check
  
  **Testing Criteria:**
  **Unit Tests:**
  - Metadata lookup logic
  - Info output formatting
  
  **Integration Tests:**
  - Info command for materialized component
  - Info command for non-existent component (error)
  - Info command with target specification
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

- [ ] Story-012: As a developer, I want helpful error messages so that I can quickly resolve issues.

  **Acceptance Criteria:**
  - Component not installed in ~/.agent-smith/: suggest install command
  - No project found: suggest creating .opencode/ or .claude/ directory
  - Missing --target flag: list valid target options
  - Component not found in project: list available components
  - All error messages include actionable next steps
  
  **Testing Criteria:**
  **Unit Tests:**
  - Error message formatting
  - Suggestion generation logic
  
  **Integration Tests:**
  - Each error scenario produces helpful message
  - Error messages include correct command examples
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

- [ ] Story-013: As a project maintainer, I want to materialize from active profile so that I can share profile-specific components with my team.

  **Acceptance Criteria:**
  - When profile is active, materialize copies from `~/.agent-smith/profiles/<profile-name>/skills/` instead of `~/.agent-smith/skills/`
  - Active profile detection uses same logic as existing profile system
  - Profile name recorded in metadata for provenance tracking
  - Clear output indicates which profile was used as source
  - Works for all component types (skills, agents, commands)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Active profile detection logic
  - Profile-aware source path resolution
  - Profile name extraction and storage
  
  **Integration Tests:**
  - Materialize with active profile
  - Materialize with no active profile (base directory)
  - Profile name correctly recorded in metadata
  - Output shows profile information
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

- [ ] Story-014: As a project maintainer, I want to materialize from specific profile so that I can choose which profile to copy from.

  **Acceptance Criteria:**
  - `--from-profile <name>` flag allows explicit profile selection
  - Overrides active profile if both specified
  - Special value `--from-profile base` materializes from `~/.agent-smith/` (no profile)
  - Error if specified profile doesn't exist with list of available profiles
  - Error if component doesn't exist in specified profile
  - Profile name recorded in metadata
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile name parsing and validation
  - Profile directory resolution
  - "base" special value handling
  
  **Integration Tests:**
  - Materialize with --from-profile flag
  - Override active profile with explicit flag
  - Error handling for non-existent profile
  - Error handling for missing component in profile
  - Base directory materialization
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

- [ ] Story-015: As a team member, I want to see which profile a component came from so that I know where to install it if needed.

  **Acceptance Criteria:**
  - `materialize list` shows profile name for each component (if materialized from profile)
  - `materialize info` displays profile information in provenance details
  - Profile shown as "base" or empty if materialized from `~/.agent-smith/`
  - Clear distinction between profile-sourced and base-sourced components
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile display formatting
  - Profile name vs base distinction
  
  **Integration Tests:**
  - List command shows profile information
  - Info command shows profile in provenance
  - Both profile and base components displayed correctly
  
  **Component Browser Tests:**
  - Not applicable (CLI tool)

---

## Functional Requirements

### Core Materialization
- FR-1: The system SHALL copy entire component directories recursively from source to `.{target}/{type}/{name}`
- FR-2: The system SHALL preserve directory structure and all files within components
- FR-3: The system SHALL support materialization of skills, agents, and commands
- FR-4: The system SHALL materialize to `.opencode/` for opencode target and `.claude/` for claudecode target

### Profile Support
- FR-5: The system SHALL detect active profile and use it as default materialization source
- FR-6: The system SHALL copy from `~/.agent-smith/profiles/<profile-name>/{type}/` when profile is active
- FR-7: The system SHALL copy from `~/.agent-smith/{type}/` when no profile is active
- FR-8: The system SHALL support `--from-profile <name>` flag to explicitly select source profile
- FR-9: The system SHALL support `--from-profile base` to materialize from base directory regardless of active profile
- FR-10: The system SHALL record source profile name in metadata
- FR-11: The system SHALL error if specified profile doesn't exist
- FR-12: The system SHALL error if component doesn't exist in specified profile

### Project Detection
- FR-13: The system SHALL auto-detect project root by walking up directory tree from current working directory
- FR-14: The system SHALL recognize `.opencode/` or `.claude/` directories as project markers
- FR-15: The system SHALL stop walking at home directory or filesystem root
- FR-16: The system SHALL support `--project-dir <path>` flag to override auto-detection

### Target Management
- FR-17: The system SHALL require `--target` flag with values: opencode, claudecode, or all
- FR-18: The system SHALL respect AGENT_SMITH_TARGET environment variable as default target
- FR-19: The system SHALL expand `--target all` to materialize to both opencode and claudecode targets
- FR-20: The system SHALL create separate metadata files for each target

### Metadata and Provenance
- FR-21: The system SHALL create `.materializations.json` file in each target directory
- FR-22: The system SHALL record: source, sourceType, sourceProfile, commitHash, originalPath, materializedAt, sourceHash, currentHash
- FR-23: The system SHALL extract metadata from lock files (base or profile-specific)
- FR-24: The system SHALL format JSON with indentation for readability
- FR-25: The system SHALL calculate content hashes for sync detection support

### Structure Creation
- FR-26: The system SHALL automatically create target directory structure on first materialize
- FR-27: The system SHALL create subdirectories: skills/, agents/, commands/
- FR-28: The system SHALL initialize empty `.materializations.json` with proper schema

### Conflict Handling
- FR-29: The system SHALL calculate hash of existing component before materializing
- FR-30: The system SHALL skip copying if existing component has identical hash (files unchanged)
- FR-31: The system SHALL error if existing component has different hash unless `--force` flag provided
- FR-32: The system SHALL overwrite existing component when `--force` flag is present

### User Interface
- FR-33: The system SHALL provide clear output showing what was materialized and where
- FR-34: The system SHALL indicate source profile in output when materializing from profile
- FR-35: The system SHALL support `--dry-run` flag to preview without making changes
- FR-36: The system SHALL support `--quiet` flag for minimal output
- FR-37: The system SHALL provide progress indication when materializing multiple components
- FR-38: The system SHALL show summary with counts of successful, skipped, and failed materializations

### Information Commands
- FR-39: The system SHALL provide `materialize list` command to show materialized components
- FR-40: The system SHALL provide `materialize info <type> <name>` command to show component provenance
- FR-41: The system SHALL group list output by target and component type
- FR-42: The system SHALL display source profile in list and info commands

---

## Non-Goals (Out of Scope)

### Phase 1 Exclusions
- No `adopt` command to import components from project back to ~/.agent-smith/ (explicitly excluded per requirements)
- No `materialize sync` command to update stale components (deferred to future)
- No `materialize status` command to show drift detection (deferred to future)
- No automatic sync or update checking
- No git integration (auto-commit, auto-add) - user manages git manually
- No automatic README.md generation in project directories

### General Exclusions
- No modification of source components in ~/.agent-smith/
- No modification of existing link behavior
- No changes to install/uninstall workflows
- No support for partial component materialization (always copies full directory)
- No compression or optimization of materialized components
- No network operations (all local copying)
- No component validation beyond hash comparison
- No automatic cleanup of orphaned metadata entries
- No migration tools for existing projects
- No multi-project management (one project at a time)

---

## Technical Implementation Notes

### Architecture
- Reuse target resolution logic from existing `link` command
- Create new `pkg/project/` package for project detection and metadata
- Create new `internal/materializer/` package for copy operations
- Extend existing lock file reading infrastructure

### File Locations
- Project metadata: `.opencode/.materializations.json` and `.claude/.materializations.json`
- Source metadata (base): `~/.agent-smith/.skill-lock.json`, `.agent-lock.json`, `.command-lock.json`
- Source metadata (profile): `~/.agent-smith/profiles/<profile-name>/.skill-lock.json`, etc.
- Component source (base): `~/.agent-smith/{skills|agents|commands}/{name}/`
- Component source (profile): `~/.agent-smith/profiles/<profile-name>/{skills|agents|commands}/{name}/`
- Component destination: `.{opencode|claude}/{skills|agents|commands}/{name}/`
- Active profile indicator: `~/.agent-smith/.active-profile`

### Hash Calculation
- Use SHA-256 hash of all file contents concatenated (sorted by path)
- Store both sourceHash (at materialization time) and currentHash (computed on check)
- Enable future sync detection without implementing sync commands now

### Metadata Schema
```json
{
  "version": 1,
  "skills": {
    "component-name": {
      "source": "github.com/user/repo",
      "sourceType": "github",
      "sourceProfile": "work",
      "commitHash": "abc123",
      "originalPath": "skills/component-name/SKILL.md",
      "materializedAt": "2024-01-15T10:30:00Z",
      "sourceHash": "sha256:...",
      "currentHash": "sha256:..."
    },
    "another-component": {
      "source": "github.com/user/other-repo",
      "sourceType": "github",
      "sourceProfile": "",
      "commitHash": "def456",
      "originalPath": "another-component.md",
      "materializedAt": "2024-01-16T11:00:00Z",
      "sourceHash": "sha256:...",
      "currentHash": "sha256:..."
    }
  },
  "agents": {},
  "commands": {}
}
```

**Note**: `sourceProfile` is empty string or omitted when materialized from base `~/.agent-smith/` directory.

### Command Structure
```
agent-smith materialize
├── skill <name>
├── agent <name>
├── command <name>
├── all
├── list
└── info <type> <name>
```

### Common Flags
- `--target <opencode|claudecode|all>` - Required, which target(s) to materialize to
- `--from-profile <name>` - Optional, materialize from specific profile (use "base" for ~/.agent-smith/)
- `--project-dir <path>` - Optional, override project directory detection
- `--force` - Optional, overwrite existing components
- `--dry-run` - Optional, preview without making changes
- `--quiet` - Optional, minimal output

---

## Success Criteria

### Functionality
- Successfully materialize skills, agents, and commands to project directories
- Auto-detect project root from any subdirectory
- Support both OpenCode and Claude Code targets
- Track full provenance metadata for all materialized components
- Handle conflicts gracefully with hash-based detection

### User Experience
- Clear, actionable error messages
- Intuitive command structure matching existing `link` command
- No manual directory setup required (auto-created)
- Dry-run mode for safe preview

### Team Workflow
- Team members clone repo and immediately have components available
- Components version-controlled in git alongside code
- Clear provenance tracking shows where components originated

### Technical Quality
- Comprehensive test coverage (unit, integration)
- Reuse existing code where possible (target resolution, lock file reading)
- Clean separation of concerns (detection, metadata, copying)
- Extensible design for future sync features

---

## Profile Materialization Workflows

### Workflow 1: Materialize from Active Profile

```bash
# Activate work profile
agent-smith profile activate work

# Install work-specific component to work profile
agent-smith install skill github.com/company/internal enterprise-auth

# Materialize from active work profile to project
cd ~/projects/company-app
agent-smith materialize skill enterprise-auth --target opencode

# Output shows:
# ✓ Materialized skill 'enterprise-auth' from profile 'work'
#   Source:      ~/.agent-smith/profiles/work/skills/enterprise-auth/
#   Destination: .opencode/skills/enterprise-auth/
```

**Result**: Component copied from work profile, metadata records `"sourceProfile": "work"`

### Workflow 2: Materialize from Specific Profile (Override Active)

```bash
# Current profile is 'work'
agent-smith profile activate work

# But want to materialize from personal profile
cd ~/projects/side-project
agent-smith materialize skill my-custom-tool --target opencode --from-profile personal

# Output shows:
# ✓ Materialized skill 'my-custom-tool' from profile 'personal'
#   Source:      ~/.agent-smith/profiles/personal/skills/my-custom-tool/
#   Destination: .opencode/skills/my-custom-tool/
```

**Result**: `--from-profile` overrides active profile

### Workflow 3: Materialize from Base (No Profile)

```bash
# Deactivate any active profile
agent-smith profile deactivate

# Materialize from base ~/.agent-smith/
cd ~/projects/my-app
agent-smith materialize skill standard-tool --target opencode

# Output shows:
# ✓ Materialized skill 'standard-tool'
#   Source:      ~/.agent-smith/skills/standard-tool/
#   Destination: .opencode/skills/standard-tool/
```

**Result**: Metadata has `"sourceProfile": ""` (empty or omitted)

### Workflow 4: Explicit Base Even with Active Profile

```bash
# Work profile is active
agent-smith profile activate work

# But want component from base directory
cd ~/projects/mixed-app
agent-smith materialize skill standard-tool --target opencode --from-profile base

# Output shows:
# ✓ Materialized skill 'standard-tool' from base
#   Source:      ~/.agent-smith/skills/standard-tool/
#   Destination: .opencode/skills/standard-tool/
```

**Result**: `--from-profile base` forces base directory regardless of active profile

### Workflow 5: Team Member Sees Profile Provenance

```bash
# Team member clones project
git clone company/project
cd project

# List what's materialized
agent-smith materialize list

# Output shows:
# Materialized Components in /Users/teammate/projects/project:
#
# OpenCode (.opencode/):
#   Skills (2):
#     • enterprise-auth  (from company/internal, profile: work)
#     • standard-tool    (from user/repo)
#
# Claude Code (.claude/):
#   Skills (0)

# See detailed provenance
agent-smith materialize info skill enterprise-auth

# Output shows:
# Skill: enterprise-auth
# Source:         github.com/company/internal
# Source Profile: work
# Commit:         abc123def
# Materialized:   2024-01-15 10:30:00
# Target:         opencode
```

**Result**: Team member knows to install to 'work' profile if they want to adopt it

---

## Future Enhancements (Not in Phase 1)

### Sync Detection and Updates
- `materialize status` - Show which components are out of sync
- `materialize sync` - Update stale components from source
- Automatic drift detection using hash comparison

### Advanced Features
- `adopt` command - Import components from project to ~/.agent-smith/
- Multi-project management - Track materializations across projects
- Component validation - Verify component structure and format
- Metadata cleanup - Remove orphaned entries

### Quality of Life
- Auto-generate project README.md explaining structure
- Git integration - Auto-commit with helpful messages
- Interactive mode - Prompt for target if not specified
- Batch operations - Materialize specific component patterns

---

## Dependencies

### Existing Code
- `pkg/config/target_manager.go` - Target resolution and detection
- `internal/metadata/lock.go` - Lock file reading
- `pkg/paths/paths.go` - Path utilities and expansion (GetProfileDir, GetActiveProfile)
- `cmd/link.go` - Reference for target flag handling
- `cmd/profile.go` - Profile activation/deactivation logic

### New Code Required
- `pkg/project/detection.go` - Project root detection
- `pkg/project/materialization.go` - Metadata management with profile support
- `pkg/project/profile_resolver.go` - Profile-aware source path resolution
- `internal/materializer/materializer.go` - Copy operations
- `cmd/materialize.go` - CLI commands with profile flags

### External Dependencies
- No new external dependencies required
- Uses existing Go standard library (os, filepath, crypto/sha256, encoding/json)

---

## Acceptance Testing Checklist

### Smoke Tests
- Materialize a skill to opencode target from project root
- Materialize an agent to claudecode target from nested subdirectory
- Materialize all components to both targets
- List materialized components
- Show info for materialized component

### Profile Scenarios
- Materialize from active profile
- Materialize from specific profile with --from-profile flag
- Materialize from base with --from-profile base
- Override active profile with explicit --from-profile
- List shows profile information
- Info displays profile in provenance

### Edge Cases
- Materialize from directory with no project (error)
- Materialize component not in ~/.agent-smith/ (error)
- Materialize component not in specified profile (error)
- Materialize with non-existent profile (error)
- Materialize identical component twice (skip)
- Materialize modified component (error without --force)
- Materialize with --force flag (overwrite)

### Multi-Target Scenarios
- Materialize same component to both targets
- List components when both targets exist
- Info command with multiple targets

### Error Handling
- Missing --target flag shows helpful error
- Invalid target value shows valid options
- Invalid profile name shows available profiles
- Component not found suggests available components
- No project found suggests creating directories

### Metadata Validation
- .materializations.json has correct structure
- Provenance data matches source lock files
- Profile name recorded correctly (or empty for base)
- Hashes calculated correctly
- Timestamps in RFC3339 format
