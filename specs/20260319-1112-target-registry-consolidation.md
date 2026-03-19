# PRD: Target Registry Consolidation

**Created**: 2026-03-19 11:12 UTC

---

## Introduction

All AI editor target definitions (opencode, claudecode, copilot, universal) are currently scattered across multiple files with hardcoded name lists repeated in at least 9 locations. Adding a new target today requires surgical edits across `target_manager.go`, `materialize/service.go`, `linker.go`, `target/service.go`, `detection.go`, `errors/helpers.go`, and `config.go` — with no compile-time guarantee that every site was updated. This PRD consolidates target metadata into a single registry so that adding a new target requires changes in at most 3 places: a new `*_target.go` file, a path constant in `paths.go`, and one registry entry.

---

## Goals

- Introduce `GetDisplayName() string` on the `Target` interface so display labels are owned by each target type
- Introduce a `builtInTargetDefs` registry in `target_manager.go` that is the single authoritative list of all built-in targets
- Eliminate every hardcoded `[]string{"opencode", "claudecode", "copilot"}` (and variants) from the codebase
- Drive `ProjectMarkers` in `detection.go` from the registry
- Consolidate `NewTargetDirectoryNotFoundError` to accept a `Target` instead of a raw string
- Remove the hardcoded `displayNames` map from `linker.go`

---

## User Stories

- [ ] Story-001: As a developer adding a new AI editor target, I want to register it in one place so that all commands, error messages, and detection logic pick it up automatically.

  **Acceptance Criteria:**
  - Adding a new entry to `builtInTargetDefs` (plus a `*_target.go` file and path constant) is sufficient for the target to appear in `GetAllTargets()`, `DetectAllTargets()`, `DetectTarget()`, `GetAvailableTargets()`, `GetAllTargetTypes()`, `ProjectMarkers`, and all display-label sites
  - No other files require modification for the new target to be fully functional
  - Existing targets (opencode, claudecode, copilot, universal) continue to behave identically

  **Testing Criteria:**
  **Unit Tests:**
  - `GetAllTargets()` returns exactly opencode, claudecode, copilot (no universal) — driven from registry
  - `GetAvailableTargets()` iterates registry, not a hardcoded slice
  - `GetAllTargetTypes()` returns all four built-in type constants from registry

- [ ] Story-002: As a developer, I want each target to own its display name so that labels like "Claude Code" never drift out of sync with the target definition.

  **Acceptance Criteria:**
  - `Target` interface exposes `GetDisplayName() string`
  - `baseTarget` provides a default implementation (title-case of `GetName()`) for custom targets
  - Each built-in target returns its canonical short name: `"OpenCode"`, `"Claude Code"`, `"Copilot"`, `"Universal"`
  - All call sites that previously used hardcoded `if/else if` chains or the `displayNames` map now call `target.GetDisplayName()`

  **Testing Criteria:**
  **Unit Tests:**
  - Each built-in target's `GetDisplayName()` returns expected value
  - `baseTarget` default returns title-cased name for arbitrary custom target names

- [ ] Story-003: As a developer, I want `NewTargetDirectoryNotFoundError` to accept a `Target` so that the error message is always consistent with the target's own metadata.

  **Acceptance Criteria:**
  - `NewTargetDirectoryNotFoundError(target config.Target)` replaces `NewTargetDirectoryNotFoundError(targetName string)`
  - Error message uses `target.GetProjectDirName()` and `target.GetDisplayName()` — no switch statement
  - All call sites updated to pass the `Target` object

  **Testing Criteria:**
  **Unit Tests:**
  - Error message for each built-in target matches expected format using the target's own metadata

- [ ] Story-004: As a developer, I want `ProjectMarkers` in `detection.go` to be derived from the registry so that new targets are automatically detected as project roots.

  **Acceptance Criteria:**
  - `ProjectMarkers` is computed from `config.GetAllTargets()` (or the registry) at init time, not hardcoded
  - `.agents/` (universal target) is included
  - Detection behavior is unchanged for existing markers (`.opencode/`, `.claude/`, `.github/`)

  **Testing Criteria:**
  **Unit Tests:**
  - `ProjectMarkers` contains the project dir name of every registered built-in target
  - Adding a hypothetical new target to the registry adds its marker automatically

- [ ] Story-005: As a developer, I want `validateConfig` to block all built-in target names from being used as custom target names, not just "opencode" and "claudecode".

  **Acceptance Criteria:**
  - `validateConfig` uses `builtInTargetNames()` (from registry) to build the reserved-name set
  - "copilot" and "universal" are now also blocked as custom target names
  - Error message lists all reserved names

  **Testing Criteria:**
  **Unit Tests:**
  - Attempting to use "copilot" or "universal" as a custom target name returns a validation error

- [ ] Story-006: As a developer, I want the hardcoded `displayNames` map removed from `linker.go` so that display labels are consistent with the target registry.

  **Acceptance Criteria:**
  - `displayName(name string)` function and its map are removed
  - Call sites build a `map[string]string` from `cl.targets` (using `GetName() → GetDisplayName()`) and use that instead

  **Testing Criteria:**
  **Unit Tests:**
  - Integration: linking flow produces correct display names for each built-in target (covered by existing linker tests if any)

---

## Functional Requirements

- FR-1: The system SHALL expose `GetDisplayName() string` on the `Target` interface.
- FR-2: The system SHALL define a `builtInTargetDefs` registry slice in `target_manager.go` containing all four built-in targets.
- FR-3: `GetAllTargets()`, `DetectTarget()`, `DetectAllTargets()`, `GetAvailableTargets()`, and `GetAllTargetTypes()` SHALL iterate `builtInTargetDefs` rather than hand-written switch/if chains.
- FR-4: `GetAllTargets()` SHALL exclude `TargetUniversal` (opt-in only); `GetAvailableTargets()` SHALL include only targets whose home directory exists.
- FR-5: `materialize/service.go` SHALL obtain the target list from `config.GetAllTargets()` and display labels from `target.GetDisplayName()` + `target.GetProjectDirName()`.
- FR-6: `errors/helpers.go` `NewTargetDirectoryNotFoundError` SHALL accept `config.Target` and derive all strings from it.
- FR-7: `pkg/project/detection.go` `ProjectMarkers` SHALL be derived from the built-in target registry.
- FR-8: `internal/linker/linker.go` SHALL remove the hardcoded `displayNames` map and use `target.GetDisplayName()` instead.
- FR-9: `pkg/services/target/service.go` `builtInNames` SHALL be derived from `config.GetAllTargetTypes()`.
- FR-10: `pkg/config/config.go` `validateConfig` SHALL use `builtInTargetNames()` for reserved-name checking.

---

## Non-Goals

- No new targets are being added in this work
- No changes to target behavior, paths, or linking logic
- No UI or CLI output format changes (labels may change only where they were previously incorrect/inconsistent)
- `universal` opt-in semantics are not changed (`--target all` does not include universal)
- No changes to `paths.go` (path constants remain as-is)
- No changes to custom target behavior beyond the `validateConfig` reserved-name fix
