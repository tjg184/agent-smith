// Package styles provides common style patterns for consistent formatting
// across the application. This centralizes formatting logic to ensure
// consistency and make maintenance easier.
package styles

import (
	"fmt"
	"strings"

	"github.com/tgaines/agent-smith/internal/formatter"
	"github.com/tgaines/agent-smith/pkg/colors"
)

// ProgressCheckingFormat returns a formatted "Checking..." message
// Format: "Checking {type}/{name}... "
func ProgressCheckingFormat(componentType, componentName string) string {
	return fmt.Sprintf("Checking %s/%s... ", colors.Muted(componentType), componentName)
}

// StatusFailedFormat returns a formatted failure status
// Format: "✗ Failed"
func StatusFailedFormat() string {
	return fmt.Sprintf("%s Failed", colors.Error(formatter.SymbolError))
}

// StatusUpToDateFormat returns a formatted "up to date" status
// Format: "✓ Up to date"
func StatusUpToDateFormat() string {
	return fmt.Sprintf("%s Up to date", colors.Success(formatter.SymbolSuccess))
}

// StatusUpdatingFormat returns a formatted "updating" status
// Format: "⟳ Updating"
func StatusUpdatingFormat() string {
	return fmt.Sprintf("%s Updating", colors.Warning(formatter.SymbolUpdating))
}

// StatusUpdatedSuccessfullyFormat returns a formatted "updated successfully" status
// Format: "✓ Updated successfully"
func StatusUpdatedSuccessfullyFormat() string {
	return fmt.Sprintf("%s Updated successfully", colors.Success(formatter.SymbolSuccess))
}

// IndentedErrorFormat returns a formatted indented error message
// Format: "  └─ error message"
func IndentedErrorFormat(message string) string {
	return fmt.Sprintf("  %s %s", colors.Muted("└─"), message)
}

// IndentedDetailFormat returns a formatted indented detail line
// Format: "  → key: value"
func IndentedDetailFormat(key, value string) string {
	return fmt.Sprintf("  %s %s: %s", colors.Muted("→"), key, value)
}

// InlineSuccessFormat returns a formatted inline success message
// Format: "Operation {type}: {name}... ✓ Done"
func InlineSuccessFormat(operation, componentType, componentName string) string {
	return fmt.Sprintf("%s %s: %s... %s", operation, componentType, componentName, colors.Success(formatter.SymbolSuccess+" Done"))
}

// InlineSuccessWithNoteFormat returns a formatted inline success message with a note
// Format: "Operation {type}: {name}... ✓ Done (note)"
func InlineSuccessWithNoteFormat(operation, componentType, componentName, note string) string {
	return fmt.Sprintf("%s %s: %s... %s", operation, componentType, componentName,
		colors.Success(formatter.SymbolSuccess+" Done")+colors.Muted(" ("+note+")"))
}

// InlineFailedFormat returns a formatted inline failure message
// Format: "Operation {type}: {name}... ✗ Failed"
func InlineFailedFormat(operation, componentType, componentName string) string {
	return fmt.Sprintf("%s %s: %s... %s", operation, componentType, componentName, colors.Error(formatter.SymbolError+" Failed"))
}

// InfoArrowFormat returns a formatted info message with arrow
// Format: "→ message"
func InfoArrowFormat(message string) string {
	return fmt.Sprintf("%s %s", colors.Info("→"), message)
}

// ComponentProgressFormat returns a formatted progress counter
// Format: "[1/5] type/name... "
func ComponentProgressFormat(current, total int, componentType, componentName string) string {
	return fmt.Sprintf("[%d/%d] %s/%s... ", current, total, colors.Muted(componentType), componentName)
}

// ProfileNoteFormat returns a formatted profile note
// Format: " (from profile: name)"
func ProfileNoteFormat(profileName string) string {
	if profileName == "" || profileName == "base" {
		return ""
	}
	return colors.Muted(fmt.Sprintf(" (from profile: %s)", profileName))
}

// SummaryTableFormat creates a formatted summary table with borders
// Width is the total table width (default: 80)
func SummaryTableFormat(title string, width int) SummaryTableBuilder {
	if width <= 0 {
		width = 80
	}
	return SummaryTableBuilder{
		title: title,
		width: width,
		rows:  []string{},
	}
}

// SummaryTableBuilder helps build formatted summary tables
type SummaryTableBuilder struct {
	title string
	width int
	rows  []string
}

