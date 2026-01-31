# PRD: Auto-Profile with Namespace Collision Handling

**Created**: 2026-01-31 14:37 UTC

---

## Introduction

Implement automatic profile generation and intelligent collision handling to prevent component naming conflicts when installing from multiple repositories. This feature enables users to safely install components from different sources without manual namespace management, while providing clear provenance and smart linking behavior based on an active profile concept.

## Goals

- Automatically generate unique profile names from repository sources
- Prevent component naming collisions across different repositories
- Provide seamless user experience with minimal manual profile management
- Enable smart linking with active profile priority
- Maintain clear component provenance (source tracking)
- Support profile-aware update and unlink operations

## User Stories

- [x] Story-001: As a user installing from a repository, I want the system to automatically create a uniquely named profile so that I don't have to manually manage namespaces.

  **Acceptance Criteria:**
  - Profile name is auto-generated from source URL with format `github-user-repo`
  - Profile name uniqueness is guaranteed through collision detection
  - Profile name follows shortest-unique algorithm (repo → user-repo → github-user-repo)
  - Local repository installs use directory name as profile name
  - Install output clearly shows which profile was created/used
  
  **Testing Criteria:**
  **Unit Tests:**
  - URL parsing logic for profile name generation
  - Collision detection algorithm validation
  - Profile name uniqueness guarantee tests
  
  **Integration Tests:**
  - End-to-end install with auto-profile creation
  - Multiple install operations with collision scenarios
  - Local path vs URL profile name generation
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

- [x] Story-002: As a first-time user, I want the first installed profile to automatically become active so that I can immediately start linking components without additional configuration.

  **Acceptance Criteria:**
  - First install sets the newly created profile as active
  - Active profile is persisted to `~/.agent-smith/.active-profile`
  - Subsequent installs do NOT change the active profile
  - Clear indicator shows which profile is active during install
  
  **Testing Criteria:**
  **Unit Tests:**
  - Active profile persistence logic tests
  - Active profile read/write file operations
  
  **Integration Tests:**
  - First install sets active profile correctly
  - Second install preserves existing active profile
  - Active profile state persists across CLI sessions
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

- [x] Story-003: As a user with multiple profiles, I want to manually switch the active profile so that I can control which profile's components are prioritized for linking.

  **Acceptance Criteria:**
  - `agent-smith profile switch <name>` command switches active profile
  - Command validates profile exists before switching
  - Clear confirmation message shows profile switch
  - Active profile persists across CLI sessions
  - Error message if profile doesn't exist
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile existence validation logic
  - Active profile update logic
  
  **Integration Tests:**
  - Profile switching end-to-end flow
  - Error handling for non-existent profiles
  - Persistence across multiple command invocations
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

- [x] Story-004: As a user with multiple profiles, I want to list all available profiles and see which is active so that I understand my current configuration.

  **Acceptance Criteria:**
  - `agent-smith profile list` command shows all profiles
  - Active profile is clearly marked with indicator (e.g., asterisk or arrow)
  - Output shows component counts per profile (agents, skills, commands)
  - Output is sorted alphabetically with active profile highlighted
  - Handles empty profiles gracefully
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile enumeration logic
  - Component counting per profile
  - Active profile marking logic
  
  **Integration Tests:**
  - List command with multiple profiles
  - List command with empty profiles
  - Active profile indication accuracy
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

- [ ] Story-005: As a user, I want to see which profile a specific component belongs to so that I understand where it came from and can make informed linking decisions.

  **Acceptance Criteria:**
  - `agent-smith info <component>` command shows component metadata
  - Output includes profile name, source URL, install date, and link status
  - Command searches across all profiles to find component
  - If component exists in multiple profiles, shows all occurrences
  - Clear formatting distinguishes metadata fields
  
  **Testing Criteria:**
  **Unit Tests:**
  - Component metadata retrieval logic
  - Multi-profile search logic
  - Metadata formatting logic
  
  **Integration Tests:**
  - Info command for single-profile component
  - Info command for multi-profile component
  - Info command for non-existent component
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

- [ ] Story-006: As a user linking a component, I want the system to automatically link from the active profile if available so that I don't need to specify the profile for common cases.

  **Acceptance Criteria:**
  - `agent-smith link <component>` checks active profile first
  - If component exists in active profile, auto-links without prompt
  - If component doesn't exist in active profile, searches other profiles
  - If found in other profiles, presents interactive prompt for selection
  - If not found anywhere, clear error message is displayed
  - Link output shows which profile the component was linked from
  
  **Testing Criteria:**
  **Unit Tests:**
  - Active profile priority logic
  - Multi-profile search logic
  - Profile selection prompt logic
  
  **Integration Tests:**
  - Link from active profile (auto-link scenario)
  - Link with prompt (multi-profile scenario)
  - Link failure (component not found scenario)
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

