# PRD: Profile Component Copy Feature

**Created**: 2026-02-01 00:52 UTC

---

## Introduction

Add functionality to copy components (skills, agents, commands) between profiles while preserving Git source metadata, enabling independent updates in both source and target profiles. This addresses a critical bug in the existing `profile add` command where lock file entries are not copied, breaking update functionality for Git-sourced components.

## Goals

- Enable easy copying of single components between profiles
- Preserve Git source metadata (lock file entries) during copy operations
- Fix the existing bug in `profile add` where lock file entries are not copied
- Maintain independent updateability for both source and target components
- Provide clear error messages and safe failure modes

## User Stories

- [x] Story-001: As a developer, I want to copy a useful skill from my work profile to my personal profile so that I can use the same tool in both contexts and update them independently.

  **Acceptance Criteria:**
  - Command accepts four arguments: component type, source profile, target profile, component name
  - Component directory is copied from source profile to target profile
  - Lock file entry is copied from source to target profile lock file
  - Success message displays component details and location
  - Both profiles can update the component independently after copy
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile name validation tests (valid/invalid patterns)
  - Component type validation tests (skills/agents/commands)
  - Component existence verification in source profile
  - Target component conflict detection tests
  
  **Integration Tests:**
  - End-to-end copy operation with Git-sourced component
  - Lock file entry preservation and readability
  - Independent update capability after copy
  - Command execution through CLI interface
  
  **Component Browser Tests:**
  - Not applicable (CLI tool, no browser UI)

- [x] Story-002: As a developer, I want clear error messages when copy operations fail so that I can understand what went wrong and how to fix it.

  **Acceptance Criteria:**
  - Error message when source profile doesn't exist shows profile name
  - Error message when target profile doesn't exist shows profile name
  - Error message when component not in source lists available components
  - Error message when component exists in target suggests workaround
  - Error message when component type invalid lists valid types
  - All error messages include example usage
  
  **Testing Criteria:**
  **Unit Tests:**
  - Error message formatting for each error condition
  - Error message content verification
  - Example usage string generation
  
  **Integration Tests:**
  - CLI error output for non-existent source profile
  - CLI error output for non-existent target profile
  - CLI error output for missing component
  - CLI error output for existing component in target
  - CLI error output for invalid component type
  
  **Component Browser Tests:**
  - Not applicable (CLI tool, no browser UI)

- [ ] Story-003: As a developer, I want to create a new profile by cherry-picking components from existing profiles so that I can create specialized toolsets for different projects.

  **Acceptance Criteria:**
  - Can copy multiple components sequentially to build new profile
  - Each copy operation is independent and atomic
  - Failed copy doesn't affect previous successful copies
  - Can copy from multiple source profiles to single target
  - Profile remains valid after multiple copy operations
  
  **Testing Criteria:**
  **Unit Tests:**
  - Multiple sequential copy operations
  - Atomic copy operation verification
  - Profile validation after multiple copies
  
  **Integration Tests:**
  - Create new profile and copy multiple components
  - Copy from different source profiles to same target
  - Verify all copied components have valid lock entries
  - Update all copied components independently
  
  **Component Browser Tests:**
  - Not applicable (CLI tool, no browser UI)

- [ ] Story-004: As a developer, I want the existing `profile add` command to preserve lock file entries so that components added from base installation remain updateable.

  **Acceptance Criteria:**
  - `profile add` command uses same lock file copying logic as `profile copy`
  - Components added via `profile add` have lock file entries in profile
  - Components added via `profile add` can be updated via `agent-smith update`
  - No breaking changes to existing `profile add` API
  - Existing test suite passes with new implementation
  
  **Testing Criteria:**
  **Unit Tests:**
  - Lock file entry copying in AddComponentToProfile method
  - Profile add operation preserves all metadata fields
  - Backward compatibility with manual components (no lock entry)
  
  **Integration Tests:**
  - Add component from base installation to profile
  - Verify lock file entry exists in profile lock file
  - Update component in profile successfully
  - Regression tests for existing profile add workflows
  
  **Component Browser Tests:**
  - Not applicable (CLI tool, no browser UI)

