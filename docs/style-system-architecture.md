# Style System Architecture

This document describes the style system architecture in agent-smith, which provides consistent and maintainable formatting across the application.

## Overview

The style system consists of three main layers:

```
┌─────────────────────────────────────────────────────┐
│                  Application Code                   │
│         (updater, linker, commands, etc.)          │
└─────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────┐
│              pkg/styles (Layer 3)                   │
│        Common Formatting Patterns                   │
│  (High-level functions like StatusFormat,          │
│   ProgressFormat, InlineFormat, etc.)              │
└─────────────────────────────────────────────────────┘
                         │
                ┌────────┴────────┐
                ▼                 ▼
┌──────────────────────┐  ┌──────────────────────┐
│  pkg/colors (Layer 2)│  │internal/formatter    │
│  Color Management    │  │  (Layer 2)           │
│  (Success, Error,    │  │  Symbols, Tables,    │
│   Warning, Info)     │  │  Box Drawing         │
└──────────────────────┘  └──────────────────────┘
                │                 │
                └────────┬────────┘
                         ▼
┌─────────────────────────────────────────────────────┐
│         External Libraries (Layer 1)                │
│     (fatih/color, go-isatty, unicode)              │
└─────────────────────────────────────────────────────┘
```

## Components

### Layer 1: External Libraries

**Purpose**: Foundation libraries for terminal capabilities

- **fatih/color**: ANSI color support
- **go-isatty**: TTY detection for auto-disabling colors in non-TTY contexts
- **unicode**: Box-drawing characters

### Layer 2: Building Blocks

#### pkg/colors

**Purpose**: Centralized color management with TTY detection

**Key Features**:
- Auto-detects TTY to enable/disable colors appropriately
- Respects `NO_COLOR` and `FORCE_COLOR` environment variables
- Provides semantic color functions (Success, Error, Warning, Info, Muted)
- Consistent color palette across the application

**Example**:
```go
colors.Success("✓")  // Green checkmark
colors.Error("✗")    // Red X
colors.Warning("⚠")  // Yellow warning
colors.Info("→")     // Cyan arrow
colors.Muted("detail") // Gray/dim text
```

#### internal/formatter

**Purpose**: Low-level formatting utilities

**Key Components**:

1. **Symbols** (`symbols.go`):
   - `SymbolSuccess = "✓"`
   - `SymbolError = "✗"`
   - `SymbolWarning = "⚠️"`
   - `SymbolUpdating = "⟳"`
   - etc.

2. **Box Drawing** (`boxes.go`, `box_table.go`):
   - Box-drawing characters (┌, ─, │, ┐, etc.)
   - Table rendering with headers and rows
   - Multi-section boxes

3. **Formatter Methods** (`formatter.go`):
   - `SuccessMsg()`, `ErrorMsg()`, `WarningMsg()`
   - `SectionHeader()`, `SubsectionHeader()`
   - `ProgressMsg()`, `ProgressComplete()`, `ProgressFailed()`

### Layer 3: Common Patterns

#### pkg/styles

**Purpose**: High-level, reusable formatting patterns

This is the **newest addition** that centralizes common formatting patterns that were previously duplicated across the codebase.

**Key Pattern Categories**:

1. **Status Messages**:
   - `StatusFailedFormat()`
   - `StatusUpToDateFormat()`
   - `StatusUpdatingFormat()`

2. **Progress Indicators**:
   - `ProgressCheckingFormat(type, name)`
   - `ComponentProgressFormat(current, total, type, name)`

3. **Inline Messages**:
   - `InlineSuccessFormat(operation, type, name)`
   - `InlineFailedFormat(operation, type, name)`

4. **Indented Details**:
   - `IndentedErrorFormat(message)`
   - `IndentedDetailFormat(key, value)`

5. **Summary Tables**:
   - `SummaryTableFormat(title, width)` - Builder pattern for tables

6. **Helper Functions**:
   - `InfoArrowFormat(message)`
   - `ProfileNoteFormat(profileName)`
   - `StatusSymbol(success)`

## Usage Guidelines

### When to Use Each Layer

#### Use pkg/styles for:
✅ Common formatting patterns used in multiple places  
✅ Status messages, progress indicators, inline success/failure  
✅ Indented details, summary tables  
✅ Profile notes, component progress counters

#### Use pkg/colors for:
✅ Direct color application when you need fine control  
✅ New patterns not yet in styles package  
✅ Low-level color manipulation

#### Use internal/formatter for:
✅ Box drawing and table rendering  
✅ Symbols (SymbolSuccess, SymbolError, etc.)  
✅ Formatter instance methods when you need writer control

#### Use external libraries directly:
❌ Avoid - use the wrapper packages instead

### Migration Path

When you find duplicated formatting code:

