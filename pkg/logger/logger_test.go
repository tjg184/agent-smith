package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelError, "ERROR"},
		{LevelWarn, "WARN"},
		{LevelInfo, "INFO"},
		{LevelDebug, "DEBUG"},
		{Level(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("Level.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNew(t *testing.T) {
	logger := New(LevelInfo)
	if logger == nil {
		t.Fatal("New() returned nil")
	}
	if logger.GetLevel() != LevelInfo {
		t.Errorf("New(LevelInfo).GetLevel() = %v, want %v", logger.GetLevel(), LevelInfo)
	}
	if !logger.showTags {
		t.Error("New() should enable showTags by default")
	}
}

func TestLogger_SetLevel(t *testing.T) {
	logger := New(LevelError)
	logger.SetLevel(LevelDebug)
	if logger.GetLevel() != LevelDebug {
		t.Errorf("After SetLevel(LevelDebug), GetLevel() = %v, want %v", logger.GetLevel(), LevelDebug)
	}
}

func TestLogger_SetOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelInfo)
	logger.SetOutput(&buf)
	logger.SetShowTags(false)

	logger.Info("test message")

	if !strings.Contains(buf.String(), "test message") {
		t.Errorf("SetOutput() did not redirect output, got: %s", buf.String())
	}
}

func TestLogger_SetErrorOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelError)
	logger.SetErrorOutput(&buf)
	logger.SetShowTags(false)

	logger.Error("error message")

	if !strings.Contains(buf.String(), "error message") {
		t.Errorf("SetErrorOutput() did not redirect error output, got: %s", buf.String())
	}
}

func TestLogger_SetPrefix(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelInfo)
	logger.SetOutput(&buf)
	logger.SetShowTags(false)
	logger.SetPrefix("[TEST] ")

	logger.Info("message")

	if !strings.Contains(buf.String(), "[TEST] message") {
		t.Errorf("SetPrefix() did not add prefix, got: %s", buf.String())
	}
}

func TestLogger_SetShowTags(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelInfo)
	logger.SetOutput(&buf)

	// Test with tags enabled (default)
	logger.Info("with tags")
	if !strings.Contains(buf.String(), "[INFO]") {
		t.Errorf("ShowTags=true should include [INFO] tag, got: %s", buf.String())
	}

	// Test with tags disabled
	buf.Reset()
	logger.SetShowTags(false)
	logger.Info("without tags")
	if strings.Contains(buf.String(), "[INFO]") {
		t.Errorf("ShowTags=false should not include [INFO] tag, got: %s", buf.String())
	}
}

func TestLogger_LevelFiltering(t *testing.T) {
	tests := []struct {
		name         string
		loggerLevel  Level
		logFunc      func(*Logger)
		shouldOutput bool
	}{
		// LevelError logger
		{"Error/Error", LevelError, func(l *Logger) { l.Error("test") }, true},
		{"Error/Warn", LevelError, func(l *Logger) { l.Warn("test") }, false},
		{"Error/Info", LevelError, func(l *Logger) { l.Info("test") }, false},
		{"Error/Debug", LevelError, func(l *Logger) { l.Debug("test") }, false},

		// LevelWarn logger
		{"Warn/Error", LevelWarn, func(l *Logger) { l.Error("test") }, true},
		{"Warn/Warn", LevelWarn, func(l *Logger) { l.Warn("test") }, true},
		{"Warn/Info", LevelWarn, func(l *Logger) { l.Info("test") }, false},
		{"Warn/Debug", LevelWarn, func(l *Logger) { l.Debug("test") }, false},

		// LevelInfo logger
		{"Info/Error", LevelInfo, func(l *Logger) { l.Error("test") }, true},
		{"Info/Warn", LevelInfo, func(l *Logger) { l.Warn("test") }, true},
		{"Info/Info", LevelInfo, func(l *Logger) { l.Info("test") }, true},
		{"Info/Debug", LevelInfo, func(l *Logger) { l.Debug("test") }, false},

		// LevelDebug logger
		{"Debug/Error", LevelDebug, func(l *Logger) { l.Error("test") }, true},
		{"Debug/Warn", LevelDebug, func(l *Logger) { l.Warn("test") }, true},
		{"Debug/Info", LevelDebug, func(l *Logger) { l.Info("test") }, true},
		{"Debug/Debug", LevelDebug, func(l *Logger) { l.Debug("test") }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(tt.loggerLevel)
			logger.SetOutput(&buf)
			logger.SetErrorOutput(&buf)

			tt.logFunc(logger)

			hasOutput := buf.Len() > 0
			if hasOutput != tt.shouldOutput {
				t.Errorf("Level filtering failed: loggerLevel=%v, shouldOutput=%v, hasOutput=%v, output=%q",
					tt.loggerLevel, tt.shouldOutput, hasOutput, buf.String())
			}
		})
	}
}

