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

// TestMaterializeAllComponentTypes verifies that skills, agents, and commands
// can all be materialized to project directories with proper metadata tracking.
// This test covers Story-002 acceptance criteria.
func TestMaterializeAllComponentTypes(t *testing.T) {
	// Save current directory and restore it after test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(originalDir)
	})

	// Create temporary directory and set HOME
	tempDir := testutil.CreateTempDir(t, "agent-smith-materialize-all-*")
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

	// Setup test components in ~/.agent-smith/
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")

	// Test data for each component type
	componentData := []struct {
		componentType string
		componentName string
		fileName      string
		content       string
	}{
		{
			componentType: "skills",
			componentName: "test-skill",
			fileName:      "SKILL.md",
			content:       "# Test Skill\nThis is a test skill for materialization.",
		},
		{
			componentType: "agents",
			componentName: "test-agent",
			fileName:      "AGENT.md",
			content:       "# Test Agent\nThis is a test agent for materialization.",
		},
		{
			componentType: "commands",
			componentName: "test-command",
			fileName:      "COMMAND.md",
			content:       "# Test Command\nThis is a test command for materialization.",
		},
	}

	// Create test components and lock files
	for _, data := range componentData {
		componentDir := filepath.Join(agentSmithDir, data.componentType, data.componentName)
		err := os.MkdirAll(componentDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create component directory")

		filePath := filepath.Join(componentDir, data.fileName)
		err = os.WriteFile(filePath, []byte(data.content), 0644)
		testutil.AssertNoError(t, err, "Failed to write component file")

		// Create lock file entry
		lockFilePath := filepath.Join(agentSmithDir, "."+data.componentType[:len(data.componentType)-1]+"-lock.json")
		lockData := map[string]interface{}{
			"version": 3,
			data.componentType: map[string]interface{}{
				data.componentName: map[string]interface{}{
					"source":       "test-repo",
					"sourceType":   "github",
					"sourceUrl":    "https://github.com/test/repo",
					"commitHash":   "abc123",
					"originalPath": data.fileName,
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

		t.Logf("Created test %s: %s", data.componentType, data.componentName)
	}

	// Test materialization for each component type
	for _, data := range componentData {
		t.Run("Materialize_"+data.componentType+"_"+data.componentName, func(t *testing.T) {
			// Change to project directory for auto-detection
			err := os.Chdir(projectDir)
			testutil.AssertNoError(t, err, "Failed to change to project directory")

			// Run materialize command
			cmd := exec.Command(binaryPath, "materialize", data.componentType[:len(data.componentType)-1], data.componentName, "--target", "opencode", "--verbose")
			output, err := cmd.CombinedOutput()
			outputStr := string(output)
			t.Logf("Materialize output:\n%s", outputStr)

			if err != nil {
				t.Fatalf("Materialize failed: %v\nOutput: %s", err, outputStr)
			}

			// Verify component was copied
			destPath := filepath.Join(opencodeDir, data.componentType, data.componentName, data.fileName)
			testutil.AssertFileExists(t, destPath)

			// Verify file content matches
			content, err := os.ReadFile(destPath)
			testutil.AssertNoError(t, err, "Failed to read materialized file")
			testutil.AssertEqual(t, data.content, string(content), "File content mismatch")

			t.Logf("Successfully verified %s was materialized", data.componentName)
		})
	}

	// Verify metadata file exists and contains all components
	t.Run("VerifyMetadata", func(t *testing.T) {
		metadataPath := filepath.Join(opencodeDir, ".component-lock.json")
		testutil.AssertFileExists(t, metadataPath)

		// Load and parse metadata
		metadataBytes, err := os.ReadFile(metadataPath)
		testutil.AssertNoError(t, err, "Failed to read metadata file")

		var metadata project.MaterializationMetadata
		err = json.Unmarshal(metadataBytes, &metadata)
		testutil.AssertNoError(t, err, "Failed to parse metadata")

		// Verify metadata structure (version 5 with nested maps)
		testutil.AssertEqual(t, 5, metadata.Version, "Incorrect metadata version")

		// Verify each component type has entries
		testutil.AssertEqual(t, 1, len(metadata.Skills), "Expected 1 skill source in metadata")
		testutil.AssertEqual(t, 1, len(metadata.Agents), "Expected 1 agent source in metadata")
		testutil.AssertEqual(t, 1, len(metadata.Commands), "Expected 1 command source in metadata")

		// Source URL from the lock file
		sourceURL := "https://github.com/test/repo"

		// Verify skill metadata (nested structure: Skills[sourceURL][componentName])
		skillsFromSource, exists := metadata.Skills[sourceURL]
		testutil.AssertTrue(t, exists, "Source URL not found in skills metadata")
		skillMeta, exists := skillsFromSource["test-skill"]
		testutil.AssertTrue(t, exists, "test-skill not found in metadata")
		testutil.AssertEqual(t, sourceURL, skillMeta.SourceUrl, "Incorrect skill source")
		testutil.AssertEqual(t, "github", skillMeta.SourceType, "Incorrect skill source type")
		testutil.AssertEqual(t, "abc123", skillMeta.CommitHash, "Incorrect skill commit hash")

		// Verify agent metadata
		agentsFromSource, exists := metadata.Agents[sourceURL]
		testutil.AssertTrue(t, exists, "Source URL not found in agents metadata")
		agentMeta, exists := agentsFromSource["test-agent"]
		testutil.AssertTrue(t, exists, "test-agent not found in metadata")
		testutil.AssertEqual(t, sourceURL, agentMeta.SourceUrl, "Incorrect agent source")
		testutil.AssertEqual(t, "github", agentMeta.SourceType, "Incorrect agent source type")
		testutil.AssertEqual(t, "abc123", agentMeta.CommitHash, "Incorrect agent commit hash")

		// Verify command metadata
		commandsFromSource, exists := metadata.Commands[sourceURL]
		testutil.AssertTrue(t, exists, "Source URL not found in commands metadata")
		commandMeta, exists := commandsFromSource["test-command"]
		testutil.AssertTrue(t, exists, "test-command not found in metadata")
		testutil.AssertEqual(t, sourceURL, commandMeta.SourceUrl, "Incorrect command source")
		testutil.AssertEqual(t, "github", commandMeta.SourceType, "Incorrect command source type")
		testutil.AssertEqual(t, "abc123", commandMeta.CommitHash, "Incorrect command commit hash")

		t.Logf("Metadata verification passed for all component types")
	})

	// Test materializing to both targets (all)
	t.Run("MaterializeToAllTargets", func(t *testing.T) {
		// Create .claude directory
		claudeDir := filepath.Join(projectDir, ".claude")
		err := os.MkdirAll(claudeDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create .claude directory")

		err = os.Chdir(projectDir)
		testutil.AssertNoError(t, err, "Failed to change to project directory")

		// Materialize a skill to all targets
		cmd := exec.Command(binaryPath, "materialize", "skill", "test-skill", "--target", "all", "--verbose")
		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		t.Logf("Materialize all output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Materialize all failed: %v\nOutput: %s", err, outputStr)
		}

		// Verify component exists in claudecode target
		destPath := filepath.Join(claudeDir, "skills", "test-skill", "SKILL.md")
		testutil.AssertFileExists(t, destPath)

		// Verify metadata exists in claudecode target
		claudeMetadataPath := filepath.Join(claudeDir, ".component-lock.json")
		testutil.AssertFileExists(t, claudeMetadataPath)

		// Load and verify Claude metadata
		metadataBytes, err := os.ReadFile(claudeMetadataPath)
		testutil.AssertNoError(t, err, "Failed to read Claude metadata")

		var claudeMetadata project.MaterializationMetadata
		err = json.Unmarshal(metadataBytes, &claudeMetadata)
		testutil.AssertNoError(t, err, "Failed to parse Claude metadata")

		testutil.AssertEqual(t, 1, len(claudeMetadata.Skills), "Expected 1 skill in Claude metadata")

		t.Logf("Successfully materialized to all targets")
	})

	// Test directory structure creation
	t.Run("VerifyDirectoryStructure", func(t *testing.T) {
		// Verify opencode structure
		for _, subdir := range []string{"skills", "agents", "commands"} {
			dirPath := filepath.Join(opencodeDir, subdir)
			testutil.AssertDirectoryExists(t, dirPath)
		}

		// Verify claude structure
		claudeDir := filepath.Join(projectDir, ".claude")
		for _, subdir := range []string{"skills", "agents", "commands"} {
			dirPath := filepath.Join(claudeDir, subdir)
			testutil.AssertDirectoryExists(t, dirPath)
		}

		t.Logf("Directory structure verified for all targets")
	})
}

// TestMaterializeComponentNotFound verifies proper error handling when
// component doesn't exist in ~/.agent-smith/
func TestMaterializeComponentNotFound(t *testing.T) {
	// Save current directory and restore it after test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(originalDir)
	})

	tempDir := testutil.CreateTempDir(t, "agent-smith-materialize-notfound-*")
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

	// Create project directory with .opencode
	projectDir := filepath.Join(tempDir, "test-project")
	opencodeDir := filepath.Join(projectDir, ".opencode")
	err = os.MkdirAll(opencodeDir, 0755)
	testutil.AssertNoError(t, err, "Failed to create project directory")

	err = os.Chdir(projectDir)
	testutil.AssertNoError(t, err, "Failed to change to project directory")

	// Try to materialize non-existent agent
	cmd = exec.Command(binaryPath, "materialize", "agent", "non-existent-agent", "--target", "opencode")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Should fail
	if err == nil {
		t.Fatalf("Expected error for non-existent agent, but command succeeded")
	}

	// Should contain helpful error message
	if !contains(outputStr, "not found") {
		t.Errorf("Expected error message to contain 'not found', got: %s", outputStr)
	}

	t.Logf("Error handling verified: %s", outputStr)
}

// TestMaterializeRecursiveDirectoryStructure verifies that nested directory
// structures are properly copied during materialization
func TestMaterializeRecursiveDirectoryStructure(t *testing.T) {
	// Save current directory and restore it after test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(originalDir)
	})

	tempDir := testutil.CreateTempDir(t, "agent-smith-materialize-recursive-*")
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
	opencodeDir := filepath.Join(projectDir, ".opencode")
	err = os.MkdirAll(opencodeDir, 0755)
	testutil.AssertNoError(t, err, "Failed to create project directory")

	// Create command with nested directory structure
	agentSmithDir := filepath.Join(tempDir, ".agent-smith")
	commandDir := filepath.Join(agentSmithDir, "commands", "complex-command")
	nestedDir := filepath.Join(commandDir, "subdir", "nested")
	err = os.MkdirAll(nestedDir, 0755)
	testutil.AssertNoError(t, err, "Failed to create nested directory")

	// Create files at different levels
	files := map[string]string{
		filepath.Join(commandDir, "COMMAND.md"):                    "# Main command",
		filepath.Join(commandDir, "subdir", "helper.js"):           "// Helper script",
		filepath.Join(commandDir, "subdir", "nested", "data.json"): `{"test": true}`,
	}

	for filePath, content := range files {
		err = os.WriteFile(filePath, []byte(content), 0644)
		testutil.AssertNoError(t, err, "Failed to write file: "+filePath)
	}

	// Create lock file
	lockFilePath := filepath.Join(agentSmithDir, ".component-lock.json")
	lockData := map[string]interface{}{
		"version": 3,
		"commands": map[string]interface{}{
			"complex-command": map[string]interface{}{
				"source":       "test-repo",
				"sourceType":   "github",
				"sourceUrl":    "https://github.com/test/commands",
				"commitHash":   "xyz789",
				"originalPath": "complex-command/COMMAND.md",
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

	err = os.Chdir(projectDir)
	testutil.AssertNoError(t, err, "Failed to change to project directory")

	// Materialize the command
	cmd = exec.Command(binaryPath, "materialize", "command", "complex-command", "--target", "opencode", "--verbose")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	t.Logf("Materialize output:\n%s", outputStr)

	if err != nil {
		t.Fatalf("Materialize failed: %v\nOutput: %s", err, outputStr)
	}

	// Verify all files were copied with correct structure
	destBase := filepath.Join(opencodeDir, "commands", "complex-command")
	for originalPath, expectedContent := range files {
		// Convert original path to destination path
		relPath, err := filepath.Rel(commandDir, originalPath)
		testutil.AssertNoError(t, err, "Failed to get relative path")
		destPath := filepath.Join(destBase, relPath)

		testutil.AssertFileExists(t, destPath)

		content, err := os.ReadFile(destPath)
		testutil.AssertNoError(t, err, "Failed to read file: "+destPath)
		testutil.AssertEqual(t, expectedContent, string(content), "Content mismatch for: "+destPath)
	}

	t.Logf("Recursive directory structure preserved correctly")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
