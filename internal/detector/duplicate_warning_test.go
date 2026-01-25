package detector_test

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tgaines/agent-smith/internal/detector"
)

// TestDuplicateComponentWarnings tests Story-005: Clear warnings when duplicate component names are detected
func TestDuplicateComponentWarnings(t *testing.T) {
	tests := []struct {
		name               string
		filesToCreate      map[string]string
		expectedDuplicates int
		expectedComponents int
		duplicateNames     []string
		description        string
	}{
		{
			name: "duplicate-skills-same-name",
			filesToCreate: map[string]string{
				"frontend/python-dev/SKILL.md": "# Python Dev Skill (Frontend)",
				"backend/python-dev/SKILL.md":  "# Python Dev Skill (Backend)",
				"lib/python-dev/SKILL.md":      "# Python Dev Skill (Lib)",
			},
			expectedDuplicates: 1, // One duplicate name: python-dev
			expectedComponents: 1, // Only first occurrence should be included
			duplicateNames:     []string{"python-dev"},
			description:        "Multiple skills with same directory name should trigger duplicate warning",
		},
		{
			name: "duplicate-agents-same-filename",
			filesToCreate: map[string]string{
				"agents/chatbot.md":        "# Chatbot Agent v1",
				"agents/v2/chatbot.md":     "# Chatbot Agent v2",
				"agents/legacy/chatbot.md": "# Chatbot Agent Legacy",
			},
			expectedDuplicates: 1, // One duplicate name: chatbot
			expectedComponents: 1, // Only first occurrence should be included
			duplicateNames:     []string{"chatbot"},
			description:        "Multiple agents with same filename should trigger duplicate warning",
		},
		{
			name: "duplicate-commands-same-filename",
			filesToCreate: map[string]string{
				"commands/deploy.md":         "# Deploy Command v1",
				"commands/staging/deploy.md": "# Deploy Command Staging",
			},
			expectedDuplicates: 1,
			expectedComponents: 1,
			duplicateNames:     []string{"deploy"},
			description:        "Multiple commands with same filename should trigger duplicate warning",
		},
		{
			name: "mixed-duplicates-multiple-types",
			filesToCreate: map[string]string{
				// Duplicate skills
				"skills/testing/SKILL.md":    "# Testing Skill v1",
				"skills/v2/testing/SKILL.md": "# Testing Skill v2",
				// Duplicate agents
				"agents/helper.md":    "# Helper Agent v1",
				"agents/v2/helper.md": "# Helper Agent v2",
				// Unique command (no duplicate)
				"commands/build.md": "# Build Command",
			},
			expectedDuplicates: 2, // testing (skill) and helper (agent)
			expectedComponents: 3, // testing (1), helper (1), build (1)
			duplicateNames:     []string{"testing", "helper"},
			description:        "Multiple component types with duplicates should all be detected",
		},
		{
			name: "no-duplicates",
			filesToCreate: map[string]string{
				"skills/python/SKILL.md": "# Python Skill",
				"skills/golang/SKILL.md": "# Golang Skill",
				"agents/chatbot.md":      "# Chatbot Agent",
				"commands/deploy.md":     "# Deploy Command",
			},
			expectedDuplicates: 0,
			expectedComponents: 4,
			duplicateNames:     []string{},
			description:        "No duplicates should result in no warnings",
		},
		{
			name: "triple-duplicate",
			filesToCreate: map[string]string{
				"v1/api/SKILL.md": "# API Skill v1",
				"v2/api/SKILL.md": "# API Skill v2",
				"v3/api/SKILL.md": "# API Skill v3",
			},
			expectedDuplicates: 1, // One duplicate name with 3 occurrences
			expectedComponents: 1, // Only first should be included
			duplicateNames:     []string{"api"},
			description:        "Three or more duplicates should all be tracked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir, err := os.MkdirTemp("", "duplicate-warning-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Create test files
			for filePath, content := range tt.filesToCreate {
				fullPath := filepath.Join(tempDir, filePath)
				dir := filepath.Dir(fullPath)

				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("Failed to create directory %s: %v", dir, err)
				}

				if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create file %s: %v", fullPath, err)
				}
			}

			// Capture log output
			var logBuf bytes.Buffer
			log.SetOutput(&logBuf)
			defer log.SetOutput(os.Stderr) // Restore default

			// Create detector and find components
			detector := detector.NewRepositoryDetector()
			components, err := detector.DetectComponentsInRepo(tempDir)
			if err != nil {
				t.Fatalf("Failed to detect components: %v", err)
			}

			// Get log output
			logOutput := logBuf.String()

			// Verify component count (should only include first occurrence of duplicates)
			if len(components) != tt.expectedComponents {
				t.Errorf("Expected %d unique components, got %d", tt.expectedComponents, len(components))
				for i, comp := range components {
					t.Logf("  Component %d: %s (%s) - %s", i+1, comp.Name, comp.Type, comp.Path)
				}
			}

			// Verify duplicate warnings are present in logs
			if tt.expectedDuplicates > 0 {
				// Check for warning indicators
				if !strings.Contains(logOutput, "WARNING") {
					t.Errorf("Expected WARNING in log output for duplicates, but not found")
					t.Logf("Log output: %s", logOutput)
				}

				if !strings.Contains(logOutput, "Duplicate") {
					t.Errorf("Expected 'Duplicate' in log output, but not found")
				}

				// Verify each expected duplicate name appears in warnings
				for _, dupName := range tt.duplicateNames {
					if !strings.Contains(logOutput, dupName) {
						t.Errorf("Expected duplicate name '%s' in warnings, but not found", dupName)
					}
				}

				// Verify "WILL BE SKIPPED" message appears
				if !strings.Contains(logOutput, "WILL BE SKIPPED") {
					t.Errorf("Expected 'WILL BE SKIPPED' message in warnings")
				}
			} else {
				// No duplicates expected - verify no duplicate warnings
				if strings.Contains(logOutput, "Duplicate component") || strings.Contains(logOutput, "Duplicate agent") || strings.Contains(logOutput, "Duplicate command") {
					t.Errorf("Unexpected duplicate warning in log output when no duplicates expected")
					t.Logf("Log output: %s", logOutput)
				}
			}

			t.Logf("Test: %s", tt.description)
			t.Logf("Components detected: %d (expected: %d)", len(components), tt.expectedComponents)
			if tt.expectedDuplicates > 0 {
				t.Logf("Duplicate warnings found in logs: ✓")
			}
		})
	}
}

