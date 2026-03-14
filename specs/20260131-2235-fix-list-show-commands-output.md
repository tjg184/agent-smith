# PRD: Fix List/Show Commands to Display Output Without --verbose Flag

**Created**: 2026-01-31 22:35 UTC

---

## Introduction

Currently, four CLI commands (`profile list`, `profile show`, `status`, `target list`) do not display output unless the `--verbose` or `--debug` flag is used. This creates a poor user experience where commands appear to do nothing. These commands use `infoPrintf()` which logs at `LevelInfo`, causing output to be suppressed at the default `LevelWarn` log level.

**Problem**: Users run list/show commands expecting immediate output, but see nothing without knowing to add `--verbose`.

**Solution**: Replace `infoPrintf()`/`infoPrintln()` with `fmt.Printf()`/`fmt.Println()` in these commands to match the behavior of already-working commands like `link list` and `link status`.

## Goals

- Make all list/show commands display output by default without requiring flags
- Ensure consistency across all CLI commands (match `link list` and `link status` behavior)
- Maintain backward compatibility with `--verbose` and `--debug` flags
- Improve user experience by providing immediate visual feedback

## User Stories

- [x] Story-001: As a user, I want to run `profile list` and immediately see all profiles without additional flags.

  **Acceptance Criteria:**
  - `./agent-smith profile list` displays all available profiles
  - Output includes profile names, active status indicator, and component counts
  - Legend and total count are displayed
  - Empty state shows helpful message with creation instructions
  - Works without --verbose or --debug flags
  
  **Testing Criteria:**
  **Unit Tests:**
  - N/A (output formatting, no logic changes)
  
  **Integration Tests:**
  - Test command outputs profiles when profiles exist
  - Test command shows empty state when no profiles exist
  - Test active profile indicator appears correctly
  - Test component counts display accurately
  
  **Component Browser Tests:**
  - N/A (CLI command, not a UI component)

- [x] Story-002: As a user, I want to run `profile show <name>` and immediately see detailed profile information without additional flags.

  **Acceptance Criteria:**
  - `./agent-smith profile show <name>` displays profile details
  - Output includes profile name, active status, location, and all components
  - Agents, skills, and commands are listed with counts
  - Empty profiles show helpful message with instructions
  - Works without --verbose or --debug flags
  
  **Testing Criteria:**
  **Unit Tests:**
  - N/A (output formatting, no logic changes)
  
  **Integration Tests:**
  - Test command displays full profile information
  - Test command shows empty state for profiles with no components
  - Test active status indicator appears correctly
  - Test all component types (agents, skills, commands) are listed
  
  **Component Browser Tests:**
  - N/A (CLI command, not a UI component)

- [x] Story-003: As a user, I want to run `status` and immediately see current configuration and component counts without additional flags.

  **Acceptance Criteria:**
  - `./agent-smith status` displays current configuration
  - Output includes active profile, detected targets, and component counts
  - Component counts shown for both base directory and active profile
  - Helpful links to other commands displayed at bottom
  - Works without --verbose or --debug flags
  
  **Testing Criteria:**
  **Unit Tests:**
  - N/A (output formatting, no logic changes)
  
  **Integration Tests:**
  - Test command displays active profile correctly
  - Test command shows detected targets
  - Test component counts are accurate
  - Test command handles no active profile gracefully
  
  **Component Browser Tests:**
  - N/A (CLI command, not a UI component)

- [x] Story-004: As a user, I want to run `target list` and immediately see all available targets without additional flags.

  **Acceptance Criteria:**
  - `./agent-smith target list` displays all built-in and custom targets
  - Output includes target names, base directories, and existence status
  - Built-in and custom targets are clearly separated
  - Legend explains status symbols
  - Works without --verbose or --debug flags
  
  **Testing Criteria:**
  **Unit Tests:**
  - N/A (output formatting, no logic changes)
  
  **Integration Tests:**
  - Test command displays all built-in targets
  - Test command displays custom targets when they exist
  - Test status symbols are accurate (directory exists/not found)
  - Test legend is displayed correctly
  
  **Component Browser Tests:**
  - N/A (CLI command, not a UI component)

