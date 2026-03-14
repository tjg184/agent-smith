# PRD: Materialize Info/List Output Consistency

## Overview

**Problem**: The `agent-smith materialize info` and `agent-smith materialize list` commands don't show output without `--verbose` or `--debug` flags, creating a poor user experience and inconsistency with other informational commands like `profile list`.

**Solution**: Replace `infoPrintf()`/`infoPrintln()` calls with `appFormatter.Info()` in the materialize info/list command handlers to ensure output is always visible, consistent with other informational commands.

## Background

### Current Behavior

When users run informational commands:

| Command | Shows Output by Default? | Method Used |
|---------|-------------------------|-------------|
| `agent-smith profile list` | ✅ YES | `appFormatter.Info()` |
| `agent-smith materialize list` | ❌ NO (requires `--verbose`) | `infoPrintf()` → `appLogger.Info()` |
| `agent-smith materialize info` | ❌ NO (requires `--verbose`) | `infoPrintf()` → `appLogger.Info()` |

### Root Cause

1. The logger defaults to `LevelWarn` when neither `--debug` nor `--verbose` flags are present
2. `infoPrintf()` calls `appLogger.Info()` which logs at `LevelInfo`
3. `LevelInfo` messages are filtered out when log level is `LevelWarn`
4. Error messages appear because they use `fmt.Println()` with structured error objects that bypass the logger

### Why This Matters

- **User Expectation**: When running a command specifically designed to display information (e.g., `info`, `list`), users expect to see output without additional flags
- **Consistency**: Other informational commands like `profile list` show output by default
- **Discoverability**: Users may think the command is broken when it shows nothing

## Goals

1. **Primary**: Ensure `materialize info` and `materialize list` always show output
2. **Secondary**: Maintain consistency with other informational commands (`profile list`, `profile show`)
3. **Tertiary**: Keep `--verbose`/`--debug` flags functional for additional diagnostic information

## Non-Goals

- Changing the default log level globally (would affect all commands)
- Modifying progress/status messages that should respect `--verbose`
- Changing error message handling

## Success Metrics

- Users can see output from `materialize info` without flags
- Users can see output from `materialize list` without flags
- Behavior matches `profile list` command
- All existing integration tests pass

## User Experience

### Before

```bash
$ agent-smith materialize info skills my-skill
[no output - appears broken]

$ agent-smith materialize info skills my-skill --verbose
✓ Provenance Information - OpenCode (.opencode/)

  Component: my-skill
  Type: skills
  ...
```

### After

```bash
$ agent-smith materialize info skills my-skill
✓ Provenance Information - OpenCode (.opencode/)

  Component: my-skill
  Type: skills
  
  Source Information:
    Repository: https://github.com/example/repo
    Source Type: github
    Commit Hash: abc123def456
    Original Path: skills/my-skill/SKILL.md
  
  Materialization:
    Materialized At: 2024-01-15T10:30:00Z
    Target Directory: /path/to/project/.opencode
  
  Sync Status:
    Source Hash: hash123...
    Current Hash: hash123...
    Status: In Sync (component is unchanged)
```

## Technical Design

### Implementation Approach

Replace `infoPrintf()` and `infoPrintln()` calls with `appFormatter.Info()` in:

1. **materialize info handler** (main.go, lines ~3079-3125)
2. **materialize list handler** (main.go, lines ~2912-2993)

### Code Changes

#### File: `main.go`

**Section 1: materialize list handler (~line 2912-2993)**

Replace:
```go
infoPrintf("Materialized Components in %s:\n\n", projectRoot)
infoPrintf("%s %s\n", green(formatter.SymbolSuccess), targetLabel)
infoPrintf("  Skills (%d):\n", len(metadata.Skills))
infoPrintf("    • %-30s (from %s)\n", name, sourceInfo)
infoPrintln("")
```

