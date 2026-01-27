# PRD: Enhanced Profile Management for Agent-Smith

## Introduction

Enhance the existing profiles feature in agent-smith to support creating profiles, adding/removing components (skills, agents, commands) to profiles, installing components directly to profiles, and making the link commands profile-aware. This builds upon the existing profile activation/deactivation system to provide a complete profile management workflow.

## Goals

- Enable users to create new empty profiles via CLI
- Allow copying existing components from ~/.agents/ into specific profiles
- Support installing components directly to profiles (bypassing ~/.agents/)
- Make link commands automatically use active profile as source when applicable
- Provide component management within profiles (add, remove, list)
- Maintain backward compatibility with existing commands and workflows
- Ensure profile operations are safe with proper validation and error handling

## User Stories

- [ ] Story-001: As a developer, I want to create a new empty profile so that I can organize my tools for different projects.

  **Acceptance Criteria:**
  - Command `agent-smith profiles create <name>` creates new profile
  - Profile name validation allows alphanumeric characters and dashes only
  - Creates directory structure at ~/.agents/profiles/<name>/{agents,skills,commands}
  - Error if profile name already exists
  - Error if profile name is invalid (empty, contains invalid characters, path traversal)
  - Success message shows created profile location
  - Empty profiles are considered valid
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile name validation logic (valid/invalid names)
  - Directory creation with proper permissions
  - Duplicate profile detection
  
  **Integration Tests:**
  - End-to-end profile creation workflow
  - Profile validation in ProfileManager
  - File system state after creation
  
  **Component Browser Tests:**
  - N/A (CLI-only feature)

- [ ] Story-002: As a developer, I want to add an existing component from ~/.agents/ to a profile so that I can reuse components across profiles.

  **Acceptance Criteria:**
  - Commands: `agent-smith profiles add skill <profile> <name>`, `agent-smith profiles add agent <profile> <name>`, `agent-smith profiles add command <profile> <name>`
  - Validates profile exists before copying
  - Validates source component exists in ~/.agents/<type>/
  - Copies entire component directory to profile directory
  - Error and abort if component already exists in profile (no overwrite)
  - Success message shows what was copied and where
  - Source component in ~/.agents/ remains unchanged
  
  **Testing Criteria:**
  **Unit Tests:**
  - Component copy logic validation
  - Duplicate component detection
  - Source validation
  
  **Integration Tests:**
  - Copy component from ~/.agents/ to profile
  - Verify directory structure after copy
  - Error handling for missing components
  
  **Component Browser Tests:**
  - N/A (CLI-only feature)

- [ ] Story-003: As a developer, I want to remove a component from a profile so that I can keep profiles clean and focused.

  **Acceptance Criteria:**
  - Commands: `agent-smith profiles remove skill <profile> <name>`, `agent-smith profiles remove agent <profile> <name>`, `agent-smith profiles remove command <profile> <name>`
  - Validates profile exists
  - Validates component exists in profile
  - Removes component directory from profile
  - Confirmation prompt before deletion (always confirm)
  - Can skip confirmation with --force flag
  - Error if component doesn't exist in profile
  - Success message shows what was removed
  
  **Testing Criteria:**
  **Unit Tests:**
  - Component removal logic
  - Confirmation prompt logic
  - Force flag handling
  
  **Integration Tests:**
  - Remove component from profile
  - Verify directory structure after removal
  - Confirmation workflow testing
  
  **Component Browser Tests:**
  - N/A (CLI-only feature)

- [ ] Story-004: As a developer, I want to install a component directly to a profile so that I can skip installing to ~/.agents/ first.

  **Acceptance Criteria:**
  - Add --profile flag to install commands: `agent-smith install skill <repo> <name> --profile <profile>`
  - Works for skill, agent, and command install commands
  - Validates profile exists before installing
  - Downloads and installs directly to ~/.agents/profiles/<profile>/<type>/<name>
  - Does not create component in ~/.agents/<type>/
  - Error if profile doesn't exist
  - Success message shows installation to profile
  - Maintains same download and detection behavior as regular installs
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile parameter handling in downloaders
  - Destination path logic with profile
  - Profile validation before download
  
  **Integration Tests:**
  - Install component to profile
  - Verify component only exists in profile, not in ~/.agents/
  - Error handling for invalid profiles
  
  **Component Browser Tests:**
  - N/A (CLI-only feature)

