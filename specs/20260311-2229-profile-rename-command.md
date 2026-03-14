# PRD: Profile Rename Command

**Created**: 2026-03-11 22:29 UTC

---

## Introduction

Add a `profile rename <old-name> <new-name>` command that allows users to rename user-created profiles. If the renamed profile was active, the active state is preserved under the new name and symlinks are cleaned up so the user can restore them with `link all`.

## Goals

- Allow user-created profiles to be renamed without data loss
- Preserve active profile state when renaming the currently active profile
- Reject renames of repo-sourced profiles (auto-created by `install all`)
- Validate new name follows existing naming rules (alphanumeric + hyphens)
- Prevent renames to already-existing profile names

## User Stories

- [x] Story-001: As a user, I want to rename an inactive profile so that I can reorganize my profiles without recreating them.

  **Acceptance Criteria:**
  - `profile rename <old-name> <new-name>` renames the profile directory
  - Old profile directory no longer exists after rename
  - New profile directory contains all original components and metadata
  - Active profile state is unchanged

  **Testing Criteria:**
  **Unit Tests:**
  - `TestRenameProfile_InactiveProfile`: directory renamed, old path gone, active state unchanged
  - `TestRenameProfile_ComponentsPreserved`: component files and subdirs intact after rename
  - `TestRenameProfile_MetadataPreserved`: `.profile-metadata` type field survives rename

- [x] Story-002: As a user, I want to rename my currently active profile so that I can rename it without having to deactivate and reactivate it manually.

  **Acceptance Criteria:**
  - CLI prompts `[y/N]` confirmation before renaming an active profile
  - After rename, `.active-profile` state file contains the new name
  - CLI instructs user to run `agent-smith link all` to restore symlinks
  - Existing symlinks to the old path are cleaned up before rename

  **Testing Criteria:**
  **Unit Tests:**
  - `TestRenameProfile_ActiveProfile`: active profile renamed, `.active-profile` updated to new name

- [x] Story-003: As a user, I want clear error messages when a rename fails so that I understand what went wrong.

  **Acceptance Criteria:**
  - Error if old profile does not exist
  - Error if new profile name already exists
  - Error if new name contains invalid characters
  - Error if attempting to rename a repo-sourced profile

  **Testing Criteria:**
  **Unit Tests:**
  - `TestRenameProfile_OldProfileNotFound`
  - `TestRenameProfile_NewNameAlreadyExists`
  - `TestRenameProfile_InvalidNewName`
  - `TestRenameProfile_RepoProfileRejected`

## Functional Requirements

- FR-1: The system SHALL expose `agent-smith profile rename <old-name> <new-name>` as a CLI subcommand
- FR-2: The system SHALL validate `<new-name>` using the existing profile name validation rules (`^[a-zA-Z0-9-]+$`)
- FR-3: The system SHALL reject renames where `<new-name>` already exists as a valid profile
- FR-4: The system SHALL reject renames of profiles with `type: "repo"` metadata
- FR-5: When the renamed profile is active, the system SHALL remove existing symlinks before renaming and update `.active-profile` to the new name
- FR-6: The system SHALL prompt for `[y/N]` confirmation before renaming an active profile
- FR-7: After renaming an active profile, the system SHALL inform the user to run `agent-smith link all` to restore symlinks

## Non-Goals

- No `--force` flag to skip confirmation (prompt is always shown for active profiles)
- No automatic re-linking after rename (user runs `link all` manually)
- Renaming repo-sourced profiles is not supported
- No update of `sourceProfile` fields in materialized project lock files (treated as informational)
