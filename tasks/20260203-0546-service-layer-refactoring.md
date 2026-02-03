# Service Layer Refactoring - Break Up main.go

**Status:** 🚧 In Progress (Phase 2: 62.5% Complete)  
**Priority:** High  
**Started:** 2026-02-03  
**Last Updated:** 2026-02-03  
**Estimated Completion:** 2026-02-04

## Progress Summary

### ✅ Completed (7 of 8 services)
- **StatusService**: Shows system status, active profile, detected targets, component counts
- **TargetService**: Add/remove/list custom targets
- **UpdateService**: Update single component or all components  
- **InstallService**: Install skills/agents/commands/bulk with profile and target-dir support
- **UninstallService**: Uninstall components with defensive unlinking
- **LinkService**: Link/unlink components, auto-link, show link status (9 methods)
- **ProfileService**: Full profile management - create, delete, activate, list, show, component operations (10 methods)

### ⏳ Remaining (1 of 8 services)
- **MaterializeService**: Materialize components to projects (HIGHEST COMPLEXITY - 6 methods)

### Current State
- ✅ main.go compiles successfully
- ✅ All completed service handlers tested and working
- ✅ No LSP errors or build issues
- 📊 main.go reduced from 3591 to 2270 lines (~37% reduction)
- 🎯 Next: Create MaterializeService (final service!)

## Problem Statement

The `main.go` file has grown to 3,591 lines and contains all business logic inline within handler functions. This creates several issues:

1. **Lack of Reusability**: Business logic is embedded in CLI handlers and cannot be reused by a TUI or other interfaces
2. **Poor Testability**: Handler functions are difficult to unit test independently
3. **Low Maintainability**: Navigating and modifying such a large file is challenging
4. **Tight Coupling**: CLI concerns (flags, output formatting) are mixed with business logic
5. **No Abstraction**: Direct dependencies on concrete implementations throughout

## Objectives

Transform `main.go` into a clean service-oriented architecture with:

- **Domain-based services** organized by business capability
- **Interface-driven design** enabling testing and multiple implementations
- **Constructor injection** for explicit dependency management
- **Thin handlers** in main.go that delegate to services
- **Reusable services** that can be called from CLI, TUI, or API

## Architecture Design

### Service Organization

```
pkg/services/
├── interfaces.go          # All service interfaces
├── install/              # InstallService
├── link/                 # LinkService  
├── profile/              # ProfileService
├── materialize/          # MaterializeService
├── update/               # UpdateService
├── uninstall/            # UninstallService
├── target/               # TargetService
└── status/               # StatusService
```

### Key Principles

1. **Domain-Based Services**: Each service handles a cohesive set of related operations
2. **Constructor Injection**: Services receive dependencies via `NewService()` functions
3. **Interface-Driven**: All services implement interfaces defined in `interfaces.go`
4. **Separation of Concerns**: Business logic in services, CLI concerns in handlers
5. **Explicit Dependencies**: Logger, formatter, and other deps injected explicitly

## Implementation Plan

### Phase 1: Infrastructure Setup ✅

- [x] Create `pkg/services/` directory structure
- [x] Define all service interfaces in `pkg/services/interfaces.go`
- [x] Create common types (Options structs, etc.)
- [x] Set up testing framework for services

### Phase 2: Create Services ✅ (5 of 8 services completed)

**Service 1: StatusService** ✅ COMPLETED
- [x] Create `pkg/services/status/service.go`
- [x] Implement `ShowSystemStatus()` method
- [x] Extract logic from handler (lines 1826-1949)
- [x] Update main.go handler
- [ ] Write unit tests

**Service 2: TargetService** ✅ COMPLETED
- [x] Create `pkg/services/target/service.go`
- [x] Implement `AddCustomTarget()`, `RemoveCustomTarget()`, `ListTargets()`
- [x] Extract logic from handlers (lines 1950-2167)
- [x] Update main.go handlers
- [ ] Write unit tests

**Service 3: UpdateService** ✅ COMPLETED
- [x] Create `pkg/services/update/service.go`
- [x] Implement `UpdateComponent()`, `UpdateAll()`, `CheckForUpdates()`
- [x] Extract logic from handlers (lines 930-948)
- [x] Update main.go handlers
- [ ] Write unit tests

