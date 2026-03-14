# PRD: Find Skill Command for Agent Smith

**Created**: 2026-02-22 12:46 UTC

---

## Introduction

Add a `find skill` command to agent-smith that queries the skills.sh remote registry API to help users discover available skills they can install. This feature provides parity with `npx skills find` while maintaining agent-smith's command structure and installation workflow, making it easy for users to discover and install skills using agent-smith commands.

## Goals

- Enable users to discover skills from the skills.sh remote registry
- Provide familiar experience similar to `npx skills find` command
- Show clear installation instructions using agent-smith commands
- Support both human-readable terminal output and machine-readable JSON
- Drive adoption by making skill discovery seamless

## User Stories

- [ ] Story-001: As a developer, I want to search for skills by keyword so I can discover relevant skills to install.

  **Acceptance Criteria:**
  - Command syntax: `agent-smith find skill <query>`
  - Query must be at least 2 characters (matching skills.sh API validation)
  - Queries the skills.sh API: `https://skills.sh/api/search?q=<query>`
  - Returns up to 20 results by default
  - Each result shows: source@skillId, install count, skills.sh URL, and installation commands
  
  **Testing Criteria:**
  **Unit Tests:**
  - Query validation (minimum 2 characters)
  - API response parsing and error handling
  - Result formatting logic
  
  **Integration Tests:**
  - End-to-end command execution with live API
  - Network error handling
  - Empty results handling

- [ ] Story-002: As a developer, I want clear installation instructions for each result so I know how to install skills using agent-smith.

  **Acceptance Criteria:**
  - Shows header with generic installation syntax
  - Each result displays specific installation command
  - Shows both `install skill <owner/repo> <skill-name>` and `install all <owner/repo>` options
  - Installation commands use gray/dimmed color to reduce visual noise
  - Commands extract owner/repo from API's `source` field correctly
  
  **Testing Criteria:**
  **Unit Tests:**
  - Installation command formatting from API response data
  - Source URL parsing and GitHub shorthand extraction

- [ ] Story-003: As a developer, I want to see the Agent Smith banner so I have consistent branding with other agent-smith commands.

  **Acceptance Criteria:**
  - Displays existing Agent Smith ASCII banner from `getBanner()`
  - Banner shown before search results
  - Maintains brand consistency with other commands
  
  **Testing Criteria:**
  **Unit Tests:**
  - Banner text is included in output

- [ ] Story-004: As a developer, I want compact, readable output so I can quickly scan results without information overload.

  **Acceptance Criteria:**
  - Compact format: banner + header + results list
  - Each result: source@skillId (bright), install count (cyan), URL (dimmed), install command (dimmed)
  - Uses box-drawing characters (└) for visual hierarchy
  - Respects terminal width for readability
  - Results sorted by install count (descending)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Output formatting matches expected format
  - Color codes applied correctly
  - Result sorting by install count

- [ ] Story-005: As a developer, I want meaningful error messages so I understand what went wrong and how to fix it.

  **Acceptance Criteria:**
  - Query too short (< 2 chars): "Error: Query must be at least 2 characters"
  - No results: "No skills found matching 'xyz'\n\nTry different keywords or visit https://skills.sh to browse all skills."
  - Network error: "Error: Failed to connect to skills.sh registry\nCheck your internet connection and try again."
  - API error: "Error: skills.sh API returned an error\nPlease try again later or visit https://skills.sh"
  
  **Testing Criteria:**
  **Unit Tests:**
  - Error message generation for each error type
  
  **Integration Tests:**
  - Network failure simulation
  - API error response handling

- [ ] Story-006: As a developer, I want JSON output option so I can script and automate skill discovery.

  **Acceptance Criteria:**
  - Flag: `--json` outputs machine-readable JSON
  - JSON structure: `{"query": "...", "count": N, "results": [{"source": "...", "skillId": "...", "name": "...", "installs": N, "url": "...", "installCommand": "...", "installAllCommand": "..."}]}`
  - No banner or colored output in JSON mode
  - Valid JSON that can be piped to `jq` or other tools
  
  **Testing Criteria:**
  **Unit Tests:**
  - JSON serialization and structure validation
  
  **Integration Tests:**
  - JSON output can be parsed by external tools

- [ ] Story-007: As a developer, I want the skills.sh URL shown but de-emphasized so I can reference it if needed without cluttering the output.

  **Acceptance Criteria:**
  - URL displayed with gray/dimmed color (color code 102)
  - URL on second line under source@skillId
  - Uses box-drawing character (└) prefix
  - URL format: `https://skills.sh/<source>/<skillId>`
  
  **Testing Criteria:**
  **Unit Tests:**
  - URL construction from API response
  - Color formatting for URL

- [ ] Story-008: As a developer, I want a limit flag so I can control how many results are shown.

  **Acceptance Criteria:**
  - Flag: `--limit <number>` or `-l <number>`
  - Default limit: 20 results
  - Limits applied to API response
  - Works with both terminal and JSON output
  
  **Testing Criteria:**
  **Unit Tests:**
  - Limit parameter parsing and application

## Functional Requirements

- FR-1: The system SHALL implement a `find` command with `skill` subcommand following agent-smith's command structure
- FR-2: The system SHALL query the skills.sh API at `https://skills.sh/api/search?q=<query>` for skill discovery
- FR-3: The system SHALL validate queries are at least 2 characters before making API calls
- FR-4: The system SHALL display up to 20 results by default, configurable via `--limit` flag
- FR-5: The system SHALL show the Agent Smith ASCII banner in terminal output mode
- FR-6: The system SHALL display results in compact format with source@skillId, install count, URL, and installation commands
- FR-7: The system SHALL provide installation commands using agent-smith syntax: `install skill <owner/repo> <skill-name>`
- FR-8: The system SHALL support `--json` flag for machine-readable output
- FR-9: The system SHALL handle network errors, API errors, and empty results with clear error messages
- FR-10: The system SHALL use color formatting (cyan for counts, gray for URLs and commands, bright white for skill identifiers)
- FR-11: The system SHALL extract GitHub repository path from API's `source` field for installation commands
- FR-12: The system SHALL show skills.sh URLs in dimmed/gray color to de-emphasize them

## Non-Goals

- No local skill search (only remote skills.sh registry)
- No interactive mode or fuzzy picker (terminal output only)
- No caching of API responses
- No find command for agents or commands (only skills in initial implementation)
- No fallback to local search on network failure
- No retry logic or exponential backoff
- No installation shortcuts (e.g., `--install` flag to directly install)
- No filtering by install count, source, or other criteria
- No pagination or infinite scroll
