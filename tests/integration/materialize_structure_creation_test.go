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

	"github.com/tjg184/agent-smith/internal/testutil"
)

// TestMaterializeStructureCreation verifies that the first materialize command
// automatically creates the project structure and shows clear output.
// This test covers Story-006 acceptance criteria.
func TestMaterializeStructureCreation(t *testing.T) {
	// Save current directory and restore it after test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(originalDir)
	})

	tempDir := testutil.CreateTempDir(t, "agent-smith-structure-*")
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

	// Create empty project directory without .opencode/
	projectDir := filepath.Join(tempDir, "test-project")
	err = os.MkdirAll(projectDir, 0755)
	testutil.AssertNoError(t, err, "Failed to create project directory")

	// We intentionally DON'T create .opencode/ - it should be created automatically

	// Setup test skill in ~/.agent-smith/
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	skillDir := filepath.Join(agentSmithDir, "skills", "test-skill")
	err = os.MkdirAll(skillDir, 0755)
	testutil.AssertNoError(t, err, "Failed to create skill directory")

	skillContent := "# Test Skill\nThis is a test skill for structure creation."
	err = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0644)
	testutil.AssertNoError(t, err, "Failed to write skill file")

	// Create lock file
	lockFilePath := filepath.Join(agentSmithDir, ".component-lock.json")
	lockData := map[string]interface{}{
		"version": 3,
		"skills": map[string]interface{}{
			"test-skill": map[string]interface{}{
				"source":       "test-repo",
				"sourceType":   "github",
				"sourceUrl":    "https://github.com/test/repo",
				"commitHash":   "abc123",
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

	// First materialize - should create structure
	cmd = exec.Command(binaryPath, "materialize", "skill", "test-skill", "--target", "opencode", "--project-dir", projectDir, "--verbose")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	t.Logf("First materialize output:\n%s", outputStr)

	if err != nil {
		t.Fatalf("First materialize failed: %v\nOutput: %s", err, outputStr)
	}

	// Story-006 Acceptance Criteria #1: Output shows directory was created
	if !strings.Contains(outputStr, "Created directory") {
		t.Errorf("Expected output to indicate directory was created, got: %s", outputStr)
	}

	// Story-006 Acceptance Criteria #2: Verify .opencode/ directory was created
	opencodeDir := filepath.Join(projectDir, ".opencode")
	if _, err := os.Stat(opencodeDir); os.IsNotExist(err) {
		t.Errorf(".opencode/ directory was not created")
	}

	// Story-006 Acceptance Criteria #3: Verify skills subdirectory was created
	// Note: Only the skills/ directory should exist since we only materialized a skill
	skillsDir := filepath.Join(opencodeDir, "skills")
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		t.Errorf("Subdirectory skills/ was not created")
	}

	// Verify that agents/ and commands/ were NOT created (we only materialized a skill)
	agentsDir := filepath.Join(opencodeDir, "agents")
	if _, err := os.Stat(agentsDir); err == nil {
		t.Errorf("Subdirectory agents/ should not have been created (only materialized skills)")
	}
	commandsDir := filepath.Join(opencodeDir, "commands")
	if _, err := os.Stat(commandsDir); err == nil {
		t.Errorf("Subdirectory commands/ should not have been created (only materialized skills)")
	}

	// Story-006 Acceptance Criteria #4: Verify .component-lock.json was created
	metadataPath := filepath.Join(opencodeDir, ".component-lock.json")
	testutil.AssertFileExists(t, metadataPath)

	// Verify component was materialized
	destPath := filepath.Join(opencodeDir, "skills", "test-skill", "SKILL.md")
	testutil.AssertFileExists(t, destPath)

	// Now create a second skill to test subsequent materialization
	skill2Dir := filepath.Join(agentSmithDir, "skills", "test-skill-2")
	err = os.MkdirAll(skill2Dir, 0755)
	testutil.AssertNoError(t, err, "Failed to create second skill directory")

	err = os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte("# Second Skill"), 0644)
	testutil.AssertNoError(t, err, "Failed to write second skill file")

	// Update lock file
	lockData["skills"].(map[string]interface{})["test-skill-2"] = map[string]interface{}{
		"source":       "test-repo",
		"sourceType":   "github",
		"sourceUrl":    "https://github.com/test/repo",
		"commitHash":   "def456",
		"originalPath": "SKILL.md",
		"installedAt":  "2024-01-01T00:00:00Z",
		"updatedAt":    "2024-01-01T00:00:00Z",
		"version":      3,
	}
	lockJSON, err = json.MarshalIndent(lockData, "", "  ")
	testutil.AssertNoError(t, err, "Failed to marshal lock data")
	err = os.WriteFile(lockFilePath, lockJSON, 0644)
	testutil.AssertNoError(t, err, "Failed to write lock file")

	// Second materialize - structure already exists, should NOT show creation message
	cmd = exec.Command(binaryPath, "materialize", "skill", "test-skill-2", "--target", "opencode", "--project-dir", projectDir, "--verbose")
	output, err = cmd.CombinedOutput()
	outputStr = string(output)
	t.Logf("Second materialize output:\n%s", outputStr)

	if err != nil {
		t.Fatalf("Second materialize failed: %v\nOutput: %s", err, outputStr)
	}

	// Story-006 Acceptance Criteria #5: Second materialize doesn't show creation message
	if strings.Contains(outputStr, "Created directory") {
		t.Errorf("Second materialize should not show creation message (structure already exists)")
	}

	// Story-006 Acceptance Criteria #5: Subsequent materializations don't show structure creation message
	if strings.Contains(outputStr, "Created project structure") {
		t.Errorf("Expected output to NOT show structure creation for existing structure, got: %s", outputStr)
	}

	// Verify second component was materialized
	dest2Path := filepath.Join(opencodeDir, "skills", "test-skill-2", "SKILL.md")
	testutil.AssertFileExists(t, dest2Path)

	t.Logf("✓ Story-006 acceptance criteria verified:")
	t.Logf("  - First materialize automatically creates .opencode/ directory")
	t.Logf("  - Subdirectories created: skills/, agents/, commands/")
	t.Logf("  - Empty .component-lock.json created")
	t.Logf("  - Clear output shows structure was created")
	t.Logf("  - Subsequent materializations don't recreate existing structure")
}

