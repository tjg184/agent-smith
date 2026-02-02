# PRD: Project Boundary-Aware Detection for Materialize Commands

**Created**: 2026-02-02 15:01 UTC

---

## Introduction

Fix project detection logic to respect project boundaries (like `.git/`, `go.mod`, `package.json`) and prevent materializing components to unintended locations such as the home directory. Currently, `agent-smith materialize all --target opencode` walks up the directory tree indefinitely and can find `.opencode/` in parent directories (including home directory), causing components to be materialized to the wrong location.

## Goals

- Respect project boundaries when searching for project root
- Stop at common project markers (`.git/`, `go.mod`, `package.json`, etc.) even if no `.opencode/` exists
- Prevent materialization to home directory when inside a project
- Create `.opencode/` at project root when it doesn't exist within project boundaries
- Provide clear errors when no project can be detected

## User Stories

- [x] Story-001: As a developer, I want materialize to stop at my project root (marked by `.git/`) so that `.opencode/` is created in my project, not my home directory.

  **Acceptance Criteria:**
  - `FindProjectRootFromDir()` detects common project boundary markers
  - Stops walking up directory tree at first project boundary marker
  - Returns project root location even when `.opencode/` doesn't exist
  - Prefers `.opencode/` over project markers when both exist
  - Never crosses project boundary to find `.opencode/` in parent directories
  
  **Testing Criteria:**
  **Unit Tests:**
  - Project boundary marker detection logic
  - Directory tree walking stops at boundaries
  - Prefers `.opencode/` when found before boundary
  - Returns boundary location when no `.opencode/` found
  
  **Integration Tests:**
  - Materialize in Git project creates `.opencode/` at repository root
  - Materialize from nested directory uses project root
  - Home directory `.opencode/` not used when inside project boundary

- [x] Story-002: As a developer, I want clear errors when no project boundary is detected so I know I'm not in a valid project directory.

  **Acceptance Criteria:**
  - Error message shown when no project markers found
  - Error includes list of supported project markers
  - Suggests using `--project-dir` flag or initializing a project marker
  - Does not fall back to current directory or home directory
  
  **Testing Criteria:**
  **Unit Tests:**
  - Error returned when no boundaries found
  - Error message includes helpful guidance
  
  **Integration Tests:**
  - Materialize fails in directory without project markers
  - Error message displayed to user
  - No `.opencode/` directory created

- [x] Story-003: As a developer, I want support for multiple project types so the detection works regardless of my tech stack.

  **Acceptance Criteria:**
  - Detects Git repositories (`.git/`)
  - Detects Go projects (`go.mod`)
  - Detects Node.js projects (`package.json`)
  - Detects Python projects (`pyproject.toml`)
  - Detects Rust projects (`Cargo.toml`)
  - Detects PHP projects (`composer.json`)
  - Detects Java projects (`pom.xml`, `build.gradle`)
  - Detects Ruby projects (`Gemfile`)
  - Detects Elixir projects (`mix.exs`)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Each project marker type is detected correctly
  - Multiple markers in same directory handled properly

- [x] Story-004: As a developer, I want the fallback logic removed from `main.go` so that project detection is handled consistently in one place.

  **Acceptance Criteria:**
  - Remove fallback to `os.Getwd()` in materialize skill command
  - Remove fallback to `os.Getwd()` in materialize all command
  - Remove fallback to `os.Getwd()` in materialize list command
  - Remove fallback to `os.Getwd()` in materialize info command
  - All commands use consistent `FindProjectRoot()` behavior
  - Commands fail with clear error when project not found
  
  **Testing Criteria:**
  **Unit Tests:**
  - Not applicable (removal of code)
  
  **Integration Tests:**
  - All materialize commands use consistent project detection
  - Commands error appropriately when no project found

- [x] Story-005: As a developer, I want updated documentation explaining how project detection works so I understand where `.opencode/` will be created.

  **Acceptance Criteria:**
  - README includes "Project Detection" section
  - Explains preferred markers (`.opencode/`, `.claude/`)
  - Explains project boundary markers with full list
  - Shows example of detection from nested directory
  - Documents `--project-dir` override flag
  - Notes that home directory `.opencode/` only used when explicitly in home
  
  **Testing Criteria:**
  **Unit Tests:**
  - Not applicable (documentation)

## Functional Requirements

- FR-1: The system SHALL define `ProjectBoundaryMarkers` list in `pkg/project/detection.go` containing common project markers

- FR-2: The system SHALL implement `hasProjectBoundaryMarker(dir string) bool` helper function to check for boundary markers

- FR-3: The `FindProjectRootFromDir()` function SHALL track the last seen project boundary marker while walking up the directory tree

- FR-4: The `FindProjectRootFromDir()` function SHALL return immediately when `.opencode/` or `.claude/` directories are found

- FR-5: The `FindProjectRootFromDir()` function SHALL return the project boundary location when boundary found but no `.opencode/`

- FR-6: The `FindProjectRootFromDir()` function SHALL return an error when no project markers or boundaries are found

- FR-7: The `FindProjectRootFromDir()` function SHALL stop at home directory boundary and not continue to filesystem root

- FR-8: The system SHALL log informational messages when using project boundary markers (e.g., "Using project root from .git: /path")

- FR-9: All materialize commands SHALL remove fallback logic to `os.Getwd()` and rely solely on `FindProjectRoot()`

- FR-10: Error messages SHALL include helpful guidance about initializing project markers or using `--project-dir` flag

## Non-Goals (Out of Scope)

- No changes to `--project-dir` flag behavior (already works correctly)
- No changes to materialization metadata format or lock file structure
- No changes to conflict handling or force flag behavior
- No changes to dry-run mode functionality
- No detection of IDE-specific directories (`.vscode/`, `.idea/`)
- No detection of dependency lock files (`package-lock.json`, `yarn.lock`) as primary markers
- No monorepo-specific handling or workspace detection
- No Git submodule special handling (will use nearest `.git` marker)
- No backward compatibility for home directory `.opencode/` usage within projects
