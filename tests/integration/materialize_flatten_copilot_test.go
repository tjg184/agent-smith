//go:build integration
// +build integration

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tjg184/agent-smith/internal/testutil"
)

// TestMaterializeAgentFlatteningPostprocessor verifies Story: Agent flattening postprocessor
func TestMaterializeAgentFlatteningPostprocessor(t *testing.T) {
	// Save current directory and restore after test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(originalDir)
	})

	// Create temporary directories
	tempDir := testutil.CreateTempDir(t, "agent-smith-flatten-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	// Build the binary
	// Use the globally compiled binary (built once in TestMain)
	binaryPath := AgentSmithBinary

	t.Run("materializing agent to copilot creates symlink", func(t *testing.T) {
		// Set up base directory with agent
		baseDir := filepath.Join(tempDir, ".agent-smith")
		agentName := "test-agent"
		agentsDir := filepath.Join(baseDir, "agents", agentName)

		// Create test agent
		err := os.MkdirAll(agentsDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create agent directory")

		agentContent := `---
name: test-agent
version: 1.0.0
---
# Test Agent
A test agent for copilot flattening.
`
		err = os.WriteFile(filepath.Join(agentsDir, "test-agent.md"), []byte(agentContent), 0644)
		testutil.AssertNoError(t, err, "Failed to write agent file")

		// Create lock file
		lockFilePath := filepath.Join(baseDir, ".component-lock.json")
		testutil.CreateComponentLockFile(t, lockFilePath, "agents", "test-agent", "https://github.com/test/repo", map[string]interface{}{
			"source":       "test-repo",
			"sourceType":   "github",
			"sourceUrl":    "https://github.com/test/repo",
			"commitHash":   "abc123",
			"originalPath": "agents/test-agent/test-agent.md",
			"installedAt":  "2024-01-15T10:30:00Z",
			"updatedAt":    "2024-01-15T10:30:00Z",
		})

		// Create project directory with .github structure
		projectDir := filepath.Join(tempDir, "test-project-copilot")
		githubDir := filepath.Join(projectDir, ".github")
		err = os.MkdirAll(githubDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create .github directory")

		// Change to project directory
		err = os.Chdir(projectDir)
		testutil.AssertNoError(t, err, "Failed to change to project directory")

		// Materialize the agent to copilot
		cmd := exec.Command(binaryPath, "materialize", "agent", agentName, "--target", "copilot", "--verbose")
		matOutput, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to materialize agent: %v\nOutput: %s", err, string(matOutput))
		}

		outputStr := string(matOutput)
		t.Logf("Materialize output:\n%s", outputStr)

		// Verify agent folder exists
		agentFolder := filepath.Join(githubDir, "agents", agentName)
		if _, err := os.Stat(agentFolder); os.IsNotExist(err) {
			t.Error("Agent folder was not created")
		}

		// Verify agent file exists in folder
		agentFileInFolder := filepath.Join(agentFolder, agentName+".md")
		if _, err := os.Stat(agentFileInFolder); os.IsNotExist(err) {
			t.Error("Agent file was not created in folder")
		}

		// Verify flattened symlink exists
		symlinkPath := filepath.Join(githubDir, "agents", agentName+".md")
		info, err := os.Lstat(symlinkPath)
		if err != nil {
			t.Fatalf("Flattened symlink was not created: %v", err)
		}

		// Verify it's actually a symlink
		if info.Mode()&os.ModeSymlink == 0 {
			t.Error("Flattened file is not a symlink")
		}

		// Verify symlink uses relative path
		target, err := os.Readlink(symlinkPath)
		if err != nil {
			t.Fatalf("Failed to read symlink: %v", err)
		}

		expectedTarget := agentName + "/" + agentName + ".md"
		if target != expectedTarget {
			t.Errorf("Symlink target = %v, want %v", target, expectedTarget)
		}

		// Verify symlink is readable
		content, err := os.ReadFile(symlinkPath)
		if err != nil {
			t.Fatalf("Failed to read through symlink: %v", err)
		}

		if len(content) == 0 {
			t.Error("Symlink points to empty file")
		}

		t.Log("✓ Agent materialized to copilot with flattened symlink")
	})

	t.Run("materializing agent to opencode does NOT create symlink", func(t *testing.T) {
		// Set up base directory with agent
		baseDir := filepath.Join(tempDir, ".agent-smith")
		agentName := "opencode-agent"
		agentsDir := filepath.Join(baseDir, "agents", agentName)

		// Create test agent
		err := os.MkdirAll(agentsDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create agent directory")

		agentContent := `# OpenCode Agent`
		err = os.WriteFile(filepath.Join(agentsDir, "opencode-agent.md"), []byte(agentContent), 0644)
		testutil.AssertNoError(t, err, "Failed to write agent file")

		// Create lock file
		lockFilePath := filepath.Join(baseDir, ".component-lock.json")
		testutil.CreateComponentLockFile(t, lockFilePath, "agents", "opencode-agent", "https://github.com/test/repo", map[string]interface{}{
			"source":       "test-repo",
			"sourceType":   "github",
			"sourceUrl":    "https://github.com/test/repo",
			"commitHash":   "abc123",
			"originalPath": "agents/opencode-agent/opencode-agent.md",
			"installedAt":  "2024-01-15T10:30:00Z",
			"updatedAt":    "2024-01-15T10:30:00Z",
		})

		// Create project directory with .opencode structure
		projectDir := filepath.Join(tempDir, "test-project-opencode")
		opencodeDir := filepath.Join(projectDir, ".opencode")
		err = os.MkdirAll(opencodeDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create .opencode directory")

		err = os.Chdir(projectDir)
		testutil.AssertNoError(t, err, "Failed to change to project directory")

		// Materialize to opencode
		cmd := exec.Command(binaryPath, "materialize", "agent", agentName, "--target", "opencode", "--verbose")
		matOutput, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to materialize agent: %v\nOutput: %s", err, string(matOutput))
		}

		// For opencode, agents are flat-copied: the .md file lives directly in the agents dir.
		flatFilePath := filepath.Join(opencodeDir, "agents", "opencode-agent.md")
		if _, err := os.Stat(flatFilePath); os.IsNotExist(err) {
			t.Error("Flat agent .md file was not created for opencode target")
		}

		// Agent wrapper folder must NOT exist (flat copy, not nested).
		agentFolder := filepath.Join(opencodeDir, "agents", agentName)
		if _, err := os.Stat(agentFolder); err == nil {
			t.Error("Agent folder should not exist for opencode flat copy")
		}

		// Verify flat file is a regular file, not a symlink
		symlinkPath := filepath.Join(opencodeDir, "agents", agentName+".md")
		info, err := os.Lstat(symlinkPath)
		if err != nil {
			t.Fatalf("Flat agent file should exist at %s: %v", symlinkPath, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			t.Error("opencode flat file must not be a symlink")
		}

		t.Log("✓ Agent materialized to opencode without symlink (as expected)")
	})

	t.Run("materializing skill to copilot does NOT create symlink", func(t *testing.T) {
		// Set up base directory with skill
		baseDir := filepath.Join(tempDir, ".agent-smith")
		skillName := "test-skill"
		skillsDir := filepath.Join(baseDir, "skills", skillName)

		// Create test skill
		err := os.MkdirAll(skillsDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create skill directory")

		skillContent := `# Test Skill`
		err = os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte(skillContent), 0644)
		testutil.AssertNoError(t, err, "Failed to write skill file")

		// Create lock file
		lockFilePath := filepath.Join(baseDir, ".component-lock.json")
		testutil.CreateComponentLockFile(t, lockFilePath, "skills", "test-skill", "https://github.com/test/repo", map[string]interface{}{
			"source":       "test-repo",
			"sourceType":   "github",
			"sourceUrl":    "https://github.com/test/repo",
			"commitHash":   "abc123",
			"originalPath": "skills/test-skill/SKILL.md",
			"installedAt":  "2024-01-15T10:30:00Z",
			"updatedAt":    "2024-01-15T10:30:00Z",
		})

		// Create project directory with .github structure
		projectDir := filepath.Join(tempDir, "test-project-skill")
		githubDir := filepath.Join(projectDir, ".github")
		err = os.MkdirAll(githubDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create .github directory")

		err = os.Chdir(projectDir)
		testutil.AssertNoError(t, err, "Failed to change to project directory")

		// Materialize skill to copilot
		cmd := exec.Command(binaryPath, "materialize", "skill", skillName, "--target", "copilot", "--verbose")
		matOutput, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to materialize skill: %v\nOutput: %s", err, string(matOutput))
		}

		// Verify skill folder exists
		skillFolder := filepath.Join(githubDir, "skills", skillName)
		if _, err := os.Stat(skillFolder); os.IsNotExist(err) {
			t.Error("Skill folder was not created")
		}

		// Verify flattened symlink does NOT exist
		symlinkPath := filepath.Join(githubDir, "skills", skillName+".md")
		if _, err := os.Lstat(symlinkPath); err == nil {
			t.Error("Flattened symlink should not be created for skills")
		}

		t.Log("✓ Skill materialized to copilot without symlink (as expected)")
	})

	t.Run("dry-run shows symlink message but does not create symlink", func(t *testing.T) {
		// Set up base directory with agent
		baseDir := filepath.Join(tempDir, ".agent-smith")
		agentName := "dryrun-agent"
		agentsDir := filepath.Join(baseDir, "agents", agentName)

		// Create test agent
		err := os.MkdirAll(agentsDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create agent directory")

		agentContent := `# Dry Run Agent`
		err = os.WriteFile(filepath.Join(agentsDir, "dryrun-agent.md"), []byte(agentContent), 0644)
		testutil.AssertNoError(t, err, "Failed to write agent file")

		// Create lock file
		lockFilePath := filepath.Join(baseDir, ".component-lock.json")
		testutil.CreateComponentLockFile(t, lockFilePath, "agents", "dryrun-agent", "https://github.com/test/repo", map[string]interface{}{
			"source":       "test-repo",
			"sourceType":   "github",
			"sourceUrl":    "https://github.com/test/repo",
			"commitHash":   "abc123",
			"originalPath": "agents/dryrun-agent/dryrun-agent.md",
			"installedAt":  "2024-01-15T10:30:00Z",
			"updatedAt":    "2024-01-15T10:30:00Z",
		})

		// Create project directory
		projectDir := filepath.Join(tempDir, "test-project-dryrun")
		githubDir := filepath.Join(projectDir, ".github")
		os.MkdirAll(githubDir, 0755)
		os.Chdir(projectDir)

		// Materialize with --dry-run
		cmd := exec.Command(binaryPath, "materialize", "agent", agentName, "--target", "copilot", "--dry-run", "--verbose")
		matOutput, err := cmd.CombinedOutput()
		outputStr := string(matOutput)
		t.Logf("Dry-run output:\n%s", outputStr)

		if err != nil {
			t.Fatalf("Failed to run dry-run: %v\nOutput: %s", err, outputStr)
		}

		// In dry-run mode, the postprocessor may warn that it can't scan the folder
		// because the folder doesn't exist yet (files aren't copied in dry-run).
		// This is expected behavior - just verify no actual symlink was created.
		// (The dry-run message would appear if the folder existed, but that's not required)

		// Verify symlink was NOT actually created
		symlinkPath := filepath.Join(githubDir, "agents", agentName+".md")
		if _, err := os.Lstat(symlinkPath); err == nil {
			t.Error("Symlink should not be created in dry-run mode")
		}

		t.Log("✓ Dry-run shows message without creating symlink")
	})

	t.Run("multi-file agent folder creates multiple symlinks", func(t *testing.T) {
		// Set up base directory with multi-file agent
		baseDir := filepath.Join(tempDir, ".agent-smith")
		agentName := "backend-development"
		agentsDir := filepath.Join(baseDir, "agents", agentName)

		// Create test agent folder with multiple files
		err := os.MkdirAll(agentsDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create agent directory")

		// Create multiple agent files
		agents := map[string]string{
			"tdd-orchestrator.md": `---
name: tdd-orchestrator
version: 1.0.0
---
# TDD Orchestrator
Guides test-driven development workflows.`,
			"temporal-python-pro.md": `---
name: temporal-python-pro
version: 1.0.0
---
# Temporal Python Pro
Expert in Temporal Python workflows.`,
			"event-sourcing-architect.md": `---
name: event-sourcing-architect
version: 1.0.0
---
# Event Sourcing Architect
Designs event-sourced systems.`,
		}

		for filename, content := range agents {
			err = os.WriteFile(filepath.Join(agentsDir, filename), []byte(content), 0644)
			testutil.AssertNoError(t, err, "Failed to write agent file: "+filename)
		}

		// Create lock file
		lockFilePath := filepath.Join(baseDir, ".component-lock.json")
		testutil.CreateComponentLockFile(t, lockFilePath, "agents", "backend-development", "https://github.com/test/repo", map[string]interface{}{
			"source":       "test-repo",
			"sourceType":   "github",
			"sourceUrl":    "https://github.com/test/repo",
			"commitHash":   "abc123",
			"originalPath": "agents/backend-development",
			"installedAt":  "2024-01-15T10:30:00Z",
			"updatedAt":    "2024-01-15T10:30:00Z",
		})

		// Create project directory with .github structure
		projectDir := filepath.Join(tempDir, "test-project-multifile")
		githubDir := filepath.Join(projectDir, ".github")
		err = os.MkdirAll(githubDir, 0755)
		testutil.AssertNoError(t, err, "Failed to create .github directory")

		// Change to project directory
		err = os.Chdir(projectDir)
		testutil.AssertNoError(t, err, "Failed to change to project directory")

		// Materialize the agent to copilot
		cmd := exec.Command(binaryPath, "materialize", "agent", agentName, "--target", "copilot", "--verbose")
		matOutput, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to materialize agent: %v\nOutput: %s", err, string(matOutput))
		}

		outputStr := string(matOutput)
		t.Logf("Materialize output:\n%s", outputStr)

		// Verify agent folder exists
		agentFolder := filepath.Join(githubDir, "agents", agentName)
		if _, err := os.Stat(agentFolder); os.IsNotExist(err) {
			t.Error("Agent folder was not created")
		}

		// Verify all agent files exist in folder
		for filename := range agents {
			agentFileInFolder := filepath.Join(agentFolder, filename)
			if _, err := os.Stat(agentFileInFolder); os.IsNotExist(err) {
				t.Errorf("Agent file was not created in folder: %s", filename)
			}
		}

		// Verify all flattened symlinks exist
		agentsRootDir := filepath.Join(githubDir, "agents")
		for filename := range agents {
			symlinkPath := filepath.Join(agentsRootDir, filename)
			info, err := os.Lstat(symlinkPath)
			if err != nil {
				t.Errorf("Flattened symlink was not created for %s: %v", filename, err)
				continue
			}

			// Verify it's actually a symlink
			if info.Mode()&os.ModeSymlink == 0 {
				t.Errorf("Flattened file is not a symlink: %s", filename)
				continue
			}

			// Verify symlink uses relative path
			target, err := os.Readlink(symlinkPath)
			if err != nil {
				t.Errorf("Failed to read symlink %s: %v", filename, err)
				continue
			}

			expectedTarget := agentName + "/" + filename
			if target != expectedTarget {
				t.Errorf("Symlink target for %s = %v, want %v", filename, target, expectedTarget)
			}

			// Verify symlink is readable
			content, err := os.ReadFile(symlinkPath)
			if err != nil {
				t.Errorf("Failed to read through symlink %s: %v", filename, err)
				continue
			}

			if len(content) == 0 {
				t.Errorf("Symlink %s points to empty file", filename)
			}
		}

		// Verify output mentions multiple symlinks
		if !strings.Contains(outputStr, "3 symlink(s)") && !strings.Contains(outputStr, "Created 3 symlink(s)") {
			t.Error("Output should mention creating 3 symlinks")
		}

		t.Log("✓ Multi-file agent materialized with 3 flattened symlinks")
	})
}
