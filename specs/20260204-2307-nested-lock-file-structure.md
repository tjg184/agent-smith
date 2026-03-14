# PRD: Nested Lock File and Materialization Structure to Prevent Name Conflicts

**Created**: 2026-02-04 23:07 UTC

---

## Introduction

The current lock file and materialization metadata structures use flat maps keyed only by component name (e.g., `skills["python-helper"]`). This creates critical name collision issues when:

1. **Lock files**: Installing two components with the same name from different sources causes one to overwrite the other
2. **Materialization files**: Materializing components with the same name from different profiles to the same target causes the second to overwrite the first

This PRD addresses these conflicts by restructuring both lock files and materialization metadata to use a nested structure grouped by source URL, ensuring each component installation is uniquely tracked regardless of name collisions.

## Goals

- Eliminate name collision issues in lock files when installing same-named components from different sources
- Eliminate name collision issues in materialization metadata when materializing same-named components from different profiles
- Support multiple components with identical names from different sources coexisting in the system
- Provide clear disambiguation when users reference ambiguous component names
- Implement auto-suffixing for filesystem names when conflicts occur during materialization
- Maintain data integrity and traceability for all component installations

## User Stories

- [ ] Story-001: As a user, I want to install skills with the same name from different repositories so that I can use specialized implementations without conflicts.

  **Acceptance Criteria:**
  - Lock file nests components by source URL (e.g., `skills["https://github.com/team-a/tools"]["python-helper"]`)
  - Installing `python-helper` from `team-a/tools` and `team-b/tools` creates two separate entries
  - Both installations are tracked independently with distinct commit hashes and metadata
  - Lock file version bumped to v4 to reflect new structure
  
  **Testing Criteria:**
  **Unit Tests:**
  - Lock file save/load with nested structure
  - Component entry insertion at correct nesting level
  - Lookup by source URL and component name
  
  **Integration Tests:**
  - Install same-named skills from two different sources
  - Verify both entries exist in lock file with correct structure

- [ ] Story-002: As a user, I want to materialize components with the same name from different profiles to the same target so that I can use multiple versions in my project.

  **Acceptance Criteria:**
  - Materialization metadata nests components by source URL
  - Materializing `python-helper` from profile `team-a` creates entry at `skills["https://github.com/team-a/tools"]["python-helper"]`
  - Materializing same-named component from different profile/source creates separate entry
  - Materialization metadata version bumped to v2
  - New field `filesystemName` tracks actual directory name on disk
  
  **Testing Criteria:**
  **Unit Tests:**
  - Materialization metadata save/load with nested structure
  - Filesystem name resolution logic
  
  **Integration Tests:**
  - Materialize same-named skills from two profiles to same target
  - Verify both entries tracked in metadata
  - Verify both directories exist on disk

- [ ] Story-003: As a user, when I materialize a component that conflicts with an existing filesystem name, I want the system to automatically suffix the new component so both can coexist on disk.

  **Acceptance Criteria:**
  - When materializing `python-helper` and name already exists on disk, use `python-helper-2`
  - Subsequent conflicts use `python-helper-3`, `python-helper-4`, etc.
  - `filesystemName` field in metadata stores actual disk name
  - User is informed which filesystem name was used
  - Display format: `Materialized skill 'python-helper' as 'python-helper-2'`
  
  **Testing Criteria:**
  **Unit Tests:**
  - Auto-suffix logic finds next available name
  - Suffix numbering increments correctly
  
  **Integration Tests:**
  - Materialize 3+ components with same name, verify suffixes
  - Verify each has correct filesystemName in metadata

- [ ] Story-004: As a user, when I reference an ambiguous component name in a command, I want clear error messages guiding me how to disambiguate so I can specify which version I meant.

  **Acceptance Criteria:**
  - Commands that accept component names check for ambiguity across all sources
  - When multiple sources have same component name, display error with all matching sources
  - Error message includes exact command syntax with `--source` flag
  - Error message shows source URLs for all matches
  - Example: `Multiple sources found for skill 'python-helper': 1) https://github.com/team-a/tools, 2) https://github.com/team-b/tools. Use --source <url> to specify.`
  
  **Testing Criteria:**
  **Unit Tests:**
  - Component lookup returns all matching sources
  - Disambiguation error message formatting
  
  **Integration Tests:**
  - Attempt to update/remove ambiguous component name without --source flag
  - Verify clear error message with all sources listed

