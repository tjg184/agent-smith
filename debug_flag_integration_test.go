//go:build integration
// +build integration

package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestDebugFlag verifies that the --debug flag enables debug output
func TestDebugFlag(t *testing.T) {
	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", "agent-smith-test")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("agent-smith-test")

	tests := []struct {
		name           string
		args           []string
		expectDebug    bool
		expectVerbose  bool
		debugKeyword   string
		verboseKeyword string
	}{
		{
			name:           "no flags should have no debug output",
			args:           []string{"status"},
			expectDebug:    false,
			expectVerbose:  false,
			debugKeyword:   "[DEBUG]",
			verboseKeyword: "Current Configuration:",
		},
		{
			name:           "--debug flag should enable debug output",
			args:           []string{"--debug", "status"},
			expectDebug:    true,
			expectVerbose:  true,
			debugKeyword:   "[DEBUG]",
			verboseKeyword: "Current Configuration:",
		},
		{
			name:           "--verbose flag should enable verbose but not debug",
			args:           []string{"--verbose", "status"},
			expectDebug:    false,
			expectVerbose:  true,
			debugKeyword:   "[DEBUG]",
			verboseKeyword: "Current Configuration:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := exec.Command("./agent-smith-test", tt.args...)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			// Run the command (we expect it might fail, but we just care about output)
			_ = cmd.Run()

			output := stdout.String() + stderr.String()

			// Check for debug output
			hasDebug := strings.Contains(output, tt.debugKeyword)
			if hasDebug != tt.expectDebug {
				if tt.expectDebug {
					t.Errorf("Expected debug output containing '%s', but didn't find it", tt.debugKeyword)
				} else {
					t.Errorf("Expected no debug output, but found '%s'", tt.debugKeyword)
				}
			}

			// Check for verbose output
			hasVerbose := strings.Contains(output, tt.verboseKeyword)
			if hasVerbose != tt.expectVerbose {
				if tt.expectVerbose {
					t.Errorf("Expected verbose output containing '%s', but didn't find it", tt.verboseKeyword)
				} else {
					t.Errorf("Expected no verbose output, but found '%s'", tt.verboseKeyword)
				}
			}
		})
	}
}

// TestDebugFlagHelp verifies that --debug flag appears in help output
func TestDebugFlagHelp(t *testing.T) {
	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", "agent-smith-test")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("agent-smith-test")

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("./agent-smith-test", "--help")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to run help command: %v", err)
	}

	output := stdout.String() + stderr.String()

	// Check that --debug flag is documented in help
	if !strings.Contains(output, "--debug") {
		t.Error("Expected --debug flag to appear in help output")
	}

	// Check that the debug flag description is present
	if !strings.Contains(output, "debug") || !strings.Contains(output, "troubleshooting") {
		t.Error("Expected debug flag description to mention debug/troubleshooting")
	}
}
