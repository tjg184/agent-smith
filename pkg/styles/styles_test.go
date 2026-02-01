package styles

import (
	"strings"
	"testing"

	"github.com/tgaines/agent-smith/internal/formatter"
	"github.com/tgaines/agent-smith/pkg/colors"
)

func TestProgressCheckingFormat(t *testing.T) {
	// Disable colors for predictable testing
	colors.Disable()
	defer colors.Enable()

	result := ProgressCheckingFormat("skills", "api-design")
	expected := "Checking skills/api-design... "
	if result != expected {
		t.Errorf("ProgressCheckingFormat() = %q, want %q", result, expected)
	}
}

func TestStatusFormats(t *testing.T) {
	// Disable colors for predictable testing
	colors.Disable()
	defer colors.Enable()

	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{
			name:     "StatusFailedFormat",
			fn:       StatusFailedFormat,
			expected: formatter.SymbolError + " Failed",
		},
		{
			name:     "StatusUpToDateFormat",
			fn:       StatusUpToDateFormat,
			expected: formatter.SymbolSuccess + " Up to date",
		},
		{
			name:     "StatusUpdatingFormat",
			fn:       StatusUpdatingFormat,
			expected: formatter.SymbolUpdating + " Updating",
		},
		{
			name:     "StatusUpdatedSuccessfullyFormat",
			fn:       StatusUpdatedSuccessfullyFormat,
			expected: formatter.SymbolSuccess + " Updated successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			if result != tt.expected {
				t.Errorf("%s() = %q, want %q", tt.name, result, tt.expected)
			}
		})
	}
}

func TestIndentedFormats(t *testing.T) {
	// Disable colors for predictable testing
	colors.Disable()
	defer colors.Enable()

	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{
			name: "IndentedErrorFormat",
			fn: func() string {
				return IndentedErrorFormat("something went wrong")
			},
			expected: "  └─ something went wrong",
		},
		{
			name: "IndentedDetailFormat",
			fn: func() string {
				return IndentedDetailFormat("key", "value")
			},
			expected: "  → key: value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			if result != tt.expected {
				t.Errorf("%s() = %q, want %q", tt.name, result, tt.expected)
			}
		})
	}
}

func TestInlineFormats(t *testing.T) {
	// Disable colors for predictable testing
	colors.Disable()
	defer colors.Enable()

	tests := []struct {
		name     string
		fn       func() string
		contains []string
	}{
		{
			name: "InlineSuccessFormat",
			fn: func() string {
				return InlineSuccessFormat("Linking", "skill", "api-design")
			},
			contains: []string{"Linking skill: api-design", formatter.SymbolSuccess, "Done"},
		},
		{
			name: "InlineSuccessWithNoteFormat",
			fn: func() string {
				return InlineSuccessWithNoteFormat("Linking", "agent", "coder", "from profile: dev")
			},
			contains: []string{"Linking agent: coder", formatter.SymbolSuccess, "Done", "(from profile: dev)"},
		},
		{
			name: "InlineFailedFormat",
			fn: func() string {
				return InlineFailedFormat("Linking", "command", "test")
			},
			contains: []string{"Linking command: test", formatter.SymbolError, "Failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("%s() = %q, should contain %q", tt.name, result, substr)
				}
			}
		})
	}
}

func TestInfoArrowFormat(t *testing.T) {
	// Disable colors for predictable testing
	colors.Disable()
	defer colors.Enable()

	result := InfoArrowFormat("Checking for updates...")
	expected := "→ Checking for updates..."
	if result != expected {
		t.Errorf("InfoArrowFormat() = %q, want %q", result, expected)
	}
}

func TestComponentProgressFormat(t *testing.T) {
	// Disable colors for predictable testing
	colors.Disable()
	defer colors.Enable()

	result := ComponentProgressFormat(3, 10, "agents", "coder")
	expected := "[3/10] agents/coder... "
	if result != expected {
		t.Errorf("ComponentProgressFormat() = %q, want %q", result, expected)
	}
}

