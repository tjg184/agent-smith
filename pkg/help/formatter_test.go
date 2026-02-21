package help

import (
	"strings"
	"testing"

	"github.com/tgaines/agent-smith/pkg/colors"
)

// TestIsSectionHeader tests section header detection
func TestIsSectionHeader(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"USAGE header", "USAGE:", true},
		{"EXAMPLES header", "EXAMPLES:", true},
		{"FLAGS header", "FLAGS:", true},
		{"QUICK START header", "QUICK START:", true},
		{"COMMAND GROUPS header", "COMMAND GROUPS:", true},
		{"REPOSITORY URL FORMATS header", "REPOSITORY URL FORMATS:", true},
		{"Not a header", "Not a header", false},
		{"Lowercase header", "usage:", false},
		{"No colon", "USAGE", false},
		{"Extra text after colon", "USAGE: extra", false},
		{"Indented header", "  USAGE:", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSectionHeader(tt.input)
			if result != tt.expected {
				t.Errorf("isSectionHeader(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestIsComment tests comment detection
func TestIsComment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Simple comment", "# This is a comment", true},
		{"Indented comment", "  # Indented comment", true},
		{"Comment with code", "# agent-smith install", true},
		{"Not a comment", "agent-smith install", false},
		{"Hashtag in middle", "This is # not a comment", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isComment(tt.input)
			if result != tt.expected {
				t.Errorf("isComment(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestIsCommandExample tests command example detection
func TestIsCommandExample(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Simple command", "agent-smith install", true},
		{"Command with args", "agent-smith install skill owner/repo", true},
		{"Indented command", "  agent-smith link all", true},
		{"Command in description", "Run agent-smith to start", true},
		{"Not a command", "This is text", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCommandExample(tt.input)
			if result != tt.expected {
				t.Errorf("isCommandExample(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestHasURL tests URL detection
func TestHasURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"HTTPS URL", "https://github.com/owner/repo", true},
		{"HTTP URL", "http://example.com", true},
		{"Git SSH URL", "git@github.com:owner/repo.git", true},
		{"GitHub shorthand", "owner/repo", true},
		{"No URL", "This is text", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasURL(tt.input)
			if result != tt.expected {
				t.Errorf("hasURL(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestGetIndentation tests indentation extraction
func TestGetIndentation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"No indentation", "text", ""},
		{"Two spaces", "  text", "  "},
		{"Four spaces", "    text", "    "},
		{"Tab indentation", "\ttext", "\t"},
		{"Mixed spaces and tabs", " \t text", " \t "},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getIndentation(tt.input)
			if result != tt.expected {
				t.Errorf("getIndentation(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestColorizeSection tests section header colorization
func TestColorizeSection(t *testing.T) {
	// Temporarily disable colors for predictable testing
	colors.Disable()
	defer colors.Enable()

	input := "USAGE:"
	result := colorizeSection(input)

	// When colors are disabled, should return unchanged
	if result != input {
		t.Errorf("colorizeSection(%q) with colors disabled = %q, want %q", input, result, input)
	}
}

// TestColorizeComment tests comment colorization
func TestColorizeComment(t *testing.T) {
	colors.Disable()
	defer colors.Enable()

	input := "# This is a comment"
	result := colorizeComment(input)

	// When colors are disabled, should return unchanged
	if result != input {
		t.Errorf("colorizeComment(%q) with colors disabled = %q, want %q", input, result, input)
	}
}

// TestColorizeCommand tests command colorization
func TestColorizeCommand(t *testing.T) {
	colors.Disable()
	defer colors.Enable()

	tests := []struct {
		name     string
		input    string
		contains []string // Should contain these substrings
	}{
		{
			name:     "Simple command",
			input:    "agent-smith install",
			contains: []string{"agent-smith", "install"},
		},
		{
			name:     "Command with parameters",
			input:    "agent-smith install skill <repo> <name>",
			contains: []string{"agent-smith", "<repo>", "<name>"},
		},
		{
			name:     "Indented command",
			input:    "  agent-smith link all",
			contains: []string{"agent-smith", "link", "all"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colorizeCommand(tt.input)
			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("colorizeCommand(%q) = %q, should contain %q", tt.input, result, substr)
				}
			}

			// Verify indentation is preserved
			originalIndent := getIndentation(tt.input)
			resultIndent := getIndentation(result)
			if originalIndent != resultIndent {
				t.Errorf("colorizeCommand(%q) lost indentation: got %q, want %q", tt.input, resultIndent, originalIndent)
			}
		})
	}
}

// TestColorizeURLs tests URL colorization
func TestColorizeURLs(t *testing.T) {
	colors.Disable()
	defer colors.Enable()

	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "HTTPS URL",
			input:    "Full GitHub URL: https://github.com/owner/repo",
			contains: []string{"https://github.com/owner/repo"},
		},
		{
			name:     "GitHub shorthand",
			input:    "GitHub shorthand: owner/repo",
			contains: []string{"owner/repo"},
		},
		{
			name:     "Git SSH",
			input:    "SSH URL: git@github.com:owner/repo.git",
			contains: []string{"git@github.com:owner/repo.git"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colorizeURLs(tt.input)
			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("colorizeURLs(%q) = %q, should contain %q", tt.input, result, substr)
				}
			}
		})
	}
}

// TestColorizeLine tests line-by-line colorization
func TestColorizeLine(t *testing.T) {
	colors.Disable()
	defer colors.Enable()

	tests := []struct {
		name  string
		input string
		// Just verify it doesn't crash and returns something
	}{
		{"Section header", "USAGE:"},
		{"Comment", "# Install a skill"},
		{"Command", "agent-smith install skill <repo>"},
		{"URL", "https://github.com/owner/repo"},
		{"Plain text", "This is plain text"},
		{"Empty line", ""},
		{"Comment with command", "# agent-smith install"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colorizeLine(tt.input)
			// Basic sanity check - result should not be empty if input wasn't
			if tt.input != "" && result == "" {
				t.Errorf("colorizeLine(%q) returned empty string", tt.input)
			}
		})
	}
}

// TestColorizeText tests the main colorization function
func TestColorizeText(t *testing.T) {
	colors.Disable()
	defer colors.Enable()

	input := `USAGE:
  agent-smith install skill <repository-url> <skill-name>

EXAMPLES:
  # Install a specific skill from GitHub
  agent-smith install skill openai/cookbook gpt-skill

  # Install from a full URL
  agent-smith install skill https://github.com/example/repo my-skill`

	result := ColorizeText(input)

	// When colors are disabled, should return unchanged
	if result != input {
		t.Errorf("ColorizeText() with colors disabled modified the text")
	}

	// Should preserve line count
	inputLines := strings.Split(input, "\n")
	resultLines := strings.Split(result, "\n")
	if len(inputLines) != len(resultLines) {
		t.Errorf("ColorizeText() changed line count: got %d, want %d", len(resultLines), len(inputLines))
	}
}

// TestColorizeTextWithColorsEnabled tests colorization with colors enabled
func TestColorizeTextWithColorsEnabled(t *testing.T) {
	// Note: When colors are enabled, we can't test exact output due to ANSI codes
	// but we can test that the function runs without errors
	colors.Enable()
	defer colors.Disable()

	input := `USAGE:
  agent-smith install all <repository-url>

EXAMPLES:
  # Install all components
  agent-smith install all owner/repo`

	result := ColorizeText(input)

	// Basic sanity checks
	if result == "" {
		t.Error("ColorizeText() with colors enabled returned empty string")
	}

	// Should contain the original text (even if with color codes)
	if !strings.Contains(result, "agent-smith") {
		t.Error("ColorizeText() lost content")
	}
}

// TestPreserveIndentation tests that indentation is preserved during colorization
func TestPreserveIndentation(t *testing.T) {
	colors.Disable()
	defer colors.Enable()

	tests := []struct {
		name  string
		input string
	}{
		{"No indent", "agent-smith install"},
		{"Two spaces", "  agent-smith install"},
		{"Four spaces", "    agent-smith install"},
		{"Six spaces", "      agent-smith install"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colorizeLine(tt.input)
			originalIndent := getIndentation(tt.input)
			resultIndent := getIndentation(result)

			if originalIndent != resultIndent {
				t.Errorf("Indentation not preserved: got %q, want %q", resultIndent, originalIndent)
			}
		})
	}
}

// TestMultiPatternLine tests lines with multiple patterns
func TestMultiPatternLine(t *testing.T) {
	colors.Disable()
	defer colors.Enable()

	input := "# Install from GitHub: agent-smith install <repo>"
	result := colorizeMultiPattern(input)

	// Should contain all parts
	if !strings.Contains(result, "#") {
		t.Error("Lost comment marker")
	}
	if !strings.Contains(result, "agent-smith") {
		t.Error("Lost command")
	}
	if !strings.Contains(result, "<repo>") {
		t.Error("Lost parameter")
	}
}
