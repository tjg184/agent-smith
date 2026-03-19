# God Object Decomposition Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Decompose three 1200–2000-line god objects into focused sub-packages and fix four quick-win issues that block testability.

**Architecture:** Each god object is split into sub-packages under its existing package root. Sub-packages expose free functions that accept the parent struct as a parameter; the parent struct and its public API are unchanged for callers. Quick wins (error returns, lock-file deduplication, component-type registry) are committed first to minimize merge conflicts.

**Tech Stack:** Go 1.23, standard library only for new files.

---

## Phase 1 — Quick Wins

### Task 1: Expose `LoadLockFile` from `internal/metadata`

The materialize service has its own private `loadLockFile` re-implementation. Before splitting the service we expose the real one from `internal/metadata` so the service can delegate.

**Files:**
- Modify: `internal/metadata/lock.go`

- [ ] Add an exported `LoadLockFile(baseDir string) (models.ComponentLockFile, error)` function that wraps `loadOrCreateLockFile`. Place it just above the private function (around line 407):

```go
// LoadLockFile reads (or creates) the component lock file for the given base directory.
func LoadLockFile(baseDir string) (models.ComponentLockFile, error) {
    lockFilePath := filepath.Join(baseDir, ".component-lock.json")
    return loadOrCreateLockFile(lockFilePath)
}
```

- [ ] Build: `go build ./...`
- [ ] Test: `go test ./internal/metadata/...`
- [ ] Commit:
```
git add internal/metadata/lock.go
git commit -m "feat(metadata): expose LoadLockFile for external callers"
```

---

### Task 2: Replace private `loadLockFile` in materialize service

**Files:**
- Modify: `pkg/services/materialize/service.go`

- [ ] Delete the `loadLockFile` method (lines 127–145) from `service.go`.
- [ ] Update `buildFilesystemNameMap` (line 100) to call `metadataPkg.LoadLockFile(baseDir)` instead of `s.loadLockFile(baseDir)`.
- [ ] Update `MaterializeAll` (line 451) to call `metadataPkg.LoadLockFile(baseDir)` instead of `s.loadLockFile(baseDir)`.
- [ ] Build: `go build ./...`
- [ ] Test: `go test ./pkg/services/materialize/...`
- [ ] Commit:
```
git add pkg/services/materialize/service.go
git commit -m "refactor(materialize): replace private loadLockFile with metadata.LoadLockFile"
```

---

### Task 3: Fix `log.Fatal` in `internal/downloader` constructors

`ForTypeWithProfile`, `ForTypeWithTargetDir`, and `baseDirForType` call `log.Fatal` on errors, killing the test process.

**Files:**
- Modify: `internal/downloader/component.go`
- Modify: callers in `internal/container/app.go`, `internal/updater/updater.go`

- [ ] Change `ForTypeWithProfile` signature to `func ForTypeWithProfile(ct models.ComponentType, profile string) (Downloader, error)`. Replace `log.Fatal` calls with `return nil, fmt.Errorf(...)`.
- [ ] Change `ForTypeWithTargetDir` signature to `func ForTypeWithTargetDir(ct models.ComponentType, targetDir string) (Downloader, error)`. Replace `log.Fatal` with `return nil, fmt.Errorf(...)`.
- [ ] Change `baseDirForType` signature to `func baseDirForType(ct models.ComponentType) (string, error)`. Replace all `log.Fatal`/`log.Fatalf` calls with `return "", fmt.Errorf(...)`. Update `ForType` to propagate the error (change its signature to `func ForType(ct models.ComponentType) (Downloader, error)`).
- [ ] Fix all call sites that used the old signatures:
  - `internal/updater/updater.go`: `downloader.ForTypeWithProfile` and `downloader.ForType` — add error handling, return the error from `UpdateComponent` / `downloadComponentWithRepo`.
  - `internal/container/app.go`: search for any direct calls (there should be none after the updater refactor, but verify).
- [ ] Build: `go build ./...`
- [ ] Test: `go test ./internal/downloader/... ./internal/updater/...`
- [ ] Commit:
```
git add internal/downloader/component.go internal/updater/updater.go
git commit -m "refactor(downloader): replace log.Fatal with error returns in constructors"
```

---

### Task 4: Fix `panic` in `internal/updater` constructors

`NewUpdateDetector` and `NewUpdateDetectorWithProfile` panic on path failures and construct their own `ProfileManager` / `LockService`, bypassing the DI container.

**Files:**
- Modify: `internal/updater/updater.go`
- Modify: `internal/container/app.go` (caller)

