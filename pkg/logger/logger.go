// Package logger provides a consistent log level system for controlling output granularity.
package logger

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/tjg184/agent-smith/pkg/colors"
	"github.com/tjg184/agent-smith/pkg/errors"
)

// Level represents the severity level of a log message.
type Level int

const (
	// LevelError represents error messages that should always be shown.
	LevelError Level = iota
	// LevelWarn represents warning messages.
	LevelWarn
	// LevelInfo represents informational messages (enabled with --verbose).
	LevelInfo
	// LevelDebug represents debug messages (enabled with --debug).
	LevelDebug
)

// String returns the string representation of the log level.
func (l Level) String() string {
	switch l {
	case LevelError:
		return "ERROR"
	case LevelWarn:
		return "WARN"
	case LevelInfo:
		return "INFO"
	case LevelDebug:
		return "DEBUG"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging with configurable output levels.
type Logger struct {
	mu       sync.RWMutex
	level    Level
	output   io.Writer
	errOut   io.Writer
	prefix   string
	showTags bool
	colorize bool
}

// New creates a new Logger with the specified minimum level.
// Messages below this level will be discarded.
func New(level Level) *Logger {
	return &Logger{
		level:    level,
		output:   os.Stdout,
		errOut:   os.Stderr,
		showTags: true,
		colorize: true,
	}
}

// SetLevel changes the minimum log level.
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current log level.
func (l *Logger) GetLevel() Level {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// SetOutput sets the output destination for info and debug messages.
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
}

// SetErrorOutput sets the output destination for error and warning messages.
func (l *Logger) SetErrorOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.errOut = w
}

// SetPrefix sets a prefix to prepend to all log messages.
func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}

// SetShowTags controls whether to show [LEVEL] tags in output.
func (l *Logger) SetShowTags(show bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.showTags = show
}

// SetColorize controls whether to use colored output.
func (l *Logger) SetColorize(colorize bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.colorize = colorize
	if !colorize {
		colors.Disable()
	} else {
		colors.Enable()
	}
}

// log is the internal logging function that handles level filtering and formatting.
func (l *Logger) log(level Level, format string, args ...interface{}) {
	l.mu.RLock()
	currentLevel := l.level
	output := l.output
	errOut := l.errOut
	prefix := l.prefix
	showTags := l.showTags
	colorize := l.colorize
	l.mu.RUnlock()

	// Filter messages below the current level
	if level > currentLevel {
		return
	}

	// Determine output destination
	writer := output
	if level <= LevelWarn {
		writer = errOut
	}

	// Build the message
	var msg string
	if showTags {
		tag := fmt.Sprintf("[%s]", level)
		// Apply color to the tag if enabled
		if colorize {
			switch level {
			case LevelError:
				tag = colors.ErrorBold(tag)
			case LevelWarn:
				tag = colors.WarningBold(tag)
			case LevelInfo:
				tag = colors.InfoBold(tag)
			case LevelDebug:
				tag = colors.Muted(tag)
			}
		}
		msg = tag + " " + prefix + fmt.Sprintf(format, args...)
	} else {
		msg = prefix + fmt.Sprintf(format, args...)
		// Apply color formatting if enabled and tags are disabled
		if colorize {
			switch level {
			case LevelError:
				msg = errors.FormatSimpleError(msg)
			case LevelWarn:
				msg = errors.FormatSimpleWarning(msg)
			}
		}
	}

	// Ensure newline at end
	if len(msg) == 0 || msg[len(msg)-1] != '\n' {
		msg += "\n"
	}

	fmt.Fprint(writer, msg)
}

// Error logs an error message. Error messages are always shown.
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// Errorf is an alias for Error.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// Warn logs a warning message. Warning messages are shown at warn level and above.
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Warnf is an alias for Warn.
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Info logs an informational message. Info messages are shown with --verbose or --debug.
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Infof is an alias for Info.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Debug logs a debug message. Debug messages are only shown with --debug.
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Debugf is an alias for Debug.
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Fatal logs an error message and exits the program with status code 1.
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
	os.Exit(1)
}

// Fatalf is an alias for Fatal.
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Fatal(format, args...)
}

// Print logs a message without any level tag, regardless of log level.
// This is useful for command output that should always be shown.
func (l *Logger) Print(format string, args ...interface{}) {
	l.mu.RLock()
	output := l.output
	l.mu.RUnlock()

	msg := fmt.Sprintf(format, args...)
	if len(msg) == 0 || msg[len(msg)-1] != '\n' {
		msg += "\n"
	}
	fmt.Fprint(output, msg)
}

// Printf is an alias for Print.
func (l *Logger) Printf(format string, args ...interface{}) {
	l.Print(format, args...)
}

// Println logs a message without any level tag, regardless of log level.
func (l *Logger) Println(args ...interface{}) {
	l.mu.RLock()
	output := l.output
	l.mu.RUnlock()

	fmt.Fprintln(output, args...)
}

// Default returns a logger configured based on common flags.
// - If debug is true, sets level to LevelDebug
// - If verbose is true, sets level to LevelInfo
// - Otherwise, sets level to LevelWarn
func Default(debug, verbose bool) *Logger {
	var level Level
	if debug {
		level = LevelDebug
	} else if verbose {
		level = LevelInfo
	} else {
		level = LevelWarn
	}
	return New(level)
}

// ErrorMsg logs a structured error message with context and suggestions.
func (l *Logger) ErrorMsg(errMsg *errors.ErrorMessage) {
	l.mu.RLock()
	errOut := l.errOut
	l.mu.RUnlock()

	fmt.Fprint(errOut, errMsg.Format())
}

// FatalMsg logs a structured error message and exits the program with status code 1.
func (l *Logger) FatalMsg(errMsg *errors.ErrorMessage) {
	l.ErrorMsg(errMsg)
	os.Exit(1)
}

// WarnMsg logs a structured warning message with context and suggestions.
func (l *Logger) WarnMsg(warnMsg *errors.ErrorMessage) {
	l.mu.RLock()
	errOut := l.errOut
	l.mu.RUnlock()

	warnMsg.AsWarning()
	fmt.Fprint(errOut, warnMsg.Format())
}
