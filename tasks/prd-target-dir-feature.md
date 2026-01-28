# PRD: Add Custom Target Directory Support to `install all`

## Overview

Add a `--target-dir` flag to the `agent-smith install all` command that allows installing components to a custom base directory instead of the default `~/.agents/`. Custom directories will be standalone and independent from the managed `~/.agents/` ecosystem.

## Motivation

Users need the ability to install components to project-local directories, create offline distributions, or test components without affecting their main `~/.agents/` installation. Currently, all installations go to `~/.agents/` or profiles, which doesn't support these use cases.

## Goals

- Enable installing components to arbitrary directories
- Support relative paths, absolute paths, and tilde expansion
- Auto-create directory structure (skills/, agents/, commands/)
- Store lock files in the target directory
- Maintain backward compatibility with existing behavior
- Keep custom directories isolated from link/update/profile commands

## Non-Goals

- Integration with `link`, `update`, or `profile` commands (custom dirs are standalone)
- Adding `--target-dir` to individual install commands (skill/agent/command)
- Managing multiple custom directories from a central registry
- Auto-discovery of custom directory installations

## User Stories

### Story 1: Project-Local Components
As a developer working on a specific project, I want to install AI components directly into my project directory so they're version-controlled with my code.

```bash
cd /my-project
./agent-smith install all https://github.com/myorg/project-tools --target-dir ./tools
# Result: ./tools/skills/, ./tools/agents/, ./tools/commands/
```

### Story 2: Testing Components
As a component author, I want to test components in isolation without affecting my main `~/.agents/` installation.

```bash
./agent-smith install all https://github.com/me/experimental --target-dir /tmp/test-components
# Test components...
rm -rf /tmp/test-components
```

### Story 3: Offline Distribution
As a systems administrator, I want to package components for offline distribution to air-gapped systems.

```bash
./agent-smith install all https://github.com/company/internal-tools --target-dir ./dist/ai-components
tar -czf components.tar.gz ./dist/ai-components
```

## Design

### Command Syntax

```bash
agent-smith install all <repository-url> [--target-dir <path>]
```

**Parameters:**
- `<repository-url>`: Git repository URL (required)
- `--target-dir <path>`: Custom installation directory (optional)
  - Short form: `-t`
  - Supports: relative paths, absolute paths, tilde expansion
  - Default: `~/.agents/` (if not specified)

### Behavior

1. **Path Resolution**:
   - Tilde expansion: `~/mydir` → `/home/user/mydir`
   - Relative paths: `./local` → `/current/dir/local`
   - Absolute paths: Used as-is

2. **Directory Creation**:
   - Auto-create target directory if it doesn't exist
   - Create subdirectories: `skills/`, `agents/`, `commands/`
   - Use appropriate permissions (755 for dirs)

3. **Component Installation**:
   - Install skills to `<target-dir>/skills/<component-name>/`
   - Install agents to `<target-dir>/agents/<component-name>/`
   - Install commands to `<target-dir>/commands/<component-name>/`

4. **Lock Files**:
   - Store `.skill-lock.json`, `.agent-lock.json`, `.command-lock.json` in `<target-dir>/`
   - Lock files track source, commit hash, and metadata

5. **Isolation**:
   - Custom directories are NOT managed by `link`, `update`, or `profile` commands
   - Users manage custom directories manually
   - No cross-contamination with `~/.agents/`

### Examples

```bash
# Default behavior (no change)
./agent-smith install all openai/cookbook
# → Installs to ~/.agents/

# Relative path
./agent-smith install all openai/cookbook --target-dir ./local-components
# → Installs to ./local-components/

# Absolute path
./agent-smith install all openai/cookbook --target-dir /opt/ai-components
# → Installs to /opt/ai-components/

# Tilde expansion
./agent-smith install all openai/cookbook --target-dir ~/projects/myapp/components
# → Installs to /home/user/projects/myapp/components/

# Short form
./agent-smith install all openai/cookbook -t ./tools
# → Installs to ./tools/
```

## Implementation

### Architecture