- [ ] Story-005: As a user, I want to use the `--source` flag to disambiguate component references so I can operate on the specific version I intend.

  **Acceptance Criteria:**
  - Add `--source <url>` flag to all component commands (update, remove, materialize)
  - Flag accepts full source URL (e.g., `https://github.com/team-a/tools`)
  - Flag accepts short form (e.g., `team-a/tools`)
  - When provided, lookup uses source URL + component name
  - Commands operate on single matching component
  - Example: `agent-smith update skill python-helper --source team-a/tools`
  
  **Testing Criteria:**
  **Unit Tests:**
  - Source URL normalization (short form to full URL)
  - Lookup with source filter
  
  **Integration Tests:**
  - Update specific component using --source flag
  - Materialize specific component using --source flag
  - Verify correct component operated on

- [ ] Story-006: As a user listing materialized components, I want to see source information for each component so I can distinguish between same-named components from different sources.

  **Acceptance Criteria:**
  - `materialize list` displays source URL or profile name for each component
  - Components with same name clearly differentiated by source
  - Display includes filesystem name if different from component name
  - Format: `python-helper (team-a → https://github.com/team-a/tools) as python-helper-2`
  
  **Testing Criteria:**
  **Unit Tests:**
  - Display formatting for component with source info
  
  **Integration Tests:**
  - Materialize multiple same-named components, list them
  - Verify source info displayed for each

- [ ] Story-007: As a developer, I want all lock file read/write operations to use the new nested structure so that the v4 format is consistently applied throughout the codebase.

  **Acceptance Criteria:**
  - `SaveLockFileEntry()` saves to nested structure `map[sourceURL]map[componentName]`
  - `LoadLockFileEntry()` searches nested structure with source URL + component name
  - `RemoveLockFileEntry()` removes from nested structure
  - `GetAllComponentNames()` aggregates names from all sources with duplicate tracking
  - All operations write lock file as version 4
  
  **Testing Criteria:**
  **Unit Tests:**
  - Each lock file operation function
  - Version field set to 4
  - Nested map initialization
  
  **Integration Tests:**
  - Full save/load cycle preserves nested structure
  - Remove operation correctly updates nested structure

- [ ] Story-008: As a developer, I want all materialization metadata read/write operations to use the new nested structure so that the v2 format is consistently applied throughout the codebase.

  **Acceptance Criteria:**
  - `AddMaterializationEntry()` saves to nested structure with filesystemName
  - `LoadMaterializationMetadata()` loads nested structure
  - `GetComponentMap()` returns nested map structure
  - `GetAllMaterializedComponents()` iterates nested structure returning flat list with source info
  - `CheckMultipleComponentsSyncStatusBatched()` handles nested structure
  - All operations write metadata as version 2
  
  **Testing Criteria:**
  **Unit Tests:**
  - Each metadata operation function
  - Version field set to 2
  - FilesystemName field populated
  
  **Integration Tests:**
  - Full save/load cycle preserves nested structure
  - Sync status check works across nested sources

- [ ] Story-009: As a developer, I want helper functions to search across all sources for a component name so that ambiguity detection and resolution is centralized.

  **Acceptance Criteria:**
  - `FindComponentSources(componentType, componentName, lockFile) []string` returns all source URLs with that component
  - `DisambiguateComponent(componentType, componentName, sourceFilter) (sourceURL, error)` handles single vs multiple matches
  - Returns error with guidance when multiple sources found and no filter provided
  - Returns single match when only one source has component
  - Returns single match when --source filter narrows to one source
  
  **Testing Criteria:**
  **Unit Tests:**
  - FindComponentSources with 0, 1, 2+ matches
  - DisambiguateComponent with various scenarios

- [ ] Story-010: As a user, I want the materialize service to resolve filesystem names automatically so that conflicts are handled transparently during materialization.

  **Acceptance Criteria:**
  - `ResolveFilesystemName()` function checks existing disk names and metadata
  - Returns base name if no conflict
  - Returns suffixed name (`-2`, `-3`, etc.) if conflict exists
  - Checks both filesystem and materialization metadata for conflicts
  - Materialization service calls this before copying directory
  
  **Testing Criteria:**
  **Unit Tests:**
  - ResolveFilesystemName with no conflicts
  - ResolveFilesystemName with 1, 2, 3+ conflicts
  
  **Integration Tests:**
  - Materialize multiple times, verify filesystem names
  - Verify metadata filesystemName field updated

- [ ] Story-011: As a developer updating the materialize service, I want all materialization operations to use nested lookups so that components from different sources are handled correctly.

  **Acceptance Criteria:**
  - `MaterializeComponent()` uses source-aware lookups
  - `ListMaterialized()` iterates nested structure
  - `ShowComponentInfo()` searches by source + name or iterates all sources
  - `ShowStatus()` handles nested data structure
  - `UpdateMaterialized()` uses nested lookups and updates
  
  **Testing Criteria:**
  **Unit Tests:**
  - Note: Service methods tested via integration tests
  
  **Integration Tests:**
  - Each materialize command with nested data
  - List, info, status, update commands with multiple sources