- [ ] Change `NewUpdateDetector()` to `NewUpdateDetector() (*UpdateDetector, error)`. Replace all `panic(...)` with `return nil, fmt.Errorf(...)`.
- [ ] Change `NewUpdateDetectorWithProfile(profile string)` to `NewUpdateDetectorWithProfile(profile string) (*UpdateDetector, error)`. Replace all `panic(...)` with `return nil, fmt.Errorf(...)`.
- [ ] `NewUpdateDetectorWithBaseDir` already returns `*UpdateDetector` with no error — leave it.
- [ ] In `internal/container/app.go`, find where `updater.NewUpdateDetector()` and `updater.NewUpdateDetectorWithProfile()` are called and add `if err != nil { return err }` handling.
- [ ] Build: `go build ./...`
- [ ] Test: `go test ./internal/updater/...`
- [ ] Commit:
```
git add internal/updater/updater.go internal/container/app.go
git commit -m "refactor(updater): replace panic with error returns in constructors"
```

---

### Task 5: Add `PluralDir()` to `models.ComponentType` and consolidate callers

The plural directory names (`"skills"`, `"agents"`, `"commands"`) are hardcoded in at least 10 files. `componentMetaTable` in `downloader` is the only authoritative mapping but is private.

**Files:**
- Modify: `internal/models/models.go`
- Modify: `internal/updater/updater.go` (remove `pluralToComponentType`, use `models.ComponentTypeFromPlural`)
- Modify: `internal/uninstaller/uninstaller.go` (hardcoded plural strings)
- Note: Do NOT touch `cmd/*.go` plural literals yet — those are Cobra `Use:` strings for the CLI UX.

- [ ] Add to `internal/models/models.go`:

```go
// PluralDir returns the directory name used for this component type (e.g. "skills", "agents", "commands").
func (ct ComponentType) PluralDir() string {
    switch ct {
    case ComponentSkill:
        return "skills"
    case ComponentAgent:
        return "agents"
    case ComponentCommand:
        return "commands"
    default:
        return string(ct) + "s"
    }
}

// ComponentTypeFromPlural converts a plural directory name to a ComponentType.
// Returns an error for unknown strings.
func ComponentTypeFromPlural(plural string) (ComponentType, error) {
    switch plural {
    case "skills":
        return ComponentSkill, nil
    case "agents":
        return ComponentAgent, nil
    case "commands":
        return ComponentCommand, nil
    default:
        return "", fmt.Errorf("unknown component type directory: %s", plural)
    }
}
```

- [ ] In `internal/downloader/component.go`, replace the `componentMetaTable` dir strings with `ct.PluralDir()` (or keep the table but derive it using `PluralDir`).
- [ ] In `internal/updater/updater.go`, replace `pluralToComponentType` with `models.ComponentTypeFromPlural` and delete the private function.
- [ ] In `internal/uninstaller/uninstaller.go`, replace the hardcoded `"skills"`, `"agents"`, `"commands"` slice literals and switch cases with `models.ComponentSkill.PluralDir()` etc. or a loop over `[]models.ComponentType{models.ComponentSkill, models.ComponentAgent, models.ComponentCommand}`.
- [ ] Build: `go build ./...`
- [ ] Test: `go test ./...`
- [ ] Commit:
```
git add internal/models/models.go internal/downloader/component.go internal/updater/updater.go internal/uninstaller/uninstaller.go
git commit -m "refactor(models): add PluralDir/ComponentTypeFromPlural, consolidate plural string literals"
```

---

### Task 6: Add `EffectiveProfile` helper

The pattern "if explicit profile given use it, else use active profile, else use base dir" is copy-pasted in `materialize/service.go::getSourceDir`, `link/service.go`, and `install/service.go`.

**Files:**
- Create: `pkg/profiles/effective.go`
- Modify: `pkg/services/materialize/service.go` (replace `getSourceDir`)
- Modify: `pkg/services/link/service.go` (replace inline profile resolution)
- Modify: `pkg/services/install/service.go` (replace inline profile resolution)

- [ ] Create `pkg/profiles/effective.go`:

