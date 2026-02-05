package formatter

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/tgaines/agent-smith/pkg/colors"
)

// Formatter handles all output formatting for the application
type Formatter struct {
	writer io.Writer
}

// New creates a new Formatter instance that writes to stdout
func New() *Formatter {
	return &Formatter{
		writer: os.Stdout,
	}
}

// NewWithWriter creates a new Formatter instance with a custom writer (for testing)
func NewWithWriter(w io.Writer) *Formatter {
	return &Formatter{
		writer: w,
	}
}

// Writer returns the underlying writer used by this formatter
func (f *Formatter) Writer() io.Writer {
	return f.writer
}

// Success prints a success message for a component installation
func (f *Formatter) Success(componentType, name string) {
	fmt.Fprintf(f.writer, "%s Installed %s: %s\n", colors.Success(SymbolSuccess), componentType, name)
}

// Error prints an error message
func (f *Formatter) Error(message string, err error) {
	if err != nil {
		fmt.Fprintf(f.writer, "%s %s: %v\n", SymbolError, message, err)
	} else {
		fmt.Fprintf(f.writer, "%s %s\n", SymbolError, message)
	}
}

// Warning prints a warning message (uses log.Printf to maintain existing behavior)
func (f *Formatter) Warning(message string, args ...interface{}) {
	log.Printf("Warning: "+message, args...)
}

// Info prints an informational message
func (f *Formatter) Info(message string, args ...interface{}) {
	fmt.Fprintf(f.writer, message+"\n", args...)
}

// ColoredSuccess returns a green-colored success symbol
func ColoredSuccess() string {
	return colors.Success(SymbolSuccess)
}

// ColoredError returns a red-colored error symbol
func ColoredError() string {
	return colors.Error(SymbolError)
}

// ColoredWarning returns a yellow-colored warning symbol
func ColoredWarning() string {
	return colors.Warning(SymbolWarning)
}

// SectionHeader prints a section header with consistent formatting
// Example: === Section Title ===
func (f *Formatter) SectionHeader(title string) {
	fmt.Fprintf(f.writer, "\n%s\n", colors.InfoBold("=== "+title+" ==="))
}

// SubsectionHeader prints a subsection header with consistent formatting
// Example: --- Subsection Title ---
func (f *Formatter) SubsectionHeader(title string) {
	fmt.Fprintf(f.writer, "\n%s\n", colors.Highlight("--- "+title+" ---"))
}

// ProgressMsg prints a progress message for ongoing operations
// Example: "Linking skill: api-design... "
func (f *Formatter) ProgressMsg(operation, item string) {
	fmt.Fprintf(f.writer, "%s: %s... ", operation, item)
}

// ProgressComplete prints a completion marker (checkmark) for progress
func (f *Formatter) ProgressComplete() {
	fmt.Fprintf(f.writer, "%s Done\n", colors.Success(SymbolSuccess))
}

// ProgressFailed prints a failure marker for progress
func (f *Formatter) ProgressFailed() {
	fmt.Fprintf(f.writer, "%s FAILED\n", colors.Error(SymbolError))
}

// SuccessMsg prints a success message with green checkmark
func (f *Formatter) SuccessMsg(message string, args ...interface{}) {
	msg := fmt.Sprintf(message, args...)
	fmt.Fprintf(f.writer, "%s %s\n", colors.Success(SymbolSuccess), msg)
}

// ErrorMsg prints an error message with red X symbol
func (f *Formatter) ErrorMsg(message string, args ...interface{}) {
	msg := fmt.Sprintf(message, args...)
	fmt.Fprintf(f.writer, "%s %s\n", colors.Error(SymbolError), msg)
}

// WarningMsg prints a warning message with yellow warning symbol
func (f *Formatter) WarningMsg(message string, args ...interface{}) {
	msg := fmt.Sprintf(message, args...)
	fmt.Fprintf(f.writer, "%s %s\n", colors.Warning(SymbolWarning), msg)
}

// InfoMsg prints an informational message with bullet point
func (f *Formatter) InfoMsg(message string, args ...interface{}) {
	msg := fmt.Sprintf(message, args...)
	fmt.Fprintf(f.writer, "• %s\n", msg)
}

