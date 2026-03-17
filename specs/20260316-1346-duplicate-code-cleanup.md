# Duplicate Code Cleanup & Refactoring

**Status:** 🔲 PENDING
**Priority:** Medium
**Created:** 2026-03-16

## Problem Statement

Across the codebase there are roughly 1,100 lines of duplicated or near-identical code spread over 15+ files. The duplications fall into three tiers:

1. **Core abstractions** — three separate downloader types (`SkillDownloader`, `AgentDownloader`, `CommandDownloader`) with identical structs, constructors, and methods; `expandHome` defined three times; target interface methods copy-pasted across five target types.
2. **Command layer** — flag registrations and `Run` closure bodies repeated verbatim for every component type in `materialize.go`, `link.go`, `unlink.go`, `install.go`, and `uninstall.go`.
3. **Linker internals & misc** — duplicated summary table rendering, legend rendering, and `linkType` symbol conversion in `linker.go`; unused dead code; inconsistent use of the `CloneShallow` abstraction.

All changes must be **behaviour-preserving** — no functional changes, no new CLI commands, no new tests required.

---

## Objectives

- Eliminate duplicate code without changing any observable behaviour
- Consolidate shared logic into well-named helpers embedded in the types that already own it
- Remove dead code (`joinStrings`, duplicate `determineDestinationFolderName`)
- Ensure the project still builds and passes existing tests after every phase

---

## Implementation Plan

### Phase 1 — Core Abstractions

#### 1.1 `baseDownloader` struct (`internal/downloader/`)

**Affected files:** `skill.go`, `agent.go`, `command.go`, `common.go`

All three downloader types have identical struct shapes and near-identical constructors. Extract a shared `baseDownloader` struct into `common.go`.

**Changes:**

- Add `baseDownloader` struct to `common.go` with fields: `baseDir`, `detector`, `cloner`, `formatter`
- Add `newBaseDownloader(baseDir string) baseDownloader` constructor helper
- Move `parseRepoURL` method from all three typed files onto `baseDownloader` (currently duplicated verbatim 3×)
- Add `detectSourceType(fullURL string) string` free function to `common.go` to replace the 9 identical inline blocks:
  ```go
  sourceType := "github"
  if strings.Contains(fullURL, "gitlab") { sourceType = "gitlab" }
  else if strings.HasPrefix(fullURL, "git@") || strings.HasPrefix(fullURL, "ssh://") { sourceType = "git" }
  ```
- Add `createComponentMarkdownFile(filePath, componentType, name, source string) error` free function to replace the three `create*File` methods
- Move `saveLockFile` onto `baseDownloader` — the three typed implementations are identical; the `componentType` string is already a parameter
- Add `withCleanup(dir string, fn func() error) error` method on `baseDownloader` to encapsulate the 9 identical `shouldCleanup` defer patterns
- Update `SkillDownloader`, `AgentDownloader`, `CommandDownloader` to embed `baseDownloader` and delete the extracted methods/functions from each typed file
- Update the three typed constructors to delegate to `newBaseDownloader`

#### 1.2 Export `ExpandHome` from `pkg/paths` (`pkg/paths/paths.go`, `pkg/config/`)

`expandHome` / `expandHomePath` is privately defined in three packages:
- `pkg/paths/paths.go`
- `pkg/config/opencode_target.go`
- `pkg/config/config.go`

**Changes:**

- Rename the existing private `expandHome` in `pkg/paths/paths.go` to exported `ExpandHome`
- Update all callers within `pkg/paths` to use `ExpandHome`
- Delete `expandHome` from `pkg/config/opencode_target.go`; replace all call sites with `paths.ExpandHome(...)`
- Delete `expandHomePath` from `pkg/config/config.go`; replace all call sites with `paths.ExpandHome(...)`

#### 1.3 `baseTarget` struct (`pkg/config/`)

All four concrete target types (`opencodeTarget`, `claudeCodeTarget`, `copilotTarget`, `universalTarget`) implement the same 8+ methods with identical bodies (~65 lines each, ~260 lines total duplication).

**Changes:**

