# PRD: Multi-Profile Link Status Command

**Created**: 2026-02-01 12:01 UTC

---

## Introduction

Add the ability to view link status across all profiles simultaneously in agent-smith. Currently, `agent-smith link status` only shows components from the active profile or base installation. This feature enables users to see which skills, agents, and commands are linked across all their profiles at once, providing a comprehensive overview of their entire component ecosystem.

## Goals

- Enable users to view link status for components across all profiles in a single command
- Provide a filterable view to focus on specific profiles when needed
- Maintain backward compatibility with existing `link status` command behavior
- Display clear profile attribution for each component in the matrix view
- Help users understand their complete component linking landscape

## User Stories

- [x] Story-001: As a user with multiple profiles, I want to see link status for all profiles at once so that I can understand my complete component ecosystem without switching between profiles.

  **Acceptance Criteria:**
  - New `--all-profiles` flag added to `agent-smith link status` command
  - Command scans all profiles plus base installation
  - Matrix view displays Component | Type | Profile | Status per Target
  - Components grouped by type (skills, agents, commands)
  - Status symbols match existing conventions (✓ ◆ ✗ - ?)
  - Summary statistics show total profiles scanned and link percentages
  - Active profile is indicated in the summary output
  
  **Testing Criteria:**
  **Unit Tests:**
  - ComponentLinker.ShowAllProfilesLinkStatus() returns correct data structure
  - Profile scanning logic correctly identifies all profiles
  - Link status detection works for components from different profiles
  - Summary calculation logic computes correct percentages
  
  **Integration Tests:**
  - Command executes successfully with --all-profiles flag
  - Output formatting matches expected matrix structure
  - All profiles (including base) are scanned
  - Link status symbols display correctly for all states
  
  **Component Browser Tests:**
  - Not applicable (CLI tool, no browser components)

- [x] Story-002: As a user managing specific profiles, I want to filter the all-profiles view to show only selected profiles so that I can focus on relevant components without visual clutter.

  **Acceptance Criteria:**
  - New `--profile` flag accepts comma-separated profile names
  - `--profile` flag requires `--all-profiles` to be set
  - Error message displayed if `--profile` used without `--all-profiles`
  - Filtered view shows only components from specified profiles
  - Summary statistics reflect filtered profile count
  - Invalid profile names in filter produce helpful error messages
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile filtering logic correctly filters by name list
  - Validation logic rejects --profile without --all-profiles
  - Invalid profile names are properly detected and reported
  
  **Integration Tests:**
  - Command executes with --all-profiles --profile=profile1,profile2
  - Output shows only components from specified profiles
  - Error handling works for invalid profile names
  - Error handling works for --profile without --all-profiles
  
  **Component Browser Tests:**
  - Not applicable (CLI tool, no browser components)

- [x] Story-003: As a developer maintaining agent-smith, I want the ComponentLinker to accept ProfileManager as a dependency so that multi-profile scanning can be implemented without breaking existing functionality.

  **Acceptance Criteria:**
  - ComponentLinker struct includes profileManager field
  - NewComponentLinker constructor accepts ProfileManager parameter (can be nil for backward compatibility)
  - All existing constructor calls updated to pass ProfileManager
  - Nil ProfileManager gracefully handled (existing single-profile behavior works)
  - No breaking changes to ComponentLinker's existing API
  
  **Testing Criteria:**
  **Unit Tests:**
  - NewComponentLinker accepts ProfileManager parameter
  - NewComponentLinker handles nil ProfileManager gracefully
  - Existing linker methods work unchanged with ProfileManager present
  
  **Integration Tests:**
  - All existing link commands continue to work
  - No regressions in link, unlink, list functionality
  
  **Component Browser Tests:**
  - Not applicable (CLI tool, no browser components)

- [ ] Story-004: As a user running the default link status command, I want the existing behavior to remain unchanged so that I can continue using familiar workflows without disruption.

  **Acceptance Criteria:**
  - `agent-smith link status` without flags shows current profile/base only (existing behavior)
  - Output format for single-profile view unchanged
  - Performance characteristics unchanged for single-profile view
  - No new flags required for existing use cases
  - Backward compatibility with all existing command patterns maintained
  
  **Testing Criteria:**
  **Unit Tests:**
  - ShowLinkStatus() method unchanged and working
  - Flag defaults ensure existing behavior when no flags set
  
  **Integration Tests:**
  - `agent-smith link status` output matches existing format
  - Performance regression testing shows no degradation
  - All existing link status test cases pass
  
  **Component Browser Tests:**
  - Not applicable (CLI tool, no browser components)

## Functional Requirements

### Core Functionality

- FR-1: The system SHALL add a `ShowAllProfilesLinkStatus(profileFilter []string)` method to ComponentLinker that scans all profiles and returns link status for all components
- FR-2: The system SHALL accept an optional list of profile names to filter the results to specific profiles only
- FR-3: The system SHALL display components in a matrix format with columns: Component Name, Type, Profile, and one column per configured target
- FR-4: The system SHALL group components by type (skills, agents, commands) in the output
- FR-5: The system SHALL use standard status symbols: ✓ (valid symlink), ◆ (copied), ✗ (broken), - (not linked), ? (unknown)
- FR-6: The system SHALL include a summary section showing total profiles scanned, total components found, and linking percentages per target
- FR-7: The system SHALL indicate the currently active profile in the summary output

### Command Interface