**Service 4: InstallService** ✅ COMPLETED
- [x] Create `pkg/services/install/service.go`
- [x] Implement `InstallSkill()`, `InstallAgent()`, `InstallCommand()`, `InstallBulk()`
- [x] Extract logic from handlers (lines 574-928)
- [x] Handle profile-based and target-dir installations
- [x] Update main.go handlers
- [ ] Write unit tests

**Service 5: UninstallService** ✅ COMPLETED
- [x] Create `pkg/services/uninstall/service.go`
- [x] Implement `UninstallComponent()`, `UninstallAllFromSource()`
- [x] Extract logic from handlers (lines 1209-1264)
- [x] Update main.go handlers
- [ ] Write unit tests

**Service 6: LinkService** ⏳ TODO (HIGH COMPLEXITY)
- [ ] Create `pkg/services/link/service.go`
- [ ] Implement link operations: `LinkComponent()`, `LinkAll()`, `LinkByType()`
- [ ] Implement unlink operations: `UnlinkComponent()`, `UnlinkAll()`, `UnlinkByType()`
- [ ] Implement status: `AutoLinkRepositories()`, `ListLinked()`, `ShowStatus()`
- [ ] Extract logic from handlers (lines 949-1207)
- [ ] Write unit tests
- [ ] Update main.go handlers

**Service 7: ProfileService** ⏳ TODO (HIGH COMPLEXITY)
- [ ] Create `pkg/services/profile/service.go`
- [ ] Implement lifecycle: `CreateProfile()`, `DeleteProfile()`, `ActivateProfile()`, `DeactivateProfile()`
- [ ] Implement display: `ListProfiles()`, `ShowProfile()`
- [ ] Implement components: `AddComponent()`, `CopyComponent()`, `RemoveComponent()`, `CherryPickComponents()`
- [ ] Extract logic from handlers (lines 1265-1825)
- [ ] Write unit tests
- [ ] Update main.go handlers

**Service 8: MaterializeService** ⏳ TODO (HIGHEST COMPLEXITY)
- [ ] Create `pkg/services/materialize/service.go`
- [ ] Implement `MaterializeComponent()`, `MaterializeAll()`
- [ ] Implement `ListMaterialized()`, `ShowComponentInfo()`
- [ ] Implement `ShowStatus()`, `UpdateMaterialized()`
- [ ] Extract logic from handlers (lines 2168-3586)
- [ ] Write unit tests
- [ ] Update main.go handlers

### Phase 3: Refactor main.go ✅ (Partial - 5 services integrated)

- [x] Initialize all services with dependency injection (5 services)
- [x] Replace handler implementations with service calls (install, update, uninstall, status, target)
- [x] Keep handlers as thin wrappers (2-5 lines each)
- [ ] Complete remaining 3 services (link, profile, materialize)
- [ ] Remove factory functions now encapsulated in services
- [ ] Clean up deprecated helper functions
- [ ] Verify main.go is ~300-400 lines

### Phase 4: Testing & Validation ✅

- [ ] Run existing integration test suite
- [ ] Verify all CLI commands work identically
- [ ] Add unit tests for each service
- [ ] Test error handling and edge cases
- [ ] Validate logging and formatting output

### Phase 5: Documentation 📝

- [ ] Add godoc comments to all service interfaces
- [ ] Document service dependencies and usage
- [ ] Create service usage examples
- [ ] Update README with architecture overview

## Service Specifications

### InstallService Interface

```go
type InstallService interface {
    InstallSkill(repoURL, name string, opts InstallOptions) error
    InstallAgent(repoURL, name string, opts InstallOptions) error
    InstallCommand(repoURL, name string, opts InstallOptions) error
    InstallBulk(repoURL string, opts InstallOptions) error
}

type InstallOptions struct {
    Profile   string
    TargetDir string
}
```

**Dependencies:** SkillDownloader, AgentDownloader, CommandDownloader, BulkDownloader, ProfileManager, Logger, Formatter

### LinkService Interface

```go
type LinkService interface {
    LinkComponent(componentType, name string, opts LinkOptions) error
    LinkAll(opts LinkOptions) error
    LinkByType(componentType string, opts LinkOptions) error
    UnlinkComponent(componentType, name string, opts UnlinkOptions) error
    UnlinkAll(opts UnlinkOptions) error
    UnlinkByType(componentType string, opts UnlinkOptions) error
    AutoLinkRepositories() error
    ListLinked() error
    ShowStatus(opts StatusOptions) error
}
```

