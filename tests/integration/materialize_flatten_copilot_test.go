//go:build integration
// +build integration

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/tjg184/agent-smith/internal/testutil"
)

// TestMaterializeAgentCopilot verifies agents/commands materialize as flat files on the copilot target,
// identical to opencode/claudecode behavior — no subdirectory, no symlink.
func TestMaterializeAgentCopilot(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	t.Cleanup(func() { os.Chdir(originalDir) })

	tempDir := testutil.CreateTempDir(t, "agent-smith-copilot-flat-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	binaryPath := AgentSmithBinary

	t.Run("agent materializes as flat file on copilot", func(t *testing.T) {
		baseDir := filepath.Join(tempDir, ".agent-smith")
		agentName := "test-agent"
		agentsDir := filepath.Join(baseDir, "agents", agentName)

		if err := os.MkdirAll(agentsDir, 0755); err != nil {
			t.Fatalf("Failed to create agent directory: %v", err)
		}

		agentContent := "# Test Agent\nA test agent."
		if err := os.WriteFile(filepath.Join(agentsDir, agentName+".md"), []byte(agentContent), 0644); err != nil {
			t.Fatalf("Failed to write agent file: %v", err)
		}

		lockFilePath := filepath.Join(baseDir, ".component-lock.json")
		testutil.CreateComponentLockFile(t, lockFilePath, "agents", agentName, "https://github.com/test/repo", map[string]interface{}{
			"source":       "test-repo",
			"sourceType":   "github",
			"sourceUrl":    "https://github.com/test/repo",
			"commitHash":   "abc123",
			"originalPath": "agents/" + agentName + "/" + agentName + ".md",
			"installedAt":  "2024-01-15T10:30:00Z",
			"updatedAt":    "2024-01-15T10:30:00Z",
		})

		projectDir := filepath.Join(tempDir, "test-project-copilot-agent")
		githubDir := filepath.Join(projectDir, ".github")
		if err := os.MkdirAll(githubDir, 0755); err != nil {
			t.Fatalf("Failed to create .github directory: %v", err)
		}

		if err := os.Chdir(projectDir); err != nil {
			t.Fatalf("Failed to change to project directory: %v", err)
		}

		cmd := exec.Command(binaryPath, "materialize", "agent", agentName, "--target", "copilot", "--verbose")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to materialize agent: %v\nOutput: %s", err, string(out))
		}
		t.Logf("Output:\n%s", string(out))

		flatFile := filepath.Join(githubDir, "agents", agentName+".md")
		info, err := os.Lstat(flatFile)
		if err != nil {
			t.Fatalf("Flat agent file not created at %s: %v", flatFile, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			t.Error("Flat agent file must not be a symlink")
		}

		subdir := filepath.Join(githubDir, "agents", agentName)
		if _, err := os.Stat(subdir); err == nil {
			t.Error("Agent subdirectory must not exist for copilot flat copy")
		}

		content, err := os.ReadFile(flatFile)
		if err != nil {
			t.Fatalf("Failed to read flat agent file: %v", err)
		}
		if string(content) != agentContent {
			t.Errorf("Content = %q, want %q", string(content), agentContent)
		}

		t.Log("✓ Agent materialized to copilot as flat file (no subdir, no symlink)")
	})

	t.Run("command materializes as flat file on copilot", func(t *testing.T) {
		baseDir := filepath.Join(tempDir, ".agent-smith")
		commandName := "my-command"
		commandsDir := filepath.Join(baseDir, "commands", commandName)

		if err := os.MkdirAll(commandsDir, 0755); err != nil {
			t.Fatalf("Failed to create command directory: %v", err)
		}

		commandContent := "# My Command"
		if err := os.WriteFile(filepath.Join(commandsDir, commandName+".md"), []byte(commandContent), 0644); err != nil {
			t.Fatalf("Failed to write command file: %v", err)
		}

		lockFilePath := filepath.Join(baseDir, ".component-lock.json")
		testutil.CreateComponentLockFile(t, lockFilePath, "commands", commandName, "https://github.com/test/repo", map[string]interface{}{
			"source":       "test-repo",
			"sourceType":   "github",
			"sourceUrl":    "https://github.com/test/repo",
			"commitHash":   "abc123",
			"originalPath": "commands/" + commandName + "/" + commandName + ".md",
			"installedAt":  "2024-01-15T10:30:00Z",
			"updatedAt":    "2024-01-15T10:30:00Z",
		})

		projectDir := filepath.Join(tempDir, "test-project-copilot-command")
		githubDir := filepath.Join(projectDir, ".github")
		if err := os.MkdirAll(githubDir, 0755); err != nil {
			t.Fatalf("Failed to create .github directory: %v", err)
		}

		if err := os.Chdir(projectDir); err != nil {
			t.Fatalf("Failed to change to project directory: %v", err)
		}

		cmd := exec.Command(binaryPath, "materialize", "command", commandName, "--target", "copilot", "--verbose")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to materialize command: %v\nOutput: %s", err, string(out))
		}

		flatFile := filepath.Join(githubDir, "commands", commandName+".md")
		info, err := os.Lstat(flatFile)
		if err != nil {
			t.Fatalf("Flat command file not created at %s: %v", flatFile, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			t.Error("Flat command file must not be a symlink")
		}

		t.Log("✓ Command materialized to copilot as flat file")
	})

	t.Run("skill materializes as directory on copilot", func(t *testing.T) {
		baseDir := filepath.Join(tempDir, ".agent-smith")
		skillName := "test-skill"
		skillsDir := filepath.Join(baseDir, "skills", skillName)

		if err := os.MkdirAll(skillsDir, 0755); err != nil {
			t.Fatalf("Failed to create skill directory: %v", err)
		}

		if err := os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte("# Test Skill"), 0644); err != nil {
			t.Fatalf("Failed to write skill file: %v", err)
		}

		lockFilePath := filepath.Join(baseDir, ".component-lock.json")
		testutil.CreateComponentLockFile(t, lockFilePath, "skills", skillName, "https://github.com/test/repo", map[string]interface{}{
			"source":       "test-repo",
			"sourceType":   "github",
			"sourceUrl":    "https://github.com/test/repo",
			"commitHash":   "abc123",
			"originalPath": "skills/" + skillName + "/SKILL.md",
			"installedAt":  "2024-01-15T10:30:00Z",
			"updatedAt":    "2024-01-15T10:30:00Z",
		})

		projectDir := filepath.Join(tempDir, "test-project-copilot-skill")
		githubDir := filepath.Join(projectDir, ".github")
		if err := os.MkdirAll(githubDir, 0755); err != nil {
			t.Fatalf("Failed to create .github directory: %v", err)
		}

		if err := os.Chdir(projectDir); err != nil {
			t.Fatalf("Failed to change to project directory: %v", err)
		}

		cmd := exec.Command(binaryPath, "materialize", "skill", skillName, "--target", "copilot", "--verbose")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to materialize skill: %v\nOutput: %s", err, string(out))
		}

		skillDir := filepath.Join(githubDir, "skills", skillName)
		if _, err := os.Stat(skillDir); os.IsNotExist(err) {
			t.Error("Skill directory was not created")
		}

		flatFile := filepath.Join(githubDir, "skills", skillName+".md")
		if _, err := os.Lstat(flatFile); err == nil {
			t.Error("Flat symlink must not be created for skills")
		}

		t.Log("✓ Skill materialized to copilot as directory (no flat symlink)")
	})

	t.Run("multi-file agent materializes as multiple flat files on copilot", func(t *testing.T) {
		baseDir := filepath.Join(tempDir, ".agent-smith")
		agentName := "backend-development"
		agentsDir := filepath.Join(baseDir, "agents", agentName)

		if err := os.MkdirAll(agentsDir, 0755); err != nil {
			t.Fatalf("Failed to create agent directory: %v", err)
		}

		agentFiles := map[string]string{
			"tdd-orchestrator.md":         "# TDD Orchestrator",
			"temporal-python-pro.md":      "# Temporal Python Pro",
			"event-sourcing-architect.md": "# Event Sourcing Architect",
		}
		for filename, content := range agentFiles {
			if err := os.WriteFile(filepath.Join(agentsDir, filename), []byte(content), 0644); err != nil {
				t.Fatalf("Failed to write %s: %v", filename, err)
			}
		}

		lockFilePath := filepath.Join(baseDir, ".component-lock.json")
		testutil.CreateComponentLockFile(t, lockFilePath, "agents", agentName, "https://github.com/test/repo", map[string]interface{}{
			"source":       "test-repo",
			"sourceType":   "github",
			"sourceUrl":    "https://github.com/test/repo",
			"commitHash":   "abc123",
			"originalPath": "agents/" + agentName,
			"installedAt":  "2024-01-15T10:30:00Z",
			"updatedAt":    "2024-01-15T10:30:00Z",
		})

		projectDir := filepath.Join(tempDir, "test-project-copilot-multifile")
		githubDir := filepath.Join(projectDir, ".github")
		if err := os.MkdirAll(githubDir, 0755); err != nil {
			t.Fatalf("Failed to create .github directory: %v", err)
		}

		if err := os.Chdir(projectDir); err != nil {
			t.Fatalf("Failed to change to project directory: %v", err)
		}

		cmd := exec.Command(binaryPath, "materialize", "agent", agentName, "--target", "copilot", "--verbose")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to materialize agent: %v\nOutput: %s", err, string(out))
		}
		t.Logf("Output:\n%s", string(out))

		agentsTargetDir := filepath.Join(githubDir, "agents")
		for filename := range agentFiles {
			flatFile := filepath.Join(agentsTargetDir, filename)
			info, err := os.Lstat(flatFile)
			if err != nil {
				t.Errorf("Flat file %s not created: %v", filename, err)
				continue
			}
			if info.Mode()&os.ModeSymlink != 0 {
				t.Errorf("Flat file %s must not be a symlink", filename)
			}
		}

		subdir := filepath.Join(agentsTargetDir, agentName)
		if _, err := os.Stat(subdir); err == nil {
			t.Error("Agent subdirectory must not exist for flat copy")
		}

		t.Log("✓ Multi-file agent materialized to copilot as flat files")
	})
}