- [ ] Story-005: As a developer, I want the link commands to automatically use my active profile so that linking matches my current context.

  **Acceptance Criteria:**
  - When profile is active, link commands source from profile instead of ~/.agents/
  - When no profile is active, link commands use ~/.agents/ (current behavior)
  - Works for all link commands: skill, agent, command, all, auto
  - Link status and list commands show source (profile name or base)
  - Always display whether linking from profile or base directory
  - No breaking changes to existing link command syntax
  - Profile awareness is automatic, no new flags required
  
  **Testing Criteria:**
  **Unit Tests:**
  - Source directory determination logic
  - Profile-aware path resolution
  - Active profile detection
  
  **Integration Tests:**
  - Link with active profile
  - Link without active profile
  - Verify correct source directories used
  - Status display shows correct source
  
  **Component Browser Tests:**
  - N/A (CLI-only feature)

- [ ] Story-006: As a developer, I want to delete a profile so that I can remove profiles I no longer need.

  **Acceptance Criteria:**
  - Command: `agent-smith profiles delete <name>`
  - Validates profile exists
  - Cannot delete active profile (must deactivate first)
  - Always prompts for confirmation before deletion
  - Can skip confirmation with --force flag
  - Removes entire profile directory
  - Success message shows what was deleted
  - Error with helpful message if trying to delete active profile
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile deletion logic
  - Active profile protection
  - Confirmation workflow
  
  **Integration Tests:**
  - Delete inactive profile
  - Attempt to delete active profile (should fail)
  - Deactivate then delete workflow
  
  **Component Browser Tests:**
  - N/A (CLI-only feature)

- [ ] Story-007: As a developer, I want to see detailed information about a profile so that I know what's in it before activating.

  **Acceptance Criteria:**
  - Command: `agent-smith profiles show <name>`
  - Displays profile name and path
  - Lists all agents with count
  - Lists all skills with count
  - Lists all commands with count
  - Shows whether profile is currently active
  - Error if profile doesn't exist
  - Clear, formatted output
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile info gathering logic
  - Component listing logic
  - Active status detection
  
  **Integration Tests:**
  - Show profile with components
  - Show empty profile
  - Show active vs inactive profiles
  
  **Component Browser Tests:**
  - N/A (CLI-only feature)

- [ ] Story-008: As a developer, I want profile name validation to prevent file system issues so that profiles always work correctly.

  **Acceptance Criteria:**
  - Profile names must match pattern: ^[a-zA-Z0-9-]+$
  - Allow letters (upper and lowercase), numbers, and dashes
  - Reject empty names
  - Reject names with spaces, special characters, or path separators
  - Reject names starting with . (hidden directories)
  - Reject path traversal attempts (../, ./, etc.)
  - Clear error messages explaining validation failures
  - Consistent validation across all profile commands
  
  **Testing Criteria:**
  **Unit Tests:**
  - Profile name validation function
  - Test valid and invalid name patterns
  - Edge cases (empty, special chars, path traversal)
  
  **Integration Tests:**
  - Attempt to create profiles with invalid names
  - Verify validation in all commands (create, add, remove, etc.)
  
  **Component Browser Tests:**
  - N/A (CLI-only feature)

## Functional Requirements

- FR-001: Implement CreateProfile(name string) in ProfileManager to create empty profile directories
- FR-002: Implement AddComponentToProfile(profileName, componentType, componentName string) to copy components
- FR-003: Implement RemoveComponentFromProfile(profileName, componentType, componentName string) to remove components
- FR-004: Implement DeleteProfile(name string) to remove entire profile with safety checks
- FR-005: Implement GetProfileInfo(name string) to retrieve detailed profile information
- FR-006: Add validateProfileName(name string) helper with strict validation rules
- FR-007: Add CLI commands under 'profiles' parent: create, add, remove, delete, show
- FR-008: Add --profile flag to install skill/agent/command commands
- FR-009: Modify downloader functions to accept optional profileName parameter
- FR-010: Modify linker functions to check active profile and use profile directories when applicable
- FR-011: Update link status/list commands to display source (profile or base)
- FR-012: Add confirmation prompts for destructive operations (remove, delete)
- FR-013: Add --force flag to skip confirmation prompts
- FR-014: Ensure all profile operations validate profile existence first
- FR-015: Ensure component operations validate component existence in source location
- FR-016: Update main.go handlers to wire up new ProfileManager methods
- FR-017: Follow existing code patterns from linker and profiles packages
- FR-018: Maintain cross-platform compatibility using fileutil helpers
- FR-019: Provide clear, informative error messages for all failure cases
- FR-020: Allow empty profiles to be created and activated

