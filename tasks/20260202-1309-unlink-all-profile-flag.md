# PRD: Add `--profile` Flag to `unlink all` Command

**Created**: 2026-02-02 13:09 UTC

---

## Introduction

Add support for the `--profile` flag to the `unlink all` command to enable unlinking components from a specific profile without switching to it first. This mirrors the functionality already present in `link all --profile` and creates symmetry between link and unlink commands.

Currently, users can link components from a specific profile using `link all --profile <name>`, but there is no equivalent way to unlink components from a specific profile. Users must either switch to the profile first or use `--all-profiles` which affects all profiles, not just one.

## Goals

- Add `--profile` flag support to `unlink all` command for profile-specific unlinking
- Maintain consistency with existing `link all --profile` behavior and flag validation
- Enable unlinking from a specific profile without switching active profile context
- Preserve backward compatibility with existing unlink command behavior

## User Stories

- [x] Story-001: As a developer with multiple profiles, I want to unlink components from a specific profile without switching to it so that I can quickly clean up linked components without disrupting my current workspace.

  **Acceptance Criteria:**
  - User can run `agent-smith unlink all --profile work`
  - Only components from the specified profile are unlinked
  - Components from other profiles and base installation remain linked
  - User receives clear feedback about which profile components are being unlinked
  - Command works with `--target` flag: `unlink all --profile work --target opencode`
  - Command works with `--force` flag: `unlink all --profile work --force`
  
  **Testing Criteria:**
  **Unit Tests:**
  - Note: Linker infrastructure and profile detection already covered by internal/linker/linker_test.go
  
  **Integration Tests:**
  - Create test profile with linked components, run unlink with --profile flag
  - Verify only specified profile's components are unlinked
  - Verify components from other profiles remain linked
  - Test combination with --target flag
  - Test combination with --force flag

- [x] Story-002: As a user, I want to receive a clear error message when I specify an invalid profile so that I can correct my command and understand what went wrong.

  **Acceptance Criteria:**
  - Command fails with error: `profile 'invalid' does not exist`
  - Error behavior matches `link all --profile` validation
  - System state remains consistent (no partial unlinking)
  
  **Testing Criteria:**
  **Integration Tests:**
  - Run unlink with non-existent profile name
  - Verify error message matches expected format
  - Verify no components were unlinked

- [x] Story-003: As a user, I want to be prevented from using conflicting flags so that I don't accidentally perform unintended operations.

  **Acceptance Criteria:**
  - Cannot use `--profile` and `--all-profiles` together
  - Error message: "Cannot use both --all-profiles and --profile flags together"
  - Validation matches `link all` flag validation behavior
  
  **Testing Criteria:**
  **Integration Tests:**
  - Run unlink with both --profile and --all-profiles flags
  - Verify error message matches expected format
  - Verify no unlinking operation was performed

- [ ] Story-004: As a maintainer, I want the help documentation and README updated so that users understand the new --profile flag capability.

  **Acceptance Criteria:**
  - `unlinkAllCmd` Long description explains `--profile` flag usage
  - Examples section includes `--profile` usage patterns
  - README.md Unlink section includes `--profile` examples
  - Help text shows flag definition: `-p, --profile string`
  
  **Testing Criteria:**
  **Integration Tests:**
  - Run `agent-smith unlink all --help` and verify --profile flag is documented
  - Verify examples are clear and accurate

## Functional Requirements

- FR-1: The `unlink all` command SHALL accept a `--profile` flag with string value specifying the profile name
- FR-2: When `--profile` is specified, the system SHALL use `NewComponentLinkerWithFilterAndProfile` to create a linker with the specified profile's base directory
- FR-3: The system SHALL validate that the specified profile exists before attempting to unlink, returning error if profile does not exist
- FR-4: The system SHALL prevent using `--profile` and `--all-profiles` flags together, returning error "Cannot use both --all-profiles and --profile flags together"
- FR-5: The `--profile` flag SHALL work in combination with `--target` and `--force` flags
- FR-6: The command help text SHALL document the `--profile` flag with description "Unlink from specific profile (bypasses active profile)"
- FR-7: All existing `unlink all` behavior (without `--profile` flag) SHALL remain unchanged for backward compatibility

## Non-Goals

- Adding `--profile` flag to `unlink skills`, `unlink agents`, or `unlink commands` (may be added in future iteration)
- Adding `--profile` flag to singular unlink commands (`unlink skill <name>`, `unlink agent <name>`, `unlink command <name>`)
- Changing existing `--all-profiles` flag behavior or semantics
- Modifying the `UnlinkAllComponents` function signature or core unlinking logic
- Adding new profile validation beyond what exists in `NewComponentLinkerWithFilterAndProfile`