- FR-8: The system SHALL add a `--all-profiles` boolean flag to the `link status` command
- FR-9: The system SHALL add a `--profile` string slice flag accepting comma-separated profile names
- FR-10: The system SHALL enforce that `--profile` flag can only be used when `--all-profiles` is also set
- FR-11: The system SHALL maintain existing behavior when neither new flag is used (backward compatibility)

### Architecture

- FR-12: The ComponentLinker struct SHALL include a profileManager field of type *profiles.ProfileManager
- FR-13: The NewComponentLinker constructor SHALL accept a ProfileManager parameter that may be nil
- FR-14: The system SHALL handle nil ProfileManager gracefully, allowing existing single-profile operations to continue working
- FR-15: All existing ComponentLinker instantiations SHALL be updated to pass a ProfileManager instance

### Profile Scanning

- FR-16: The system SHALL use ProfileManager.ScanProfiles() to discover all available profiles
- FR-17: The system SHALL include the base installation (non-profile components) in the scan results
- FR-18: The system SHALL use existing getProfileFromPath() function to determine profile attribution for symlinks
- FR-19: The system SHALL scan all component types (skills, agents, commands) for each profile

### Error Handling

- FR-20: The system SHALL display a clear error message when `--profile` is used without `--all-profiles`
- FR-21: The system SHALL display helpful error messages for invalid profile names in the filter
- FR-22: The system SHALL handle missing or empty profiles directories gracefully
- FR-23: The system SHALL continue execution if individual component link status checks fail, logging warnings

## Non-Goals (Out of Scope)

- No modifications to the linking/unlinking logic itself
- No changes to how profiles are created or managed
- No automatic profile switching based on link status
- No export functionality for link status data (JSON, CSV, etc.)
- No filtering by component type (only profile filtering)
- No filtering by link status (e.g., "show only broken links")
- No detailed path information in matrix view (only in existing list view)
- No per-profile link status subcommand (e.g., `agent-smith profiles link-status <name>`)
- No graphical/visual representations beyond ASCII table
- No historical link status tracking or comparison

## Implementation Notes

### File Structure
```
internal/linker/
  ├── linker.go          # Add ShowAllProfilesLinkStatus() method
  └── status.go          # Existing LinkStatus struct (no changes)

pkg/profiles/
  └── manager.go         # Existing ProfileManager (no changes)

cmd/
  └── [link-command].go  # Add flags and update command logic
```

### Data Flow
1. User runs `agent-smith link status --all-profiles [--profile=p1,p2]`
2. Link command parses flags and validates
3. Command calls `linker.ShowAllProfilesLinkStatus(profileFilter)`
4. Linker uses ProfileManager to scan all profiles
5. For each profile, linker scans all component types
6. For each component, linker checks link status across all targets
7. Results compiled into matrix data structure
8. Matrix formatted and displayed with summary statistics

### Display Format Example
```
=== Link Status Across All Profiles ===

Component               Type      Profile         OPENCODE     CLAUDECODE
------------------------------------------------------------------------

Skills:
  api-design            skills    work-profile    ✓            ✓
  typescript-utils      skills    base            ✓            -
  python-ml             skills    data-profile    -            -

Agents:
  backend-dev           agents    work-profile    ✓            ✓
  code-reviewer         agents    base            ✓            ✓

Commands:
  docker-helper         commands  base            ✓            -

Legend:
  ✓  Valid symlink
  ◆  Copied directory
  ✗  Broken link
  -  Not linked
  ?  Unknown status

Summary:
  Profiles scanned: 3 (base + 2 custom)
  Total components: 6
  OPENCODE: 5/6 linked (83%)
  CLAUDECODE: 3/6 linked (50%)

Active Profile: work-profile
```

## Testing Scenarios

### Scenario 1: Multiple Profiles with Various Link States
**Given**: User has 3 profiles (work, personal, experimental) plus base installation
**When**: User runs `agent-smith link status --all-profiles`
**Then**: Display shows all components from all 4 sources with correct link status

### Scenario 2: Profile Filtering
**Given**: User has multiple profiles but wants to focus on specific ones
**When**: User runs `agent-smith link status --all-profiles --profile=work,personal`
**Then**: Display shows only components from work and personal profiles

### Scenario 3: No Profiles Exist
**Given**: User has only base installation, no custom profiles
**When**: User runs `agent-smith link status --all-profiles`
**Then**: Display shows components from base installation only, summary indicates "1 profile (base)"

### Scenario 4: Invalid Flag Combination
**Given**: User attempts to filter without all-profiles flag
**When**: User runs `agent-smith link status --profile=work`
**Then**: Error message displayed: "--profile flag requires --all-profiles"

### Scenario 5: Backward Compatibility
**Given**: User runs existing command without new flags
**When**: User runs `agent-smith link status`
**Then**: Output matches existing single-profile behavior exactly

## Success Metrics

- All existing link status tests pass (no regressions)
- New command successfully displays components from multiple profiles
- Profile filtering works correctly with various filter combinations
- Performance acceptable even with 10+ profiles containing 50+ components each
- No breaking changes to ComponentLinker API
- Clear, readable matrix output for users with multiple profiles

## Dependencies

- Existing `ProfileManager` in `pkg/profiles/manager.go`
- Existing `ComponentLinker` in `internal/linker/linker.go`
- Existing `LinkStatus` struct in `internal/linker/status.go`
- Cobra command framework for flag parsing

## Future Considerations (Not in Scope)

- Export link status data to JSON/CSV for external processing
- Historical tracking of link status changes over time
- Automatic link health monitoring and alerting
- Bulk link operations based on link status (e.g., "link all unlinked components")
- Link status filtering (e.g., show only broken links across all profiles)
- Visual graph representation of profile linking relationships