- Create `pkg/config/base_target.go` with a `baseTarget` struct containing fields `baseDir string` and `projectDirName string`
- Move the following methods onto `baseTarget`:
  - `GetGlobalBaseDir() (string, error)`
  - `GetGlobalSkillsDir() (string, error)`
  - `GetGlobalAgentsDir() (string, error)`
  - `GetGlobalCommandsDir() (string, error)`
  - `GetGlobalComponentDir(componentType string) (string, error)`
  - `GetDetectionConfigPath() (string, error)`
  - `GetProjectBaseDir(projectRoot string) string`
  - `GetProjectComponentDir(projectRoot, componentType string) (string, error)`
- Embed `baseTarget` in each of the four concrete target types; delete the now-redundant method implementations from each typed file
- `customTarget` can remain as-is since it has meaningfully different behaviour

---

### Phase 2 — Command Layer

#### 2.1 Shared flag helpers (`cmd/`)

Create `cmd/flags.go` (new file) with helpers that register repeated flag sets:

```go
func addMaterializeFlags(cmd *cobra.Command)         // target, project-dir, force, dry-run, profile
func addSourceFlag(cmd *cobra.Command)               // source
func addLinkTargetFlags(cmd *cobra.Command)          // to, profile
func addUnlinkFlags(cmd *cobra.Command)              // target, profile
func addForceFlag(cmd *cobra.Command)                // force
func addInstallFlags(cmd *cobra.Command)             // profile, install-dir
func addUninstallComponentFlags(cmd *cobra.Command)  // profile, source
```

#### 2.2 `cmd/materialize.go` cleanup

Currently registers the same 4-5 flags 6+ times and has identical `Run` closures differing only in the component type string.

**Changes:**

- Replace all repeated flag registration blocks with calls to `addMaterializeFlags(cmd)` / `addSourceFlag(cmd)`
- Add factory functions:
  ```go
  func makeMaterializeComponentRun(componentType string) func(*cobra.Command, []string)
  func makeMaterializeTypeRun(componentType string) func(*cobra.Command, []string)
  ```
- Assign `Run: makeMaterializeComponentRun("skills")` etc. for each subcommand

#### 2.3 `cmd/link.go` cleanup

Registers the same 2 flags 6 times; `Run` closures are identical except for the component type string.

**Changes:**

- Replace repeated flag blocks with `addLinkTargetFlags(cmd)`
- Add factory functions:
  ```go
  func makeLinkRun(componentType string) func(*cobra.Command, []string)
  func makeLinkTypeRun(componentType string) func(*cobra.Command, []string)
  ```

#### 2.4 `cmd/unlink.go` cleanup

Registers the same 2-3 flags 6 times; `Run` closures identical except for component type.

**Changes:**

- Replace repeated flag blocks with `addUnlinkFlags(cmd)` / `addForceFlag(cmd)`
- Add factory functions:
  ```go
  func makeUnlinkRun(componentType string) func(*cobra.Command, []string)
  func makeUnlinkBulkRun(componentType string) func(*cobra.Command, []string)
  ```

#### 2.5 `cmd/install.go` cleanup

Registers the same 2 flags (`profile`, `install-dir`) 4 times.

**Changes:**

- Replace repeated flag blocks with `addInstallFlags(cmd)`

#### 2.6 `cmd/uninstall.go` cleanup

Registers the same 2 flags (`profile`, `source`) 3 times; `Run` closures differ only in component type.

**Changes:**

- Replace repeated flag blocks with `addUninstallComponentFlags(cmd)`
- Add factory function:
  ```go
  func makeUninstallRun(componentType string) func(*cobra.Command, []string)
  ```

---

### Phase 3 — Linker Internals & Misc

#### 3.1 `internal/linker/linker.go` — extract private helpers

Three blocks of code are duplicated between `LinkComponentsByType`/`LinkAllComponents` and between `ShowLinkStatus`/`ShowAllProfilesLinkStatus`.

**Changes:**

- Extract `renderLinkSummary(successCount, failedCount, skippedCount int, failedComponents []string)` private method — replaces the identical ~24-line results table block in both `LinkComponentsByType` and `LinkAllComponents`
- Extract `renderLinkStatusLegend()` private method — replaces the identical 5-item legend slice defined in both `ShowLinkStatus` and `ShowAllProfilesLinkStatus`
- Extract `linkStatusSymbol(linkType string, valid bool) string` private function — replaces the identical `switch linkType` block in both `ShowLinkStatus` and `ShowAllProfilesLinkStatus`

#### 3.2 Consistent `CloneShallow` usage (`internal/downloader/`, `internal/updater/`)