```go
package profiles

import (
    "fmt"
    "path/filepath"

    "github.com/tjg184/agent-smith/pkg/paths"
)

// ResolveEffectiveDir returns the base directory and profile name to use for an operation.
// If explicit is non-empty and not "base", validates the profile exists and returns its dir.
// If explicit is "base", returns the base ~/.agent-smith dir with no profile name.
// Otherwise falls back to the active profile (if any), then the base dir.
func ResolveEffectiveDir(pm *ProfileManager, explicit string) (dir string, profileName string, err error) {
    baseDir, err := paths.GetAgentsDir()
    if err != nil {
        return "", "", fmt.Errorf("failed to get agent-smith directory: %w", err)
    }

    if explicit != "" {
        if explicit == "base" {
            return baseDir, "", nil
        }

        profiles, err := pm.ScanProfiles()
        if err != nil {
            return "", "", fmt.Errorf("failed to scan profiles: %w", err)
        }

        for _, p := range profiles {
            if p.Name == explicit {
                profilesDir, err := paths.GetProfilesDir()
                if err != nil {
                    return "", "", fmt.Errorf("failed to get profiles directory: %w", err)
                }
                return filepath.Join(profilesDir, explicit), explicit, nil
            }
        }

        return "", "", fmt.Errorf("profile '%s' not found", explicit)
    }

    active, err := pm.GetActiveProfile()
    if err != nil {
        return "", "", fmt.Errorf("failed to check active profile: %w", err)
    }

    if active != "" {
        profilesDir, err := paths.GetProfilesDir()
        if err != nil {
            return "", "", fmt.Errorf("failed to get profiles directory: %w", err)
        }
        return filepath.Join(profilesDir, active), active, nil
    }

    return baseDir, "", nil
}
```

- [ ] In `pkg/services/materialize/service.go`, replace `getSourceDir` body with a call to `profiles.ResolveEffectiveDir(s.profileManager, profile)` and delete the old implementation.
- [ ] In `pkg/services/link/service.go`, find the inline profile resolution and replace with `profiles.ResolveEffectiveDir(...)`.
- [ ] In `pkg/services/install/service.go`, do the same.
- [ ] Build: `go build ./...`
- [ ] Test: `go test ./...`
- [ ] Commit:
```
git add pkg/profiles/effective.go pkg/services/materialize/service.go pkg/services/link/service.go pkg/services/install/service.go
git commit -m "refactor(profiles): extract ResolveEffectiveDir to eliminate copy-pasted profile resolution"
```

---

## Phase 2 — Split `internal/linker/linker.go`

The goal is to move methods off `linker.go` into sub-package files. Each sub-package exposes free functions that accept `*ComponentLinker` as their first parameter. Since Go does not allow methods to be defined in external packages on a type, we use free functions and have the parent methods delegate.

The `ComponentLinker` struct, its constructor, and small utilities stay in the root `internal/linker` package. The sub-packages live at `internal/linker/sync`, `internal/linker/unlink`, `internal/linker/display`, `internal/linker/profilepicker`. **Because these sub-packages are children of `internal/linker`, they cannot import `internal/linker` (that would be circular).** Instead they receive all needed data as parameters.

**Strategy:** Rather than delegating from `*ComponentLinker` methods to sub-package functions (which would create a circular import), the sub-packages define their own types/functions that accept raw data (dirs, targets, formatter, profileManager). The parent `*ComponentLinker` methods pass their fields in. This is the clean cut.

### Task 7: Create `internal/linker/sync` sub-package

Move symlink creation and link operations into a new sub-package.

**Files:**
- Create: `internal/linker/sync/sync.go`
- Modify: `internal/linker/linker.go` (remove moved functions, delegate)

- [ ] Create `internal/linker/sync/sync.go` containing the implementation of:
  - `createSymlink(src, dst string) error`
  - `createJunction(src, dst string) error`  
  - `copyDirectory(src, dst string) error`
  - `copyFile(src, dst string) error`
  - `LinkComponent(agentsDir string, targets []config.Target, formatter *formatter.Formatter, pm linkerProfileManager, componentType, componentName string) error`
  - `LinkComponentsByType(agentsDir string, targets []config.Target, formatter *formatter.Formatter, pm linkerProfileManager, componentType string) error`
  - `LinkAllComponents(agentsDir string, targets []config.Target, formatter *formatter.Formatter, pm linkerProfileManager) error`
  - `DetectAndLinkLocalRepositories(agentsDir string, targets []config.Target, formatter *formatter.Formatter, detector *detector.RepositoryDetector) error`
  
  Define a `linkerProfileManager` interface within `sync` for the profile operations needed.

