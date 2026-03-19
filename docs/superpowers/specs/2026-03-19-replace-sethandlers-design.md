# Design: Replace SetHandlers with Handlers Struct

**Date:** 2026-03-19  
**Branch:** refactor_duplicates  
**Status:** Approved

## Problem

`cmd/handlers.go` contains a `SetHandlers` function with 44 positional parameters. It is the wiring point between `internal/container/app.go` (which constructs services) and `cmd/` (which calls them). The function signature is impossible to read, fragile to extend, and a maintenance hazard.

## Goal

Eliminate the 44-parameter `SetHandlers` function. Replace it with a grouped `Handlers` struct and a single `Register(*Handlers)` function that performs the same assignment. No behavioral change; no change to call sites in `cmd/*.go`.

## Approach: Handlers Struct + Register

### `cmd/handlers.go`

Replace the flat `SetHandlers(f1, f2, ..., f44)` with:

1. **Domain-grouped sub-structs** — one per command group:
   - `InstallHandlers` — AddSkill, AddAgent, AddCommand, AddAll
   - `UpdateHandlers` — Update, UpdateAll
   - `LinkHandlers` — Link, LinkAll, LinkType, AutoLink, ListLinks, LinkStatus
   - `UnlinkHandlers` — Unlink, UnlinkWithProfile, UnlinkAll, UnlinkAllWithProfile, UnlinkType, UnlinkTypeWithProfile
   - `UninstallHandlers` — Uninstall, UninstallAll
   - `ProfileHandlers` — List, Show, Create, Delete, Activate, Deactivate, Add, Copy, Remove, CherryPick, Share, Rename
   - `StatusHandlers` — Status
   - `TargetHandlers` — Add, Remove, List
   - `MaterializeHandlers` — Component, Type, All, List, Info, Status, Update
   - `FindHandlers` — FindSkill

2. **`Handlers` root struct** with one field per sub-struct.

3. **`Register(h *Handlers)`** — replaces `SetHandlers`. Assigns all 44 package-level vars from the struct fields. Same logic, named fields instead of positional args.

The 44 package-level `var` declarations and all `handle*` call sites in `cmd/*.go` are **unchanged**.

### `internal/container/app.go`

Replace the `cmd.SetHandlers(f1, f2, ..., f44)` call with a `cmd.Handlers{}` struct literal (using named fields), then `cmd.Register(&h)`.

The struct literal is split into domain blocks matching the sub-structs, making the wiring self-documenting.

## Constraints

- No changes to `cmd/*.go` call sites (e.g., `handleAddSkill(...)`)
- No changes to `pkg/services/` or any service interface
- No import cycle introduced (cmd imports nothing from container; container imports cmd as before)
- Pure refactor — identical runtime behavior

## Files Changed

| File | Change |
|------|--------|
| `cmd/handlers.go` | Replace `SetHandlers` func + `var` block with sub-structs, `Handlers`, `Register` |
| `internal/container/app.go` | Replace `cmd.SetHandlers(...)` call with `cmd.Register(&cmd.Handlers{...})` |

## Verification

- `go build ./...` must pass
- `go test ./...` must pass
- No changes to test files required