func TestLogger_ErrorOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	logger := New(LevelDebug)
	logger.SetOutput(&stdout)
	logger.SetErrorOutput(&stderr)
	logger.SetShowTags(false)

	// Error and Warn should go to stderr
	logger.Error("error message")
	logger.Warn("warn message")

	// Info and Debug should go to stdout
	logger.Info("info message")
	logger.Debug("debug message")

	if !strings.Contains(stderr.String(), "error message") {
		t.Error("Error() should write to stderr")
	}
	if !strings.Contains(stderr.String(), "warn message") {
		t.Error("Warn() should write to stderr")
	}
	if !strings.Contains(stdout.String(), "info message") {
		t.Error("Info() should write to stdout")
	}
	if !strings.Contains(stdout.String(), "debug message") {
		t.Error("Debug() should write to stdout")
	}
}

func TestLogger_FormattingAliases(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelDebug)
	logger.SetOutput(&buf)
	logger.SetErrorOutput(&buf)
	logger.SetShowTags(false)

	// Test all formatting aliases
	logger.Errorf("error %d", 1)
	logger.Warnf("warn %d", 2)
	logger.Infof("info %d", 3)
	logger.Debugf("debug %d", 4)

	output := buf.String()
	expectations := []string{"error 1", "warn 2", "info 3", "debug 4"}
	for _, expected := range expectations {
		if !strings.Contains(output, expected) {
			t.Errorf("Formatting alias failed, expected %q in output: %s", expected, output)
		}
	}
}

func TestLogger_Print(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelError) // Set to Error level
	logger.SetOutput(&buf)

	// Print should always output regardless of level
	logger.Print("always visible")

	if !strings.Contains(buf.String(), "always visible") {
		t.Error("Print() should always output regardless of log level")
	}
	if strings.Contains(buf.String(), "[") {
		t.Error("Print() should not include level tags")
	}
}

func TestLogger_Printf(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelError)
	logger.SetOutput(&buf)

	logger.Printf("formatted %s %d", "test", 42)

	if !strings.Contains(buf.String(), "formatted test 42") {
		t.Errorf("Printf() formatting failed, got: %s", buf.String())
	}
}

func TestLogger_Println(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelError)
	logger.SetOutput(&buf)

	logger.Println("line", "one")

	output := buf.String()
	if !strings.Contains(output, "line") || !strings.Contains(output, "one") {
		t.Errorf("Println() failed, got: %s", output)
	}
}

func TestLogger_AutomaticNewline(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelInfo)
	logger.SetOutput(&buf)
	logger.SetShowTags(false)

	// Test without newline
	logger.Info("no newline")
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Errorf("Info() should add automatic newline, got %d lines", len(lines))
	}

	// Test with newline
	buf.Reset()
	logger.Info("with newline\n")
	lines = strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Errorf("Info() should not double newline, got %d lines", len(lines))
	}
}

