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
	// Calculate padding to align values (keeping space for borders and padding)
	labelWidth := 30
	valueWidth := b.width - labelWidth - 6 // 6 = borders (2) + padding (4)

	row := fmt.Sprintf("│ %-*s %-*v│", labelWidth, label, valueWidth, value)
	b.rows = append(b.rows, row)
	return b
}

// AddRowWithSymbol adds a row with a colored symbol
func (b *SummaryTableBuilder) AddRowWithSymbol(symbol, label string, value interface{}) *SummaryTableBuilder {
	// Calculate padding to align values
	labelWidth := 30
	valueWidth := b.width - labelWidth - 6 // 6 = borders (2) + padding (4)

	labelWithSymbol := fmt.Sprintf("%s %s", symbol, label)
	row := fmt.Sprintf("│ %-*s %-*v│", labelWidth, labelWithSymbol, valueWidth, value)
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
