// Package errors provides colored error formatting with context for better user experience.
package errors

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

var (
	// Error styling
	errorPrefix = color.New(color.FgRed, color.Bold).SprintFunc()
	errorIcon   = errorPrefix("✗")
	errorText   = color.New(color.FgRed).SprintFunc()

	// Warning styling
	warnPrefix = color.New(color.FgYellow, color.Bold).SprintFunc()
	warnIcon   = warnPrefix("⚠")
	warnText   = color.New(color.FgYellow).SprintFunc()

	// Context styling
	contextPrefix = color.New(color.FgCyan, color.Bold).SprintFunc()
	contextText   = color.New(color.FgCyan).SprintFunc()

	// Suggestion styling
	suggestionPrefix = color.New(color.FgGreen, color.Bold).SprintFunc()
	suggestionText   = color.New(color.FgGreen).SprintFunc()

	// Code/path styling
	codeText = color.New(color.FgWhite, color.Bold).SprintFunc()

	// Dim text for additional details
	dimText = color.New(color.Faint).SprintFunc()
)

// ErrorMessage represents a structured error message with context and suggestions.
type ErrorMessage struct {
	Message    string
	Context    string
	Suggestion string
	Example    string
	Details    []string
	IsWarning  bool
}

// Format returns the fully formatted error message with colors.
func (e *ErrorMessage) Format() string {
	var sb strings.Builder

	// Main error/warning message
	if e.IsWarning {
		sb.WriteString(fmt.Sprintf("%s %s\n", warnIcon, warnText(e.Message)))
	} else {
		sb.WriteString(fmt.Sprintf("%s %s\n", errorIcon, errorText(e.Message)))
	}

	// Add context if provided
	if e.Context != "" {
		sb.WriteString(fmt.Sprintf("\n%s %s\n", contextPrefix("Context:"), contextText(e.Context)))
	}

	// Add details if provided
	if len(e.Details) > 0 {
		sb.WriteString("\n")
		for _, detail := range e.Details {
			sb.WriteString(fmt.Sprintf("  %s %s\n", dimText("•"), detail))
		}
	}

	// Add suggestion if provided
	if e.Suggestion != "" {
		sb.WriteString(fmt.Sprintf("\n%s %s\n", suggestionPrefix("Suggestion:"), suggestionText(e.Suggestion)))
	}

	// Add example if provided
	if e.Example != "" {
		sb.WriteString(fmt.Sprintf("\n%s\n", codeText("  $ "+e.Example)))
	}

	return sb.String()
}

// String returns the formatted error message.
func (e *ErrorMessage) String() string {
	return e.Format()
}

// New creates a new error message with just the main message.
func New(message string) *ErrorMessage {
	return &ErrorMessage{
		Message: message,
	}
}

// NewWithContext creates an error message with context.
func NewWithContext(message, context string) *ErrorMessage {
	return &ErrorMessage{
		Message: message,
		Context: context,
	}
}

// WithContext adds context to an error message.
func (e *ErrorMessage) WithContext(context string) *ErrorMessage {
	e.Context = context
	return e
}

// WithSuggestion adds a suggestion to an error message.
func (e *ErrorMessage) WithSuggestion(suggestion string) *ErrorMessage {
	e.Suggestion = suggestion
	return e
}

// WithExample adds an example command to an error message.
func (e *ErrorMessage) WithExample(example string) *ErrorMessage {
	e.Example = example
	return e
}

// WithDetails adds detail lines to an error message.
func (e *ErrorMessage) WithDetails(details ...string) *ErrorMessage {
	e.Details = append(e.Details, details...)
	return e
}

// AsWarning marks this as a warning instead of an error.
func (e *ErrorMessage) AsWarning() *ErrorMessage {
	e.IsWarning = true
	return e
}

// FormatSimpleError formats a simple error message with color.
func FormatSimpleError(message string) string {
	return fmt.Sprintf("%s %s", errorIcon, errorText(message))
}

// FormatSimpleWarning formats a simple warning message with color.
func FormatSimpleWarning(message string) string {
	return fmt.Sprintf("%s %s", warnIcon, warnText(message))
}

// FormatCode formats code/path text for display.
func FormatCode(text string) string {
	return codeText(text)
}

// FormatDim formats dimmed text for additional details.
func FormatDim(text string) string {
	return dimText(text)
}

// Disable removes all color formatting (useful for testing or piped output).
func Disable() {
	color.NoColor = true
}

// Enable enables color formatting.
func Enable() {
	color.NoColor = false
}