- [ ] Story-012: As a developer, I want the profile manager to handle nested lock files when copying components between profiles so that source information is preserved.

  **Acceptance Criteria:**
  - `copyComponentWithMetadata()` reads from nested source structure
  - Writes to target profile's nested structure
  - Preserves source URL and all metadata fields
  - Handles cases where target already has same component from different source
  
  **Testing Criteria:**
  **Unit Tests:**
  - Note: Lock file operations already unit tested
  
  **Integration Tests:**
  - Copy component from one profile to another
  - Verify nested structure maintained in both profiles

- [ ] Story-013: As a developer, I want the updater service to search nested lock files when checking for updates so that all installed versions can be checked.

  **Acceptance Criteria:**
  - `LoadMetadata()` searches nested structure
  - When component name provided without source, checks all sources with that name
  - Returns update status for each source
  - With `--source` flag, checks only specified source
  
  **Testing Criteria:**
  **Unit Tests:**
  - Note: Lock file operations already unit tested
  
  **Integration Tests:**
  - Check for updates with same-named components from multiple sources
  - Verify update check runs for all sources or filtered source

- [ ] Story-014: As a developer, I want the linker service to handle nested lock files when loading component metadata so that linking operations work correctly.

  **Acceptance Criteria:**
  - `loadFromLockFile()` searches nested structure
  - Handles ambiguous names by iterating all sources
  - Uses source information when available for precise lookup
  
  **Testing Criteria:**
  **Unit Tests:**
  - Note: Lock file operations already unit tested
  
  **Integration Tests:**
  - Link component with nested lock file structure
  - Verify correct component linked when multiple sources present

- [ ] Story-015: As a user, I want all CLI commands that accept component names to support the `--source` flag so that I can disambiguate across all operations.

  **Acceptance Criteria:**
  - Add `--source` flag to: install, update, remove, materialize, info commands
  - Flag is optional; only required when ambiguous
  - Flag accepts full URL or short form (owner/repo)
  - Help text updated to document flag
  - Example commands in help text show --source usage
  
  **Testing Criteria:**
  **Unit Tests:**
  - Note: Flag parsing handled by cobra
  
  **Integration Tests:**
  - Run each command with --source flag
  - Verify correct component targeted

- [ ] Story-016: As a developer, I want all integration tests updated to use the new lock file and materialization structure so that tests pass with v4/v2 formats.

  **Acceptance Criteria:**
  - All test fixtures updated to v4 lock file format
  - All test fixtures updated to v2 materialization format
  - Tests that create mock lock files use nested structure
  - Tests that assert on file contents expect nested structure
  - Tests verify source URL nesting
  - Tests verify filesystemName field in materialization metadata
  
  **Testing Criteria:**
  **Unit Tests:**
  - Note: Test updates verified by test suite passing
  
  **Integration Tests:**
  - All existing integration tests pass
  - New tests for multi-source scenarios added

## Functional Requirements

- FR-1: The system SHALL store lock file data in a nested structure: `map[componentType]map[sourceURL]map[componentName]Entry`
- FR-2: The system SHALL store materialization metadata in a nested structure: `map[componentType]map[sourceURL]map[componentName]Entry`
- FR-3: The system SHALL bump lock file version to 4 when writing lock files
- FR-4: The system SHALL bump materialization metadata version to 2 when writing metadata files
- FR-5: The system SHALL add `filesystemName` field to materialization metadata entries
- FR-6: The system SHALL automatically suffix filesystem names when conflicts are detected (e.g., `-2`, `-3`)
- FR-7: The system SHALL return disambiguation errors when component names are ambiguous and no --source filter provided
- FR-8: The system SHALL support `--source <url>` flag on all component commands for disambiguation
- FR-9: The system SHALL accept both full URLs and short forms (owner/repo) for --source flag
- FR-10: The system SHALL display source information when listing materialized components
- FR-11: The system SHALL display filesystem names when different from component names
- FR-12: The system SHALL NOT provide backward compatibility for v3 lock files or v1 materialization metadata
- FR-13: The system SHALL aggregate component names across all sources when listing components
- FR-14: The system SHALL track duplicate component names in display output

## Non-Goals (Out of Scope)

- No backward compatibility or migration from v3 lock files to v4
- No backward compatibility or migration from v1 materialization metadata to v2
- No automatic resolution of ambiguous component names (always require explicit --source flag)
- No profile-aware default disambiguation (e.g., preferring active profile)
- No support for renaming components during installation via `--as` flag
- No nested filesystem structure (disk structure remains flat with auto-suffixing)
- No UUID-based or composite key approaches for uniqueness
- No interactive prompts for disambiguation (error-with-guidance only)
- No support for updating all instances of ambiguous name in single command (must specify --source)
