# PRD: Standardize Materialize Profile Flag

## Overview

**Status**: ✅ Implemented  
**Priority**: P2 - Quality Improvement  
**Type**: Refactoring  
**Estimated Effort**: Small (1-2 hours)

## Problem Statement

The `materialize` command uses `--from-profile` flag, which is inconsistent with all other commands in the agent-smith CLI that use `--profile` / `-p`. This inconsistency creates a confusing user experience and violates the principle of least surprise.

### Current State
- **Install commands**: Use `--profile` / `-p`
- **Link commands**: Use `--profile` / `-p`
- **Unlink commands**: Use `--profile` / `-p`
- **Uninstall commands**: Use `--profile` / `-p`
- **Update command**: Use `--profile` / `-p`
- **Materialize commands**: Use `--from-profile` (no short flag) ❌

### Pain Points
1. Users must remember different flag names for different commands
2. No short flag `-p` available for materialize (requires more typing)
3. Documentation is inconsistent across commands
4. Violates CLI design principle: consistent interface patterns

## Goals

1. **Consistency**: All commands should use `--profile` / `-p` flag
2. **Simplicity**: Shorter flag name and addition of `-p` shortcut
3. **Backward Incompatible**: Clean break (no backward compatibility needed)
4. **Preserve Behavior**: Keep special `base` value for forcing base directory

## Non-Goals

