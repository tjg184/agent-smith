# PRD: Update Profile-Builder Skill for Profile-Based Architecture

**Created**: 2026-02-01 03:23 UTC

---

## Introduction

The profile-builder skill needs to be updated to work with agent-smith's current profile-based architecture. Agent-smith now stores components in `~/.agent-smith/profiles/<profile-name>/` directories rather than the base `~/.agent-smith/` directories. Additionally, new commands (`profile copy` and `profile cherry-pick`) have been added that this skill should leverage.

The skill currently assumes components live in `~/.agent-smith/skills/`, `~/.agent-smith/agents/`, and `~/.agent-smith/commands/`, which are now typically empty. The skill must be updated to:
1. Scan profiles for available components
2. Use new agent-smith commands for copying components
3. Maintain its value proposition of intelligent, template-driven automation
4. Differentiate itself from the manual `cherry-pick` command

## Goals

- Update component scanner to scan profiles instead of base directories
- Leverage new `agent-smith profile copy` command for copying components between profiles
- Document the relationship between profile-builder and cherry-pick commands
- Maintain template-driven intelligent automation as the key value proposition
- Ensure all examples and documentation reflect profile-based architecture
- Maximize use of agent-smith commands while preserving skill functionality

## User Stories

- [ ] Story-001: As a developer, I want the component scanner to discover components from all installed profiles so I can see everything available across my profile installations.

  **Acceptance Criteria:**
  - `scan_skills()` function scans `~/.agent-smith/profiles/*/skills/*` instead of `~/.agent-smith/skills/`
  - `scan_agents()` function scans `~/.agent-smith/profiles/*/agents/*` instead of `~/.agent-smith/agents/`
  - `scan_commands()` function scans `~/.agent-smith/profiles/*/commands/*` instead of `~/.agent-smith/commands/`
  - Scanner returns deduplicated component names across all profiles
  - Scanner handles empty profiles gracefully without errors
  
  **Testing Criteria:**
  **Unit Tests:**
  - Test scanner with multiple profiles containing overlapping components
  - Test scanner with empty profiles directory
  - Test scanner with profiles containing no components
  
  **Integration Tests:**
  - Test scanner against actual ~/.agent-smith/profiles structure
  - Verify deduplication works correctly with same component in multiple profiles
  
  **Component Browser Tests:**
  - Not applicable (bash script functionality)

- [ ] Story-002: As a developer, I want to know which profile contains each component so I can track component sources when creating new profiles.

  **Acceptance Criteria:**
  - New `find_profiles_with_skill()` function returns list of profiles containing a specific skill
  - New `find_profiles_with_agent()` function returns list of profiles containing a specific agent
  - New `find_profiles_with_command()` function returns list of profiles containing a specific command
  - New `list_all_profiles()` function calls `agent-smith profile list` to show all profiles
  - Functions handle components that exist in multiple profiles
  
  **Testing Criteria:**
  **Unit Tests:**
  - Test finding skill in single profile
  - Test finding skill in multiple profiles
  - Test finding non-existent skill returns empty result
  
  **Integration Tests:**
  - Test against actual profile structure with real components
  - Verify agent-smith command integration works correctly
  
  **Component Browser Tests:**
  - Not applicable (bash script functionality)

- [ ] Story-003: As a developer, I want profile-builder to use `agent-smith profile copy` command so component metadata and lock files are preserved correctly.

  **Acceptance Criteria:**
  - Workflow A Step A6 updated to use `./agent-smith profile copy skills <source-profile> <target-profile> <skill-name>`
  - Workflow uses `profile copy` for agents and commands as well
  - Manual `cp -r` commands removed from documentation
  - Error handling for cases where component doesn't exist in source profile
  - Source profile automatically detected using `find_profiles_with_skill()` helper
  
  **Testing Criteria:**
  **Unit Tests:**
  - Not applicable (documented workflow)
  
  **Integration Tests:**
  - Test copying skills using agent-smith profile copy command
  - Test copying agents using agent-smith profile copy command
  - Test copying commands using agent-smith profile copy command
  - Verify lock files preserved after copy
  
  **Component Browser Tests:**
  - Not applicable (bash command integration)

- [ ] Story-004: As a user, I want to understand how profile-builder differs from cherry-pick so I can choose the right tool for my needs.

  **Acceptance Criteria:**
  - "How It Works" section includes comparison to cherry-pick command
  - Clear explanation that cherry-pick is for manual selection, profile-builder is for template-driven automation
  - Examples showing when to use each tool
  - Table comparing features of both approaches
  - Statement that they are complementary, not redundant
  
  **Testing Criteria:**
  **Unit Tests:**
  - Not applicable (documentation)
  
  **Integration Tests:**
  - Not applicable (documentation)
  
  **Component Browser Tests:**
  - Not applicable (documentation)

