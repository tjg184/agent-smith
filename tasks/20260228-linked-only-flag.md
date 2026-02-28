# Add --linked-only Flag to agent-smith link status

**Status:** ✅ Complete  
**Created:** 2026-02-28  
**Completed:** 2026-02-28

## Problem

When running `agent-smith link status --all-profiles`, the output can be very long and difficult to parse because it shows all components, including those that aren't linked to any targets (showing `-` across all columns). Users need a way to focus on only the components that are actually linked.

## Solution

Add a `--linked-only` flag to the `agent-smith link status` command that filters out components with no links across any targets. The flag works with both single-profile and all-profiles views while maintaining full backward compatibility.

## Requirements

### Functional Requirements

- [ ] ✅ Add `--linked-only` flag to `agent-smith link status` command
- [ ] ✅ Filter out components where all targets show `-` (not linked)
- [ ] ✅ Show components with at least one link (`✓`, `◆`, or `✗`)
- [ ] ✅ Work with single-profile view
- [ ] ✅ Work with `--all-profiles` view
- [ ] ✅ Work with `--profile` filter
- [ ] ✅ Maintain backward compatibility (default shows all components)

### Non-Functional Requirements

- [ ] ✅ No breaking changes to existing behavior
- [ ] ✅ All existing tests pass
- [ ] ✅ New tests cover filtering scenarios
- [ ] ✅ Clear help text and examples
- [ ] ✅ Minimal performance overhead

## Technical Design

### Architecture

The feature adds an optional boolean flag that flows through these layers:

1. **CLI Layer** (`cmd/root.go`): Flag definition and parsing
2. **Handler Layer** (`main.go`): Parameter passing
3. **Service Layer** (`pkg/services/`): Options struct and routing
4. **Display Layer** (`internal/linker/`): Filtering logic

### Implementation Details

#### 1. Options Struct
**File:** `pkg/services/interfaces.go`

```go
type LinkStatusOptions struct {
    AllProfiles   bool     // Show status for all profiles
    ProfileFilter []string // Filter by specific profile names
    LinkedOnly    bool     // Show only components with at least one link
}
```

#### 2. Flag Definition
**File:** `cmd/root.go`

```go
linkStatusCmd.Flags().Bool("linked-only", false, "Show only components that have at least one link")
```

#### 3. Filtering Logic
**Files:** `internal/linker/linker.go` (2 methods)

- `ShowLinkStatus(linkedOnly bool)` - Single profile view
- `ShowAllProfilesLinkStatus(profileFilter []string, linkedOnly bool)` - All profiles view

Both methods check each component's link status and skip rendering if `linkedOnly=true` and all targets show `-`.

### Definition of "Linked"

- **Linked:** Component has at least one target showing `✓`, `◆`, or `✗`
- **Unlinked:** All targets show `-`

## Implementation Tasks

### Phase 1: Core Implementation ✅
- [x] Update `LinkStatusOptions` struct in `pkg/services/interfaces.go`
- [x] Add `--linked-only` flag to `cmd/root.go`
- [x] Update handler signature in `main.go` (3 locations)
- [x] Update service to pass `LinkedOnly` in `pkg/services/link/service.go`
- [x] Add filtering to `ShowLinkStatus()` in `internal/linker/linker.go`
- [x] Add filtering to `ShowAllProfilesLinkStatus()` in `internal/linker/linker.go`

### Phase 2: Testing ✅
- [x] Update existing unit tests (`show_link_status_test.go`)
- [x] Update profile manager tests (`profile_manager_test.go`)
- [x] Create new test file `linked_only_test.go` with:
  - [x] Test filtering with mixed linked/unlinked components
  - [x] Test edge case: all components unlinked
  - [x] Test edge case: partially linked components
- [x] Run all linker tests - all passing
- [x] Manual testing of flag combinations

### Phase 3: Documentation ✅
- [x] Update command help text with examples
- [x] Create PRD document

## Files Modified

1. ✅ `pkg/services/interfaces.go` - Added `LinkedOnly` field
2. ✅ `cmd/root.go` - Added flag, updated examples, updated handler signature
3. ✅ `main.go` - Updated handler implementation
4. ✅ `pkg/services/link/service.go` - Pass `LinkedOnly` to linker methods
5. ✅ `internal/linker/linker.go` - Added filtering logic to 2 methods
6. ✅ `internal/linker/show_link_status_test.go` - Updated test calls
7. ✅ `internal/linker/profile_manager_test.go` - Updated test calls
8. ✅ `internal/linker/linked_only_test.go` - NEW: Comprehensive tests

## Testing

### Unit Tests ✅
All tests passing in `./internal/linker/...`:
- `TestShowLinkStatus_LinkedOnly` - Verifies filtering works correctly
- `TestShowLinkStatus_LinkedOnlyAllUnlinked` - Tests all unlinked edge case
- `TestShowLinkStatus_LinkedOnlyMixedStatuses` - Tests partial linking
- All existing tests updated and passing

### Manual Testing ✅
- ✅ `agent-smith link status` - Shows all components (unchanged)
- ✅ `agent-smith link status --linked-only` - Filters to linked only
- ✅ `agent-smith link status --all-profiles --linked-only` - Works with all profiles
- ✅ `agent-smith link status --profile work --linked-only` - Works with profile filter
- ✅ Help text displays correctly
- ✅ Legend and summary display correctly

## Usage Examples

```bash
# Default - shows all components (unchanged behavior)
agent-smith link status

# Show only linked components (single profile)
agent-smith link status --linked-only

# Show only linked components (all profiles)
agent-smith link status --all-profiles --linked-only

# Show only linked components for specific profile
agent-smith link status --profile myprofile --linked-only

# Combine with multiple profiles
agent-smith link status --all-profiles --profile work,dev --linked-only
```

## Success Metrics

- ✅ All existing tests pass
- ✅ New tests cover flag behavior
- ✅ Manual testing shows correct filtering
- ✅ No breaking changes to existing behavior
- ✅ Binary builds successfully

## Risks & Mitigations

| Risk | Impact | Mitigation | Status |
|------|--------|------------|--------|
| Breaking existing behavior | High | Comprehensive test coverage, optional flag | ✅ Mitigated |
| Performance impact | Low | Simple iteration, minimal overhead | ✅ No impact |
| User confusion | Medium | Clear flag description, maintain legend | ✅ Documented |

## References

- User request: Long output hard to read with many unlinked components
- Related commands: `agent-smith link status`, `agent-smith link list`

## Notes

Feature is complete and working as expected. The `--linked-only` flag successfully filters out unlinked components while maintaining full backward compatibility.
