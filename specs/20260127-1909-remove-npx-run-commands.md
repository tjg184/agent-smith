# PRD: Remove npx and run Commands from Agent Smith

## Introduction

Remove the `npx` and `run` commands along with the entire executor package from Agent Smith. These commands were designed to execute components without installing them, but they fundamentally conflict with the tool's purpose: skills, agents, and commands are markdown documentation files for AI coding environments, not executable programs. The executor package attempts to find and run executable files, which doesn't align with documentation-based components.

## Goals

- Remove non-functional npx/run commands that don't align with component purpose
- Delete entire executor package (~350+ lines) to reduce codebase complexity
- Update help text to accurately reflect tool capabilities
- Simplify Agent Smith's purpose: download, manage, and link AI component documentation
- Maintain all core functionality (install, link, update, profiles)

## User Stories

- [x] Story-001: As a developer, I want the npx command removed from Agent Smith so that the CLI doesn't offer functionality that doesn't work with documentation-based components.

  **Acceptance Criteria:**
  - `npx` command definition removed from cmd/root.go (lines 169-176)
  - Attempting `./agent-smith npx <target>` returns command not found error
  - Help text (`./agent-smith --help`) does not list npx command
  - Build succeeds with no compilation errors

  **Testing Criteria:**
  **Unit Tests:**
  - No unit tests exist for executor package (verify removal is clean)
  
  **Integration Tests:**
  - Build test: `go build` succeeds without errors
  - Command verification: `./agent-smith --help` does not show npx
  - Error handling: `./agent-smith npx test` returns appropriate error
  
  **Manual Testing:**
  - Help output verification
  - Command discovery testing

- [x] Story-002: As a developer, I want the run command removed from Agent Smith so that the CLI has a clear, focused purpose.

  **Acceptance Criteria:**
  - `run` command definition removed from cmd/root.go (lines 178-185)
  - Attempting `./agent-smith run <target>` returns command not found error
  - Help text does not list run command
  - Build succeeds with no compilation errors

  **Testing Criteria:**
  **Unit Tests:**
  - No unit tests exist for run command (verify removal is clean)
  
  **Integration Tests:**
  - Build test: `go build` succeeds without errors
  - Command verification: `./agent-smith --help` does not show run
  - Error handling: `./agent-smith run test` returns appropriate error
  
  **Manual Testing:**
  - Help output verification
  - Command discovery testing

- [x] Story-003: As a developer, I want the executor package removed so that the codebase doesn't contain unused infrastructure for executing files.

  **Acceptance Criteria:**
  - Delete entire `internal/executor/` directory
  - Remove `internal/executor/executor.go` file (~312 lines)
  - Remove import statement for executor package from main.go
  - No remaining references to executor package in codebase
  - Build succeeds with no import errors

  **Testing Criteria:**
  **Unit Tests:**
  - No unit tests reference executor package
  
  **Integration Tests:**
  - Codebase grep: `grep -r "internal/executor" --include="*.go" .` returns no results
  - Build test: `go build` succeeds without import errors
  - Package verification: Ensure no orphaned references
  
  **Manual Testing:**
  - Directory structure verification
  - Import dependency validation

- [x] Story-004: As a developer, I want the executeComponent function removed from main.go so that there are no orphaned handler functions.

  **Acceptance Criteria:**
  - Remove `executeComponent()` function definition from main.go (lines 200-204)
  - Remove function comment about "npx-like functionality" (line 200)
  - Remove `handleRun` parameter from SetHandlers call in main.go (around lines 341-344)
  - Build succeeds with no unused function warnings

  **Testing Criteria:**
  **Unit Tests:**
  - No unit tests for executeComponent function
  
  **Integration Tests:**
  - Build test: `go build` succeeds without warnings
  - Function reference check: `grep -r "executeComponent" --include="*.go" .` returns no results
  - Handler chain validation
  
  **Manual Testing:**
  - Handler function verification
  - Build output inspection

- [x] Story-005: As a developer, I want handleRun handler removed from cmd/root.go so that the command handler infrastructure is clean.

  **Acceptance Criteria:**
  - Remove `handleRun` variable declaration from cmd/root.go (around line 777)
  - Remove `run` parameter from SetHandlers function signature (around line 805)
  - Remove `handleRun = run` assignment from SetHandlers body (line 831)
  - All calls to SetHandlers updated to remove run parameter
  - Build succeeds with no undefined variable errors

  **Testing Criteria:**
  **Unit Tests:**
  - No unit tests for handleRun handler
  
  **Integration Tests:**
  - Build test: `go build` succeeds without errors
  - Handler reference check: `grep -r "handleRun" --include="*.go" .` returns no results
  - SetHandlers signature validation
  
  **Manual Testing:**
  - Handler chain verification
  - Function parameter validation

