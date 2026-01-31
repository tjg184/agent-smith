# PRD: Conventional Commit Converter

**Created**: 2026-01-31 12:38 UTC

---

## Introduction

Implement an AI-powered tool that automatically converts all commits in the current branch to proper conventional commit format by analyzing commit diffs and rewriting git history. This solves the problem of inconsistent commit messages in repositories and ensures compliance with Conventional Commits specification without manual effort.

## Goals

- Automatically analyze all commits in the current branch using AI (LLM)
- Determine appropriate conventional commit type and scope from diff analysis
- Rewrite git history with properly formatted conventional commits
- Always prompt user for breaking change confirmation before marking commits
- Preserve commit authorship and timestamps during rewrite
- Provide clear feedback and preview before making changes

## User Stories

- [ ] Story-001: As a developer, I want to convert all commits in my current branch to conventional format so that my commit history is standardized.

  **Acceptance Criteria:**
  - Tool identifies all commits in current branch (from divergence point to HEAD)
  - Excludes commits that already follow conventional commit format
  - Displays count of commits to be converted
  - Shows branch name and commit range being analyzed
  
  **Testing Criteria:**
  **Unit Tests:**
  - Git command execution and parsing logic
  - Branch detection and commit range calculation
  - Conventional commit format detection regex
  
  **Integration Tests:**
  - Git repository integration with test fixtures
  - Multi-commit branch scenarios
  - Edge cases (single commit, no commits, etc.)
  
  **Component Browser Tests:**
  - N/A (CLI tool)

- [ ] Story-002: As a developer, I want the tool to use AI to analyze each commit's changes so that it can determine the correct conventional commit type automatically.

  **Acceptance Criteria:**
  - For each commit, extracts full diff content
  - Sends diff to LLM with prompt for conventional commit analysis
  - LLM returns type (feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert), scope (optional), and description
  - Handles API errors gracefully with retry logic
  - Supports streaming or batch processing for performance
  
  **Testing Criteria:**
  **Unit Tests:**
  - Diff extraction logic
  - LLM prompt construction
  - Response parsing and validation
  - Error handling and retry logic
  
  **Integration Tests:**
  - Mock LLM API integration tests
  - Real LLM API tests with sample commits
  - Various commit types (file additions, deletions, modifications)
  
  **Component Browser Tests:**
  - N/A (CLI tool)

- [ ] Story-003: As a developer, I want to preview the proposed conventional commit messages before any changes are made so that I can verify they are correct.

  **Acceptance Criteria:**
  - Displays side-by-side comparison: original message vs. proposed conventional message
  - Shows commit SHA, author, and date for each commit
  - Includes diff summary (files changed, insertions, deletions)
  - Provides option to accept all, reject all, or review individually
  - Allows manual editing of proposed messages before applying
  
  **Testing Criteria:**
  **Unit Tests:**
  - Message formatting and display logic
  - User input parsing and validation
  - Edit mode functionality
  
  **Integration Tests:**
  - Interactive CLI workflow testing
  - Multiple commits preview scenarios
  
  **Component Browser Tests:**
  - N/A (CLI tool)

- [ ] Story-004: As a developer, I want the tool to always ask me about breaking changes so that I can mark them appropriately with BREAKING CHANGE notation.

  **Acceptance Criteria:**
  - For each commit, prompts user: "Does this commit contain breaking changes? (y/n)"
  - If yes, prompts for breaking change description
  - Appends "BREAKING CHANGE: <description>" to commit body
  - Adds "!" suffix to type if breaking (e.g., "feat!" or "fix!")
  - Skips prompt if user provides --no-breaking-prompt flag
  
  **Testing Criteria:**
  **Unit Tests:**
  - Breaking change prompt logic
  - Commit message formatting with breaking changes
  - Flag parsing for skip option
  
  **Integration Tests:**
  - Interactive breaking change workflow
  - Commit body formatting with BREAKING CHANGE footer
  
  **Component Browser Tests:**
  - N/A (CLI tool)