- [ ] In `internal/linker/linker.go`, update the existing `LinkComponent`, `LinkComponentsByType`, `LinkAllComponents`, `DetectAndLinkLocalRepositories`, `createSymlink`, `createJunction`, `copyDirectory`, `copyFile` methods to delegate to `sync.*`.
- [ ] Build: `go build ./...`
- [ ] Test: `go test ./...`
- [ ] Commit:
```
git add internal/linker/sync/ internal/linker/linker.go
git commit -m "refactor(linker): extract link operations into internal/linker/sync sub-package"
```

---

### Task 8: Create `internal/linker/unlink` sub-package

**Files:**
- Create: `internal/linker/unlink/unlink.go`
- Modify: `internal/linker/linker.go`

- [ ] Create `internal/linker/unlink/unlink.go` containing:
  - `UnlinkComponent(agentsDir string, targets []config.Target, formatter *formatter.Formatter, pm unlinkerProfileManager, componentType, componentName, targetFilter string) error`
  - `UnlinkComponentsByType(agentsDir string, targets []config.Target, formatter *formatter.Formatter, pm unlinkerProfileManager, componentType, targetFilter string, force bool) error`
  - `UnlinkAllComponents(agentsDir string, targets []config.Target, formatter *formatter.Formatter, pm unlinkerProfileManager, targetFilter string, force bool, allProfiles bool) error`
  - Private helpers: `isSymlinkFromCurrentProfile`, `isSymlinkFromAgentSmith`, `anyProfilesExist`

- [ ] In `internal/linker/linker.go`, delegate `UnlinkComponent`, `UnlinkComponentsByType`, `UnlinkAllComponents` to `unlink.*`.
- [ ] Build: `go build ./...`
- [ ] Test: `go test ./...`
- [ ] Commit:
```
git add internal/linker/unlink/ internal/linker/linker.go
git commit -m "refactor(linker): extract unlink operations into internal/linker/unlink sub-package"
```

---

### Task 9: Create `internal/linker/display` sub-package

**Files:**
- Create: `internal/linker/display/display.go`
- Modify: `internal/linker/linker.go`

- [ ] Create `internal/linker/display/display.go` containing:
  - `ListLinkedComponents(agentsDir string, targets []config.Target, formatter *formatter.Formatter, pm displayProfileManager) error`
  - `ShowLinkStatus(agentsDir string, targets []config.Target, formatter *formatter.Formatter, pm displayProfileManager, linkedOnly bool) error`
  - `ShowAllProfilesLinkStatus(agentsDir string, targets []config.Target, formatter *formatter.Formatter, pm displayProfileManager, profileFilter []string, linkedOnly bool) error`
  - Private helpers: `renderLinkSummary`, `linkStatusLegendItems`, `linkStatusSymbol`

- [ ] In `internal/linker/linker.go`, delegate `ListLinkedComponents`, `ShowLinkStatus`, `ShowAllProfilesLinkStatus` to `display.*`.
- [ ] Build: `go build ./...`
- [ ] Test: `go test ./...`
- [ ] Commit:
```
git add internal/linker/display/ internal/linker/linker.go
git commit -m "refactor(linker): extract display operations into internal/linker/display sub-package"
```

---

### Task 10: Create `internal/linker/profilepicker` sub-package + `io.Reader` for prompts

**Files:**
- Create: `internal/linker/profilepicker/profilepicker.go`
- Modify: `internal/linker/linker.go`

- [ ] Create `internal/linker/profilepicker/profilepicker.go` containing:
  - `ProfileMatch` struct
  - `SearchComponentInProfiles(pm pickerProfileManager, componentType, componentName string) ([]ProfileMatch, error)`
  - `PromptProfileSelection(componentType, componentName string, matches []ProfileMatch, in io.Reader, out io.Writer) (profileName string, profileDir string, err error)` — replace `fmt.Scanln` with `bufio.NewReader(in).ReadString('\n')`

- [ ] In `internal/linker/linker.go`, update `searchComponentInProfiles` and `promptProfileSelection` to delegate to `profilepicker.*`, passing `os.Stdin` / `os.Stdout` as defaults.
- [ ] Build: `go build ./...`
- [ ] Test: `go test ./...`
- [ ] Commit:
```
git add internal/linker/profilepicker/ internal/linker/linker.go
git commit -m "refactor(linker): extract profile picker into sub-package, accept io.Reader for prompts"
```

---

## Phase 3 — Split `pkg/profiles/manager.go`

Same strategy: sub-packages under `pkg/profiles/*` expose free functions, `manager.go` delegates. Sub-packages that don't need the linker don't import it.

### Task 11: Create `pkg/profiles/activation` sub-package