- [x] Story-006: As a user, I want the main help text updated so that it accurately describes what Agent Smith does without mentioning execution functionality.

  **Acceptance Criteria:**
  - Remove "Execute components without installation (npx-like)" from feature list in cmd/root.go (line 17)
  - Help text focuses on: download, manage, link, and update components
  - Root command description accurately reflects tool purpose
  - `./agent-smith --help` output is clean and accurate

  **Testing Criteria:**
  **Unit Tests:**
  - No unit tests for help text content
  
  **Integration Tests:**
  - Help text verification: Parse help output and verify no execution mentions
  - Feature list validation: Ensure remaining features are accurate
  
  **Manual Testing:**
  - `./agent-smith --help` manual review
  - Feature description accuracy check
  - User experience validation

- [x] Story-007: As a developer, I want comprehensive validation that npx/run removal is complete so that no broken references remain in the codebase.

  **Acceptance Criteria:**
  - Run grep search for "npx" in all Go files: no results
  - Run grep search for "executeComponent" in all Go files: no results
  - Run grep search for "internal/executor" in all Go files: no results
  - Run grep search for "handleRun" in all Go files: no results
  - Build succeeds with zero warnings or errors
  - All existing commands (install, link, update, profiles) still function correctly

  **Testing Criteria:**
  **Unit Tests:**
  - Run existing unit test suite: all tests pass
  
  **Integration Tests:**
  - Grep validation: Automated search for removed references
  - Build validation: `go build` exits with code 0
  - Command smoke tests: Verify install/link/update/profiles commands
  - Import validation: No orphaned package imports
  
  **Manual Testing:**
  - Manual execution of core commands
  - Help text review for all command groups
  - End-to-end workflow testing

## Functional Requirements

- FR-1: The system must not include npx command in available commands
- FR-2: The system must not include run command in available commands
- FR-3: The system must not include internal/executor package in codebase
- FR-4: The system must not include executeComponent function in main.go
- FR-5: The system must not include handleRun handler in cmd/root.go
- FR-6: The system must not mention execution functionality in help text
- FR-7: The system must build successfully with zero errors or warnings
- FR-8: The system must maintain all existing core functionality (install, link, update, profiles)
- FR-9: The help output must accurately describe tool capabilities
- FR-10: The codebase must contain no references to removed functionality

## Non-Goals

- Updating PRD documentation files (leave as historical reference)
- Adding deprecation notices or migration guides
- Creating backward compatibility aliases
- Modifying any core functionality (install, link, update, profiles)
- Changing handler implementations beyond removing executeComponent
- Adding new features or commands
- Updating external documentation or README files

## Technical Details

### Files to Modify

| File | Action | Lines Affected | Description |
|------|--------|----------------|-------------|
| `internal/executor/executor.go` | DELETE | All (312) | Remove entire file |
| `internal/executor/` | DELETE | Directory | Remove package |
| `cmd/root.go` | MODIFY | 169-185, 17, ~777, ~805, ~831 | Remove commands & handler refs |
| `main.go` | MODIFY | 200-204, ~341-344 | Remove executeComponent & handler param |

### Validation Commands

```bash
# Verify no references remain
grep -r "npx" --include="*.go" .
grep -r "executeComponent" --include="*.go" .
grep -r "internal/executor" --include="*.go" .
grep -r "handleRun" --include="*.go" .

# Build and test
go build
./agent-smith --help
./agent-smith install --help
./agent-smith link --help
```

### Impact Assessment

**Removed:**
- ~350+ lines of code (executor package + command definitions)
- 2 CLI commands (npx, run)
- 1 handler function (executeComponent)
- 1 handler variable (handleRun)
- 1 complete package (internal/executor)

**Unchanged:**
- Core functionality (install, link, update, profiles)
- Component detection and downloading
- Profile management
- Symlink creation
- Update detection
- All existing tests

### Success Criteria

- Build succeeds: `go build` exits with code 0
- No broken references: All grep searches return zero results
- Help text accurate: No mention of execution/npx/run functionality
- Core commands work: install, link, update, profiles all functional
- Clean codebase: No orphaned imports or unused functions

## Risk Assessment

**Low Risk**: Users affected by command removal
- **Mitigation**: Commands were non-functional for documentation-based components
- **Impact**: Users get "command not found" error, can remove from scripts

**Low Risk**: Future desire to execute scripts
- **Mitigation**: Can be re-added later with proper design
- **Impact**: Clean slate for future implementation if needed

**No Risk**: Core functionality disruption
- **Mitigation**: Only removing execution-related code, core features untouched
- **Impact**: Zero impact on working features
