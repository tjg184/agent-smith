package linker

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/tjg184/agent-smith/internal/detector"
	"github.com/tjg184/agent-smith/pkg/config"
	"github.com/tjg184/agent-smith/pkg/paths"
)

// TestPromptProfileSelection_ValidSelection tests that valid user input is handled correctly
func TestPromptProfileSelection_ValidSelection(t *testing.T) {
	// Create test matches
	matches := []ProfileMatch{
		{
			ProfileName: "test-profile-1",
			ProfilePath: "/path/to/profile1",
			IsActive:    true,
			SourceUrl:   "https://github.com/test/repo1",
		},
		{
			ProfileName: "test-profile-2",
			ProfilePath: "/path/to/profile2",
			IsActive:    false,
			SourceUrl:   "https://github.com/test/repo2",
		},
	}

	// Create linker (we just need a basic instance)
	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		t.Fatalf("Failed to get agents directory: %v", err)
	}

	tempTargetDir, err := os.MkdirTemp("", "agent-smith-target-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp target dir: %v", err)
	}
	defer os.RemoveAll(tempTargetDir)

	targets := []config.Target{
		&mockTarget{name: "test-target", baseDir: tempTargetDir},
	}

	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(agentsDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	// Test cases
	testCases := []struct {
		name           string
		input          string
		expectedPath   string
		expectedName   string
		expectError    bool
		errorSubstring string
	}{
		{
			name:         "Valid selection - first option",
			input:        "1\n",
			expectedPath: "/path/to/profile1",
			expectedName: "test-profile-1",
			expectError:  false,
		},
		{
			name:         "Valid selection - second option",
			input:        "2\n",
			expectedPath: "/path/to/profile2",
			expectedName: "test-profile-2",
			expectError:  false,
		},
		{
			name:           "Invalid selection - zero",
			input:          "0\n",
			expectError:    true,
			errorSubstring: "invalid selection",
		},
		{
			name:           "Invalid selection - out of range",
			input:          "3\n",
			expectError:    true,
			errorSubstring: "invalid selection",
		},
		{
			name:           "Invalid selection - negative",
			input:          "-1\n",
			expectError:    true,
			errorSubstring: "invalid selection",
		},
		{
			name:           "Invalid selection - non-numeric",
			input:          "abc\n",
			expectError:    true,
			errorSubstring: "invalid selection",
		},
		{
			name:           "Cancel with 'c'",
			input:          "c\n",
			expectError:    true,
			errorSubstring: "cancelled",
		},
		{
			name:           "Cancel with empty input",
			input:          "\n",
			expectError:    true,
			errorSubstring: "cancelled",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Mock stdin
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r

			// Write test input
			go func() {
				w.Write([]byte(tc.input))
				w.Close()
			}()

			// Capture stdout
			oldStdout := os.Stdout
			rOut, wOut, _ := os.Pipe()
			os.Stdout = wOut

			// Execute the prompt
			profilePath, profileName, err := linker.promptProfileSelection("skills", "test-skill", matches)

			// Restore stdin/stdout
			os.Stdin = oldStdin
			wOut.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, rOut)
			output := buf.String()

			// Verify results
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', but got no error", tc.errorSubstring)
				} else if !strings.Contains(err.Error(), tc.errorSubstring) {
					t.Errorf("Expected error containing '%s', got: %v", tc.errorSubstring, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if profilePath != tc.expectedPath {
					t.Errorf("Expected path '%s', got '%s'", tc.expectedPath, profilePath)
				}
				if profileName != tc.expectedName {
					t.Errorf("Expected name '%s', got '%s'", tc.expectedName, profileName)
				}
			}

			// Verify output contains required elements (for non-error cases)
			if !tc.expectError {
				if !strings.Contains(output, "test-skill") {
					t.Error("Output should contain component name")
				}
				if !strings.Contains(output, "test-profile-1") {
					t.Error("Output should contain first profile name")
				}
				if !strings.Contains(output, "test-profile-2") {
					t.Error("Output should contain second profile name")
				}
				if !strings.Contains(output, "https://github.com/test/repo1") {
					t.Error("Output should contain first profile source URL")
				}
				if !strings.Contains(output, "https://github.com/test/repo2") {
					t.Error("Output should contain second profile source URL")
				}
				if !strings.Contains(output, "(active)") {
					t.Error("Output should indicate which profile is active")
				}
			}
		})
	}
}

