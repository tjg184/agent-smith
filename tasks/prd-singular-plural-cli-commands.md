# PRD: Singular/Plural CLI Command Consistency

## Introduction

Standardize the agent-smith CLI command syntax to use both singular and plural forms consistently across `link` and `unlink` commands. Currently, `link` uses singular forms (`skill`, `agent`, `command`) while `unlink` uses plural forms (`skills`, `agents`, `commands`), creating confusion about whether commands operate on one item or all items.

## Goals

- Eliminate confusion between singular and plural command forms
- Make the CLI self-documenting through grammatical consistency
- Prevent accidental operations (e.g., unlinking all when user meant one)
- Maintain backward compatibility with existing `unlink` syntax
- Provide helpful error messages to guide users to correct syntax

## User Stories

- [x] Story-001: As a CLI user, I want singular commands to operate on one specific component so that the grammar matches the operation.

  **Acceptance Criteria:**
  - `link skill <name>` links one specific skill
  - `link agent <name>` links one specific agent
  - `link command <name>` links one specific command
  - `unlink skill <name>` unlinks one specific skill (new)
  - `unlink agent <name>` unlinks one specific agent (new)
  - `unlink command <name>` unlinks one specific command (new)
  - Singular form requires a name argument
  - Error shown if name is not provided with singular form

  **Testing Criteria:**
  **Unit Tests:**
  - Command argument validation tests
  - Singular form name requirement tests
  
  **Integration Tests:**
  - Single component link/unlink operations
  - Error handling for missing name argument
  
  **Component Browser Tests:**
  - N/A (CLI tool, no browser components)

- [x] Story-002: As a CLI user, I want plural commands to operate on all components of a type so that the grammar clearly indicates bulk operations.

  **Acceptance Criteria:**
  - `link skills` links all skills
  - `link agents` links all agents
  - `link commands` links all commands
  - `unlink skills` unlinks all skills (existing behavior)
  - `unlink agents` unlinks all agents (existing behavior)
  - `unlink commands` unlinks all commands (existing behavior)
  - Plural form forbids name argument
  - Error shown with helpful message if name is provided with plural form

  **Testing Criteria:**
  **Unit Tests:**
  - Command argument validation for plural forms
  - Bulk operation logic tests
  
  **Integration Tests:**
  - All components link/unlink operations
  - Error handling for unexpected name argument
  
  **Component Browser Tests:**
  - N/A (CLI tool, no browser components)

- [x] Story-003: As a CLI user, I want helpful error messages when I use the wrong form so that I can quickly correct my command.

  **Acceptance Criteria:**
  - When running `link skills <name>`, show: "Error: 'skills' is for linking all skills. To link one skill, use: agent-smith link skill <name>"
  - When running `unlink skills <name>`, show same pattern
  - When running `link skill` (no name), show: "Error: 'skill' requires a component name. To link all skills, use: agent-smith link skills"
  - Error messages include 2-3 usage examples showing correct syntax
  - Suggested commands use the actual component name from the failed command

  **Testing Criteria:**
  **Unit Tests:**
  - Error message formatting tests
  - Command suggestion generation tests
  
  **Integration Tests:**
  - Error display for all incorrect syntax variations
  - Verification of suggested command accuracy
  
  **Component Browser Tests:**
  - N/A (CLI tool, no browser components)

- [x] Story-004: As an existing user, I want my current `unlink skills <name>` commands to continue working so that I don't have breaking changes.

  **Acceptance Criteria:**
  - `unlink skills <name>` continues to work (backward compatible)
  - `unlink agents <name>` continues to work (backward compatible)
  - `unlink commands <name>` continues to work (backward compatible)
  - Both old and new syntax work identically
  - No deprecation warnings shown (full backward compatibility)
  - Documentation shows new singular forms as recommended syntax

  **Testing Criteria:**
  **Unit Tests:**
  - Backward compatibility argument parsing tests
  - Both syntax paths lead to same code
  
  **Integration Tests:**
  - Old syntax operations complete successfully
  - New syntax operations complete successfully
  - Results are identical between old and new syntax
  
  **Component Browser Tests:**
  - N/A (CLI tool, no browser components)

- [x] Story-005: As a developer, I want clear command argument validation so that invalid usage is caught early with helpful feedback.

  **Acceptance Criteria:**
  - Cobra command validation checks argument count
  - Custom validation distinguishes between singular/plural usage
  - Validation function `validateSingularUsage()` checks name is provided
  - Validation function `validatePluralUsage()` checks name is not provided
  - All validation errors return formatted error messages with examples
  - Help text clearly documents singular vs plural behavior

  **Testing Criteria:**
  **Unit Tests:**
  - Validation function unit tests
  - Argument count validation tests
  - Error message format tests
  
  **Integration Tests:**
  - End-to-end validation for all command variations
  - Help text display verification
  
  **Component Browser Tests:**
  - N/A (CLI tool, no browser components)

## Functional Requirements

### Command Syntax

**FR-1: Singular Forms (One Component)**
- `agent-smith link skill <name>` - Link one skill (existing)
- `agent-smith link agent <name>` - Link one agent (existing)
- `agent-smith link command <name>` - Link one command (existing)
- `agent-smith unlink skill <name>` - Unlink one skill (NEW)
- `agent-smith unlink agent <name>` - Unlink one agent (NEW)
- `agent-smith unlink command <name>` - Unlink one command (NEW)