1. **Identify the pattern**: What is being formatted? Is it a status, progress, detail line?
2. **Check pkg/styles**: Does a function already exist for this pattern?
3. **Use existing function**: Replace inline formatting with styles function
4. **Add new function**: If no function exists and the pattern is used in 2+ places, add it to pkg/styles

## Example: Refactoring Common Patterns

### Before Refactoring

**In updater.go**:
```go
fmt.Printf("Checking %s/%s... ", colors.Muted(componentType), componentName)
// ... do work ...
fmt.Printf("%s Failed\n", colors.Error(formatter.SymbolError))
fmt.Printf("  %s %v\n", colors.Muted("└─"), err)
```

**In linker.go**:
```go
fmt.Printf("Checking %s/%s... ", colors.Muted(componentType), componentName)
// ... do work ...
fmt.Printf("%s Failed\n", colors.Error(formatter.SymbolError))
fmt.Printf("  %s %v\n", colors.Muted("→"), err)
```

**Problem**: Same pattern duplicated in multiple files with slight variations

### After Refactoring

**In pkg/styles/styles.go**:
```go
func ProgressCheckingFormat(componentType, componentName string) string {
    return fmt.Sprintf("Checking %s/%s... ", colors.Muted(componentType), componentName)
}

func StatusFailedFormat() string {
    return fmt.Sprintf("%s Failed", colors.Error(formatter.SymbolError))
}

func IndentedErrorFormat(message string) string {
    return fmt.Sprintf("  %s %s", colors.Muted("└─"), message)
}
```

**In updater.go and linker.go**:
```go
fmt.Print(styles.ProgressCheckingFormat(componentType, componentName))
// ... do work ...
fmt.Printf("%s\n", styles.StatusFailedFormat())
fmt.Printf("%s\n", styles.IndentedErrorFormat(err.Error()))
```

**Benefits**:
- ✅ Pattern defined once, used everywhere
- ✅ Consistent formatting (both files use same detail arrow)
- ✅ Easy to change (modify one function to update all usages)
- ✅ More readable (descriptive function names)

## Testing Strategy

### Unit Tests

Each layer has comprehensive unit tests:

1. **pkg/colors**: Tests color enable/disable, TTY detection
2. **internal/formatter**: Tests box drawing, table rendering, symbols
3. **pkg/styles**: Tests all pattern functions with colors disabled

### Integration Tests

Application code tests verify:
- Correct usage of styles patterns
- Output formatting in real scenarios
- Color behavior in TTY and non-TTY contexts

## Best Practices

### 1. Use Descriptive Function Names

✅ **Good**: `StatusUpToDateFormat()`  
❌ **Bad**: `Status1()` or `UpToDate()`

### 2. Return Strings, Don't Print

✅ **Good**: `return fmt.Sprintf(...)`  
❌ **Bad**: `fmt.Printf(...)`

This allows callers to compose output and control printing.

### 3. Keep Functions Focused

Each function should format one specific pattern. Don't create mega-functions that try to do everything.

### 4. Document with Examples

Every function should have a comment showing example output:

```go
// StatusUpToDateFormat returns a formatted "up to date" status
// Format: "✓ Up to date"
func StatusUpToDateFormat() string {
    // ...
}
```

### 5. Test with Colors Disabled

Always test formatting functions with colors disabled to ensure predictable output:

```go
func TestStatusUpToDateFormat(t *testing.T) {
    colors.Disable()
    defer colors.Enable()
    
    result := StatusUpToDateFormat()
    expected := "✓ Up to date"
    if result != expected {
        t.Errorf("got %q, want %q", result, expected)
    }
}
```

## Future Improvements

### Potential Enhancements

1. **Context-aware formatting**: Adapt formatting based on terminal width
2. **Localization support**: Translate messages while keeping formatting
3. **Theme system**: Allow users to customize colors and symbols
4. **Animation support**: Progress bars, spinners
5. **Structured logging integration**: Output to structured log formats

### Planned Additions to pkg/styles

- List formatting (bullet points, numbered lists)
- Tree formatting (directory structures)
- Diff formatting (before/after comparisons)
- Multi-column layouts
- Error with stack trace formatting

## Migration Progress

### Completed

✅ Created pkg/styles package with common patterns  
✅ Updated internal/updater to use styles  
✅ Updated internal/linker to use styles  
✅ Comprehensive unit tests for all patterns  
✅ Documentation and examples

### In Progress

🔄 Gradually migrate other modules to use pkg/styles

### Future

- [ ] Migrate downloader package
- [ ] Migrate command handlers
- [ ] Add more patterns as duplicates are identified
- [ ] Create style guide document

## Conclusion

The style system provides a robust, maintainable foundation for consistent formatting across agent-smith. By centralizing common patterns in pkg/styles, we:

1. Reduce code duplication
2. Ensure consistent formatting
3. Make it easier to change formatting globally
4. Improve code readability
5. Enable better testing

As you work on the codebase, look for opportunities to use existing patterns or add new ones to pkg/styles when you find duplication.
