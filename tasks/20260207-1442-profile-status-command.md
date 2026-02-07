# PRD: Rename `profile show` to `profile status` and Make Profile Argument Optional

**Created**: 2026-02-07 14:42 UTC

## Introduction

The goal is to increase CLI command clarity and consistency by renaming the `agent-smith profile show` command to `agent-smith profile status`, and making the profile argument optional with a default to the active profile. If no profile is specified and none is active, the command should output a clear error, matching the behavior when a profile does not exist. This brings `profile status` in line with other commands that use the active profile as default.

## Goals

- Improve CLI consistency by using `status` rather than `show`.
- Make profile argument optional and default to active profile if available.
- Reduce user confusion and streamline profile inspection.
- Output clear error if neither a profile nor active profile is available.

## User Stories

- [ ] Story-001: As a user, I want to run `agent-smith profile status` without arguments to see details for my active profile, so that I don’t need to remember its name.

  **Acceptance Criteria:**
  - Running `agent-smith profile status` with no argument shows details for active profile.
  - If no profile is active, output a clear error message.
  - If profile is specified and exists, show its details.
  - If profile specified but does not exist, output error.

  **Testing Criteria:**
  **Unit Tests:**
  - Functionality for detecting and displaying the active profile.
  - Error output for nonexistent profile.
  - Argument parsing for optional profile name.
  
  **Integration Tests:**
  - CLI workflow tests covering default (active) profile, explicit profile, and error paths.

- [ ] Story-002: As a user, I want documentation and help output to reflect the command rename and new default behavior, so that I understand how to use it intuitively.

  **Acceptance Criteria:**
  - All references in README, SKILL.md, PRODUCT_SNAPSHOT.md, CHANGELOG.md updated.
  - CLI help and usage output reflects `status` command.
  - Examples use new command and show argument-default logic.

  **Testing Criteria:**
  **Unit Tests:**
  - Automated check for presence/absence of old "show" references in docs.
  
  **Integration Tests:**
  - CLI help output test with/without arguments.

## Functional Requirements

1. FR-1: The command `agent-smith profile status [profile-name]` SHALL replace all uses of `agent-smith profile show <profile-name>`.
2. FR-2: The profile name argument SHALL be optional.
3. FR-3: If profile name is not provided, the command SHALL default to the active profile.
4. FR-4: If no profile is specified and none is active, the command SHALL output a clear error.
5. FR-5: Documentation, examples, and help output SHALL reference the new command and updated behavior.
6. FR-6: All legacy references to `show` for profile inspection SHALL be removed.

## Non-Goals

- The command will not fallback to the base/global profile if no profile is specified or active.
- No changes will be made to other profile-related commands.
- No new aliases for "show" will be added.
- No support for multi-profile output is introduced.

---

This PRD supports an implementation that is simple, robust, and aligned with expectations for command consistency. The focus is on clarity and explicit error messaging without expanding scope unnecessarily.