**Files:**
- Create: `pkg/profiles/activation/activation.go`
- Modify: `pkg/profiles/manager.go`

- [ ] Create `pkg/profiles/activation/activation.go` with:
  - `GetActiveProfile(profilesDir string) (string, error)`
  - `ActivateProfile(profilesDir string, profileName string) (*ProfileActivationResult, error)` — define `ProfileActivationResult` here
  - `ActivateProfileWithResult(profilesDir string, profileName string) (*ProfileActivationResult, error)`
  - `DeactivateProfile(profilesDir string) error`

- [ ] In `pkg/profiles/manager.go`, delegate `GetActiveProfile`, `ActivateProfile`, `ActivateProfileWithResult`, `DeactivateProfile` to `activation.*`.
- [ ] Build: `go build ./...`
- [ ] Test: `go test ./pkg/profiles/...`
- [ ] Commit:
```
git add pkg/profiles/activation/ pkg/profiles/manager.go
git commit -m "refactor(profiles): extract profile activation into pkg/profiles/activation sub-package"
```

---

### Task 12: Create `pkg/profiles/profilemeta` sub-package

**Files:**
- Create: `pkg/profiles/profilemeta/profilemeta.go`
- Modify: `pkg/profiles/manager.go`

- [ ] Create `pkg/profiles/profilemeta/profilemeta.go` with:
  - `ProfileMetadata` struct (move from manager.go)
  - `Save(profileDir string, sourceURL string) error` (the repo-type variant)
  - `SaveUser(profileDir string) error` (the user-type variant)
  - `Load(profileDir string) (*ProfileMetadata, error)`
  - `GetProfileType(profileDir string) string`
  - `GenerateNameFromRepo(repoURL string) (string, error)`
  - `FindBySourceURL(profilesDir string, sourceURL string) (string, error)`
  - Private: `sanitizeForProfileName`, `validateProfileName`

- [ ] In `pkg/profiles/manager.go`, delegate these methods to `profilemeta.*`.
- [ ] Build: `go build ./...`
- [ ] Test: `go test ./pkg/profiles/...`
- [ ] Commit:
```
git add pkg/profiles/profilemeta/ pkg/profiles/manager.go
git commit -m "refactor(profiles): extract profile metadata into pkg/profiles/profilemeta sub-package"
```

---

### Task 13: Create `pkg/profiles/scanner` sub-package

**Files:**
- Create: `pkg/profiles/scanner/scanner.go`
- Modify: `pkg/profiles/manager.go`

- [ ] Create `pkg/profiles/scanner/scanner.go` with:
  - `ScanProfiles(profilesDir string) ([]*Profile, error)` — define minimal `Profile` type or import from root
  - `CountComponents(profileDir string) (int, error)`
  - `GetComponentNames(profileDir string, componentType string) ([]string, error)`
  - `GetComponentSource(profileDir string, componentType string, componentName string) (string, error)`

- [ ] In `pkg/profiles/manager.go`, delegate `ScanProfiles`, `CountComponents`, `GetComponentNames`, `GetComponentSource` to `scanner.*`.
- [ ] Build: `go build ./...`
- [ ] Test: `go test ./pkg/profiles/...`
- [ ] Commit:
```
git add pkg/profiles/scanner/ pkg/profiles/manager.go
git commit -m "refactor(profiles): extract profile scanner into pkg/profiles/scanner sub-package"
```

---

### Task 14: Create `pkg/profiles/copy` sub-package

**Files:**
- Create: `pkg/profiles/copy/copy.go`
- Modify: `pkg/profiles/manager.go`

- [ ] Create `pkg/profiles/copy/copy.go` with:
  - `CopyComponentBetweenProfiles(profilesDir string, lockService services.ComponentLockService, src, dst, componentType, componentName string) error` — replace the private hand-rolled JSON parse with `metadata.LoadLockFile`
  - `AddComponentToProfile(profilesDir string, profileName, componentType, componentName string) error`
  - `RemoveComponentFromProfile(profilesDir string, profileName, componentType, componentName string) error`
  - Private: `copyDirectory(src, dst string) error`

- [ ] In `pkg/profiles/manager.go`, delegate these methods to `copy.*`.
- [ ] Build: `go build ./...`
- [ ] Test: `go test ./pkg/profiles/...`
- [ ] Commit:
```
git add pkg/profiles/copy/ pkg/profiles/manager.go
git commit -m "refactor(profiles): extract component copy into pkg/profiles/copy sub-package"
```

---

