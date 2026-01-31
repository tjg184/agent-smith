# PRD: Profile Reuse on Repeated Install

**Created**: 2026-01-31 22:41 UTC

---

## Introduction

Currently, when users run `./agent-smith install all <repository-url>` multiple times with the same repository URL, the system creates new profiles with hash suffixes (e.g., `anthropics-skills`, `anthropics-skills-82a901`, `anthropics-skills-82a901-2`). This behavior is confusing and creates profile clutter.

This PRD defines a solution to detect and reuse existing profiles when installing from the same repository URL, providing a cleaner and more intuitive user experience.

## Goals

- Detect when a profile already exists for a given repository URL
- Reuse and update the existing profile instead of creating duplicates
- Normalize repository URLs to recognize variations (HTTPS, SSH, shorthand) as the same source
- Provide clear feedback when updating an existing profile
- Allow users to override and force creation of new profiles when needed
- Minimize breaking changes to existing functionality

## User Stories

- [ ] Story-001: As a user, I want repeated installs from the same repository to update my existing profile so that I don't accumulate duplicate profiles.

  **Acceptance Criteria:**
  - When running `install all <repo-url>` for a repository that already has a profile, the system detects the existing profile
  - The system displays a message: "Found existing profile 'profile-name' from this repository. Updating..."
  - Components are installed to the existing profile directory
  - Existing components with the same name are overwritten with new versions
  - Components in the profile that are not in the new install are preserved
  - No new profile with hash suffix is created
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile detection by source URL logic
  - URL normalization function tests
  - Metadata file read/write operations
  
  **Integration Tests:**
  - Install all command with existing profile
  - Install all command with new repository
  - Component merge behavior validation
  
  **Component Browser Tests:**
  - N/A (CLI-only feature)

- [ ] Story-002: As a user, I want repository URL variations to be recognized as the same source so that HTTPS, SSH, and shorthand formats are treated consistently.

  **Acceptance Criteria:**
  - `https://github.com/owner/repo`, `git@github.com:owner/repo`, and `owner/repo` are normalized to the same repository
  - Trailing slashes and `.git` suffixes are removed during normalization
  - GitLab and Bitbucket URLs are also normalized
  - URL comparison is case-insensitive for domain names
  - Normalized URL is used for profile detection and metadata storage
  
  **Testing Criteria:**
  **Unit Tests:**
  - URL normalization for GitHub HTTPS format
  - URL normalization for GitHub SSH format
  - URL normalization for shorthand format (owner/repo)
  - URL normalization for GitLab and Bitbucket
  - Case-insensitive domain comparison
  - Trailing slash and .git suffix removal
  
  **Integration Tests:**
  - Install with HTTPS URL, then SSH URL → same profile detected
  - Install with shorthand, then full URL → same profile detected
  
  **Component Browser Tests:**
  - N/A (CLI-only feature)

- [ ] Story-003: As a developer, I want profile metadata to include the source repository URL so that the system can detect duplicates.

  **Acceptance Criteria:**
  - When a profile is created via `install all`, a `.profile-metadata` file is created in the profile directory
  - The metadata file contains the normalized repository URL in plain text format
  - The metadata file is UTF-8 encoded and human-readable
  - When scanning for existing profiles, the system reads and parses metadata files
  - Profiles without metadata files are skipped during duplicate detection (backward compatibility)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Metadata file creation with normalized URL
  - Metadata file parsing and validation
  - Handling of missing metadata files (backward compatibility)
  - UTF-8 encoding validation
  
  **Integration Tests:**
  - Profile creation generates metadata file
  - Profile detection reads metadata correctly
  - Legacy profiles without metadata are handled gracefully
  
  **Component Browser Tests:**
  - N/A (CLI-only feature)

- [ ] Story-004: As a user, I want to force creation of a new profile from the same repository when needed so that I can maintain multiple configurations.

  **Acceptance Criteria:**
  - The `install all` command accepts an optional `--profile <custom-name>` flag
  - When `--profile` is specified, the system creates a new profile with the given name
  - The new profile is created even if a profile from the same repository already exists
  - If the custom profile name already exists, an error is shown
  - Metadata is still saved for the new profile
  - The behavior without the flag remains as described in Story-001
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile name validation logic
  - Custom profile name handling
  - Duplicate custom name detection
  
  **Integration Tests:**
  - Install all with --profile flag creates new profile
  - Install all with --profile and existing name shows error
  - Install all without --profile updates existing profile
  
  **Component Browser Tests:**
  - N/A (CLI-only feature)

