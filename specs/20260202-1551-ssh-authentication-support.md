# PRD: SSH Authentication Support for Private Repositories

**Created**: 2026-02-02 15:51 UTC

---

## Introduction

Agent-smith currently fails when cloning private repositories via SSH URLs (e.g., `git@github.com:owner/repo.git`) because go-git doesn't automatically use the system's SSH agent like the `git` CLI does. This prevents users from installing components from private repositories, even though SSH authentication is properly configured on their system.

## Problem Statement

When running `agent-smith install all git@github.com:tjg184/skills.git`, the command fails with:
```
Failed to bulk download components: failed to clone repository for bulk detection: authentication required: Repository not found
```

This occurs because:
1. The go-git library requires explicit SSH authentication configuration
2. Agent-smith's `git.PlainClone` calls don't include an `Auth` field
3. The system's SSH agent is available and working (verified by successful `git clone` operations)

## Goals

- Enable SSH authentication for private repository access via ssh-agent
- Maintain compatibility with existing public repository installations
- Ensure all git clone operations throughout the codebase use authentication
- Avoid breaking HTTPS or local path repository access

## User Stories

- [ ] Story-001: As a developer, I want to install components from private SSH repositories so that I can use agent-smith with my organization's private component libraries.

  **Acceptance Criteria:**
  - SSH URLs (git@github.com:owner/repo.git format) successfully authenticate via ssh-agent
  - Public repositories continue to work without authentication
  - Existing HTTPS and local path installations remain functional
  - Authentication errors provide clear, actionable error messages
  
  **Testing Criteria:**
  **Unit Tests:**
  - Test `GetAuthMethod` returns SSH auth for SSH URLs
  - Test `GetAuthMethod` returns nil for public/local repos
  - Test `isSSHURL` correctly identifies SSH URL formats (git@ and ssh://)
  - Test graceful fallback when ssh-agent is unavailable

- [ ] Story-002: As a developer, I want authentication to work consistently across all install commands so that I don't encounter different behavior between `install all`, `install skill`, `install agent`, etc.

  **Acceptance Criteria:**
  - BulkDownloader uses authentication for bulk installs
  - SkillDownloader uses authentication for skill installs  
  - AgentDownloader uses authentication for agent installs
  - CommandDownloader uses authentication for command installs
  - Updater uses authentication for update operations
  - All code paths use the centralized auth helper
  
  **Testing Criteria:**
  **Unit Tests:**
  - Test `ToGoGitCloneOptions` adds auth to CloneOptions for SSH URLs
  - Test auth helper is used in all downloader implementations
  - Note: Actual cloning is covered by integration tests in existing codebase

- [ ] Story-003: As a developer, I want clear error messages when SSH authentication fails so that I can quickly diagnose and fix configuration issues.

  **Acceptance Criteria:**
  - Missing ssh-agent errors suggest checking SSH configuration
  - Missing SSH keys errors indicate which key locations were checked
  - Authentication failures distinguish between "no auth" and "wrong auth"
  - Error messages include suggestions for resolution (e.g., "Check your SSH key configuration")
  
  **Testing Criteria:**
  **Unit Tests:**
  - Test error messages from `getSSHAuth` when ssh-agent unavailable
  - Test error messages from `getDefaultSSHKeys` when no keys found
  - Test error handling in `ToGoGitCloneOptions` when auth fails

## Functional Requirements

- **FR-1**: The system SHALL attempt SSH agent authentication when detecting SSH URLs (starting with "git@" or "ssh://")

- **FR-2**: The system SHALL fall back to default SSH key file authentication (~/.ssh/id_rsa, ~/.ssh/id_ed25519, etc.) when ssh-agent is unavailable

- **FR-3**: The system SHALL return nil auth for HTTPS and local path URLs to maintain current behavior for public repositories

- **FR-4**: The system SHALL apply authentication consistently across all git clone operations in:
  - `internal/downloader/bulk.go`
  - `internal/downloader/agent.go`
  - `internal/downloader/skill.go`
  - `internal/downloader/command.go`
  - `internal/updater/updater.go`

- **FR-5**: The system SHALL provide clear error messages when SSH authentication fails, including suggestions for resolution

- **FR-6**: The system SHALL not break existing functionality for public repositories, HTTPS URLs, or local paths

## Implementation Details

### New File: `internal/git/auth.go`

Create a new authentication helper module with:

- `GetAuthMethod(url string)` - Main entry point, returns appropriate auth for URL type
- `isSSHURL(url string)` - Detects SSH URLs (git@ or ssh:// format)
- `isHTTPSURL(url string)` - Detects HTTPS URLs
- `getSSHAuth()` - Attempts ssh-agent authentication
- `getDefaultSSHKeys()` - Falls back to ~/.ssh key files
- `isSSHAgentAvailable()` - Checks if SSH_AUTH_SOCK is accessible

### Modified Files

1. **`internal/git/clone.go`**:
   - Update `ToGoGitCloneOptions()` to call `GetAuthMethod()` and add auth to CloneOptions

2. **`internal/downloader/bulk.go`**:
   - Add `gitpkg` import
   - Update `PlainClone` call to include auth from `GetAuthMethod()`

3. **`internal/downloader/agent.go`**:
   - Update `PlainClone` calls (lines ~142 and ~319) to include auth

4. **`internal/downloader/skill.go`**:
   - Already uses `CloneShallow()` helper - auth will work automatically via `ToGoGitCloneOptions()`

5. **`internal/downloader/command.go`**:
   - Update `PlainClone` calls (lines ~131 and ~295) to include auth

6. **`internal/updater/updater.go`**:
   - Update `PlainClone` call (line ~156) to include auth

### Files Already Complete

- ✅ `internal/git/auth.go` - Created with comprehensive SSH auth support
- ✅ `internal/git/clone.go` - Updated `ToGoGitCloneOptions()` to use auth
- ✅ `internal/downloader/bulk.go` - Updated with auth and import

### Files Remaining

- ⏳ `internal/downloader/agent.go` - 2 PlainClone calls need auth
- ⏳ `internal/downloader/command.go` - 2 PlainClone calls need auth
- ⏳ `internal/updater/updater.go` - 1 PlainClone call needs auth

## Non-Goals (Out of Scope)

- **HTTPS token authentication** - Not implementing GITHUB_TOKEN or credential helper support in this PRD (can be added later if needed)
- **SSH key passphrase prompts** - Will use ssh-agent for passwordless auth; passphrase-protected keys must be added to ssh-agent first
- **Custom SSH key paths** - Will only check standard ~/.ssh/ locations
- **Git credential helper integration** - Not using git's credential system
- **SSH key generation** - Users must have SSH keys already configured
- **Multi-factor authentication** - SSH auth only, no OTP or other MFA methods

## Success Criteria

The feature is complete when:

1. Running `agent-smith install all git@github.com:tjg184/skills.git` successfully clones the private repository
2. All unit tests pass for the new auth module
3. Existing public repository installations continue to work without changes
4. All git clone operations in the codebase use the authentication helper
5. Clear error messages guide users when SSH authentication fails

## Testing Strategy

### Unit Tests (Required)

- Test `GetAuthMethod()` for different URL types (SSH, HTTPS, local)
- Test SSH URL detection (`isSSHURL()`)
- Test ssh-agent availability checks
- Test error handling for missing ssh-agent and missing keys
- Test `ToGoGitCloneOptions()` adds auth for SSH URLs
- Test `ToGoGitCloneOptions()` doesn't add auth for public repos

### Manual Verification (Post-Implementation)

1. Verify private SSH repository cloning works
2. Verify public repositories still work
3. Verify HTTPS URLs still work
4. Verify local paths still work
5. Verify error messages are helpful when SSH auth fails

## Dependencies

- Existing `github.com/go-git/go-git/v5` dependency (already in go.mod)
- Existing `github.com/xanzy/ssh-agent` dependency (already in go.mod)
- System ssh-agent must be running and have keys loaded

## Rollout Plan

1. Implement and test auth module
2. Update all downloader and updater files
3. Run unit tests to verify functionality
4. Manual testing with private repository
5. Commit and document changes