### Task 15: Create `pkg/profiles/cherrypick` sub-package + `io.Reader` for prompts

**Files:**
- Create: `pkg/profiles/cherrypick/cherrypick.go`
- Modify: `pkg/profiles/manager.go`

- [ ] Create `pkg/profiles/cherrypick/cherrypick.go` with:
  - `ComponentItem` struct
  - `GetAllAvailableComponents(profilesDir string, excludeProfile string) ([]ComponentItem, error)`
  - `PromptComponentSelection(items []ComponentItem, in io.Reader, out io.Writer) ([]ComponentItem, error)` — replace `fmt.Scanln` with `bufio.NewReader(in).ReadString('\n')`
  - `CherryPickComponents(pm cherryPickDeps, targetProfile string, in io.Reader, out io.Writer) error`

- [ ] In `pkg/profiles/manager.go`, delegate `GetAllAvailableComponents`, `PromptComponentSelection`, `CherryPickComponents` to `cherrypick.*`, passing `os.Stdin`/`os.Stdout` as defaults.
- [ ] Build: `go build ./...`
- [ ] Test: `go test ./pkg/profiles/...`
- [ ] Commit:
```
git add pkg/profiles/cherrypick/ pkg/profiles/manager.go
git commit -m "refactor(profiles): extract cherry-pick into sub-package, accept io.Reader for prompts"
```

---

## Phase 4 — Split `pkg/services/materialize/service.go`

By Phase 4 the `loadLockFile` duplication is already gone (Task 2) and `getSourceDir` is already replaced (Task 6). The remaining split separates the write-path from the read/display path.

### Task 16: Create `pkg/services/materialize/sync` sub-package

**Files:**
- Create: `pkg/services/materialize/sync/sync.go`
- Modify: `pkg/services/materialize/service.go`

- [ ] Create `pkg/services/materialize/sync/sync.go` containing the implementations of:
  - `MaterializeComponent(s materializeDeps, componentType, componentName string, opts services.MaterializeOptions) error`
  - `MaterializeAll(s materializeDeps, opts services.MaterializeOptions) error`
  - `MaterializeByType(s materializeDeps, componentType string, opts services.MaterializeOptions) error`
  - `UpdateMaterialized(s materializeDeps, componentType, componentName string, opts services.MaterializeOptions) error`
  - `buildFilesystemNameMap(baseDir string) (map[string]componentInfo, error)` — using `metadataPkg.LoadLockFile`
  
  Define `materializeDeps` as an interface providing formatter, logger, profileManager (as `profiles.ResolveEffectiveDir`-compatible), and postprocessorRegistry access.

- [ ] In `pkg/services/materialize/service.go`, replace the method bodies of `MaterializeComponent`, `MaterializeAll`, `MaterializeByType`, `UpdateMaterialized` with delegation calls to `sync.*`.
- [ ] Build: `go build ./...`
- [ ] Test: `go test ./pkg/services/materialize/...`
- [ ] Commit:
```
git add pkg/services/materialize/sync/ pkg/services/materialize/service.go
git commit -m "refactor(materialize): extract write-path into pkg/services/materialize/sync sub-package"
```

---

### Task 17: Create `pkg/services/materialize/status` sub-package

**Files:**
- Create: `pkg/services/materialize/status/status.go`
- Modify: `pkg/services/materialize/service.go`

- [ ] Create `pkg/services/materialize/status/status.go` containing:
  - `ShowStatus(s statusDeps, opts services.MaterializeStatusOptions) error`
  - `ShowComponentInfo(s statusDeps, componentType, componentName string, opts services.MaterializeOptions) error`
  - `ListMaterialized(s statusDeps, opts services.MaterializeOptions) error`

- [ ] In `pkg/services/materialize/service.go`, delegate `ShowStatus`, `ShowComponentInfo`, `ListMaterialized` to `status.*`.
- [ ] Build: `go build ./...`
- [ ] Test: `go test ./pkg/services/materialize/...`
- [ ] Commit:
```
git add pkg/services/materialize/status/ pkg/services/materialize/service.go
git commit -m "refactor(materialize): extract status/display into pkg/services/materialize/status sub-package"
```

---

## Final Verification

- [ ] `go build ./...` — clean build
- [ ] `go test ./...` — all tests pass
- [ ] `go vet ./...` — no vet errors
- [ ] Review that no file exceeds ~400 lines (the god objects should now be gone)
- [ ] Commit if any cleanup needed:
```
git commit -m "chore: final cleanup after god object decomposition"
```
