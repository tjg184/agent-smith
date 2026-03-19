// Package colors provides a centralized color system with proper TTY detection
// and consistent color definitions across the application.
package colors

import (
	"os"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

var (
	Success     func(a ...interface{}) string
	SuccessBold func(a ...interface{}) string

	Error     func(a ...interface{}) string
	ErrorBold func(a ...interface{}) string

	Warning     func(a ...interface{}) string
	WarningBold func(a ...interface{}) string

	Info     func(a ...interface{}) string
	InfoBold func(a ...interface{}) string

	Highlight     func(a ...interface{}) string
	HighlightBold func(a ...interface{}) string

	Muted func(a ...interface{}) string
	Dim   func(a ...interface{}) string

	Code func(a ...interface{}) string

	enabled bool
)

func init() {
	Init(true)
}

func Init(autoDetect bool) {
	if autoDetect {
		enabled = shouldEnableColors()
	} else {
		enabled = true
	}

	color.NoColor = !enabled
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

func shouldEnableColors() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}

	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

// initColorFunctions initializes all color function pointers.
// When colors are disabled, these functions return the input unchanged.
func initColorFunctions() {
	if enabled {
		Success = color.New(color.FgGreen).SprintFunc()
		SuccessBold = color.New(color.FgGreen, color.Bold).SprintFunc()

		Error = color.New(color.FgRed).SprintFunc()
		ErrorBold = color.New(color.FgRed, color.Bold).SprintFunc()

		Warning = color.New(color.FgYellow).SprintFunc()
		WarningBold = color.New(color.FgYellow, color.Bold).SprintFunc()

		Info = color.New(color.FgCyan).SprintFunc()
		InfoBold = color.New(color.FgCyan, color.Bold).SprintFunc()

		Highlight = color.New(color.FgBlue).SprintFunc()
		HighlightBold = color.New(color.FgBlue, color.Bold).SprintFunc()

		Muted = color.New(color.FgHiBlack).SprintFunc()
		Dim = color.New(color.Faint).SprintFunc()

		Code = color.New(color.FgWhite, color.Bold).SprintFunc()
	} else {
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