```
┌─────────────────────────────────────────┐
│ cmd/root.go                             │
│ - Add --target-dir flag                 │
│ - Update Run handler                    │
└─────────────┬───────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────┐
│ main.go                                 │
│ - Update handleAddAll signature         │
│ - Pass targetDir to BulkDownloader      │
└─────────────┬───────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────┐
│ internal/downloader/bulk.go             │
│ - Add NewBulkDownloaderWithTargetDir()  │
│ - Add resolveTargetDir() helper         │
└─────────────┬───────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────┐
│ internal/downloader/{skill,agent,       │
│                       command}.go       │
│ - Add NewXxxDownloaderWithBaseDir()     │
└─────────────────────────────────────────┘
```

### Files to Modify

| File | Changes | Lines Added |
|------|---------|-------------|
| `cmd/root.go` | Add --target-dir flag, update command | ~15 |
| `main.go` | Update handler signature | ~5 |
| `internal/downloader/bulk.go` | Add constructor + path resolver | ~40 |
| `internal/downloader/skill.go` | Add constructor | ~20 |
| `internal/downloader/agent.go` | Add constructor | ~20 |
| `internal/downloader/command.go` | Add constructor | ~20 |

**Total**: ~120 lines of new code across 6 files

### Key Functions

**1. Path Resolution (`bulk.go`)**
```go
func resolveTargetDir(targetDir string) string {
    // Handle empty (use default)
    // Expand ~ to home directory
    // Convert relative to absolute
    // Return resolved path
}
```

**2. Constructor Factory (`bulk.go`)**
```go
func NewBulkDownloaderWithTargetDir(targetDir string) *BulkDownloader {
    resolvedDir := resolveTargetDir(targetDir)
    return &BulkDownloader{
        skillDownloader: NewSkillDownloaderWithBaseDir(resolvedDir),
        // ... other downloaders
    }
}
```

**3. Base Dir Constructors (skill/agent/command.go)**
```go
func NewSkillDownloaderWithBaseDir(baseDir string) *SkillDownloader {
    skillsDir := filepath.Join(baseDir, "skills")
    // Create directory + return downloader
}
```

### Error Handling

| Error Condition | Behavior |
|----------------|----------|
| Empty targetDir | Use default `~/.agents/` |
| Invalid path | Return error with clear message |
| Permission denied | Return error: "Cannot create directory: permission denied" |
| Path is a file | Return error: "Target path exists and is not a directory" |
| Disk full | Return error from OS |

## Testing

### Manual Test Cases

1. **Default behavior** (no flag)
   ```bash
   ./agent-smith install all https://github.com/anthropics/skills
   # Verify: installs to ~/.agents/
   ```

2. **Relative path**
   ```bash
   ./agent-smith install all https://github.com/anthropics/skills --target-dir ./test-local
   # Verify: ./test-local/skills/, ./test-local/agents/, etc. exist
   ```

3. **Absolute path**
   ```bash
   ./agent-smith install all https://github.com/anthropics/skills --target-dir /tmp/test-abs
   # Verify: /tmp/test-abs/skills/ exists
   ```

4. **Tilde expansion**
   ```bash
   ./agent-smith install all https://github.com/anthropics/skills --target-dir ~/test-tilde
   # Verify: expands to /home/user/test-tilde/
   ```

5. **Lock files**
   ```bash
   ./agent-smith install all https://github.com/anthropics/skills -t ./test-locks
   # Verify: ./test-locks/.skill-lock.json exists and contains correct data
   ```

6. **Short form flag**
   ```bash
   ./agent-smith install all https://github.com/anthropics/skills -t ./test-short
   # Verify: works identically to --target-dir
   ```

7. **Path with spaces**
   ```bash
   ./agent-smith install all https://github.com/anthropics/skills --target-dir "./my components"
   # Verify: handles spaces correctly
   ```

8. **Non-existent parent**
   ```bash
   ./agent-smith install all https://github.com/anthropics/skills --target-dir ./a/b/c/components
   # Verify: creates full directory hierarchy
   ```

### Edge Cases