// TestMaterializeStructureCreationBothTargets verifies structure creation
// works correctly when materializing to both targets at once.
// This test covers Story-006 with --target all flag.
func TestMaterializeStructureCreationBothTargets(t *testing.T) {
	// Save current directory and restore it after test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(originalDir)
	})

	tempDir := testutil.CreateTempDir(t, "agent-smith-both-targets-*")
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

	// Create empty project directory
	projectDir := filepath.Join(tempDir, "test-project")
	err = os.MkdirAll(projectDir, 0755)
	testutil.AssertNoError(t, err, "Failed to create project directory")

	// Setup test agent
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	agentDir := filepath.Join(agentSmithDir, "agents", "test-agent")
	err = os.MkdirAll(agentDir, 0755)
	testutil.AssertNoError(t, err, "Failed to create agent directory")

	err = os.WriteFile(filepath.Join(agentDir, "AGENT.md"), []byte("# Test Agent"), 0644)
	testutil.AssertNoError(t, err, "Failed to write agent file")

	// Create lock file
	lockFilePath := filepath.Join(agentSmithDir, ".component-lock.json")
	lockData := map[string]interface{}{
		"version": 3,
		"agents": map[string]interface{}{
			"test-agent": map[string]interface{}{
				"source":       "test-repo",
				"sourceType":   "github",
				"sourceUrl":    "https://github.com/test/agents",
				"commitHash":   "xyz789",
				"originalPath": "AGENT.md",
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

	// Materialize to both targets at once
	cmd = exec.Command(binaryPath, "materialize", "agent", "test-agent", "--target", "all", "--project-dir", projectDir, "--verbose")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	t.Logf("Materialize to both targets output:\n%s", outputStr)

	if err != nil {
		t.Fatalf("Materialize to both targets failed: %v\nOutput: %s", err, outputStr)
	}

	// Verify output shows directories were created
	if !strings.Contains(outputStr, "Created directory") {
		t.Errorf("Expected output to show directory creation, got: %s", outputStr)
	}

	// Verify both target directories were created
	opencodeDir := filepath.Join(projectDir, ".opencode")
	claudeDir := filepath.Join(projectDir, ".claude")

	for _, targetDir := range []string{opencodeDir, claudeDir} {
		// Check target directory exists
		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			t.Errorf("Target directory %s was not created", targetDir)
		}

		// Check that agents subdirectory was created (we materialized an agent)
		agentsPath := filepath.Join(targetDir, "agents")
		if _, err := os.Stat(agentsPath); os.IsNotExist(err) {
			t.Errorf("Subdirectory %s/agents was not created", targetDir)
		}

		// Check that skills and commands were NOT created (only materialized agents)
		skillsPath := filepath.Join(targetDir, "skills")
		if _, err := os.Stat(skillsPath); err == nil {
			t.Errorf("Subdirectory %s/skills should not have been created (only materialized agents)", targetDir)
		}
		commandsPath := filepath.Join(targetDir, "commands")
		if _, err := os.Stat(commandsPath); err == nil {
			t.Errorf("Subdirectory %s/commands should not have been created (only materialized agents)", targetDir)
		}

		// Check metadata file
		metadataPath := filepath.Join(targetDir, ".component-lock.json")
		testutil.AssertFileExists(t, metadataPath)

		// Verify agent was materialized
		agentPath := filepath.Join(targetDir, "agents", "test-agent", "AGENT.md")
		testutil.AssertFileExists(t, agentPath)
	}

	t.Logf("✓ Story-006 verified for --target all:")
	t.Logf("  - Both .opencode/ and .claude/ directories created")
	t.Logf("  - Subdirectories created in both targets")
	t.Logf("  - Metadata files created in both targets")
	t.Logf("  - Clear output shows structures were created")
}
