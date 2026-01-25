# PRD: Refactor main.go into Standard Go Package Structure

## Introduction

Refactor the monolithic `main.go` file (~4200 lines) into a well-organized, standard Go package structure that separates concerns, improves testability, and follows Go best practices.

## Goals

- Reduce main.go from 4200+ lines to < 200 lines
- Separate concerns by domain/responsibility into focused packages
- Improve testability with dependency injection
- Maintain zero breaking changes to CLI interface
- Enable easier future extensibility (plugins, new component types)

## User Stories

- [ ] **Story-001**: As a developer maintaining Agent Smith, I want all data models in a dedicated package so that I can understand types without searching through business logic. - Create `internal/models/` package, Extract ComponentType, DetectedComponent, ComponentDetectionPattern, DetectionConfig, ComponentLockEntry, ComponentLockFile, ComponentMetadata, ComponentFrontmatter types, Update all imports to use models package.

- [ ] **Story-002**: As a developer adding new features, I want file system utilities in a separate package so that I can reuse them without coupling to main business logic. - Create `internal/fileutil/` package, Extract cross-platform permission functions, Extract file/directory copying utilities, Extract frontmatter parsing functions, Add package-level documentation.

- [ ] **Story-003**: As a developer working on path management, I want centralized path constants so that directory locations are consistent across the codebase. - Create `pkg/paths/` package, Extract directory path constants (AgentsDir, SkillsSubDir, etc.), Create path helper functions (GetAgentsDir, GetSkillsDir, etc.), Replace hardcoded paths throughout codebase.

- [ ] **Story-004**: As a developer maintaining detection logic, I want component detection isolated in its own package so that I can test and extend detection patterns independently. - Create `internal/detector/` package, Extract RepositoryDetector struct and methods, Separate pattern matching into patterns.go, Separate component detection into components.go, Add interfaces for testability.

- [ ] **Story-005**: As a developer fixing download bugs, I want downloader implementations separated by type so that I can modify skill downloads without affecting agent downloads. - Create `internal/downloader/` package, Extract SkillDownloader to skill.go, Extract AgentDownloader to agent.go, Extract CommandDownloader to command.go, Extract BulkDownloader to bulk.go, Create common.go for shared utilities.

- [ ] **Story-006**: As a developer maintaining git operations, I want git functionality isolated so that I can mock it in tests. - Create `internal/git/` package, Extract clone operations to clone.go, Extract repository info to info.go, Create clean interfaces for testing.

- [ ] **Story-007**: As a developer working on metadata, I want lock file and metadata operations in a dedicated package so that I can ensure consistency. - Create `internal/metadata/` package, Extract lock file operations to lock.go, Extract legacy metadata handling to legacy.go, Extract hash computation to hash.go.

- [ ] **Story-008**: As a developer maintaining linking functionality, I want component linking isolated so that I can extend it without touching other code. - Create `internal/linker/` package, Extract ComponentLinker to linker.go, Extract link status analysis to status.go, Implement dependency injection.

- [ ] **Story-009**: As a developer maintaining update functionality, I want update detection in its own package so that I can improve it independently. - Create `internal/updater/` package, Extract UpdateDetector to updater.go, Implement clean interfaces.

- [ ] **Story-010**: As a developer maintaining npx-like execution, I want executor functionality isolated so that I can extend execution without affecting downloads. - Create `internal/executor/` package, Extract ComponentExecutor to executor.go, Wire up to CLI handlers.

- [ ] **Story-011**: As a developer reviewing code, I want main.go to be minimal so that I can quickly understand the application entry point. - Refactor main.go to contain only handler implementations, Move all business logic to appropriate packages, Ensure main.go is < 200 lines, Add clear comments for each handler.

- [ ] **Story-012**: As a developer writing tests, I want dependency injection throughout the codebase so that I can mock expensive operations. - Replace direct instantiation with constructor injection, Define interfaces for major components, Enable easy mocking of git, file system, and network operations.

## Functional Requirements

- **FR-1**: All existing CLI commands must work exactly as before
- **FR-2**: All existing tests must continue to pass
- **FR-3**: No changes to lock file or metadata file formats
- **FR-4**: Existing user installations must continue to work without migration
- **FR-5**: Package structure must follow Go standard layout patterns
- **FR-6**: All packages in `internal/` to prevent external dependencies
- **FR-7**: Each package must have a single, clear responsibility
- **FR-8**: Packages must use dependency injection for testability
- **FR-9**: No performance degradation from refactoring
- **FR-10**: Clear package-level documentation for all new packages

## Non-Goals

- No changes to CLI command syntax or behavior
- No changes to user-facing error messages
- No new features during refactoring
- No changes to external dependencies
- No database or storage backend changes
- No changes to git repository interactions