- [ ] Story-007: As a user with component collisions across profiles, I want an interactive prompt to choose which profile to link from so that I maintain control over component selection.

  **Acceptance Criteria:**
  - Interactive prompt lists all profiles containing the component
  - Prompt shows profile name and component source URL for each option
  - User can select by number (1, 2, 3, etc.)
  - Selection is validated before proceeding with link
  - User can cancel the prompt (Ctrl+C or empty input)
  - Active profile option is clearly indicated in the list
  
  **Testing Criteria:**
  **Unit Tests:**
  - Prompt generation logic
  - Selection validation logic
  - Cancellation handling logic
  
  **Integration Tests:**
  - Interactive prompt with multiple options
  - Selection validation and error handling
  - Cancellation behavior
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

- [ ] Story-008: As a user, I want to explicitly specify a profile when linking so that I can bypass the active profile priority when needed.

  **Acceptance Criteria:**
  - `agent-smith link <component> --profile <name>` explicitly links from specified profile
  - Bypasses active profile priority and collision prompts
  - Validates that component exists in specified profile
  - Clear error if component not found in specified profile
  - Clear error if profile doesn't exist
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile-specific link logic
  - Profile and component existence validation
  
  **Integration Tests:**
  - Explicit profile link success scenario
  - Explicit profile link with non-existent profile
  - Explicit profile link with non-existent component
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

- [ ] Story-009: As a user with multiple components in my active profile, I want to bulk link all of them so that I don't have to link each component individually.

  **Acceptance Criteria:**
  - `agent-smith link --all` links all components from active profile
  - Output shows progress for each component being linked
  - Handles partial failures gracefully (continues linking remaining components)
  - Summary shows total linked, skipped, and failed counts
  - Clear error messages for any failed links
  
  **Testing Criteria:**
  **Unit Tests:**
  - Bulk link enumeration logic
  - Error handling for partial failures
  - Summary generation logic
  
  **Integration Tests:**
  - Bulk link all components successfully
  - Bulk link with some failures (partial success)
  - Bulk link with no active profile set
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

- [ ] Story-010: As a user updating components, I want to update only components in my active profile so that I can safely update without affecting other profiles.

  **Acceptance Criteria:**
  - `agent-smith update` updates all components in active profile only
  - Output clearly shows which profile is being updated
  - Checks source repository for updates for each component
  - Updates component files and lock file metadata (commit hash, timestamps)
  - Skips components with no updates available
  - Summary shows updated, unchanged, and failed counts
  
  **Testing Criteria:**
  **Unit Tests:**
  - Active profile component enumeration
  - Update detection logic
  - Lock file update logic
  
  **Integration Tests:**
  - Update all in active profile with available updates
  - Update all with no updates available
  - Update with no active profile set (error scenario)
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

- [ ] Story-011: As a user unlinking components, I want to unlink only from the active profile by default so that I maintain control over which profile's components are affected.

  **Acceptance Criteria:**
  - `agent-smith unlink <component>` unlinks from active profile by default
  - If component doesn't exist in active profile, searches other profiles and prompts
  - `--profile` flag allows explicit profile specification
  - Clear confirmation of which profile the component was unlinked from
  - Error handling for component not found scenarios
  
  **Testing Criteria:**
  **Unit Tests:**
  - Active profile unlink logic
  - Multi-profile search for unlink
  - Profile-specific unlink validation
  
  **Integration Tests:**
  - Unlink from active profile successfully
  - Unlink with prompt (component in other profiles)
  - Explicit profile unlink with --profile flag
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

- [ ] Story-012: As a user viewing my installed components, I want the list grouped by profile with the active profile highlighted so that I can easily see my component organization.

  **Acceptance Criteria:**
  - `agent-smith list` command groups components by profile
  - Active profile is marked with clear indicator (*, →, or [ACTIVE])
  - Each profile section shows profile name and component count
  - Components within each profile are sorted alphabetically
  - Empty profiles are shown with "(empty)" indicator
  - `--profile` flag filters to show only specified profile
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile grouping logic
  - Active profile marking logic
  - Component sorting within profiles
  
  **Integration Tests:**
  - List all components across profiles
  - List with --profile filter
  - List with no components installed
  
  **Component Browser Tests:**
  - Not applicable (CLI-only feature)

## Functional Requirements

**Profile Name Generation:**

- FR-1: The system SHALL parse source URLs to extract platform, user/org, and repository name
- FR-2: The system SHALL generate profile names following the pattern: `{platform}-{user}-{repo}`
- FR-3: The system SHALL implement shortest-unique algorithm: try `{repo}` first, then `{user}-{repo}`, then `{platform}-{user}-{repo}` if collisions detected
- FR-4: The system SHALL use directory name as profile name for local path installations
- FR-5: The system SHALL validate profile name uniqueness before creating new profile

**Active Profile Management:**

- FR-6: The system SHALL store active profile in `~/.agent-smith/.active-profile` file
- FR-7: The system SHALL set first installed profile as active automatically
- FR-8: The system SHALL preserve existing active profile on subsequent installs
- FR-9: The system SHALL validate active profile exists before operations that depend on it
- FR-10: The system SHALL provide clear error when no active profile is set for profile-dependent operations

