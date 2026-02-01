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