## Technical Solution

### Proposed Package Structure

```
agent-smith/
├── cmd/
│   └── root.go                    # Cobra CLI root (already exists)
├── main.go                         # Minimal main - calls cmd.Execute()
│
├── internal/
│   ├── models/                    # Data structures and domain models
│   │   ├── component.go           # Component types & DetectedComponent
│   │   ├── config.go              # Detection configuration
│   │   ├── lock.go                # Lock file structures
│   │   └── metadata.go            # Metadata structures
│   │
│   ├── detector/                  # Component detection logic
│   │   ├── detector.go            # RepositoryDetector
│   │   ├── patterns.go            # Detection patterns & configuration
│   │   └── components.go          # Component detection logic
│   │
│   ├── downloader/                # Component downloaders
│   │   ├── skill.go               # SkillDownloader
│   │   ├── agent.go               # AgentDownloader
│   │   ├── command.go             # CommandDownloader
│   │   ├── bulk.go                # BulkDownloader
│   │   └── common.go              # Shared utilities
│   │
│   ├── linker/                    # OpenCode linking
│   │   ├── linker.go              # ComponentLinker
│   │   └── status.go              # Link status & analysis
│   │
│   ├── updater/                   # Update detection
│   │   └── updater.go             # UpdateDetector
│   │
│   ├── executor/                  # npx-like execution
│   │   └── executor.go            # ComponentExecutor
│   │
│   ├── metadata/                  # Metadata & lock files
│   │   ├── lock.go                # Lock file operations
│   │   ├── legacy.go              # Legacy metadata
│   │   └── hash.go                # Hash computation
│   │
│   ├── git/                       # Git operations
│   │   ├── clone.go               # Cloning operations
│   │   └── info.go                # Repository info
│   │
│   └── fileutil/                  # File system utilities
│       ├── permissions.go         # Cross-platform permissions
│       ├── copy.go                # File/directory copying
│       └── frontmatter.go         # YAML frontmatter parsing
│
└── pkg/
    └── paths/                     # Path constants & helpers
        └── paths.go
```

### Migration Phases

#### Phase 1: Extract Models (Low Risk, 2-3 hours)
```
Tasks:
1. Create internal/models/ directory
2. Extract component.go (ComponentType, DetectedComponent)
3. Extract config.go (ComponentDetectionPattern, DetectionConfig)
4. Extract lock.go (ComponentLockEntry, ComponentLockFile)
5. Extract metadata.go (ComponentMetadata, ComponentFrontmatter)
6. Update imports in main.go
7. Run all tests

Validation: All existing tests pass
```

#### Phase 2: Extract Utilities (Low Risk, 2-3 hours)
```
Tasks:
1. Create internal/fileutil/ directory
2. Extract permissions.go (getCrossPlatformPermissions, etc.)
3. Extract copy.go (copyFile, copyDirectory, symlink functions)
4. Extract frontmatter.go (parseFrontmatter, determineComponentName)
5. Create pkg/paths/ directory
6. Extract path constants and helper functions
7. Update all imports
8. Run all tests

Validation: All existing tests pass
```

#### Phase 3: Extract Detector (Medium Risk, 4-5 hours)
```
Tasks:
1. Create internal/detector/ directory
2. Extract detector.go (RepositoryDetector struct, core methods)
3. Extract patterns.go (pattern matching, config operations)
4. Extract components.go (detectComponentsInRepo, pattern matching)
5. Update all dependencies to use detector package
6. Add dependency injection
7. Run detection tests

Validation: All detection tests pass
```

#### Phase 4: Extract Downloaders (Medium Risk, 6-8 hours)
```
Tasks:
1. Create internal/downloader/ directory
2. Extract common.go (shared utilities)
3. Extract skill.go (SkillDownloader)
4. Extract agent.go (AgentDownloader)
5. Extract command.go (CommandDownloader)
6. Extract bulk.go (BulkDownloader)
7. Inject dependencies (detector, metadata, git)
8. Run download tests

Validation: End-to-end download tests pass
```

#### Phase 5: Extract Metadata & Git (Low Risk, 3-4 hours)
```
Tasks:
1. Create internal/metadata/ directory
2. Extract lock.go (saveLockFile, loadFromLockFile)
3. Extract legacy.go (saveMetadata, loadMetadata)
4. Extract hash.go (computeGitHubTreeSHA, computeLocalFolderHash)
5. Create internal/git/ directory
6. Extract clone.go (git cloning operations)
7. Extract info.go (repository information)
8. Run tests

Validation: Metadata and git operations work correctly
```

