# Styles Package

The `styles` package provides common style patterns for consistent formatting across the application. It centralizes formatting logic to ensure consistency and make maintenance easier.

## Purpose

This package extracts common formatting patterns that were previously duplicated throughout the codebase. By centralizing these patterns:

1. **Consistency**: All parts of the application use the same formatting conventions
2. **Maintainability**: Changing a format pattern only requires updating one location
3. **Readability**: Code becomes more readable by using descriptive function names instead of inline formatting
4. **Testability**: Formatting logic can be tested independently

## Key Features

### Status Messages
- `StatusFailedFormat()` - Formats failure status with error symbol
- `StatusUpToDateFormat()` - Formats "up to date" status with success symbol
- `StatusUpdatingFormat()` - Formats "updating" status with updating symbol
- `StatusUpdatedSuccessfullyFormat()` - Formats success status

### Progress Indicators
- `ProgressCheckingFormat(type, name)` - Formats "Checking..." messages
- `ComponentProgressFormat(current, total, type, name)` - Formats "[1/5] type/name..."

### Inline Messages
- `InlineSuccessFormat(operation, type, name)` - "Operation type: name... ✓ Done"
- `InlineSuccessWithNoteFormat(operation, type, name, note)` - With additional note
- `InlineFailedFormat(operation, type, name)` - "Operation type: name... ✗ Failed"

### Indented Details
- `IndentedErrorFormat(message)` - "  └─ error message"
- `IndentedDetailFormat(key, value)` - "  → key: value"

### Summary Tables
- `SummaryTableFormat(title, width)` - Creates a builder for formatted summary tables
- Tables use box-drawing characters for professional appearance

### Helper Functions
- `InfoArrowFormat(message)` - Formats info messages with arrow
- `ProfileNoteFormat(profileName)` - Formats profile notes (only if not "base")
- `StatusSymbol(success)` - Returns colored symbol based on success/failure
- `CounterRowFormat(symbol, label, count)` - Formats counter rows for summaries
- `ProgressMessage(action, type, name, status)` - Formats progress messages for operations
- `SummaryStats(success, skipped, failed)` - Formats summary statistics with colored symbols
- `ComponentCount(type, count)` - Formats component counts with proper pluralization
- `CommandHint(command, description)` - Formats command hints for "Next steps" sections

## Usage Examples

### Basic Status Messages

```go
import "github.com/tgaines/agent-smith/pkg/styles"

// Check for updates
fmt.Print(styles.ProgressCheckingFormat("skills", "api-design"))
// Output: Checking skills/api-design... 

// Show status
fmt.Printf("%s\n", styles.StatusUpToDateFormat())
// Output: ✓ Up to date

fmt.Printf("%s\n", styles.StatusFailedFormat())
// Output: ✗ Failed
```

### Progress with Counter

```go
// Show progress for component 3 out of 10
fmt.Print(styles.ComponentProgressFormat(3, 10, "agents", "coder"))
// Output: [3/10] agents/coder... 
```

### Inline Success/Failure

```go
// Success with profile note
profileNote := styles.ProfileNoteFormat("dev")
fmt.Printf("%s%s\n", styles.InlineSuccessFormat("Linking", "skill", "api-design"), profileNote)
// Output: Linking skill: api-design... ✓ Done (from profile: dev)

// Failure
fmt.Printf("%s\n", styles.InlineFailedFormat("Linking", "agent", "coder"))
// Output: Linking agent: coder... ✗ Failed
```

### Indented Details

```go
// Show error details
fmt.Printf("%s\n", styles.IndentedErrorFormat("File not found"))
// Output:   └─ File not found

// Show key-value details
fmt.Printf("%s\n", styles.IndentedDetailFormat("opencode", "/path/to/component"))
// Output:   → opencode: /path/to/component
```

### Progress Messages

```go
// Format a progress message
msg := styles.ProgressMessage("Linking", "skill", "api-design", colors.Success(formatter.SymbolSuccess+" Done"))
fmt.Println(msg)
// Output: Linking skill: api-design... ✓ Done
```

### Summary Statistics

