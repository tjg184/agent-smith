# PRD: Decouple Opencode Dependency

## Introduction

Refactor the Agent Smith CLI to abstract the target environment. Currently, `opencode` paths and logic are hardcoded throughout the codebase. This change introduces a configuration/target interface where `opencode` is the first and default implementation. This enables future support for other environments (targets) while maintaining `opencode` as the default behavior, ensuring extensibility without regression.

## Goals

- Abstract hardcoded `opencode` paths into a configurable Target interface
- Implement `OpencodeTarget` as the default implementation
- Refactor `ComponentLinker` to operate on the generic Target interface
- Ensure zero regression for existing `opencode` users

## User Stories

- [ ] Story-001: As a developer, I want to abstract the hardcoded `opencode` paths into a configuration interface so that the CLI can support multiple targets in the future.

  **Acceptance Criteria:**
  - `pkg/paths` no longer contains hardcoded `~/.config/opencode` global constants exposed directly
  - A new `Target` or `Configuration` interface is defined in `internal/models` or `pkg/config`
  - An `OpencodeConfiguration` struct implements this interface, providing the existing paths
  - The application initializes with `OpencodeConfiguration` by default

  **Testing Criteria:**
  **Unit Tests:**
  - Verify configuration interface returns correct paths for Opencode implementation
  - Ensure default configuration loading works as expected
  
  **Integration Tests:**
  - Verify application startup initializes correct paths without flags

- [ ] Story-002: As a developer, I want to refactor the Linker to use a generic Target interface so that linking logic is not coupled to `opencode` directory structures.

  **Acceptance Criteria:**
  - `ComponentLinker` struct accepts a `Target` interface instead of a raw `opencodeDir` string
  - Logic specific to `opencode` folder structure (e.g., `skills/`, `agents/` subdirs) is moved to the `OpencodeTarget` implementation
  - The linker calls `target.GetLinkPath(componentType, name)` instead of constructing paths manually

  **Testing Criteria:**
  **Unit Tests:**
  - Mock Target interface to verify Linker interaction
  - Test OpencodeTarget path generation logic
  
  **Integration Tests:**
  - Link a component and verify it still lands in `~/.config/opencode` via the new abstraction

- [ ] Story-003: As a user, I want the CLI to default to `opencode` behavior without extra configuration so that my existing workflows remain unchanged.

  **Acceptance Criteria:**
  - Running `agent-smith link` without arguments uses the Opencode target
  - Existing CLI flags and commands work identically to before
  - No manual configuration file is required for the default case

  **Testing Criteria:**
  **Integration Tests:**
  - Full regression test of `link`, `unlink`, and `list-links` commands
  - Verify help text reflects the default behavior

## Functional Requirements

- FR-1: The system must define a `Target` interface that specifies methods for getting component paths (`GetSkillsDir`, `GetAgentsDir`, etc.).
- FR-2: The system must provide a default `OpencodeTarget` implementation that preserves current path logic (`~/.config/opencode`).
- FR-3: The `ComponentLinker` must use the active `Target` to resolve destination paths.
- FR-4: The root command initialization must instantiate the `OpencodeTarget` if no other target is specified.

## Non-Goals

- Implementing a second target (e.g., "Personal" or "System") is out of scope for this PRD; this is purely refactoring for enablement.
- Changing the actual directory structure of `opencode` is out of scope.