- [ ] Story-005: As a developer reading the skill documentation, I want to see accurate paths and examples that reflect the profile-based architecture so I'm not confused by outdated references.

  **Acceptance Criteria:**
  - All references to `~/.agent-smith/skills/` updated to `~/.agent-smith/profiles/*/skills/`
  - All references to `~/.agent-smith/agents/` updated to `~/.agent-smith/profiles/*/agents/`
  - All references to `~/.agent-smith/commands/` updated to `~/.agent-smith/profiles/*/commands/`
  - Examples show profile structure with sample profiles like "anthropics-skills" and "wshobson-agents"
  - Code examples use correct profile-aware paths
  - Explanatory diagrams show profile directory structure
  
  **Testing Criteria:**
  **Unit Tests:**
  - Not applicable (documentation)
  
  **Integration Tests:**
  - Grep SKILL.md for old path patterns to ensure none remain
  - Grep README.md for old path patterns to ensure none remain
  
  **Component Browser Tests:**
  - Not applicable (documentation)

- [ ] Story-006: As a user, I want the skill recommendations to show which profile contains each component so I know where components will be copied from.

  **Acceptance Criteria:**
  - Step A4 recommendations format includes source profile in parentheses
  - Format: `✓ api-design-principles (from: wshobson-agents)`
  - Shows source profile for skills, agents, and commands
  - Handles components found in multiple profiles by showing first match
  - Clear visual distinction between component name and source profile
  
  **Testing Criteria:**
  **Unit Tests:**
  - Not applicable (documented workflow)
  
  **Integration Tests:**
  - Test recommendation display with components from multiple profiles
  - Verify source profile detection works correctly
  
  **Component Browser Tests:**
  - Not applicable (skill workflow output)

- [ ] Story-007: As a developer, I want clear documentation about which agent-smith commands the skill uses and why so I understand the implementation approach.

  **Acceptance Criteria:**
  - New "Implementation Notes" section in SKILL.md
  - Table showing which operations use which agent-smith commands
  - Explanation of when filesystem operations are used instead of commands
  - Justification for hybrid approach (commands for actions, filesystem for queries)
  - List of all agent-smith commands leveraged by the skill
  
  **Testing Criteria:**
  **Unit Tests:**
  - Not applicable (documentation)
  
  **Integration Tests:**
  - Not applicable (documentation)
  
  **Component Browser Tests:**
  - Not applicable (documentation)

- [ ] Story-008: As a user, I want README.md to explain the profile architecture and how profile-builder fits into the ecosystem so I understand the big picture.

  **Acceptance Criteria:**
  - New "How Agent-Smith Profiles Work" section in README.md
  - New "Profile Architecture" ASCII diagram showing directory structure
  - New "Profile-Builder vs Cherry-Pick" comparison section
  - Table showing which scenarios suit which tool
  - "Current Component Counts" updated to show per-profile counts
  - Examples updated to show profile-based workflow
  
  **Testing Criteria:**
  **Unit Tests:**
  - Not applicable (documentation)
  
  **Integration Tests:**
  - Not applicable (documentation)
  
  **Component Browser Tests:**
  - Not applicable (documentation)

- [ ] Story-009: As a maintainer, I want component-scanner.sh to provide helper functions for profile-aware operations so the skill can efficiently query profile contents.

  **Acceptance Criteria:**
  - New `list_profiles()` function lists all profile directories
  - New `get_profile_skills()` function lists skills in a specific profile
  - New `get_profile_agents()` function lists agents in a specific profile
  - New `get_profile_commands()` function lists commands in a specific profile
  - New CLI cases added to handle new functions
  - All new functions follow existing naming and error handling patterns
  
  **Testing Criteria:**
  **Unit Tests:**
  - Test each new function with valid profile name
  - Test functions with non-existent profile name
  - Test functions with empty profile
  
  **Integration Tests:**
  - Test against actual ~/.agent-smith/profiles structure
  - Verify output format matches expectations
  
  **Component Browser Tests:**
  - Not applicable (bash script functionality)

- [ ] Story-010: As a developer, I want all six templates to remain unchanged so existing template patterns continue to work with the new profile-based scanning.

  **Acceptance Criteria:**
  - No changes to templates/java-backend.yaml
  - No changes to templates/python-ml.yaml
  - No changes to templates/react-frontend.yaml
  - No changes to templates/nodejs-fullstack.yaml
  - No changes to templates/mobile-react-native.yaml
  - No changes to templates/devops-platform.yaml
  - Template keyword patterns work with profile-based component scanning
  
  **Testing Criteria:**
  **Unit Tests:**
  - Verify template files are unchanged (git diff shows no changes)
  
  **Integration Tests:**
  - Test each template with profile-based scanner to ensure matching still works
  - Verify keyword patterns correctly match components from profiles
  
  **Component Browser Tests:**
  - Not applicable (template files)

