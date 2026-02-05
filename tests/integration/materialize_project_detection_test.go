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

// TestMaterializeFromNestedSubdirectory verifies that project auto-detection
// works from nested subdirectories within a project.
// This test covers Story-003 acceptance criteria for auto-detection.
func TestMaterializeFromNestedSubdirectory(t *testing.T) {
	// Save current directory and restore it after test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(originalDir)
	})

	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-materialize-nested-*")
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

	// Create project with nested directory structure
	projectDir := filepath.Join(tempDir, "test-project")
	opencodeDir := filepath.Join(projectDir, ".opencode")
	nestedDir := filepath.Join(projectDir, "src", "components", "deep")
	err = os.MkdirAll(nestedDir, 0755)
	testutil.AssertNoError(t, err, "Failed to create nested directory")
	err = os.MkdirAll(opencodeDir, 0755)
	testutil.AssertNoError(t, err, "Failed to create .opencode directory")

	// Setup test skill in ~/.agent-smith/
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	skillDir := filepath.Join(agentSmithDir, "skills", "test-skill")
	err = os.MkdirAll(skillDir, 0755)
	testutil.AssertNoError(t, err, "Failed to create skill directory")

	skillContent := "# Test Skill\nThis is a test skill."
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

	// Change to deeply nested directory
	err = os.Chdir(nestedDir)
	testutil.AssertNoError(t, err, "Failed to change to nested directory")

	t.Logf("Working directory: %s", nestedDir)
	t.Logf("Project root should be: %s", projectDir)

	// Run materialize from nested directory (should auto-detect project root)
	cmd = exec.Command(binaryPath, "materialize", "skill", "test-skill", "--target", "opencode", "--verbose")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	t.Logf("Materialize output:\n%s", outputStr)

	if err != nil {
		t.Fatalf("Materialize failed from nested directory: %v\nOutput: %s", err, outputStr)
	}

	// Verify component was materialized to project root
	destPath := filepath.Join(opencodeDir, "skills", "test-skill", "SKILL.md")
	testutil.AssertFileExists(t, destPath)

	content, err := os.ReadFile(destPath)
	testutil.AssertNoError(t, err, "Failed to read materialized file")
	testutil.AssertEqual(t, skillContent, string(content), "File content mismatch")

	t.Logf("Successfully materialized from nested subdirectory")
}

// TestMaterializeWithProjectDirOverride verifies that the --project-dir flag
// allows overriding auto-detection.
// This test covers Story-003 acceptance criteria for --project-dir flag.
func TestMaterializeWithProjectDirOverride(t *testing.T) {
	// Save current directory and restore it after test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(originalDir)
	})

	tempDir := testutil.CreateTempDir(t, "agent-smith-materialize-override-*")
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

	// Create two separate project directories
	project1Dir := filepath.Join(tempDir, "project1")
	project1Opencode := filepath.Join(project1Dir, ".opencode")
	err = os.MkdirAll(project1Opencode, 0755)
	testutil.AssertNoError(t, err, "Failed to create project1 .opencode")

	project2Dir := filepath.Join(tempDir, "project2")
	project2Opencode := filepath.Join(project2Dir, ".opencode")
	err = os.MkdirAll(project2Opencode, 0755)
	testutil.AssertNoError(t, err, "Failed to create project2 .opencode")

	// Setup test agent in ~/.agent-smith/
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	agentDir := filepath.Join(agentSmithDir, "agents", "test-agent")
	err = os.MkdirAll(agentDir, 0755)
	testutil.AssertNoError(t, err, "Failed to create agent directory")

	agentContent := "# Test Agent\nThis is a test agent."
	err = os.WriteFile(filepath.Join(agentDir, "AGENT.md"), []byte(agentContent), 0644)
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
				"commitHash":   "def456",
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

	// Change to project1 directory
	err = os.Chdir(project1Dir)
	testutil.AssertNoError(t, err, "Failed to change to project1 directory")

	t.Logf("Current directory: %s", project1Dir)
	t.Logf("Overriding to materialize to: %s", project2Dir)

	// Run materialize with --project-dir pointing to project2
	// Should ignore auto-detection (project1) and use project2 instead
	cmd = exec.Command(binaryPath, "materialize", "agent", "test-agent", "--target", "opencode", "--project-dir", project2Dir, "--verbose")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	t.Logf("Materialize output:\n%s", outputStr)

	if err != nil {
		t.Fatalf("Materialize with --project-dir failed: %v\nOutput: %s", err, outputStr)
	}

	// Verify component was materialized to project2, NOT project1
	dest2Path := filepath.Join(project2Opencode, "agents", "test-agent", "AGENT.md")
	testutil.AssertFileExists(t, dest2Path)

	dest1Path := filepath.Join(project1Opencode, "agents", "test-agent", "AGENT.md")
	if _, err := os.Stat(dest1Path); !os.IsNotExist(err) {
		t.Errorf("Component should not exist in project1 (auto-detected location), but it does")
	}

	content, err := os.ReadFile(dest2Path)
	testutil.AssertNoError(t, err, "Failed to read materialized file")
	testutil.AssertEqual(t, agentContent, string(content), "File content mismatch")

	t.Logf("Successfully overrode project directory with --project-dir flag")
}