## Non-Goals (Out of Scope)

- No profile templates or presets in initial version
- No --from-current flag to clone current state (future enhancement)
- No profile sharing or export/import functionality
- No profile versioning or history tracking
- No component update/upgrade operations within profiles (use regular update command)
- No bulk operations (adding multiple components at once)
- No profile renaming command (can add later)
- No profile description or metadata fields
- No integration with git or version control for profiles
- No remote profile repositories or sharing
- No profile inheritance or composition
- No profile-specific configuration or settings
- No changes to existing profile activation/deactivation logic
- No changes to component detection or download logic (except destination path)

## Technical Implementation Notes

### Files to Modify

1. **pkg/profiles/manager.go**
   - Add: CreateProfile, AddComponentToProfile, RemoveComponentFromProfile, DeleteProfile, GetProfileInfo
   - Add: validateProfileName helper
   - Use fileutil.CopyDirectoryContents for copying components
   - Use fileutil.CreateDirectoryWithPermissions for creating directories

2. **pkg/profiles/profiles.go**
   - Add: ProfileInfo struct
   - Add: ListComponents method
   - Modify: IsValid to allow empty profiles

3. **cmd/root.go**
   - Add subcommands: profilesCreateCmd, profilesAddCmd, profilesRemoveCmd, profilesDeleteCmd, profilesShowCmd
   - Add --profile flag to installSkillCmd, installAgentCmd, installCommandCmd
   - Add --force flag to profilesRemoveCmd and profilesDeleteCmd

4. **main.go**
   - Add handlers: handleProfilesCreate, handleProfilesAdd, handleProfilesRemove, handleProfilesDelete, handleProfilesShow
   - Modify install handlers to accept optional profile parameter
   - Pass ProfileManager to linker functions

5. **internal/linker/linker.go**
   - Add: getComponentSourceDir helper
   - Modify: Link, LinkAll, LinkType to be profile-aware
   - Add ProfileManager parameter to linker functions

6. **internal/linker/status.go**
   - Modify: ListLinks to display profile source
   - Modify: ShowLinkStatus to show active profile info

7. **internal/downloader/skill.go, agent.go, command.go**
   - Add: profileName parameter to download functions
   - Modify: destination path logic to check profile parameter

### Validation Rules

Profile names must:
- Match regex: ^[a-zA-Z0-9-]+$
- Not be empty
- Not start with . or -
- Not contain path separators (/, \)
- Not be . or ..

### Error Handling

- Profile doesn't exist: "Profile '<name>' not found"
- Profile already exists: "Profile '<name>' already exists"
- Invalid profile name: "Invalid profile name '<name>': must contain only letters, numbers, and dashes"
- Component doesn't exist in source: "Component '<type>/<name>' not found in ~/.agents/"
- Component already exists in profile: "Component '<type>/<name>' already exists in profile '<profile>'"
- Cannot delete active profile: "Cannot delete active profile '<name>'. Please deactivate it first."

### Backward Compatibility

All existing commands work unchanged:
- `agent-smith install skill <repo> <name>` - Still installs to ~/.agents/
- `agent-smith link skill` - Uses active profile if present, otherwise ~/.agents/
- `agent-smith profiles list|activate|deactivate` - No changes
- Empty profiles now valid (change from current behavior)

## Success Criteria

- All 8 user stories implemented and tested
- Unit tests added for new ProfileManager methods
- Integration tests for complete workflows
- Manual testing checklist passes
- Documentation updated with new commands
- All existing tests continue to pass
- No breaking changes to existing functionality
- Error messages are clear and actionable
