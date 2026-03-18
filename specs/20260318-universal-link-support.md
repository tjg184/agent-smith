# Spec: Universal Target Support for Link Command

**Created**: 2026-03-18
**Extends**: `20260203-1334-universal-target-support.md` (materialize-only implementation)

---

## Overview

Extend `agent-smith link` to support the `universal` target (`~/.agents/`), matching the same capability already provided by `materialize`. Universal is target-agnostic storage usable by any AI assistant.

---

## User Stories

- [ ] Story-001: As a developer, I want to link components to `~/.agents/` so that any AI assistant can access them without editor-specific config.

  **Acceptance Criteria:**
  - `agent-smith link skill <name> --to universal` creates a symlink in `~/.agents/skills/`
  - `agent-smith link agent <name> --to universal` creates a symlink in `~/.agents/agents/`
  - `agent-smith link command <name> --to universal` creates a symlink in `~/.agents/commands/`
  - `~/.agents/` is created if it does not exist
  - Works with `--profile` flag

  **Testing Criteria:**
  - Integration: link skill to universal, verify symlink exists in `~/.agents/skills/`
  - Integration: link when `~/.agents/` absent, verify directory is created

- [ ] Story-002: As a developer, I want `link all` to auto-include universal when `~/.agents/` already exists.

  **Acceptance Criteria:**
  - `agent-smith link all` includes universal in targets when `~/.agents/` exists on disk
  - When `~/.agents/` does not exist, `link all` does NOT include universal (no auto-creation)
  - `--to universal` always includes universal (creates `~/.agents/` if needed)

  **Testing Criteria:**
  - Integration: create `~/.agents/`, run `link all`, verify universal symlinks created
  - Integration: no `~/.agents/`, run `link all`, verify no `~/.agents/` created

- [ ] Story-003: As a developer, I want `--to` flag help text to mention `universal`.

  **Acceptance Criteria:**
  - `agent-smith link --help` shows `universal` in `--to` valid values
  - Subcommand long help includes `--to universal` examples

---

## Functional Requirements

- FR-1: `DetectAllTargets()` SHALL include `UniversalTarget` when `~/.agents/` exists on disk
- FR-2: When `--to universal` is explicitly passed and `~/.agents/` does not exist, the link service SHALL create it and proceed (not error)
- FR-3: `addLinkTargetFlags` `--to` description SHALL list `universal` as a valid option
- FR-4: `cmd/link.go` help text SHALL include `universal` in FLAGS description and at least one example
- FR-5: `--to all` behavior follows `DetectAllTargets()` — includes universal only if `~/.agents/` already exists (consistent with editor targets)

## Non-Goals

- No changes to `materialize` command
- No `--to all` forced inclusion of universal (stays detection-based)
- No project-local `.agents/` linking (that is materialize's domain)

---

## Files to Modify

| File | Change |
|---|---|
| `pkg/config/target_manager.go` | Add universal detection in `DetectAllTargets` |
| `pkg/services/link/service.go` | Bootstrap `~/.agents/` when `--to universal` explicit and dir absent |
| `cmd/flags.go` | Update `--to` description to include `universal` |
| `cmd/link.go` | Update help text / examples |

---

## Implementation Checklist

- [x] `DetectAllTargets` includes universal when `~/.agents/` exists
- [x] `createLinkerWithFilterAndProfile` bootstraps universal target when explicitly requested
- [x] `addLinkTargetFlags` description updated
- [x] `cmd/link.go` help text updated
- [x] Build passes (`go build ./...`)
- [x] Existing tests pass
