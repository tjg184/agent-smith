# PRD: Materialize Status and Update Commands

**Created**: 2026-02-03 01:24 UTC

---

## Introduction

Add `materialize status` and `materialize update` commands to help users keep materialized components in sync with their source components after running `agent-smith update all`. Currently, when users update source components in `~/.agent-smith/` or profiles, the materialized copies in `.opencode/` or `.claude/` directories become out of sync, with no way to detect or fix this.

## Goals

- Enable users to check which materialized components are out of sync with their sources
- Provide a smart update mechanism that only re-materializes changed components
- Warn users when materialized component sources are no longer installed
- Maintain consistency between source components and project-materialized copies
- Follow existing `materialize` command patterns for user experience

## User Stories

- [ ] Story-001: As a developer, I want to check which materialized components are out of sync so that I know what needs updating after running update all.

  **Acceptance Criteria:**
  - Command `agent-smith materialize status` shows all materialized components
  - Components marked as "in sync" (✓), "out of sync" (⚠), or "source missing" (✗)
  - Output grouped by target (.opencode/, .claude/) and component type
  - Shows summary count (X in sync, Y out of sync, Z source missing)
  - Uses concise output format by default
  
  **Testing Criteria:**
  **Unit Tests:**
  - CheckComponentSyncStatus returns correct status for in sync, out of sync, and missing sources
  
  **Integration Tests:**
  - Status shows "in sync" for freshly materialized components
  - Status shows "out of sync" after source component is updated
  - Status shows "source missing" after component is uninstalled
  - Status respects --target flag to check specific target only
  - Status works with both base and profile sources

- [ ] Story-002: As a developer, I want to update only out-of-sync materialized components so that I don't waste time re-copying unchanged files.

  **Acceptance Criteria:**
  - Command `agent-smith materialize update` updates only components where source hash differs
  - Skips components that are already in sync
  - Shows which components were updated vs skipped
  - Updates metadata with new hashes and timestamp
  - Shows summary count (X updated, Y already in sync, Z skipped)
  
  **Testing Criteria:**
  **Unit Tests:**
  - UpdateMaterializationEntry correctly updates metadata fields
  
  **Integration Tests:**
  - Update skips in-sync components in smart mode
  - Update only re-materializes components with changed sources
  - Update refreshes metadata with new sourceHash and currentHash
  - Update respects --target flag to update specific target only
  - Update re-reads lock files to get latest commit hash

- [ ] Story-003: As a developer, I want to force re-materialize all components so that I can ensure everything is freshly copied regardless of sync status.

  **Acceptance Criteria:**
  - Flag `--force` causes update to re-materialize all components
  - Ignores sync status check when --force is used
  - Shows which components were force-updated
  - Updates all metadata entries
  
  **Testing Criteria:**
  **Integration Tests:**
  - Update with --force re-materializes all components regardless of sync status

- [ ] Story-004: As a developer, I want to preview what would be updated so that I can verify changes before applying them.

  **Acceptance Criteria:**
  - Flag `--dry-run` shows what would be updated without making changes
  - Output clearly indicates "Would update..." for dry-run mode
  - Does not modify any files or metadata
  - Shows same summary format as real update
  
  **Testing Criteria:**
  **Integration Tests:**
  - Dry-run shows preview without modifying files or metadata

- [ ] Story-005: As a developer, I want clear warnings when materialized components have missing sources so that I understand why they can't be updated.

  **Acceptance Criteria:**
  - Components with missing sources show warning message
  - Warning includes component name and type
  - Update continues processing other components after warning
  - Summary includes count of skipped components due to missing sources
  
  **Testing Criteria:**
  **Integration Tests:**
  - Update warns and skips components with missing sources
  - Update continues processing remaining components after encountering missing source

- [ ] Story-006: As a developer, I want status and update commands to work with specific targets so that I can manage .opencode/ and .claude/ separately.

  **Acceptance Criteria:**
  - Flag `--target <name>` limits operation to specific target directory
  - Works with both status and update commands
  - Only processes specified target, skips others
  
  **Testing Criteria:**
  **Integration Tests:**
  - Status and update respect --target flag for specific target operations

- [ ] Story-007: As a developer, I want helper functions to check component sync status so that both status and update commands can reuse the same logic.

  **Acceptance Criteria:**
  - Function CheckComponentSyncStatus checks if source exists and compares hashes
  - Returns "in_sync", "out_of_sync", or "source_missing"
  - Resolves source path from metadata (profile or base directory)
  - Handles missing source directories gracefully
  
  **Testing Criteria:**
  **Unit Tests:**
  - CheckComponentSyncStatus handles all status cases correctly
  - Source path resolution works for base and profile sources

