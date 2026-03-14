# PRD: Unified Component Lock Format

**Created**: 2026-02-05 01:01 UTC

---

## Introduction

Currently, agent-smith maintains separate lock file structures for installations (`ComponentLockEntry` in `~/.agent-smith/.skill-lock.json`) and materializations (`MaterializedComponentMetadata` in `.opencode/.materializations.json`). These structures are ~85% similar but use different field names and lack feature parity. This creates code duplication, makes changes harder to propagate, and prevents beneficial features (like drift detection) from being available everywhere.

This PRD unifies both structures into a single `ComponentEntry` type and uses `.component-lock.json` as the standard filename in both locations, enabling feature parity, code simplification, and future extensibility.

## Goals

- Unify `ComponentLockEntry` and `MaterializedComponentMetadata` into single `ComponentEntry` struct
- Use `.component-lock.json` filename in both `~/.agent-smith/` and project directories (`.opencode/`, `.claude/`, etc.)
- Enable drift detection for both installs and materializations
- Enable filesystem name conflict resolution for both contexts
- Simplify codebase by eliminating duplicate lock file handling code
- Maintain all existing functionality while adding feature parity

## User Stories

- [ ] Story-001: As a developer, I want a unified ComponentEntry struct so that I don't maintain duplicate code for lock files.

  **Acceptance Criteria:**
  - Single `ComponentEntry` struct defined in `internal/models/models.go`
  - Includes all fields from both old structures (installedAt, materializedAt, updatedAt, sourceHash, currentHash, filesystemName, sourceProfile, components, detection)
  - Old `ComponentLockEntry` and `MaterializedComponentMetadata` types removed or marked deprecated
  - ComponentLockFile uses `map[string]map[string]ComponentEntry` structure
  
  **Testing Criteria:**
  **Unit Tests:**
  - Verify ComponentEntry JSON marshaling/unmarshaling with all fields
  - Verify omitempty fields are excluded when empty
  - Verify version field defaults correctly

- [ ] Story-002: As a developer, I want unified lock file paths so that both installs and materializations use .component-lock.json.

  **Acceptance Criteria:**
  - `pkg/paths/paths.go` defines `ComponentLockFile = ".component-lock.json"`
  - Old constants (SkillLockFile, AgentLockFile, CommandLockFile) removed
  - `GetComponentLockPath()` returns `.component-lock.json` path for given base directory
  - Works for both `~/.agent-smith/` and project directories
  
  **Testing Criteria:**
  **Unit Tests:**
  - Verify GetComponentLockPath returns correct path for home directory
  - Verify GetComponentLockPath returns correct path for project directories

- [ ] Story-003: As a developer, I want unified save functions so that both installs and materializations write to the same format.

  **Acceptance Criteria:**
  - `SaveComponentEntry()` function in `internal/metadata/lock.go` handles both install and materialize cases
  - Sets `installedAt` and `updatedAt` for install operations
  - Sets `materializedAt` for materialize operations
  - Calculates and stores `sourceHash` and `currentHash` for both operations
  - Accepts `filesystemName`, `sourceProfile`, `components`, `detection` as optional parameters
  - Writes to `.component-lock.json` with version 5
  
  **Testing Criteria:**
  **Unit Tests:**
  - Verify SaveComponentEntry creates new lock file with correct structure
  - Verify SaveComponentEntry updates existing entries preserving installedAt
  - Verify SaveComponentEntry populates install-specific fields correctly
  - Verify SaveComponentEntry populates materialize-specific fields correctly
  - Verify sourceHash and currentHash are calculated correctly
  
  **Integration Tests:**
  - Verify concurrent writes to same lock file don't corrupt data

- [ ] Story-004: As a developer, I want unified load functions so that reading lock files works consistently.

  **Acceptance Criteria:**
  - `LoadComponentEntry()` function loads from `.component-lock.json`
  - `LoadComponentEntryBySource()` function loads specific source entries
  - Functions work for both home directory and project directory lock files
  - Returns unified `ComponentEntry` struct
  - Handles missing files gracefully (returns empty structure)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Verify LoadComponentEntry reads existing lock file correctly
  - Verify LoadComponentEntry handles missing lock file
  - Verify LoadComponentEntryBySource filters by source URL correctly
  - Verify LoadComponentEntry handles multiple sources correctly

- [ ] Story-005: As a developer, I want unified remove functions so that uninstall and unmaterialize use the same code.

  **Acceptance Criteria:**
  - `RemoveComponentEntry()` function removes from `.component-lock.json`
  - `RemoveComponentEntryBySource()` function removes specific source entries
  - Handles removing last component from a source (cleans up empty source map)
  - Works for both home directory and project directory lock files
  
  **Testing Criteria:**
  **Unit Tests:**
  - Verify RemoveComponentEntry removes entry correctly
  - Verify RemoveComponentEntry cleans up empty source maps
  - Verify RemoveComponentEntryBySource removes only specified source