// ListItem prints an indented list item
func (f *Formatter) ListItem(item string, args ...interface{}) {
	msg := fmt.Sprintf(item, args...)
	fmt.Fprintf(f.writer, "  • %s\n", msg)
}

// DetailItem prints a detail line with indentation
func (f *Formatter) DetailItem(key, value string) {
	fmt.Fprintf(f.writer, "    %s: %s\n", colors.Muted(key), value)
}

// EmptyLine prints an empty line for spacing
func (f *Formatter) EmptyLine() {
	fmt.Fprintln(f.writer)
}

// Summary prints a summary section with key-value pairs
func (f *Formatter) Summary(title string, items map[string]interface{}) {
	f.SectionHeader(title)
	for key, value := range items {
		fmt.Fprintf(f.writer, "%s: %v\n", key, value)
	}
}

// CounterSummary prints a summary with counters using box-drawing table
func (f *Formatter) CounterSummary(total, success, failed, skipped int) {
	// Create summary table with box-drawing characters
	table := NewBoxTable(f.writer, []string{"Status", "Count"})

	// Add total row (always shown)
	table.AddRow([]string{"Total", fmt.Sprintf("%d", total)})

	// Add success row if there are successes
	if success > 0 {
		table.AddRow([]string{colors.Success(SymbolSuccess + " Successful"), fmt.Sprintf("%d", success)})
	}

	// Add failed row if there are failures
	if failed > 0 {
		table.AddRow([]string{colors.Error(SymbolError + " Failed"), fmt.Sprintf("%d", failed)})
	}

	// Add skipped row if there are skipped items
	if skipped > 0 {
		table.AddRow([]string{colors.Warning(SymbolWarning + " Skipped"), fmt.Sprintf("%d", skipped)})
	}

	table.Render()
}

// InlineSuccess prints an inline success message
// Format: "Operation: item... ✓ Done"
func (f *Formatter) InlineSuccess(operation, item string) {
	fmt.Fprintf(f.writer, "%s: %s... %s\n", operation, item, colors.Success(SymbolSuccess+" Done"))
}

// InlineSuccessWithNote prints an inline success message with an additional note
// Format: "Operation: item... ✓ Done (note)"
func (f *Formatter) InlineSuccessWithNote(operation, item, note string) {
	fmt.Fprintf(f.writer, "%s: %s... %s\n", operation, item, colors.Success(SymbolSuccess+" Done")+colors.Muted(" ("+note+")"))
}

// InlineFailed prints an inline failure message
// Format: "Operation: item... ✗ Failed"
func (f *Formatter) InlineFailed(operation, item string) {
	fmt.Fprintf(f.writer, "%s: %s... %s\n", operation, item, colors.Error(SymbolError+" Failed"))
}

// StatusSuccess prints a standalone success status message
// Format: "✓ Successfully completed operation"
func (f *Formatter) StatusSuccess(message string, args ...interface{}) {
	msg := fmt.Sprintf(message, args...)
	fmt.Fprintf(f.writer, "%s %s\n", colors.Success(SymbolSuccess), msg)
}

// StatusError prints a standalone error status message
// Format: "✗ Error: something went wrong"
func (f *Formatter) StatusError(message string, args ...interface{}) {
	msg := fmt.Sprintf(message, args...)
	fmt.Fprintf(f.writer, "%s %s\n", colors.Error(SymbolError), msg)
}

// StatusUpToDate prints an "up to date" status message
// Format: "✓ Up to date"
func (f *Formatter) StatusUpToDate() {
	fmt.Fprintf(f.writer, "%s Up to date\n", colors.Success(SymbolSuccess))
}

// StatusUpdating prints an "updating" status message
// Format: "⟳ Updating"
func (f *Formatter) StatusUpdating() {
	fmt.Fprintf(f.writer, "%s Updating\n", colors.Warning(SymbolUpdating))
}

// IndentedDetail prints an indented detail line
// Format: "  → key: value"
func (f *Formatter) IndentedDetail(key, value string) {
	fmt.Fprintf(f.writer, "  %s %s: %s\n", colors.Muted("→"), key, value)
}

