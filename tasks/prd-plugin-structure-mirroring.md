# PRD: Plugin Structure Mirroring for Agent Downloads

## Introduction

Fix the multi-component agent download duplication issue in agent-smith by implementing plugin-aware structure mirroring. Currently, when downloading agents from repositories with plugin structures (like `plugins/ui-design/agents/*.md`), agent-smith creates duplicate nested directories with replicated files. This results in 27 duplicate files for the three UI design agents. The system should auto-detect `plugins/` directories in repositories and preserve their structure in `~/.agents/`, while maintaining backward compatibility with flat repository structures.

## Goals

- Eliminate file duplication in multi-component plugin downloads (currently 27 duplicate files for ui-design agents)
- Mirror source repository structure in `~/.agents/` for plugin-based repos
- Auto-detect both plugin-based and flat repository structures
- Apply consistent logic to agents, commands, and skills
- Maintain backward compatibility with existing flat repository installations
- Keep lock file format with minimal extensions (add optional `pluginPath` field)
- Implement simple update strategy (re-download entire plugin structure)

## User Stories

- [ ] Story-001: As an agent-smith developer, I want to extend data structures to track plugin information so that the system can preserve repository structure throughout the download pipeline.

  **Acceptance Criteria:**
  - Add `FilePath` field to `DetectedComponent` struct to store full relative path to component file
  - Add `PluginPath` field to `ComponentMetadata` struct with JSON tag `pluginPath,omitempty`
  - Add `PluginPath` field to `ComponentLockEntry` struct with JSON tag `pluginPath,omitempty`
  - All fields are properly initialized and stored during component detection
  - Lock file format remains backward compatible with existing entries
  
  **Testing Criteria:**
  **Unit Tests:**
  - Struct field initialization and JSON marshaling tests
  - Backward compatibility tests for lock file reading with/without pluginPath
  
  **Integration Tests:**
  - End-to-end test verifying plugin path propagation through pipeline
  - Lock file compatibility tests with mixed old/new format entries

- [ ] Story-002: As an agent-smith developer, I want to update component detection to capture full file paths so that the download logic can copy only relevant files instead of entire directories.

  **Acceptance Criteria:**
  - Update `detectComponentForPattern()` function signature to return four values: `(string, string, string, bool)` with third return being filePath
  - Update all return statements in `detectComponentForPattern()` to include `fullRelPath` as third return value
  - Update `detectComponentsInRepo()` function call on line 713 to capture `filePath` from detection result
  - Store `FilePath` in `DetectedComponent` struct when creating components (lines 745-750)
  - Update duplicate tracking logic to store `FilePath` (lines 735-740)
  - Update additional agent detection section to store `fullRelPath` as `FilePath` (lines 791-796)
  - Update additional command detection section similarly (around lines 807-843)
  - All detected components have accurate FilePath information
  
  **Testing Criteria:**
  **Unit Tests:**
  - Test `detectComponentForPattern()` returns correct filePath for various input paths
  - Test `DetectedComponent` struct properly stores all path fields
  
  **Integration Tests:**
  - Test component detection across plugin and flat repository structures
  - Verify FilePath accuracy for nested plugin paths like "plugins/ui-design/agents/accessibility-expert.md"

- [ ] Story-003: As an agent-smith developer, I want to add plugin detection helper functions so that the system can identify and extract plugin paths from component paths.

  **Acceptance Criteria:**
  - Create `extractPluginPath()` function that takes componentPath and returns plugin directory path (e.g., "plugins/ui-design") or empty string
  - Function correctly identifies "plugins/" in path and extracts plugin name
  - Create `detectCommonPluginPath()` function that takes slice of DetectedComponent and returns common plugin path or empty string
  - Function verifies all components share the same plugin path before returning it
  - Both functions handle edge cases (empty inputs, no plugins, mixed structures)
  - Functions use `filepath.ToSlash()` for cross-platform compatibility
  
  **Testing Criteria:**
  **Unit Tests:**
  - Test `extractPluginPath()` with various path formats (plugins/foo, no plugins, nested paths)
  - Test `detectCommonPluginPath()` with same plugin, different plugins, no plugins
  - Test cross-platform path handling (Windows/Unix separators)
  
  **Integration Tests:**
  - Test functions with real repository detection results
  - Verify correct plugin detection for wshobson/agents ui-design components

- [ ] Story-004: As an agent-smith user, I want the download logic to mirror plugin structures so that I can see the exact repository organization in my local .agents directory without file duplication.

  **Acceptance Criteria:**
  - Rewrite `downloadAgentWithRepo()` to detect plugin structures using `detectCommonPluginPath()`
  - When plugin detected, copy entire plugin directory once to `~/.agents/plugins/{plugin-name}/`
  - Save metadata with `pluginPath` field to track plugin structure
  - Update lock file entries for all components in plugin with pluginPath information
  - Maintain existing behavior for single-agent downloads (no plugin structure)
  - Maintain existing behavior for multi-agent non-plugin downloads (monorepo structure)
  - No duplicate files created during plugin downloads
  
  **Testing Criteria:**
  **Unit Tests:**
  - Test plugin path detection logic
  - Test metadata generation with pluginPath field
  
  **Integration Tests:**
  - Download accessibility-expert from wshobson/agents and verify plugin structure created
  - Verify only one copy of ui-design plugin directory exists
  - Download second agent from same plugin and verify structure reused
  - Download flat repository agent and verify existing behavior maintained
  
  **Component Browser Tests:**
  - Manual verification of directory structure in ~/.agents/plugins/
  - Verify no duplicate .md files exist in plugin structure

