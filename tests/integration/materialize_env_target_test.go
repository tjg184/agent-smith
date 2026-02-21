//go:build integration
// +build integration

package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/tjg184/agent-smith/internal/testutil"
	"github.com/tjg184/agent-smith/pkg/project"
)

// TestMaterializeWithEnvTarget verifies that the AGENT_SMITH_TARGET environment
// variable can be used to set a default target when --target flag is not provided.
// This test covers Story-005 acceptance criteria:
// - AGENT_SMITH_TARGET environment variable can set default target
func TestMaterializeWithEnvTarget(t *testing.T) {
	// Save current directory and restore it after test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(originalDir)
	})

	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-env-target-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	// Build agent-smith binary
	binaryPath := filepath.Join(tempDir, "agent-smith")
	repoRoot := filepath.Join("..", "..")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = repoRoot
	if output, buildErr := cmd.CombinedOutput(); buildErr != nil {
		t.Fatalf("Failed to build agent-smith: %v\nOutput: %s", buildErr, string(output))
	}

	// Create project directory
	projectDir := filepath.Join(tempDir, "test-project")
	err = os.MkdirAll(projectDir, 0755)
	testutil.AssertNoError(t, err, "Failed to create project directory")

	// Create .opencode directory to mark it as a project
	opencodeDir := filepath.Join(projectDir, ".opencode")
	err = os.MkdirAll(opencodeDir, 0755)
	testutil.AssertNoError(t, err, "Failed to create .opencode directory")

	// Create .claude directory
	claudeDir := filepath.Join(projectDir, ".claude")
	err = os.MkdirAll(claudeDir, 0755)
	testutil.AssertNoError(t, err, "Failed to create .claude directory")

	// Setup test skill in ~/.agent-smith/
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	skillsDir := filepath.Join(agentSmithDir, "skills")
	testSkillDir := filepath.Join(skillsDir, "env-test-skill")

	err = os.MkdirAll(testSkillDir, 0755)
	testutil.AssertNoError(t, err, "Failed to create test skill directory")

	skillContent := "# Env Test Skill\nThis is a test skill for environment variable testing."
	err = os.WriteFile(filepath.Join(testSkillDir, "SKILL.md"), []byte(skillContent), 0644)
	testutil.AssertNoError(t, err, "Failed to write SKILL.md")

	// Create lock file
	lockFilePath := filepath.Join(agentSmithDir, ".component-lock.json")
	lockData := map[string]interface{}{
		"version": 3,
		"skills": map[string]interface{}{
			"env-test-skill": map[string]interface{}{
				"source":       "test-repo",
				"sourceType":   "github",
				"sourceUrl":    "https://github.com/test/env-repo",
				"commitHash":   "env123",
				"originalPath": "SKILL.md",
				"installedAt":  "2024-01-01T00:00:00Z",
				"updatedAt":    "2024-01-01T00:00:00Z",
				"version":      3,
			},
		},
	}
	lockJSON, err := json.MarshalIndent(lockData, "", "  ")
	testutil.AssertNoError(t, err, "Failed to marshal lock data")
	err = os.WriteFile(lockFilePath, lockJSON, 0644)
	testutil.AssertNoError(t, err, "Failed to write lock file")

	t.Logf("Created test skill at: %s", testSkillDir)

	// Test 1: Materialize with AGENT_SMITH_TARGET=opencode (no --target flag)
	t.Run("EnvTargetOpencode", func(t *testing.T) {
		err := os.Chdir(projectDir)
		testutil.AssertNoError(t, err, "Failed to change to project directory")

		// Run materialize command with AGENT_SMITH_TARGET environment variable
		cmd := exec.Command(binaryPath, "materialize", "skill", "env-test-skill", "--verbose")
		cmd.Env = append(os.Environ(), "AGENT_SMITH_TARGET=opencode")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Materialize output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Materialize failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify component was copied to opencode target
		destPath := filepath.Join(opencodeDir, "skills", "env-test-skill", "SKILL.md")
		testutil.AssertFileExists(t, destPath)

		// Verify metadata exists in opencode target
		metadataPath := filepath.Join(opencodeDir, ".component-lock.json")
		testutil.AssertFileExists(t, metadataPath)

		// Load and verify metadata
		metadataBytes, err := os.ReadFile(metadataPath)
		testutil.AssertNoError(t, err, "Failed to read metadata")

		var metadata project.MaterializationMetadata
		err = json.Unmarshal(metadataBytes, &metadata)
		testutil.AssertNoError(t, err, "Failed to parse metadata")

		testutil.AssertEqual(t, 1, len(metadata.Skills), "Expected 1 skill in metadata")

		t.Logf("Successfully materialized with AGENT_SMITH_TARGET=opencode")
	})

	// Test 2: Materialize with AGENT_SMITH_TARGET=claudecode (no --target flag)
	t.Run("EnvTargetClaudecode", func(t *testing.T) {
		// Create another test skill
		testSkillDir2 := filepath.Join(skillsDir, "env-test-skill-2")
		err := os.MkdirAll(testSkillDir2, 0755)
		testutil.AssertNoError(t, err, "Failed to create second test skill directory")

		skillContent2 := "# Env Test Skill 2\nThis is a second test skill."
		err = os.WriteFile(filepath.Join(testSkillDir2, "SKILL.md"), []byte(skillContent2), 0644)
		testutil.AssertNoError(t, err, "Failed to write SKILL.md for skill 2")

		// Update lock file
		lockData["skills"].(map[string]interface{})["env-test-skill-2"] = map[string]interface{}{
			"source":       "test-repo-2",
			"sourceType":   "github",
			"sourceUrl":    "https://github.com/test/env-repo-2",
			"commitHash":   "env456",
			"originalPath": "SKILL.md",
			"installedAt":  "2024-01-01T00:00:00Z",
			"updatedAt":    "2024-01-01T00:00:00Z",
			"version":      3,
		}
		lockJSON, err := json.MarshalIndent(lockData, "", "  ")
		testutil.AssertNoError(t, err, "Failed to marshal updated lock data")
		err = os.WriteFile(lockFilePath, lockJSON, 0644)
		testutil.AssertNoError(t, err, "Failed to update lock file")

		err = os.Chdir(projectDir)
		testutil.AssertNoError(t, err, "Failed to change to project directory")

		// Run materialize command with AGENT_SMITH_TARGET=claudecode
		cmd := exec.Command(binaryPath, "materialize", "skill", "env-test-skill-2", "--verbose")
		cmd.Env = append(os.Environ(), "AGENT_SMITH_TARGET=claudecode")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Materialize output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Materialize failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify component was copied to claudecode target
		destPath := filepath.Join(claudeDir, "skills", "env-test-skill-2", "SKILL.md")
		testutil.AssertFileExists(t, destPath)

		// Verify metadata exists in claudecode target
		metadataPath := filepath.Join(claudeDir, ".component-lock.json")
		testutil.AssertFileExists(t, metadataPath)

		t.Logf("Successfully materialized with AGENT_SMITH_TARGET=claudecode")
	})

	// Test 3: Verify --target flag overrides AGENT_SMITH_TARGET
	t.Run("FlagOverridesEnvVar", func(t *testing.T) {
		// Create third test skill
		testSkillDir3 := filepath.Join(skillsDir, "env-test-skill-3")
		err := os.MkdirAll(testSkillDir3, 0755)
		testutil.AssertNoError(t, err, "Failed to create third test skill directory")

		skillContent3 := "# Env Test Skill 3\nThis is a third test skill."
		err = os.WriteFile(filepath.Join(testSkillDir3, "SKILL.md"), []byte(skillContent3), 0644)
		testutil.AssertNoError(t, err, "Failed to write SKILL.md for skill 3")

		// Update lock file
		lockData["skills"].(map[string]interface{})["env-test-skill-3"] = map[string]interface{}{
			"source":       "test-repo-3",
			"sourceType":   "github",
			"sourceUrl":    "https://github.com/test/env-repo-3",
			"commitHash":   "env789",
			"originalPath": "SKILL.md",
			"installedAt":  "2024-01-01T00:00:00Z",
			"updatedAt":    "2024-01-01T00:00:00Z",
			"version":      3,
		}
		lockJSON, err := json.MarshalIndent(lockData, "", "  ")
		testutil.AssertNoError(t, err, "Failed to marshal updated lock data")
		err = os.WriteFile(lockFilePath, lockJSON, 0644)
		testutil.AssertNoError(t, err, "Failed to update lock file")

		err = os.Chdir(projectDir)
		testutil.AssertNoError(t, err, "Failed to change to project directory")

		// Run materialize with AGENT_SMITH_TARGET=claudecode but --target opencode
		// Should materialize to opencode (flag overrides env var)
		cmd := exec.Command(binaryPath, "materialize", "skill", "env-test-skill-3", "--target", "opencode", "--verbose")
		cmd.Env = append(os.Environ(), "AGENT_SMITH_TARGET=claudecode")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Materialize output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Materialize failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify component was copied to opencode target (not claudecode)
		destPath := filepath.Join(opencodeDir, "skills", "env-test-skill-3", "SKILL.md")
		testutil.AssertFileExists(t, destPath)

		// Verify it was NOT copied to claudecode target
		claudeDestPath := filepath.Join(claudeDir, "skills", "env-test-skill-3", "SKILL.md")
		if _, err := os.Stat(claudeDestPath); err == nil {
			t.Errorf("Component should not exist in claudecode target, but it does")
		}

		t.Logf("Successfully verified --target flag overrides AGENT_SMITH_TARGET")
	})

	// Test 4: Verify AGENT_SMITH_TARGET=all works
	t.Run("EnvTargetAll", func(t *testing.T) {
		// Create fourth test skill
		testSkillDir4 := filepath.Join(skillsDir, "env-test-skill-4")
		err := os.MkdirAll(testSkillDir4, 0755)
		testutil.AssertNoError(t, err, "Failed to create fourth test skill directory")

		skillContent4 := "# Env Test Skill 4\nThis is a fourth test skill."
		err = os.WriteFile(filepath.Join(testSkillDir4, "SKILL.md"), []byte(skillContent4), 0644)
		testutil.AssertNoError(t, err, "Failed to write SKILL.md for skill 4")

		// Update lock file
		lockData["skills"].(map[string]interface{})["env-test-skill-4"] = map[string]interface{}{
			"source":       "test-repo-4",
			"sourceType":   "github",
			"sourceUrl":    "https://github.com/test/env-repo-4",
			"commitHash":   "env101",
			"originalPath": "SKILL.md",
			"installedAt":  "2024-01-01T00:00:00Z",
			"updatedAt":    "2024-01-01T00:00:00Z",
			"version":      3,
		}
		lockJSON, err := json.MarshalIndent(lockData, "", "  ")
		testutil.AssertNoError(t, err, "Failed to marshal updated lock data")
		err = os.WriteFile(lockFilePath, lockJSON, 0644)
		testutil.AssertNoError(t, err, "Failed to update lock file")

		err = os.Chdir(projectDir)
		testutil.AssertNoError(t, err, "Failed to change to project directory")

		// Run materialize with AGENT_SMITH_TARGET=all
		cmd := exec.Command(binaryPath, "materialize", "skill", "env-test-skill-4", "--verbose")
		cmd.Env = append(os.Environ(), "AGENT_SMITH_TARGET=all")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Materialize output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Materialize failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify component was copied to BOTH targets
		opencodeDestPath := filepath.Join(opencodeDir, "skills", "env-test-skill-4", "SKILL.md")
		testutil.AssertFileExists(t, opencodeDestPath)

		claudeDestPath := filepath.Join(claudeDir, "skills", "env-test-skill-4", "SKILL.md")
		testutil.AssertFileExists(t, claudeDestPath)

		t.Logf("Successfully materialized with AGENT_SMITH_TARGET=all to both targets")
	})
}