- [ ] Story-005: As a user, I want clear feedback when my profile is being updated so that I understand what the system is doing.

  **Acceptance Criteria:**
  - When updating an existing profile, display: "Found existing profile 'profile-name' from this repository. Updating..."
  - When creating a new profile, display: "Creating profile: profile-name"
  - After installation, display component count: "✓ Updated profile profile-name (5 skills, 3 agents, 2 commands)"
  - If components were overwritten, mention it: "Overwrote 2 existing components"
  - Messages use consistent formatting and symbols
  
  **Testing Criteria:**
  **Unit Tests:**
  - Message formatting helper functions
  - Component counting logic
  
  **Integration Tests:**
  - Verify correct messages displayed for update scenario
  - Verify correct messages displayed for new profile scenario
  - Verify component count is accurate
  
  **Component Browser Tests:**
  - N/A (CLI-only feature)

- [ ] Story-006: As a developer, I want the ProfileManager to expose methods for finding profiles by source URL so that the installation logic can detect duplicates.

  **Acceptance Criteria:**
  - ProfileManager has a `FindProfileBySourceURL(repoURL string) (*Profile, error)` method
  - ProfileManager has a `SaveProfileMetadata(profileName, sourceURL string) error` method
  - ProfileManager has a `LoadProfileMetadata(profileName string) (string, error)` method that returns the source URL
  - Methods handle errors gracefully (missing files, corrupt data, etc.)
  - Methods are well-documented with godoc comments
  
  **Testing Criteria:**
  **Unit Tests:**
  - FindProfileBySourceURL with matching profile
  - FindProfileBySourceURL with no matching profile
  - FindProfileBySourceURL with multiple profiles (returns first match)
  - SaveProfileMetadata creates correct file structure
  - LoadProfileMetadata reads URL correctly
  - Error handling for corrupt/missing metadata
  
  **Integration Tests:**
  - ProfileManager methods work end-to-end
  - Metadata persistence across process restarts
  
  **Component Browser Tests:**
  - N/A (internal API)

- [ ] Story-007: As a user, I want existing components preserved during updates so that my manually installed components are not deleted.

  **Acceptance Criteria:**
  - When updating a profile, existing component directories are scanned
  - New components from the repository are added
  - Components with the same name are overwritten (files replaced)
  - Components in the profile that are not in the repository are left untouched
  - No components are deleted during the update process
  - The system does not distinguish between manually added and repository components
  
  **Testing Criteria:**
  **Unit Tests:**
  - Component merge logic
  - Component overwrite behavior
  - Component preservation logic
  
  **Integration Tests:**
  - Update profile with new components → new components added
  - Update profile with modified components → components overwritten
  - Update profile with extra components → extra components preserved
  - Combined scenario with new, modified, and extra components
  
  **Component Browser Tests:**
  - N/A (CLI-only feature)

## Functional Requirements

- FR-1: The system SHALL normalize repository URLs by removing trailing slashes, `.git` suffixes, and converting SSH URLs to a canonical format before comparison
- FR-2: The system SHALL create a `.profile-metadata` file in the profile directory containing the normalized repository URL when a profile is created via `install all`
- FR-3: The system SHALL scan existing profiles for `.profile-metadata` files and compare normalized URLs to detect duplicates before creating a new profile
- FR-4: The system SHALL update the existing profile when a duplicate repository URL is detected, unless the `--profile` flag is specified
- FR-5: The system SHALL preserve existing components during profile updates and only overwrite components with matching names
- FR-6: The system SHALL display informative messages indicating whether a profile is being created or updated
- FR-7: The system SHALL accept an optional `--profile <custom-name>` flag to override duplicate detection and create a new profile with the specified name
- FR-8: The system SHALL handle legacy profiles without metadata files gracefully by treating them as non-duplicates
- FR-9: The ProfileManager SHALL expose `FindProfileBySourceURL`, `SaveProfileMetadata`, and `LoadProfileMetadata` methods for metadata operations
- FR-10: The BulkDownloader SHALL call `SaveProfileMetadata` after successfully downloading components to a profile