func TestProfileNoteFormat(t *testing.T) {
	// Disable colors for predictable testing
	colors.Disable()
	defer colors.Enable()

	tests := []struct {
		name        string
		profileName string
		expected    string
	}{
		{
			name:        "With profile name",
			profileName: "dev",
			expected:    " (from profile: dev)",
		},
		{
			name:        "With base profile",
			profileName: "base",
			expected:    "",
		},
		{
			name:        "Empty profile",
			profileName: "",
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProfileNoteFormat(tt.profileName)
			if result != tt.expected {
				t.Errorf("ProfileNoteFormat(%q) = %q, want %q", tt.profileName, result, tt.expected)
			}
		})
	}
}

func TestSummaryTableBuilder(t *testing.T) {
	// Disable colors for predictable testing
	colors.Disable()
	defer colors.Enable()

	table := SummaryTableFormat("Test Summary", 60)
	table.AddRow("Total components", 10)
	table.AddRow("Successful", 8)
	table.AddRow("Failed", 2)

	result := table.Build()

	// Check for essential elements
	if !strings.Contains(result, "┌") {
		t.Error("SummaryTableBuilder should contain top-left corner")
	}
	if !strings.Contains(result, "└") {
		t.Error("SummaryTableBuilder should contain bottom-left corner")
	}
	if !strings.Contains(result, "Test Summary") {
		t.Error("SummaryTableBuilder should contain title")
	}
	if !strings.Contains(result, "Total components") {
		t.Error("SummaryTableBuilder should contain row data")
	}
}

func TestSummaryTableBuilderWithSymbol(t *testing.T) {
	// Disable colors for predictable testing
	colors.Disable()
	defer colors.Enable()

	table := SummaryTableFormat("Results", 60)
	table.AddRowWithSymbol(formatter.SymbolSuccess, "Passed", 5)
	table.AddRowWithSymbol(formatter.SymbolError, "Failed", 1)

	result := table.Build()

	// Check for symbols
	if !strings.Contains(result, formatter.SymbolSuccess) {
		t.Error("SummaryTableBuilder should contain success symbol")
	}
	if !strings.Contains(result, formatter.SymbolError) {
		t.Error("SummaryTableBuilder should contain error symbol")
	}
}

func TestStatusSymbol(t *testing.T) {
	// Disable colors for predictable testing
	colors.Disable()
	defer colors.Enable()

	tests := []struct {
		name     string
		success  bool
		expected string
	}{
		{
			name:     "Success",
			success:  true,
			expected: formatter.SymbolSuccess,
		},
		{
			name:     "Failure",
			success:  false,
			expected: formatter.SymbolError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StatusSymbol(tt.success)
			if result != tt.expected {
				t.Errorf("StatusSymbol(%v) = %q, want %q", tt.success, result, tt.expected)
			}
		})
	}
}

func TestCounterRowFormat(t *testing.T) {
	// Disable colors for predictable testing
	colors.Disable()
	defer colors.Enable()

	result := CounterRowFormat(formatter.SymbolSuccess, "Total", 42)
	if !strings.Contains(result, formatter.SymbolSuccess) {
		t.Error("CounterRowFormat should contain symbol")
	}
	if !strings.Contains(result, "Total:") {
		t.Error("CounterRowFormat should contain label with colon")
	}
	if !strings.Contains(result, "42") {
		t.Error("CounterRowFormat should contain count")
	}
}

func TestSummaryTableBuilderDefaultWidth(t *testing.T) {
	// Disable colors for predictable testing
	colors.Disable()
	defer colors.Enable()

	// Test with width <= 0 (should use default)
	table := SummaryTableFormat("Test", 0)
	table.AddRow("Item", "Value")
	result := table.Build()

	if result == "" {
		t.Error("SummaryTableBuilder with default width should produce output")
	}
	if !strings.Contains(result, "Test") {
		t.Error("SummaryTableBuilder should contain title")
	}
}
