# AGENTS.md - Guidelines for LLMs Working on This Codebase

**⚠️ REQUIRED READING: All AI agents and LLMs must read and follow these guidelines before making changes to this codebase.**

---

### Godoc

Keep godoc only if it explains non-obvious behavior. A function named `determineDestinationFolderName` doesn't need "determines destination folder name" in its doc.

---

## Coding Style

### Naming

- Use full words over abbreviations when clarity matters
- Boolean names should read naturally: `isValid`, `hasPermission`, `canProceed`
- Function names should describe what they do: `validateAndTransformUserData`, not `processData`

### Error Handling

Handle errors immediately with context:

```go
result, err := someOperation()
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

### Code Structure

- Keep functions focused (< 50 lines)
- Use early returns to reduce nesting
- Deeply nested conditionals indicate a refactor is needed

---

## Project Conventions

### Key Directories

- `cmd/`: CLI commands (Cobra)
- `pkg/`: Public packages
- `internal/`: Private implementation
- `tests/integration/`: Integration tests

### Architecture

- **Profiles**: Environment contexts (dev, prod, testing)
- **Components**: Reusable units (agents, commands, skills) from git repos
- **Linkers**: Connect components to editors (OpenCode, Claude Code)
- **Targets**: Editor configurations for deployment
- **Lock Service**: Tracks component versions

### Testing

- Integration tests in `tests/integration/`
- Descriptive test names: `profile_add_lock_preservation_test.go`
- Use table-driven tests; mock external dependencies

---

## Before Submitting

- [ ] No comments restating what code does
- [ ] Names are self-documenting
- [ ] Godoc adds value beyond the signature
- [ ] Error messages provide context
- [ ] Functions are focused with early returns
- [ ] Tests cover new functionality
