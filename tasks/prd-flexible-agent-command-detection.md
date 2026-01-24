# PRD: Flexible Agent & Command Detection System

## Introduction

Implement a flexible, frontmatter-based detection system for AI agents and commands that can handle various repository structures. The system will scan repositories recursively for `/agents/` and `/commands/` directories, parse YAML frontmatter for metadata, and store components in a flat structure that can be symlinked to OpenCode's configuration directory.

## Goals

- Support multiple repository formats (plugins, flat, nested structures)
- Parse YAML frontmatter for component metadata (name, description, model, etc.)
- Store agents/commands in flat directory structure (`~/.agents/agents/`, `~/.agents/commands/`)
- Copy entire component directories to preserve context and resources
- Provide clear warnings for duplicate component names
- Remove dependency on marker files (`AGENT.md`, `COMMAND.md`)
- Maintain backward compatibility with existing skills storage approach

## User Stories

- [x] Story-001: As a developer, I want to download agents from repositories with various structures so that I can use agents regardless of how they're organized.

  **Acceptance Criteria:**
  - System detects any `/agents/` directory anywhere in repository tree
  - System detects any `/commands/` directory anywhere in repository tree
  - Detection works for flat, nested, and plugin-based structures
  - Ignores test/example directories (test, tests, examples, docs)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Directory pattern matching validation
  - Path ignore logic verification
  - Detection logic for various structures
  
  **Integration Tests:**
  - End-to-end detection with sample repositories
  - Multiple structure format validation
  - Ignore path effectiveness verification

- [x] Story-002: As a developer, I want agents/commands to use YAML frontmatter for metadata so that I can specify custom names and configuration.

  **Acceptance Criteria:**
  - Parse YAML frontmatter delimited by `---`
  - Support `name`, `description`, `model`, `mode` fields
  - Gracefully handle missing or malformed frontmatter
  - Log warnings for YAML parsing errors
  - Forward-compatible with unknown frontmatter fields
  
  **Testing Criteria:**
  **Unit Tests:**
  - YAML parser with valid frontmatter
  - YAML parser with malformed frontmatter
  - YAML parser with no frontmatter
  - YAML parser with extra unknown fields
  
  **Integration Tests:**
  - End-to-end with various frontmatter formats
  - Fallback behavior validation

- [x] Story-003: As a developer, I want component names determined by priority (frontmatter > filename) so that I have flexibility in naming.

  **Acceptance Criteria:**
  - Use frontmatter `name` field if present
  - Fall back to filename (without .md extension) if no frontmatter name
  - Skip special files (README.md, index.md, main.md)
  - Handle edge cases (no extension, empty name)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Name determination with frontmatter name
  - Name determination without frontmatter
  - Special file name handling
  - Edge case validation
  
  **Integration Tests:**
  - End-to-end naming validation
  - Component uniqueness verification

- [x] Story-004: As a developer, I want agents/commands stored in flat directory structure so that linking to OpenCode is straightforward.

  **Acceptance Criteria:**
  - Store agents in `~/.agents/agents/{name}/{name}.md`
  - Store commands in `~/.agents/commands/{name}/{name}.md`
  - Copy entire source directory to preserve resources
  - Create parent directories if they don't exist
  
  **Testing Criteria:**
  **Unit Tests:**
  - Directory creation logic
  - Path construction validation
  
  **Integration Tests:**
  - End-to-end storage with actual repositories
  - Directory structure verification
  - Resource preservation validation

- [x] Story-005: As a developer, I want clear warnings when duplicate component names are detected so that I can resolve conflicts.

  **Acceptance Criteria:**
  - Detect duplicate names during detection phase
  - Log warning with list of conflicting file paths
  - Use only the first occurrence
  - Provide actionable suggestions (rename, frontmatter name)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Duplicate detection logic
  - Warning message formatting
  
  **Integration Tests:**
  - End-to-end with duplicate components
  - First occurrence selection verification

- [x] Story-006: As a developer, I want the old AGENT.md/COMMAND.md marker detection removed so that the codebase is simpler.

  **Acceptance Criteria:**
  - Remove `ExactFiles` from agent detection config
  - Remove `ExactFiles` from command detection config
  - Remove hardcoded fallback detection logic (lines 618-666)
  - Skills still use SKILL.md marker (unchanged)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Verify marker files no longer trigger detection
  - Skills marker detection still works
  
  **Integration Tests:**
  - End-to-end without marker files
  - Backward compatibility with skills