`skill.go` correctly uses `gitpkg.CloneShallow(cloner, path, url)`. `agent.go`, `command.go`, `bulk.go`, and `internal/updater/updater.go` bypass this abstraction and manually build `git.CloneOptions` inline (~8 lines each, 4-6 occurrences).

**Changes:**

- Replace all inline `cloneOpts` construction blocks in `agent.go`, `command.go`, `bulk.go`, and `updater.go` with calls to `gitpkg.CloneShallow()`

#### 3.3 Dead code removal (`main.go`)

Two unused symbols exist in `main.go`:

- `joinStrings` — a hand-rolled reimplementation of `strings.Join` that is defined but never called anywhere; delete it
- Private `determineDestinationFolderName` — a simpler variant of the exported function already in `internal/downloader/common.go`; verify it is unused and delete it

#### 3.4 Consistent `getTargetNames` usage (`main.go`, `internal/linker/linker.go`)

`getTargetNames` is already defined in `main.go` but multiple places inline equivalent `for` loops.

**Changes:**

- Replace any inline target-name extraction loops with calls to `getTargetNames(targets)`
- If `linker.go` also needs this, expose it from the `config` package or pass the names in as a parameter

---

## Success Criteria

1. `go build ./...` succeeds with no errors after every phase
2. `go vet ./...` reports no new issues
3. All existing integration tests pass without modification
4. No observable change in CLI behaviour (same flags, same output, same error messages)
5. Estimated ~1,100 lines of duplicate code removed across 15+ files
6. No new files created beyond `cmd/flags.go` and `pkg/config/base_target.go`

---

## Non-Goals

- No new features or CLI commands
- No changes to public interfaces or exported function signatures (except exporting `ExpandHome`)
- No new tests (existing tests are the safety net)
- No changes to integration test files
- No performance optimisations beyond what naturally follows from consolidation

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Subtle behaviour difference in `withCleanup` vs inline defer | Medium | Verify with existing integration tests |
| `ExpandHome` export changes call-site behaviour | Low | Exported functions behave identically; only visibility changes |
| Missed call sites when deleting duplicate methods | Medium | `go build` will catch any unresolved references at compile time |
| `CloneShallow` wrapper not equivalent to inline options | Medium | Review `CloneShallow` implementation before replacing each call site |

---

## File Change Summary

| File | Change Type |
|------|-------------|
| `internal/downloader/common.go` | Extend — add `baseDownloader`, `detectSourceType`, `createComponentMarkdownFile` |
| `internal/downloader/skill.go` | Simplify — embed `baseDownloader`, remove duplicated methods |
| `internal/downloader/agent.go` | Simplify — embed `baseDownloader`, remove duplicated methods, fix `CloneShallow` |
| `internal/downloader/command.go` | Simplify — embed `baseDownloader`, remove duplicated methods, fix `CloneShallow` |
| `internal/downloader/bulk.go` | Fix — replace inline clone options with `CloneShallow` |
| `internal/updater/updater.go` | Fix — replace inline clone options with `CloneShallow` |
| `pkg/paths/paths.go` | Export — rename `expandHome` → `ExpandHome` |
| `pkg/config/opencode_target.go` | Simplify — embed `baseTarget`, remove duplicate methods, remove `expandHome` |
| `pkg/config/claudecode_target.go` | Simplify — embed `baseTarget`, remove duplicate methods |
| `pkg/config/copilot_target.go` | Simplify — embed `baseTarget`, remove duplicate methods |
| `pkg/config/universal_target.go` | Simplify — embed `baseTarget`, remove duplicate methods |
| `pkg/config/config.go` | Fix — remove `expandHomePath`, call `paths.ExpandHome` |
| `pkg/config/base_target.go` | New — `baseTarget` struct with shared interface methods |
| `cmd/flags.go` | New — shared flag registration helpers |
| `cmd/materialize.go` | Simplify — use helpers and factory `Run` functions |
| `cmd/link.go` | Simplify — use helpers and factory `Run` functions |
| `cmd/unlink.go` | Simplify — use helpers and factory `Run` functions |
| `cmd/install.go` | Simplify — use `addInstallFlags` |
| `cmd/uninstall.go` | Simplify — use helpers and factory `Run` function |
| `internal/linker/linker.go` | Simplify — extract `renderLinkSummary`, `renderLinkStatusLegend`, `linkStatusSymbol` |
| `main.go` | Simplify — delete `joinStrings`, delete unused `determineDestinationFolderName` |
