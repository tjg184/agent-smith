# PRD: Fix Base Installation Unlink Filtering

**Created**: 2026-02-07 18:30 UTC

---

## Introduction

When running `agent-smith unlink all` with no active profile (using base installation at `~/.agent-smith/`), the command incorrectly unlinks components from all profiles instead of only base components. This happens because the `isSymlinkFromCurrentProfile()` function uses prefix matching that incorrectly matches profile paths (which are subdirectories of the base installation).

## Goals

- Fix the `isSymlinkFromCurrentProfile()` function to correctly distinguish base components from profile components
- Ensure base installation only unlinks base components, not profile components
- Maintain backward compatibility for profile-based unlink operations
- Add unit tests to prevent regression

## User Stories

- [ ] Story-001: As a user with no active profile, when I run `unlink all`, only base components are unlinked so that profile components remain untouched.

  **Acceptance Criteria:**
  - Running `agent-smith unlink all` without an active profile only unlinks components from `~/.agent-smith/`
  - Components from `~/.agent-smith/profiles/*` are correctly skipped with appropriate messaging
  - The confirmation message shows the correct count of base-only components
  - Existing profile-based unlink functionality continues to work correctly
  
  **Testing Criteria:**
  **Unit Tests:**
  - Test `isSymlinkFromCurrentProfile()` returns `false` for profile symlinks when using base installation
  - Test `isSymlinkFromCurrentProfile()` returns `true` for base symlinks when using base installation
  - Test `isSymlinkFromCurrentProfile()` returns `true` for matching profile symlinks when using a profile
  
  **Integration Tests:**
  - Test `unlink all` with no active profile only removes base component symlinks
  - Test `unlink all` with active profile only removes that profile's symlinks

## Functional Requirements

- FR-1: The `isSymlinkFromCurrentProfile()` function SHALL use profile name comparison instead of path prefix matching
- FR-2: When using base installation (`cl.agentsDir` = `~/.agent-smith`), the function SHALL only match symlinks pointing to paths where `getProfileFromPath()` returns `"base"`
- FR-3: When using a profile, the function SHALL only match symlinks pointing to that specific profile
- FR-4: The fix SHALL NOT change any other unlink functionality or command-line interface

## Non-Goals

- No changes to the command-line interface or flags
- No support for `--profile base` flag (out of scope for this bug fix)
- No changes to symlink detection or other linker functionality
- No changes to profile management or activation logic

## Technical Details

### Root Cause

In `/path/to/agent-smith/internal/linker/linker.go`, the `isSymlinkFromCurrentProfile()` function:

```go
return strings.HasPrefix(target, agentsDir), nil
```

When `agentsDir` = `/Users/<user>/.agent-smith` (base installation), and a symlink points to `/Users/<user>/.agent-smith/profiles/work/skills/mcp-builder`, the prefix check returns `true` because profile directories are physically inside `~/.agent-smith/`.

### Solution

Replace prefix matching with explicit profile name comparison:

```go
// Get profile names and compare
currentProfile := getProfileFromPath(cl.agentsDir)
targetProfile := GetProfileNameFromSymlink(symlinkPath)
return currentProfile == targetProfile, nil
```

This ensures:
- Base installation (`currentProfile == "base"`) only matches symlinks pointing to base directories (paths not under `profiles/`)
- Profile installations only match symlinks pointing to that specific profile
- The logic is consistent and predictable

## References

- Bug location: `/path/to/agent-smith/internal/linker/linker.go:1658-1677`
- Related function: `getProfileFromPath()` in `/path/to/agent-smith/internal/linker/status.go:32`
- Related function: `GetProfileNameFromSymlink()` in `/path/to/agent-smith/internal/linker/status.go:52`