**Profile Commands:**

- FR-11: The system SHALL implement `agent-smith profile use <name>` to switch active profile
- FR-12: The system SHALL implement `agent-smith profile list` to show all profiles with counts
- FR-13: The system SHALL implement `agent-smith info <component>` to show component metadata and profile information
- FR-14: The system SHALL clearly indicate active profile in all list/info outputs

**Linking with Collision Handling:**

- FR-15: The system SHALL prioritize active profile when linking components
- FR-16: The system SHALL search all profiles if component not found in active profile
- FR-17: The system SHALL present interactive prompt when component exists in multiple profiles
- FR-18: The system SHALL support `--profile` flag to explicitly specify profile for linking
- FR-19: The system SHALL support `--all` flag to bulk link all components from active profile
- FR-20: The system SHALL display source profile name in link operation output

**Update Command:**

- FR-21: The system SHALL update only components in active profile by default
- FR-22: The system SHALL check source repository for updates for each component
- FR-23: The system SHALL update lock file metadata (commit hash, timestamps) after successful updates
- FR-24: The system SHALL provide summary of updated, unchanged, and failed components

**Unlink Command:**

- FR-25: The system SHALL unlink components from active profile by default
- FR-26: The system SHALL search other profiles if component not in active profile and present prompt
- FR-27: The system SHALL support `--profile` flag for explicit profile specification during unlink

**List Command Enhancements:**

- FR-28: The system SHALL group components by profile in list output
- FR-29: The system SHALL mark active profile with visual indicator
- FR-30: The system SHALL support `--profile` flag to filter list to specific profile
- FR-31: The system SHALL show component counts per profile

## Non-Goals (Out of Scope)

- No automatic profile merging or component deduplication
- No profile deletion on empty (profiles remain even when all components uninstalled)
- No automatic profile switching based on working directory or context
- No profile-based access control or permissions
- No profile import/export functionality (future enhancement)
- No profile renaming command (future enhancement)
- No git-style profile branching or forking
- No rollback/undo for profile operations
- No profile-level configuration or settings beyond active state
- No multi-profile installation (install to multiple profiles at once)

## Technical Implementation Notes

### Profile Name Generation Algorithm

```
func GenerateProfileName(sourceURL string) string {
    platform, user, repo := parseURL(sourceURL)
    
    // Try shortest name first
    candidates := []string{
        repo,
        fmt.Sprintf("%s-%s", user, repo),
        fmt.Sprintf("%s-%s-%s", platform, user, repo),
    }
    
    for _, candidate := range candidates {
        if !profileExists(candidate) {
            return candidate
        }
    }
    
    // If all exist, append number
    return fmt.Sprintf("%s-%s-%s-%d", platform, user, repo, getNextNumber())
}
```

### Active Profile File Format

File: `~/.agent-smith/.active-profile`

Content: Single line containing profile name (no newlines or additional data)

Example: `github-anomalyco-opencode-plugins`

### Interactive Prompt Format

```
⚠️  Component "api-testing" found in multiple profiles:

  1. github-anomalyco-repo-1 (active)
     Source: https://github.com/anomalyco/repo-1
     
  2. github-user-repo-2
     Source: https://github.com/user/repo-2

Select profile to link from [1-2]: _
```

### Profile List Output Format

```
$ agent-smith profile list

Available Profiles:

→ github-anomalyco-opencode-plugins (active)
  Skills: 45 | Agents: 12 | Commands: 8
  
  github-user-custom-extensions
  Skills: 3 | Agents: 1 | Commands: 2
  
  local-workspace-plugins
  Skills: 0 | Agents: 0 | Commands: 1
```

### Component Info Output Format

```
$ agent-smith info api-testing

Component: api-testing (skill)
Profile: github-anomalyco-opencode-plugins (active)
Source: https://github.com/anomalyco/opencode-plugins
Installed: 2026-01-29 14:30 UTC
Updated: 2026-01-31 10:15 UTC
Status: Linked to opencode, claudecode
```

## Dependencies

- Existing profile infrastructure (`~/.agent-smith/profiles/`)
- Existing lock file system (`.agent-lock.json`, `.skill-lock.json`, etc.)
- Existing detector/downloader/linker modules
- Cobra CLI framework for command implementation

## Success Metrics

- Users can install from multiple repositories without manual namespace management
- Zero component naming collisions across different source repositories
- Active profile behavior is intuitive (first install sets active, subsequent installs preserve)
- Linking workflow requires minimal user input for common cases (active profile auto-link)
- Profile switching is simple and discoverable
- Component provenance is always clear (users know which repo a component came from)

## Migration Notes

Since backward compatibility is not a requirement:

- Fresh installations will automatically use the new profile-per-source model
- Existing users can migrate by re-installing components (old installations will be replaced)
- No migration script needed - users can manually `uninstall` old components and `install` with new system
- Lock files will be recreated with profile-aware structure on fresh installs
