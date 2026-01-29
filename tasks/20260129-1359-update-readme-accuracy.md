# PRD: Update README to Be More Accurate

**Created**: 2026-01-29 13:59 UTC

---

## Introduction

The agent-smith README.md currently has several critical omissions and inaccuracies that prevent users from discovering important features. This PRD focuses on addressing high-priority documentation gaps including undocumented commands (`link auto`, `link list`), missing CLI flags (`--target-dir`), and incorrect directory structure documentation.

## Goals

- Document all existing CLI commands and flags accurately
- Update directory structure diagram to match actual implementation
- Provide standard examples for previously undocumented features
- Maintain consistency with existing README style and organization
- Ensure users can discover all available functionality

## User Stories

- [x] Story-001: As a developer, I want to learn about the `link auto` command so that I can automatically detect and link components from my current repository.

  **Acceptance Criteria:**
  - Documentation added to the "Link" command section (lines 70-102)
  - Example showing `agent-smith link auto` usage
  - Brief explanation of automatic detection behavior (scans for SKILL.md, /agents/, /commands/ patterns)
  - Example workflow showing when to use `link auto` vs `link all`
  
  **Testing Criteria:**
  **Unit Tests:**
  - Not applicable (documentation only)
  
  **Integration Tests:**
  - Verify `link auto` command still works as documented after README update
  
  **Component Browser Tests:**
  - Not applicable (documentation only)

- [x] Story-002: As a developer, I want to learn about the `link list` command so that I can see which components are linked where.

  **Acceptance Criteria:**
  - Documentation added to the "Link" command section (lines 70-102)
  - Example showing `agent-smith link list` usage
  - Brief explanation that it shows linked components across all targets
  - Note about displaying symlinks, copied directories, and broken links
  
  **Testing Criteria:**
  **Unit Tests:**
  - Not applicable (documentation only)
  
  **Integration Tests:**
  - Verify `link list` command output format matches documentation
  
  **Component Browser Tests:**
  - Not applicable (documentation only)

- [x] Story-003: As a developer, I want to know about the `--target-dir` flag for install commands so that I can use project-local installations for testing.

  **Acceptance Criteria:**
  - Documentation added to the "Install" command section (lines 44-68)
  - Example showing `agent-smith install skill owner/repo skill-name --target-dir ./test-components`
  - Brief explanation of use cases (testing, project-local components, isolated from ~/.agents/)
  - Available on all install commands (skill, agent, command, all)
  
  **Testing Criteria:**
  **Unit Tests:**
  - Not applicable (documentation only)
  
  **Integration Tests:**
  - Verify `--target-dir` flag works with all install command types
  
  **Component Browser Tests:**
  - Not applicable (documentation only)

- [x] Story-004: As a developer, I want the directory structure documentation to be accurate so that I understand where agent-smith stores configuration files.

  **Acceptance Criteria:**
  - Add `config.json` to directory tree diagram (line 255)
  - Add `.active-profile` to directory tree diagram (line 255)
  - Place both files at the same level as existing lock files
  - Keep existing directory structure intact, only add missing files
  - Maintain consistent formatting with existing tree structure
  
  **Testing Criteria:**
  **Unit Tests:**
  - Not applicable (documentation only)
  
  **Integration Tests:**
  - Verify actual filesystem structure matches documented structure
  
  **Component Browser Tests:**
  - Not applicable (documentation only)

- [x] Story-005: As a developer, I want the install command documentation to clarify required parameters so that I don't get errors from missing arguments.

  **Acceptance Criteria:**
  - Update install examples to show required `name` parameter explicitly
  - Update line 51: `agent-smith install skill owner/repo skill-name`
  - Update line 54: `agent-smith install agent owner/repo agent-name`
  - Update line 57: `agent-smith install command owner/repo command-name`
  - Add note that name parameter is required for individual component installs
  
  **Testing Criteria:**
  **Unit Tests:**
  - Not applicable (documentation only)
  
  **Integration Tests:**
  - Verify install commands require name parameter as documented
  
  **Component Browser Tests:**
  - Not applicable (documentation only)

## Functional Requirements

- FR-1: The README SHALL document the `link auto` command with usage examples in the Link command section
- FR-2: The README SHALL document the `link list` command with usage examples in the Link command section
- FR-3: The README SHALL document the `--target-dir` flag for all install commands with testing use case examples
- FR-4: The README SHALL include `config.json` and `.active-profile` in the directory structure diagram
- FR-5: The README SHALL show required `name` parameter in install command examples
- FR-6: All new documentation SHALL maintain consistency with existing README style and format
- FR-7: All examples SHALL be accurate and tested against actual implementation

## Non-Goals

- No documentation of medium or low priority items (--all-targets flag, plural command backward compatibility)
- No creation of new "Advanced Features" section (inline updates only per user preference)
- No comprehensive examples with multiple use cases (standard examples only)
- No complete restructuring of directory structure section (minimal file additions only)
- No Ralphy YAML export (PRD markdown only per user preference)
- No updates to CONFIG.md or TESTING.md (README.md only)
- No changes to command implementation (documentation only)

## Implementation Notes

### Specific Line Ranges to Update

1. **Install Section (lines 44-68)**: Add `--target-dir` flag documentation
2. **Link Section (lines 70-102)**: Add `link auto` and `link list` commands
3. **Directory Structure (lines 251-271)**: Add `config.json` and `.active-profile` to tree
4. **Install Examples**: Clarify required `name` parameter throughout

### Style Guidelines

- Match existing README tone (concise, practical, example-driven)
- Use existing formatting conventions (bash code blocks, bullet lists, bold headers)
- Keep examples short and focused on common use cases
- Maintain alphabetical or logical ordering where established

### Testing Validation

After implementation:
- Verify all documented commands work as described
- Test all example commands against actual CLI
- Confirm directory structure matches filesystem reality
- Check that no new inaccuracies are introduced
