# PRD: Makefile to Justfile Migration

**Created**: 2026-02-02 15:38 UTC

---

## Introduction

Migrate the project's build automation from GNU Make (Makefile) to Just (justfile) for improved developer experience with cleaner syntax, better defaults, and cross-platform compatibility. This change modernizes the build tooling while maintaining all existing functionality.

## Goals

- Replace Makefile with justfile while preserving all existing recipes
- Update all documentation references from Make to Just
- Add new `install` command for building and installing the binary
- Maintain backward compatibility in terms of available commands
- Improve developer experience with Just's better help output

## User Stories

- [x] Story-001: As a developer, I want to replace the Makefile with a justfile so that I can use modern build automation with cleaner syntax.

  **Acceptance Criteria:**
  - Remove existing Makefile
  - Create new justfile with all existing recipes (test, test-unit, test-integration, test-all, test-verbose, test-integration-verbose, coverage, coverage-unit, coverage-integration, build, clean)
  - Default recipe shows list of available commands (`just --list`)
  - All recipes maintain identical functionality to their Make counterparts
  
  **Testing Criteria:**
  **Unit Tests:**
  - N/A - build automation change
  
  **Integration Tests:**
  - Manual verification: Run `just test`, `just build`, `just clean` to verify commands work
  - Verify `just` (no args) displays available recipes

- [x] Story-002: As a developer, I want updated documentation so that I know to use `just` instead of `make` commands.

  **Acceptance Criteria:**
  - Update TESTING.md Quick Reference section to use `just` commands
  - Change "Quick Start with Makefile" to "Quick Start with justfile"
  - Update all command examples from `make` to `just` throughout TESTING.md
  - Update CI/CD Integration section with `just` commands
  - Update Recommended Workflow section
  
  **Testing Criteria:**
  **Unit Tests:**
  - N/A - documentation change
  
  **Integration Tests:**
  - Manual verification: Review TESTING.md for any remaining `make` references

- [x] Story-003: As a developer, I want test files updated so that examples reflect justfile instead of Makefile.

  **Acceptance Criteria:**
  - Update internal/detector/patterns_test.go test case from "Makefile" to "justfile"
  - Test case maintains same logic (testing files without extensions)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Run `just test` to verify pattern tests still pass with new example

- [x] Story-004: As a developer, I want an install command so that I can build and install the binary to my GOPATH in one step.

  **Acceptance Criteria:**
  - Add `install` recipe to justfile
  - Recipe runs `go install .` to build and install binary
  - Display confirmation message showing installation path
  - Update is independent of existing `build` recipe
  
  **Testing Criteria:**
  **Unit Tests:**
  - N/A - build automation change
  
  **Integration Tests:**
  - Manual verification: Run `just install` and verify binary installed to $GOPATH/bin

## Functional Requirements

- FR-1: The justfile SHALL provide all recipes that existed in the Makefile
- FR-2: The system SHALL use `just` command syntax throughout all documentation
- FR-3: The install recipe SHALL execute `go install .` and display installation path
- FR-4: The default recipe SHALL display available commands via `just --list`
- FR-5: All test files SHALL reference justfile instead of Makefile in examples

## Non-Goals

- No changes to actual test logic or test coverage
- No changes to CI/CD pipeline configuration (beyond command syntax)
- No changes to the Go build process itself
- No additional recipes beyond the new `install` command

## Implementation Summary

**Completed Tasks:**

1. ✅ Created justfile with all existing recipes from Makefile
2. ✅ Removed Makefile from repository
3. ✅ Updated TESTING.md with justfile references (4 sections updated)
4. ✅ Updated test case in internal/detector/patterns_test.go
5. ✅ Added new `install` recipe for building and installing binary
6. ✅ Verified all commands maintain identical functionality

**Files Modified:**
- Created: `justfile`
- Removed: `Makefile`
- Updated: `TESTING.md` (4 sections)
- Updated: `internal/detector/patterns_test.go` (1 test case)

**Commands Available:**
- `just` - List all available recipes
- `just test` - Run unit tests
- `just test-integration` - Run integration tests
- `just test-all` - Run all tests
- `just test-verbose` - Run unit tests with verbose output
- `just test-integration-verbose` - Run integration tests with verbose output
- `just coverage` - Run unit tests with coverage
- `just coverage-integration` - Run integration tests with coverage
- `just build` - Build the agent-smith binary locally
- `just install` - Build and install agent-smith to $GOPATH/bin
- `just clean` - Clean build artifacts and test files