// AddRow adds a row to the summary table
func (b *SummaryTableBuilder) AddRow(label string, value interface{}) *SummaryTableBuilder {
	// Format the complete content
	content := fmt.Sprintf("%s %v", label, value)

	// Calculate visible length and padding (following box_table.go pattern)
	innerWidth := b.width - 2      // Subtract borders
	contentWidth := innerWidth - 2 // Subtract space padding on each side
	visLen := formatter.VisibleLength(content)
	padding := contentWidth - visLen

	// Build row with proper ANSI-aware padding (following boxes.go line 135 pattern)
	row := fmt.Sprintf("│ %s%s │", content, strings.Repeat(" ", padding))
	b.rows = append(b.rows, row)
	return b
}

// AddRowWithSymbol adds a row with a colored symbol
func (b *SummaryTableBuilder) AddRowWithSymbol(symbol, label string, value interface{}) *SummaryTableBuilder {
	// Format the complete content
	content := fmt.Sprintf("%s %s %v", symbol, label, value)

	// Calculate visible length and padding (following box_table.go pattern)
	innerWidth := b.width - 2      // Subtract borders
	contentWidth := innerWidth - 2 // Subtract space padding on each side
	visLen := formatter.VisibleLength(content)
	padding := contentWidth - visLen

	// Build row with proper ANSI-aware padding (following boxes.go line 135 pattern)
	row := fmt.Sprintf("│ %s%s │", content, strings.Repeat(" ", padding))
	b.rows = append(b.rows, row)
	return b
}

// Build constructs the final table string
func (b *SummaryTableBuilder) Build() string {
	var result strings.Builder

	// Top border
	innerWidth := b.width - 2
	result.WriteString("┌")
	result.WriteString(strings.Repeat("─", innerWidth))
	result.WriteString("┐\n")

	// Title row (centered)
	if b.title != "" {
		titlePadding := (innerWidth - len(b.title)) / 2
		remainingPadding := innerWidth - len(b.title) - titlePadding
		result.WriteString("│")
		result.WriteString(strings.Repeat(" ", titlePadding))
		result.WriteString(b.title)
		result.WriteString(strings.Repeat(" ", remainingPadding))
		result.WriteString("│\n")

		// Separator after title
		result.WriteString("├")
		result.WriteString(strings.Repeat("─", innerWidth))
		result.WriteString("┤\n")
	}

	// Rows
	for _, row := range b.rows {
		result.WriteString(row)
		result.WriteString("\n")
	}

	// Bottom border
	result.WriteString("└")
	result.WriteString(strings.Repeat("─", innerWidth))
	result.WriteString("┘")

	return result.String()
}

// StatusSymbol returns a colored status symbol based on the condition
func StatusSymbol(success bool) string {
	if success {
		return colors.Success(formatter.SymbolSuccess)
	}
	return colors.Error(formatter.SymbolError)
}

// CounterRowFormat formats a counter row for summary tables
// Format: "✓ Label: value" or "✗ Label: value"
func CounterRowFormat(symbol, label string, count int) string {
	return fmt.Sprintf("%s %-30s %-45d", symbol, label+":", count)
}

// ProgressMessage formats a progress message for operations
// Format: "{action} {type}: {name}... {status}"
// Example: "Linking skill: api-design... ✓ Done"
func ProgressMessage(action, componentType, componentName, status string) string {
	return fmt.Sprintf("%s %s: %s... %s", action, componentType, componentName, status)
}

// SummaryStats formats summary statistics for operations
// Returns a formatted string with success, skipped, and failed counts
func SummaryStats(success, skipped, failed int) string {
	var parts []string

	if success > 0 {
		parts = append(parts, fmt.Sprintf("%s %d successful", colors.Success(formatter.SymbolSuccess), success))
	}
	if skipped > 0 {
		parts = append(parts, fmt.Sprintf("%s %d skipped", colors.Warning(formatter.SymbolWarning), skipped))
	}
	if failed > 0 {
		parts = append(parts, fmt.Sprintf("%s %d failed", colors.Error(formatter.SymbolError), failed))
	}

	if len(parts) == 0 {
		return "No operations performed"
	}

	return strings.Join(parts, ", ")
}

// ComponentCount formats a component count with proper pluralization
// Format: "X agents" or "1 agent", "Y skills" or "1 skill"
func ComponentCount(componentType string, count int) string {
	if count == 1 {
		// Singular form
		return fmt.Sprintf("%d %s", count, componentType)
	}
	// Plural form - add 's' if not already plural
	plural := componentType
	if !strings.HasSuffix(componentType, "s") {
		plural = componentType + "s"
	}
	return fmt.Sprintf("%d %s", count, plural)
}

// CommandHint formats a command hint with description
// Format: "• command - description" (command in cyan)
func CommandHint(command, description string) string {
	return fmt.Sprintf("  %s %s - %s", colors.Muted("•"), colors.Info(command), description)
}