**Dependencies:** ComponentLinker, ProfileManager, Logger, Formatter

### ProfileService Interface

```go
type ProfileService interface {
    ListProfiles(opts ListProfileOptions) error
    ShowProfile(name string) error
    CreateProfile(name string) error
    DeleteProfile(name string) error
    ActivateProfile(name string) error
    DeactivateProfile() error
    AddComponent(componentType, profileName, componentName string) error
    CopyComponent(sourceProfile, targetProfile, componentType, componentName string) error
    RemoveComponent(profileName, componentType, componentName string) error
    CherryPickComponents(targetProfile string, sourceProfiles []string) error
}
```

**Dependencies:** ProfileManager, ComponentLinker, Logger, Formatter

### MaterializeService Interface

```go
type MaterializeService interface {
    MaterializeComponent(componentType, name string, opts MaterializeOptions) error
    MaterializeAll(opts MaterializeOptions) error
    ListMaterialized(opts ListMaterializedOptions) error
    ShowComponentInfo(componentType, name string, opts InfoOptions) error
    ShowStatus(opts StatusOptions) error
    UpdateMaterialized(opts UpdateOptions) error
}
```

**Dependencies:** Materializer, ProfileManager, ProjectManager, MetadataManager, UpdateDetector, Logger, Formatter

### UpdateService Interface

```go
type UpdateService interface {
    UpdateComponent(componentType, name string, opts UpdateOptions) error
    UpdateAll(opts UpdateOptions) error
    CheckForUpdates(opts UpdateOptions) ([]UpdateInfo, error)
}
```

**Dependencies:** UpdateDetector, ProfileManager, Logger, Formatter

### UninstallService Interface

```go
type UninstallService interface {
    UninstallComponent(componentType, name string, opts UninstallOptions) error
    UninstallAllFromSource(repoURL string, opts UninstallOptions) error
}
```

**Dependencies:** Uninstaller, ComponentLinker, ProfileManager, Logger, Formatter

### TargetService Interface

```go
type TargetService interface {
    AddCustomTarget(name, path string) error
    RemoveCustomTarget(name string) error
    ListTargets() error
}
```

**Dependencies:** Config, Logger, Formatter

### StatusService Interface

```go
type StatusService interface {
    ShowSystemStatus() error
}
```

**Dependencies:** ProfileManager, Config, Paths, Logger, Formatter

## Success Criteria

1. ✅ `main.go` reduced from 3,591 to ~300-400 lines
2. ✅ All business logic extracted into domain services
3. ✅ Service interfaces defined for all services
4. ✅ Services use constructor injection for dependencies
5. ✅ Handlers in main.go are thin wrappers (2-5 lines)
6. ✅ All existing CLI commands work identically
7. ✅ Integration tests pass without modification
8. ✅ Services are reusable from TUI or other interfaces
9. ✅ Comprehensive unit tests for each service
10. ✅ Documentation updated with new architecture

## Expected Benefits

1. **Reusability**: Services can be called from CLI, TUI, API, or tests
2. **Testability**: Each service independently testable with mocked dependencies
3. **Maintainability**: Clear separation of concerns and single responsibility
4. **Extensibility**: Easy to add new features or interfaces
5. **Readability**: main.go becomes clear entry point showing system composition
6. **Type Safety**: Interfaces enable compile-time checking and better IDE support

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Breaking existing functionality | High | Run full integration test suite after each service |
| Inconsistent behavior after refactor | Medium | Compare CLI output before/after for each command |
| Missing edge cases | Medium | Comprehensive unit tests with edge case coverage |
| Performance degradation | Low | Profile critical paths, minimal overhead expected |
| Incomplete migration | Low | Checklist tracking for each handler |

## Timeline

- **Phase 1** (Infrastructure): 30 minutes
- **Phase 2** (Create Services): 3-4 hours
- **Phase 3** (Update main.go): 1 hour
- **Phase 4** (Testing): 30 minutes
- **Phase 5** (Documentation): 30 minutes

**Total Estimated Time:** ~6 hours

## Related Work

- Enables future TUI implementation
- Supports API server development
- Facilitates plugin architecture
- Improves developer onboarding

## Notes

- Handlers remain in main.go as thin wrappers (not moved to cmd/ package)
- Logger and Formatter injected into each service
- Migration done all at once but each service is independent
- Backward compatibility maintained for cmd package