- [ ] Story-006: As a developer, I want materialization functions to use the unified format so that project lock files match install lock files.

  **Acceptance Criteria:**
  - `pkg/project/materialization.go` functions updated to use `ComponentEntry`
  - `SaveMaterializationMetadata()` writes to `.component-lock.json`
  - `LoadMaterializationMetadata()` reads from `.component-lock.json`
  - `AddMaterializationEntry()` creates `ComponentEntry` with materialize-specific fields
  - Sets `materializedAt`, `sourceHash`, `currentHash`, `filesystemName`, `sourceProfile`
  - Existing functionality (drift detection, filesystem naming) preserved
  
  **Testing Criteria:**
  **Unit Tests:**
  - Verify SaveMaterializationMetadata writes correct format
  - Verify LoadMaterializationMetadata reads unified format
  - Verify AddMaterializationEntry populates all required fields

- [ ] Story-007: As a developer, I want install operations to support drift detection so that I can detect local modifications.

  **Acceptance Criteria:**
  - Install operations calculate `sourceHash` from source directory at install time
  - Install operations calculate `currentHash` from installed directory after copy
  - Hash calculation uses same algorithm as materialization (directory content hash)
  - Hashes stored in `.component-lock.json` for installed components
  
  **Testing Criteria:**
  **Unit Tests:**
  - Verify sourceHash calculated correctly from source directory
  - Verify currentHash calculated correctly from installed directory
  - Verify hash calculation is deterministic and consistent

- [ ] Story-008: As a developer, I want install operations to support filesystem name tracking so that conflict resolution works everywhere.

  **Acceptance Criteria:**
  - Install operations populate `filesystemName` field with actual directory name
  - Uses same conflict resolution logic as materialization (adds -2, -3 suffixes)
  - `ResolveFilesystemName()` function works for both install and materialize contexts
  - Filesystem name checked for uniqueness within component type directory
  
  **Testing Criteria:**
  **Unit Tests:**
  - Verify ResolveFilesystemName returns base name when no conflict
  - Verify ResolveFilesystemName returns suffixed name when conflict exists
  - Verify ResolveFilesystemName reuses existing name for same component+source

- [ ] Story-009: As a developer, I want all callers updated to use the unified format so that the codebase is consistent.

  **Acceptance Criteria:**
  - All references to old `ComponentLockEntry` updated to `ComponentEntry`
  - All references to old `MaterializedComponentMetadata` updated to `ComponentEntry`
  - All code using `.skill-lock.json`, `.agent-lock.json`, `.command-lock.json` updated to use `.component-lock.json`
  - All code using `.materializations.json` updated to use `.component-lock.json`
  - Install operations in `internal/installer/` use unified functions
  - Materialize operations in `pkg/services/materialize/` use unified functions
  - Update operations in `internal/updater/` use unified functions
  - Uninstall operations in `internal/uninstaller/` use unified functions
  
  **Testing Criteria:**
  **Integration Tests:**
  - Verify install command writes to `.component-lock.json`
  - Verify materialize command writes to `.component-lock.json`
  - Verify uninstall command removes from `.component-lock.json`
  - Verify update command reads from `.component-lock.json`
  - Verify list commands read from `.component-lock.json`

- [ ] Story-010: As a developer, I want comprehensive integration tests so that the unified format works end-to-end.

  **Acceptance Criteria:**
  - Test install → materialize workflow with unified format
  - Test drift detection works for both installs and materializations
  - Test filesystem name conflict resolution in both contexts
  - Test uninstall removes correct entries from unified lock file
  - Test multiple sources with same component name work correctly
  - Test concurrent operations don't corrupt lock files
  
  **Testing Criteria:**
  **Integration Tests:**
  - Install component, verify .component-lock.json created correctly
  - Materialize component, verify project .component-lock.json created correctly
  - Modify installed component, verify drift detected via hash comparison
  - Install two components with same name from different sources, verify both tracked
  - Uninstall one source of component, verify other source remains

## Functional Requirements

- FR-1: The system SHALL use a single `ComponentEntry` struct for both installs and materializations
- FR-2: The system SHALL write all lock files to `.component-lock.json` in the appropriate directory
- FR-3: The system SHALL use version 5 for the unified lock file format
- FR-4: The system SHALL calculate `sourceHash` and `currentHash` for all install operations
- FR-5: The system SHALL calculate `sourceHash` and `currentHash` for all materialize operations
- FR-6: The system SHALL track `filesystemName` for both installs and materializations
- FR-7: The system SHALL preserve all existing functionality (drift detection, conflict resolution, profile tracking)
- FR-8: The system SHALL maintain backward compatibility by gracefully handling missing lock files
- FR-9: The system SHALL handle concurrent writes to lock files without data corruption
- FR-10: The system SHALL remove empty source maps when last component from a source is removed

## Non-Goals

- No automatic migration of existing lock files (users start fresh)
- No backward compatibility for reading old `.skill-lock.json`, `.agent-lock.json`, `.command-lock.json` files
- No backward compatibility for reading old `.materializations.json` files
- No changes to component detection or installation logic (only lock file format)
- No changes to user-facing command interface or output
- No performance optimizations beyond what unified code naturally provides