## Non-Goals (Out of Scope)

- Automatic cleanup of old duplicate profiles created before this feature
- Interactive prompt asking users to choose between update or create new
- Version tracking or rollback capabilities for profile updates
- Detection of outdated components or automatic update notifications
- Migration of existing profiles to add metadata retroactively
- Support for `--clean` flag to remove components not in the repository
- Diff view showing changes between old and new components
- Profile aliases or multiple names pointing to the same profile
- Metadata beyond the source URL (timestamps, commit hashes, etc.)
- Integration with Git to track repository state or branch information

## Implementation Notes

### File Structure

```
~/.agent-smith/profiles/
  ├── anthropics-skills/
  │   ├── .profile-metadata          # Contains: https://github.com/anthropics/skills
  │   ├── skills/
  │   ├── agents/
  │   └── commands/
  └── openai-cookbook/
      ├── .profile-metadata          # Contains: https://github.com/openai/cookbook
      ├── skills/
      └── agents/
```

### Metadata File Format

Simple plain text format with UTF-8 encoding:

```
https://github.com/anthropics/skills
```

### URL Normalization Rules

1. Convert to lowercase domain (github.com, gitlab.com, etc.)
2. Remove trailing slashes
3. Remove `.git` suffix
4. Convert SSH format `git@github.com:owner/repo` to `https://github.com/owner/repo`
5. Expand shorthand `owner/repo` to `https://github.com/owner/repo` (assume GitHub)
6. Preserve path case-sensitivity (owner/repo names)

### Backward Compatibility

- Existing profiles without `.profile-metadata` files will continue to work
- They will be treated as unique profiles during duplicate detection
- Next time they are updated (if possible), metadata will be added

### Error Handling

- If metadata file is corrupt or unreadable, treat profile as non-duplicate and log a warning
- If metadata file cannot be written during installation, show a warning but continue
- If `--profile` flag specifies an existing profile name, show error and abort

## Acceptance Criteria Summary

### Must Have (P0)
- ✅ Detect existing profiles by repository URL
- ✅ Update existing profile instead of creating duplicates
- ✅ Normalize repository URL variations
- ✅ Store repository URL in profile metadata
- ✅ Preserve existing components during updates
- ✅ Add `--profile` flag to override duplicate detection

### Should Have (P1)
- ✅ Clear user feedback messages
- ✅ ProfileManager API methods for metadata operations
- ✅ Backward compatibility with profiles without metadata

### Nice to Have (P2)
- ⬜ Display overwrite count in completion message
- ⬜ Validate metadata file integrity

## Testing Strategy

### Unit Tests
- URL normalization function with various input formats
- Metadata file read/write operations
- Profile detection logic
- Component merge behavior

### Integration Tests
- End-to-end install all workflow with existing profile
- End-to-end install all workflow with new profile
- Install all with --profile flag
- Legacy profile handling (no metadata)

### Manual Testing Checklist
- [ ] Install from GitHub HTTPS URL twice → should reuse profile
- [ ] Install from GitHub SSH URL after HTTPS → should reuse profile
- [ ] Install from shorthand format after full URL → should reuse profile
- [ ] Install with --profile flag → should create new profile
- [ ] Install with existing profile, add new component to repo → should add to profile
- [ ] Install with existing profile, modify component in repo → should overwrite
- [ ] Profile without metadata → should be treated as unique
- [ ] Invalid metadata file → should log warning and continue

## Rollout Plan

### Phase 1: Core Infrastructure (Week 1)
- Implement URL normalization utility
- Create ProfileMetadata struct and operations
- Add metadata methods to ProfileManager
- Unit tests for all new functions

### Phase 2: Integration (Week 2)
- Modify BulkDownloader to save metadata
- Update install all handler to detect existing profiles
- Implement component merge logic
- Add --profile flag support

### Phase 3: Polish & Testing (Week 3)
- Add informative user messages
- Integration tests
- Manual testing and bug fixes
- Documentation updates

### Phase 4: Release
- Code review
- Merge to main branch
- Release notes
- Update README with new behavior

## Success Metrics

- Zero duplicate profiles created when installing from the same repository URL
- User feedback indicates improved understanding of profile behavior
- No regression in existing installation workflows
- All tests passing (unit + integration)