1. Maintaining backward compatibility with `--from-profile`
2. Changing the underlying profile selection logic
3. Modifying how active profiles work
4. Updating test behavior (tests don't use CLI flags directly)

## Success Metrics

- ✅ All materialize commands use consistent `--profile` / `-p` flag
- ✅ All documentation updated to reflect new flag name
- ✅ Code compiles without errors
- ✅ Existing behavior preserved (optional flag, falls back to active profile)

## Technical Design

### Flag Specification

**Old (Inconsistent)**:
```bash
agent-smith materialize skill api-design --target opencode --from-profile work
```

**New (Consistent)**:
```bash
agent-smith materialize skill api-design --target opencode --profile work
agent-smith materialize skill api-design --target opencode -p work  # Short flag
```

### Affected Commands

All materialize subcommands:
- `materialize skill <name>`
- `materialize agent <name>`
- `materialize command <name>`
- `materialize all`

### Special Value: `base`

The special value `--profile base` (formerly `--from-profile base`) will be preserved. This allows users to force materialization from the base `~/.agent-smith/` directory even when a profile is active.

**Example**:
```bash
# Even if "work" profile is active, materialize from base
agent-smith materialize skill standard-tool --target opencode --profile base
```

### Behavior Specification

| Scenario | Flag Value | Active Profile | Source Directory |
|----------|------------|----------------|------------------|
| No flag specified | - | None | `~/.agent-smith/` |
| No flag specified | - | `work` | `~/.agent-smith/profiles/work/` |
| Flag specified | `personal` | `work` | `~/.agent-smith/profiles/personal/` (override) |
| Flag specified | `base` | `work` | `~/.agent-smith/` (force base) |

**Logic Flow**:
1. If `--profile <name>` is specified:
   - If `<name>` is `base` → use `~/.agent-smith/`
   - Else validate profile exists → use `~/.agent-smith/profiles/<name>/`
2. Else (no flag):
   - If active profile exists → use active profile directory
   - Else → use base `~/.agent-smith/`

### Implementation Changes

#### 1. Flag Definitions (`cmd/root.go`)
**Changes Required**: 4 commands × 1 line each

```go
// OLD
materializeSkillCmd.Flags().String("from-profile", "", "Materialize from specific profile (use 'base' for ~/.agent-smith/)")

// NEW
materializeSkillCmd.Flags().StringP("profile", "p", "", "Materialize from specific profile (use 'base' for ~/.agent-smith/)")
```

**Lines to change**:
- Line 1661: `materialize skill`
- Line 1701: `materialize agent`
- Line 1741: `materialize command`
- Line 1784: `materialize all`

#### 2. Flag Reading (`cmd/root.go`)
**Changes Required**: 4 commands × 2 lines each (read flag + pass to function)

```go
// OLD
fromProfile, _ := cmd.Flags().GetString("from-profile")
handleMaterializeComponent("skills", args[0], target, projectDir, force, dryRun, fromProfile)

// NEW
profile, _ := cmd.Flags().GetString("profile")
handleMaterializeComponent("skills", args[0], target, projectDir, force, dryRun, profile)
```

#### 3. Implementation Logic (`main.go`)
**Changes Required**: 2 function closures with identical logic

**Function signature**:
```go
// OLD
func(componentType, componentName, target, projectDir string, force, dryRun bool, fromProfile string)

// NEW
func(componentType, componentName, target, projectDir string, force, dryRun bool, profile string)
```

**Variable renames**:
- `fromProfile` → `profile` (parameter and all references)

**Comment updates**:
- `--from-profile flag` → `--profile flag`
- `via --from-profile` → `via --profile`

#### 4. Help Text Examples (`cmd/root.go`)
**Line 1766**:
```bash
# OLD
agent-smith materialize all --target opencode --from-profile work

# NEW
agent-smith materialize all --target opencode --profile work
```

#### 5. Documentation (`README.md`)
**Lines 340, 343**:
```bash
# OLD
agent-smith materialize skill api-design --target opencode --from-profile work
agent-smith materialize skill mcp-builder --target opencode --from-profile base

# NEW
agent-smith materialize skill api-design --target opencode --profile work
agent-smith materialize skill mcp-builder --target opencode --profile base
```

#### 6. Task Documentation (`tasks/20260202-0241-component-materialization.md`)
**Updates Required**: ~10 occurrences

- Acceptance criteria (lines 329, 331)
- Testing criteria (line 343)
- Functional requirements FR-8, FR-9 (line 388)
- CLI parameter documentation (line 531)
- Scenario examples (lines 598, 606, 634, 642, 735-737)

### Files Modified

| File | Type | Lines Changed |
|------|------|---------------|
| `cmd/root.go` | Code | ~16 lines (8 flag definitions + 8 flag reads) |
| `main.go` | Code | ~20 lines (signatures + comments + logic) |
| `README.md` | Docs | 2 lines |
| `tasks/20260202-0241-component-materialization.md` | Docs | ~10 lines |
| **Total** | - | **~48 lines** |

### Testing Strategy

#### Verification Steps

1. **Compile check**:
   ```bash
   go build -o /tmp/agent-smith .
   ```

2. **Help text verification**:
   ```bash
   ./agent-smith materialize skill --help
   ./agent-smith materialize agent --help
   ./agent-smith materialize command --help
   ./agent-smith materialize all --help
   ```
   
   Expected: All should show `--profile, -p` flag

3. **Functional testing**:
   ```bash
   # Test default behavior (uses active profile)
   agent-smith profile activate work
   agent-smith materialize skill test-skill --target opencode
   
   # Test explicit profile with long flag
   agent-smith materialize skill test-skill --target opencode --profile personal
   
   # Test explicit profile with short flag
   agent-smith materialize skill test-skill --target opencode -p personal
   
   # Test base override
   agent-smith materialize skill test-skill --target opencode --profile base
   ```

4. **Integration tests**:
   ```bash
   go test -tags=integration ./tests/integration/materialize_*.go
   ```
   
   Expected: All tests pass (no changes needed to tests)

## Risks & Mitigations

### Risk: Breaking Change
**Impact**: High  
**Likelihood**: High  
**Mitigation**: This is intentional. No backward compatibility maintained. Users will see clear error if using old flag.

**Error message**:
```
Error: unknown flag: --from-profile
```

Users can easily fix by replacing `--from-profile` with `--profile` or `-p`.

### Risk: Documentation Drift
**Impact**: Medium  
**Likelihood**: Low  
**Mitigation**: All documentation updated in single PR. Searched codebase for all occurrences.

## Open Questions

None - implementation complete.

## Implementation Checklist

- [x] Update flag definitions in `cmd/root.go` (4 commands)
- [x] Update flag reading in `cmd/root.go` (4 commands)
- [x] Update function signatures in `main.go`
- [x] Update variable names and comments in `main.go`
- [x] Update help text examples in `cmd/root.go`
- [x] Update `README.md` documentation
- [x] Update task documentation
- [x] Verify code compiles
- [ ] Manual functional testing
- [ ] Run integration test suite

## Future Considerations

1. **Deprecation Warning**: If we ever want to add backward compatibility in the future, we could add a hidden `--from-profile` alias that prints a deprecation warning.

2. **Tab Completion**: Consider adding shell completion scripts that suggest `-p` as the short flag for profile across all commands.

3. **Migration Guide**: If this breaks many user scripts, consider adding a migration guide or script to help users update their automation.

## References

- Original issue discussion: User feedback on CLI inconsistency
- Related PR: Materialize command implementation
- CLI Design Principles: [POSIX/GNU Guidelines](https://www.gnu.org/prep/standards/html_node/Command_002dLine-Interfaces.html)

---

**Document Version**: 1.0  
**Last Updated**: 2025-01-02  
**Author**: Agent Smith Development Team
