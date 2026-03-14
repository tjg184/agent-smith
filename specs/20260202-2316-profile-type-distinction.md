# PRD: Profile Type Distinction and Visual Indicators

**Created**: 2026-02-02 23:16 UTC

---

## Introduction

Agent-smith currently creates profiles both automatically (when installing from repositories) and manually (when users create custom profiles). This dual-purpose system lacks clear visual distinction, making it confusing for users to understand which profiles are repo-sourced namespaces versus user-created collections. This PRD outlines improvements to clearly distinguish profile types while maintaining current functionality.

## Problem Statement

Currently, profiles serve two distinct purposes without clear differentiation:
1. **Repository Namespaces**: Auto-created during `install all`, tied to source repositories, used for easy repo-wide updates
2. **User Collections**: Manually created via `profiles create`, used for organizing and cherry-picking components across repos

Users cannot easily tell these apart when listing profiles, leading to confusion about profile purpose and management strategies.

## Goals

- Clearly distinguish repo-sourced profiles from user-created profiles in all UI interactions
- Maintain backward compatibility with existing profile functionality
- Enhance profile metadata to track profile type and creation source
- Improve user experience with visual indicators (emojis) for profile types
- Add filtering capabilities to view profiles by type
- Update documentation and help text to clarify the dual-purpose nature of profiles

## User Stories

- [ ] Story-001: As a user, I want to see visual indicators distinguishing repo-sourced profiles from user-created profiles so that I understand each profile's purpose at a glance.

  **Acceptance Criteria:**
  - Profile listings show 📦 emoji for repo-sourced profiles
  - Profile listings show 👤 emoji for user-created profiles
  - Source URL is displayed for repo-sourced profiles
  - Visual format is consistent across all profile list commands
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile type detection returns correct type based on metadata
  - Emoji selection logic returns correct emoji for each profile type
  
  **Integration Tests:**
  - Profile list command displays correct emojis for mixed profile types
  - Source URL only appears for repo-sourced profiles

- [ ] Story-002: As a developer, I want profile metadata to include an explicit type field so that the system can reliably distinguish between repo-sourced and user-created profiles.

  **Acceptance Criteria:**
  - ProfileMetadata struct includes Type field with values "repo" or "user"
  - CreateProfile() saves metadata with type="user"
  - CreateProfileWithMetadata() saves metadata with type="repo"
  - LoadProfileMetadata() correctly reads and returns profile type
  - Profile type is persisted in .profile-metadata JSON file
  
  **Testing Criteria:**
  **Unit Tests:**
  - ProfileMetadata serialization includes type field
  - CreateProfile saves correct user type
  - CreateProfileWithMetadata saves correct repo type
  - LoadProfileMetadata parses type field correctly
  
  **Integration Tests:**
  - Created profiles persist type across application restarts
  - Profile type survives metadata file read/write cycles

- [ ] Story-003: As a user, I want to filter profile listings by type so that I can focus on either my repository installations or my custom collections.

  **Acceptance Criteria:**
  - profiles list command accepts --type flag
  - --type repo shows only repo-sourced profiles
  - --type user shows only user-created profiles
  - No --type flag shows all profiles (default behavior)
  - Clear error message if invalid type value provided
  - Filter respects existing profile listing format and sorting
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile type filtering logic correctly filters by type
  - Invalid type values are rejected with clear error
  
  **Integration Tests:**
  - profiles list --type repo shows only repo profiles
  - profiles list --type user shows only user profiles
  - profiles list with no flag shows all profiles
  - Filtered lists maintain correct emoji indicators

- [ ] Story-004: As a user, I want install all command to create profiles marked as repo-sourced so that I can distinguish them from my manually created profiles.

  **Acceptance Criteria:**
  - install all automatically creates profile with type="repo"
  - Profile metadata includes normalized source URL
  - Profile name is generated from repository URL
  - Existing profile reuse logic still works (finds by source URL)
  - Success message indicates profile was created from repository
  
  **Testing Criteria:**
  **Unit Tests:**
  - GenerateProfileNameFromRepo produces valid profile names
  - URL normalization handles various URL formats
  
  **Integration Tests:**
  - install all creates profile with correct type and metadata
  - Installing from same repo twice reuses existing profile
  - Profile metadata persists source URL correctly

- [ ] Story-005: As a user, I want profiles create command to create profiles marked as user-created so that I can manage my custom component collections separately from repository installations.

  **Acceptance Criteria:**
  - profiles create command creates profile with type="user"
  - No source URL is saved for user-created profiles
  - User-created profiles support all existing operations (add, remove, activate, etc.)
  - Success message clarifies this is a user-created profile
  
  **Testing Criteria:**
  **Unit Tests:**
  - CreateProfile saves metadata with type="user"
  - User profiles do not include source URL in metadata
  
  **Integration Tests:**
  - profiles create makes profile with correct type
  - User profiles support cherry-picking components
  - User profiles can be activated and deactivated

- [ ] Story-006: As a user, I want updated help text and documentation that explains the dual-purpose nature of profiles so that I understand when to use each type.

  **Acceptance Criteria:**
  - profiles command help describes repo-sourced vs user-created profiles
  - install all help mentions automatic profile creation
  - profiles create help clarifies purpose of user-created profiles
  - profiles list help documents --type flag options
  - Examples show both profile types in action
  
  **Testing Criteria:**
  **Unit Tests:**
  - N/A (documentation only)
  
  **Integration Tests:**
  - Help text is accessible via --help flag
  - Examples in help text are accurate and runnable

## Functional Requirements