func TestDefault(t *testing.T) {
	tests := []struct {
		name          string
		debug         bool
		verbose       bool
		expectedLevel Level
	}{
		{"default", false, false, LevelWarn},
		{"verbose", false, true, LevelInfo},
		{"debug", true, false, LevelDebug},
		{"debug+verbose", true, true, LevelDebug}, // debug takes precedence
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := Default(tt.debug, tt.verbose)
			if logger.GetLevel() != tt.expectedLevel {
				t.Errorf("Default(%v, %v).GetLevel() = %v, want %v",
					tt.debug, tt.verbose, logger.GetLevel(), tt.expectedLevel)
			}
		})
	}
}

func TestLogger_Concurrency(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelDebug)
	logger.SetOutput(&buf)
	logger.SetErrorOutput(&buf)

	// Test concurrent access to logger
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			logger.SetLevel(LevelInfo)
			logger.Info("concurrent %d", n)
			logger.GetLevel()
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Just ensure no race conditions occurred (test passes if it doesn't panic)
	if buf.Len() == 0 {
		t.Error("Expected some output from concurrent logging")
	}
}

func TestLogger_ColorSupport(t *testing.T) {
	tests := []struct {
		name     string
		level    Level
		showTags bool
		colorize bool
		logFunc  func(*Logger)
		wantANSI bool
	}{
		{
			name:     "Error with colors and tags",
			level:    LevelError,
			showTags: true,
			colorize: true,
			logFunc:  func(l *Logger) { l.Error("test") },
			wantANSI: true,
		},
		{
			name:     "Warn with colors and tags",
			level:    LevelWarn,
			showTags: true,
			colorize: true,
			logFunc:  func(l *Logger) { l.Warn("test") },
			wantANSI: true,
		},
		{
			name:     "Info with colors and tags",
			level:    LevelInfo,
			showTags: true,
			colorize: true,
			logFunc:  func(l *Logger) { l.Info("test") },
			wantANSI: true,
		},
		{
			name:     "Debug with colors and tags",
			level:    LevelDebug,
			showTags: true,
			colorize: true,
			logFunc:  func(l *Logger) { l.Debug("test") },
			wantANSI: true,
		},
		{
			name:     "Error without colors",
			level:    LevelError,
			showTags: true,
			colorize: false,
			logFunc:  func(l *Logger) { l.Error("test") },
			wantANSI: false,
		},
		{
			name:     "Error without tags but with colors",
			level:    LevelError,
			showTags: false,
			colorize: true,
			logFunc:  func(l *Logger) { l.Error("test") },
			wantANSI: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(tt.level)
			logger.SetOutput(&buf)
			logger.SetErrorOutput(&buf)
			logger.SetShowTags(tt.showTags)
			logger.SetColorize(tt.colorize)

			tt.logFunc(logger)

			output := buf.String()
			hasANSI := strings.Contains(output, "\x1b[")

			if hasANSI != tt.wantANSI {
				t.Errorf("Color support failed: wantANSI=%v, hasANSI=%v, output=%q",
					tt.wantANSI, hasANSI, output)
			}

			// Ensure message contains "test"
			if !strings.Contains(output, "test") {
				t.Errorf("Output should contain 'test', got: %q", output)
			}
		})
	}
}

func TestLogger_SetColorize(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelInfo)
	logger.SetOutput(&buf)

	// Test with colors enabled (default)
	logger.Info("with colors")
	withColors := buf.String()

	// Test with colors disabled
	buf.Reset()
	logger.SetColorize(false)
	logger.Info("without colors")
	withoutColors := buf.String()

	// With colors should have ANSI codes, without should not
	hasANSIWith := strings.Contains(withColors, "\x1b[")
	hasANSIWithout := strings.Contains(withoutColors, "\x1b[")

	if !hasANSIWith {
		t.Error("SetColorize(true) should produce ANSI color codes")
	}
	if hasANSIWithout {
		t.Error("SetColorize(false) should not produce ANSI color codes")
	}
}