- Empty string for targetDir → should use default `~/.agents/`
- Target exists as file → should error
- No write permissions → should error with clear message
- Symlinks in path → should resolve correctly
- Very long paths → should handle OS limits
- Unicode in path → should handle properly

## Documentation

### Help Text Update

```
NAME:
   agent-smith install all - Download all components from a git repository

USAGE:
   agent-smith install all <repository-url> [--target-dir <path>]

FLAGS:
   --target-dir <path>, -t <path>
       Install to a custom standalone directory instead of ~/.agents/
       
       Custom directories are independent and not managed by link/update/profile
       commands. Useful for project-local components, testing, or distribution.
       
       Supports: relative paths, absolute paths, tilde expansion
       
       Examples:
         --target-dir ./local         # Relative to current directory
         --target-dir /opt/components # Absolute path
         --target-dir ~/myproject     # Expands ~ to home directory

EXAMPLES:
   # Install to default managed directory
   agent-smith install all openai/cookbook

   # Install to project directory
   agent-smith install all openai/cookbook --target-dir ./tools

   # Install for offline distribution
   agent-smith install all company/tools --target-dir ./dist/components

NOTES:
   Custom directories (--target-dir) are standalone installations:
   - NOT managed by 'link', 'update', or 'profile' commands
   - Lock files stored in target directory for reference
   - User manages custom directories manually
   
   For managed installations with full link/update/profile support,
   use default installation (no --target-dir) or profiles.
```

### README Section

Add a new section to README.md:

````markdown
## Custom Target Directories

Install components to custom directories for project-local use or distribution:

```bash
# Install to project directory
./agent-smith install all github.com/org/tools --target-dir ./project-tools

# Result: ./project-tools/skills/, ./project-tools/agents/, ./project-tools/commands/
```

### Important Notes

- Custom directories are **standalone** and independent
- They are **not** managed by `link`, `update`, or `profile` commands
- Lock files stored in target directory for reference
- Use default installation (no flag) for full management features

### Use Cases

- **Project-local components**: Version-control with your project
- **Testing**: Isolate experiments from main installation
- **Distribution**: Package components for offline deployment
- **Development**: Test components during development

### Managed vs Custom Installations

| Feature | Managed (~/.agents/) | Custom (--target-dir) |
|---------|---------------------|----------------------|
| Install | ✓ | ✓ |
| Link to editors | ✓ | ✗ |
| Update tracking | ✓ | ✗ |
| Profile support | ✓ | ✗ |
| Use case | Daily workflow | Project-local, distribution |
````

## Backward Compatibility

✅ **Fully backward compatible**

- Existing commands work exactly as before
- `install all` without flag → installs to `~/.agents/` (current behavior)
- Individual `install skill/agent/command` commands → unchanged
- `--profile` flag for individual installs → unchanged
- Link, update, profile commands → unchanged

## Security Considerations

1. **Path Traversal**: Resolved paths are validated before use
2. **Permissions**: Respect OS file permissions, fail clearly if insufficient
3. **Symlink Following**: Use `filepath.Clean()` to resolve safely
4. **Directory Creation**: Only create directories, never modify existing files
5. **Lock File Security**: Lock files contain only metadata, no sensitive data

## Future Enhancements (Out of Scope)

- Add `--target-dir` to individual install commands
- Support installing to profiles with `install all --profile <name>`
- Registry/discovery system for multiple custom directories
- Integration of custom dirs with link/update commands
- Validation that target directory is a valid agent-smith structure
- `agent-smith list-installations` to show all known installations

## Success Metrics

- Feature is used for project-local installations
- No regression in existing install functionality
- Clear documentation prevents confusion about isolation
- Users report successful use in test/distribution scenarios

## Rollout Plan

1. Implement feature in 6 files (~120 lines)
2. Manual testing with all test cases
3. Update help text and documentation
4. Merge to main branch
5. Include in next release notes

## Open Questions

None - design is complete and approved.

## Approval

- [x] User Requirements Gathered
- [x] Architecture Approach Selected (Isolated Custom Dirs)
- [x] Implementation Plan Defined
- [x] Documentation Planned
- [x] Ready for Implementation
