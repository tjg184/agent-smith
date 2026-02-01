package formatter

import (
	"fmt"
	"io"
	"log"
	"os"

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

// CounterSummary prints a summary with counters
func (f *Formatter) CounterSummary(total, success, failed, skipped int) {
	fmt.Fprintf(f.writer, "Total: %d\n", total)
	if success > 0 {
		fmt.Fprintf(f.writer, "%s Successful: %d\n", colors.Success(SymbolSuccess), success)
	}
	if failed > 0 {
		fmt.Fprintf(f.writer, "%s Failed: %d\n", colors.Error(SymbolError), failed)
	}
	if skipped > 0 {
		fmt.Fprintf(f.writer, "%s Skipped: %d\n", colors.Warning(SymbolWarning), skipped)
	}
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
