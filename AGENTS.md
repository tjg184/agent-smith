# AGENTS.md - Guidelines for LLMs Working on This Codebase

**⚠️ REQUIRED READING: All AI agents and LLMs must read and follow these guidelines before making changes to this codebase.**

This document establishes coding standards and conventions for AI-assisted development on the agent-smith project. Following these guidelines ensures consistency, maintainability, and high code quality.

---

## Code Commenting Policy

**Primary Principle: Write self-documenting code. Only add comments when something is genuinely unclear.**

### When to Add Comments ✅

1. **Complex or non-obvious logic**
   - Algorithm explanations that aren't immediately clear
   - Performance optimizations with trade-offs
   - Mathematical formulas or calculations

2. **Explaining the "why" behind a decision**
   - Business logic rationale
   - Design decisions that might seem counterintuitive
   - Why one approach was chosen over alternatives

3. **Documenting workarounds or hacks**
   - Temporary fixes that need future attention
   - Bug workarounds for external dependencies
   - Platform-specific quirks

4. **Something cannot be made clear through code alone**
   - External API contracts
   - State machine transitions
   - Non-obvious edge cases

### When NOT to Add Comments ❌

1. **Never explain what obvious code does**
   ```go
   // BAD: Obvious and redundant
   // Increment counter by 1
   counter++
   
   // BAD: Restates the function name
   // Execute runs the root command
   func Execute() { ... }
   
   // BAD: Obvious from context
   // Color definitions
   cyan := color.New(color.FgCyan)
   ```

2. **Never restate code in English**
   ```go
   // BAD: Just repeating what code says
   // Loop through all files
   for _, file := range files { ... }
   
   // BAD: Obvious from variable name
   // Get the parent directory
   parentDir := filepath.Dir(currentDir)
   ```

3. **Never add boilerplate documentation for simple functions**
   ```go
   // BAD: Name already explains everything
   // NewLockService creates a new ComponentLockService
   func NewLockService() services.ComponentLockService { ... }
   
   // GOOD: Adds non-obvious context
   // NewProfileManager creates a ProfileManager that monitors the active 
   // profile file for changes, allowing hot-switching between environments
   func NewProfileManager() *ProfileManager { ... }
   ```

4. **Never add section markers without value**
   ```go
   // BAD: Doesn't add information
   // Initialize services
   lockService := NewLockService()
   profileManager := NewProfileManager()
   
   // GOOD: Skip the comment, code is clear
   lockService := NewLockService()
   profileManager := NewProfileManager()
   ```

### Godoc Documentation Guidelines

For exported functions and types, apply the same "self-documenting" principle:

- ✅ **Keep godoc if:** Function has complex behavior, non-obvious side effects, or parameters that need clarification
- ❌ **Remove godoc if:** Function name and signature fully explain what it does
- ✅ **Keep deprecation notices:** Always document deprecated APIs with alternatives

```go
// GOOD: Adds valuable context
// determineDestinationFolderName walks up from the component file directory,
// skipping component-type names (agents/commands/skills) to find the first
// non-component-type directory for preserving optional hierarchy.
func determineDestinationFolderName(componentFilePath string) string { ... }

// BAD: Redundant with function name
// Execute runs the root command
func Execute() { ... }

// Should be:
func Execute() { ... }  // No comment needed
```

---

## Coding Style Rules

### Naming Conventions

Use descriptive names that eliminate the need for comments:

```go
// GOOD: Name explains intent
func detectAvailableTargets() []Target { ... }
func validateProfileConfiguration(cfg Config) error { ... }

// BAD: Vague names requiring comments
// processData processes the user data and returns results
func processData(d Data) Results { ... }

// Should be:
func validateAndTransformUserData(d Data) Results { ... }
```

### Variable Naming

- Use full words over abbreviations when clarity matters
- Single-letter variables are fine for short scopes (loops, closures)
- Boolean variables should read naturally in conditionals

```go
// GOOD
isValidProfile := profile.Validate()
if isValidProfile { ... }

componentCount := len(components)
for i := 0; i < componentCount; i++ { ... }

// AVOID
p := profile.Validate()  // What is p?
if p { ... }
```

### Error Handling

Follow Go idioms:

```go
// GOOD: Handle errors immediately
result, err := someOperation()
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Wrap errors with context
if err := validateConfig(cfg); err != nil {
    return fmt.Errorf("invalid configuration: %w", err)
}
```

### Code Structure

- Keep functions focused and short (prefer < 50 lines)
- Extract complex logic into well-named helper functions
- Use early returns to reduce nesting

```go
// GOOD: Early returns, clear flow
func processComponent(c Component) error {
    if c == nil {
        return errors.New("component is nil")
    }
    
    if !c.IsValid() {
        return errors.New("component failed validation")
    }
    
    return c.Process()
}

// AVOID: Deep nesting
func processComponent(c Component) error {
    if c != nil {
        if c.IsValid() {
            return c.Process()
        } else {
            return errors.New("component failed validation")
        }
    }
    return errors.New("component is nil")
}
```

---

## Project-Specific Conventions

### Architecture Overview

Understanding these concepts helps you write better code:

