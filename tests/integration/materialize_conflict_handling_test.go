//go:build integration
// +build integration

package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tgaines/agent-smith/internal/testutil"
)

// TestMaterializeConflictHandling verifies Story-007 acceptance criteria:
// - If component already exists in target directory, skip silently if files are identical (hash match)
// - If files differ, error with message: "Component exists and differs. Use --force to overwrite"
// - `--force` flag allows overwriting existing components
// - Hash comparison determines if files are identical
// - Clear output indicates when components are skipped vs copied
func TestMaterializeConflictHandling(t *testing.T) {
	// Save current directory and restore it after test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(originalDir)
	})

	// Create temporary directories
	tempDir := testutil.CreateTempDir(t, "agent-smith-conflict-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	// Build the binary
	binaryPath := filepath.Join(tempDir, "agent-smith")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = filepath.Join(originalDir, "../..")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, string(output))
	}

	// Set up test environment
	baseDir := filepath.Join(tempDir, ".agent-smith")
	skillsDir := filepath.Join(baseDir, "skills")
	skillName := "conflict-test-skill"
	testSkillDir := filepath.Join(skillsDir, skillName)

	// Create directories
	mkdirErr := os.MkdirAll(testSkillDir, 0755)
	testutil.AssertNoError(t, mkdirErr, "Failed to create test skill directory")

	// Create test skill
	skillContent := "# Conflict Test Skill\n\nThis is a test skill for conflict handling."
	err = os.WriteFile(filepath.Join(testSkillDir, "SKILL.md"), []byte(skillContent), 0644)
	testutil.AssertNoError(t, err, "Failed to write SKILL.md")

	t.Logf("Created test skill at: %s", testSkillDir)

	// Create lock file
	lockFilePath := filepath.Join(baseDir, ".skill-lock.json")
	lockData := map[string]interface{}{
		"version": 3,
		"skills": map[string]interface{}{
			skillName: map[string]interface{}{
				"sourceUrl":    "github.com/test/conflict",
				"sourceType":   "github",
				"commitHash":   "abc123",
				"originalPath": "SKILL.md",
				"timestamp":    "2024-01-01T00:00:00Z",
			},
		},
	}
	lockJSON, _ := json.MarshalIndent(lockData, "", "  ")
	err = os.WriteFile(lockFilePath, lockJSON, 0644)
	testutil.AssertNoError(t, err, "Failed to write lock file")

	t.Run("Skip_When_Component_Already_Exists_And_Identical", func(t *testing.T) {
		// Setup: Create temporary project directory
		projectDir := testutil.CreateTempDir(t, "materialize-conflict-identical")

		// Create .opencode directory
		opencodeDir := filepath.Join(projectDir, ".opencode")
		if err := os.MkdirAll(opencodeDir, 0755); err != nil {
			t.Fatalf("Failed to create .opencode directory: %v", err)
		}

		// First materialization - should succeed
		cmd := exec.Command(binaryPath, "materialize", "skill", skillName, "--target", "opencode", "--project-dir", projectDir, "--verbose")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("First materialize output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("First materialize failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify component was materialized
		materializedPath := filepath.Join(opencodeDir, "skills", skillName, "SKILL.md")
		testutil.AssertFileExists(t, materializedPath)

		// Second materialization - should skip (identical files)
		cmd = exec.Command(binaryPath, "materialize", "skill", skillName, "--target", "opencode", "--project-dir", projectDir, "--verbose")
		output, err = cmd.CombinedOutput()
		outputStr = string(output)
		t.Logf("Second materialize output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Second materialize failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify output shows skip message
		if !strings.Contains(outputStr, "Skipped") || !strings.Contains(outputStr, "already exists and identical") {
			t.Errorf("Expected skip message for identical component, got: %s", outputStr)
		}

		t.Logf("✓ Story-007 AC#1: Identical component skipped silently")
	})

	t.Run("Error_When_Component_Exists_And_Differs_Without_Force", func(t *testing.T) {
		// Setup: Create temporary project directory
		projectDir := testutil.CreateTempDir(t, "materialize-conflict-differs")

		// Create .opencode directory
		opencodeDir := filepath.Join(projectDir, ".opencode")
		if err := os.MkdirAll(opencodeDir, 0755); err != nil {
			t.Fatalf("Failed to create .opencode directory: %v", err)
		}

		// First materialization - should succeed
		cmd := exec.Command(binaryPath, "materialize", "skill", skillName, "--target", "opencode", "--project-dir", projectDir, "--verbose")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("First materialize output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("First materialize failed: %v\nOutput: %s", err, outputStr)
		}

		// Modify the materialized component to make it differ
		materializedPath := filepath.Join(opencodeDir, "skills", skillName, "SKILL.md")
		modifiedContent := "# Modified Skill\n\nThis content has been changed."
		if err := os.WriteFile(materializedPath, []byte(modifiedContent), 0644); err != nil {
			t.Fatalf("Failed to modify materialized skill: %v", err)
		}

		// Second materialization without --force - should fail
		cmd = exec.Command(binaryPath, "materialize", "skill", skillName, "--target", "opencode", "--project-dir", projectDir, "--verbose")
		output, err = cmd.CombinedOutput()
		outputStr = string(output)
		t.Logf("Second materialize output:\n%s", outputStr)

		if err == nil {
			t.Fatalf("Expected materialize to fail when component differs, but it succeeded")
		}

		// Verify error message
		if !strings.Contains(outputStr, "already exists") && !strings.Contains(outputStr, "differs") {
			t.Errorf("Expected error about differing component, got: %s", outputStr)
		}

		if !strings.Contains(outputStr, "--force") {
			t.Errorf("Expected error to mention --force flag, got: %s", outputStr)
		}

		t.Logf("✓ Story-007 AC#2: Error when component differs without --force")
	})

	t.Run("Force_Flag_Overwrites_Existing_Component", func(t *testing.T) {
		// Setup: Create temporary project directory
		projectDir := testutil.CreateTempDir(t, "materialize-conflict-force")

		// Create .opencode directory
		opencodeDir := filepath.Join(projectDir, ".opencode")
		if err := os.MkdirAll(opencodeDir, 0755); err != nil {
			t.Fatalf("Failed to create .opencode directory: %v", err)
		}

		// First materialization - should succeed
		cmd := exec.Command(binaryPath, "materialize", "skill", skillName, "--target", "opencode", "--project-dir", projectDir, "--verbose")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("First materialize output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("First materialize failed: %v\nOutput: %s", err, outputStr)
		}

		// Modify the materialized component to make it differ
		materializedPath := filepath.Join(opencodeDir, "skills", skillName, "SKILL.md")
		modifiedContent := "# Modified Skill\n\nThis content has been changed."
		if err := os.WriteFile(materializedPath, []byte(modifiedContent), 0644); err != nil {
			t.Fatalf("Failed to modify materialized skill: %v", err)
		}

		// Read modified content to verify it's different
		modifiedData, err := os.ReadFile(materializedPath)
		if err != nil {
			t.Fatalf("Failed to read modified skill: %v", err)
		}
		if string(modifiedData) != modifiedContent {
			t.Fatalf("Modified content doesn't match expected")
		}

		// Second materialization with --force - should succeed and overwrite
		cmd = exec.Command(binaryPath, "materialize", "skill", skillName, "--target", "opencode", "--project-dir", projectDir, "--force", "--verbose")
		output, err = cmd.CombinedOutput()
		outputStr = string(output)
		t.Logf("Second materialize with --force output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Materialize with --force failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify output shows overwrite message
		if !strings.Contains(outputStr, "Overwriting") || !strings.Contains(outputStr, "--force") {
			t.Errorf("Expected overwrite message with --force, got: %s", outputStr)
		}

		// Verify the component was overwritten (content should match original)
		finalData, err := os.ReadFile(materializedPath)
		if err != nil {
			t.Fatalf("Failed to read final skill: %v", err)
		}
		if string(finalData) == modifiedContent {
			t.Errorf("Component was not overwritten - still contains modified content")
		}
		if !strings.Contains(string(finalData), "Conflict Test Skill") {
			t.Errorf("Component doesn't contain expected original content: %s", string(finalData))
		}

		t.Logf("✓ Story-007 AC#3: --force flag successfully overwrites existing component")
	})

	t.Run("Hash_Comparison_Determines_Identity", func(t *testing.T) {
		// Setup: Create temporary project directory
		projectDir := testutil.CreateTempDir(t, "materialize-conflict-hash")

		// Create .opencode directory
		opencodeDir := filepath.Join(projectDir, ".opencode")
		if err := os.MkdirAll(opencodeDir, 0755); err != nil {
			t.Fatalf("Failed to create .opencode directory: %v", err)
		}

		// First materialization
		cmd := exec.Command(binaryPath, "materialize", "skill", skillName, "--target", "opencode", "--project-dir", projectDir, "--verbose")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("First materialize failed: %v\nOutput: %s", err, string(output))
		}

		// Touch the file to change timestamp but not content
		materializedPath := filepath.Join(opencodeDir, "skills", skillName, "SKILL.md")
		originalContent, err := os.ReadFile(materializedPath)
		if err != nil {
			t.Fatalf("Failed to read materialized skill: %v", err)
		}

		// Write the same content back (changes timestamp, not content)
		if err := os.WriteFile(materializedPath, originalContent, 0644); err != nil {
			t.Fatalf("Failed to touch file: %v", err)
		}

		// Second materialization - should skip (hash is identical even though timestamp differs)
		cmd = exec.Command(binaryPath, "materialize", "skill", skillName, "--target", "opencode", "--project-dir", projectDir, "--verbose")
		output, err = cmd.CombinedOutput()
		outputStr := string(output)

		if err != nil {
			t.Fatalf("Second materialize failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify skip message (hash comparison should detect identical content)
		if !strings.Contains(outputStr, "Skipped") || !strings.Contains(outputStr, "identical") {
			t.Errorf("Expected skip due to identical hash, got: %s", outputStr)
		}

		t.Logf("✓ Story-007 AC#4: Hash comparison correctly identifies identical content")
	})
}