// TestPromptProfileSelection_AcceptanceCriteria verifies all Story-007 acceptance criteria
func TestPromptProfileSelection_AcceptanceCriteria(t *testing.T) {
	// This test validates all acceptance criteria for Story-007:
	// ✓ Interactive prompt lists all profiles containing the component
	// ✓ Prompt shows profile name and component source URL for each option
	// ✓ User can select by number (1, 2, 3, etc.)
	// ✓ Selection is validated before proceeding with link
	// ✓ User can cancel the prompt (Ctrl+C or empty input)
	// ✓ Active profile option is clearly indicated in the list

	matches := []ProfileMatch{
		{
			ProfileName: "github-user-repo1",
			ProfilePath: "/home/user/.agent-smith/profiles/github-user-repo1",
			IsActive:    true,
			SourceUrl:   "https://github.com/user/repo1",
		},
		{
			ProfileName: "github-org-repo2",
			ProfilePath: "/home/user/.agent-smith/profiles/github-org-repo2",
			IsActive:    false,
			SourceUrl:   "https://github.com/org/repo2",
		},
		{
			ProfileName: "local-workspace",
			ProfilePath: "/home/user/.agent-smith/profiles/local-workspace",
			IsActive:    false,
			SourceUrl:   "", // Local install has no source URL
		},
	}

	agentsDir, err := paths.GetAgentsDir()
	if err != nil {
		t.Fatalf("Failed to get agents directory: %v", err)
	}

	tempTargetDir, err := os.MkdirTemp("", "agent-smith-target-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp target dir: %v", err)
	}
	defer os.RemoveAll(tempTargetDir)

	targets := []config.Target{
		&mockTarget{name: "test-target", baseDir: tempTargetDir},
	}

	det := detector.NewRepositoryDetector()
	linker, err := NewComponentLinker(agentsDir, targets, det, nil)
	if err != nil {
		t.Fatalf("Failed to create linker: %v", err)
	}

	t.Run("AC1: Lists all profiles containing the component", func(t *testing.T) {
		oldStdin := os.Stdin
		r, w, _ := os.Pipe()
		os.Stdin = r

		go func() {
			w.Write([]byte("1\n"))
			w.Close()
		}()

		oldStdout := os.Stdout
		rOut, wOut, _ := os.Pipe()
		os.Stdout = wOut

		linker.promptProfileSelection("skills", "test-component", matches)

		os.Stdin = oldStdin
		wOut.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, rOut)
		output := buf.String()

		for i, match := range matches {
			expectedNumber := fmt.Sprintf("%d.", i+1)
			if !strings.Contains(output, expectedNumber) {
				t.Errorf("Output should contain numbered option '%s' for profile %s", expectedNumber, match.ProfileName)
			}
			if !strings.Contains(output, match.ProfileName) {
				t.Errorf("Output should contain profile name '%s'", match.ProfileName)
			}
		}
	})

	t.Run("AC2: Shows profile name and source URL", func(t *testing.T) {
		oldStdin := os.Stdin
		r, w, _ := os.Pipe()
		os.Stdin = r

		go func() {
			w.Write([]byte("1\n"))
			w.Close()
		}()

		oldStdout := os.Stdout
		rOut, wOut, _ := os.Pipe()
		os.Stdout = wOut

		linker.promptProfileSelection("agents", "api-handler", matches)

		os.Stdin = oldStdin
		wOut.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, rOut)
		output := buf.String()

		// Check profile names
		if !strings.Contains(output, "github-user-repo1") {
			t.Error("Output should contain profile name 'github-user-repo1'")
		}
		if !strings.Contains(output, "github-org-repo2") {
			t.Error("Output should contain profile name 'github-org-repo2'")
		}

		// Check source URLs
		if !strings.Contains(output, "https://github.com/user/repo1") {
			t.Error("Output should contain source URL for first profile")
		}
		if !strings.Contains(output, "https://github.com/org/repo2") {
			t.Error("Output should contain source URL for second profile")
		}
	})

	t.Run("AC3: User can select by number", func(t *testing.T) {
		oldStdin := os.Stdin
		r, w, _ := os.Pipe()
		os.Stdin = r

		go func() {
			w.Write([]byte("2\n"))
			w.Close()
		}()

		os.Stdout, _ = os.Open(os.DevNull) // Suppress output
		defer func() { os.Stdout = os.Stderr }()

		profilePath, profileName, err := linker.promptProfileSelection("skills", "test", matches)

		os.Stdin = oldStdin

		if err != nil {
			t.Errorf("Selection by number should work, got error: %v", err)
		}
		if profileName != "github-org-repo2" {
			t.Errorf("Expected profile 'github-org-repo2', got '%s'", profileName)
		}
		if profilePath != "/home/user/.agent-smith/profiles/github-org-repo2" {
			t.Errorf("Expected correct profile path, got '%s'", profilePath)
		}
	})

	t.Run("AC4: Selection is validated", func(t *testing.T) {
		testCases := []string{"0\n", "4\n", "-1\n", "999\n", "abc\n"}

		for _, input := range testCases {
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r

			go func() {
				w.Write([]byte(input))
				w.Close()
			}()

			os.Stdout, _ = os.Open(os.DevNull)
			defer func() { os.Stdout = os.Stderr }()

			_, _, err := linker.promptProfileSelection("skills", "test", matches)

			os.Stdin = oldStdin

			if err == nil {
				t.Errorf("Invalid input '%s' should be rejected", strings.TrimSpace(input))
			}
			if !strings.Contains(err.Error(), "invalid selection") {
				t.Errorf("Error should mention 'invalid selection', got: %v", err)
			}
		}
	})

	t.Run("AC5: User can cancel the prompt", func(t *testing.T) {
		cancelInputs := []string{"c\n", "C\n", "\n"}

		for _, input := range cancelInputs {
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r

			go func() {
				w.Write([]byte(input))
				w.Close()
			}()

			os.Stdout, _ = os.Open(os.DevNull)
			defer func() { os.Stdout = os.Stderr }()

			_, _, err := linker.promptProfileSelection("skills", "test", matches)

			os.Stdin = oldStdin

			if err == nil {
				t.Errorf("Cancel input '%s' should return an error", strings.TrimSpace(input))
			}
			if !strings.Contains(err.Error(), "cancelled") {
				t.Errorf("Error should mention 'cancelled', got: %v", err)
			}
		}
	})

	t.Run("AC6: Active profile is clearly indicated", func(t *testing.T) {
		oldStdin := os.Stdin
		r, w, _ := os.Pipe()
		os.Stdin = r

		go func() {
			w.Write([]byte("1\n"))
			w.Close()
		}()

		oldStdout := os.Stdout
		rOut, wOut, _ := os.Pipe()
		os.Stdout = wOut

		linker.promptProfileSelection("skills", "test-component", matches)

		os.Stdin = oldStdin
		wOut.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, rOut)
		output := buf.String()

		// Check that "(active)" indicator appears for the active profile
		if !strings.Contains(output, "(active)") {
			t.Error("Output should contain '(active)' indicator")
		}

		// Verify it appears on the same line as the active profile name
		lines := strings.Split(output, "\n")
		foundActiveIndicator := false
		for _, line := range lines {
			if strings.Contains(line, "github-user-repo1") && strings.Contains(line, "(active)") {
				foundActiveIndicator = true
				break
			}
		}
		if !foundActiveIndicator {
			t.Error("Active indicator should appear on the same line as the active profile name")
		}
	})
}
