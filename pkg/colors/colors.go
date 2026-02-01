// Package colors provides a centralized color system with proper TTY detection
// and consistent color definitions across the application.
package colors

import (
	"os"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

var (
	// Success colors (green)
	Success     func(a ...interface{}) string
	SuccessBold func(a ...interface{}) string

	// Error colors (red)
	Error     func(a ...interface{}) string
	ErrorBold func(a ...interface{}) string

	// Warning colors (yellow)
	Warning     func(a ...interface{}) string
	WarningBold func(a ...interface{}) string

	// Info colors (cyan)
	Info     func(a ...interface{}) string
	InfoBold func(a ...interface{}) string

	// Highlight colors (blue)
	Highlight     func(a ...interface{}) string
	HighlightBold func(a ...interface{}) string

	// Secondary/muted colors (gray/dim)
	Muted func(a ...interface{}) string
	Dim   func(a ...interface{}) string

	// Code/path styling (white bold)
	Code func(a ...interface{}) string

	// enabled tracks whether colors are enabled
	enabled bool
)

func init() {
	// Initialize with auto-detection
	Init(true)
}

// Init initializes the color system with optional auto-detection.
// If autoDetect is true, it will check TTY status and NO_COLOR environment variable.
// If autoDetect is false, colors will be enabled unconditionally.
func Init(autoDetect bool) {
	if autoDetect {
		enabled = shouldEnableColors()
	} else {
		enabled = true
	}

	// Configure fatih/color library
	color.NoColor = !enabled

	// Initialize all color functions
	initColorFunctions()
}

// Enable forces colors to be enabled, regardless of TTY or environment variables.
func Enable() {
	enabled = true
	color.NoColor = false
	initColorFunctions()
}

// Disable forces colors to be disabled.
func Disable() {
	enabled = false
	color.NoColor = true
	initColorFunctions()
}

// IsEnabled returns whether colors are currently enabled.
func IsEnabled() bool {
	return enabled
}

// shouldEnableColors determines if colors should be enabled based on:
// 1. NO_COLOR environment variable (if set, colors are disabled)
// 2. FORCE_COLOR environment variable (if set, colors are enabled)
// 3. TTY detection (colors enabled only if stdout is a TTY)
func shouldEnableColors() bool {
	// Check NO_COLOR environment variable (highest priority for disabling)
	// https://no-color.org/
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check FORCE_COLOR environment variable (overrides TTY detection)
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}

	// Check if stdout is a TTY
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

// initColorFunctions initializes all color function pointers.
// When colors are disabled, these functions return the input unchanged.
func initColorFunctions() {
	if enabled {
		// Success (green)
		Success = color.New(color.FgGreen).SprintFunc()
		SuccessBold = color.New(color.FgGreen, color.Bold).SprintFunc()

		// Error (red)
		Error = color.New(color.FgRed).SprintFunc()
		ErrorBold = color.New(color.FgRed, color.Bold).SprintFunc()

		// Warning (yellow)
		Warning = color.New(color.FgYellow).SprintFunc()
		WarningBold = color.New(color.FgYellow, color.Bold).SprintFunc()

		// Info (cyan)
		Info = color.New(color.FgCyan).SprintFunc()
		InfoBold = color.New(color.FgCyan, color.Bold).SprintFunc()

		// Highlight (blue)
		Highlight = color.New(color.FgBlue).SprintFunc()
		HighlightBold = color.New(color.FgBlue, color.Bold).SprintFunc()

		// Muted/Dim (gray/faint)
		Muted = color.New(color.FgHiBlack).SprintFunc()
		Dim = color.New(color.Faint).SprintFunc()

		// Code (white bold)
		Code = color.New(color.FgWhite, color.Bold).SprintFunc()
	} else {
		// When disabled, all functions return input unchanged
		noop := func(a ...interface{}) string {
			return color.New().SprintFunc()(a...)
		}

		Success = noop
		SuccessBold = noop
		Error = noop
		ErrorBold = noop
		Warning = noop
		WarningBold = noop
		Info = noop
		InfoBold = noop
		Highlight = noop
		HighlightBold = noop
		Muted = noop
		Dim = noop
		Code = noop
	}
}
