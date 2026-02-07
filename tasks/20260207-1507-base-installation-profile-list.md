# PRD: Show Base Installation in Profile List

**Created**: 2026-02-07 15:07 UTC

---

## Introduction

Add base installation (`~/.agent-smith/`) as a visible item in `agent-smith profile list` with label "(no profile)", and update all terminology from "base" to "(no profile)" throughout the codebase for consistency. This addresses the user confusion where base components exist but aren't visible in the profile list, while other commands like `link status --all-profiles` show "base" as a profile.

## Goals

- Make base installation visible in `agent-smith profile list` output
- Use consistent terminology "(no profile)" instead of "base" across all commands
- Define a constant for the "(no profile)" label to enable easy future changes
- Maintain backward compatibility with zero breaking changes
- Provide clear distinction between base installation and profiles

## User Stories

- [ ] Story-001: As a user with components in base installation, I want to see them listed in `agent-smith profile list` so that I can discover all my installed components in one place.

  **Acceptance Criteria:**
  - Base installation appears as first row in profile list with label "(no profile)"
  - Symbol ⊙ (circled dot) used to identify base installation
  - Component count shows skills, agents, and commands in base
  - Only shown if base has components (empty base hidden)
  - Base never shows active indicator (✓)
  
  **Testing Criteria:**
  **Unit Tests:**
  - scanBaseInstallation returns profile with correct counts
  - scanBaseInstallation returns nil for empty base
  
  **Integration Tests:**
  - Profile list displays base when components exist
  - Profile list excludes base when empty

- [ ] Story-002: As a developer, I want a constant for "(no profile)" label so that the terminology can be easily changed in the future without hunting through the codebase.

  **Acceptance Criteria:**
  - Constant `BaseProfileName = "(no profile)"` defined in paths package
  - All code uses constant instead of hardcoded string
  - Constant documented with comment explaining usage
  
  **Testing Criteria:**
  **Unit Tests:**
  - Constant value verified in paths package test

- [ ] Story-003: As a user running `agent-smith link status --all-profiles`, I want to see "(no profile)" instead of "base" so that terminology is consistent with profile list.

  **Acceptance Criteria:**
  - Profile column shows "(no profile)" instead of "base"
  - Summary says "base installation + X profile(s)"
  - Source description uses "base installation" in parenthetical
  - All messages use "(no profile)" when referring to base by name
  
  **Testing Criteria:**
  **Unit Tests:**
  - getProfileFromPath returns BaseProfileName constant for base paths
  
  **Integration Tests:**
  - Link status output contains "(no profile)" not "base"

- [ ] Story-004: As a user, I want the profile list count to distinguish between base and profiles so that I understand what I'm looking at.

  **Acceptance Criteria:**
  - Count shows "X profile(s) + base installation" when both exist
  - Count shows "base installation only" when no profiles
  - Count shows "X profile(s)" when no base (empty or no components)
  - Legend includes "⊙ - Base installation (no profile)"
  
  **Testing Criteria:**
  **Unit Tests:**
  - Count formatting logic handles all combinations
  
  **Integration Tests:**
  - Count displays correctly for various scenarios

- [ ] Story-005: As a user running `agent-smith profile show "(no profile)"`, I want to see base installation details so that I can inspect what's installed in base.

  **Acceptance Criteria:**
  - Command accepts "(no profile)" as valid profile name
  - Shows base directory path and component list
  - Clear indication this is base installation, not a profile
  
  **Testing Criteria:**
  **Integration Tests:**
  - profile show "(no profile)" displays base details

- [ ] Story-006: As a user filtering profiles by type, I want base to be excluded from `--type=repo` and `--type=user` filters so that I only see the types I requested.

  **Acceptance Criteria:**
  - Base has type "base" (not "repo" or "user")
  - `--type=repo` excludes base from results
  - `--type=user` excludes base from results
  - No `--type=base` filter needed (base always shown without filter)
  
  **Testing Criteria:**
  **Unit Tests:**
  - GetProfileType returns "base" for BaseProfileName constant
  
  **Integration Tests:**
  - Type filters properly exclude base

## Functional Requirements

- FR-1: The system SHALL define a constant `BaseProfileName = "(no profile)"` in the paths package for consistent labeling
- FR-2: The system SHALL scan base installation (`~/.agent-smith/`) for components and create a pseudo-profile entry if components exist
- FR-3: The system SHALL display base installation as first row in profile list with ⊙ symbol
- FR-4: The system SHALL NOT display base installation if no components exist in skills, agents, or commands directories
- FR-5: The system SHALL use BaseProfileName constant in all locations previously using "base" string
- FR-6: The system SHALL update link status display to use "(no profile)" in the Profile column
- FR-7: The system SHALL update summary counts to say "base installation + X profile(s)"
- FR-8: The system SHALL assign type "base" to the base installation pseudo-profile
- FR-9: The system SHALL exclude base from `--type=repo` and `--type=user` filters
- FR-10: The system SHALL never show active indicator (✓) on base installation row
- FR-11: The system SHALL support `agent-smith profile show "(no profile)"` to display base installation details
- FR-12: The system SHALL update all comments and documentation to use "(no profile)" terminology

## Non-Goals

- No `--type=base` filter option (base shows by default, unnecessary complexity)
- No `--profile="(no profile)"` filtering in profile list (edge case, low value)
- No ability to activate base as active profile (use `profile deactivate` instead)
- No ability to delete or rename base installation
- No migration of existing base components to profiles
- No changes to lock file format or component storage structure
- No changes to install/link/unlink behavior (only display changes)

## Technical Implementation Notes

### Constant Definition
```go
// pkg/paths/paths.go
const BaseProfileName = "(no profile)"
```

### Key Files to Modify
1. `pkg/paths/paths.go` - Add BaseProfileName constant
2. `pkg/services/profile/service.go` - Add scanBaseInstallation(), update ListProfiles()
3. `internal/linker/linker.go` - Replace "base" with BaseProfileName constant
4. `internal/linker/status.go` - Update comments and returns
5. `pkg/profiles/manager.go` - Handle BaseProfileName in GetProfileType()

### Testing Locations
- Unit: `pkg/paths/paths_test.go`, `pkg/services/profile/service_test.go`, `internal/linker/status_test.go`
- Integration: `tests/integration/e2e_workflow_integration_test.go`, existing profile tests

## Success Metrics

- Zero breaking changes (all existing commands work identically)
- Base installation visible in profile list when components exist
- Consistent "(no profile)" terminology across all commands
- Single constant definition for easy future changes
- All tests passing with updated expectations
- User confusion eliminated (base components discoverable)

## Dependencies

- Existing `ProfileManager` in `pkg/profiles/manager.go`
- Existing `paths` package in `pkg/paths/paths.go`
- Existing profile list implementation in `pkg/services/profile/service.go`
- Existing link status implementation in `internal/linker/linker.go`

## Future Enhancements (Not in Scope)

- Add `--type=base` filter to show only base installation
- Support `--profile="(no profile)"` filtering in multi-profile views
- Add `profile migrate-to-base` command to move profile components to base
- Add `profile migrate-from-base` command to move base components to profile
- Historical tracking of base vs profile usage
- Metrics on base installation size and component counts