```go
// Show operation summary
summary := styles.SummaryStats(5, 2, 1)
fmt.Println(summary)
// Output: ✓ 5 successful, ⚠ 2 skipped, ✗ 1 failed

// No operations
summary := styles.SummaryStats(0, 0, 0)
fmt.Println(summary)
// Output: No operations performed
```

### Component Counts

```go
// Proper pluralization
fmt.Println(styles.ComponentCount("agent", 1))   // Output: 1 agent
fmt.Println(styles.ComponentCount("agent", 5))   // Output: 5 agents
fmt.Println(styles.ComponentCount("skill", 3))   // Output: 3 skills
fmt.Println(styles.ComponentCount("command", 0)) // Output: 0 commands
```

### Command Hints

```go
// Format a command hint for "Next steps"
hint := styles.CommandHint("agent-smith link", "Link components to targets")
fmt.Println(hint)
// Output:   • agent-smith link - Link components to targets (command in cyan)
```

### Summary Tables

```go
import (
    "fmt"
    "github.com/tgaines/agent-smith/internal/formatter"
    "github.com/tgaines/agent-smith/pkg/colors"
    "github.com/tgaines/agent-smith/pkg/styles"
)

// Build a summary table
table := styles.SummaryTableFormat("Update Summary", 80)
table.AddRow("Total components checked:", 10)
table.AddRowWithSymbol(colors.Success(formatter.SymbolSuccess), "Already up to date:", 8)
table.AddRowWithSymbol(colors.Error(formatter.SymbolError), "Failed:", 2)

fmt.Println(table.Build())
```

Output:
```
┌──────────────────────────────────────────────────────────────────────────────┐
│                            Update Summary                                    │
├──────────────────────────────────────────────────────────────────────────────┤
│ Total components checked:      10                                            │
│ ✓ Already up to date:          8                                             │
│ ✗ Failed:                      2                                             │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Before and After

### Before (without styles package)

```go
fmt.Printf("Checking %s/%s... ", colors.Muted(componentType), componentName)

if err != nil {
    fmt.Printf("%s Failed\n", colors.Error(formatter.SymbolError))
    fmt.Printf("  %s %v\n\n", colors.Muted("└─"), err)
    return err
}

fmt.Printf("%s Up to date\n\n", colors.Success(formatter.SymbolSuccess))
```

### After (with styles package)

```go
fmt.Print(styles.ProgressCheckingFormat(componentType, componentName))

if err != nil {
    fmt.Printf("%s\n", styles.StatusFailedFormat())
    fmt.Printf("%s\n\n", styles.IndentedErrorFormat(err.Error()))
    return err
}

fmt.Printf("%s\n\n", styles.StatusUpToDateFormat())
```

## Design Principles

1. **Descriptive Names**: Function names clearly describe what they format
2. **Composability**: Functions can be combined to create complex output
3. **Color Integration**: Works seamlessly with the `colors` package
4. **Symbol Integration**: Uses symbols from the `formatter` package
5. **Testability**: All functions are unit tested with color disabled for predictable output

## Testing

The package includes comprehensive unit tests that verify:
- Correct formatting of all patterns
- Proper color integration (tested with colors disabled)
- Edge cases (empty strings, special characters, etc.)
- Table building and rendering

Run tests:
```bash
go test ./pkg/styles/...
```

## Related Packages

- **colors**: Provides color functions (Success, Error, Warning, Info, Muted, etc.)
- **formatter**: Provides symbols (SymbolSuccess, SymbolError, etc.) and Formatter methods
- **internal/formatter**: Provides box-drawing utilities and table formatting

## Future Enhancements

Potential additions to consider:

1. **List Formatting**: Standardized bullet point lists
2. **Tree Formatting**: Hierarchical tree structures
3. **Diff Formatting**: Show before/after comparisons
4. **Progress Bars**: Animated progress indicators
5. **Spinner Integration**: Loading spinners for long operations

## Contributing

When adding new formatting patterns:

1. Add the function to `styles.go`
2. Add comprehensive tests to `styles_test.go`
3. Update this README with examples
4. Use descriptive function names that indicate what they format
5. Ensure compatibility with color enable/disable
