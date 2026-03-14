# PRD: Multi-File Agent Flattening Support

**Created**: 2026-02-04  
**Status**: In Progress  
**Priority**: High  
**Type**: Bug Fix + Enhancement

---

## Problem Statement

The current `AgentFlattenPostprocessor` assumes all agents follow a single-file pattern where the folder name matches the file name (e.g., `my-agent/my-agent.md`). However, the `wshobson-agents` repository uses **category folders** containing **multiple agent files** with different names:

```
agents/
├── backend-development/
│   ├── tdd-orchestrator.md
│   ├── temporal-python-pro.md
│   └── event-sourcing-architect.md
└── ui-design/
    ├── ui-designer.md
    └── accessibility-expert.md
```

When materializing to GitHub Copilot:
1. Postprocessor looks for `backend-development/backend-development.md` ❌
2. File doesn't exist, logs warning, skips symlink creation
3. **Result:** No symlinks created, Copilot can't detect agents

---

## Solution

Update the postprocessor to **scan for all `.md` files** in agent folders and create individual flat symlinks for each agent file, enabling GitHub Copilot to detect all agents.

**Desired Result:**
```
.github/agents/
├── backend-development/
│   ├── tdd-orchestrator.md
│   ├── temporal-python-pro.md
│   └── event-sourcing-architect.md
├── tdd-orchestrator.md → backend-development/tdd-orchestrator.md ✅
├── temporal-python-pro.md → backend-development/temporal-python-pro.md ✅
└── event-sourcing-architect.md → backend-development/event-sourcing-architect.md ✅
```

---

## Requirements

### Functional Requirements

1. **Multi-File Support**
   - Scan agent folders for all `.md` files (not just `componentName.md`)
   - Create individual symlink for each agent file found
   - Use relative paths: `filename.md` → `componentName/filename.md`

