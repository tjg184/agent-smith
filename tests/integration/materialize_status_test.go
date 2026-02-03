//go:build integration
// +build integration

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tgaines/agent-smith/internal/testutil"
)

// TestMaterializeStatus tests the materialize status command
func TestMaterializeStatus(t *testing.T) {
	// Save current directory and restore it after test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(originalDir)
	})

	// Create temporary directories
	tempDir := testutil.CreateTempDir(t, "agent-smith-status-*")
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

	t.Run("AC1: Shows in sync for freshly materialized components", func(t *testing.T) {
		// Set up test environment
		baseDir := filepath.Join(tempDir, ".agent-smith")
		skillName := "test-skill-ac1"
		skillsDir := filepath.Join(baseDir, "skills", skillName)

		// Create test skill
		err := os.MkdirAll(skillsDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create skill directory")

		skillContent := `---
name: test-skill-ac1
version: 1.0.0
---
# Test Skill AC1
A test skill for status checking.
`
		err = os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte(skillContent), 0644)
		testutil.AssertNoError(t, err, "Failed to write SKILL.md")

		// Create lock file
		lockFilePath := filepath.Join(baseDir, ".skill-lock.json")
		lockData := map[string]interface{}{
			"version": 3,
			"skills": map[string]interface{}{
				skillName: map[string]interface{}{
					"source":       "test-repo",
					"sourceType":   "github",
					"sourceUrl":    "https://github.com/test/repo",
					"commitHash":   "abc123def456",
					"originalPath": fmt.Sprintf("skills/%s/SKILL.md", skillName),
					"installedAt":  "2024-01-15T10:30:00Z",
				},
			},
		}
		lockJSON, err := json.MarshalIndent(lockData, "", "  ")
		testutil.AssertNoError(t, err, "Failed to marshal lock data")
		err = os.WriteFile(lockFilePath, lockJSON, 0644)
		testutil.AssertNoError(t, err, "Failed to write lock file")

		// Create project directory
		projectDir := filepath.Join(tempDir, "test-project-ac1")
		opencodeDir := filepath.Join(projectDir, ".opencode")
		err = os.MkdirAll(opencodeDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create project directory")

		// Change to project directory
		err = os.Chdir(projectDir)
		testutil.AssertNoError(t, err, "Failed to change to project directory")

		// Materialize the skill
		cmd := exec.Command(binaryPath, "materialize", "skill", skillName, "--target", "opencode")
		output, err := cmd.CombinedOutput()
		testutil.AssertNoError(t, err, fmt.Sprintf("Failed to materialize skill: %s", string(output)))

		// Run status command
		cmd = exec.Command(binaryPath, "materialize", "status", "--target", "opencode")
		output, err = cmd.CombinedOutput()
		testutil.AssertNoError(t, err, fmt.Sprintf("Failed to run status command: %s", string(output)))

		outputStr := string(output)
		t.Logf("Status output:\n%s", outputStr)

		// Verify output contains "in sync"
		if !strings.Contains(outputStr, "in sync") {
			t.Errorf("Expected 'in sync' in output, got: %s", outputStr)
		}

		// Verify output contains green checkmark symbol
		if !strings.Contains(outputStr, "✓") {
			t.Errorf("Expected green checkmark (✓) in output, got: %s", outputStr)
		}

		// Verify summary shows 1 in sync
		if !strings.Contains(outputStr, "1 in sync") {
			t.Errorf("Expected summary to show '1 in sync', got: %s", outputStr)
		}

		t.Log("✓ AC1: Shows 'in sync' for freshly materialized components")
	})

	t.Run("AC2: Shows out of sync after source is updated", func(t *testing.T) {
		// Set up test environment
		baseDir := filepath.Join(tempDir, ".agent-smith")
		skillName := "test-skill-ac2"
		skillsDir := filepath.Join(baseDir, "skills", skillName)

		// Create test skill
		err := os.MkdirAll(skillsDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create skill directory")

		skillContent := `---
name: test-skill-ac2
version: 1.0.0
---
# Test Skill AC2
Original content.
`
		err = os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte(skillContent), 0644)
		testutil.AssertNoError(t, err, "Failed to write SKILL.md")

		// Create lock file
		lockFilePath := filepath.Join(baseDir, ".skill-lock.json")
		lockData := map[string]interface{}{
			"version": 3,
			"skills": map[string]interface{}{
				skillName: map[string]interface{}{
					"source":       "test-repo",
					"sourceType":   "github",
					"sourceUrl":    "https://github.com/test/repo",
					"commitHash":   "abc123def456",
					"originalPath": fmt.Sprintf("skills/%s/SKILL.md", skillName),
					"installedAt":  "2024-01-15T10:30:00Z",
				},
			},
		}
		lockJSON, err := json.MarshalIndent(lockData, "", "  ")
		testutil.AssertNoError(t, err, "Failed to marshal lock data")
		err = os.WriteFile(lockFilePath, lockJSON, 0644)
		testutil.AssertNoError(t, err, "Failed to write lock file")

		// Create project directory
		projectDir := filepath.Join(tempDir, "test-project-ac2")
		opencodeDir := filepath.Join(projectDir, ".opencode")
		err = os.MkdirAll(opencodeDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create project directory")

		// Change to project directory
		err = os.Chdir(projectDir)
		testutil.AssertNoError(t, err, "Failed to change to project directory")

		// Materialize the skill
		cmd := exec.Command(binaryPath, "materialize", "skill", skillName, "--target", "opencode")
		output, err := cmd.CombinedOutput()
		testutil.AssertNoError(t, err, fmt.Sprintf("Failed to materialize skill: %s", string(output)))

		// Modify the source skill to make it out of sync
		updatedContent := `---
name: test-skill-ac2
version: 2.0.0
---
# Test Skill AC2
Updated content with new features.
`
		err = os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte(updatedContent), 0644)
		testutil.AssertNoError(t, err, "Failed to update SKILL.md")

		// Run status command
		cmd = exec.Command(binaryPath, "materialize", "status", "--target", "opencode")
		output, err = cmd.CombinedOutput()
		testutil.AssertNoError(t, err, fmt.Sprintf("Failed to run status command: %s", string(output)))

		outputStr := string(output)
		t.Logf("Status output:\n%s", outputStr)

		// Verify output contains "out of sync"
		if !strings.Contains(outputStr, "out of sync") {
			t.Errorf("Expected 'out of sync' in output, got: %s", outputStr)
		}

		// Verify output contains warning symbol
		if !strings.Contains(outputStr, "⚠") {
			t.Errorf("Expected warning symbol (⚠) in output, got: %s", outputStr)
		}

		// Verify summary shows 1 out of sync
		if !strings.Contains(outputStr, "1 out of sync") {
			t.Errorf("Expected summary to show '1 out of sync', got: %s", outputStr)
		}

		t.Log("✓ AC2: Shows 'out of sync' after source is updated")
	})

	t.Run("AC3: Shows source missing when component is uninstalled", func(t *testing.T) {
		// Set up test environment
		baseDir := filepath.Join(tempDir, ".agent-smith")
		skillName := "test-skill-ac3"
		skillsDir := filepath.Join(baseDir, "skills", skillName)

		// Create test skill
		err := os.MkdirAll(skillsDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create skill directory")

		skillContent := `---
name: test-skill-ac3
version: 1.0.0
---
# Test Skill AC3
Will be removed.
`
		err = os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte(skillContent), 0644)
		testutil.AssertNoError(t, err, "Failed to write SKILL.md")

		// Create lock file
		lockFilePath := filepath.Join(baseDir, ".skill-lock.json")
		lockData := map[string]interface{}{
			"version": 3,
			"skills": map[string]interface{}{
				skillName: map[string]interface{}{
					"source":       "test-repo",
					"sourceType":   "github",
					"sourceUrl":    "https://github.com/test/repo",
					"commitHash":   "abc123def456",
					"originalPath": fmt.Sprintf("skills/%s/SKILL.md", skillName),
					"installedAt":  "2024-01-15T10:30:00Z",
				},
			},
		}
		lockJSON, err := json.MarshalIndent(lockData, "", "  ")
		testutil.AssertNoError(t, err, "Failed to marshal lock data")
		err = os.WriteFile(lockFilePath, lockJSON, 0644)
		testutil.AssertNoError(t, err, "Failed to write lock file")

		// Create project directory
		projectDir := filepath.Join(tempDir, "test-project-ac3")
		opencodeDir := filepath.Join(projectDir, ".opencode")
		err = os.MkdirAll(opencodeDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create project directory")

		// Change to project directory
		err = os.Chdir(projectDir)
		testutil.AssertNoError(t, err, "Failed to change to project directory")

		// Materialize the skill
		cmd := exec.Command(binaryPath, "materialize", "skill", skillName, "--target", "opencode")
		output, err := cmd.CombinedOutput()
		testutil.AssertNoError(t, err, fmt.Sprintf("Failed to materialize skill: %s", string(output)))

		// Remove the source skill to simulate uninstallation
		err = os.RemoveAll(skillsDir)
		testutil.AssertNoError(t, err, "Failed to remove source skill")

		// Run status command
		cmd = exec.Command(binaryPath, "materialize", "status", "--target", "opencode")
		output, err = cmd.CombinedOutput()
		testutil.AssertNoError(t, err, fmt.Sprintf("Failed to run status command: %s", string(output)))

		outputStr := string(output)
		t.Logf("Status output:\n%s", outputStr)

		// Verify output contains "source missing"
		if !strings.Contains(outputStr, "source missing") {
			t.Errorf("Expected 'source missing' in output, got: %s", outputStr)
		}

		// Verify output contains red X symbol
		if !strings.Contains(outputStr, "✗") {
			t.Errorf("Expected red X (✗) in output, got: %s", outputStr)
		}

		// Verify summary shows 1 source missing
		if !strings.Contains(outputStr, "1 source missing") {
			t.Errorf("Expected summary to show '1 source missing', got: %s", outputStr)
		}

		t.Log("✓ AC3: Shows 'source missing' when component is uninstalled")
	})

	t.Run("AC4: Respects --target flag to check specific target only", func(t *testing.T) {
		// Set up test environment
		baseDir := filepath.Join(tempDir, ".agent-smith")
		skillName := "test-skill-ac4"
		skillsDir := filepath.Join(baseDir, "skills", skillName)

		// Create test skill
		err := os.MkdirAll(skillsDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create skill directory")

		skillContent := `---
name: test-skill-ac4
version: 1.0.0
---
# Test Skill AC4
For target testing.
`
		err = os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte(skillContent), 0644)
		testutil.AssertNoError(t, err, "Failed to write SKILL.md")

		// Create lock file
		lockFilePath := filepath.Join(baseDir, ".skill-lock.json")
		lockData := map[string]interface{}{
			"version": 3,
			"skills": map[string]interface{}{
				skillName: map[string]interface{}{
					"source":       "test-repo",
					"sourceType":   "github",
					"sourceUrl":    "https://github.com/test/repo",
					"commitHash":   "abc123def456",
					"originalPath": fmt.Sprintf("skills/%s/SKILL.md", skillName),
					"installedAt":  "2024-01-15T10:30:00Z",
				},
			},
		}
		lockJSON, err := json.MarshalIndent(lockData, "", "  ")
		testutil.AssertNoError(t, err, "Failed to marshal lock data")
		err = os.WriteFile(lockFilePath, lockJSON, 0644)
		testutil.AssertNoError(t, err, "Failed to write lock file")

		// Create project directory
		projectDir := filepath.Join(tempDir, "test-project-ac4")
		opencodeDir := filepath.Join(projectDir, ".opencode")
		claudeDir := filepath.Join(projectDir, ".claude")
		err = os.MkdirAll(opencodeDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create .opencode directory")
		err = os.MkdirAll(claudeDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create .claude directory")

		// Change to project directory
		err = os.Chdir(projectDir)
		testutil.AssertNoError(t, err, "Failed to change to project directory")

		// Materialize to opencode only
		cmd := exec.Command(binaryPath, "materialize", "skill", skillName, "--target", "opencode")
		output, err := cmd.CombinedOutput()
		testutil.AssertNoError(t, err, fmt.Sprintf("Failed to materialize skill: %s", string(output)))

		// Run status command for opencode target only
		cmd = exec.Command(binaryPath, "materialize", "status", "--target", "opencode")
		output, err = cmd.CombinedOutput()
		testutil.AssertNoError(t, err, fmt.Sprintf("Failed to run status command: %s", string(output)))

		outputStr := string(output)
		t.Logf("Status output for opencode:\n%s", outputStr)

		// Verify output contains opencode reference
		if !strings.Contains(outputStr, "OpenCode") && !strings.Contains(outputStr, ".opencode") {
			t.Errorf("Expected OpenCode reference in output, got: %s", outputStr)
		}

		// Verify output does NOT contain claude reference
		if strings.Contains(outputStr, "Claude") || strings.Contains(outputStr, ".claude") {
			t.Errorf("Expected NO Claude reference when using --target opencode, got: %s", outputStr)
		}

		t.Log("✓ AC4: Respects --target flag to check specific target only")
	})

	t.Run("AC5: Works with both base and profile sources", func(t *testing.T) {
		// Set up test environment
		profileName := "test-profile"
		baseDir := filepath.Join(tempDir, ".agent-smith", "profiles", profileName)
		skillName := "test-skill-ac5"
		skillsDir := filepath.Join(baseDir, "skills", skillName)

		// Create test skill in profile
		err := os.MkdirAll(skillsDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create skill directory")

		skillContent := `---
name: test-skill-ac5
version: 1.0.0
---
# Test Skill AC5
From profile.
`
		err = os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte(skillContent), 0644)
		testutil.AssertNoError(t, err, "Failed to write SKILL.md")

		// Create lock file in profile
		lockFilePath := filepath.Join(baseDir, ".skill-lock.json")
		lockData := map[string]interface{}{
			"version": 3,
			"skills": map[string]interface{}{
				skillName: map[string]interface{}{
					"source":       "test-repo",
					"sourceType":   "github",
					"sourceUrl":    "https://github.com/test/repo",
					"commitHash":   "abc123def456",
					"originalPath": fmt.Sprintf("skills/%s/SKILL.md", skillName),
					"installedAt":  "2024-01-15T10:30:00Z",
				},
			},
		}
		lockJSON, err := json.MarshalIndent(lockData, "", "  ")
		testutil.AssertNoError(t, err, "Failed to marshal lock data")
		err = os.WriteFile(lockFilePath, lockJSON, 0644)
		testutil.AssertNoError(t, err, "Failed to write lock file")

		// Create project directory
		projectDir := filepath.Join(tempDir, "test-project-ac5")
		opencodeDir := filepath.Join(projectDir, ".opencode")
		err = os.MkdirAll(opencodeDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create project directory")

		// Change to project directory
		err = os.Chdir(projectDir)
		testutil.AssertNoError(t, err, "Failed to change to project directory")

		// Materialize the skill from profile
		cmd := exec.Command(binaryPath, "materialize", "skill", skillName, "--target", "opencode", "--profile", profileName)
		output, err := cmd.CombinedOutput()
		testutil.AssertNoError(t, err, fmt.Sprintf("Failed to materialize skill: %s", string(output)))

		// Run status command
		cmd = exec.Command(binaryPath, "materialize", "status", "--target", "opencode")
		output, err = cmd.CombinedOutput()
		testutil.AssertNoError(t, err, fmt.Sprintf("Failed to run status command: %s", string(output)))

		outputStr := string(output)
		t.Logf("Status output:\n%s", outputStr)

		// Verify output shows in sync for profile-sourced component
		if !strings.Contains(outputStr, "in sync") {
			t.Errorf("Expected 'in sync' in output for profile-sourced component, got: %s", outputStr)
		}

		t.Log("✓ AC5: Works with both base and profile sources")
	})

	fmt.Println("✓ All materialize status acceptance criteria verified")
}
