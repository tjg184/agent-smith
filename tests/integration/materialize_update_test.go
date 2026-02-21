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
	"time"

	"github.com/tjg184/agent-smith/internal/testutil"
	"github.com/tjg184/agent-smith/pkg/project"
)

// TestMaterializeUpdate tests the materialize update command
func TestMaterializeUpdate(t *testing.T) {
	// Save current directory and restore it after test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(originalDir)
	})

	// Create temporary directories
	tempDir := testutil.CreateTempDir(t, "agent-smith-update-*")
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

	t.Run("AC1: Only updates out-of-sync components (smart mode)", func(t *testing.T) {
		// Set up test environment
		baseDir := filepath.Join(tempDir, ".agent-smith")

		// Create two skills - one will be updated, one stays the same
		skill1Name := "skill-updated"
		skill2Name := "skill-unchanged"

		// Create skill 1
		skill1Dir := filepath.Join(baseDir, "skills", skill1Name)
		err := os.MkdirAll(skill1Dir, 0755)
		testutil.AssertNoError(t, err, "Failed to create skill1 directory")

		skill1Content := `---
name: skill-updated
version: 1.0.0
---
# Skill Updated
Original content.
`
		err = os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte(skill1Content), 0644)
		testutil.AssertNoError(t, err, "Failed to write skill1 SKILL.md")

		// Create skill 2
		skill2Dir := filepath.Join(baseDir, "skills", skill2Name)
		err = os.MkdirAll(skill2Dir, 0755)
		testutil.AssertNoError(t, err, "Failed to create skill2 directory")

		skill2Content := `---
name: skill-unchanged
version: 1.0.0
---
# Skill Unchanged
Will not change.
`
		err = os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte(skill2Content), 0644)
		testutil.AssertNoError(t, err, "Failed to write skill2 SKILL.md")

		// Create lock file
		lockFilePath := filepath.Join(baseDir, ".component-lock.json")
		lockData := map[string]interface{}{
			"version": 3,
			"skills": map[string]interface{}{
				skill1Name: map[string]interface{}{
					"source":       "test-repo",
					"sourceType":   "github",
					"sourceUrl":    "https://github.com/test/repo",
					"commitHash":   "abc123def456",
					"originalPath": fmt.Sprintf("skills/%s/SKILL.md", skill1Name),
					"installedAt":  "2024-01-15T10:30:00Z",
				},
				skill2Name: map[string]interface{}{
					"source":       "test-repo",
					"sourceType":   "github",
					"sourceUrl":    "https://github.com/test/repo",
					"commitHash":   "xyz789abc012",
					"originalPath": fmt.Sprintf("skills/%s/SKILL.md", skill2Name),
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

		// Materialize both skills
		cmd := exec.Command(binaryPath, "materialize", "all", "--target", "opencode")
		output, err := cmd.CombinedOutput()
		testutil.AssertNoError(t, err, fmt.Sprintf("Failed to materialize skills: %s", string(output)))

		// Update skill1 in source
		updatedSkill1Content := `---
name: skill-updated
version: 2.0.0
---
# Skill Updated
Updated content with new features.
`
		err = os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte(updatedSkill1Content), 0644)
		testutil.AssertNoError(t, err, "Failed to update skill1")

		// Run update command (smart mode - should only update skill1)
		cmd = exec.Command(binaryPath, "materialize", "update", "--target", "opencode")
		output, err = cmd.CombinedOutput()
		testutil.AssertNoError(t, err, fmt.Sprintf("Failed to run update command: %s", string(output)))

		outputStr := string(output)
		t.Logf("Update output:\n%s", outputStr)

		// Verify skill1 was updated
		if !strings.Contains(outputStr, fmt.Sprintf("Updated %s", skill1Name)) {
			t.Errorf("Expected '%s' to be updated, got: %s", skill1Name, outputStr)
		}

		// Verify skill2 was skipped
		if !strings.Contains(outputStr, fmt.Sprintf("Skipped %s", skill2Name)) {
			t.Errorf("Expected '%s' to be skipped, got: %s", skill2Name, outputStr)
		}

		// Verify summary shows 1 updated, 1 skipped
		if !strings.Contains(outputStr, "1 updated") {
			t.Errorf("Expected summary to show '1 updated', got: %s", outputStr)
		}
		if !strings.Contains(outputStr, "1 already in sync") {
			t.Errorf("Expected summary to show '1 already in sync', got: %s", outputStr)
		}

		// Verify materialized skill1 has updated content
		materializedSkill1 := filepath.Join(opencodeDir, "skills", skill1Name, "SKILL.md")
		content, err := os.ReadFile(materializedSkill1)
		testutil.AssertNoError(t, err, "Failed to read materialized skill1")
		if !strings.Contains(string(content), "Updated content with new features") {
			t.Errorf("Expected materialized skill1 to have updated content, got: %s", string(content))
		}

		t.Log("✓ AC1: Only updates out-of-sync components (smart mode)")
	})

	t.Run("AC2: Updates all components with --force flag", func(t *testing.T) {
		// Set up test environment
		baseDir := filepath.Join(tempDir, ".agent-smith")
		skillName := "skill-force-ac2"
		skillsDir := filepath.Join(baseDir, "skills", skillName)

		// Create test skill
		err := os.MkdirAll(skillsDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create skill directory")

		skillContent := `---
name: skill-force-ac2
version: 1.0.0
---
# Skill Force AC2
Original content.
`
		err = os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte(skillContent), 0644)
		testutil.AssertNoError(t, err, "Failed to write SKILL.md")

		// Create lock file
		lockFilePath := filepath.Join(baseDir, ".component-lock.json")
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

		// Run update command with --force (should update even though in sync)
		cmd = exec.Command(binaryPath, "materialize", "update", "--target", "opencode", "--force")
		output, err = cmd.CombinedOutput()
		testutil.AssertNoError(t, err, fmt.Sprintf("Failed to run update command: %s", string(output)))

		outputStr := string(output)
		t.Logf("Update output:\n%s", outputStr)

		// Verify skill was updated (not skipped)
		if !strings.Contains(outputStr, fmt.Sprintf("Updated %s", skillName)) {
			t.Errorf("Expected '%s' to be updated with --force, got: %s", skillName, outputStr)
		}

		// Verify NO "already in sync" skips with --force
		if strings.Contains(outputStr, "already in sync") {
			t.Errorf("Expected no 'already in sync' skips with --force, got: %s", outputStr)
		}

		t.Log("✓ AC2: Updates all components with --force flag")
	})

	t.Run("AC3: Skips and warns about missing sources", func(t *testing.T) {
		// Set up test environment
		baseDir := filepath.Join(tempDir, ".agent-smith")
		skillName := "skill-missing-ac3"
		skillsDir := filepath.Join(baseDir, "skills", skillName)

		// Create test skill
		err := os.MkdirAll(skillsDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create skill directory")

		skillContent := `---
name: skill-missing-ac3
version: 1.0.0
---
# Skill Missing AC3
Will be removed.
`
		err = os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte(skillContent), 0644)
		testutil.AssertNoError(t, err, "Failed to write SKILL.md")

		// Create lock file
		lockFilePath := filepath.Join(baseDir, ".component-lock.json")
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

		// Remove source skill to simulate uninstallation
		err = os.RemoveAll(skillsDir)
		testutil.AssertNoError(t, err, "Failed to remove source skill")

		// Run update command
		cmd = exec.Command(binaryPath, "materialize", "update", "--target", "opencode")
		output, err = cmd.CombinedOutput()
		testutil.AssertNoError(t, err, fmt.Sprintf("Failed to run update command: %s", string(output)))

		outputStr := string(output)
		t.Logf("Update output:\n%s", outputStr)

		// Verify warning about missing source
		if !strings.Contains(outputStr, "Skipped") && !strings.Contains(outputStr, "source no longer installed") {
			t.Errorf("Expected warning about missing source, got: %s", outputStr)
		}

		// Verify summary shows skipped (source missing)
		if !strings.Contains(outputStr, "source missing") {
			t.Errorf("Expected summary to show 'source missing', got: %s", outputStr)
		}

		t.Log("✓ AC3: Skips and warns about missing sources")
	})

	t.Run("AC4: Dry-run shows preview without making changes", func(t *testing.T) {
		// Set up test environment
		baseDir := filepath.Join(tempDir, ".agent-smith")
		skillName := "skill-dryrun-ac4"
		skillsDir := filepath.Join(baseDir, "skills", skillName)

		// Create test skill
		err := os.MkdirAll(skillsDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create skill directory")

		originalContent := `---
name: skill-dryrun-ac4
version: 1.0.0
---
# Skill Dry Run AC4
Original content.
`
		err = os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte(originalContent), 0644)
		testutil.AssertNoError(t, err, "Failed to write SKILL.md")

		// Create lock file
		lockFilePath := filepath.Join(baseDir, ".component-lock.json")
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
		err = os.MkdirAll(opencodeDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create project directory")

		// Change to project directory
		err = os.Chdir(projectDir)
		testutil.AssertNoError(t, err, "Failed to change to project directory")

		// Materialize the skill
		cmd := exec.Command(binaryPath, "materialize", "skill", skillName, "--target", "opencode")
		output, err := cmd.CombinedOutput()
		testutil.AssertNoError(t, err, fmt.Sprintf("Failed to materialize skill: %s", string(output)))

		// Update source skill
		updatedContent := `---
name: skill-dryrun-ac4
version: 2.0.0
---
# Skill Dry Run AC4
Updated content.
`
		err = os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte(updatedContent), 0644)
		testutil.AssertNoError(t, err, "Failed to update skill")

		// Run update command with --dry-run
		cmd = exec.Command(binaryPath, "materialize", "update", "--target", "opencode", "--dry-run")
		output, err = cmd.CombinedOutput()
		testutil.AssertNoError(t, err, fmt.Sprintf("Failed to run dry-run update: %s", string(output)))

		outputStr := string(output)
		t.Logf("Dry-run output:\n%s", outputStr)

		// Verify dry-run indicator in output
		if !strings.Contains(outputStr, "DRY RUN") && !strings.Contains(outputStr, "Would update") {
			t.Errorf("Expected dry-run indicator in output, got: %s", outputStr)
		}

		// Verify materialized file was NOT actually updated
		materializedSkill := filepath.Join(opencodeDir, "skills", skillName, "SKILL.md")
		content, err := os.ReadFile(materializedSkill)
		testutil.AssertNoError(t, err, "Failed to read materialized skill")
		if strings.Contains(string(content), "Updated content") {
			t.Errorf("Expected materialized skill to remain unchanged in dry-run, but it was updated: %s", string(content))
		}
		if !strings.Contains(string(content), "Original content") {
			t.Errorf("Expected materialized skill to have original content, got: %s", string(content))
		}

		t.Log("✓ AC4: Dry-run shows preview without making changes")
	})

	t.Run("AC5: Updates metadata with new hashes and timestamp", func(t *testing.T) {
		// Set up test environment
		baseDir := filepath.Join(tempDir, ".agent-smith")
		skillName := "skill-metadata-ac5"
		skillsDir := filepath.Join(baseDir, "skills", skillName)

		// Create test skill
		err := os.MkdirAll(skillsDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create skill directory")

		skillContent := `---
name: skill-metadata-ac5
version: 1.0.0
---
# Skill Metadata AC5
Original content.
`
		err = os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte(skillContent), 0644)
		testutil.AssertNoError(t, err, "Failed to write SKILL.md")

		// Create lock file
		lockFilePath := filepath.Join(baseDir, ".component-lock.json")
		lockData := map[string]interface{}{
			"version": 3,
			"skills": map[string]interface{}{
				skillName: map[string]interface{}{
					"source":       "test-repo",
					"sourceType":   "github",
					"sourceUrl":    "https://github.com/test/repo",
					"commitHash":   "oldcommit123",
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

		// Materialize the skill
		cmd := exec.Command(binaryPath, "materialize", "skill", skillName, "--target", "opencode")
		output, err := cmd.CombinedOutput()
		testutil.AssertNoError(t, err, fmt.Sprintf("Failed to materialize skill: %s", string(output)))

		// Load initial metadata
		initialMetadata, err := project.LoadMaterializationMetadata(opencodeDir)
		testutil.AssertNoError(t, err, "Failed to load initial metadata")

		// Access nested structure: Skills[sourceURL][componentName]
		sourceURL := "https://github.com/test/repo"
		skillsFromSource, exists := initialMetadata.Skills[sourceURL]
		testutil.AssertTrue(t, exists, "Source URL not found in initial metadata")
		initialEntry, exists := skillsFromSource[skillName]
		testutil.AssertTrue(t, exists, "Skill not found in initial metadata")

		initialSourceHash := initialEntry.SourceHash
		initialTimestamp := initialEntry.MaterializedAt

		// Update source skill
		updatedContent := `---
name: skill-metadata-ac5
version: 2.0.0
---
# Skill Metadata AC5
Updated content for metadata testing.
`
		err = os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte(updatedContent), 0644)
		testutil.AssertNoError(t, err, "Failed to update skill")

		// Update lock file with new commit hash
		lockData["skills"].(map[string]interface{})[skillName].(map[string]interface{})["commitHash"] = "newcommit456"
		lockJSON, err = json.MarshalIndent(lockData, "", "  ")
		testutil.AssertNoError(t, err, "Failed to marshal updated lock data")
		err = os.WriteFile(lockFilePath, lockJSON, 0644)
		testutil.AssertNoError(t, err, "Failed to write updated lock file")

		// Add a small sleep to ensure timestamp will be different
		time.Sleep(1 * time.Second)

		// Run update command
		cmd = exec.Command(binaryPath, "materialize", "update", "--target", "opencode")
		output, err = cmd.CombinedOutput()
		testutil.AssertNoError(t, err, fmt.Sprintf("Failed to run update command: %s", string(output)))

		// Load updated metadata
		updatedMetadata, err := project.LoadMaterializationMetadata(opencodeDir)
		testutil.AssertNoError(t, err, "Failed to load updated metadata")

		// Access nested structure
		updatedSkillsFromSource, exists := updatedMetadata.Skills[sourceURL]
		testutil.AssertTrue(t, exists, "Source URL not found in updated metadata")
		updatedEntry, exists := updatedSkillsFromSource[skillName]
		testutil.AssertTrue(t, exists, "Skill not found in updated metadata")

		// Verify sourceHash was updated
		if updatedEntry.SourceHash == initialSourceHash {
			t.Errorf("Expected sourceHash to change after update, but it remained: %s", initialSourceHash)
		}

		// Verify currentHash was updated
		if updatedEntry.CurrentHash == initialEntry.CurrentHash {
			t.Errorf("Expected currentHash to change after update")
		}

		// Verify timestamp was updated
		if updatedEntry.MaterializedAt == initialTimestamp {
			t.Errorf("Expected timestamp to change after update, but it remained: %s", initialTimestamp)
		}

		// Verify commit hash was updated from lock file
		if updatedEntry.CommitHash != "newcommit456" {
			t.Errorf("Expected commit hash to be updated to 'newcommit456', got: %s", updatedEntry.CommitHash)
		}

		t.Log("✓ AC5: Updates metadata with new hashes and timestamp")
	})

	t.Run("AC6: Respects --target flag to update specific target only", func(t *testing.T) {
		// Set up test environment
		baseDir := filepath.Join(tempDir, ".agent-smith")
		skillName := "skill-target-ac6"
		skillsDir := filepath.Join(baseDir, "skills", skillName)

		// Create test skill
		err := os.MkdirAll(skillsDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create skill directory")

		skillContent := `---
name: skill-target-ac6
version: 1.0.0
---
# Skill Target AC6
For target testing.
`
		err = os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte(skillContent), 0644)
		testutil.AssertNoError(t, err, "Failed to write SKILL.md")

		// Create lock file
		lockFilePath := filepath.Join(baseDir, ".component-lock.json")
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

		// Create project directory with both targets
		projectDir := filepath.Join(tempDir, "test-project-ac6")
		opencodeDir := filepath.Join(projectDir, ".opencode")
		claudeDir := filepath.Join(projectDir, ".claude")
		err = os.MkdirAll(opencodeDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create .opencode directory")
		err = os.MkdirAll(claudeDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create .claude directory")

		// Change to project directory
		err = os.Chdir(projectDir)
		testutil.AssertNoError(t, err, "Failed to change to project directory")

		// Materialize to both targets
		cmd := exec.Command(binaryPath, "materialize", "skill", skillName, "--target", "opencode")
		output, err := cmd.CombinedOutput()
		testutil.AssertNoError(t, err, fmt.Sprintf("Failed to materialize to opencode: %s", string(output)))

		cmd = exec.Command(binaryPath, "materialize", "skill", skillName, "--target", "claudecode")
		output, err = cmd.CombinedOutput()
		testutil.AssertNoError(t, err, fmt.Sprintf("Failed to materialize to claude: %s", string(output)))

		// Update source skill
		updatedContent := `---
name: skill-target-ac6
version: 2.0.0
---
# Skill Target AC6
Updated content.
`
		err = os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte(updatedContent), 0644)
		testutil.AssertNoError(t, err, "Failed to update skill")

		// Run update command for opencode target only
		cmd = exec.Command(binaryPath, "materialize", "update", "--target", "opencode")
		output, err = cmd.CombinedOutput()
		testutil.AssertNoError(t, err, fmt.Sprintf("Failed to run update command: %s", string(output)))

		outputStr := string(output)
		t.Logf("Update output:\n%s", outputStr)

		// Verify output mentions opencode
		if !strings.Contains(outputStr, "OpenCode") && !strings.Contains(outputStr, ".opencode") {
			t.Errorf("Expected OpenCode reference in output, got: %s", outputStr)
		}

		// Verify opencode was updated
		opencodeSkill := filepath.Join(opencodeDir, "skills", skillName, "SKILL.md")
		opencodeContent, err := os.ReadFile(opencodeSkill)
		testutil.AssertNoError(t, err, "Failed to read opencode skill")
		if !strings.Contains(string(opencodeContent), "Updated content") {
			t.Errorf("Expected opencode skill to be updated, got: %s", string(opencodeContent))
		}

		// Verify claude was NOT updated
		claudeSkill := filepath.Join(claudeDir, "skills", skillName, "SKILL.md")
		claudeContent, err := os.ReadFile(claudeSkill)
		testutil.AssertNoError(t, err, "Failed to read claude skill")
		if strings.Contains(string(claudeContent), "Updated content") {
			t.Errorf("Expected claude skill to remain unchanged when using --target opencode, but it was updated")
		}

		t.Log("✓ AC6: Respects --target flag to update specific target only")
	})

	fmt.Println("✓ All materialize update acceptance criteria verified")
}
