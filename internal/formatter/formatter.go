package formatter

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/fatih/color"
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
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Fprintf(f.writer, "%s Installed %s: %s\n", green(SymbolSuccess), componentType, name)
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
	green := color.New(color.FgGreen).SprintFunc()
	return green(SymbolSuccess)
}

// ColoredError returns a red-colored error symbol
func ColoredError() string {
	red := color.New(color.FgRed).SprintFunc()
	return red(SymbolError)
}

// ColoredWarning returns a yellow-colored warning symbol
func ColoredWarning() string {
	yellow := color.New(color.FgYellow).SprintFunc()
	return yellow(SymbolWarning)
}

// SectionHeader prints a section header with consistent formatting
// Example: === Section Title ===
func (f *Formatter) SectionHeader(title string) {
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	fmt.Fprintf(f.writer, "\n%s\n", cyan("=== "+title+" ==="))
}

// SubsectionHeader prints a subsection header with consistent formatting
// Example: --- Subsection Title ---
func (f *Formatter) SubsectionHeader(title string) {
	blue := color.New(color.FgBlue).SprintFunc()
	fmt.Fprintf(f.writer, "\n%s\n", blue("--- "+title+" ---"))
}

// ProgressMsg prints a progress message for ongoing operations
// Example: "Linking skill: api-design... "
func (f *Formatter) ProgressMsg(operation, item string) {
	fmt.Fprintf(f.writer, "%s: %s... ", operation, item)
}

// ProgressComplete prints a completion marker (checkmark) for progress
func (f *Formatter) ProgressComplete() {
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Fprintf(f.writer, "%s Done\n", green(SymbolSuccess))
}

// ProgressFailed prints a failure marker for progress
func (f *Formatter) ProgressFailed() {
	red := color.New(color.FgRed).SprintFunc()
	fmt.Fprintf(f.writer, "%s FAILED\n", red(SymbolError))
}

// SuccessMsg prints a success message with green checkmark
func (f *Formatter) SuccessMsg(message string, args ...interface{}) {
	green := color.New(color.FgGreen).SprintFunc()
	msg := fmt.Sprintf(message, args...)
	fmt.Fprintf(f.writer, "%s %s\n", green(SymbolSuccess), msg)
}

// ErrorMsg prints an error message with red X symbol
func (f *Formatter) ErrorMsg(message string, args ...interface{}) {
	red := color.New(color.FgRed).SprintFunc()
	msg := fmt.Sprintf(message, args...)
	fmt.Fprintf(f.writer, "%s %s\n", red(SymbolError), msg)
}

// WarningMsg prints a warning message with yellow warning symbol
func (f *Formatter) WarningMsg(message string, args ...interface{}) {
	yellow := color.New(color.FgYellow).SprintFunc()
	msg := fmt.Sprintf(message, args...)
	fmt.Fprintf(f.writer, "%s %s\n", yellow(SymbolWarning), msg)
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
	gray := color.New(color.FgHiBlack).SprintFunc()
	fmt.Fprintf(f.writer, "    %s: %s\n", gray(key), value)
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
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Fprintf(f.writer, "Total: %d\n", total)
	if success > 0 {
		fmt.Fprintf(f.writer, "%s Successful: %d\n", green(SymbolSuccess), success)
	}
	if failed > 0 {
		fmt.Fprintf(f.writer, "%s Failed: %d\n", red(SymbolError), failed)
	}
	if skipped > 0 {
		fmt.Fprintf(f.writer, "%s Skipped: %d\n", yellow(SymbolWarning), skipped)
	}
}