// TestMaterializeNoProjectFound verifies error handling when no project
// is found in the directory tree.
// This test covers Story-003 acceptance criteria for error handling.
func TestMaterializeNoProjectFound(t *testing.T) {
	// Save current directory and restore it after test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(originalDir)
	})

	tempDir := testutil.CreateTempDir(t, "agent-smith-materialize-noproject-*")
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

	// Create a directory WITHOUT .opencode or .claude
	nonProjectDir := filepath.Join(tempDir, "not-a-project", "src")
	err = os.MkdirAll(nonProjectDir, 0755)
	testutil.AssertNoError(t, err, "Failed to create non-project directory")

	// Setup test command in ~/.agent-smith/
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	commandDir := filepath.Join(agentSmithDir, "commands", "test-command")
	err = os.MkdirAll(commandDir, 0755)
	testutil.AssertNoError(t, err, "Failed to create command directory")

	err = os.WriteFile(filepath.Join(commandDir, "COMMAND.md"), []byte("# Test Command"), 0644)
	testutil.AssertNoError(t, err, "Failed to write command file")

	// Create lock file
	lockFilePath := filepath.Join(agentSmithDir, ".component-lock.json")
	lockData := map[string]interface{}{
		"version": 3,
		"commands": map[string]interface{}{
			"test-command": map[string]interface{}{
				"source":       "test-repo",
				"sourceType":   "github",
				"sourceUrl":    "https://github.com/test/commands",
				"commitHash":   "ghi789",
				"originalPath": "COMMAND.md",
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

	// Change to non-project directory
	err = os.Chdir(nonProjectDir)
	testutil.AssertNoError(t, err, "Failed to change to non-project directory")

	t.Logf("Working directory: %s (no .opencode or .claude in tree)", nonProjectDir)

	// Try to materialize without a project
	cmd = exec.Command(binaryPath, "materialize", "command", "test-command", "--target", "opencode")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	t.Logf("Error output:\n%s", outputStr)

	// Should fail
	if err == nil {
		t.Fatalf("Expected error when no project found, but command succeeded")
	}

	// Verify error message contains helpful information
	if !strings.Contains(outputStr, "no project boundary detected") {
		t.Errorf("Expected error message about project boundary not detected, got: %s", outputStr)
	}

	// Should list supported project markers
	if !strings.Contains(outputStr, "Supported project markers:") {
		t.Errorf("Expected error message to list supported project markers, got: %s", outputStr)
	}

	// Should mention preferred markers
	if !strings.Contains(outputStr, ".opencode/") || !strings.Contains(outputStr, ".claude/") {
		t.Errorf("Expected error message to mention .opencode/ and .claude/, got: %s", outputStr)
	}

	// Should suggest fixes
	if !strings.Contains(outputStr, "To fix this:") {
		t.Errorf("Expected error message to suggest fixes, got: %s", outputStr)
	}

	// Should suggest --project-dir flag
	if !strings.Contains(outputStr, "--project-dir") {
		t.Errorf("Expected error message to mention --project-dir flag, got: %s", outputStr)
	}

	// Should suggest git init
	if !strings.Contains(outputStr, "git init") {
		t.Errorf("Expected error message to mention git init, got: %s", outputStr)
	}

	t.Logf("Error handling verified: clear message when no project found")
}

// TestMaterializeStopsAtHomeDirectory verifies that project detection
// stops at the home directory and doesn't continue to filesystem root.
// This test covers Story-003 acceptance criteria for boundary detection.
func TestMaterializeStopsAtHomeDirectory(t *testing.T) {
	// Save current directory and restore it after test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(originalDir)
	})

	tempDir := testutil.CreateTempDir(t, "agent-smith-materialize-home-*")
	oldHome := os.Getenv("HOME")

	// Set HOME to tempDir (this becomes our "home directory" boundary)
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

	// Create a subdirectory in "home" without a project
	subDir := filepath.Join(tempDir, "documents", "work")
	err = os.MkdirAll(subDir, 0755)
	testutil.AssertNoError(t, err, "Failed to create subdirectory")

	// Setup test skill
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	skillDir := filepath.Join(agentSmithDir, "skills", "boundary-test")
	err = os.MkdirAll(skillDir, 0755)
	testutil.AssertNoError(t, err, "Failed to create skill directory")

	err = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Boundary Test"), 0644)
	testutil.AssertNoError(t, err, "Failed to write skill file")

	// Create lock file
	lockFilePath := filepath.Join(agentSmithDir, ".component-lock.json")
	lockData := map[string]interface{}{
		"version": 3,
		"skills": map[string]interface{}{
			"boundary-test": map[string]interface{}{
				"source":       "test-repo",
				"sourceType":   "github",
				"sourceUrl":    "https://github.com/test/boundary",
				"commitHash":   "jkl012",
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

	// Change to subdirectory (below "home" without a project)
	err = os.Chdir(subDir)
	testutil.AssertNoError(t, err, "Failed to change to subdirectory")

	t.Logf("Working directory: %s", subDir)
	t.Logf("HOME directory: %s", tempDir)

	// Try to materialize - should fail at home directory, not continue up
	cmd = exec.Command(binaryPath, "materialize", "skill", "boundary-test", "--target", "opencode")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	t.Logf("Error output:\n%s", outputStr)

	// Should fail with project not found
	if err == nil {
		t.Fatalf("Expected error when reaching home directory without project, but command succeeded")
	}

	// Verify it stopped at home directory (no project boundary detected)
	if !strings.Contains(outputStr, "no project boundary detected") {
		t.Errorf("Expected error about project boundary not detected, got: %s", outputStr)
	}

	t.Logf("Correctly stopped at home directory boundary")
}

// TestMaterializeWithRelativeProjectDir verifies that --project-dir
// works with relative paths.
func TestMaterializeWithRelativeProjectDir(t *testing.T) {
	// Save current directory and restore it after test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(originalDir)
	})

	tempDir := testutil.CreateTempDir(t, "agent-smith-materialize-relative-*")
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

	// Create project structure
	workDir := filepath.Join(tempDir, "workspace")
	projectDir := filepath.Join(workDir, "my-project")
	opencodeDir := filepath.Join(projectDir, ".opencode")
	err = os.MkdirAll(opencodeDir, 0755)
	testutil.AssertNoError(t, err, "Failed to create project directory")

	// Setup test skill
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	skillDir := filepath.Join(agentSmithDir, "skills", "relative-test")
	err = os.MkdirAll(skillDir, 0755)
	testutil.AssertNoError(t, err, "Failed to create skill directory")

	skillContent := "# Relative Test Skill"
	err = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0644)
	testutil.AssertNoError(t, err, "Failed to write skill file")

	// Create lock file
	lockFilePath := filepath.Join(agentSmithDir, ".component-lock.json")
	lockData := map[string]interface{}{
		"version": 3,
		"skills": map[string]interface{}{
			"relative-test": map[string]interface{}{
				"source":       "test-repo",
				"sourceType":   "github",
				"sourceUrl":    "https://github.com/test/relative",
				"commitHash":   "mno345",
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

	// Change to workspace directory
	err = os.Chdir(workDir)
	testutil.AssertNoError(t, err, "Failed to change to workspace directory")

	t.Logf("Working directory: %s", workDir)
	t.Logf("Using relative path: ./my-project")

	// Run materialize with relative --project-dir
	cmd = exec.Command(binaryPath, "materialize", "skill", "relative-test", "--target", "opencode", "--project-dir", "./my-project", "--verbose")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	t.Logf("Materialize output:\n%s", outputStr)

	if err != nil {
		t.Fatalf("Materialize with relative --project-dir failed: %v\nOutput: %s", err, outputStr)
	}

	// Verify component was materialized
	destPath := filepath.Join(opencodeDir, "skills", "relative-test", "SKILL.md")
	testutil.AssertFileExists(t, destPath)

	content, err := os.ReadFile(destPath)
	testutil.AssertNoError(t, err, "Failed to read materialized file")
	testutil.AssertEqual(t, skillContent, string(content), "File content mismatch")

	t.Logf("Successfully used relative path with --project-dir")
}