- [x] Story-007: As a developer, I want components downloaded by copying entire directories so that support files and resources are preserved.

  **Acceptance Criteria:**
  - Copy entire parent directory containing component file
  - Preserve directory structure within component directory
  - Include all files (markdown, templates, references, etc.)
  - Use existing `copyDirectoryContents` function
  
  **Testing Criteria:**
  **Unit Tests:**
  - Directory copy logic validation
  
  **Integration Tests:**
  - End-to-end with complex component directories
  - Resource file preservation verification
  - Nested directory handling

- [x] Story-008: As a developer, I want the update-all command to re-clone repositories so that updates are reliable and simple.

  **Acceptance Criteria:**
  - Update command removes old component directory
  - Update command re-clones source repository
  - Update command re-runs detection and download
  - Update command preserves lock file entries
  
  **Testing Criteria:**
  **Unit Tests:**
  - Update logic flow validation
  
  **Integration Tests:**
  - End-to-end update with modified repository
  - Lock file preservation verification
  - Component replacement validation

## Functional Requirements

### Detection

- FR-1: The system must walk the entire repository tree to find agents and commands
- FR-2: The system must detect any directory named `agents` or `agents/` in the path
- FR-3: The system must detect any directory named `commands` or `commands/` in the path
- FR-4: The system must ignore directories: .git, node_modules, .vscode, .idea, target, build, dist, test, tests, __tests__, examples, example, docs, .next, .nuxt, coverage
- FR-5: The system must process only `.md` files within agents/commands directories
- FR-6: The system must skip directories when encountering ignore patterns

### Frontmatter Parsing

- FR-7: The system must parse YAML frontmatter delimited by `---\n` at file start
- FR-8: The system must handle files without frontmatter (return nil, no error)
- FR-9: The system must log warnings for malformed YAML and continue processing
- FR-10: The system must extract `name`, `description`, `model`, `mode` fields from frontmatter
- FR-11: The system must ignore unknown frontmatter fields for forward compatibility

### Component Naming

- FR-12: The system must use frontmatter `name` field if present
- FR-13: The system must fall back to filename (minus .md extension) if no frontmatter name
- FR-14: The system must skip files named README.md, index.md, main.md
- FR-15: The system must handle edge cases (no extension, empty names) gracefully

### Storage

- FR-16: The system must store agents in `~/.agents/agents/{component-name}/`
- FR-17: The system must store commands in `~/.agents/commands/{component-name}/`
- FR-18: The system must copy the entire source directory to destination
- FR-19: The system must create parent directories if they don't exist
- FR-20: The system must use component name for directory name (kebab-case or as-is)

### Duplicate Handling

- FR-21: The system must track all components by name during detection
- FR-22: The system must detect when multiple components share the same name
- FR-23: The system must log warning with full paths of all duplicates
- FR-24: The system must use only the first detected occurrence
- FR-25: The system must provide actionable suggestions in warnings

### Linking

- FR-26: The system must create symlinks from `~/.config/opencode/agents/{name}` to `~/.agents/agents/{name}/{name}.md`
- FR-27: The system must create symlinks from `~/.config/opencode/commands/{name}` to `~/.agents/commands/{name}/{name}.md`
- FR-28: The system must handle existing symlinks gracefully (recreate or skip)

### Lock File

- FR-29: The system must update `.agent-lock.json` with agent entries
- FR-30: The system must update `.command-lock.json` with command entries
- FR-31: The system must include `componentPath` field tracking source location
- FR-32: The system must maintain version 3 lock file format

## Technical Specifications

### Data Structures

```go
// ComponentFrontmatter represents YAML frontmatter
type ComponentFrontmatter struct {
    Name        string `yaml:"name"`
    Description string `yaml:"description"`
    Model       string `yaml:"model"`
    Mode        string `yaml:"mode"`
}

// DetectedComponent already exists
type DetectedComponent struct {
    Type       ComponentType
    Name       string
    Path       string
    SourceFile string
}
```

### Key Functions

```go
// New functions to implement
func parseFrontmatter(filePath string) (*ComponentFrontmatter, error)
func determineComponentName(frontmatter *ComponentFrontmatter, fileName string) string
func (rd *RepositoryDetector) detectAgent(filePath, repoPath string) *DetectedComponent
func (rd *RepositoryDetector) detectCommand(filePath, repoPath string) *DetectedComponent
func (rd *RepositoryDetector) shouldIgnoreDir(dirPath string) bool
func (rd *RepositoryDetector) detectSkills(repoPath string) []DetectedComponent

// Modified functions
func (rd *RepositoryDetector) detectComponentsInRepo(repoPath string) ([]DetectedComponent, error)
func createDefaultDetectionConfig() *DetectionConfig
```