- [ ] Story-005: As a developer, I want the tool to rewrite git history with the new conventional commit messages so that my branch has clean, standardized commits.

  **Acceptance Criteria:**
  - Uses git rebase interactive or git filter-branch to rewrite commits
  - Preserves original author name, email, and timestamp
  - Preserves original committer information
  - Updates commit messages to conventional format
  - Maintains commit order and parent relationships
  - Handles merge commits appropriately (skip or warn)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Git rebase command construction
  - Author/committer preservation logic
  - Merge commit detection
  
  **Integration Tests:**
  - Full rewrite workflow on test repository
  - Verification of preserved metadata
  - Complex branch scenarios with merges
  
  **Component Browser Tests:**
  - N/A (CLI tool)

- [ ] Story-006: As a developer, I want safety checks and warnings before rewriting history so that I don't accidentally lose work or create conflicts.

  **Acceptance Criteria:**
  - Checks if branch has unpushed commits (warn if already pushed)
  - Verifies working directory is clean (no uncommitted changes)
  - Creates backup branch before rewriting (e.g., backup/original-branch-name)
  - Provides clear warning about history rewrite consequences
  - Requires explicit confirmation before proceeding
  - Provides rollback instructions if something goes wrong
  
  **Testing Criteria:**
  **Unit Tests:**
  - Safety check validation logic
  - Backup branch creation
  - User confirmation prompts
  
  **Integration Tests:**
  - Dirty working directory detection
  - Backup branch creation and verification
  - Rollback scenario testing
  
  **Component Browser Tests:**
  - N/A (CLI tool)

- [ ] Story-007: As a developer, I want the tool to follow the standard Conventional Commits specification so that my commits are compatible with automated tools.

  **Acceptance Criteria:**
  - Supports all standard types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert
  - Format: `<type>[optional scope]: <description>`
  - Optional body and footer sections preserved from original commit
  - Scope extracted from file paths or AI analysis (e.g., "api", "ui", "docs")
  - Description is clear, concise, and imperative mood ("add" not "added")
  - Maximum subject line length of 72 characters (warn if exceeded)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Conventional commit format validation
  - Type and scope parsing
  - Subject line length validation
  - Message formatting logic
  
  **Integration Tests:**
  - Various conventional commit formats
  - Edge cases (very long descriptions, special characters)
  
  **Component Browser Tests:**
  - N/A (CLI tool)

- [ ] Story-008: As a developer, I want comprehensive error handling and logging so that I can troubleshoot issues if the conversion fails.

  **Acceptance Criteria:**
  - Logs all LLM API calls and responses
  - Logs git commands executed
  - Provides clear error messages for common failures (git conflicts, API errors, invalid input)
  - Saves session log to file for debugging
  - Includes --verbose flag for detailed output
  - Graceful handling of partial failures (ability to resume)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Logging functionality
  - Error message formatting
  - Verbose flag handling
  
  **Integration Tests:**
  - Error scenarios (API timeout, git conflicts, etc.)
  - Log file creation and content verification
  
  **Component Browser Tests:**
  - N/A (CLI tool)

## Functional Requirements

- FR-1: The system SHALL analyze all commits in the current branch from the divergence point with the base branch to HEAD
- FR-2: The system SHALL use an LLM API to analyze commit diffs and determine conventional commit type, scope, and description
- FR-3: The system SHALL always prompt the user for breaking change confirmation for each commit
- FR-4: The system SHALL display a preview of all proposed changes before rewriting history
- FR-5: The system SHALL rewrite git history using git rebase or equivalent mechanism
- FR-6: The system SHALL preserve original author, committer, and timestamp metadata
- FR-7: The system SHALL create a backup branch before making any changes
- FR-8: The system SHALL verify working directory is clean before proceeding
- FR-9: The system SHALL follow the standard Conventional Commits specification format
- FR-10: The system SHALL provide comprehensive error handling and logging
- FR-11: The system SHALL warn users if commits have already been pushed to remote
- FR-12: The system SHALL handle merge commits appropriately (skip or warn)

## Non-Goals

- No support for converting commits across all branches (only current branch)
- No pattern matching or heuristic-based analysis (AI-only approach)
- No option to create new commits instead of rewriting history
- No custom conventional commit format support (standard spec only)
- No automatic detection of breaking changes (always prompt user)
- No support for other commit conventions (Angular, Ember, etc.) beyond standard Conventional Commits
- No GUI or web interface (CLI tool only)
- No commit message templates or custom rules configuration
- No integration with specific git hosting platforms (GitHub, GitLab, etc.)
- No automated testing of generated commit messages against CI/CD rules