- [ ] Story-008: As a developer, I want helper functions to iterate all materialized components so that I can process them consistently.

  **Acceptance Criteria:**
  - Function GetAllMaterializedComponents returns flat list from metadata
  - Returns component type, name, and metadata for each component
  - Works with all three component types (skills, agents, commands)
  
  **Testing Criteria:**
  **Unit Tests:**
  - GetAllMaterializedComponents returns all components from metadata

- [ ] Story-009: As a developer, I want to update materialization metadata entries so that I can refresh hashes and timestamps after re-materializing.

  **Acceptance Criteria:**
  - Function UpdateMaterializationEntry updates sourceHash, currentHash, materializedAt
  - Re-reads lock files to get latest commit hash
  - Preserves other metadata fields (source, sourceType, sourceProfile, originalPath)
  
  **Testing Criteria:**
  **Unit Tests:**
  - UpdateMaterializationEntry updates correct metadata fields
  
  **Integration Tests:**
  - Metadata updated with new hashes after re-materialization

## Functional Requirements

- FR-1: The system SHALL provide `materialize status` command to show sync status of all materialized components
- FR-2: The system SHALL mark components as "in sync", "out of sync", or "source missing" based on hash comparison
- FR-3: The system SHALL provide `materialize update` command to re-materialize out-of-sync components
- FR-4: The system SHALL skip in-sync components by default (smart mode)
- FR-5: The system SHALL provide `--force` flag to re-materialize all components regardless of sync status
- FR-6: The system SHALL provide `--dry-run` flag to preview updates without making changes
- FR-7: The system SHALL warn and skip components with missing sources without failing the entire operation
- FR-8: The system SHALL update metadata with new hashes and timestamp after re-materialization
- FR-9: The system SHALL re-read lock files to get latest commit hash during updates
- FR-10: The system SHALL support `--target` flag to operate on specific target directories
- FR-11: The system SHALL calculate source hash by calling CalculateDirectoryHash on source directory
- FR-12: The system SHALL compare source hash with metadata sourceHash to determine sync status
- FR-13: The system SHALL resolve source path from metadata sourceProfile (profile or base directory)
- FR-14: The system SHALL provide CheckComponentSyncStatus helper for status checking
- FR-15: The system SHALL provide GetAllMaterializedComponents helper for iteration
- FR-16: The system SHALL provide UpdateMaterializationEntry helper for metadata updates
- FR-17: The system SHALL display summary counts for all operations
- FR-18: The system SHALL use concise output format (no verbose details by default)

## Non-Goals (Out of Scope)

- No verbose mode with detailed provenance information (keep output concise)
- No non-zero exit codes based on sync status (not needed for now)
- No "last checked" timestamp separate from "materialized at" in metadata
- No automatic update hints after `agent-smith update all` completes
- No support for updating specific individual components (update all or nothing)
- No automatic cleanup of components with missing sources
- No migration of existing materialized components to new metadata format
- No bi-directional sync (adopt command to import components back to ~/.agent-smith/)

## Technical Implementation Notes

### File Changes

- `cmd/root.go`: Add materializeStatusCmd and materializeUpdateCmd command definitions
- `cmd/root.go`: Add handler signatures and wire up in SetHandlers
- `main.go`: Implement handleMaterializeStatus and handleMaterializeUpdate
- `pkg/project/materialization.go`: Add CheckComponentSyncStatus, GetAllMaterializedComponents, UpdateMaterializationEntry
- `tests/integration/materialize_status_test.go`: New integration tests for status command
- `tests/integration/materialize_update_test.go`: New integration tests for update command
- `README.md`: Add workflow section showing update all → status → update

### Key Design Decisions

- **Smart by default**: Only update out-of-sync components unless --force specified
- **Warn on missing sources**: Don't fail entire operation, just skip and warn
- **Concise output**: No verbose mode needed initially
- **Reuse existing patterns**: Follow materialize all/list/info command structure
- **Hash comparison**: Compare CalculateDirectoryHash(source) with metadata.SourceHash

### Edge Cases

- Source in profile but profile deactivated: Still locate via metadata sourceProfile
- Source moved between base and profile: Use metadata sourceProfile to find
- Lock file missing/corrupted: Show warning and skip component
- Permission errors during copy: Fail with clear error message
- Project has no materialized components: Show "No materialized components found"
- Multiple targets (.opencode/ and .claude/): Process both independently

## Success Criteria

- User can run `agent-smith materialize status` to see sync status of all materialized components
- User can run `agent-smith materialize update` to update only out-of-sync components
- User can run `agent-smith materialize update --force` to re-materialize everything
- User can run `agent-smith materialize update --dry-run` to preview changes
- All integration tests pass
- Documentation updated with new workflow examples