- [ ] Story-005: As a developer, I want to test new component versions in an experimental profile while keeping stable versions in production so that I can safely evaluate updates.

  **Acceptance Criteria:**
  - Can copy component from production profile to experimental profile
  - Update component in experimental profile independently
  - Production profile component remains at original version
  - Can verify different versions in each profile
  - Can update production profile when experimental proves stable
  
  **Testing Criteria:**
  **Unit Tests:**
  - Independent lock file management per profile
  - Version tracking in separate lock files
  
  **Integration Tests:**
  - Copy component between profiles with different versions
  - Update component in one profile, verify other unchanged
  - Verify commit hash differs between profiles after update
  - Full workflow test: copy, update experimental, update production
  
  **Component Browser Tests:**
  - Not applicable (CLI tool, no browser UI)

## Functional Requirements

- FR-1: The system SHALL implement `agent-smith profile copy <type> <source-profile> <target-profile> <component-name>` command
- FR-2: The system SHALL validate profile names match regex `^[a-zA-Z0-9-]+$`
- FR-3: The system SHALL validate component type is one of: skills, agents, commands
- FR-4: The system SHALL verify source profile exists and is valid
- FR-5: The system SHALL verify target profile exists and is valid
- FR-6: The system SHALL verify component exists in source profile
- FR-7: The system SHALL verify component does not exist in target profile
- FR-8: The system SHALL copy component directory from source to target
- FR-9: The system SHALL copy lock file entry from source to target lock file
- FR-10: The system SHALL preserve all lock file metadata fields (commitHash, sourceUrl, installedAt, etc.)
- FR-11: The system SHALL handle missing lock file entries gracefully (manually created components)
- FR-12: The system SHALL display success message with component details
- FR-13: The system SHALL abort operation and display clear error on any validation failure
- FR-14: The system SHALL update `AddComponentToProfile` to use lock file copying logic
- FR-15: The system SHALL enable independent updates for copied components in both profiles

## Non-Goals

- Bulk copy operations (copy all components at once)
- Copy with conflict resolution (overwrite existing components with --force flag)
- Copy from `~/.agent-smith/` base to profiles (already handled by `profile add`)
- Move operations (copy + delete from source)
- Interactive component selection mode
- Dry-run preview mode
- Component version comparison between profiles
- Automatic conflict resolution strategies

## Technical Implementation

### Architecture Changes

**New Method**: `copyComponentWithMetadata` in `pkg/profiles/manager.go`
```go
func (pm *ProfileManager) copyComponentWithMetadata(
    sourceBaseDir, targetBaseDir, componentType, componentName string,
) error
```

**New Method**: `CopyComponentBetweenProfiles` in `pkg/profiles/manager.go`
```go
func (pm *ProfileManager) CopyComponentBetweenProfiles(
    sourceProfile, targetProfile, componentType, componentName string,
) error
```

**Modified Method**: `AddComponentToProfile` in `pkg/profiles/manager.go`
- Replace manual file copying with `copyComponentWithMetadata` helper

**New CLI Command**: `profile copy` in `cmd/root.go`
- Add as subcommand under `profilesCmd`

**New Handler**: `handleProfilesCopy` in `cmd/profiles.go`

### Lock File Handling

Lock file entries contain critical Git source metadata:
```json
{
  "source": "https://github.com/anthropics/skills",
  "sourceType": "github",
  "sourceUrl": "https://github.com/anthropics/skills",
  "originalPath": "skills/api-design/SKILL.md",
  "commitHash": "69c0b1a0674149f27b61b2635f935524b6add202",
  "installedAt": "2026-01-31T18:24:22-06:00",
  "updatedAt": "2026-01-31T18:24:22-06:00",
  "version": 3,
  "components": 1,
  "detection": "single"
}
```

**Copy Process**:
1. Read source profile lock file using `metadata.LoadLockFileEntry`
2. Extract entry for component name
3. Read target profile lock file (create if missing)
4. Write entry to target lock file using `metadata.SaveLockFileEntry`
5. Preserve all metadata fields exactly as-is