// IndentedError prints an indented error message
// Format: "  ✗ error message"
func (f *Formatter) IndentedError(message string, args ...interface{}) {
	msg := fmt.Sprintf(message, args...)
	fmt.Fprintf(f.writer, "  %s %s\n", colors.Error(SymbolError), msg)
}

// IndentedSuccess prints an indented success message
// Format: "  ✓ success message"
func (f *Formatter) IndentedSuccess(message string, args ...interface{}) {
	msg := fmt.Sprintf(message, args...)
	fmt.Fprintf(f.writer, "  %s %s\n", colors.Success(SymbolSuccess), msg)
}

// PlainWarning prints a plain warning message with consistent prefix
// Format: "Warning: message"
func (f *Formatter) PlainWarning(message string, args ...interface{}) {
	msg := fmt.Sprintf(message, args...)
	fmt.Fprintf(f.writer, "Warning: %s\n", msg)
}

// SuccessWithDetail prints a success message with additional detail information
// Format: "✓ Successfully {message}"
//
//	"  → detail"
func (f *Formatter) SuccessWithDetail(componentType, name, detail string) {
	fmt.Fprintf(f.writer, "%s Successfully %s: %s\n", colors.Success(SymbolSuccess), componentType, name)
	if detail != "" {
		fmt.Fprintf(f.writer, "  %s %s\n", colors.Muted("→"), detail)
	}
}

// ErrorWithContext prints an error message with context and optional suggestion
// Format: "✗ Error: message"
//
//	"  └─ error details"
//	"  Try: suggestion"
func (f *Formatter) ErrorWithContext(message string, err error, suggestion string) {
	fmt.Fprintf(f.writer, "%s %s\n", colors.Error(SymbolError), colors.Error(message))
	if err != nil {
		fmt.Fprintf(f.writer, "  %s %v\n", colors.Muted("└─"), err)
	}
	if suggestion != "" {
		fmt.Fprintf(f.writer, "  Try: %s\n", suggestion)
	}
}

// Section prints a section header with consistent formatting
// Format: "Section Title"
//
//	"─────────────" (underline)
func (f *Formatter) Section(title string) {
	fmt.Fprintf(f.writer, "\n%s\n", colors.InfoBold(title))
	fmt.Fprintf(f.writer, "%s\n", colors.Muted(strings.Repeat(BoxHorizontal, len(title))))
}

// Divider prints a visual separator line
// Format: "────────────────────────────────────────"
func (f *Formatter) Divider() {
	fmt.Fprintf(f.writer, "%s\n", colors.Muted(strings.Repeat(BoxHorizontal, 40)))
}

// KeyValue prints a key-value pair with consistent alignment
// Format: "Key:     value"
func (f *Formatter) KeyValue(key, value string) {
	fmt.Fprintf(f.writer, "%-15s %s\n", colors.Muted(key+":"), value)
}

// List prints a bulleted list of items
// Format: "• item1"
//
//	"• item2"
func (f *Formatter) List(items []string) {
	for _, item := range items {
		fmt.Fprintf(f.writer, "• %s\n", item)
	}
}

// NextSteps prints a "Next steps" section with common follow-up commands
// Format: "Next steps:"
//
//	"  • command1: description"
//	"  • command2: description"
func (f *Formatter) NextSteps(commands map[string]string) {
	f.EmptyLine()
	fmt.Fprintln(f.writer, "Next steps:")
	for cmd, desc := range commands {
		fmt.Fprintf(f.writer, "  • %s: %s\n", colors.Info(cmd), desc)
	}
}

// LegendItem represents a single item in a legend table
type LegendItem struct {
	Symbol      string
	Description string
}

// DisplayLegendTable displays a legend in a two-column box table format
// This provides a more professional appearance compared to bullet-point lists
func (f *Formatter) DisplayLegendTable(items []LegendItem) {
	if len(items) == 0 {
		return
	}

	table := NewBoxTable(f.writer, []string{"Symbol", "Meaning"})
	for _, item := range items {
		table.AddRow([]string{item.Symbol, item.Description})
	}
	table.Render()
}