1. **Profiles**: Environment contexts that isolate component installations (dev, prod, testing)
2. **Components**: Reusable units (agents, commands, skills) installed from git repositories
3. **Linkers**: Services that connect components to target editors (OpenCode, Claude Code)
4. **Targets**: AI editor configurations where components are deployed
5. **Lock Service**: Tracks installed components and their versions for consistency

### Key Directories

- `cmd/`: Cobra CLI command implementations
- `pkg/`: Public, reusable packages (linker, profiles, services)
- `internal/`: Private implementation details
- `tests/integration/`: End-to-end integration tests
- `schemas/`: JSON schemas for configuration validation

### Testing Approach

- Integration tests are in `tests/integration/`
- Test file names should be descriptive: `profile_add_lock_preservation_test.go`
- Use table-driven tests for multiple scenarios
- Mock external dependencies (git, filesystem) when appropriate

---

## LLM-Specific Guidance

### Before Making Changes

1. **Read existing code first** - Understand patterns before modifying
2. **Follow established conventions** - Match the style of surrounding code
3. **Trust readable code** - Don't add comments as a safety net for unclear code
4. **Refactor for clarity** - Improve names and structure instead of adding comments

### During Refactoring

1. **Remove comments that became obsolete** - If you improve code clarity, remove the comment
2. **Don't add comments to explain refactored code** - If it needs explanation, refactor more
3. **Update godoc only if behavior changed** - Don't add docs to unchanged functions

### Code Review Checklist

Before submitting changes, verify:

- [ ] No comments that restate what code does
- [ ] Variable and function names are self-explanatory
- [ ] Godoc comments add value beyond the signature
- [ ] Complex logic has "why" comments, not "what" comments
- [ ] Error messages provide useful context
- [ ] Code follows Go idioms and conventions
- [ ] Tests cover new functionality

### Common Anti-Patterns to Avoid

1. **Over-commenting during implementation**
   - Don't add a comment for every step
   - Let the code flow naturally

2. **Defensive commenting**
   - Don't add comments because you're unsure if code is clear
   - Refactor until it IS clear

3. **Apologetic comments**
   - Don't write "// Hack:" or "// TODO: This is ugly"
   - Fix it now or create a proper TODO with a ticket reference

4. **Change log comments**
   - Don't comment why code changed - that's what git history is for
   - Comments should explain current state, not history

---

## Examples: Before and After

### Example 1: Obvious Variable Declarations

```go
// BEFORE: Unnecessary comments
// Color definitions for terminal output
cyan := color.New(color.FgCyan).SprintFunc()
yellow := color.New(color.FgYellow).SprintFunc()
green := color.New(color.FgGreen).SprintFunc()

// AFTER: Clear without comments
cyan := color.New(color.FgCyan).SprintFunc()
yellow := color.New(color.FgYellow).SprintFunc()
green := color.New(color.FgGreen).SprintFunc()
```

### Example 2: Self-Evident Control Flow

```go
// BEFORE: Comments state the obvious
// Walk up the directory tree
for {
    // Check if current directory name is a component type
    isComponentType := false
    for _, ct := range componentTypes {
        if dirName == ct {
            isComponentType = true
            break
        }
    }
    
    // If not a component type name, use it
    if !isComponentType && dirName != "." && dirName != "" {
        return dirName
    }
    
    // Go up one directory
    parentDir := filepath.Dir(currentDir)
    
    // Check if we've reached the root
    if parentDir == currentDir {
        return "root"
    }
}

// AFTER: Clear code needs no comments
for {
    isComponentType := false
    for _, ct := range componentTypes {
        if dirName == ct {
            isComponentType = true
            break
        }
    }
    
    if !isComponentType && dirName != "." && dirName != "" {
        return dirName
    }
    
    parentDir := filepath.Dir(currentDir)
    
    if parentDir == currentDir {
        return "root"
    }
}
```

### Example 3: Redundant Godoc

```go
// BEFORE: Adds no value
// NewComponentLinker creates a new ComponentLinker with dependencies injected
func NewComponentLinker() (*linker.ComponentLinker, error) { ... }

// AFTER: Skip obvious godoc
func NewComponentLinker() (*linker.ComponentLinker, error) { ... }
```

### Example 4: Valuable Comment (Keep This!)

```go
// GOOD: Explains non-obvious behavior
// determineDestinationFolderName uses a hierarchy heuristic to preserve
// optional directory structure. It walks up from the component file,
// skipping component-type names (agents/commands/skills), and returns
// the first non-component-type directory name found.
func determineDestinationFolderName(componentFilePath string) string { ... }
```

---

## Summary: The Golden Rule

**If you're about to add a comment, first try:**

1. Renaming variables/functions to be more descriptive
2. Extracting complex logic to a well-named function
3. Simplifying the code structure
4. Restructuring the control flow

**Only add the comment if none of those solve the clarity problem.**

---

## Questions or Exceptions?

If you encounter a situation not covered by these guidelines:

1. Look for similar patterns in the existing codebase
2. Favor clarity and simplicity
3. When in doubt, less is more - err on the side of fewer comments
4. Trust that future developers (human or AI) can read the code

**Remember: Comments are a last resort, not a first instinct.**