### Dependencies

- Add `gopkg.in/yaml.v3` to `go.mod`
- Import yaml package in `main.go`

### File Changes Summary

| File | Change | Description |
|------|--------|-------------|
| `go.mod` | Add dependency | `gopkg.in/yaml.v3 v3.0.1` |
| `main.go` | Add import | YAML parser import |
| `main.go` | Add struct | `ComponentFrontmatter` |
| `main.go` | Add functions | `parseFrontmatter()`, `determineComponentName()` |
| `main.go` | Add functions | `detectAgent()`, `detectCommand()`, `shouldIgnoreDir()` |
| `main.go` | Add function | `detectSkills()` |
| `main.go` | Modify | `createDefaultDetectionConfig()` - remove ExactFiles |
| `main.go` | Replace | `detectComponentsInRepo()` - new detection logic |
| `main.go` | Modify | Download functions - copy entire directory |

## Non-Goals (Out of Scope)

- No plugin-level grouping or tracking
- No plugin.json metadata parsing
- No namespace prefix for duplicate names (just warn)
- No .agentignore file support
- No size limits on component directories
- No standalone component download (individual file download)
- No automatic migration of existing components
- No validation of frontmatter field values
- No type field requirement in frontmatter (rely on directory location)

## Implementation Phases

### Phase 1: Dependencies & Data Structures
- Add YAML dependency to go.mod
- Add ComponentFrontmatter struct
- Add import statement

### Phase 2: Frontmatter Parsing
- Implement parseFrontmatter() function
- Implement determineComponentName() function
- Add error handling and warnings

### Phase 3: Detection Logic
- Implement detectAgent() function
- Implement detectCommand() function
- Implement shouldIgnoreDir() function
- Implement detectSkills() function

### Phase 4: Update Detection Config
- Remove ExactFiles from agent detection
- Remove ExactFiles from command detection
- Add more ignore paths

### Phase 5: Rewrite Main Detection
- Replace detectComponentsInRepo() logic
- Add duplicate detection and warnings
- Integrate new detection functions

### Phase 6: Update Download Logic
- Modify agent downloader to copy directories
- Modify command downloader to copy directories
- Ensure existing skills logic unchanged

### Phase 7: Testing
- Test with wshobson/agents repository
- Test with flat structure repositories
- Test with nested structure repositories
- Test duplicate detection
- Test frontmatter parsing edge cases
- Test directory copying

### Phase 8: Documentation
- Update README with frontmatter format
- Create example repository structures
- Document migration path for users

## Success Criteria

- Successfully downloads agents from wshobson/agents repository (74 plugins)
- Correctly detects agents/commands in various repository structures
- Parses YAML frontmatter and uses names appropriately
- Stores components in flat directory structure
- Copies entire component directories with resources
- Warns appropriately for duplicate names
- Links components correctly to OpenCode directories
- Maintains backward compatibility with skills
- All tests pass successfully

## Testing Strategy

### Unit Tests
- YAML frontmatter parsing (valid, invalid, missing)
- Component name determination logic
- Directory ignore logic
- Duplicate detection logic
- Path construction validation

### Integration Tests
- End-to-end with wshobson/agents repo
- End-to-end with flat structure repo
- End-to-end with nested structure repo
- Update-all command flow
- Link-all command flow

### Manual Testing
```bash
# Build
go build -o agent-smith main.go

# Test detection
./agent-smith add-all wshobson/agents

# Verify storage
ls -la ~/.agents/agents/
ls -la ~/.agents/commands/

# Test linking
./agent-smith link-all

# Verify links
ls -la ~/.config/opencode/agents/
ls -la ~/.config/opencode/commands/

# Test updates
./agent-smith update-all
```

## Risk Assessment

### High Risk
- Breaking changes to existing detection logic
- Repository structure variations not covered

### Medium Risk
- YAML parsing errors with edge cases
- Performance with very large repositories
- Duplicate name handling user confusion

### Low Risk
- Directory copying edge cases
- Lock file format changes

### Mitigation Strategies
- Comprehensive testing with real-world repositories
- Clear error messages and warnings
- Graceful fallbacks for parsing errors
- Preserve existing skills logic unchanged
