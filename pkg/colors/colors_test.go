package colors

import (
	"os"
	"testing"

	"github.com/fatih/color"
)

func TestInit(t *testing.T) {
	// Save original state
	origNoColor := os.Getenv("NO_COLOR")
	origForceColor := os.Getenv("FORCE_COLOR")
	defer func() {
		os.Setenv("NO_COLOR", origNoColor)
		os.Setenv("FORCE_COLOR", origForceColor)
	}()

	tests := []struct {
		name       string
		autoDetect bool
		noColor    string
		forceColor string
		wantColor  bool // expected color state when autoDetect is true
	}{
		{
			name:       "auto detect with NO_COLOR set",
			autoDetect: true,
			noColor:    "1",
			forceColor: "",
			wantColor:  false,
		},
		{
			name:       "auto detect with FORCE_COLOR set",
			autoDetect: true,
			noColor:    "",
			forceColor: "1",
			wantColor:  true,
		},
		{
			name:       "NO_COLOR takes precedence over FORCE_COLOR",
			autoDetect: true,
			noColor:    "1",
			forceColor: "1",
			wantColor:  false,
		},
		{
			name:       "auto detect disabled - colors enabled",
			autoDetect: false,
			noColor:    "1",
			forceColor: "",
			wantColor:  true, // Should ignore environment
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("NO_COLOR", tt.noColor)
			os.Setenv("FORCE_COLOR", tt.forceColor)

			// Initialize
			Init(tt.autoDetect)

			if tt.autoDetect {
				if enabled != tt.wantColor {
					t.Errorf("Init() enabled = %v, want %v", enabled, tt.wantColor)
				}
			} else {
				// When autoDetect is false, colors should always be enabled
				if !enabled {
					t.Error("Init(false) should enable colors")
				}
			}

			// Verify color.NoColor matches our state
			if color.NoColor == enabled {
				t.Errorf("color.NoColor = %v, should be inverse of enabled = %v", color.NoColor, enabled)
			}
		})
	}
}

func TestEnableDisable(t *testing.T) {
	// Test Enable
	Disable()
	if enabled {
		t.Error("Disable() failed to disable colors")
	}
	if !color.NoColor {
		t.Error("Disable() failed to set color.NoColor")
	}

	Enable()
	if !enabled {
		t.Error("Enable() failed to enable colors")
	}
	if color.NoColor {
		t.Error("Enable() failed to unset color.NoColor")
	}

	// Test Disable
	Disable()
	if enabled {
		t.Error("Disable() failed to disable colors")
	}
	if !color.NoColor {
		t.Error("Disable() failed to set color.NoColor")
	}
}

func TestIsEnabled(t *testing.T) {
	Enable()
	if !IsEnabled() {
		t.Error("IsEnabled() should return true after Enable()")
	}

	Disable()
	if IsEnabled() {
		t.Error("IsEnabled() should return false after Disable()")
	}
}

func TestColorFunctions(t *testing.T) {
	tests := []struct {
		name  string
		fn    func(a ...interface{}) string
		input string
	}{
		{"Success", Success, "test"},
		{"SuccessBold", SuccessBold, "test"},
		{"Error", Error, "test"},
		{"ErrorBold", ErrorBold, "test"},
		{"Warning", Warning, "test"},
		{"WarningBold", WarningBold, "test"},
		{"Info", Info, "test"},
		{"InfoBold", InfoBold, "test"},
		{"Highlight", Highlight, "test"},
		{"HighlightBold", HighlightBold, "test"},
		{"Muted", Muted, "test"},
		{"Dim", Dim, "test"},
		{"Code", Code, "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_enabled", func(t *testing.T) {
			Enable()
			result := tt.fn(tt.input)
			// When enabled, result should contain ANSI codes (be longer than input)
			if len(result) <= len(tt.input) {
				t.Errorf("%s() with colors enabled should add ANSI codes, got %q", tt.name, result)
			}
		})

		t.Run(tt.name+"_disabled", func(t *testing.T) {
			Disable()
			result := tt.fn(tt.input)
			// When disabled, result should equal input
			if result != tt.input {
				t.Errorf("%s() with colors disabled = %q, want %q", tt.name, result, tt.input)
			}
		})
	}
}

func TestShouldEnableColors(t *testing.T) {
	// Save original state
	origNoColor := os.Getenv("NO_COLOR")
	origForceColor := os.Getenv("FORCE_COLOR")
	defer func() {
		os.Setenv("NO_COLOR", origNoColor)
		os.Setenv("FORCE_COLOR", origForceColor)
	}()

	tests := []struct {
		name       string
		noColor    string
		forceColor string
		want       bool
	}{
		{
			name:       "NO_COLOR set disables colors",
			noColor:    "1",
			forceColor: "",
			want:       false,
		},
		{
			name:       "FORCE_COLOR enables colors",
			noColor:    "",
			forceColor: "1",
			want:       true,
		},
		{
			name:       "NO_COLOR has priority over FORCE_COLOR",
			noColor:    "1",
			forceColor: "1",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("NO_COLOR", tt.noColor)
			os.Setenv("FORCE_COLOR", tt.forceColor)

			// Note: When neither is set, it falls back to TTY detection
			// which we can't reliably test here, so we only test the env var logic
			if tt.noColor != "" || tt.forceColor != "" {
				got := shouldEnableColors()
				if got != tt.want {
					t.Errorf("shouldEnableColors() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestInitColorFunctions(t *testing.T) {
	// Test that initColorFunctions doesn't panic
	Enable()
	initColorFunctions()

	Disable()
	initColorFunctions()

	// Verify all functions are callable
	funcs := []func(a ...interface{}) string{
		Success, SuccessBold, Error, ErrorBold,
		Warning, WarningBold, Info, InfoBold,
		Highlight, HighlightBold, Muted, Dim, Code,
	}

	for _, fn := range funcs {
		if fn == nil {
			t.Error("Color function should not be nil after initColorFunctions()")
		}
		// Should not panic
		_ = fn("test")
	}
}