#### Phase 6: Extract Linker & Updater (Medium Risk, 4-5 hours)
```
Tasks:
1. Create internal/linker/ directory
2. Extract linker.go (ComponentLinker)
3. Extract status.go (LinkStatus, link analysis)
4. Create internal/updater/ directory
5. Extract updater.go (UpdateDetector)
6. Inject dependencies
7. Run link and update tests

Validation: Link and update operations work
```

#### Phase 7: Extract Executor (Low Risk, 2-3 hours)
```
Tasks:
1. Create internal/executor/ directory
2. Extract executor.go (ComponentExecutor)
3. Wire up to CLI handlers
4. Run execution tests

Validation: npx-like execution works
```

#### Phase 8: Clean Up main.go (Final, 3-4 hours)
```
Tasks:
1. Refactor main.go to minimal size (< 200 lines)
2. Keep only handler implementations
3. Update cmd/root.go to use new package structure
4. Add comprehensive package documentation
5. Run full integration test suite
6. Update developer documentation

Validation: All tests pass, CLI works identically
```

### Dependency Injection Pattern

**Before (tightly coupled):**
```go
type SkillDownloader struct {
    baseDir  string
    detector *RepositoryDetector
}

func NewSkillDownloader() *SkillDownloader {
    return &SkillDownloader{
        baseDir:  filepath.Join(home, ".agents", "skills"),
        detector: NewRepositoryDetector(),
    }
}
```

**After (dependency injection):**
```go
type SkillDownloader struct {
    baseDir  string
    detector detector.Interface
    metadata metadata.Manager
    git      git.Client
}

func New(baseDir string, det detector.Interface, meta metadata.Manager, git git.Client) *SkillDownloader {
    return &SkillDownloader{
        baseDir:  baseDir,
        detector: det,
        metadata: meta,
        git:      git,
    }
}
```

## Success Criteria

### Code Structure
- [ ] main.go is < 200 lines
- [ ] All packages follow single responsibility principle
- [ ] Each package has clear, documented interfaces
- [ ] All business logic moved out of main.go
- [ ] Dependency injection used throughout

### Testing
- [ ] All existing tests pass without modification
- [ ] Test coverage maintained or improved (target: 80%+)
- [ ] Each package has unit tests
- [ ] Integration tests validate package interactions

### Functionality
- [ ] All CLI commands work identically to before
- [ ] No breaking changes to user experience
- [ ] Lock files and metadata remain compatible
- [ ] Existing installations work without migration
- [ ] No performance degradation (benchmark verification)

### Documentation
- [ ] Each package has package-level documentation
- [ ] Exported functions have clear godoc comments
- [ ] README updated with new package structure
- [ ] Developer guide includes architecture overview

### Quality Gates
- [ ] Code passes all linter checks
- [ ] No new warnings or errors
- [ ] All imports properly organized
- [ ] No circular dependencies
- [ ] Clean separation of concerns verified

## Acceptance Criteria

### AC-1: Models Package
```
GIVEN the internal/models/ package is created
WHEN all type definitions are extracted
THEN all imports use the models package
AND all existing tests pass
AND no circular dependencies exist
```

### AC-2: Utilities Package
```
GIVEN the internal/fileutil/ package is created
WHEN file utilities are extracted
THEN all file operations use the fileutil package
AND cross-platform compatibility is maintained
AND all existing tests pass
```

### AC-3: Paths Package
```
GIVEN the pkg/paths/ package is created
WHEN path constants are extracted
THEN all hardcoded paths are replaced
AND path helpers work on all platforms
AND no paths are hardcoded in main.go
```

### AC-4: Detector Package
```
GIVEN the internal/detector/ package is created
WHEN detection logic is extracted
THEN component detection works identically
AND detection patterns are configurable
AND detection can be easily tested with mocks
AND all detection tests pass
```

### AC-5: Downloader Package
```
GIVEN the internal/downloader/ package is created
WHEN downloaders are extracted
THEN skill/agent/command downloads work identically
AND dependencies are injected
AND downloaders can be tested independently
AND all download tests pass
```

### AC-6: Git Package
```
GIVEN the internal/git/ package is created
WHEN git operations are extracted
THEN git cloning works identically
AND git operations can be mocked for testing
AND repository operations are isolated
```

### AC-7: Metadata Package
```
GIVEN the internal/metadata/ package is created
WHEN metadata operations are extracted
THEN lock files work identically
AND legacy metadata is supported
AND hash computation works correctly
AND all metadata tests pass
```

### AC-8: Linker Package
```
GIVEN the internal/linker/ package is created
WHEN linking logic is extracted
THEN component linking works identically
AND link status analysis works correctly
AND all link tests pass
```

### AC-9: Updater Package
```
GIVEN the internal/updater/ package is created
WHEN update logic is extracted
THEN update detection works identically
AND update operations work correctly
AND all update tests pass
```

