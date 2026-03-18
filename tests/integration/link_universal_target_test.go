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

// createBaseSkill creates a skill directly in ~/.agent-smith/skills/<name>/ (base installation).
func createBaseSkill(t *testing.T, baseDir, skillName string) {
	t.Helper()
	skillDir := filepath.Join(baseDir, ".agent-smith", "skills", skillName)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("failed to create base skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Skill: "+skillName), 0644); err != nil {
		t.Fatalf("failed to write skill file: %v", err)
	}
}

// createBaseAgent creates an agent directly in ~/.agent-smith/agents/<name>/.
func createBaseAgent(t *testing.T, baseDir, agentName string) {
	t.Helper()
	agentDir := filepath.Join(baseDir, ".agent-smith", "agents", agentName)
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		t.Fatalf("failed to create base agent dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agentDir, agentName+".md"), []byte("# Agent: "+agentName), 0644); err != nil {
		t.Fatalf("failed to write agent file: %v", err)
	}
}

// TestE2E_UniversalLink_ExplicitTargetCreatesDir verifies that `link skill --to universal`
// creates ~/.agents/ when it does not exist and symlinks the skill into it.
func TestE2E_UniversalLink_ExplicitTargetCreatesDir(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-universal-link-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	binaryPath := AgentSmithBinary

	// No ~/.agents/ — it should be created by the link command.
	universalSkillsDir := filepath.Join(tempDir, ".agents", "skills")

	skillName := "universal-test-skill"
	createBaseSkill(t, tempDir, skillName)

	t.Run("Step1_LinkSkillToUniversal", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "skill", skillName, "--to", "universal")
		output, err := cmd.CombinedOutput()
		t.Logf("Output:\n%s", output)
		if err != nil {
			t.Fatalf("link skill --to universal failed: %v\nOutput: %s", err, output)
		}
	})

	t.Run("Step2_SymlinkExistsInUniversalDir", func(t *testing.T) {
		symlinkPath := filepath.Join(universalSkillsDir, skillName)
		if _, err := os.Lstat(symlinkPath); err != nil {
			t.Errorf("expected symlink at %s: %v", symlinkPath, err)
		}
	})

	t.Run("Step3_UniversalDirWasCreated", func(t *testing.T) {
		if _, err := os.Stat(universalSkillsDir); err != nil {
			t.Errorf("expected ~/.agents/skills/ to exist: %v", err)
		}
	})
}

// TestE2E_UniversalLink_AllComponentTypes verifies that agents, skills, and commands can all be
// linked to the universal target.
func TestE2E_UniversalLink_AllComponentTypes(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-universal-all-types-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	binaryPath := AgentSmithBinary

	skillName := "universal-skill"
	agentName := "universal-agent"

	createBaseSkill(t, tempDir, skillName)
	createBaseAgent(t, tempDir, agentName)

	t.Run("Step1_LinkSkill", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "skill", skillName, "--to", "universal")
		output, err := cmd.CombinedOutput()
		t.Logf("Output:\n%s", output)
		if err != nil {
			t.Fatalf("link skill --to universal failed: %v\nOutput: %s", err, output)
		}
		assertSymlinkExists(t, filepath.Join(tempDir, ".agents", "skills", skillName))
	})

	t.Run("Step2_LinkAgent", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "agent", agentName, "--to", "universal")
		output, err := cmd.CombinedOutput()
		t.Logf("Output:\n%s", output)
		if err != nil {
			t.Fatalf("link agent --to universal failed: %v\nOutput: %s", err, output)
		}
		// Agents are linked as flat .md files: <agentName>-<file>.md
		agentsDir := filepath.Join(tempDir, ".agents", "agents")
		entries, err := os.ReadDir(agentsDir)
		if err != nil {
			t.Fatalf("failed to read ~/.agents/agents/: %v", err)
		}
		if len(entries) == 0 {
			t.Error("expected at least one agent entry in ~/.agents/agents/")
		}
	})
}

// TestE2E_UniversalLink_LinkAllAutoIncludes verifies that `link all` automatically includes
// the universal target when ~/.agents/ already exists on disk.
func TestE2E_UniversalLink_LinkAllAutoIncludes(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-universal-auto-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	binaryPath := AgentSmithBinary

	// Pre-create ~/.agents/ to trigger auto-detection.
	universalDir := filepath.Join(tempDir, ".agents")
	if err := os.MkdirAll(universalDir, 0755); err != nil {
		t.Fatalf("failed to create ~/.agents: %v", err)
	}

	skillName := "auto-universal-skill"
	createBaseSkill(t, tempDir, skillName)

	t.Run("Step1_LinkAll", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "all")
		output, err := cmd.CombinedOutput()
		t.Logf("Output:\n%s", output)
		if err != nil {
			t.Fatalf("link all failed: %v\nOutput: %s", err, output)
		}
	})

	t.Run("Step2_SkillSymlinkedInUniversal", func(t *testing.T) {
		symlinkPath := filepath.Join(universalDir, "skills", skillName)
		if _, err := os.Lstat(symlinkPath); err != nil {
			t.Errorf("expected skill symlink in ~/.agents/skills/ after link all: %v", err)
		}
	})
}

// TestE2E_UniversalLink_LinkAllDoesNotCreateDirWhenAbsent verifies that `link all` does NOT
// create ~/.agents/ when it does not already exist.
func TestE2E_UniversalLink_LinkAllDoesNotCreateDirWhenAbsent(t *testing.T) {
	tempDir := testutil.CreateTempDir(t, "agent-smith-e2e-universal-no-create-*")
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Setenv("HOME", oldHome) })

	binaryPath := AgentSmithBinary

	// Pre-create an editor target so link all has at least one target and doesn't error.
	opencodeDir := filepath.Join(tempDir, ".config", "opencode")
	if err := os.MkdirAll(opencodeDir, 0755); err != nil {
		t.Fatalf("failed to create opencode dir: %v", err)
	}

	skillName := "no-universal-skill"
	createBaseSkill(t, tempDir, skillName)

	t.Run("Step1_LinkAll", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "link", "all")
		output, err := cmd.CombinedOutput()
		t.Logf("Output:\n%s", output)
		if err != nil {
			t.Fatalf("link all failed: %v\nOutput: %s", err, output)
		}
	})

	t.Run("Step2_UniversalDirNotCreated", func(t *testing.T) {
		universalDir := filepath.Join(tempDir, ".agents")
		if _, err := os.Stat(universalDir); !os.IsNotExist(err) {
			t.Errorf("expected ~/.agents/ to NOT exist after link all (no pre-existing dir), but it does")
		}
	})
}
