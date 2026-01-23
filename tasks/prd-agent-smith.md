# PRD: Agent Smith - Universal Opencode Component Manager

## Introduction

Agent Smith is a universal component manager for AI coding environments that provides seamless downloading, management, and linking of agents, commands, and skills. It maintains full compatibility with the existing npx add-skill ecosystem while extending functionality to cover all component types. The tool operates with a two-tier system: global canonical storage in `~/.agents/` and automatic symlinking to environment-specific directories, enabling developers to efficiently manage components across multiple projects with intelligent change detection and flexible repository format support. 

**Current Scope**: Starting with OpenCode support as the initial environment. The architecture is designed for extensibility to support Claude Code, Cursor, and other AI coding environments in future releases, but the initial implementation focuses solely on OpenCode compatibility and functionality.

## Goals

- Create a universal component manager that replaces multiple scattered tools across AI coding environments
- Improve developer experience with seamless component management for OpenCode (initial release focus)
- Establish a standardized system for managing agents, commands, and skills across AI platforms
- Maintain full compatibility with existing npx add-skill ecosystem
- Support flexible repository format detection and linking to OpenCode directories (initial scope)
- Provide intelligent updates via GitHub tree SHA change detection
- Enable bulk operations and efficient component management
- Build extensible architecture foundation for future addition of new AI coding environments

## User Stories

- [x] Story-001: As a developer, I want to download skills from git repositories so I can have a centralized component library.

  **Acceptance Criteria:**
  - Download skill directories (not entire repos) from GitHub, GitLab, and direct URLs with automatic source type detection
  - Perform shallow cloning for efficiency
  - Store skill directories in ~/.agents/skills/[skill-name]/ with metadata tracking
  - Create SKILL.md files in proper format for opencode compatibility

- [x] Story-002: As a developer, I want to download agents from git repositories so I can manage agent definitions centrally.

  **Acceptance Criteria:**
  - Download agent directories (not individual .md files) with source tracking
  - Store in ~/.agents/agents/[agent-name]/ with lock file management
  - Support multiple source formats (GitHub shorthand, full URLs, local paths)
  - Support opencode agent format with frontmatter and proper directory structure

- [x] Story-003: As a developer, I want to download commands from git repositories so I can maintain a command library.

  **Acceptance Criteria:**
  - Download command directories (not individual .md files) with metadata tracking
  - Store in ~/.agents/commands/[command-name]/ with lock file integration
  - Support batch downloading from repositories
  - Support opencode command format with frontmatter and proper file structure

- [x] Story-004: As a developer, I want to link downloaded components to my opencode repository so I can use them in projects.

  **Acceptance Criteria:**
  - Create relative symlinks from ~/.agents/[type]/[name]/ to ~/.config/opencode/[type]/[name]/
  - Auto-detect opencode repository structure and create ~/.config/opencode/ if needed
  - Handle conflicts by skipping existing files with user notification
  - Support cross-platform symlink creation with Windows junction fallbacks