// TestDuplicateWarningFormat tests the format and clarity of duplicate warnings
func TestDuplicateWarningFormat(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "warning-format-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create duplicate components
	testFiles := map[string]string{
		"v1/testing/SKILL.md": "# Testing Skill v1",
		"v2/testing/SKILL.md": "# Testing Skill v2",
	}

	for filePath, content := range testFiles {
		fullPath := filepath.Join(tempDir, filePath)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Capture both stdout and log output
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stderr)

	// Capture stdout (for the summary display)
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	// Create detector and find components
	detector := detector.NewRepositoryDetector()
	_, err = detector.DetectComponentsInRepo(tempDir)
	if err != nil {
		t.Fatalf("Failed to detect components: %v", err)
	}

	// Close writer and read stdout
	w.Close()
	var stdoutBuf bytes.Buffer
	stdoutBuf.ReadFrom(r)
	stdoutOutput := stdoutBuf.String()

	// Restore stdout
	os.Stdout = oldStdout

	// Get log output
	logOutput := logBuf.String()

	// Verify warning format in logs
	t.Run("log-warning-format", func(t *testing.T) {
		requiredElements := []string{
			"⚠️",                // Warning emoji
			"WARNING",           // Warning keyword
			"Duplicate",         // Duplicate keyword
			"testing",           // Component name
			"WILL BE SKIPPED",   // Action indicator
			"First occurrence:", // First occurrence indicator
		}

		for _, element := range requiredElements {
			if !strings.Contains(logOutput, element) {
				t.Errorf("Log warning missing required element: '%s'", element)
			}
		}

		t.Logf("Log warning format validated successfully")
	})

	// Verify summary format in stdout
	t.Run("summary-format", func(t *testing.T) {
		requiredElements := []string{
			"WARNING",                   // Warning keyword
			"Duplicate Component Names", // Clear title
			"testing",                   // Component name
			"USED",                      // First occurrence indicator
			"SKIPPED",                   // Duplicate indicator
			"Resolution Required",       // Action section
			"Rename or remove",          // Resolution guidance
		}

		for _, element := range requiredElements {
			if !strings.Contains(stdoutOutput, element) {
				t.Errorf("Summary missing required element: '%s'", element)
				t.Logf("Stdout output:\n%s", stdoutOutput)
			}
		}

		// Check for box drawing characters (visual formatting)
		if !strings.Contains(stdoutOutput, "╔") || !strings.Contains(stdoutOutput, "╚") {
			t.Errorf("Summary should have box drawing characters for visual emphasis")
		}

		t.Logf("Summary format validated successfully")
	})

	t.Logf("=== Log Output ===\n%s", logOutput)
	t.Logf("=== Stdout Output ===\n%s", stdoutOutput)
}

// TestDuplicateResolutionGuidance tests that warnings provide actionable guidance
func TestDuplicateResolutionGuidance(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "resolution-guidance-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create duplicate components
	testFiles := map[string]string{
		"frontend/api/SKILL.md": "# API Skill Frontend",
		"backend/api/SKILL.md":  "# API Skill Backend",
	}

	for filePath, content := range testFiles {
		fullPath := filepath.Join(tempDir, filePath)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create detector and find components
	detector := detector.NewRepositoryDetector()
	components, err := detector.DetectComponentsInRepo(tempDir)
	if err != nil {
		t.Fatalf("Failed to detect components: %v", err)
	}

	// Close writer and read stdout
	w.Close()
	var stdoutBuf bytes.Buffer
	stdoutBuf.ReadFrom(r)
	stdoutOutput := stdoutBuf.String()

	// Restore stdout
	os.Stdout = oldStdout

	// Verify only first occurrence was included
	if len(components) != 1 {
		t.Errorf("Expected 1 component (first occurrence only), got %d", len(components))
	}

	// Verify guidance is actionable
	actionableGuidance := []string{
		"Resolution Required",
		"first occurrence",
		"USED",
		"SKIPPED",
		"Rename or remove",
	}

	for _, guidance := range actionableGuidance {
		if !strings.Contains(stdoutOutput, guidance) {
			t.Errorf("Missing actionable guidance: '%s'", guidance)
		}
	}

	// Verify both file paths are listed
	if !strings.Contains(stdoutOutput, "frontend/api") {
		t.Errorf("First occurrence path should be listed")
	}

	if !strings.Contains(stdoutOutput, "backend/api") {
		t.Errorf("Duplicate occurrence path should be listed")
	}

	t.Logf("Resolution guidance validated successfully")
	t.Logf("Output:\n%s", stdoutOutput)
}