### Error Handling

| Error Condition | Behavior |
|----------------|----------|
| Invalid profile name | Abort with validation error |
| Invalid component type | Abort with valid types list |
| Source profile missing | Abort with profile not found error |
| Target profile missing | Abort with profile not found error |
| Component not in source | Abort with component not found error |
| Component exists in target | Abort with conflict error |
| Lock file read failure | Continue with warning (manual component) |
| Lock file write failure | Warning only, operation continues |
| File copy failure | Abort, cleanup partial copy |

### Command Usage

```bash
# Basic usage
agent-smith profile copy skills work-profile personal-profile api-design

# Copy agent between profiles
agent-smith profile copy agents team-profile solo-profile code-reviewer

# Copy command between profiles
agent-smith profile copy commands dev-profile prod-profile test-runner

# Error examples
agent-smith profile copy plugins work personal tool  # Invalid type
agent-smith profile copy skills work personal api-design  # Already exists
agent-smith profile copy skills missing personal api-design  # Profile not found
```

### Success Message Format

```
Copying component...
✓ Successfully copied skill 'api-design' from 'work-profile' to 'personal-profile'

Component details:
  Type: skills
  Name: api-design
  Source: https://github.com/example/skills
  Location: ~/.agent-smith/profiles/personal-profile/skills/api-design

Both profiles can now update this component independently.
```

## Dependencies

**Existing Code**:
- `pkg/profiles/manager.go` - Profile management, validation, file operations
- `internal/metadata/lock.go` - Lock file read/write operations
- `cmd/root.go` - CLI command definitions using Cobra
- `cmd/profiles.go` - Profile command handlers
- `internal/updater/updater.go` - Update logic that requires lock files

**External Libraries**:
- `github.com/spf13/cobra` - CLI framework
- Standard Go libraries: `os`, `path/filepath`, `encoding/json`

**No New Dependencies Required**

## Risks & Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Lock file corruption | High | Low | Validate JSON before write, atomic operations |
| Partial copy on failure | Medium | Low | Cleanup on error, transactional copy |
| Breaking `profile add` | High | Low | Comprehensive regression testing |
| Lock format changes | Medium | Low | Version lock format, graceful handling |
| Data loss during copy | High | Very Low | Never modify source, only copy |

## Success Metrics

**Functional**:
- User can copy components between profiles successfully
- Lock file entries preserved with 100% accuracy
- Update works independently in both profiles
- `profile add` bug fixed with no regressions

**Performance**:
- Copy operation completes in < 5 seconds for typical components
- Lock file operations add < 100ms overhead

**Quality**:
- Test coverage > 85%
- Zero P0/P1 bugs after release
- Update success rate matches original component update rate

## Implementation Phases

**Phase 1: Core Functionality**
- Add `copyComponentWithMetadata` helper
- Add `CopyComponentBetweenProfiles` method
- Add `profile copy` CLI command
- Add `handleProfilesCopy` handler
- Unit tests for core methods

**Phase 2: Bug Fix**
- Update `AddComponentToProfile` implementation
- Regression tests for `profile add`
- Verify existing workflows unaffected

**Phase 3: Integration & Testing**
- End-to-end integration tests
- Error scenario testing
- Update independence verification
- CLI help text and examples

**Phase 4: Documentation**
- Update command help text
- Add usage examples
- Update README if needed

## Future Enhancements (Out of Scope)

- Add `--force` flag to overwrite existing components
- Add `profile copy-all` for bulk operations
- Add interactive component selection mode
- Add `--dry-run` flag for preview
- Add component version comparison tools
- Add profile merge capabilities

---

## Appendix: Related Documentation

**Related Files**:
- `pkg/profiles/manager.go` - Profile management implementation
- `internal/metadata/lock.go` - Lock file format specification
- `internal/updater/updater.go` - Update logic requiring lock files

**Related Issues**:
- Bug: `profile add` doesn't copy lock file entries
- Enhancement: Easy component sharing between profiles