2. **File Filtering**
   - Ignore documentation files: `README.md`, `LICENSE.md`, `DOCS.md`, `CHANGELOG.md`
   - Only process top-level `.md` files (don't recurse into subdirectories)

3. **Conflict Handling**
   - If two components have files with the same name, warn and skip duplicate
   - First component wins, subsequent components log warning
   - Non-fatal error (continue materialization)

4. **Backward Compatibility**
   - Must still support single-file pattern: `my-agent/my-agent.md`
   - No regression in existing functionality

5. **Cleanup**
   - Remove **all** symlinks pointing to files within a component folder
   - Cleanup should handle multi-file case

### Non-Functional Requirements

1. **Performance:** Minimal overhead (folder scan is fast)
2. **Idempotency:** Re-running materialize should not cause errors
3. **Error Handling:** All errors properly categorized as fatal vs non-fatal
4. **Logging:** Clear messages for users about what's happening

---

## User Stories

### Story-001: Multi-File Agent Flattening

**As a developer**, I want to materialize agents from multi-file category folders so that all agents are accessible to GitHub Copilot as flat files.

**Acceptance Criteria:**
- [ ] Postprocessor scans for all `.md` files in agent folder
- [ ] Creates individual symlink for each agent file found
- [ ] Symlinks use relative paths for portability
- [ ] Documentation files (README, LICENSE, etc.) are ignored
- [ ] Works with `wshobson-agents` repository structure

**Testing:**
- Unit test: Multi-file folder creates multiple symlinks
- Unit test: Documentation files are filtered out
- Integration test: Real multi-file agent creates all symlinks

---

### Story-002: Name Conflict Detection

**As a developer**, I want clear warnings when agent files from different components have the same name so that I can understand which agents are accessible.

**Acceptance Criteria:**
- [ ] Postprocessor tracks which symlinks have been created
- [ ] If conflict detected, logs clear warning message
- [ ] First component wins, duplicate is skipped
- [ ] Non-fatal error (materialization continues)

**Testing:**
- Unit test: Name conflict logs warning and skips
- Integration test: Multiple components with same filename

---

### Story-003: Backward Compatibility

**As a developer**, I want single-file agents to continue working so that existing repositories are not affected.

**Acceptance Criteria:**
- [ ] Single-file pattern still works: `my-agent/my-agent.md`
- [ ] All existing unit tests pass
- [ ] All existing integration tests pass
- [ ] No regression in behavior

**Testing:**
- Unit test: Single-file agent creates symlink
- Integration test: Existing test scenarios pass

---

### Story-004: Multi-Symlink Cleanup

**As a developer**, I want force overwrites to clean up all symlinks from multi-file agents so that re-materialization doesn't leave orphaned symlinks.

**Acceptance Criteria:**
- [ ] Cleanup scans for all symlinks targeting component folder
- [ ] Removes all symlinks (not just single file)
- [ ] Works with both single-file and multi-file patterns
- [ ] Non-fatal errors (logs warnings but continues)

**Testing:**
- Unit test: Cleanup removes all symlinks for component
- Integration test: Force overwrite recreates all symlinks

---

## Technical Design

### Architecture Changes

**Modified Components:**
1. `PostprocessContext` - Add `SymlinkRegistry` field
2. `AgentFlattenPostprocessor.Process()` - Scan for multiple files
3. `AgentFlattenPostprocessor.Cleanup()` - Remove multiple symlinks
4. `service.go` - Initialize `SymlinkRegistry` once per materialization

### Data Structures

```go
// PostprocessContext provides context for postprocessing
type PostprocessContext struct {
    ComponentType   string
    ComponentName   string
    Target          string
    TargetDir       string
    DestPath        string
    DryRun          bool
    Formatter       *formatter.Formatter
    
    // NEW: Track symlinks created to detect conflicts
    SymlinkRegistry map[string]string // filename -> componentName
}
```

### New Helper Function

```go
// findAgentMarkdownFiles scans directory for agent .md files
// Returns absolute paths to all valid agent files
// Excludes: README.md, LICENSE.md, DOCS.md, CHANGELOG.md
func findAgentMarkdownFiles(dir string) ([]string, error)
```

### Algorithm: Process()

```
1. Scan folder for all .md files (excluding ignored patterns)
2. If no files found:
   - Log info message
   - Return nil (non-fatal)
3. For each .md file:
   a. Check SymlinkRegistry for name conflict
   b. If conflict exists:
      - Log warning with details
      - Skip this file
      - Continue to next
   c. If symlink already exists:
      - Check if target is correct (idempotent)
      - If wrong target, remove and recreate
   d. If regular file exists:
      - Return fatal error (user must resolve)
   e. Create symlink with relative path
   f. Register in SymlinkRegistry
   g. Log success message
4. Return nil
```

### Algorithm: Cleanup()

```
1. List all entries in agents directory
2. For each entry:
   a. If not a symlink, skip
   b. Read symlink target
   c. Check if target starts with "componentName/"
   d. If yes, remove symlink
   e. Log any errors as warnings (non-fatal)
3. Return nil
```

---

## Implementation Plan

### Phase 1: Core Logic Updates (45 min)

**File:** `pkg/services/materialize/postprocessor.go`
- [ ] Add `SymlinkRegistry map[string]string` to `PostprocessContext`

**File:** `pkg/services/materialize/agent_flatten_postprocessor.go`
- [ ] Add `findAgentMarkdownFiles()` helper function
- [ ] Update `Process()` to scan for multiple files
- [ ] Add conflict detection logic
- [ ] Update logging messages

**Estimated LOC:** ~150 lines

---

### Phase 2: Cleanup Updates (15 min)

**File:** `pkg/services/materialize/agent_flatten_postprocessor.go`
- [ ] Update `Cleanup()` to remove multiple symlinks
- [ ] Scan symlinks by target prefix

**Estimated LOC:** ~30 lines

---

### Phase 3: Integration Point (5 min)

**File:** `pkg/services/materialize/service.go`
- [ ] Initialize `SymlinkRegistry: make(map[string]string)` in `PostprocessContext`
- [ ] Ensure registry is shared across all postprocessors in same run

**Estimated LOC:** ~5 lines

---

### Phase 4: Unit Tests (60 min)

**File:** `pkg/services/materialize/agent_flatten_postprocessor_test.go`

**New Tests:**
1. [ ] `TestAgentFlattenPostprocessor_Process_MultipleFiles`
2. [ ] `TestAgentFlattenPostprocessor_Process_MixedFiles`
3. [ ] `TestAgentFlattenPostprocessor_Process_IgnoredFiles`
4. [ ] `TestAgentFlattenPostprocessor_Process_NameConflict`
5. [ ] `TestAgentFlattenPostprocessor_Process_NoMarkdownFiles`
6. [ ] `TestAgentFlattenPostprocessor_Cleanup_MultipleSymlinks`

**Updated Tests:**
7. [ ] `TestAgentFlattenPostprocessor_Process_Success` - Ensure backward compat

**Estimated LOC:** ~300 lines

---

### Phase 5: Integration Tests (30 min)

**File:** `tests/integration/materialize_flatten_copilot_test.go`

**New Test:**
- [ ] `TestMaterializeAgentFlatteningPostprocessor/multi-file_agent_creates_multiple_symlinks`

**Estimated LOC:** ~80 lines

---

### Phase 6: Documentation (15 min)

**File:** `tasks/20260204-agent-flattening-postprocessor.md`
- [ ] Add Story-006: Multi-File Agent Support
- [ ] Update technical design examples
- [ ] Document ignored file patterns
- [ ] Document conflict handling

**Estimated LOC:** ~50 lines

---

### Phase 7: Testing & Verification (30 min)

- [ ] Run unit tests: `go test ./pkg/services/materialize/... -v`
- [ ] Run integration tests: `go test -tags integration ./tests/integration/... -v`
- [ ] Build binary: `go build -o agent-smith`
- [ ] Install globally
- [ ] Clean test project: `rm -rf /Users/tgaines/dev/git/opencode/.github`
- [ ] Re-materialize: `agent-smith materialize all --target copilot --verbose`
- [ ] Verify symlinks: `ls -la .github/agents/ | grep "^l"`
- [ ] Count symlinks (should be 50+)
- [ ] Test readability: `cat .github/agents/tdd-orchestrator.md`

---

## Error Handling

### Non-Fatal Errors (Log Warning, Continue)
- No `.md` files found in folder
- All files are ignored (README, LICENSE, etc.)
- Name conflict (duplicate filename from different component)
- Cannot create symlink (permissions issue)
- Symlink already exists with correct target (idempotent)
- Cannot remove symlink during cleanup

### Fatal Errors (Stop Materialization)
- Regular file exists where symlink should be created (user conflict)
- Cannot stat destination path (I/O error)

---

## Edge Cases

### 1. Multi-File with One Matching Name
**Scenario:** Folder `backend-dev/` contains `backend-dev.md`, `tdd.md`, `temporal.md`
**Behavior:** Create symlinks for all 3 files (including matching name)

### 2. Empty Agent Folder
**Scenario:** Agent folder exists but contains no `.md` files
**Behavior:** Log info, skip symlink creation, non-fatal

### 3. Only Documentation Files
**Scenario:** Folder only contains `README.md`
**Behavior:** Treat as empty folder (no valid agent files)

### 4. Name Conflict
**Scenario:** `backend/api.md` and `scaffolding/api.md`
**Behavior:** First creates symlink, second warns and skips

### 5. Mixed Single and Multi-File
**Scenario:** Materializing 5 agents, 3 single-file, 2 multi-file
**Behavior:** All work correctly, symlinks created for all agent files

### 6. Force Overwrite Multi-File
**Scenario:** Force overwrite of multi-file agent
**Behavior:** Cleanup removes all 3 old symlinks, Process creates 3 new ones

---

## Test Coverage

### Unit Tests
- ✅ Multi-file folder scanning
- ✅ File filtering (ignore patterns)
- ✅ Name conflict detection
- ✅ Empty folder handling
- ✅ Backward compatibility (single-file)
- ✅ Multi-symlink cleanup
- ✅ Idempotency

### Integration Tests
- ✅ Real multi-file agent materialization
- ✅ Multiple symlinks created and readable
- ✅ Symlinks use correct relative paths
- ✅ Backward compatibility with existing tests

### Manual Testing
- ✅ `wshobson-agents` materialization
- ✅ 50+ symlinks created
- ✅ All symlinks readable via `cat`
- ✅ GitHub Copilot detection (if available)

---

## Success Criteria

- [ ] All unit tests pass (17 tests total)
- [ ] All integration tests pass (5 tests total)
- [ ] Binary compiles without errors
- [ ] Materializing `wshobson-agents` creates 50+ symlinks
- [ ] `ls -la .github/agents/ | grep "^l"` shows all symlinks
- [ ] All symlinks are readable
- [ ] No regressions with single-file agents
- [ ] GitHub Copilot can detect agents (manual verification)

---

## Files Changed

### Modified (2 files)
1. `pkg/services/materialize/postprocessor.go` - Add SymlinkRegistry
2. `pkg/services/materialize/agent_flatten_postprocessor.go` - Multi-file support

### Modified - Tests (2 files)
3. `pkg/services/materialize/agent_flatten_postprocessor_test.go` - New tests
4. `tests/integration/materialize_flatten_copilot_test.go` - Multi-file test

### Modified - Integration (1 file)
5. `pkg/services/materialize/service.go` - Initialize registry

### Modified - Documentation (2 files)
6. `tasks/20260204-agent-flattening-postprocessor.md` - Update original PRD
7. `tasks/20260204-multi-file-agent-flattening.md` - This PRD

**Total:** 7 files, ~620 LOC changes

---

## Timeline

| Phase | Task | Time | Status |
|-------|------|------|--------|
| 1 | Core logic updates | 45 min | Pending |
| 2 | Cleanup updates | 15 min | Pending |
| 3 | Integration point | 5 min | Pending |
| 4 | Unit tests | 60 min | Pending |
| 5 | Integration tests | 30 min | Pending |
| 6 | Documentation | 15 min | Pending |
| 7 | Testing & verification | 30 min | Pending |

**Total Estimate:** 3 hours

---

## Ignored File Patterns

Files with these names will be skipped during scanning:
- `README.md`
- `LICENSE.md`
- `DOCS.md`
- `CHANGELOG.md`

Case-insensitive matching (e.g., `readme.md`, `ReadMe.MD` also ignored)

---

## Conflict Warning Format

```
⚠️ Name conflict: api.md (from api-scaffolding) conflicts with existing api.md (from backend-development)
   Skipping symlink for api-scaffolding/api.md
```

---

## Dependencies

- Go 1.20+ (for `os.ReadDir`)
- Existing postprocessor infrastructure
- Agent materialization system

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Name conflicts common | Users confused about missing agents | Clear warning messages with both components |
| Performance with many files | Slower materialization | Scanning is fast; O(n) per component |
| Backward compatibility break | Existing users affected | Comprehensive testing; single-file still works |
| Registry memory usage | Large profiles use more memory | Registry only lives during materialization run |

---

## Future Enhancements

1. **Smart Conflict Resolution:** Prefix with category name instead of skipping
2. **Configurable Ignore Patterns:** Let users customize which files to skip
3. **Windows Support:** Test and verify symlink behavior on Windows
4. **Copilot Metadata:** Add metadata files for Copilot agent detection
5. **Subdirectory Support:** Optionally recurse into subdirectories

---

## References

- Original PRD: `tasks/20260204-agent-flattening-postprocessor.md`
- Postprocessor Interface: `pkg/services/materialize/postprocessor.go`
- Test Repository: `wshobson-agents` (https://github.com/wshobson/agents)

---

**Last Updated:** 2026-02-04  
**Next Review:** After implementation complete