**FR-2: Plural Forms (All Components)**
- `agent-smith link skills` - Link all skills (NEW)
- `agent-smith link agents` - Link all agents (NEW)
- `agent-smith link commands` - Link all commands (NEW)
- `agent-smith unlink skills` - Unlink all skills (existing)
- `agent-smith unlink agents` - Unlink all agents (existing)
- `agent-smith unlink commands` - Unlink all commands (existing)

**FR-3: Backward Compatibility**
- `agent-smith unlink skills <name>` - Continue working as singular (existing)
- `agent-smith unlink agents <name>` - Continue working as singular (existing)
- `agent-smith unlink commands <name>` - Continue working as singular (existing)

### Validation Rules

**FR-4: Singular Form Validation**
- Singular commands MUST have exactly one name argument
- Error if name is missing: "Error: 'skill' requires a component name. To link all skills, use: agent-smith link skills"

**FR-5: Plural Form Validation**
- Plural commands MUST NOT have a name argument
- Error if name provided: "Error: 'skills' is for linking all skills. To link one skill, use: agent-smith link skill <name>"
- Exception: `unlink skills <name>` allowed for backward compatibility

**FR-6: Error Messages**
- All errors include verbose explanation with 2-3 examples
- Examples show both the correct command and alternative options
- Use actual component name from failed command in suggestions

### Implementation Details

**FR-7: New Cobra Subcommands**
- Add `link skills` subcommand alongside existing `link skill`
- Add `link agents` subcommand alongside existing `link agent`
- Add `link commands` subcommand alongside existing `link command`
- Add `unlink skill` subcommand alongside existing `unlink skills`
- Add `unlink agent` subcommand alongside existing `unlink agents`
- Add `unlink command` subcommand alongside existing `unlink commands`

**FR-8: Shared Implementation**
- Singular and plural forms call the same underlying functions
- `link skill <name>` and `unlink skill <name>` → `handleLink("skills", name, targetFilter)`
- `link skills` and `unlink skills` → `handleLinkType("skills", targetFilter)`
- No code duplication between singular/plural command handlers

**FR-9: Help Text**
- Update all help text to mention both singular and plural forms
- Show examples using both forms
- Clarify when to use each form

## Non-Goals (Out of Scope)

- No changes to `link all` or `unlink all` commands
- No changes to `install` command syntax
- No changes to `update` command syntax
- No changes to `profiles` command syntax
- No deprecation warnings (full backward compatibility maintained)
- No changes to command functionality, only syntax additions
- No automatic command correction (user must re-run with correct syntax)

## Technical Implementation Notes

### File Changes Required
- `cmd/root.go` - Add new subcommands for singular/plural forms
- Add validation helpers for argument checking
- Update help text across all affected commands

### Command Structure
```
link
├── skill [name]      # Existing: one skill
├── skills            # NEW: all skills
├── agent [name]      # Existing: one agent
├── agents            # NEW: all agents
├── command [name]    # Existing: one command
├── commands          # NEW: all commands
└── all               # Existing: all components

unlink
├── skill [name]      # NEW: one skill
├── skills [name]     # Existing: all skills OR one skill (backward compat)
├── agent [name]      # NEW: one agent
├── agents [name]     # Existing: all agents OR one agent (backward compat)
├── command [name]    # NEW: one command
├── commands [name]   # Existing: all commands OR one command (backward compat)
└── all               # Existing: all components
```

### Validation Logic
```go
func validateSingularUsage(cmd *cobra.Command, args []string, formName string) error {
    if len(args) == 0 {
        return fmt.Errorf("Error: '%s' requires a component name.\n\n"+
            "To %s all %ss, use:\n  agent-smith %s %ss\n\n"+
            "Examples:\n"+
            "  agent-smith %s %s my-component\n"+
            "  agent-smith %s %ss",
            formName, cmd.Use, formName, cmd.Parent().Use, formName,
            cmd.Parent().Use, formName, cmd.Parent().Use, formName)
    }
    return nil
}

func validatePluralUsage(cmd *cobra.Command, args []string, singularForm string) error {
    if len(args) > 0 {
        return fmt.Errorf("Error: '%ss' is for %sing all %ss.\n\n"+
            "To %s one %s, use:\n  agent-smith %s %s %s\n\n"+
            "Examples:\n"+
            "  agent-smith %s %ss\n"+
            "  agent-smith %s %s my-component",
            singularForm, cmd.Parent().Use, singularForm,
            cmd.Parent().Use, singularForm, cmd.Parent().Use, singularForm, args[0],
            cmd.Parent().Use, singularForm, cmd.Parent().Use, singularForm)
    }
    return nil
}
```

## Success Criteria

- All 6 new subcommands (`link skills`, `link agents`, `link commands`, `unlink skill`, `unlink agent`, `unlink command`) work correctly
- Singular forms require name argument, plural forms forbid it
- Error messages show verbose explanations with examples
- Backward compatibility maintained for all existing `unlink` commands
- Help text updated to document both forms
- No breaking changes for existing users
- Unit tests pass for all validation logic
- Integration tests verify both singular and plural forms work correctly