- [x] Story-005: As a developer, I want to verify all changes are consistent and no regressions are introduced.

  **Acceptance Criteria:**
  - All four commands tested without flags show output
  - All four commands still work with --verbose flag
  - All four commands still work with --debug flag
  - Output format matches original verbose output exactly
  - No other commands are affected by changes
  
  **Testing Criteria:**
  **Unit Tests:**
  - N/A (output formatting, no logic changes)
  
  **Integration Tests:**
  - Test each command without flags
  - Test each command with --verbose flag
  - Test each command with --debug flag
  - Test commands with no profiles/targets (empty states)
  - Test commands with multiple profiles/targets
  
  **Component Browser Tests:**
  - N/A (CLI command, not a UI component)

## Functional Requirements

- FR-1: The system SHALL replace all `infoPrintf()` calls with `fmt.Printf()` in profile list handler (main.go lines 935-981)
- FR-2: The system SHALL replace all `infoPrintln()` calls with `fmt.Println()` in profile list handler (main.go lines 935-981)
- FR-3: The system SHALL replace all `infoPrintf()` calls with `fmt.Printf()` in profile show handler (main.go lines 1014-1060)
- FR-4: The system SHALL replace all `infoPrintln()` calls with `fmt.Println()` in profile show handler (main.go lines 1014-1060)
- FR-5: The system SHALL replace all `infoPrintf()` calls with `fmt.Printf()` in status handler (main.go lines 1226-1274)
- FR-6: The system SHALL replace all `infoPrintln()` calls with `fmt.Println()` in status handler (main.go lines 1226-1274)
- FR-7: The system SHALL replace all `infoPrintf()` calls with `fmt.Printf()` in target list handler (main.go lines 1360-1427)
- FR-8: The system SHALL replace all `infoPrintln()` calls with `fmt.Println()` in target list handler (main.go lines 1360-1427)
- FR-9: The system SHALL NOT modify any other `infoPrintf()` or `infoPrintln()` calls outside these four handlers
- FR-10: The system SHALL maintain exact output formatting (no visual changes to output)

## Technical Details

### Files to Modify
- **File**: `/path/to/agent-smith/main.go`
- **Scope**: Only modify the four handler functions for profile list, profile show, status, and target list

### Line Ranges to Update
1. **Profile List Handler**: Lines 916-982 (function starting at line 916)
2. **Profile Show Handler**: Lines 983-1062 (function starting at line 983)
3. **Status Handler**: Lines 1220-1276 (function starting around line 1220)
4. **Target List Handler**: Lines 1349-1428 (function starting around line 1349)

### Replacement Pattern
- Replace: `infoPrintf(...)` → `fmt.Printf(...)`
- Replace: `infoPrintln(...)` → `fmt.Println(...)`

### Intentionally NOT Changed
Other uses of `infoPrintf`/`infoPrintln` in main.go are intentionally verbose-only and should NOT be changed:
- Installation progress messages (lines 468, 517-518, 546, 591-592, 618, 663-664, etc.)
- Profile creation confirmations (lines 715, 723, 740-741, 1086-1087)
- Target add/remove confirmations (lines 1307-1314, 1346-1348)
- Component download/link operations feedback

These are operational feedback that users may want to suppress for quiet operations.

## Non-Goals

- No changes to log level system or logger implementation
- No changes to `--verbose` or `--debug` flag behavior
- No changes to operational feedback messages (downloads, installs, links)
- No changes to error or warning messages
- No changes to `link list` or `link status` commands (already working correctly)
- No visual formatting changes to output

## Success Criteria

- All four commands produce output when run without flags
- Output matches exactly what currently shows with --verbose flag
- Commands still work correctly with --verbose and --debug flags
- No regressions in other commands
- User can immediately see results from list/show commands