- [x] Story-005: As a developer, I want the tool to work with any repository format so I'm not limited to standard structure.

  **Acceptance Criteria:**
  - Perform recursive content-based detection for skills (*/SKILL.md), agents (*/AGENT.md), and commands (*/COMMAND.md)
  - Identify components by file patterns: skills/*/SKILL.md, agents/*/AGENT.md, commands/*/COMMAND.md
  - Support custom repository layouts through flexible parsing and user-configurable patterns
  - Handle monorepo structures with multiple component types in same repository

- [x] Story-006: As a developer, I want intelligent update detection so I know when components change.

  **Acceptance Criteria:**
  - Use GitHub tree SHA for precise change detection
  - Compare stored hashes with current repository state
  - Only re-download when actual content changes

- [x] Story-007: As a developer, I want npx add-skill compatibility so I can use both tools together.

  **Acceptance Criteria:**
  - Use same ~/.agents/ directory structure for component storage
  - Maintain compatible .skill-lock.json format for skills
  - Extend pattern with .agent-lock.json and .command-lock.json for other component types
  - Support existing .skill-lock.json files from npx add-skill installations
  - Coexist peacefully with npx add-skill tool (no conflicts or overwrites)

- [x] Story-008: As a developer, I want bulk operations so I can manage many components efficiently.

  **Acceptance Criteria:**
  - Download all component types from repositories with add-all command
  - Link all components to ~/.config/opencode/ with single command
  - Perform batch updates across all installed components with configurable concurrency
  - Support bulk operations like add-all, link-all, update-all

- [x] Story-009: As a developer, I want cross-platform support so I can work on any OS.

  **Acceptance Criteria:**
  - Support native symlinks on macOS/Linux with automatic detection
  - Use junctions as fallback on Windows when symlinks fail
  - Handle path normalization across platforms (forward/backward slashes, drive letters)
  - Graceful degradation to file copies when symlinks/junctions are not possible
  - Proper permission handling across different operating systems

- [x] Story-010: As a developer, I want a CLI interface similar to npx so I have a familiar experience.

  **Acceptance Criteria:**
  - Use add-agent, add-command, add-skill subcommands matching npx add-skill patterns
  - Provide list, search, update, link, and help commands
  - Include comprehensive help with examples and auto-completion
  - Support global (--global, -g) and force (--force, -f) flags like npx add-skill
  - Provide dry-run functionality to preview actions before execution

- [x] Story-011: As a developer, I want comprehensive source type support so I can download from anywhere.

  **Acceptance Criteria:**
  - Parse GitHub shorthand (owner/repo) and expand to full GitHub URLs
  - Handle GitLab URLs (gitlab.com/owner/repo) with proper API integration
  - Support any direct git URL via generic provider system
  - Work with local file paths for testing and local development
  - Validate all source types before attempting download

- [x] Story-012: As a developer, I want automatic repository detection so linking is effortless.

  **Acceptance Criteria:**
  - Detect when current directory is within an opencode repository (walk up to find .git)
  - Automatically create ~/.config/opencode/ structure if it doesn't exist
  - Fall back to recursive component search when standard layouts aren't found
  - Link components to appropriate opencode directories based on component type
  - Validate that opencode can properly load the linked components

## Technical Requirements

- TR-1: The system must be implemented in Go (Golang) programming language
- TR-2: The system must be cross-platform compatible (Linux, macOS, Windows)
- TR-3: The system must have extensible architecture foundation for supporting multiple AI coding environments (future releases)
- TR-4: The system must support OpenCode directory structure and component formats (initial release scope)

## Functional Requirements

- FR-1: The system must support downloading components from multiple git source types (GitHub shorthand, GitLab URLs, direct git URLs, local paths)
- FR-2: The system must maintain separate lock files for each component type (.skill-lock.json, .agent-lock.json, .command-lock.json) in project root
- FR-3: The system must store components in ~/.agents/ with subdirectories for skills/, agents/, and commands/ (one directory per component)
- FR-4: The system must create relative symlinks from ~/.config/opencode/ to ~/.agents/ global storage for component availability
- FR-5: The system must use GitHub tree SHA for precise change detection and only re-download when content actually changes
- FR-6: The system must detect repository formats through recursive content-based analysis for all component types
- FR-7: The system must maintain full compatibility with existing npx add-skill ecosystem and coexist peacefully
- FR-8: The system must support cross-platform symlink creation with Windows junction fallbacks and graceful degradation
- FR-9: The system must provide CLI interface with add-skill, add-agent, add-command, list, search, update, and link commands
- FR-10: The system must handle conflicts by skipping existing files with clear user notification and --force override option
- FR-11: The system must support bulk operations for downloading and linking multiple components with configurable concurrency
- FR-12: The system must validate component content before installation and linking (SKILL.md/AGENT.md/COMMAND.md format validation)
- FR-13: The system must create ~/.config/opencode/ directory structure automatically if it doesn't exist
- FR-14: The system must support opencode component formats (skills: SKILL.md with frontmatter, commands: .md with frontmatter, agents: .md with frontmatter)
- FR-15: The system must support OpenCode directory detection and target configuration (~/.config/opencode/)
- FR-16: The system must provide foundation architecture for future environment extensions
- FR-17: The system must maintain clear separation between core functionality and environment-specific implementations

## Non-Goals

- No social media integration or authentication
- No web-based user interface or dashboard
- No automatic background updates or notifications
- No enterprise features like team-based component sharing
- No plugin system for custom component types beyond skills/agents/commands
- No dependency management between components
- No version constraints or semantic versioning support (beyond basic change detection)
- No workspace or team-level component management
- No integration with IDEs or editors beyond basic CLI
- No content validation beyond basic structure checking and frontmatter validation
- No backup or restore functionality for component libraries
- No analytics or usage tracking beyond basic error reporting
- No modification of existing npx add-skill installations or lock files
- No hard dependencies on specific AI environments - should work with any supported environment
- Initial release scope limited to OpenCode only (extensible architecture for future environments)