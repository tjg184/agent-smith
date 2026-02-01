# Formatter Package - Visual Hierarchy Guide

This package provides consistent visual hierarchy utilities for all CLI commands in agent-smith.

## Overview

The formatter package ensures a consistent user experience across all CLI output by providing standardized formatting methods for:
- Section headers and subsections
- Success, error, and warning messages
- Progress indicators
- List and detail formatting
- Summary displays

## Usage

### Basic Setup

```go
import "github.com/tgaines/agent-smith/internal/formatter"

// Create a formatter instance
f := formatter.New()
```

### Section Headers

Use section headers to organize output into clear, scannable sections:

```go
// Main section header: === Title ===
f.SectionHeader("Installation Summary")

// Subsection header: --- Title ---
f.SubsectionHeader("Skills")
```

**Output:**
```
=== Installation Summary ===

--- Skills ---
```

### Status Messages

Display operation results with consistent symbols and colors:

```go
// Success message (green ✓)
f.SuccessMsg("Successfully installed component: %s", name)

// Error message (red ✗)
f.ErrorMsg("Failed to link component: %s", err)

// Warning message (yellow ⚠️)
f.WarningMsg("Component already exists, skipping")

// Info message (• bullet)
f.InfoMsg("Checking for updates...")
```

**Output:**
```
✓ Successfully installed component: api-design
✗ Failed to link component: permission denied
⚠️ Component already exists, skipping
• Checking for updates...
```

### Progress Indicators

Show real-time progress for long-running operations:

```go
f.ProgressMsg("Installing", "api-design")
// ... do work ...
f.ProgressComplete() // or f.ProgressFailed()
```

**Output:**
```
Installing: api-design... ✓ Done
```

### Lists and Details

Create hierarchical information displays:

```go
// List items (2-space indent)
f.ListItem("api-design → /path/to/target")
f.ListItem("code-review → /path/to/target")

// Detail items (4-space indent, dimmed key)
f.DetailItem("Source", "/home/user/.agent-smith/skills")
f.DetailItem("Target", "/path/to/opencode")
```

**Output:**
```
  • api-design → /path/to/target
  • code-review → /path/to/target
    Source: /home/user/.agent-smith/skills
    Target: /path/to/opencode
```

### Summary Displays

Display operation summaries with counters:

```go
f.SectionHeader("Summary")
f.CounterSummary(10, 7, 2, 1)
```

**Output:**
```
=== Summary ===
Total: 10
✓ Successful: 7
✗ Failed: 2
⚠️ Skipped: 1
```

### Spacing

Add empty lines for better visual separation:

```go
f.EmptyLine()
```

## Complete Example

Here's a complete example showing how to format a typical CLI command output:

```go
func linkCommand() {
    f := formatter.New()
    
    f.SectionHeader("Linking Components")
    
    // Progress for each component
    for _, component := range components {
        f.ProgressMsg("Linking", component.Name)
        
        if err := linkComponent(component); err != nil {
            f.ProgressFailed()
            f.DetailItem("Error", err.Error())
        } else {
            f.ProgressComplete()
            f.DetailItem("Path", component.Path)
        }
    }
    
    // Summary
    f.EmptyLine()
    f.SectionHeader("Summary")
    f.CounterSummary(total, success, failed, skipped)
}
```

**Output:**
```
=== Linking Components ===
Linking: api-design... ✓ Done
    Path: /path/to/opencode/skills/api-design
Linking: code-review... ✗ FAILED
    Error: permission denied

=== Summary ===
Total: 2
✓ Successful: 1
✗ Failed: 1
```

## Symbol Constants

The package provides these standard symbols:

```go
formatter.SymbolSuccess   // ✓
formatter.SymbolError     // ✗
formatter.SymbolWarning   // ⚠️
formatter.SymbolCopied    // ◆
formatter.SymbolNotLinked // -
formatter.SymbolUnknown   // ?
```

## Color Helpers

Get colored symbols for custom formatting:

```go
formatter.ColoredSuccess()  // Green ✓
formatter.ColoredError()    // Red ✗
formatter.ColoredWarning()  // Yellow ⚠️
```

## Box Tables

For structured tabular data, use box tables:

```go
table := formatter.NewBoxTable(os.Stdout, []string{"Name", "Status", "Path"})
table.AddRow([]string{"api-design", "✓", "/path/to/target"})
table.AddRow([]string{"code-review", "✗", "broken link"})
table.Render()
```

**Output:**
```
┌────────────┬────────┬─────────────────┐
│ Name       │ Status │ Path            │
├────────────┼────────┼─────────────────┤
│ api-design │ ✓      │ /path/to/target │
│ code-review│ ✗      │ broken link     │
└────────────┴────────┴─────────────────┘
```

## Migration Guide

### Before (inconsistent):
```go
fmt.Printf("\n=== %s ===\n", title)
fmt.Printf("✓ Successfully installed %s\n", name)
fmt.Printf("Error: %v\n", err)
```

### After (consistent):
```go
f := formatter.New()
f.SectionHeader(title)
f.SuccessMsg("Successfully installed %s", name)
f.ErrorMsg("Installation failed: %v", err)
```

## Benefits

1. **Consistency**: All commands use the same visual language
2. **Scannability**: Headers and symbols make output easy to scan
3. **Accessibility**: Consistent colors and symbols aid comprehension
4. **Maintainability**: Centralized formatting logic
5. **Testability**: Writer injection enables easy testing

## Testing

The formatter supports custom writers for testing:

```go
buf := &bytes.Buffer{}
f := formatter.NewWithWriter(buf)
f.SuccessMsg("test")
// Assert on buf.String()
```