- [ ] Story-005: As an agent-smith developer, I want to apply identical plugin mirroring logic to skills and commands so that all component types benefit from structure preservation.

  **Acceptance Criteria:**
  - Apply same rewrite pattern from Story-004 to `downloadSkillWithRepo()` function (around lines 1200-1300)
  - Apply same rewrite pattern to `downloadCommandWithRepo()` function (around lines 2100-2200)
  - Both functions use `detectCommonPluginPath()` for plugin detection
  - Both functions copy entire plugin structure when detected
  - Both functions save metadata and lock entries with pluginPath
  - Both functions maintain backward compatibility for non-plugin structures
  
  **Testing Criteria:**
  **Unit Tests:**
  - Test skill download with plugin structure
  - Test command download with plugin structure
  
  **Integration Tests:**
  - Download skill from ui-design plugin and verify structure mirrored
  - Download command from ui-design plugin and verify structure mirrored
  - Verify skills and commands share same plugin directory with agents
  
  **Component Browser Tests:**
  - Manual verification that skills and commands exist in same plugin directory
  - Verify no duplication across component types

- [ ] Story-006: As an agent-smith user, I want proper symlinks created to plugin-based components so that I can use agents/commands/skills installed from plugin structures.

  **Acceptance Criteria:**
  - Update linking logic to load metadata and check for `pluginPath` field
  - When pluginPath present, create symlink to specific file in plugin directory (e.g., `plugins/ui-design/agents/accessibility-expert.md`)
  - When pluginPath absent, use existing directory-based linking logic
  - Symlinks created in `~/.config/opencode/{type}s/` point to correct files
  - Update `isMonorepoContainer()` to recognize plugin structures and return false for them
  - Plugin structures not treated as monorepo containers
  
  **Testing Criteria:**
  **Unit Tests:**
  - Test metadata loading and pluginPath detection
  - Test symlink path generation for plugin vs non-plugin structures
  
  **Integration Tests:**
  - Download plugin-based agent and verify symlink created correctly
  - Verify symlink points to actual .md file in plugin directory
  - Test that plugin structures bypass monorepo container logic
  
  **Component Browser Tests:**
  - Manual verification of symlinks in ~/.config/opencode/agents/
  - Follow symlinks and verify they point to correct plugin files
  - Verify agents from plugin structure are usable in OpenCode

- [ ] Story-007: As an agent-smith user, I want comprehensive integration testing so that the plugin mirroring system works correctly for real-world repositories.

  **Acceptance Criteria:**
  - Test downloading accessibility-expert from wshobson/agents creates `~/.agents/plugins/ui-design/` structure
  - Verify plugin directory contains all three agent .md files, commands, and skills
  - Verify metadata file includes `pluginPath: "plugins/ui-design"`
  - Test downloading design-system-architect reuses existing plugin structure
  - Test downloading bash-pro (multi-component non-plugin) maintains existing monorepo behavior
  - Test downloading single-agent flat repo maintains existing behavior
  - Verify cross-platform path handling works on Windows and Unix systems
  - No duplicate files created in any test scenario
  
  **Testing Criteria:**
  **Unit Tests:**
  - Mock repository structures for testing various scenarios
  - Test edge cases (empty plugins, nested plugins, mixed structures)
  
  **Integration Tests:**
  - Full download workflow test with wshobson/agents repository
  - Test multiple agents from same plugin downloaded sequentially
  - Test agents from different plugins in same repository
  - Test update workflow (re-download entire plugin structure)
  
  **Component Browser Tests:**
  - Manual end-to-end testing with real wshobson/agents repository
  - Verify complete workflow: download, link, use agent in OpenCode
  - Verify directory structure matches source repository exactly
  - Verify no broken symlinks or duplicate files remain

## Functional Requirements

- FR-1: The system must detect plugin-based repository structures by identifying "plugins/" in component paths
- FR-2: When downloading components from plugin structures, the system must copy the entire plugin directory once to `~/.agents/plugins/{plugin-name}/`
- FR-3: The system must store plugin path information in metadata and lock files using the `pluginPath` field
- FR-4: The system must create symlinks to individual component files within plugin directories
- FR-5: The system must apply plugin mirroring logic consistently to agents, commands, and skills
- FR-6: The system must maintain backward compatibility with existing flat repository installations
- FR-7: When updating plugin-based components, the system must re-download the entire plugin structure
- FR-8: The system must use cross-platform path handling with `filepath.Join()` and `filepath.ToSlash()`
- FR-9: The system must prevent plugin structures from being treated as monorepo containers
- FR-10: The system must eliminate file duplication by copying plugin structures only once regardless of how many components are downloaded from the same plugin

## Non-Goals (Out of Scope)

- No migration script for existing broken installations (users will clean up manually)
- No selective file updates within plugin structures (always re-download entire plugin)
- No backward compatibility for legacy .agent-metadata.json format (focus on lock file format only)
- No UI changes or user-facing configuration options for plugin handling
- No support for custom plugin directory names or structures outside the "plugins/" convention
- No optimization for incremental updates of large plugin structures
- No automatic cleanup of orphaned plugin directories when all components are uninstalled
- No validation of plugin structure integrity or completeness