With:
```go
appFormatter.Info("Materialized Components in %s:\n", projectRoot)
appFormatter.Info("%s %s", green(formatter.SymbolSuccess), targetLabel)
appFormatter.Info("  Skills (%d):", len(metadata.Skills))
appFormatter.Info("    • %-30s (from %s)", name, sourceInfo)
appFormatter.EmptyLine()
```

**Section 2: materialize info handler (~line 3079-3125)**

Replace:
```go
infoPrintf("\n%s Provenance Information - %s\n\n", green(formatter.SymbolSuccess), bold(targetLabel))
infoPrintf("  %s: %s\n", cyan("Component"), componentName)
infoPrintln("")
```

With:
```go
appFormatter.EmptyLine()
appFormatter.Info("%s Provenance Information - %s", green(formatter.SymbolSuccess), bold(targetLabel))
appFormatter.EmptyLine()
appFormatter.Info("  %s: %s", cyan("Component"), componentName)
appFormatter.EmptyLine()
```

### Testing Strategy

1. **Manual Testing**:
   - Run `agent-smith materialize info skills <name>` without flags - should show output
   - Run `agent-smith materialize list` without flags - should show output
   - Verify error cases still show clear messages

2. **Integration Tests**:
   - All existing `materialize_info_test.go` tests should pass
   - All existing `materialize_list_profile_test.go` tests should pass
   - Tests use `--verbose` but should work without it after fix

3. **Regression Testing**:
   - Verify `--verbose` and `--debug` flags still work
   - Verify other commands aren't affected
   - Verify `profile list` still works as expected

## Implementation Plan

### Phase 1: Update materialize list handler
1. Replace all `infoPrintf()` calls with `appFormatter.Info()` in materialize list handler
2. Replace all `infoPrintln("")` calls with `appFormatter.EmptyLine()`
3. Test manually without flags

### Phase 2: Update materialize info handler
1. Replace all `infoPrintf()` calls with `appFormatter.Info()` in materialize info handler
2. Replace all `infoPrintln("")` calls with `appFormatter.EmptyLine()`
3. Test manually without flags

### Phase 3: Testing & Validation
1. Run integration tests: `go test -tags=integration ./tests/integration/materialize_info_test.go`
2. Run integration tests: `go test -tags=integration ./tests/integration/materialize_list_profile_test.go`
3. Manual testing of both commands with and without flags
4. Verify error cases

### Phase 4: Documentation
1. Update any documentation that mentions `--verbose` requirement for these commands
2. Update help text if needed

## Alternatives Considered

### Alternative 1: Change default log level to LevelInfo
**Rejected**: Would affect ALL commands and could clutter output for operations where verbose should be opt-in.

### Alternative 2: Command-specific log level override
**Rejected**: More complex implementation, harder to maintain, less obvious to future developers.

### Alternative 3: Use fmt.Print() directly
**Rejected**: Bypasses the formatter system which provides consistent styling and structure.

## Dependencies

- No external dependencies
- Uses existing `appFormatter.Info()` infrastructure
- Compatible with current logging system

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Breaking existing workflows | Low | Commands now show expected output; no breaking changes |
| Test failures | Medium | Review and update test expectations if needed |
| Inconsistent formatting | Low | `appFormatter` provides consistent formatting |

## Open Questions

None - implementation path is clear and consistent with existing patterns.

## Appendix

### Related Code Locations

- **materialize list handler**: `main.go` lines 2891-2995
- **materialize info handler**: `main.go` lines 2998-3156  
- **logger implementation**: `pkg/logger/logger.go`
- **formatter implementation**: `pkg/formatter/formatter.go`
- **integration tests**: `tests/integration/materialize_info_test.go`, `tests/integration/materialize_list_profile_test.go`

### Example Commands Tested

```bash
# Should show output after fix
agent-smith materialize info skills agent-smith-profile-builder
agent-smith materialize list

# Should still work with flags
agent-smith materialize info skills my-skill --verbose
agent-smith materialize info skills my-skill --debug
agent-smith materialize list --verbose
```