- FR-1: The system SHALL add a Type field to ProfileMetadata struct with valid values "repo" or "user"
- FR-2: The CreateProfile function SHALL save profile metadata with type="user"
- FR-3: The CreateProfileWithMetadata function SHALL save profile metadata with type="repo"
- FR-4: The LoadProfileMetadata function SHALL correctly parse and return the type field
- FR-5: The profiles list command SHALL display 📦 emoji for repo-sourced profiles
- FR-6: The profiles list command SHALL display 👤 emoji for user-created profiles
- FR-7: The profiles list command SHALL display source URL for repo-sourced profiles only
- FR-8: The profiles list command SHALL accept --type flag with values "repo", "user", or no value
- FR-9: The --type repo flag SHALL filter to show only repo-sourced profiles
- FR-10: The --type user flag SHALL filter to show only user-created profiles
- FR-11: Invalid --type values SHALL produce clear error messages
- FR-12: The install all command SHALL create profiles with type="repo" and save source URL
- FR-13: The profiles create command SHALL create profiles with type="user" and no source URL
- FR-14: All profile management operations (activate, deactivate, delete, copy, etc.) SHALL work identically for both profile types
- FR-15: Profile type SHALL persist across application restarts via .profile-metadata file
- FR-16: Help text for profiles command SHALL explain repo-sourced vs user-created distinction
- FR-17: Help text for install all SHALL mention automatic profile creation
- FR-18: Help text for profiles create SHALL clarify user-created profile purpose

## Technical Implementation Notes

### Files to Modify

1. **pkg/profiles/manager.go**
   - Line 24-27: Add Type field to ProfileMetadata struct
   - Line 46-74: Update SaveProfileMetadata() to include type
   - Line 77-101: Update LoadProfileMetadata() to parse type field
   - Line 979-1024: Update CreateProfile() to save type="user"
   - Line 1026-1042: Update CreateProfileWithMetadata() to save type="repo"
   - Add GetProfileType(profileName string) (string, error) helper

2. **main.go**
   - Profile listing logic (around line 1490-1650):
     - Load metadata for each profile
     - Display emoji indicators based on type
     - Show source URL for repo types only
     - Implement --type flag filtering

3. **cmd/root.go**
   - Update profilesCmd Long description to explain profile types
   - Add --type flag to profiles list subcommand
   - Update install all command help text
   - Update profiles create command help text

### ProfileMetadata Structure

```go
type ProfileMetadata struct {
    Type      string `json:"type"`       // "repo" or "user"
    SourceURL string `json:"source_url"` // Only for type="repo"
}
```

### Profile List Output Format

```
Available Profiles:
  📦 example-skills (https://github.com/example/skills)
     3 skills
     
  📦 owner-repo (https://github.com/owner/repo)
     2 agents, 1 skill
     
  👤 my-work
     5 components (2 agents, 3 skills)
     
  👤 personal
     1 skill
```

## Non-Goals (Out of Scope)

- No migration logic for existing profiles (user will recreate)
- No "repo" subcommand aliases (keeping simple with --type flags only)
- No changes to profile functionality beyond visual distinction and metadata
- No automatic type detection or inference for profiles without metadata
- No changes to profile activation/deactivation behavior
- No changes to component installation workflows beyond metadata tracking
- No additional metadata fields beyond Type (keeping minimal)
- No profile renaming or type conversion features
- No analytics or usage tracking of profile types
- No changes to profile directory structure or file layout

## Success Metrics

- Users can visually distinguish profile types in list output
- Profile metadata correctly tracks type for new profiles
- Filtering by type works reliably
- Help text clearly explains the two use cases
- No breaking changes to existing profile workflows
- All existing tests pass with new metadata field

## Testing Strategy

### Unit Tests
- Profile metadata serialization/deserialization with type field
- Profile type detection logic
- Emoji selection based on profile type
- Type filtering logic
- URL normalization for repo profiles

### Integration Tests
- End-to-end profile creation with correct type
- Profile listing with emoji indicators
- Type filtering in profile list command
- Metadata persistence across restarts
- install all creates repo-typed profiles
- profiles create creates user-typed profiles

## Dependencies

- Existing profile infrastructure (pkg/profiles)
- Existing installation workflows (main.go)
- Cobra command framework (cmd/root.go)

## Risks and Mitigations

**Risk**: Emoji rendering issues in some terminals
**Mitigation**: Emojis are widely supported in modern terminals; fallback not needed based on user preference

**Risk**: Confusion during transition period with mixed old/new profiles
**Mitigation**: User will recreate all profiles, avoiding migration complexity

## Documentation Updates Required

- Update profiles command help text
- Update install all command help text
- Update profiles create command help text
- Add --type flag documentation to profiles list
- Update README.md with profile type explanation (if applicable)

---

## Appendix: Example Usage

### Creating and Viewing Different Profile Types

```bash
# Install from repository (creates repo-sourced profile)
$ agent-smith install all https://github.com/example/skills
Creating profile: example-skills
📦 Profile 'example-skills' created from repository

# Create user profile
$ agent-smith profiles create my-work
Creating profile 'my-work'...
👤 Profile 'my-work' created

# List all profiles
$ agent-smith profiles list
Available Profiles:
  📦 example-skills (https://github.com/example/skills)
     3 skills
     
  👤 my-work
     0 components

# Filter by type
$ agent-smith profiles list --type repo
Repository Profiles:
  📦 example-skills (https://github.com/example/skills)
     3 skills

$ agent-smith profiles list --type user
User Profiles:
  👤 my-work
     0 components
```
