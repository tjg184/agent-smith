// Package errors provides colored error formatting with context for better user experience.
package errors

import (
	"fmt"
	"strings"

	"github.com/tjg184/agent-smith/pkg/colors"
)

var (
	// Error styling
	errorIcon string
	// Warning styling
	warnIcon string
)

func init() {
	// Initialize icons with colors
	errorIcon = colors.ErrorBold("✗")
	warnIcon = colors.WarningBold("⚠")
}

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
		sb.WriteString(fmt.Sprintf("%s %s\n", warnIcon, colors.Warning(e.Message)))
	} else {
		sb.WriteString(fmt.Sprintf("%s %s\n", errorIcon, colors.Error(e.Message)))
	}

	// Add context if provided
	if e.Context != "" {
		sb.WriteString(fmt.Sprintf("\n%s %s\n", colors.InfoBold("Context:"), colors.Info(e.Context)))
	}

	// Add details if provided
	if len(e.Details) > 0 {
		sb.WriteString("\n")
		for _, detail := range e.Details {
			sb.WriteString(fmt.Sprintf("  %s %s\n", colors.Dim("•"), detail))
		}
	}

	// Add suggestion if provided
	if e.Suggestion != "" {
		sb.WriteString(fmt.Sprintf("\n%s %s\n", colors.SuccessBold("Suggestion:"), colors.Success(e.Suggestion)))
	}

	// Add example if provided
	if e.Example != "" {
		sb.WriteString(fmt.Sprintf("\n%s\n", colors.Code("  $ "+e.Example)))
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
	return fmt.Sprintf("%s %s", errorIcon, colors.Error(message))
}

// FormatSimpleWarning formats a simple warning message with color.
func FormatSimpleWarning(message string) string {
	return fmt.Sprintf("%s %s", warnIcon, colors.Warning(message))
}

// FormatCode formats code/path text for display.
func FormatCode(text string) string {
	return colors.Code(text)
}

// FormatDim formats dimmed text for additional details.
func FormatDim(text string) string {
	return colors.Dim(text)
}

// Disable removes all color formatting (useful for testing or piped output).
func Disable() {
	colors.Disable()
	// Reinitialize icons without colors
	errorIcon = "✗"
	warnIcon = "⚠"
}

// Enable enables color formatting.
func Enable() {
	colors.Enable()
	// Reinitialize icons with colors
	errorIcon = colors.ErrorBold("✗")
	warnIcon = colors.WarningBold("⚠")
}