### AC-10: Executor Package
```
GIVEN the internal/executor/ package is created
WHEN executor logic is extracted
THEN npx-like execution works identically
AND executor can find local components
AND executor can run from repository
AND all execution tests pass
```

### AC-11: Minimal main.go
```
GIVEN all business logic is extracted
WHEN main.go is refactored
THEN main.go is < 200 lines
AND main.go contains only handler implementations
AND main.go is easy to understand
AND all CLI commands work identically
```

### AC-12: Integration Tests
```
GIVEN all packages are extracted
WHEN integration tests are run
THEN all existing tests pass
AND CLI functionality is identical
AND no performance degradation occurs
AND all error messages are unchanged
```

## Testing Strategy

### Unit Tests (per package)
- **models/**: Test type conversions and validations
- **fileutil/**: Test file operations with temp directories
- **paths/**: Test path generation across platforms
- **detector/**: Test detection patterns with fixtures
- **downloader/**: Test with mocked git and file operations
- **git/**: Test with local repositories
- **metadata/**: Test lock file read/write operations
- **linker/**: Test symlink creation and analysis
- **updater/**: Test update detection with fixtures
- **executor/**: Test execution with mocked components

### Integration Tests
- Test end-to-end skill download
- Test end-to-end agent download
- Test bulk download with multi-component repos
- Test linking workflow
- Test update workflow
- Test npx-like execution

### Coverage Goals
- **detector/**: 85%+ coverage
- **downloader/**: 80%+ coverage
- **linker/**: 80%+ coverage
- **fileutil/**: 90%+ coverage
- **Overall**: 80%+ coverage maintained

## Risks and Mitigation

### Risk 1: Migration Complexity
**Risk**: Large refactor could introduce bugs
**Likelihood**: Medium
**Impact**: High
**Mitigation**: 
- Incremental phases with testing after each
- Keep main.go working in parallel during phases 1-7
- Each phase in separate branch
- Comprehensive test suite before starting

### Risk 2: Breaking Internal APIs
**Risk**: Third-party code might depend on internals
**Likelihood**: Low
**Impact**: Medium
**Mitigation**:
- Everything in `internal/` is not public API by Go convention
- Document that internal packages are unstable
- Provide migration guide if needed

### Risk 3: Performance Regression
**Risk**: Additional abstraction layers could slow performance
**Likelihood**: Low
**Impact**: Medium
**Mitigation**:
- Benchmark critical paths before/after
- Profile if performance issues detected
- Optimize hot paths if needed

### Risk 4: Circular Dependencies
**Risk**: Poor package design could create import cycles
**Likelihood**: Medium
**Impact**: High
**Mitigation**:
- Models package has zero dependencies
- Use interfaces to break cycles
- Review dependency graph after each phase

## Rollback Plan

If critical issues arise during refactoring:

1. **Each phase in separate branch**: Can revert to previous phase
2. **Parallel working main.go**: Keep original working during phases 1-7
3. **Tagged releases**: Tag stable points for easy rollback
4. **Validation gates**: Don't proceed to next phase until tests pass
5. **Final validation**: Only delete old code in phase 8 after full validation

## Timeline and Resources

### Total Effort: 26-35 hours

**Phase 1 (Models)**: 2-3 hours  
**Phase 2 (Utilities)**: 2-3 hours  
**Phase 3 (Detector)**: 4-5 hours  
**Phase 4 (Downloaders)**: 6-8 hours  
**Phase 5 (Metadata & Git)**: 3-4 hours  
**Phase 6 (Linker & Updater)**: 4-5 hours  
**Phase 7 (Executor)**: 2-3 hours  
**Phase 8 (Cleanup)**: 3-4 hours  

### Resources Required
- 1 Senior Go Developer (lead refactoring)
- Test environment for validation
- CI/CD pipeline for automated testing

## Future Enhancements

Once refactoring is complete, these become easier:

1. **Plugin System**: Clean interfaces enable third-party plugins
2. **Multiple Storage Backends**: Abstract storage layer
3. **Remote Execution**: Execute components on remote hosts
4. **Caching Layer**: Intelligent download caching
5. **Parallel Downloads**: Concurrent component downloads
6. **Structured Error Handling**: Rich error context
7. **Telemetry**: Usage analytics (opt-in)
8. **Enhanced Configuration**: Better config file support

## References

- [Go Package Layout](https://github.com/golang-standards/project-layout)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [SOLID Principles in Go](https://dave.cheney.net/2016/08/20/solid-go-design)

---

**Document Version**: 1.0  
**Created**: 2026-01-25  
**Status**: Draft - Ready for Implementation