## Functional Requirements

- FR-1: The component scanner SHALL scan `~/.agent-smith/profiles/` for all component discovery operations
- FR-2: The skill SHALL use `agent-smith profile copy` command for all component copying operations between profiles
- FR-3: The skill SHALL use `agent-smith profile create` command for creating new profiles
- FR-4: The skill SHALL use `agent-smith profile activate` and `agent-smith link all` commands for profile activation
- FR-5: The skill SHALL provide functions to identify which profiles contain specific components
- FR-6: The skill SHALL handle cases where components exist in multiple profiles by selecting the first match
- FR-7: The documentation SHALL clearly differentiate profile-builder from cherry-pick command
- FR-8: The documentation SHALL explain that profile-builder provides template-driven automation while cherry-pick provides manual selection
- FR-9: All documentation SHALL use correct profile-based paths (not base directory paths)
- FR-10: The skill SHALL preserve existing template files without modifications
- FR-11: The skill SHALL maximize use of agent-smith commands for all state-changing operations
- FR-12: The skill SHALL use filesystem operations for fast pattern matching and component discovery
- FR-13: The recommendations display SHALL include source profile information for each component
- FR-14: The skill SHALL gracefully handle empty profiles and missing components

## Non-Goals

- No changes to template YAML files (keep existing keyword patterns)
- No creation of new templates beyond the existing six
- No implementation of the skill workflows in this task (only documentation and helper functions)
- No integration with repository installation workflows (users install profiles separately via `agent-smith install`)
- No modifications to agent-smith's core profile management code
- No changes to how agent-smith commands work (only how we use them)
- No interactive UI enhancements (keep existing text-based interface)
- No support for scanning base ~/.agent-smith/ directories (profiles only)

## Technical Notes

### Files to Update

1. **`lib/component-scanner.sh`** (~50 lines changed)
   - Update `scan_skills()`, `scan_agents()`, `scan_commands()` to scan profiles
   - Add `list_profiles()`, `get_profile_skills()`, `get_profile_agents()`, `get_profile_commands()`
   - Add `find_profiles_with_skill()`, `find_profiles_with_agent()`, `find_profiles_with_command()`
   - Update main CLI switch case to handle new functions

2. **`SKILL.md`** (~100 lines changed)
   - Update "How It Works" section with profile architecture explanation
   - Add cherry-pick comparison and differentiation
   - Update Workflow A Step A3 to scan profiles
   - Update Workflow A Step A4 to show source profiles
   - Update Workflow A Step A6 to use `agent-smith profile copy`
   - Add "Implementation Notes" section
   - Update all path references throughout
   - Update all code examples

3. **`README.md`** (~80 lines changed)
   - Add "How Agent-Smith Profiles Work" section
   - Add "Profile Architecture" diagram
   - Add "Profile-Builder vs Cherry-Pick" comparison section
   - Update "Current Component Counts" section
   - Update all examples to show profile-based workflow
   - Update all path references

4. **Templates** (0 changes)
   - No modifications needed

### Agent-Smith Commands to Leverage

The skill should use these agent-smith commands where possible:
- `agent-smith profile list` - Display available profiles
- `agent-smith profile show <name>` - Show profile details
- `agent-smith profile create <name>` - Create new empty profile
- `agent-smith profile copy <type> <source> <target> <name>` - Copy component between profiles
- `agent-smith profile remove <type> <profile> <name>` - Remove component from profile
- `agent-smith profile activate <name>` - Activate a profile
- `agent-smith link all` - Link components to editors

### Parallel Execution Strategy

Stories can be executed in parallel groups:

**Group 0 (Foundation)**: Must complete first
- Story-001: Update scanner to scan profiles
- Story-002: Add profile tracking functions
- Story-009: Add helper functions to scanner

**Group 1 (Independent Documentation)**: Can run in parallel after Group 0
- Story-004: Document cherry-pick comparison
- Story-007: Document implementation approach
- Story-008: Update README with architecture info
- Story-010: Verify templates unchanged

**Group 2 (Integration)**: Requires Group 0 and Group 1
- Story-003: Update workflow to use profile copy
- Story-005: Update all path references
- Story-006: Update recommendations format

## Success Criteria

The profile-builder skill update is successful when:
- Component scanner correctly discovers components from all installed profiles
- All agent-smith profile commands are properly leveraged
- Documentation clearly differentiates profile-builder from cherry-pick
- All path references reflect profile-based architecture
- Templates remain unchanged and continue to work
- Skill provides clear value proposition of template-driven automation
- No references to deprecated base directory paths remain
- Source profile information is displayed in recommendations
